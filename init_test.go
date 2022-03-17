package phphttpd_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitPhpHttpd(t *testing.T) {
	suite := spec.New("php-httpd", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Build", testBuild)
	suite("Detect", testDetect, spec.Sequential())
	suite("Config", testConfig, spec.Sequential())
	suite.Run(t)
}
