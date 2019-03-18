package main

import (
	"flag"
	"os"

	server "github.com/redhat-developer/git-service/pkg/cmd/server"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/component-base/logs"
)

var log = logs.NewLogger("git-api-server: ")

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	stopCh := genericapiserver.SetupSignalHandler()
	options := server.NewGitServerOptions(os.Stdout, os.Stderr)
	cmd := server.NewCommandStartGitServer(options, stopCh)
	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
