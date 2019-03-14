package serverstrings

import (
	"fmt"
	"io"
	"net"

	"github.com/spf13/cobra"

	"github.com/sbryzak/kubepoc/pkg/apiserver"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
)

const defaultEtcdPathPrefix = "/registry/kubepoc.bryzak.com"

type PocServerOptions struct {
	RecommendedOptions *genericoptions.RecommendedOptions
	StdOut             io.Writer
	StdErr             io.Writer
}

func NewPocServerOptions(out, errOut io.Writer) *PocServerOptions {
	o := &PocServerOptions{
		RecommendedOptions: genericoptions.NewRecommendedOptions(defaultEtcdPathPrefix,
			apiserver.Codecs.LegacyCodec(apiserver.SchemeGroupVersion), nil),
		StdOut: out,
		StdErr: errOut,
	}
	return o
}

// NewCommandStartPocServer provides a CLI handler for 'start master' command
// with a default PocServerOptions.
func NewCommandStartPocServer(defaults *PocServerOptions, stopCh <-chan struct{}) *cobra.Command {
	o := *defaults
	cmd := &cobra.Command{
		Short: "Launch Poc API server",
		Long:  "Launch Poc API server",
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(); err != nil {
				return err
			}
			if err := o.Validate(args); err != nil {
				return err
			}
			if err := o.RunPocServer(stopCh); err != nil {
				return err
			}
			return nil
		},
	}

	flags := cmd.Flags()
	o.RecommendedOptions.AddFlags(flags)

	return cmd
}

func (o PocServerOptions) Validate(args []string) error {
	errors := []error{}
	errors = append(errors, o.RecommendedOptions.Validate()...)
	return utilerrors.NewAggregate(errors)
}

func (o *PocServerOptions) Complete() error {
	return nil
}

func (o *PocServerOptions) Config() (*apiserver.Config, error) {
	// TODO have a "real" external address
	if err := o.RecommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	serverConfig := genericapiserver.NewRecommendedConfig(apiserver.Codecs)
	if err := o.RecommendedOptions.ApplyTo(serverConfig); err != nil {
		return nil, err
	}

	config := &apiserver.Config{
		GenericConfig: serverConfig,
		ExtraConfig:   apiserver.ExtraConfig{},
	}
	return config, nil
}

func (o PocServerOptions) RunPocServer(stopCh <-chan struct{}) error {
	config, err := o.Config()
	if err != nil {
		return err
	}

	server, err := config.Complete().New()
	if err != nil {
		return err
	}
	return server.GenericAPIServer.PrepareRun().Run(stopCh)
}
