This app runs `php-fpm` in the background, and then runs `httpd` in the
foreground.
The goal of this fixture is to show that the HTTPD process and configuration is
properly configured to work with FPM.
