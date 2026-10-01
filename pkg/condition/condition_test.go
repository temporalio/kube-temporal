package condition_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	itest "github.com/temporalio/kube-temporal/internal/testing"
	"github.com/temporalio/kube-temporal/pkg/condition"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kstatus "sigs.k8s.io/cli-utils/pkg/kstatus/status"
)

func TestFirstOfType(t *testing.T) {
	cases := []struct {
		name string
		res  condition.Manager
		ct   fmt.Stringer
		exp  *metav1.Condition
	}{
		{
			name: "empty conditions yields nil",
			res:  &itest.Resource{},
			ct:   kstatus.ConditionReconciling,
		},
		{
			name: "not found condition yields nil",
			res: &itest.Resource{
				Status: itest.ResourceStatusWithConditions(
					&metav1.Condition{
						Type: kstatus.ConditionStalled.String(),
					},
				),
			},
			ct: kstatus.ConditionReconciling,
		},
		{
			name: "found condition yields condition",
			res: &itest.Resource{
				Status: itest.ResourceStatusWithConditions(
					&metav1.Condition{
						Type: kstatus.ConditionReconciling.String(),
					},
				),
			},
			ct: kstatus.ConditionReconciling,
			exp: &metav1.Condition{
				Type: kstatus.ConditionReconciling.String(),
			},
		},
		{
			name: "multiple matched conditions yields first condition",
			res: &itest.Resource{
				Status: itest.ResourceStatusWithConditions(
					&metav1.Condition{
						Type:   kstatus.ConditionReconciling.String(),
						Reason: "1",
					},
					&metav1.Condition{
						Type:   kstatus.ConditionReconciling.String(),
						Reason: "2",
					},
				),
			},
			ct: kstatus.ConditionReconciling,
			exp: &metav1.Condition{
				Type:   kstatus.ConditionReconciling.String(),
				Reason: "1",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(tt *testing.T) {
			require := require.New(tt)
			got := condition.FirstOfType(tc.res, tc.ct)
			if tc.exp == nil {
				require.Nil(got)
			} else {
				require.Equal(tc.exp, got)
			}
		})
	}
}
