package cmd

import (
	"context"

	"github.com/spf13/cobra"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	addonsv1alpha1 "github.com/peak-scale/break-the-glass/api/v1alpha1"
)

var activateCmd = &cobra.Command{
	Use:   "activate",
	Short: "activate a BreakRequest",
	Args:  cobra.ExactArgs(1),
	Example: `
  # exprire an existing BreakRequest
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

		return ar.ActiveRequest(ctx, k8sClient, user)
	},
}
