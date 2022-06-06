package phphttpd

import (
	"os"

	"github.com/paketo-buildpacks/packit/v2"
)

func Detect() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {

		// only pass detection if $BP_PHP_SERVER is set to httpd
		if os.Getenv("BP_PHP_SERVER") != "httpd" {
			return packit.DetectResult{}, packit.Fail.WithMessage("BP_PHP_SERVER is not set to 'httpd'")
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{},
				Provides: []packit.BuildPlanProvision{
					{
						Name: PhpHttpdConfig,
					},
				},
			},
		}, nil
	}
}
