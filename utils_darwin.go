package main

import (
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
)

// /usr/lib is not writable on recent darwin
var runuAuxFileDir = "/usr/local/lib/runu"

func setupNetwork(spec *specs.Spec) (*lklIfInfo, error) {
	logrus.Infof("no netns detected: no addr configuration, skipping")
	return nil, nil
}
