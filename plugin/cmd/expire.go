package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"k8s.io/client-go/util/retry"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	addonsv1alpha1 "github.com/peak-scale/break-the-glass/api/v1alpha1"
)

var expireCmd = &cobra.Command{
	Use:   "expire",
	Short: "expire an BreakRequest",
	Args:  cobra.ExactArgs(1),
	Example: `
  # expire an existing BreakRequest
  kubectl accessdev expire grant-admin --namespace default
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name = args[0]

		ctx := context.Background()
		cfg, err := config.GetConfig()
		if err != nil {
			return err
		}
		k8sClient, err := ctrlclient.New(cfg, ctrlclient.Options{Scheme: scheme})
		if err != nil {
			return err
		}

		ar := &addonsv1alpha1.BreakRequest{}
		if err := k8sClient.Get(ctx, ctrlclient.ObjectKey{Name: name, Namespace: namespace}, ar); err != nil {
			return err
		}

		user := &addonsv1alpha1.AccessEntity{Type: addonsv1alpha1.AccessEntityTypeUser, Name: cfg.Username}

		if err := k8sClient.Get(ctx, ctrlclient.ObjectKey{Name: name, Namespace: namespace}, ar); err != nil {
			return err
		}

		return retry.OnError(
			retry.DefaultRetry,
			func(err error) bool {
				return ctrlclient.IgnoreNotFound(err) == nil
			},
			func() error {
				if err := k8sClient.Get(ctx, ctrlclient.ObjectKey{Name: name, Namespace: namespace}, ar); err != nil {
					return err
				}

				if err := ar.ExpireRequest(user); err != nil {
					return err
				}

				return k8sClient.Status().Update(ctx, ar)
			},
		)
	},
}
