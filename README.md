autocamper
==========
Tool that reserves campsites so you don't have to.

Building:
=========

```shell
export GOPATH=`pwd`
# Disregard the error message
go get github.com/tstromberg/autocamper
go get github.com/PuerkitoBio/goquery
go get github.com/steveyen/gkvlite
cd github.com/tstromberg/autocamper/cmd/campquery
go run campquery.go --help
```
