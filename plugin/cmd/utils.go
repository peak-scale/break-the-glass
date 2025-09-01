package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/yaml"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"

	addonsv1alpha1 "github.com/peak-scale/break-the-glass/api/v1alpha1"
)

func printAccessRequestApprovalTable(
	br *addonsv1alpha1.BreakRequest,
	brp *addonsv1alpha1.BreakRequestStatusReviewProperties,
	color bool,
) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.Style().Title.Align = text.AlignCenter

	durStr := "Unlimited"
	if brp.Duration.Duration != 0 {
		durStr = brp.Duration.Duration.String()
	}

	keepStr := "Undefined"
	if brp.KeepFor != 0 {
		keepStr = brp.KeepFor.String()
	}

	t.AppendHeader(table.Row{"Field", "Value"})
	t.AppendRows([]table.Row{
		{"Name", colorizeValue(br.Name, color)},
		{"Namespace", colorizeValue(br.Namespace, color)},
		{"ExtendedDuration", colorizeValue(durStr, color)},
		{"Keep", colorizeValue(keepStr, color)},
	})

	// Example: printing .status.items nicely as YAML
	for name, item := range brp.Items {
		content := prettyRawExtension(item)
		if color {
			content = colorizeYAML(content)
		}
		t.AppendSeparator()
		// Multi-line cells are supported; keep them as one cell.
		t.AppendRow(table.Row{
			fmt.Sprintf("Status Item %q", name),
			content,
		})
	}

	t.Render()
}

// PrettyRawExtension returns human-readable YAML for a RawExtension.
// - If Object is non-nil, it marshals that.
// - Else converts JSON -> YAML.
func prettyRawExtension(re runtime.Unstructured) string {
	j, err := json.Marshal(re)
	if err == nil {
		if y, errY := yaml.JSONToYAML(j); errY == nil {
			return string(y)
		}
		return string(j) // fallback to JSON string
	}
	return "-"
}

// colorizeValue applies ANSI colors for YAML using chroma and returns a string suitable for terminal output.
func colorizeValue(src string, color bool) string {
	if !color || src == "" {
		return src
	}
	return colorize(src, chroma.Literator(chroma.Token{Type: chroma.NameTag, Value: src}))
}

// colorizeYAML applies ANSI colors for YAML using chroma and returns a string suitable for terminal output.
func colorizeYAML(src string) string {
	if src == "" {
		return src
	}

	lexer := lexers.Get("yaml")
	if lexer == nil {
		return src
	}

	it, err := lexer.Tokenise(nil, src)
	if err != nil {
		return src
	}
	return colorize(src, it)
}

func colorize(src string, it chroma.Iterator) string {
	// Choose a style; "dracula", "native", "github", etc. Fall back to "native".
	style := styles.Get("native")
	if style == nil {
		style = styles.Fallback
	}
	// Use terminal16m for truecolor; fall back to standard terminal if not supported.
	formatter := formatters.Get("terminal16m")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	var buf strings.Builder
	if err := formatter.Format(&buf, style, it); err != nil {
		return src
	}
	return buf.String()
}

func newK8sClient() (*rest.Config, ctrlclient.Client, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, nil, err
	}
	cl, err := ctrlclient.New(cfg, ctrlclient.Options{Scheme: scheme})
	return cfg, cl, err
}

func runBreakRequestAction(
	action func(br *addonsv1alpha1.BreakRequest, user *addonsv1alpha1.AccessEntity) error,
) error {
	ctx := context.Background()
	cfg, k8sClient, err := newK8sClient()
	if err != nil {
		return err
	}

	user := &addonsv1alpha1.AccessEntity{
		Type: addonsv1alpha1.AccessEntityTypeUser,
		Name: cfg.Username,
	}

	return retry.OnError(
		retry.DefaultRetry,
		func(err error) bool {
			return ctrlclient.IgnoreNotFound(err) == nil
		},
		func() error {
			br := &addonsv1alpha1.BreakRequest{}
			if err := k8sClient.Get(ctx, ctrlclient.ObjectKey{Name: name, Namespace: namespace}, br); err != nil {
				return err
			}

			if err := action(br, user); err != nil {
				return err
			}
			return k8sClient.Status().Update(ctx, br)
		},
	)
}
