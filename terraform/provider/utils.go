package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
)

func getUserHomeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, err
}

func ParseConfigfile(fileName string, config interface{}) error {
	if strings.Contains(fileName, "~") {
		usrHomeDir, err := getUserHomeDir()
		if err != nil {
			return err
		}
		fileName = strings.Replace(fileName, "~", usrHomeDir, -1)
	}

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return err
	}

	yamlFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		return err
	}
	return nil
}