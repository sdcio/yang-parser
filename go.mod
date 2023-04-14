module github.com/steiler/yang-parser

go 1.20

require (
	github.com/danos/encoding v0.0.0-20210701125528-66857fd8c8ea
	github.com/danos/mgmterror v0.0.0-20210701125710-6fcf751e367d
	github.com/danos/utils v0.0.0-20210701125856-7935e3348d7c
	github.com/sirupsen/logrus v1.9.0
	github.com/iptecharch/schema-server v0.0.0-20230403115201-914d55f66653
)

replace github.com/iptecharch/schema-server v0.0.0-20230403115201-914d55f66653 => /home/mava/projects/schema-server

require (
	github.com/kr/pretty v0.3.1 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
)
