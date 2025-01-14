package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rancher/wrangler/pkg/kubeconfig"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
)

var resourcesCmd = &cobra.Command{
	Use:   "resources",
	Short: "List all available resources that can be watched",
	RunE:  listResources,
}

func init() {
	rootCmd.AddCommand(resourcesCmd)
}

func listResources(cmd *cobra.Command, args []string) error {
	clientConfig := kubeconfig.GetNonInteractiveClientConfigWithContext(kubeConfig, context)
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return err
	}

	discovery, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return err
	}

	// Get all API resources
	resources, err := discovery.ServerPreferredResources()
	if err != nil {
		return err
	}

	// Group resources by API group
	groups := make(map[string][]schema.GroupVersionResource)
	for _, list := range resources {
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			continue
		}

		for _, r := range list.APIResources {
			if !strings.Contains(r.Verbs.String(), "watch") {
				continue
			}

			// Skip subresources
			if strings.Contains(r.Name, "/") {
				continue
			}

			gvr := schema.GroupVersionResource{
				Group:    gv.Group,
				Version:  gv.Version,
				Resource: r.Name,
			}
			groups[gv.Group] = append(groups[gv.Group], gvr)
		}
	}

	// Sort groups
	groupNames := make([]string, 0, len(groups))
	for group := range groups {
		if group == "" {
			group = "core"
		}
		groupNames = append(groupNames, group)
	}
	sort.Strings(groupNames)

	// Print resources by group
	for _, group := range groupNames {
		displayGroup := group
		if displayGroup == "core" {
			displayGroup = ""
			fmt.Printf("\nCore Resources:\n")
		} else {
			fmt.Printf("\n%s Resources:\n", displayGroup)
		}

		resources := groups[group]
		sort.Slice(resources, func(i, j int) bool {
			return resources[i].Resource < resources[j].Resource
		})

		for _, r := range resources {
			if displayGroup == "" {
				fmt.Printf("  %s\n", r.Resource)
			} else {
				fmt.Printf("  %s.%s\n", r.Resource, displayGroup)
			}
		}
	}

	fmt.Println("\nUse 'wtfk8s watch <resource>' to watch a specific resource")
	return nil
}
