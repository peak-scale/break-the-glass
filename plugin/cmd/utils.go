package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"

	addonsv1alpha1 "github.com/peak-scale/break-the-glass/api/v1alpha1"
)

func printAccessRequestApprovalTable(ar *addonsv1alpha1.BreakRequest, arp *addonsv1alpha1.BreakRequestStatusReviewProperties) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.Style().Title.Align = text.AlignCenter

	durStr := "Unlimited"
	if arp.Duration.Duration != 0 {
		durStr = arp.Duration.Duration.String()
	}

	keepStr := "Undefined"
	if arp.KeepFor.Duration != 0 {
		keepStr = arp.KeepFor.Duration.String()
	}

	t.AppendHeader(table.Row{"Field", "Value"})
	t.AppendRows([]table.Row{
		{"Name", ar.Name},
		{"Namespace", ar.Namespace},
		{"Duration", durStr},
		{"Keep", keepStr},
	})

	// Example: printing .status.items nicely as YAML
	for i, item := range arp.Items {
		pretty := prettyRawExtension(item.RawExtension)
		// Multi-line cells are supported; keep them as one cell.
		t.AppendRow(table.Row{
			fmt.Sprintf("Status Item %d", i+1),
			pretty,
		})
	}

	t.Render()
}

// PrettyRawExtension returns human-readable YAML for a RawExtension.
// - If Object is non-nil, it marshals that.
// - Else if Raw contains JSON, it converts JSON -> YAML.
// - Else it returns Raw as string (best-effort).
func prettyRawExtension(re runtime.RawExtension) string {
	// 1) Prefer the typed Object if available
	if re.Object != nil {
		j, err := json.Marshal(re.Object)
		if err == nil {
			if y, errY := yaml.JSONToYAML(j); errY == nil {
				return string(y)
			}
			return string(j) // fallback to JSON string
		}
	}

	// 2) If Raw looks like JSON, convert to YAML
	if len(re.Raw) > 0 {
		if json.Valid(re.Raw) {
			if y, err := yaml.JSONToYAML(re.Raw); err == nil {
				return string(y)
			}
			return string(re.Raw) // fallback to JSON string
		}
		// 3) Not JSON? Return as-is (may already be YAML or plain text)
		return string(re.Raw)
	}

	return "—"
}
