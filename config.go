package phphttpd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/paketo-buildpacks/packit/v2/scribe"
)

type HttpdConfig struct {
	ServerAdmin          string
	DisableHTTPSRedirect bool
	AppRoot              string
	WebDirectory         string
	FpmSocket            string
	UserInclude          string
}

type Config struct {
	logger scribe.Emitter
}

func NewConfig(logger scribe.Emitter) Config {
	return Config{
		logger: logger,
	}
}

func (c Config) Write(layerPath, workingDir, cnbPath string) (string, error) {
	tmpl, err := template.New("httpd.conf").ParseFiles(filepath.Join(cnbPath, "config", "httpd.conf"))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTTPD config template: %w", err)
	}

	// Configuration set by this buildpack

	// If there's a user-provided HTTPD conf, include it in the base configuration.
	userPath := filepath.Join(workingDir, ".httpd.conf.d", "*.conf")
	_, err = os.Stat(filepath.Join(workingDir, ".httpd.conf.d"))
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			// untested
			return "", fmt.Errorf("failed to stat %s/.httpd.conf.d: %w", workingDir, err)
		}
		userPath = ""
	}
	if userPath != "" {
		c.logger.Debug.Subprocess(fmt.Sprintf("Including user-provided HTTPD configuration from: %s", userPath))
	}

	serverAdmin := os.Getenv("BP_PHP_SERVER_ADMIN")
	if serverAdmin == "" {
		serverAdmin = "admin@localhost"
	}
	c.logger.Debug.Subprocess(fmt.Sprintf("Server admin: %s", serverAdmin))

	webDir := os.Getenv("BP_PHP_WEB_DIR")
	if webDir == "" {
		webDir = "htdocs"
	}
	c.logger.Debug.Subprocess(fmt.Sprintf("Web directory: %s", webDir))

	enableHTTPSRedirect := true
	enableHTTPSRedirectStr, ok := os.LookupEnv("BP_PHP_ENABLE_HTTPS_REDIRECT")
	if ok {
		enableHTTPSRedirect, err = strconv.ParseBool(enableHTTPSRedirectStr)
		if err != nil {
			return "", fmt.Errorf("failed to pase $BP_PHP_ENABLE_HTTPS_REDIRECT into boolean: %w", err)
		}
	}
	c.logger.Debug.Subprocess(fmt.Sprintf("Enable HTTPS redirect: %t", enableHTTPSRedirect))
	fpmSocket := "127.0.0.1:9000"

	data := HttpdConfig{
		ServerAdmin:          serverAdmin,
		AppRoot:              workingDir,
		WebDirectory:         webDir,
		FpmSocket:            fpmSocket,
		DisableHTTPSRedirect: !enableHTTPSRedirect,
		UserInclude:          userPath,
	}

	var b bytes.Buffer
	err = tmpl.Execute(&b, data)
	if err != nil {
		// not tested
		return "", err
	}

	f, err := os.OpenFile(filepath.Join(layerPath, "httpd.conf"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, &b)
	if err != nil {
		// not tested
		return "", err
	}

	return f.Name(), nil
}
