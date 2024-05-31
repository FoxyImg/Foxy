package env

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"reflect"
	"strconv"
)

func ptr(v reflect.Value) reflect.Value {
	pt := reflect.PointerTo(v.Type())
	pv := reflect.New(pt.Elem())
	pv.Elem().Set(v)
	return pv
}

func LoadEnvironment(config interface{}) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	v := reflect.ValueOf(config).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)

		envKey := field.Tag.Get("env")
		if envKey != "" {
			switch field.Type.String() {
			case "int":
				envVal := os.Getenv(envKey)
				if envVal != "" {
					envIntVal, intErr := strconv.Atoi(envVal)
					if intErr != nil {
						log.Fatal("Error parsing int env var", intErr)
					}

					v.FieldByName(field.Name).SetInt(int64(envIntVal))
				}
			case "string":
				envVal := os.Getenv(envKey)
				if envVal == "" && v.Field(i).String() == "" {
					log.Fatal("Missing required env var: " + envKey)
				}

				if envVal != "" {
					v.FieldByName(field.Name).SetString(envVal)
				}
			case "*string":
				envVal := os.Getenv(envKey)
				if envVal != "" {
					v.FieldByName(field.Name).Set(ptr(reflect.ValueOf(envVal)))
				}
			case "bool":
				envVal := os.Getenv(envKey)
				if envVal == "" {
					log.Fatal("Missing required env var: " + envKey)
				} else {
					v.FieldByName(field.Name).SetBool(envVal == "true")
				}

			case "*bool":
				envVal := os.Getenv(envKey)
				if envVal != "" {
					boolVal := envVal == "true"
					v.FieldByName(field.Name).Set(ptr(reflect.ValueOf(boolVal)))
				}
			}
		}
	}
}
