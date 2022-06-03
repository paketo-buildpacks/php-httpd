package phphttpd

import (
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

//go:generate faux --interface ConfigWriter --output fakes/config_writer.go

// ConfigWriter sets up the HTTPD configuration file with defaults, and adds in
// user-set environment variables.
type ConfigWriter interface {
	Write(layerPath, workingDir, cnbPath string) (string, error)
}

// Build will return a packit.BuildFunc that will be invoked during the build
// phase of the buildpack lifecycle.
//
// Build will create a layer dedicated to PHP HTTPD configuration, configure default HTTPD
// settings, incorporate other configuration sources, and make the
// configuration available at both build-time and
// launch-time.
func Build(config ConfigWriter, logger scribe.Emitter) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		logger.Debug.Process("Getting the layer associated with the HTTPD configuration")
		phpHttpdLayer, err := context.Layers.Get(PhpHttpdConfigLayer)
		if err != nil {
			return packit.BuildResult{}, err
		}
		logger.Debug.Subprocess(phpHttpdLayer.Path)
		logger.Debug.Break()

		phpHttpdLayer, err = phpHttpdLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Process("Setting up the HTTPD configuration file")
		httpdConfigPath, err := config.Write(phpHttpdLayer.Path, context.WorkingDir, context.CNBPath)
		if err != nil {
			return packit.BuildResult{}, err
		}
		logger.Break()

		planner := draft.NewPlanner()
		phpHttpdLayer.Launch, phpHttpdLayer.Build = planner.MergeLayerTypes(PhpHttpdConfig, context.Plan.Entries)

		phpHttpdLayer.SharedEnv.Default("PHP_HTTPD_PATH", httpdConfigPath)
		logger.EnvironmentVariables(phpHttpdLayer)

		return packit.BuildResult{
			Layers: []packit.Layer{phpHttpdLayer},
		}, nil
	}
}
