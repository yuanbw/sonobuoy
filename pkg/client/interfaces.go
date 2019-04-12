/*
Copyright 2018 Heptio Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"io"
	"time"

	"github.com/heptio/sonobuoy/pkg/config"
	"github.com/heptio/sonobuoy/pkg/plugin/aggregation"
	"github.com/heptio/sonobuoy/pkg/plugin/manifest"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// LogConfig are the input options for viewing a Sonobuoy run's logs.
type LogConfig struct {
	// Follow determines if the logs should be followed or not (tail -f).
	Follow bool
	// Namespace is the namespace the sonobuoy aggregator is running in.
	Namespace string
	// Out is the writer to write to.
	Out io.Writer
}

// GenConfig are the input options for generating a Sonobuoy manifest.
type GenConfig struct {
	E2EConfig            *E2EConfig
	Config               *config.Config
	Image                string
	Namespace            string
	EnableRBAC           bool
	ImagePullPolicy      string
	KubeConformanceImage string
	SSHKeyPath           string
	SSHUser              string

	// DynamicPlugins are plugins which we know by name and whose manifest
	// YAML are generated dynamically using the GenConfig settings.
	DynamicPlugins []string

	// StaticPlugins are plugins whose manifest YAML has been provided
	// explicitly and will be written without further consideration of other
	// GenConfig settings.
	StaticPlugins []*manifest.Manifest
}

// E2EConfig is the configuration of the E2E tests.
type E2EConfig struct {
	Focus    string
	Skip     string
	Parallel string

	// CustomRegistries is the contents of a yaml file which will be
	// used as KUBE_TEST_REPO_LIST which overrides which registries
	// e2e tests use.
	CustomRegistries string
}

// RunConfig are the input options for running Sonobuoy.
type RunConfig struct {
	GenConfig
	Wait time.Duration
}

// DeleteConfig are the input options for cleaning up a Sonobuoy run.
type DeleteConfig struct {
	Namespace  string
	EnableRBAC bool
	DeleteAll  bool
	Wait       time.Duration
}

// RetrieveConfig are the input options for retrieving a Sonobuoy run's results.
type RetrieveConfig struct {
	// Namespace is the namespace the sonobuoy aggregator is running in.
	Namespace string
}

// PreflightConfig are the options passed to PreflightChecks.
type PreflightConfig struct {
	Namespace string
}

// SonobuoyKubeAPIClient is the interface Sonobuoy uses to communicate with a kube-apiserver.
type SonobuoyKubeAPIClient interface {
	CreateObject(*unstructured.Unstructured) (*unstructured.Unstructured, error)
	Name(*unstructured.Unstructured) (string, error)
	Namespace(*unstructured.Unstructured) (string, error)
	ResourceVersion(*unstructured.Unstructured) (string, error)
}

// SonobuoyClient is a high-level interface to Sonobuoy operations.
type SonobuoyClient struct {
	RestConfig    *rest.Config
	client        kubernetes.Interface
	dynamicClient SonobuoyKubeAPIClient
}

// NewSonobuoyClient creates a new SonobuoyClient
func NewSonobuoyClient(restConfig *rest.Config, skc SonobuoyKubeAPIClient) (*SonobuoyClient, error) {
	sc := &SonobuoyClient{
		RestConfig:    restConfig,
		client:        nil,
		dynamicClient: skc,
	}
	return sc, nil
}

// Client creates or retrieves an existing kubernetes client from the SonobuoyClient's RESTConfig.
func (s *SonobuoyClient) Client() (kubernetes.Interface, error) {
	if s.client == nil {
		client, err := kubernetes.NewForConfig(s.RestConfig)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't create kubernetes client")
		}
		s.client = client
	}
	return s.client, nil
}

// Make sure SonobuoyClient implements the interface
var _ Interface = &SonobuoyClient{}

// Interface is the main contract that we will give to external consumers of this library
// This will provide a consistent look/feel to upstream and allow us to expose sonobuoy behavior
// to other automation systems.
type Interface interface {
	// Run generates the manifest, then tries to apply it to the cluster.
	// returns created resources or an error
	Run(cfg *RunConfig) error
	// GenerateManifest fills in a template with a Sonobuoy config
	GenerateManifest(cfg *GenConfig) ([]byte, error)
	// RetrieveResults copies results from a sonobuoy run into a Reader in tar format.
	RetrieveResults(cfg *RetrieveConfig) (io.Reader, <-chan error)
	// GetStatus determines the status of the sonobuoy run in order to assist the user.
	GetStatus(namespace string) (*aggregation.Status, error)
	// LogReader returns a reader that contains a merged stream of sonobuoy logs.
	LogReader(cfg *LogConfig) (*Reader, error)
	// Delete removes a sonobuoy run, namespace, and all associated resources.
	Delete(cfg *DeleteConfig) error
	// PreflightChecks runs a number of preflight checks to confirm the environment is good for Sonobuoy
	PreflightChecks(cfg *PreflightConfig) []error
}
