package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testOffline(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker
		source string
		name   string
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("when the buildpack is run with pack build", func() {
		var (
			image     occam.Image
			container occam.Container
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("generates a functional httpd config file", func() {
			var (
				logs fmt.Stringer
				err  error
			)

			image, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					offlineHttpdBuildpack,
					offlineBuildpack,
					buildPlanBuildpack,
					procfileBuildpack,
				).
				WithEnv(map[string]string{
					"BP_LOG_LEVEL":  "DEBUG",
					"BP_PHP_SERVER": "httpd",
				}).
				WithNetwork("none").
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, buildpackInfo.Buildpack.Name)),
				"  Getting the layer associated with the HTTPD configuration",
				"    /layers/paketo-buildpacks_php-httpd/php-httpd-config",
				"",
				"  Setting up the HTTPD configuration file",
				"    Including user-provided HTTPD configuration from: /workspace/.httpd.conf.d/*.conf",
				"    Server admin: admin@localhost",
				"    Web directory: htdocs",
				"    Enable HTTPS redirect: true",
				"",
				"  Configuring launch environment",
				MatchRegexp(fmt.Sprintf(`    PHP_HTTPD_PATH -> "/layers/%s/php-httpd-config/httpd.conf"`, strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"))),
			))

			container, err = docker.Container.Run.
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				WithPublishAll().
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(container).Should(Serve(ContainSubstring("Hello World!")).OnPort(8080))
		})
	})
}
