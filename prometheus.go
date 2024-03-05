package main

import (
	"github.com/prometheus/prometheus/prompb"
	"github.com/sirupsen/logrus"
)

func processWriteRequest(req *prompb.WriteRequest) (map[string][][]byte, error) {
	logrus.WithField("var", req).Debugln()
	return Serialize(serializer, req)
}
