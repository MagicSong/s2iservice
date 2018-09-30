package run

import (
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/s2iservice/pkg/api"
	"github.com/s2iservice/pkg/api/describe"
	"github.com/s2iservice/pkg/api/validation"
	"github.com/s2iservice/pkg/build/strategies"
	"github.com/s2iservice/pkg/docker"
	util "github.com/s2iservice/pkg/utils"
)

//RunS2I Just run the command
func RunS2I(cfg *api.Config) {
	if len(cfg.AsDockerfile) > 0 {
		if cfg.RunImage {
			fmt.Fprintln(os.Stderr, "ERROR: --run cannot be used with --as-dockerfile")
			return
		}
		if len(cfg.RuntimeImage) > 0 {
			fmt.Fprintln(os.Stderr, "ERROR: --runtime-image cannot be used with --as-dockerfile")
			return
		}
	}

	if cfg.Incremental && len(cfg.RuntimeImage) > 0 {
		fmt.Fprintln(os.Stderr, "ERROR: Incremental build with runtime image isn't supported")
		return
	}
	//set default image pull policy
	if len(cfg.BuilderPullPolicy) == 0 {
		cfg.BuilderPullPolicy = api.DefaultBuilderPullPolicy
	}
	if len(cfg.PreviousImagePullPolicy) == 0 {
		cfg.PreviousImagePullPolicy = api.DefaultPreviousImagePullPolicy
	}
	if len(cfg.RuntimeImagePullPolicy) == 0 {
		cfg.RuntimeImagePullPolicy = api.DefaultRuntimeImagePullPolicy
	}

	if errs := validation.ValidateConfig(cfg); len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", e)
		}
		fmt.Println()
		cmd.Help()
		return
	}

	// Attempt to read the .dockercfg and extract the authentication for
	// docker pull
	if r, err := os.Open(cfg.DockerCfgPath); err == nil {
		defer r.Close()
		auths := docker.LoadImageRegistryAuth(r)
		cfg.PullAuthentication = docker.GetImageRegistryAuth(auths, cfg.BuilderImage)
		if cfg.Incremental {
			cfg.IncrementalAuthentication = docker.GetImageRegistryAuth(auths, cfg.Tag)
		}
		if len(cfg.RuntimeImage) > 0 {
			cfg.RuntimeAuthentication = docker.GetImageRegistryAuth(auths, cfg.RuntimeImage)
		}
	}

	if len(cfg.EnvironmentFile) > 0 {
		result, err := util.ReadEnvironmentFile(cfg.EnvironmentFile)
		if err != nil {
			glog.Warningf("Unable to read environment file %q: %v", cfg.EnvironmentFile, err)
		} else {
			for name, value := range result {
				cfg.Environment = append(cfg.Environment, api.EnvironmentSpec{Name: name, Value: value})
			}
		}
	}

	if len(oldScriptsFlag) != 0 {
		glog.Warning("DEPRECATED: Flag --scripts is deprecated, use --scripts-url instead")
		cfg.ScriptsURL = oldScriptsFlag
	}
	if len(oldDestination) != 0 {
		glog.Warning("DEPRECATED: Flag --location is deprecated, use --destination instead")
		cfg.Destination = oldDestination
	}

	if networkMode != "" {
		cfg.DockerNetworkMode = api.DockerNetworkMode(networkMode)
	}

	client, err := docker.NewEngineAPIClient(cfg.DockerConfig)
	if err != nil {
		glog.Fatal(err)
	}

	d := docker.New(client, cfg.PullAuthentication)
	err = d.CheckReachable()
	if err != nil {
		glog.Fatal(err)
	}

	glog.V(2).Infof("\n%s\n", describe.Config(client, cfg))

	builder, _, err := strategies.GetStrategy(client, cfg)
	s2ierr.CheckError(err)
	result, err := builder.Build(cfg)
	if err != nil {
		glog.V(0).Infof("Build failed")
		s2ierr.CheckError(err)
	} else {
		if len(cfg.AsDockerfile) > 0 {
			glog.V(0).Infof("Application dockerfile generated in %s", cfg.AsDockerfile)
		} else {
			glog.V(0).Infof("Build completed successfully")

		}
	}

	for _, message := range result.Messages {
		glog.V(1).Infof(message)
	}
}
