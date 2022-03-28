package phphttpd_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
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

		buffer        *bytes.Buffer
		config        *fakes.ConfigWriter
		entryResolver *fakes.EntryResolver

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

		clock := chronos.DefaultClock

		buffer = bytes.NewBuffer(nil)
		logEmitter := scribe.NewEmitter(buffer)

		entryResolver = &fakes.EntryResolver{}

		config = &fakes.ConfigWriter{}
		config.WriteCall.Returns.String = "some-workspace/httpd.conf"

		build = phphttpd.Build(entryResolver, config, clock, logEmitter)
	})

	it.After(func() {
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	it("writes a config file into its layer", func() {
		result, err := build(packit.BuildContext{
			WorkingDir: workingDir,
			CNBPath:    cnbDir,
			Stack:      "some-stack",
			BuildpackInfo: packit.BuildpackInfo{
				Name:    "Some Buildpack",
				Version: "some-version",
			},
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{Name: "some-entry"},
				},
			},
			Layers: packit.Layers{Path: layerDir},
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(config.WriteCall.Receives.LayerPath).To(Equal(filepath.Join(layerDir, "php-httpd-config")))
		Expect(config.WriteCall.Receives.WorkingDir).To(Equal(workingDir))
		Expect(config.WriteCall.Receives.CnbPath).To(Equal(cnbDir))

		Expect(entryResolver.MergeLayerTypesCall.Receives.Name).To(Equal(phphttpd.PhpHttpdConfig))
		Expect(entryResolver.MergeLayerTypesCall.Receives.Entries).To(Equal([]packit.BuildpackPlanEntry{
			{Name: "some-entry"},
		}))

		Expect(result.Layers).To(HaveLen(1))
		Expect(result.Layers[0].Name).To(Equal("php-httpd-config"))
		Expect(result.Layers[0].Path).To(Equal(filepath.Join(layerDir, "php-httpd-config")))
		Expect(result.Layers[0].SharedEnv).To(Equal(packit.Environment{
			"PHP_HTTPD_PATH.default": "some-workspace/httpd.conf",
		}))

		Expect(result.Layers[0].Build).To(BeFalse())
		Expect(result.Layers[0].Cache).To(BeFalse())
		Expect(result.Layers[0].Launch).To(BeFalse())
	})

	context("when httpd-config is required at launch time", func() {
		it.Before(func() {
			entryResolver.MergeLayerTypesCall.Returns.Launch = true
		})
		it("makes the layer available at launch time", func() {
			result, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{Name: "some-entry"},
					},
				},
				Layers: packit.Layers{Path: layerDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(HaveLen(1))
			Expect(result.Layers[0].Name).To(Equal("php-httpd-config"))
			Expect(result.Layers[0].Path).To(Equal(filepath.Join(layerDir, "php-httpd-config")))
			Expect(result.Layers[0].SharedEnv).To(Equal(packit.Environment{
				"PHP_HTTPD_PATH.default": "some-workspace/httpd.conf",
			}))

			Expect(result.Layers[0].Launch).To(BeTrue())
			Expect(result.Layers[0].Build).To(BeFalse())
			Expect(result.Layers[0].Cache).To(BeFalse())
		})
	})

	context("failure cases", func() {
		context("when config layer cannot be gotten", func() {
			it.Before(func() {
				err := os.WriteFile(filepath.Join(layerDir, fmt.Sprintf("%s.toml", phphttpd.PhpHttpdConfigLayer)), nil, 0000)
				Expect(err).NotTo(HaveOccurred())
			})
			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					CNBPath:    cnbDir,
					Stack:      "some-stack",
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "some-version",
					},
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{Name: phphttpd.PhpHttpdConfigLayer},
						},
					},
					Layers: packit.Layers{Path: layerDir},
				})
				Expect(err).To(MatchError(ContainSubstring("failed to parse layer content metadata")))
			})
		})

		context("when config file cannot be written", func() {
			it.Before(func() {
				config.WriteCall.Returns.Error = errors.New("config writing error")
			})
			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					CNBPath:    cnbDir,
					Stack:      "some-stack",
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "some-version",
					},
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{Name: phphttpd.PhpHttpdConfigLayer},
						},
					},
					Layers: packit.Layers{Path: layerDir},
				})
				Expect(err).To(MatchError(ContainSubstring("config writing error")))
			})
		})
	})
}
