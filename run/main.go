package main

import (
	"os"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	phphttpd "github.com/paketo-buildpacks/php-httpd"
)

func main() {
	logEmitter := scribe.NewEmitter(os.Stdout).WithLevel(os.Getenv("BP_LOG_LEVEL"))
	config := phphttpd.NewConfig(logEmitter)
	entryResolver := draft.NewPlanner()

	packit.Run(
		phphttpd.Detect(),
		phphttpd.Build(entryResolver, config, chronos.DefaultClock, logEmitter),
	)
}
