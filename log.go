package main

import (
	"github.com/op/go-logging"
	"io"
	"os"
)

func getLogger(pf io.Writer) (*logging.Logger, error) {

	log, err := logging.GetLogger("app")
	if err != nil {
		return nil, err
	}

	var stdFormat = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x} %{message}`,
	)
	var fileFormat = logging.MustStringFormatter(
		`%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x} %{message}`,
	)

	pfBackend := logging.NewLogBackend(pf, "", 0)
	backend := logging.NewLogBackend(os.Stderr, "", 0)

	fileFormatter := logging.NewBackendFormatter(pfBackend, fileFormat)
	backendFormatter := logging.NewBackendFormatter(backend, stdFormat)

	logging.SetBackend(fileFormatter, backendFormatter)

	return log, nil
}

func openLogFile(file string) (*os.File, error) {

	pf, err := os.OpenFile(file, os.O_WRONLY, 0666)

	if err != nil {

		if !os.IsNotExist(err) {
			return nil, err
		}

		pf, err = os.Create(file)
		if err != nil {
			return nil, err
		}
	}

	return pf, nil
}
