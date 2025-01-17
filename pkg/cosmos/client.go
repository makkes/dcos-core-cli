package cosmos

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"sort"

	"github.com/antihax/optional"
	"github.com/dcos/client-go/dcos"
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/httpclient"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
)

// Client is a diagnostics client for DC/OS.
type Client struct {
	cosmos *dcos.CosmosApiService
}

// NewClient creates a new Cosmos client.
func NewClient(ctx api.Context, baseClient *httpclient.Client) (*Client, error) {
	dcosConfigStore := dcos.NewConfigStore(&dcos.ConfigStoreOpts{
		Fs: ctx.Fs(),
	})
	dcosConfig := dcos.NewConfig(dcosConfigStore)
	dcosConfig.SetURL(baseClient.BaseURL().String())
	pluginutil.SetConfigFromEnv(dcosConfig)

	dcosClient, err := dcos.NewClientWithConfig(dcosConfig)
	if err != nil {
		return nil, err
	}
	return &Client{
		cosmos: dcosClient.Cosmos,
	}, nil
}

// PackageDescribe returns the content of '/package/describe'.
func (c *Client) PackageDescribe(name string, version string) (*Description, error) {
	desc, _, err := c.cosmos.PackageDescribe(context.TODO(), &dcos.PackageDescribeOpts{
		CosmosPackageDescribeV1Request: optional.NewInterface(dcos.CosmosPackageDescribeV1Request{
			PackageName:    name,
			PackageVersion: version,
		}),
	})
	if err != nil {
		return nil, err
	}

	backwardCompatibleDesc := Description{
		Package: desc.Package,
	}
	return &backwardCompatibleDesc, nil
}

// PackageListVersions returns the versions of a package.
func (c *Client) PackageListVersions(name string) ([]string, error) {
	list, _, err := c.cosmos.PackageListVersions(context.TODO(), dcos.CosmosPackageListVersionsV1Request{
		PackageName:            name,
		IncludePackageVersions: true,
	})

	if err != nil {
		return nil, err
	}

	var versions []string
	for version := range list.Results {
		versions = append(versions, version)
	}

	sort.Strings(versions)

	return versions, nil
}

// PackageRender returns a rendered package.
func (c *Client) PackageRender(appID string, name string, version string, optionsPath string) (map[string]interface{}, error) {
	var optionsInterface map[string]interface{}
	if optionsPath != "" {
		options, err := ioutil.ReadFile(optionsPath)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(options), &optionsInterface)
		if err != nil {
			log.Fatal(err)
		}
	}

	render, _, err := c.cosmos.PackageRender(context.TODO(), &dcos.PackageRenderOpts{
		CosmosPackageRenderV1Request: optional.NewInterface(dcos.CosmosPackageRenderV1Request{
			AppId:          appID,
			PackageName:    name,
			PackageVersion: version,
			Options:        optionsInterface,
		}),
	})

	if err != nil {
		return nil, err
	}

	return render.MarathonJson, nil
}

// PackageSearch returns the packages found using the given query.
func (c *Client) PackageSearch(query string) (*SearchResult, error) {
	desc, _, err := c.cosmos.PackageSearch(context.TODO(), dcos.CosmosPackageSearchV1Request{Query: query})
	if err != nil {
		return nil, err
	}

	backwardCompatibleSearchResult := SearchResult{
		Packages: desc.Packages,
	}
	return &backwardCompatibleSearchResult, nil
}

// PackageAddRepo adds a package repository.
func (c *Client) PackageAddRepo(name string, uri string, index int) ([]dcos.CosmosPackageRepo, error) {
	addRepoRequest := dcos.CosmosPackageAddRepoV1Request{
		Name: name,
		Uri:  uri,
	}

	if index > 0 {
		index32 := int32(index)
		addRepoRequest.Index = &index32
	}

	desc, _, err := c.cosmos.PackageRepositoryAdd(context.TODO(), &dcos.PackageRepositoryAddOpts{
		CosmosPackageAddRepoV1Request: optional.NewInterface(addRepoRequest),
	})
	if err != nil {
		return nil, err
	}
	return desc.Repositories, nil
}

// PackageDeleteRepo deletes a package repository.
func (c *Client) PackageDeleteRepo(name string) error {
	_, _, err := c.cosmos.PackageRepositoryDelete(context.TODO(), &dcos.PackageRepositoryDeleteOpts{
		CosmosPackageDeleteRepoV1Request: optional.NewInterface(dcos.CosmosPackageDeleteRepoV1Request{
			Name: name,
		}),
	})
	return err
}

// PackageListRepo returns a list of package repositories.
func (c *Client) PackageListRepo() ([]dcos.CosmosPackageRepo, error) {
	desc, _, err := c.cosmos.PackageRepositoryList(context.TODO(), nil)
	if err != nil {
		return nil, err
	}
	return desc.Repositories, nil
}
