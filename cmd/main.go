package main

import (
	"flag"
	"fmt"
	"os"

	server "github.com/redhat-developer/git-service/pkg/cmd/server"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/component-base/logs"
	//"github.com/golang/glog"
	//"k8s.io/apiserver/pkg/util/logs"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	stopCh := genericapiserver.SetupSignalHandler()
	options := server.NewPocServerOptions(os.Stdout, os.Stderr)
	cmd := server.NewCommandStartPocServer(options, stopCh)
	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		//glog.Fatal(err)
	}
}
