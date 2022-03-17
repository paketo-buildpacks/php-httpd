package phphttpd_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/scribe"
	phphttpd "github.com/paketo-buildpacks/php-httpd"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testConfig(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layerDir   string
		workingDir string
		cnbDir     string
		config     phphttpd.Config
	)

	it.Before(func() {
		var err error
		layerDir, err = os.MkdirTemp("", "php-httpd-layer")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chmod(layerDir, os.ModePerm)).To(Succeed())

		workingDir, err = os.MkdirTemp("", "workingDir")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.MkdirAll(filepath.Join(cnbDir, "config"), os.ModePerm)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(cnbDir, "config", "httpd.conf"), []byte(`
ServerAdmin {{.ServerAdmin}}
DocumentRoot {{.AppRoot}}/{{.WebDirectory}}
FPMSocket {{.FpmSocket}}
DisableHTTPSRedirect {{.DisableHTTPSRedirect }}
{{ if ne .UserInclude "" }}
IncludeOptional {{ .UserInclude }}
{{- end}}
`), os.ModePerm)).To(Succeed())

		logEmitter := scribe.NewEmitter(bytes.NewBuffer(nil))
		config = phphttpd.NewConfig(logEmitter)
	})

	it.After(func() {
		Expect(os.RemoveAll(layerDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	it("writes an httpd.conf file into the layer dir", func() {
		path, err := config.Write(layerDir, workingDir, cnbDir)
		Expect(err).NotTo(HaveOccurred())

		Expect(path).To(Equal(filepath.Join(layerDir, "httpd.conf")))
		Expect(filepath.Join(layerDir, "httpd.conf")).To(BeARegularFile())

		contents, err := os.ReadFile(filepath.Join(layerDir, "httpd.conf"))
		Expect(err).NotTo(HaveOccurred())

		Expect(string(contents)).To(ContainSubstring("ServerAdmin admin@localhost"))
		Expect(string(contents)).To(ContainSubstring(fmt.Sprintf("DocumentRoot %s/htdocs", workingDir)))
		Expect(string(contents)).To(ContainSubstring("FPMSocket 127.0.0.1:9000"))
		Expect(string(contents)).To(ContainSubstring("DisableHTTPSRedirect false"))
		Expect(string(contents)).NotTo(ContainSubstring(fmt.Sprintf("IncludeOptional %s/.httpd.conf.d/*.conf", workingDir)))
	})

	context("there is a user-provided conf file", func() {
		it.Before(func() {
			Expect(os.MkdirAll(filepath.Join(workingDir, ".httpd.conf.d"), os.ModePerm)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(workingDir, ".httpd.conf.d", "user.conf"), nil, os.ModePerm)).To(Succeed())
		})
		it.After(func() {
			Expect(os.RemoveAll(filepath.Join(workingDir, ".httpd.conf.d"))).To(Succeed())
		})
		it("writes an httpd.conf with the user included conf into layerDir", func() {
			path, err := config.Write(layerDir, workingDir, cnbDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(path).To(Equal(filepath.Join(layerDir, "httpd.conf")))
			Expect(filepath.Join(layerDir, "httpd.conf")).To(BeARegularFile())

			contents, err := os.ReadFile(filepath.Join(layerDir, "httpd.conf"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(ContainSubstring(fmt.Sprintf("IncludeOptional %s/.httpd.conf.d/*.conf", workingDir)))
		})
	})

	context("all config env. vars are set", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_PHP_SERVER_ADMIN", "some-server-admin")).To(Succeed())
			Expect(os.Setenv("BP_PHP_ENABLE_HTTPS_REDIRECT", "false")).To(Succeed())
			Expect(os.Setenv("BP_PHP_WEB_DIR", "some-web-dir")).To(Succeed())
		})
		it.After(func() {
			Expect(os.Unsetenv("BP_PHP_SERVER_ADMIN")).To(Succeed())
			Expect(os.Unsetenv("BP_PHP_ENABLE_HTTPS_REDIRECT")).To(Succeed())
			Expect(os.Unsetenv("BP_PHP_WEB_DIR")).To(Succeed())
		})
		it("writes an httpd.conf that includes the env var values", func() {
			path, err := config.Write(layerDir, workingDir, cnbDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(path).To(Equal(filepath.Join(layerDir, "httpd.conf")))

			contents, err := os.ReadFile(filepath.Join(layerDir, "httpd.conf"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(ContainSubstring("ServerAdmin some-server-admin"))
			Expect(string(contents)).To(ContainSubstring(fmt.Sprintf("DocumentRoot %s/some-web-dir", workingDir)))
			Expect(string(contents)).To(ContainSubstring("DisableHTTPSRedirect true"))
		})
	})

	context("failure cases", func() {
		context("when template is not parseable", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(cnbDir, "config", "httpd.conf"), []byte(`
{{ .UserInclude
		`), os.ModePerm)).To(Succeed())
			})
			it("returns an error", func() {
				_, err := config.Write(layerDir, workingDir, cnbDir)
				Expect(err).To(MatchError(ContainSubstring("unclosed action")))
			})
		})

		context("when the BP_PHP_ENABLE_HTTPS_REDIRECT value cannot be parsed into a bool", func() {
			it.Before(func() {
				Expect(os.Setenv("BP_PHP_ENABLE_HTTPS_REDIRECT", "blah")).To(Succeed())
			})
			it.After(func() {
				Expect(os.Unsetenv("BP_PHP_ENABLE_HTTPS_REDIRECT")).To(Succeed())
			})
			it("returns an error", func() {
				_, err := config.Write(layerDir, workingDir, cnbDir)
				Expect(err).To(MatchError(ContainSubstring("failed to pase $BP_PHP_ENABLE_HTTPS_REDIRECT into boolean:")))
			})
		})

		context("when conf file can't be opened for writing", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(layerDir, "httpd.conf"), nil, 0400)).To(Succeed())
			})
			it("returns an error", func() {
				_, err := config.Write(layerDir, workingDir, cnbDir)
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})
	})
}
