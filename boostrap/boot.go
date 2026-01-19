package boostrap

import (
	"os"

	"github.com/socle-lab/core"
)

func Boot(applicationName string) (*core.Core, error) {
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	// init socle
	socleApp := &core.Core{}
	err = socleApp.New(path, applicationName)
	if err != nil {
		return nil, err
	}

	return socleApp, nil

}
