package cmd

import (
	"github.com/peak-scale/break-the-glass/api/v1alpha1"
	"github.com/spf13/cobra"
)

var expireCmd = &cobra.Command{
	Use:   "expire",
	Short: "expire a BreakRequest",
	Args:  cobra.ExactArgs(1),
	Example: `
  # expire an existing BreakRequest
  kubectl break-glass expire grant-admin --namespace default
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name = args[0]
		return runBreakRequestAction(
			func(br *v1alpha1.BreakRequest, user *v1alpha1.AccessEntity) error {
				return br.ExpireRequest(user)
			},
		)
	},
}
