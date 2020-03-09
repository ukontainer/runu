package main

import (
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
)

func setupNetwork(spec *specs.Spec) (*lklIfInfo, error) {
	logrus.Infof("no netns detected: no addr configuration, skipping")
	return nil, nil
}
