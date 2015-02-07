timetrap
========

mkdir -p $GOSPACE/src/github.com/trebogeer/

cd $GOSPACE/src/github.com/trebogeer/

go get ./...

go build


agents
========
cd $GOSPACE/src/github.com/trebogeer/agents/mstats

go build mstat2.go

