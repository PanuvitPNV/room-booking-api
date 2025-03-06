// pkg/report/template_funcs.go
package report

import (
	"html/template"
	"sort"
)

// TemplateFuncs returns the template functions used by the timeline report
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		// Calculate percentage
		"percentage": func(part, total int) float64 {
			if total == 0 {
				return 0
			}
			return float64(part) / float64(total) * 100
		},

		// Multiply two floats
		"multiply": func(a, b float64) float64 {
			return a * b
		},

		// Get top N actions
		"topActions": func(actions map[string]int, n int) map[string]int {
			type kv struct {
				Key   string
				Value int
			}

			if len(actions) <= n {
				return actions
			}

			var sorted []kv
			for k, v := range actions {
				sorted = append(sorted, kv{k, v})
			}

			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i].Value > sorted[j].Value
			})

			result := make(map[string]int)
			for i := 0; i < n && i < len(sorted); i++ {
				result[sorted[i].Key] = sorted[i].Value
			}

			return result
		},
	}
}
