package cmd

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/futuretea/kubectl-yadt/pkg/differ"
	"github.com/futuretea/kubectl-yadt/pkg/watcher"
	"github.com/manifoldco/promptui"
	"github.com/rancher/wrangler/pkg/clients"
	"github.com/rancher/wrangler/pkg/generic"
	"github.com/rancher/wrangler/pkg/kubeconfig"
	"github.com/rancher/wrangler/pkg/signals"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/klog/v2"
)

var (
	debug    bool
	noStatus bool
	noMeta   bool
)

var watchCmd = &cobra.Command{
	Use:   "watch [resources...]",
	Short: "Watch Kubernetes resources and show changes",
	Long: `Watch Kubernetes resources and print the delta in changes.
Example: kubectl-yadt watch pods deployments`,
	RunE: watchRun,
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug logging")
	watchCmd.Flags().BoolVar(&noStatus, "no-status", false, "Ignore status changes")
	watchCmd.Flags().BoolVar(&noMeta, "no-meta", false, "Ignore metadata changes")
}

func watchRun(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		resources, err := selectResource()
		if err != nil {
			return err
		}
		args = resources
	}

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// Disable klog output
	klog.SetOutput(io.Discard)

	ctx := signals.SetupSignalContext()
	restConfig := kubeconfig.GetNonInteractiveClientConfigWithContext(kubeConfig, context)

	logrus.WithFields(logrus.Fields{
		"context":   context,
		"namespace": namespace,
	}).Info("Creating kubernetes client")

	clients, err := clients.New(restConfig, &generic.FactoryOptions{
		Namespace: namespace,
	})
	if err != nil {
		logrus.WithError(err).Debug("Failed to create kubernetes client")
		return err
	}

	watcher, err := watcher.New(clients)
	if err != nil {
		logrus.WithError(err).Debug("Failed to create watcher")
		return err
	}

	for _, arg := range args {
		logrus.WithField("resource", arg).Debug("Adding resource to watch")
		watcher.MatchName(arg)
	}

	differ, err := differ.New(clients)
	if err != nil {
		logrus.WithError(err).Debug("Failed to create differ")
		return err
	}
	differ.SetIgnoreStatus(noStatus)
	differ.SetIgnoreMeta(noMeta)

	logrus.Debug("Starting to watch resources")
	objs, err := watcher.Start(ctx)
	if err != nil {
		logrus.WithError(err).Debug("Failed to start watcher")
		return err
	}

	go func() {
		for obj := range objs {
			if err := differ.Print(obj); err != nil {
				logrus.WithError(err).Debug("Failed to print diff")
			}
		}
	}()

	if err := clients.Start(ctx); err != nil {
		logrus.WithError(err).Debug("Failed to start clients")
		return err
	}

	<-ctx.Done()
	logrus.Debug("Shutting down")
	return nil
}

func selectResource() ([]string, error) {
	clientConfig := kubeconfig.GetNonInteractiveClientConfigWithContext(kubeConfig, context)
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	discovery, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	resources, err := discovery.ServerPreferredResources()
	if err != nil {
		return nil, err
	}

	var items []string
	items = append(items, "Done")
	for _, list := range resources {
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			continue
		}

		for _, r := range list.APIResources {
			if !strings.Contains(r.Verbs.String(), "watch") {
				continue
			}
			if strings.Contains(r.Name, "/") {
				continue
			}

			if gv.Group == "" {
				items = append(items, r.Name)
			} else {
				items = append(items, fmt.Sprintf("%s.%s", r.Name, gv.Group))
			}
		}
	}

	sort.Strings(items[1:])

	var selected []string
	for {
		prompt := promptui.Select{
			Label: fmt.Sprintf("Select Resources (%d selected)", len(selected)),
			Items: items,
			Searcher: func(input string, index int) bool {
				item := items[index]
				return strings.Contains(strings.ToLower(item), strings.ToLower(input))
			},
			Size: 20,
		}

		_, result, err := prompt.Run()
		if err != nil {
			return nil, err
		}

		if result == "Done" {
			break
		}

		found := false
		for _, s := range selected {
			if s == result {
				found = true
				break
			}
		}
		if !found {
			selected = append(selected, result)
		}
	}

	if len(selected) == 0 {
		return nil, fmt.Errorf("no resources selected")
	}

	return selected, nil
}
