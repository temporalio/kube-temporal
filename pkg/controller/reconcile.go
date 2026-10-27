package controller

import (
	"context"
	"errors"
	"fmt"

	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrlrt "sigs.k8s.io/controller-runtime"

	ctrlerrors "github.com/temporalio/kube-temporal/pkg/controller/errors"
	ctrlreqlog "github.com/temporalio/kube-temporal/pkg/controller/request/log"
	"github.com/temporalio/kube-temporal/pkg/controller/requeue"
	"github.com/temporalio/kube-temporal/pkg/kube/condition"
	"github.com/temporalio/kube-temporal/pkg/types"
)

// Reconcile performs a full reconciliation for the object referred to by the
// Request.
//
// If the returned error is non-nil, the Result is ignored and the request will
// be requeued using exponential backoff. The only exception is if the error is
// a TerminalError in which case no requeuing happens.
//
// If the error is nil and the returned Result has a non-zero
// result.RequeueAfter, the request will be requeued after the specified
// duration.
//
// If the error is nil and result.RequeueAfter is zero and result.Requeue is
// true, the request will be requeued using exponential backoff.
func (c *Controller) Reconcile(
	ctx context.Context,
	req ctrlrt.Request,
) (ctrlrt.Result, error) {
	if req.Name == "" {
		return ctrlrt.Result{}, fmt.Errorf("request.Name required")
	}
	desired, err := c.r.ResourceFromRequest(ctx, req)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// resource wasn't found. just ignore these.
			return ctrlrt.Result{}, nil
		}
		return ctrlrt.Result{}, err
	}

	ctx = ctrlreqlog.ToContext(ctx, c.logger, desired)

	var latest types.Resource

	if desired.IsBeingDeleted() {
		latest, err = c.resourceDelete(ctx, desired)
	} else {
		latest, err = c.reconcileResource(ctx, desired)
	}

	return c.reconcileComplete(ctx, desired, latest, err)
}

// resconcileResource ensures that whatever observed state of the external
// resource backing the supplied Resource matches the supplied desired state in
// the Resource.
//
// `desired` is NEVER mutated by this method.
//
// It returns a copy of the resource that represents the latest observed state.
func (c *Controller) reconcileResource(
	ctx context.Context,
	desired types.Resource,
) (types.Resource, error) {
	var err error
	log := ctrlreqlog.FromContext(ctx)
	exit := log.Trace("resource.sync")
	defer func() {
		exit(err)
	}()

	// Before we do anything at all, ensure that we set a finalizer on the
	// resource and immediately patch the resource's metadata.
	addedFinalizer, err := c.resourceEnsureFinalizer(ctx, desired)
	if err != nil {
		return nil, err
	}
	if addedFinalizer {
		// in resourceEnsureFinalizer, we immediately patched the resource's
		// metadata back to the Kubernetes API server. To avoid needing to
		// track metadata.resourceVersion when attempting to update the
		// resource at any point later, we simply requeue the resource
		// immediately.
		return nil, requeue.Needed(nil)
	}

	// We make a copy of the desired resource and mutate this copy in
	// various submethods. `desired` is never mutated.
	latest := desired.DeepCopy()

	// We clear the resource's Conditions collection at the start of each
	// reconcile loop to ensure that the Status.Conditions always contains the
	// freshest indication of current state.
	condition.Clear(latest)
	defer func() {
		c.resourceConditionsSync(ctx, latest, err)
	}()

	// The derived Controller is responsible for fetching the latest observed
	// state for our resource and mutating it accordingly.
	err = c.r.ResourceSetLatest(ctx, latest)
	if err != nil {
		return latest, err
	}

	// TODO(jaypipes): Use a friendlier delta library that allows us to log the
	// differences between two [types.Resource] instead of just whether they
	// are semantically the same.
	if equality.Semantic.DeepEqual(
		desired.ClientObject(),
		latest.ClientObject(),
	) {
		return latest, nil
	}

	log.Debug("detected difference in desired state")

	err = c.r.ResourceSync(ctx, desired, latest)
	if err != nil {
		return latest, err
	}

	return latest, nil
}

// reconcileComplete patches the latest state of the resource back to the
// Kubernetes API and translates the reconcile error into a Result.
//
// `desired` is NEVER mutated by this method.
// `latest` is NEVER mutated by this method.
//
// If `latest` is not nil, its Status is ALWAYS patched, whether or not `err`
// is nil, so conditions set during a long-running or failed reconcile are
// saved. Its metadata is patched only when `err` is nil.
func (c *Controller) reconcileComplete(
	ctx context.Context,
	desired types.Resource,
	latest types.Resource,
	err error,
) (ctrlrt.Result, error) {
	if !lo.IsNil(latest) {
		// Patch status before metadata: removing the last finalizer lets the
		// API server delete the object, after which a status patch would fail.
		// NotFound is ignored because the object may already be gone.
		patchErr := c.resourceStatusPatch(ctx, desired, latest)
		if patchErr != nil && !apierrors.IsNotFound(patchErr) {
			return ctrlrt.Result{}, fmt.Errorf("failed patching status: %w", patchErr)
		}
		if err == nil {
			diff, mdErr := c.resourceMetadataDifferent(desired, latest)
			if mdErr != nil {
				return ctrlrt.Result{}, mdErr
			}
			if diff {
				patchErr = c.resourceMetadataPatch(ctx, desired, latest)
				if patchErr != nil && !apierrors.IsNotFound(patchErr) {
					return ctrlrt.Result{}, fmt.Errorf("failed patching metadata: %w", patchErr)
				}
			}
		}
	}

	if err == nil {
		if desired.IsBeingDeleted() {
			// We don't requeue on a successful resource deletion.
			return ctrlrt.Result{}, nil
		}
		// By default, we requeue on success after a configurable time to keep
		// resources in sync.
		return ctrlrt.Result{
			RequeueAfter: c.cfg.Reconcile.RequeueOnSuccessDuration,
		}, nil
	}

	var terminalErr *ctrlerrors.TerminalError
	if errors.As(err, &terminalErr) {
		// We don't requeue on a terminal error.
		return ctrlrt.Result{}, nil
	}

	var requeueNeededAfter *requeue.RequeueNeededAfter
	if errors.As(err, &requeueNeededAfter) {
		return ctrlrt.Result{RequeueAfter: requeueNeededAfter.Duration()}, nil
	}

	// The error returned by the reconciler may have indicated an immediate
	// requeue is needed.
	var requeueNeeded *requeue.RequeueNeeded
	if errors.As(err, &requeueNeeded) {
		return ctrlrt.Result{Requeue: true}, nil
	}

	// Let the controller runtime determine the best requeue behaviour when an
	// error occurred.
	return ctrlrt.Result{}, err
}
