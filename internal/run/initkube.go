package run

import (
	"errors"
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/pelotech/drone-helm3/internal/env"
)

// InitKube is a step in a helm Plan that initializes the kubernetes config file.
type InitKube struct {
	*config
	templateFilename string
	configFilename   string
	template         *template.Template
	configFile       io.WriteCloser
	values           kubeValues
}

type kubeValues struct {
	SkipTLSVerify  bool
	Certificate    string
	APIServer      string
	Namespace      string
	ServiceAccount string
	Token          string
	KubeConfig     string
	KubeContext    string
}

// NewInitKube creates a InitKube using the given Config and filepaths. No validation is performed at this time.
func NewInitKube(cfg env.Config, templateFile, configFile string) *InitKube {
	return &InitKube{
		config: newConfig(cfg),
		values: kubeValues{
			SkipTLSVerify:  cfg.SkipTLSVerify,
			Certificate:    cfg.Certificate,
			APIServer:      cfg.APIServer,
			Namespace:      cfg.Namespace,
			ServiceAccount: cfg.ServiceAccount,
			Token:          cfg.KubeToken,
			KubeConfig:     cfg.KubeConfig,
			KubeContext:    cfg.KubeContext,
		},
		templateFilename: templateFile,
		configFilename:   configFile,
	}
}

// Execute generates a kubernetes config file from drone-helm3's template.
func (i *InitKube) Execute() error {
	if i.debug {
		fmt.Fprintf(i.stderr, "writing kubeconfig file to %s\n", i.configFilename)
	}

	if i.values.KubeConfig != "" {
		if i.debug {
			fmt.Fprintf(i.stderr, "Writing values of Kubeconfig to file")
			//fmt.Fprintf(i.stderr, "KubeConfig file \n %s \n", i.values.KubeConfig)
		}
		os.WriteFile(i.configFilename, []byte(i.values.KubeConfig), 0644)
		return nil
	}
	defer i.configFile.Close()
	return i.template.Execute(i.configFile, i.values)
}

// Prepare ensures all required configuration is present and that the config file is writable.
func (i *InitKube) Prepare() error {
	var err error

	if i.values.KubeConfig == "" {
		if i.values.APIServer == "" {
			return errors.New("an API Server is needed to deploy")
		}
		if i.values.Token == "" {
			return errors.New("token is needed to deploy")
		}

		if i.values.ServiceAccount == "" {
			i.values.ServiceAccount = "helm"
		}

		if i.debug {
			fmt.Fprintf(i.stderr, "loading kubeconfig template from %s\n", i.templateFilename)
		}
		i.template, err = template.ParseFiles(i.templateFilename)
		if err != nil {
			return fmt.Errorf("could not load kubeconfig template: %w", err)
		}

		if i.debug {
			if _, err := os.Stat(i.configFilename); err != nil {
				// non-nil err here isn't an actual error state; the kubeconfig just doesn't exist
				fmt.Fprint(i.stderr, "creating ")
			} else {
				fmt.Fprint(i.stderr, "truncating ")
			}
			fmt.Fprintf(i.stderr, "kubeconfig file at %s\n", i.configFilename)
		}

		i.configFile, err = os.Create(i.configFilename)
		if err != nil {
			return fmt.Errorf("could not open kubeconfig file for writing: %w", err)
		}
	} else {
		fmt.Fprintf(i.stderr, "Kubeconfig is present and then will be written in %s\n", i.configFilename)
	}
	return nil
}
