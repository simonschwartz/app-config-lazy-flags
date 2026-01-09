package appconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	"github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	"github.com/aws/aws-sdk-go-v2/service/appconfigdata"
)

type ConfigClient interface {
	ListApplications(
		context.Context,
		*appconfig.ListApplicationsInput,
		...func(*appconfig.Options),
	) (*appconfig.ListApplicationsOutput, error)
	ListEnvironments(
		context.Context,
		*appconfig.ListEnvironmentsInput,
		...func(*appconfig.Options),
	) (*appconfig.ListEnvironmentsOutput, error)
	ListConfigurationProfiles(
		context.Context,
		*appconfig.ListConfigurationProfilesInput,
		...func(*appconfig.Options),
	) (*appconfig.ListConfigurationProfilesOutput, error)
}

type DataClient interface {
	StartConfigurationSession(
		context.Context,
		*appconfigdata.StartConfigurationSessionInput,
		...func(*appconfigdata.Options),
	) (*appconfigdata.StartConfigurationSessionOutput, error)
	GetLatestConfiguration(
		context.Context,
		*appconfigdata.GetLatestConfigurationInput,
		...func(*appconfigdata.Options),
	) (*appconfigdata.GetLatestConfigurationOutput, error)
}

type App struct {
	Description *string
	Id          *string
	Name        *string
}

type AppFlagConfig struct {
	ApplicationId *string
	Id            *string
	Name          *string
}

type AppEnvironments struct {
	ApplicationId *string
	Id            *string
	Name          *string
	State         types.EnvironmentState
}

type Flag struct {
	Enabled bool `json:"enabled"`
}

type Flags map[string]Flag

type Client struct {
	configClient ConfigClient
	dataClient   DataClient
}

func New(cfg aws.Config) *Client {
	configClient := appconfig.NewFromConfig(cfg)
	dataClient := appconfigdata.NewFromConfig(cfg)

	return &Client{
		configClient: configClient,
		dataClient:   dataClient,
	}
}

func (c *Client) ListApps(ctx context.Context) ([]App, error) {
	apps, err := c.configClient.ListApplications(ctx, &appconfig.ListApplicationsInput{MaxResults: aws.Int32(50)})
	if err != nil {
		return []App(nil), fmt.Errorf("failed to list applications: %w", err)
	}

	var applications []App
	for _, app := range apps.Items {
		applications = append(applications, App{
			Description: app.Description,
			Id:          app.Id,
			Name:        app.Name,
		})
	}

	return applications, nil
}

func (c *Client) ListAppFlagConfigs(ctx context.Context, appId string) ([]AppFlagConfig, error) {
	configs, err := c.configClient.ListConfigurationProfiles(ctx, &appconfig.ListConfigurationProfilesInput{ApplicationId: &appId})
	if err != nil {
		return []AppFlagConfig(nil), fmt.Errorf("failed to list application configs: %w", err)
	}

	var flagConfigs []AppFlagConfig
	for _, config := range configs.Items {
		if *config.Type == "AWS.AppConfig.FeatureFlags" {
			flagConfigs = append(flagConfigs, AppFlagConfig{
				ApplicationId: config.ApplicationId,
				Id:            config.Id,
				Name:          config.Name,
			})
		}
	}

	return flagConfigs, nil
}

func (c *Client) ListAppEnvironments(ctx context.Context, appId string) ([]AppEnvironments, error) {
	envs, err := c.configClient.ListEnvironments(ctx, &appconfig.ListEnvironmentsInput{ApplicationId: &appId})
	if err != nil {
		return []AppEnvironments(nil), fmt.Errorf("failed to list application environments: %w", err)
	}

	var appEnvs []AppEnvironments
	for _, env := range envs.Items {
		appEnvs = append(appEnvs, AppEnvironments{
			ApplicationId: env.ApplicationId,
			Id:            env.Id,
			Name:          env.Name,
			State:         env.State,
		})
	}

	return appEnvs, nil
}

// Careful - this request costs $$$
func (c *Client) GetLatestFlagConfig(ctx context.Context, appId, configId, envId string, minPollInterval int32) (Flags, error) {
	session, err := c.dataClient.StartConfigurationSession(ctx, &appconfigdata.StartConfigurationSessionInput{
		ApplicationIdentifier:                &appId,
		ConfigurationProfileIdentifier:       &configId,
		EnvironmentIdentifier:                &envId,
		RequiredMinimumPollIntervalInSeconds: &minPollInterval,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to establish configuration session: %w", err)
	}

	res, err := c.dataClient.GetLatestConfiguration(ctx, &appconfigdata.GetLatestConfigurationInput{ConfigurationToken: session.InitialConfigurationToken})
	if err != nil {
		return nil, fmt.Errorf("failed to get feature flags: %w", err)
	}

	var flags Flags

	err = json.Unmarshal(res.Configuration, &flags)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal feature flags: %w", err)
	}

	log.Printf("config: %v", flags)
	return flags, nil
}

type Result struct {
	EnvName  string
	EnvState types.EnvironmentState
	Flags    Flags
	Err      error
}

func (c *Client) GetFlags(ctx context.Context, appId string, configId string) ([]Result, error) {
	envs, err := c.ListAppEnvironments(ctx, appId)
	if err != nil {
		return nil, fmt.Errorf("failed to list app environments: %w", err)
	}
	results := make(chan Result, len(envs))

	var wg sync.WaitGroup
	for _, env := range envs {
		wg.Add(1)
		go func(env AppEnvironments) {
			defer wg.Done()
			flags, err := c.GetLatestFlagConfig(ctx, appId, configId, *env.Id, 60)
			results <- Result{
				EnvName:  *env.Name,
				EnvState: env.State,
				Flags:    flags,
				Err:      err,
			}
		}(env)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	res := make([]Result, 0, len(envs))
	for result := range results {
		res = append(res, result)
	}

	return res, nil
}
