package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	addonsv1alpha1 "github.com/peak-scale/break-the-glass/api/v1alpha1"
)

var (
	approveFlag, denyFlag                 bool
	message                               string
	startTimeStr, durationStr, keepForStr string
)

const (
	denyValue    = "deny"
	approveValue = "approve"
)

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Review an BreakRequest",
	Args:  cobra.ExactArgs(1),
	Example: `
  # interactive review
  kubectl accessdev review grant-admin --namespace default

  # non-interactive approve/deny
  kubectl accessdev review grant-admin --namespace default --approve
  kubectl accessdev review grant-admin --namespace default --deny
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

		if ar.Status.Phase != addonsv1alpha1.RequestPhaseRequested {
			return fmt.Errorf("BreakRequest %s is not in Requested phase (already reviewed), current phase: %s", name, ar.Status.Phase)
		}

		props, err := ar.GetReviewProperties(ctx, k8sClient)
		if err != nil {
			return err
		}

		// Parse Flags and Overwrite

		if keepForStr != "" {
			d, err := time.ParseDuration(keepForStr)
			if err != nil {
				return fmt.Errorf("invalid duration %q: %w", keepForStr, err)
			}
			props.KeepFor = metav1.Duration{Duration: d}
		}

		if durationStr != "" {
			d, err := time.ParseDuration(durationStr)
			if err != nil {
				return fmt.Errorf("invalid duration %q: %w", durationStr, err)
			}
			props.Duration = metav1.Duration{Duration: d}
		}

		// Validate Action

		action := ""
		if approveFlag {
			action = approveValue
		} else if denyFlag {
			action = denyValue
		} else {
			printAccessRequestApprovalTable(ar, props)

			var input string
			for {
				fmt.Print("Approve this request? [y/n]: ")
				fmt.Scanln(&input)

				input = strings.ToLower(strings.TrimSpace(input))
				if input == "y" {
					action = approveValue
					break
				} else if input == "n" {
					action = denyValue
					break
				} else {
					fmt.Println("Invalid input. Please type 'y' or 'n'.")
				}
			}
		}

		user := &addonsv1alpha1.AccessEntity{Type: addonsv1alpha1.AccessEntityTypeUser, Name: cfg.Username}

		return retry.OnError(
			retry.DefaultRetry,
			func(err error) bool {
				return ctrlclient.IgnoreNotFound(err) == nil
			},
			func() error {
				if err := k8sClient.Get(ctx, ctrlclient.ObjectKey{Name: name, Namespace: namespace}, ar); err != nil {
					return err
				}
				switch action {
				case approveValue:
					if err := ar.ApproveRequest(user, props, message); err != nil {
						return err
					}
				case denyValue:
					if err := ar.DenyRequest(user, message); err != nil {
						return err
					}
				}

				return k8sClient.Status().Update(ctx, ar)
			},
		)
	},
}

func init() {
	// Register the flag to the `approve` subcommand
	reviewCmd.Flags().BoolVar(&approveFlag, "approve", false, "Approve the request")
	reviewCmd.Flags().BoolVar(&denyFlag, "deny", false, "Deny the request")
	reviewCmd.Flags().StringVarP(&message, "message", "m", "", "Optional review message")
	reviewCmd.Flags().StringVar(&startTimeStr, "start-time", "", "Start time (RFC3339 format, e.g. 2025-07-15T14:45:00Z)")
	reviewCmd.Flags().StringVar(&durationStr, "duration", "", "The Duration this request is available for (e.g. 5m, 1h30m) [Overrites the value from the request, if defined]")
	reviewCmd.Flags().StringVar(&keepForStr, "keep-for", "", "The Duration this request is archived for (e.g. 5m, 1h30m) [Overrites the value from the request, if defined]")

}
