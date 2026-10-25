package controller_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	tmock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/kube-temporal/pkg/testutil"
	"github.com/temporalio/kube-temporal/pkg/testutil/fixtures"
	"github.com/temporalio/kube-temporal/pkg/testutil/kube"
	"github.com/temporalio/kube-temporal/pkg/testutil/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlrt "sigs.k8s.io/controller-runtime"
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

type assertFunc func(
	*testing.T,
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

func TestController_Reconcile_Delete(t *testing.T) {
	conn := fixtures.NewConnection(
		testResourceName,
		fixtures.ConnectionWithKubeNamespace(testKubeNamespace),
	)
	connDeleted := fixtures.NewConnection(
		testResourceName,
		fixtures.ConnectionWithKubeNamespace(testKubeNamespace),
	)
	delTime := metav1.NewTime(time.Now())
	connDeleted.DeletionTimestamp = &delTime

	cases := []struct {
		name   string
		req    ctrlrt.Request
		setup  setupFunc
		assert assertFunc
		exp    ctrlrt.Result
		expErr string
	}{
		{
			name: "delete not called when no deletion ts",
			req: kube.NewResourceRequest(
				testKubeNamespace, testResourceName,
			),
			setup: func(
				ctx context.Context,
				req ctrlrt.Request,
				r *mock.Reconciler,
			) {
				cr := fixtures.ResourceFromConnection(conn)
				r.On(
					"ResourceFromRequest", ctx, req,
				).Return(
					cr, nil,
				)
				r.On(
					"ResourceLatest", tmock.Anything, tmock.Anything,
				).Return(
					nil,
				)
			},
			assert: func(t *testing.T, r *mock.Reconciler) {
				r.AssertNotCalled(t, "ResourceDelete")
			},
			exp: ctrlrt.Result{},
		},
		{
			name: "delete called when deletion ts",
			req: kube.NewResourceRequest(
				testKubeNamespace, testResourceName,
			),
			setup: func(
				ctx context.Context,
				req ctrlrt.Request,
				r *mock.Reconciler,
			) {
				cr := fixtures.ResourceFromConnection(connDeleted)
				r.On(
					"ResourceFromRequest", ctx, req,
				).Return(
					cr, nil,
				)
				r.On(
					"ResourceLatest", tmock.Anything, tmock.Anything,
				).Return(
					nil,
				)
				r.On(
					"ResourceDelete", tmock.Anything, cr,
				).Return(
					nil,
				)
			},
			assert: func(t *testing.T, r *mock.Reconciler) {
				r.AssertNotCalled(t, "ResourceDelete")
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
			if tc.assert != nil {
				tc.assert(tt, r)
			}
			require.EqualExportedValues(tc.exp, got)
		})
	}
}
