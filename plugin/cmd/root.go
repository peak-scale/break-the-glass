package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	addonsv1alpha1 "github.com/peak-scale/break-the-glass/api/v1alpha1"
)

var (
	name      string
	namespace string
)

var rootCmd = &cobra.Command{
	Use:   "break",
	Short: "Manage BreakRequests",
}

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(addonsv1alpha1.AddToScheme(scheme))
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().
		StringVarP(&namespace, "namespace", "n", "default", "Namespace of the AccessRequest")

	// Add subcommands
	rootCmd.AddCommand(reviewCmd)
	rootCmd.AddCommand(activateCmd)
	rootCmd.AddCommand(expireCmd)
}
