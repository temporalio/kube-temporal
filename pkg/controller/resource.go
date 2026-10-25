package controller

import (
	"context"
	"errors"
	"fmt"
	"maps"

	"github.com/samber/lo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlrtutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	ctrlerrors "github.com/temporalio/kube-temporal/pkg/controller/errors"
	ctrlreqlog "github.com/temporalio/kube-temporal/pkg/controller/request/log"
	"github.com/temporalio/kube-temporal/pkg/kube/condition"
	"github.com/temporalio/kube-temporal/pkg/resource"
	"github.com/temporalio/kube-temporal/pkg/types"
)

// resourceEnsureFinalizer ensures that the supplied Resource has a finalizer
// protecting its deletion. If the finalizer was added, this method immediately
// patches the resources's metadata containing the finalizer annotation back to
// the Kubernetes API and returne whether the finalizer was added to the
// Resource.
func (c *Controller) resourceEnsureFinalizer(
	ctx context.Context,
	desired types.Resource,
) (bool, error) {
	log := ctrlreqlog.FromContext(ctx)

	fname := c.FinalizerName(desired)
	obj := desired.ClientObject()
	if ctrlrtutil.ContainsFinalizer(obj, fname) {
		return false, nil
	}
	ctrlrtutil.AddFinalizer(obj, fname)
	log.Debug("added finalizer")

	patch := client.MergeFrom(obj)
	err := c.client.Patch(ctx, obj, patch)
	if err != nil {
		return false, fmt.Errorf("failed patching metadata: %w", err)
	}
	log.Debug("patched metadata")
	return true, nil
}

// resourceDelete ensures that the supplied Resource's backing Temporal Cluster
// is destroyed along with all child dependent resources.
//
// Returns a copy of the Resource that has been updated with a removed
// finalizer in its metadata.
//
// `desired` is NEVER mutated by this method.
func (c *Controller) resourceDelete(
	ctx context.Context,
	desired types.Resource,
) (types.Resource, error) {
	var err error
	log := ctrlreqlog.FromContext(ctx)
	exit := log.Trace("resource.delete")
	defer func() {
		exit(err)
	}()

	// Ask the associated Reconciler to clean up any externally-managed or
	// owned child resources before we remove the finalizer on the supplied
	// Resource which will complete the process of deletion.
	err = c.r.ResourceDelete(ctx, desired)
	if err != nil {
		return nil, err
	}

	latest := desired.DeepCopy()

	fname := c.FinalizerName(desired)
	obj := latest.ClientObject()
	if !ctrlrtutil.ContainsFinalizer(obj, fname) {
		// Nothing to do... resource already deleted.
		return latest, nil
	}
	ctrlrtutil.RemoveFinalizer(obj, fname)
	log.Debug("removed finalizer")

	return latest, nil
}

// resourceConditionsSync examines the supplied [types.Resource]'s collection
// of Condition objects and ensures that appropriate Stalled condition is
// present.
//
// `latest` IS ALWAYS MUTATED by this method.
func (c *Controller) resourceConditionsSync(
	ctx context.Context,
	latest types.Resource,
	reconcileErr error, // error returned from resourceSync.
) {
	if lo.IsNil(latest) {
		return
	}

	var err error
	log := ctrlreqlog.FromContext(ctx)
	exit := log.Trace("resource.conditions.sync")
	defer func() {
		exit(err)
	}()

	if errors.Is(reconcileErr, &ctrlerrors.TerminalError{}) {
		// If we got a terminal error while reconciling the resource, we set
		// its Stalled condition to True to indicate that another reconcile
		// will not change the status of the resource without a change to its
		// desired state.
		condition.SetStalled(
			latest,
			metav1.ConditionFalse,
			"TerminalError",
			fmt.Sprintf(
				"reconciliation encountered a terminal error: %s",
				reconcileErr.Error(),
			),
		)
	} else {
		condition.SetStalled(
			latest,
			metav1.ConditionFalse,
			"ReconcileSucceeded",
			"reconcile succeeded",
		)
	}
}

// resourceStatusPatch patches the Status of the resource in the Kubernetes API
// to match the supplied latest resource.
//
// `desired` is NEVER mutated by this method.
// `latest` is NEVER mutated by this method.
func (c *Controller) resourceStatusPatch(
	ctx context.Context,
	desired types.Resource,
	latest types.Resource,
) error {
	var err error
	log := ctrlreqlog.FromContext(ctx)
	exit := log.Trace("resource.status.patch")
	defer func() {
		exit(err)
	}()

	dobj := desired.DeepCopy().ClientObject()
	lobj := latest.DeepCopy().ClientObject()
	patch := client.MergeFrom(dobj)

	return c.client.Status().Patch(ctx, lobj, patch)
}

// resourceMetadataDifferent returns true if there supplied resources have
// different metadata.
func (c *Controller) resourceMetadataDifferent(
	a, b types.Resource,
) (bool, error) {
	if lo.IsNil(a) != lo.IsNil(b) {
		return true, nil
	}
	amd, err := resource.CleanMetadata(a)
	if err != nil {
		return false, fmt.Errorf(
			"failed getting clean metadata: %w", err,
		)
	}
	bmd, err := resource.CleanMetadata(b)
	if err != nil {
		return false, fmt.Errorf(
			"failed getting clean metadata: %w", err,
		)
	}
	return maps.Equal(amd, bmd), nil
}

// resourceMetadataPatch patches the metadata of the resource in the Kubernetes
// API to match the supplied latest resource.
//
// `desired` is NEVER mutated by this method.
// `latest` is NEVER mutated by this method.
func (c *Controller) resourceMetadataPatch(
	ctx context.Context,
	desired types.Resource,
	latest types.Resource,
) error {
	var err error
	log := ctrlreqlog.FromContext(ctx)
	exit := log.Trace("resource.metadata.patch")
	defer func() {
		exit(err)
	}()

	dobj := desired.DeepCopy().ClientObject()
	lobj := latest.DeepCopy().ClientObject()
	patch := client.MergeFrom(dobj)

	return c.client.Patch(ctx, lobj, patch)
}
