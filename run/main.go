package main

import (
	"os"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	phphttpd "github.com/paketo-buildpacks/php-httpd"
)

func main() {
	logEmitter := scribe.NewEmitter(os.Stdout).WithLevel(os.Getenv("BP_LOG_LEVEL"))
	config := phphttpd.NewConfig(logEmitter)

	packit.Run(
		phphttpd.Detect(),
		phphttpd.Build(config, logEmitter),
	)
}
