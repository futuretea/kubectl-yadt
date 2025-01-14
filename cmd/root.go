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
	Use:   "wtfk8s",
	Short: "A tool for watching Kubernetes resources",
	Long: `wtfk8s is a command line tool that helps you monitor changes 
in your Kubernetes cluster resources.`,
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
