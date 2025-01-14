package printer

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type DiffPrinter struct {
	lastPrintTime time.Time
	showTimestamp bool
	// Color functions
	added    func(a ...interface{}) string
	removed  func(a ...interface{}) string
	modified func(a ...interface{}) string
	header   func(a ...interface{}) string
}

func NewPrinter(showTimestamp bool) *DiffPrinter {
	return &DiffPrinter{
		showTimestamp: showTimestamp,
		added:         color.New(color.FgGreen).SprintFunc(),
		removed:       color.New(color.FgRed).SprintFunc(),
		modified:      color.New(color.FgYellow).SprintFunc(),
		header:        color.New(color.FgCyan).SprintFunc(),
	}
}

func (p *DiffPrinter) Print(oldObj, newObj *unstructured.Unstructured) error {
	// Deep copy objects to avoid modifying originals
	oldData := oldObj.DeepCopy()
	newData := newObj.DeepCopy()

	// Get fields for comparison
	oldFields := make(map[string]interface{})
	newFields := make(map[string]interface{})

	// Always include spec
	if spec, ok := oldData.Object["spec"]; ok {
		oldFields["spec"] = spec
	}
	if spec, ok := newData.Object["spec"]; ok {
		newFields["spec"] = spec
	}

	// Include status if not ignored
	if status, ok := oldData.Object["status"]; ok {
		oldFields["status"] = status
	}
	if status, ok := newData.Object["status"]; ok {
		newFields["status"] = status
	}

	// Compare filtered objects and check if there are any differences
	if p.areObjectsEqual(oldFields, newFields) {
		return nil
	}

	// Format header
	name := newObj.GetName()
	namespace := newObj.GetNamespace()
	if namespace != "" {
		name = namespace + "/" + name
	}

	resource := newObj.GetKind()
	if group := newObj.GetAPIVersion(); group != "" {
		resource = fmt.Sprintf("%s.%s", strings.ToLower(resource), group)
	}

	// Print header
	timestamp := ""
	if p.showTimestamp {
		now := time.Now()
		if now.Sub(p.lastPrintTime) > time.Second {
			timestamp = now.Format("15:04:05 ")
			p.lastPrintTime = now
		}
	}

	fmt.Printf("%s%s\n", timestamp, p.header(fmt.Sprintf("diff %s %s", resource, name)))
	fmt.Printf("%s\n", p.header(strings.Repeat("-", 80)))

	// For first time objects (no old version)
	if oldObj == nil {
		fmt.Printf("%s\n", p.added("+ New Resource"))
		// Print spec if exists
		if spec, ok := newObj.Object["spec"]; ok {
			p.printSection("spec", spec, true)
		}
		// Print status if exists
		if status, ok := newObj.Object["status"]; ok {
			p.printSection("status", status, true)
		}
		fmt.Println()
		return nil
	}

	// Print the diff
	p.printDiff("", oldFields, newFields, "")
	fmt.Println()
	return nil
}

func (p *DiffPrinter) printDiff(path string, old, new interface{}, indent string) {
	switch {
	case old == nil && new == nil:
		return
	case old == nil:
		p.printValue(path, new, indent, true)
	case new == nil:
		p.printValue(path, old, indent, false)
	default:
		switch oldVal := old.(type) {
		case map[string]interface{}:
			if newVal, ok := new.(map[string]interface{}); ok {
				p.printMapDiff(path, oldVal, newVal, indent)
			} else {
				p.printValue(path, old, indent, false)
				p.printValue(path, new, indent, true)
			}
		case []interface{}:
			if newVal, ok := new.([]interface{}); ok {
				p.printSliceDiff(path, oldVal, newVal, indent)
			} else {
				p.printValue(path, old, indent, false)
				p.printValue(path, new, indent, true)
			}
		default:
			if old != new {
				p.printValue(path, old, indent, false)
				p.printValue(path, new, indent, true)
			}
		}
	}
}

func (p *DiffPrinter) printMapDiff(path string, old, new map[string]interface{}, indent string) {
	// Get all keys
	keys := make(map[string]bool)
	for k := range old {
		keys[k] = true
	}
	for k := range new {
		keys[k] = true
	}

	// Sort keys for consistent output
	sortedKeys := make([]string, 0, len(keys))
	for k := range keys {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	// Print diff for each key
	for _, k := range sortedKeys {
		oldVal, oldOk := old[k]
		newVal, newOk := new[k]

		newPath := k
		if path != "" {
			newPath = path + "." + k
		}

		switch {
		case !oldOk:
			p.printDiff(newPath, nil, newVal, indent)
		case !newOk:
			p.printDiff(newPath, oldVal, nil, indent)
		default:
			p.printDiff(newPath, oldVal, newVal, indent)
		}
	}
}

func (p *DiffPrinter) printSliceDiff(path string, old, new []interface{}, indent string) {
	maxLen := len(old)
	if len(new) > maxLen {
		maxLen = len(new)
	}

	for i := 0; i < maxLen; i++ {
		var oldVal, newVal interface{}
		if i < len(old) {
			oldVal = old[i]
		}
		if i < len(new) {
			newVal = new[i]
		}

		newPath := fmt.Sprintf("%s[%d]", path, i)
		p.printDiff(newPath, oldVal, newVal, indent+"  ")
	}
}

func (p *DiffPrinter) printValue(path string, val interface{}, indent string, isAdd bool) {
	if val == nil {
		return
	}

	var prefix string
	var colorFunc func(a ...interface{}) string
	if isAdd {
		prefix = "+"
		colorFunc = p.added
	} else {
		prefix = "-"
		colorFunc = p.removed
	}

	switch v := val.(type) {
	case map[string]interface{}:
		if path == "" {
			// This is the root object, print its fields directly
			for k, fieldVal := range v {
				p.printDiff(k, nil, fieldVal, indent)
			}
		} else {
			b, _ := json.MarshalIndent(v, indent, "  ")
			lines := strings.Split(string(b), "\n")
			for _, line := range lines {
				if line == "{" || line == "}" {
					continue
				}
				fmt.Printf("%s%s%s\n", colorFunc(prefix), indent, line)
			}
		}
	case []interface{}:
		b, _ := json.MarshalIndent(v, indent, "  ")
		lines := strings.Split(string(b), "\n")
		for _, line := range lines {
			if line == "[" || line == "]" {
				continue
			}
			fmt.Printf("%s%s%s\n", colorFunc(prefix), indent, line)
		}
	default:
		if path != "" {
			fmt.Printf("%s%s%s: %v\n", colorFunc(prefix), indent, path, v)
		} else {
			fmt.Printf("%s%s%v\n", colorFunc(prefix), indent, v)
		}
	}
}

func (p *DiffPrinter) printSection(name string, val interface{}, isAdd bool) {
	prefix := "+"
	if !isAdd {
		prefix = "-"
	}
	colorFunc := p.added
	if !isAdd {
		colorFunc = p.removed
	}

	fmt.Printf("%s %s:\n", colorFunc(prefix), name)
	if m, ok := val.(map[string]interface{}); ok {
		b, _ := json.MarshalIndent(m, "  ", "  ")
		lines := strings.Split(string(b), "\n")
		for _, line := range lines {
			if line == "{" || line == "}" {
				continue
			}
			fmt.Printf("%s  %s\n", colorFunc(prefix), line)
		}
	}
}

// areObjectsEqual compares two objects recursively
func (p *DiffPrinter) areObjectsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v1 := range a {
		v2, ok := b[k]
		if !ok {
			return false
		}

		switch val1 := v1.(type) {
		case map[string]interface{}:
			if val2, ok := v2.(map[string]interface{}); !ok || !p.areObjectsEqual(val1, val2) {
				return false
			}
		case []interface{}:
			if val2, ok := v2.([]interface{}); !ok || !p.areSlicesEqual(val1, val2) {
				return false
			}
		default:
			if v1 != v2 {
				return false
			}
		}
	}

	return true
}

// areSlicesEqual compares two slices recursively
func (p *DiffPrinter) areSlicesEqual(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		switch v1 := a[i].(type) {
		case map[string]interface{}:
			if v2, ok := b[i].(map[string]interface{}); !ok || !p.areObjectsEqual(v1, v2) {
				return false
			}
		case []interface{}:
			if v2, ok := b[i].([]interface{}); !ok || !p.areSlicesEqual(v1, v2) {
				return false
			}
		default:
			if a[i] != b[i] {
				return false
			}
		}
	}

	return true
}
