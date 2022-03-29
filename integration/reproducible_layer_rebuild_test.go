package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testReproducibleLayerRebuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		docker occam.Docker
		pack   occam.Pack

		imageIDs     map[string]struct{}
		containerIDs map[string]struct{}

		name   string
		source string
	)

	it.Before(func() {
		var err error
		name, err = occam.RandomName()
		Expect(err).NotTo(HaveOccurred())

		docker = occam.NewDocker()
		pack = occam.NewPack()
		imageIDs = map[string]struct{}{}
		containerIDs = map[string]struct{}{}
	})

	it.After(func() {
		for id := range containerIDs {
			Expect(docker.Container.Remove.Execute(id)).To(Succeed())
		}

		for id := range imageIDs {
			Expect(docker.Image.Remove.Execute(id)).To(Succeed())
		}

		Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
		Expect(os.RemoveAll(source)).To(Succeed())
	})

	context("when an app is rebuilt", func() {
		it("creates a layer with the same SHA as before", func() {
			var (
				err  error
				logs fmt.Stringer

				firstImage  occam.Image
				secondImage occam.Image
			)

			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			build := pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					httpdBuildpack,
					phpBuildpack,
					phpFpmBuildpack,
					buildpack,
					buildPlanBuildpack,
					procfileBuildpack,
				).
				WithEnv(map[string]string{
					"BP_LOG_LEVEL":  "DEBUG",
					"BP_PHP_SERVER": "httpd",
				})

			firstImage, logs, err = build.Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[firstImage.ID] = struct{}{}

			Expect(firstImage.Buildpacks).To(HaveLen(6))

			Expect(firstImage.Buildpacks[3].Key).To(Equal(buildpackInfo.Buildpack.ID))
			Expect(firstImage.Buildpacks[3].Layers).To(HaveKey("php-httpd-config"))

			Expect(logs.String()).To(ContainSubstring("  Executing build process"))

			// Second pack build
			secondImage, _, err = build.Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[secondImage.ID] = struct{}{}

			Expect(secondImage.Buildpacks).To(HaveLen(6))

			Expect(secondImage.Buildpacks[3].Key).To(Equal(buildpackInfo.Buildpack.ID))
			Expect(secondImage.Buildpacks[3].Layers).To(HaveKey("php-httpd-config"))

			Expect(secondImage.Buildpacks[3].Layers["php-httpd-config"].SHA).To(Equal(firstImage.Buildpacks[3].Layers["php-httpd-config"].SHA))
		})
	})
}
