package phphttpd_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	phphttpd "github.com/paketo-buildpacks/php-httpd"
	"github.com/paketo-buildpacks/php-httpd/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layerDir   string
		workingDir string
		cnbDir     string

		buffer *bytes.Buffer
		config *fakes.ConfigWriter

		buildContext     packit.BuildContext
		expectedPhpLayer packit.Layer

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		layerDir, err = os.MkdirTemp("", "layer")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		buffer = bytes.NewBuffer(nil)
		logEmitter := scribe.NewEmitter(buffer)

		config = &fakes.ConfigWriter{}
		config.WriteCall.Returns.String = "some-workspace/httpd.conf"

		buildContext = packit.BuildContext{
			WorkingDir: workingDir,
			CNBPath:    cnbDir,
			Stack:      "some-stack",
			BuildpackInfo: packit.BuildpackInfo{
				Name:    "Some Buildpack",
				Version: "some-version",
			},
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: phphttpd.PhpHttpdConfig,
					},
				},
			},
			Layers: packit.Layers{Path: layerDir},
		}

		expectedPhpLayer = packit.Layer{
			Path:   filepath.Join(layerDir, phphttpd.PhpHttpdConfigLayer),
			Name:   phphttpd.PhpHttpdConfigLayer,
			Build:  false,
			Launch: false,
			Cache:  false,
			SharedEnv: packit.Environment{
				"PHP_HTTPD_PATH.default": "some-workspace/httpd.conf",
			},
			BuildEnv:         packit.Environment{},
			LaunchEnv:        packit.Environment{},
			ProcessLaunchEnv: map[string]packit.Environment{},
		}

		build = phphttpd.Build(config, logEmitter)
	})

	it.After(func() {
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	it("writes a config file into its layer", func() {
		result, err := build(buildContext)
		Expect(err).NotTo(HaveOccurred())

		Expect(config.WriteCall.Receives.LayerPath).To(Equal(filepath.Join(layerDir, "php-httpd-config")))
		Expect(config.WriteCall.Receives.WorkingDir).To(Equal(workingDir))
		Expect(config.WriteCall.Receives.CnbPath).To(Equal(cnbDir))

		Expect(result.Layers).To(HaveLen(1))
		Expect(result.Layers[0]).To(Equal(expectedPhpLayer))
	})

	context("when httpd-config is required at launch time", func() {
		it.Before(func() {
			buildContext.Plan.Entries[0].Metadata = map[string]interface{}{
				"launch": true,
			}

			expectedPhpLayer.Launch = true
		})

		it("makes the layer available at launch time", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(HaveLen(1))
			Expect(result.Layers[0]).To(Equal(expectedPhpLayer))
		})
	})

	context("failure cases", func() {
		context("when config layer cannot be gotten", func() {
			it.Before(func() {
				err := os.WriteFile(filepath.Join(layerDir, fmt.Sprintf("%s.toml", phphttpd.PhpHttpdConfigLayer)), nil, 0000)
				Expect(err).NotTo(HaveOccurred())
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("failed to parse layer content metadata")))
			})
		})

		context("when config file cannot be written", func() {
			it.Before(func() {
				config.WriteCall.Returns.Error = errors.New("config writing error")
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("config writing error")))
			})
		})
	})
}
