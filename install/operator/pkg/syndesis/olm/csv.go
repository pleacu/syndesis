/*
 * Copyright (C) 2020 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package olm

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/syndesisio/syndesis/install/operator/pkg/syndesis/capabilities"
	"github.com/syndesisio/syndesis/install/operator/pkg/syndesis/configuration"

	"github.com/syndesisio/syndesis/install/operator/pkg/generator"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"gopkg.in/yaml.v2"
)

type csv struct {
	config   *configuration.Config
	operator string
	body     []byte

	// Variable needed to build the CSV
	version  string
	maturity string

	// Dependant on whether it is community or productized
	name           string
	displayName    string
	support        string
	description    string
	maintainerName string
	maintainerMail string
	provider       string
}

type CSVOut struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string
	Metadata   Metadata
	Spec       Spec
}

type Spec struct {
	DisplayName               string `yaml:"displayName"`
	Description               string
	Keywords                  []string
	Version                   string
	Maturity                  string
	Maintainers               []Maintainer
	Provider                  Provider
	Labels                    Labels
	Selector                  Selector
	Icon                      []Icon
	Links                     []Link
	InstallModes              []InstallMode `yaml:"installModes"`
	Install                   Install
	Customresourcedefinitions CustomResourceDefinitions
	RelatedImages             []Image `yaml:"relatedImages"`
}

type Metadata struct {
	Name        string
	Namespace   string
	Annotations MetadataAnnotations
}

type MetadataAnnotations struct {
	Capabilities   string
	Categories     string
	Certified      string
	CreatedAt      string `yaml:"createdAt"`
	ContainerImage string `yaml:"containerImage"`
	Support        string
	Description    string
	Repository     string
	AlmExamples    string `yaml:"alm-examples"`
}

type Maintainer struct {
	Name  string
	Email string
}

type Provider struct {
	Name string
}

type Labels struct {
	Name string
}

type Selector struct {
	MatchLabels Label
}

type Link struct {
	Name string
	URL  string `json:"url"`
}

type Image struct {
	Name  string
	Image string
}

type InstallMode struct {
	Type      string
	Supported bool
}
type Label struct {
	Name string
}

type Icon struct {
	Base64data string
	Mediatype  string
}

type Install struct {
	Strategy string
	Spec     InstallSpec
}

type InstallSpec struct {
	ClusterPermissions []InstallSpecPermission
	Permissions        []InstallSpecPermission
	Deployments        []InstallSpecDeployment
}

type InstallSpecPermission struct {
	ServiceAccountName string
	Rules              interface{}
}

type InstallSpecDeployment struct {
	Name string
	Spec interface{}
}

type CustomResourceDefinitions struct {
	Owned []CustomResourceDefinition
}

type CustomResourceDefinition struct {
	Name        string
	Version     string
	Kind        string
	DisplayName string
	Description string
}

// In order to build the body for both upstream and downstream,
// set variables accordingly
func (c *csv) setVariables() {
	c.version = c.config.Version
	c.maturity = "alpha"

	// Both of these are required to ensure permissions are correctly added to manifest
	c.config.ApiServer.ImageStreams = true
	c.config.ApiServer.Routes = true
	c.config.ApiServer.EmbeddedProvider = true
	c.config.ApiServer.OlmSupport = true
	c.config.ApiServer.ConsoleLink = true

	// Dependant on whether it is community or productized
	c.name = "fuse-online-operator"
	c.displayName = "Red Hat Integration - Fuse Online"
	c.support = "Fuse Online"
	c.description = "Manages the installation of Fuse Online, a flexible and customizable open source platform that provides core integration capabilities as a service."
	c.maintainerName = "Jon Anstey"
	c.maintainerMail = "janstey@redhat.com"
	c.provider = "Red Hat"

	if !c.config.Productized {
		c.name = "syndesis-operator"
		c.displayName = "Syndesis"
		c.support = "Syndesis"
		c.description = "Manages the installation of Syndesis, a flexible and customizable open source platform that provides core integration capabilities as a service."
		c.maintainerName = "Syndesis team"
		c.maintainerMail = "syndesis@googlegroups.com"
		c.provider = "Syndesis team"
	}
}

// Build the content of the csv file
func (c *csv) build() (err error) {
	target := "productized"
	if !c.config.Productized {
		target = "community"
	}
	c.setVariables()

	alm, err := ioutil.ReadFile(filepath.Join("pkg", "syndesis", "olm", "assets", "alm-examples"))
	descriptionLong, _ := ioutil.ReadFile(filepath.Join("pkg", "syndesis", "olm", "assets", target, "description"))
	icon, _ := ioutil.ReadFile(filepath.Join("pkg", "syndesis", "olm", "assets", "icon"))
	rules, err := c.loadRoleFromTemplate()
	if err != nil {
		return err
	}

	clusterrules, err := c.loadClusterRoleFromTemplate()
	if err != nil {
		return err
	}

	deployment, err := c.loadDeploymentFromTemplate()
	if err != nil {
		return err
	}

	relatedImages, err := c.assembleRelatedImages()
	if err != nil {
		return err
	}

	co := CSVOut{
		APIVersion: "operators.coreos.com/v1alpha1",
		Kind:       "ClusterServiceVersion",
		Metadata: Metadata{
			Name:      c.name + ".v" + c.version,
			Namespace: "placeholder",
			Annotations: MetadataAnnotations{
				Capabilities:   "Seamless Upgrades",
				Categories:     "Integration & Delivery",
				Certified:      "false",
				CreatedAt:      time.Now().String(),
				ContainerImage: c.operator,
				Support:        c.support,
				Description:    c.description,
				Repository:     "https://github.com/syndesisio/syndesis/",
				AlmExamples:    string(alm),
			},
		},
		Spec: Spec{
			DisplayName: c.displayName,
			Description: string(descriptionLong),
			Keywords:    []string{"camel", "integration", "syndesis", "fuse", "online"},
			Version:     c.version,
			Maturity:    c.maturity,
			Maintainers: []Maintainer{
				{
					Name:  c.maintainerName,
					Email: c.maintainerMail,
				},
			},
			Provider: Provider{Name: c.provider},
			Labels:   Labels{Name: c.name},
			Selector: Selector{MatchLabels: Label{Name: c.name}},
			Icon: []Icon{
				{
					Base64data: string(icon),
					Mediatype:  "image/svg+xml",
				},
			},
			Links: []Link{
				{
					Name: "Red Hat Fuse Online Documentation",
					URL:  "https://access.redhat.com/documentation/en-us/red-hat-fuse",
				}, {
					Name: "Upstream project Syndesis",
					URL:  "https://github.com/syndesisio/syndesis",
				}, {
					Name: "Upstream Syndesis Operator",
					URL:  "https://github.com/syndesisio/syndesis/tree/master/install/operator",
				},
			},
			InstallModes: []InstallMode{
				{
					Type:      "OwnNamespace",
					Supported: true,
				}, {
					Type:      "SingleNamespace",
					Supported: true,
				}, {
					Type:      "MultiNamespace",
					Supported: false,
				}, {
					Type:      "AllNamespaces",
					Supported: false,
				},
			},
			Install: Install{
				Strategy: "deployment",
				Spec: InstallSpec{
					ClusterPermissions: []InstallSpecPermission{{
						ServiceAccountName: "syndesis-operator",
						Rules:              clusterrules,
					}},
					Permissions: []InstallSpecPermission{{
						ServiceAccountName: "syndesis-operator",
						Rules:              rules,
					}},
					Deployments: []InstallSpecDeployment{{
						Name: c.name,
						Spec: deployment,
					}},
				},
			},
			Customresourcedefinitions: CustomResourceDefinitions{
				Owned: []CustomResourceDefinition{{
					Name:        "syndesises.syndesis.io",
					Version:     "v1beta2",
					Kind:        "Syndesis",
					DisplayName: "Syndesis CRD",
					Description: "Syndesis CRD",
				}},
			},
			RelatedImages: relatedImages,
		},
	}

	c.body, err = yaml.Marshal(co)
	return
}

// Load role to apply to syndesis operator, from template file. This role
// is later applied to the syndesis-operator service account, by OLM
func (c *csv) loadRoleFromTemplate() (r interface{}, err error) {
	context := struct {
		Kind string
		Role string
	}{
		Kind: "",
		Role: "",
	}

	g, err := generator.Render("./install/role.yml.tmpl", context)
	if err != nil {
		return nil, err
	}

	mjson, err := g[0].MarshalJSON()
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	if err := yaml.Unmarshal(mjson, &m); err != nil {
		return nil, err
	}

	r = m["rules"]
	return
}

// Load cluster role to apply to syndesis operator, from template file. This role
// is later applied to the syndesis-operator service account, by OLM and allows
// the operator to query for OLM artifacts across namespaces, eg. subscriptions.
func (c *csv) loadClusterRoleFromTemplate() (r interface{}, err error) {
	context := struct {
		Kind      string
		Role      string
		ApiServer capabilities.ApiServerSpec
	}{
		Kind:      "",
		Role:      "",
		ApiServer: c.config.ApiServer,
	}

	resources, err := generator.Render("./install/cluster_role_kafka.yml.tmpl", context)
	if err != nil {
		return nil, err
	}

	olm, err := generator.Render("./install/cluster_role_olm.yml.tmpl", context)
	if err != nil {
		return nil, err
	}

	resources = append(resources, olm...)

	pubapi, err := generator.Render("./install/cluster_role_public_api.yml.tmpl", context)
	if err != nil {
		return nil, err
	}
	resources = append(resources, pubapi...)

	if len(resources) == 0 {
		return r, nil
	}

	m := make([]map[string]interface{}, 0, 0)
	for _, resource := range resources {
		rules, exists, _ := unstructured.NestedFieldNoCopy(resource.UnstructuredContent(), "rules")
		if !exists {
			return nil, fmt.Errorf("Cannot validate 'rules' in %s", resource.GetName())
		}

		ruleMaps, ok := rules.([]interface{})
		if !ok || len(ruleMaps) == 0 {
			return nil, fmt.Errorf("Cannot validate rule maps in %s", resource.GetName())
		}

		for _, ruleMap := range ruleMaps {
			ruleMap, ok := ruleMap.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("Cannot validate 'rule map' in %s", resource.GetName())
			}
			m = append(m, ruleMap)
		}
	}

	r = m
	return
}

// Load syndesis-operator deployment from template file
func (c *csv) loadDeploymentFromTemplate() (r interface{}, err error) {
	context := struct {
		RelatedImages   bool
		DatabaseImage   string
		OperatorImage   string
		AmqImage        string
		CamelKImage     string
		TodoImage       string
		OauthImage      string
		UiImage         string
		S2iImage        string
		PrometheusImage string
		UpgradeImage    string
		MetaImage       string
		ServerImage     string
		ExporterImage   string
	}{
		RelatedImages:   true,
		OperatorImage:   c.operator,
		DatabaseImage:   c.config.Syndesis.Components.Database.Image,
		CamelKImage:     c.config.Syndesis.Addons.CamelK.Image,
		TodoImage:       c.config.Syndesis.Addons.Todo.Image,
		AmqImage:        c.config.Syndesis.Components.AMQ.Image,
		OauthImage:      c.config.Syndesis.Components.Oauth.Image,
		UiImage:         c.config.Syndesis.Components.UI.Image,
		MetaImage:       c.config.Syndesis.Components.Meta.Image,
		ServerImage:     c.config.Syndesis.Components.Server.Image,
		S2iImage:        c.config.Syndesis.Components.S2I.Image,
		PrometheusImage: c.config.Syndesis.Components.Prometheus.Image,
		UpgradeImage:    c.config.Syndesis.Components.Upgrade.Image,
		ExporterImage:   c.config.Syndesis.Components.Database.Exporter.Image,
	}

	g, err := generator.Render("./install/deployment.yml.tmpl", context)
	if err != nil {
		return nil, err
	}

	mjson, err := g[0].MarshalJSON()
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	if err := yaml.Unmarshal(mjson, &m); err != nil {
		return nil, err
	}

	r = m["spec"]
	return
}

func (c *csv) assembleRelatedImages() ([]Image, error) {
	images := []Image{}

	images = append(images, Image{Name: "syndesis-operator", Image: c.operator})
	images = append(images, Image{Name: "postgres-version", Image: c.config.Syndesis.Components.Database.Image})
	images = append(images, Image{Name: "todo", Image: c.config.Syndesis.Addons.Todo.Image})
	images = append(images, Image{Name: "oauth", Image: c.config.Syndesis.Components.Oauth.Image})
	images = append(images, Image{Name: "ui", Image: c.config.Syndesis.Components.UI.Image})
	images = append(images, Image{Name: "s2i", Image: c.config.Syndesis.Components.S2I.Image})
	images = append(images, Image{Name: "prometheus", Image: c.config.Syndesis.Components.Prometheus.Image})
	images = append(images, Image{Name: "upgrade", Image: c.config.Syndesis.Components.Upgrade.Image})
	images = append(images, Image{Name: "meta", Image: c.config.Syndesis.Components.Meta.Image})
	images = append(images, Image{Name: "database", Image: c.config.Syndesis.Components.Database.Image})
	images = append(images, Image{Name: "psql_exporter", Image: c.config.Syndesis.Components.Database.Exporter.Image})
	images = append(images, Image{Name: "server", Image: c.config.Syndesis.Components.Server.Image})
	images = append(images, Image{Name: "amq", Image: c.config.Syndesis.Components.AMQ.Image})
	return images, nil
}
