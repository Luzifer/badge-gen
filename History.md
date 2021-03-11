# 1.16.0 / 2021-03-11

  * Add Twitch service
  * Add documentation for configuration
  * Improve config loading
  * Allow service handler to disable itself

# 1.15.1 / 2021-03-11

  * [#1] Fix: Allow URI encoded slash in badge fields

# 1.15.0 / 2021-03-07

  * Update Dockerfile for go modules
  * Port to go modules, remove prometheus support
  * Remove old vendoring

# 1.14.1 / 2018-10-09

  * Update Dockerfile

# 1.14.0 / 2018-10-09

  * Deps: Update vendors
  * Add AUR service

# 1.13.1 / 2018-06-01

  * Fix: Background too bright for text

# 1.13.0 / 2018-06-01

  * Add LiberaPay support
  * Add Github stars by repo
  * Replace ctxhttp with native ctx support
  * Vendor: Switch to using `dep` for vendoring
  * Meta: Replace license stub with full text

# 1.12.0 / 2018-03-21

  * Gratipay service was discontinued

# 1.11.1 / 2016-07-07

  * fix missing godeps

# 1.11.0 / 2016-07-07

  * add metrics collection

# 1.10.0 / 2016-07-07

  * add Gratipay integration

# 1.9.0 / 2016-07-07

  * display example path on hover
  * add GitHub downloads integration
  * stop rendering badges for HTTP304 responses
  * fix loading list instead of latest release

# 1.8.0 / 2016-07-06

  * add GitHub latest-release integration
  * add GitHub latest-tag integration

# 1.7.0 / 2016-07-06

  * add caching to beerpay API
  * add service "beerpay"
  * Ignore character case when sorting
  * add color-naming map

# 1.6.0 / 2016-07-05

  * Add config storage
  * Add caching results of API calls
  * Added GitHub service with license command
  * Allow multiple documentation entries per handler

# 1.5.1 / 2016-06-29

  * Fix missing vendor libraries

# 1.5.0 / 2016-06-29

  * Deprecate /v1/badge route
  * Add context to timeout badge generators

# 1.4.3 / 2016-06-29

  * Shrink the version number

# 1.4.2 / 2016-06-29

  * Show program version in footer

# 1.4.1 / 2016-06-29

  * Fix title on demo page
  * Move documentation to the demo page

# 1.4.0 / 2016-06-28

  * Add demo page with / route

# 1.3.0 / 2016-06-28

  * Document new service plugin
  * Prevent long-time caching for GitHub
  * Add "travis" badge service

# 1.2.0 / 2016-06-28

  * Update Dockerfile to build binary in container
  * Add static badge plugin
  * Refactor app, add support for service hooks

# 1.1.0 / 2015-05-25

  * Added Godeps
  * Added svg minifying

# 1.0.0 / 2015-05-25

  * Initial version
