campwiz
==========
Command-line interface that lists Bay Area campsites that are available on a particular date, using:

* Santa Clara County Parks
* San Mateo County Parks
* Reserve America
* Reserve California

Building:
=========

```shell
go get -u github.com/tstromberg/campwiz/pkg/cw
```

Usage:
======

```shell
cw --dates 2021-02-05 --nights 1
```