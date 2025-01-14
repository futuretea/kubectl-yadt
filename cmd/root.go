package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	kubeConfig string
	context    string
	namespace  string
)

var rootCmd = &cobra.Command{
	Use:   "yadt",
	Short: "Watch and diff Kubernetes resources in real-time",
	Long: `A tool for watching Kubernetes resources and showing changes in real-time.
This tool can be used both as a standalone command and as a kubectl plugin.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set default level to ERROR
		logrus.SetLevel(logrus.ErrorLevel)
		// Use JSON formatter
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&kubeConfig, "kubeconfig", "", "Kube config location")
	rootCmd.PersistentFlags().StringVar(&context, "context", "", "Kube config context")
	rootCmd.PersistentFlags().StringVar(&namespace, "namespace", "", "Limit to namespace")
}
