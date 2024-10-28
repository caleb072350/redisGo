package config

import (
	"bufio"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type PropertyHolder struct {
	Bind           string `cfg:"bind"`
	Port           int    `cfg:"port"`
	AppendOnly     bool   `cfg:"appendOnly"`
	AppendFilename string `cfg:"appendFilename"`
	MaxClients     int    `cfg:"maxClients"`
}

var Properties *PropertyHolder

func LoadConfig(configFilename string) *PropertyHolder {
	// open config file
	file, err := os.Open(configFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// load config
	rawMap := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		pivot := strings.Index(line, " ")
		if pivot > 0 && pivot < len(line)-1 {
			key := line[:pivot]
			value := line[pivot+1:]
			rawMap[strings.ToLower(key)] = value
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// parse config
	config := &PropertyHolder{}
	t := reflect.TypeOf(config)
	v := reflect.ValueOf(config)
	n := t.Elem().NumField()
	for i := 0; i < n; i++ {
		field := t.Elem().Field(i)
		fieldVal := v.Elem().Field(i)
		key, ok := field.Tag.Lookup("cfg")
		if !ok {
			key = field.Name
		}
		value, ok := rawMap[strings.ToLower(key)]
		if ok {
			switch field.Type.Kind() {
			case reflect.String:
				fieldVal.SetString(value)
			case reflect.Int:
				intValue, err := strconv.ParseInt(value, 10, 64)
				if err == nil {
					fieldVal.SetInt(intValue)
				}
			case reflect.Bool:
				boolVal := "yes" == value
				fieldVal.SetBool(boolVal)
			}
		}
	}
	return config
}

func SetupConfig(configFilename string) {
	Properties = LoadConfig(configFilename)
}
