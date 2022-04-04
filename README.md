# PHP HTTPD Cloud Native Buildpack
A Cloud Native Buildpack for configuring HTTPD settings for PHP apps.

The buildpack generates the HTTPD configuration file
with the minimal set of options to get HTTPD to work, and incorporates
configuration from users and environment variables. The final HTTPD configuration file
is available at
`/layers/paketo-buildpacks_php-httpd/php-httpd-config/httpd.conf`, or
locatable through the buildpack-set `$PHP_HTTPD_PATH` environment variable at launch-time.

## Integration

The PHP HTTPD CNB provides `php-httpd-config`, which can be required by subsequent
buildpacks. In order to configure HTTPD, the user must declare the intention to
use HTTPD as the web-server by setting the `$BP_PHP_SERVER` environment
variable to `httpd` at build-time.

```shell
pack build my-httpd-app --env BP_PHP_SERVER="httpd"
```

## HTTPD Configuration Sources
The base configuration file generated in this buildpack includes some default
configuration, and an `IncludeOption` section for user-included configuration.

#### User Included Configuration
User-included configuration should be found in the application source directory
under `<app-directory>/.httpd.conf.d/*.conf`. If a file at this path exists, it
will be included in an `IncludeOptional` section at the bottom of the generated
HTTPD configuration.

#### Environment Variables
The following environment variables can be used to override default settings in
the HTTPD configuration file.

| Variable | Default |
| -------- | -------- |
| `BP_PHP_SERVER_ADMIN`     | admin@localhost    |
| `BP_PHP_ENABLE_HTTPS_REDIRECT`   | true    |
| `BP_PHP_WEB_DIR`    | htdocs    |

## Usage

To package this buildpack for consumption:

```
$ ./scripts/package.sh
```

This builds the buildpack's Go source using `GOOS=linux` by default. You can
supply another value as the first argument to `package.sh`.

## Run Tests

To run all unit tests, run:
```
./scripts/unit.sh
```

To run all integration tests, run:
```
./scripts/integration.sh
```

## Debug Logs
For extra debug logs from the image build process, set the `$BP_LOG_LEVEL`
environment variable to `DEBUG` at build-time (ex. `pack build my-app --env
BP_LOG_LEVEL=DEBUG` or through a  [`project.toml`
file](https://github.com/buildpacks/spec/blob/main/extensions/project-descriptor.md).
