package controller_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	tmock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	tcorev1 "github.com/temporalio/kube-temporal/api/core/v1"
	"github.com/temporalio/kube-temporal/pkg/controller/requeue"
	"github.com/temporalio/kube-temporal/pkg/kube/condition"
	"github.com/temporalio/kube-temporal/pkg/testutil"
	"github.com/temporalio/kube-temporal/pkg/testutil/fixtures"
	"github.com/temporalio/kube-temporal/pkg/testutil/kube"
	"github.com/temporalio/kube-temporal/pkg/testutil/mock"
	"github.com/temporalio/kube-temporal/pkg/types"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrlrt "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlrtfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testKubeNamespace = "test-namespace"
	testResourceName  = "test-conn"
)

type setupFunc func(
	context.Context,
	ctrlrt.Request,
	*mock.Reconciler,
)

func TestController_Reconcile_ResourceFromRequest(t *testing.T) {
	cases := []struct {
		name   string
		req    ctrlrt.Request
		setup  setupFunc
		exp    ctrlrt.Result
		expErr string
	}{
		{
			name:   "fail if request is empty",
			req:    ctrlrt.Request{},
			expErr: "request.Name required",
		},
		{
			name: "fail if resourceFromRequest errors",
			req: kube.NewResourceRequest(
				testKubeNamespace, testResourceName,
			),
			setup: func(
				ctx context.Context,
				req ctrlrt.Request,
				r *mock.Reconciler,
			) {
				r.On(
					"ResourceFromRequest", ctx, req,
				).Return(
					nil, fmt.Errorf("failed reading resource"),
				)
			},
			expErr: "failed reading resource",
		},
		{
			name: "empty result when not found",
			req: kube.NewResourceRequest(
				testKubeNamespace, testResourceName,
			),
			setup: func(
				ctx context.Context,
				req ctrlrt.Request,
				r *mock.Reconciler,
			) {
				r.On(
					"ResourceFromRequest", ctx, req,
				).Return(
					nil, kube.NewNotFoundError(testResourceName),
				)
			},
			exp: ctrlrt.Result{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(tt *testing.T) {
			require := require.New(tt)
			r := new(mock.Reconciler)
			r.On("SetConfig", tmock.Anything)
			r.On("SetClient", tmock.Anything)
			r.On("SetReader", tmock.Anything)
			kc := ctrlrtfake.NewFakeClient()
			c := testutil.NewController(kc, kc, r)
			ctx := context.TODO()
			if tc.setup != nil {
				tc.setup(ctx, tc.req, r)
			}
			got, err := c.Reconcile(ctx, tc.req)
			if tc.expErr != "" {
				require.ErrorContains(err, tc.expErr)
				return
			}
			require.Nil(err)
			require.EqualExportedValues(tc.exp, got)
		})
	}
}

// testFinalizerName is the finalizer the Controller attaches to Connection
// resources: the lowercased GroupKind followed by "/finalizer".
const testFinalizerName = "connection.core.temporal.io/finalizer"

func TestController_Reconcile_Delete(t *testing.T) {
	newConn := func() *tcorev1.Connection {
		return fixtures.NewConnection(
			testResourceName,
			fixtures.ConnectionWithKubeNamespace(testKubeNamespace),
		)
	}
	newConnDeleted := func() *tcorev1.Connection {
		c := newConn()
		delTime := metav1.NewTime(time.Now())
		c.DeletionTimestamp = &delTime
		// The fake client refuses objects with a deletion timestamp and no
		// finalizers. A resource being deleted has the Controller's finalizer.
		c.Finalizers = []string{testFinalizerName}
		return c
	}

	cases := []struct {
		name   string
		obj    *tcorev1.Connection
		setup  func(r *mock.Reconciler, cr types.Resource)
		assert func(t *testing.T, r *mock.Reconciler, kc client.Client, cr types.Resource)
		exp    ctrlrt.Result
	}{
		{
			name: "delete not called when no deletion ts",
			obj:  newConn(),
			assert: func(t *testing.T, r *mock.Reconciler, kc client.Client, _ types.Resource) {
				r.AssertNotCalled(t, "ResourceDelete", tmock.Anything, tmock.Anything)
				// The Controller adds its finalizer and requeues before it
				// reads the latest state.
				r.AssertNotCalled(t, "ResourceSetLatest", tmock.Anything, tmock.Anything)
				require.Contains(t, getConnection(t, kc).Finalizers, testFinalizerName)
			},
			exp: ctrlrt.Result{Requeue: true},
		},
		{
			name: "delete called when deletion ts",
			obj:  newConnDeleted(),
			setup: func(r *mock.Reconciler, cr types.Resource) {
				r.On("ResourceDelete", tmock.Anything, cr).Return(cr, nil)
			},
			assert: func(t *testing.T, r *mock.Reconciler, kc client.Client, cr types.Resource) {
				r.AssertCalled(t, "ResourceDelete", tmock.Anything, cr)
				// Removing the last finalizer lets the API server delete the
				// object.
				err := kc.Get(
					context.TODO(),
					client.ObjectKey{Namespace: testKubeNamespace, Name: testResourceName},
					&tcorev1.Connection{},
				)
				require.True(t, apierrors.IsNotFound(err), "expected NotFound, got %v", err)
			},
			exp: ctrlrt.Result{},
		},
		{
			name: "delete in progress keeps finalizer, saves status and requeues",
			obj:  newConnDeleted(),
			setup: func(r *mock.Reconciler, cr types.Resource) {
				latest := cr.DeepCopy()
				condition.SetReconciling(
					latest, metav1.ConditionTrue, "deleting", "Deleting",
				)
				r.On("ResourceDelete", tmock.Anything, cr).Return(
					latest,
					requeue.NeededAfter(errors.New("deleting"), 15*time.Second),
				)
			},
			assert: func(t *testing.T, r *mock.Reconciler, kc client.Client, cr types.Resource) {
				r.AssertCalled(t, "ResourceDelete", tmock.Anything, cr)
				stored := getConnection(t, kc)
				require.Contains(t, stored.Finalizers, testFinalizerName)
				c := condition.Reconciling(fixtures.ResourceFromConnection(stored))
				require.NotNil(t, c)
				require.Equal(t, metav1.ConditionTrue, c.Status)
				require.Equal(t, "Deleting", c.Reason)
				require.Equal(t, "deleting", c.Message)
			},
			exp: ctrlrt.Result{RequeueAfter: 15 * time.Second},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(tt *testing.T) {
			require := require.New(tt)

			scheme := runtime.NewScheme()
			require.NoError(tcorev1.AddToScheme(scheme))
			kc := ctrlrtfake.NewClientBuilder().
				WithScheme(scheme).
				WithStatusSubresource(&tcorev1.Connection{}).
				WithObjects(tc.obj).
				Build()

			r := new(mock.Reconciler)
			r.On("SetConfig", tmock.Anything)
			c := testutil.NewController(kc, kc, r)
			ctx := context.TODO()
			req := kube.NewResourceRequest(testKubeNamespace, testResourceName)

			// Read the object back so the resource carries the stored
			// resourceVersion, as a real ResourceFromRequest would.
			cr := fixtures.ResourceFromConnection(getConnection(tt, kc))
			r.On("ResourceFromRequest", ctx, req).Return(cr, nil)
			if tc.setup != nil {
				tc.setup(r, cr)
			}

			got, err := c.Reconcile(ctx, req)
			require.NoError(err)
			require.EqualExportedValues(tc.exp, got)
			tc.assert(tt, r, kc, cr)
		})
	}
}

// getConnection returns the test Connection stored in the fake client.
func getConnection(t *testing.T, kc client.Client) *tcorev1.Connection {
	t.Helper()
	conn := &tcorev1.Connection{}
	err := kc.Get(
		context.TODO(),
		client.ObjectKey{Namespace: testKubeNamespace, Name: testResourceName},
		conn,
	)
	require.NoError(t, err)
	return conn
}
