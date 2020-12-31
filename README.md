campwiz
==========
Command-line interface that lists Bay Area campsites that are available on a particular date, using:

* Santa Clara County Parks
* San Mateo County Parks
* Reserve America
* Reserve California

![screenshot](campwiz.png)

Requirements:
=============
* go v1.14+
* macOS, Windows, or any UNIX flavor

CLI usage:
==========

To search campsites near San Francisco with a minimum rating for a particular set of dates:

```shell
 go run cmd/cw/cw.go --dates 2021-01-15,2021-01-29 --min_rating 7 \
   --nights 2 --max_distance 150
```

Webserver usage:
================

```shell
go run cmd/server/server.go
```

By default, campwiz listens on port 8080.

Roadmap:
========
- Flush cache on error
- SQL persistent cache interface
- Integrate additional metadata sources (Google Maps, Bing, Yelp)
- Filtering by campsite type