/*
 * Copyright 2018 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	// tl "github.com/tricksterproxy/trickster/pkg/logging"
)


type envToConfig struct {
	key		string
	value	string
}
// var tomlTagsToConfigMap map[string]interface{}

// func init() {
// 	tomlTagsToConfigMap = make(map[string]interface{})
// 	c := Config{}
// 	refType := reflect.TypeOf(c)
// 	for i:=0; i < refType.NumField(); i++ {
// 		// refType.Field(i).Tag.Get("toml")
// 		if tag, ok := refType.Field(i).Tag.Lookup("toml"); ok {
// 			tomlTagsToConfigMap[tag] = reflect.New(refType.Field(i).Type.Elem())
// 		}
// 	}
// }

const (
	// Environment variables
	evOriginURL   = "TRK_ORIGIN_URL"
	evProvider    = "TRK_ORIGIN_TYPE"
	evProxyPort   = "TRK_PROXY_PORT"
	evMetricsPort = "TRK_METRICS_PORT"
	evLogLevel    = "TRK_LOG_LEVEL"
)

func (c *Config) loadEnvVars() {
	// Origin
	if x := os.Getenv(evOriginURL); x != "" {
		c.providedOriginURL = x
	}

	if x := os.Getenv(evProvider); x != "" {
		c.providedProvider = x
	}

	// Proxy Port
	if x := os.Getenv(evProxyPort); x != "" {
		if y, err := strconv.ParseInt(x, 10, 64); err == nil {
			c.Frontend.ListenPort = int(y)
		}
	}

	// Metrics Port
	if x := os.Getenv(evMetricsPort); x != "" {
		if y, err := strconv.ParseInt(x, 10, 64); err == nil {
			c.Metrics.ListenPort = int(y)
		}
	}

	// LogLevel
	if x := os.Getenv(evLogLevel); x != "" {
		c.Logging.LogLevel = x
	}
	c.loadEnvVarsNew()

}

func (c *Config)loadEnvVarsNew() {
	configMap := make(map[string]map[string]string)
	for _, e := range os.Environ() {
		option := strings.SplitN(e, "=", 2) // split: TRX_main_instance_id="1234" into TRX_main_instance_id and "1234"
		n := strings.SplitN(option[0], "_", 3) // split: TRX_main_instance_id into TRX, main & instance_id
		if n[0] == "TRX" {
      		if _, ok := configMap[n[1]]; !ok {
        		configMap[n[1]] = make(map[string]string)
      		}
      		configMap[n[1]][n[2]] = option[1]
		}
	}
	//TODO: Not cool! try with pointers instead of this
	cfg := *c
	fmt.Printf("\n Config BEFORE:  \n %+v \n", cfg.Backends)
	refValC := reflect.ValueOf(cfg)
	for i:=0; i < reflect.TypeOf(cfg).NumField(); i++ {
		if tagName, ok := reflect.TypeOf(cfg).Field(i).Tag.Lookup("toml"); ok {
		  if _, ok:=configMap[tagName]; ok {
			//   fmt.Println(tagName)
			if refValC.Field(i).Kind() == reflect.Map {
				iter := refValC.Field(i).MapRange()
				for iter.Next() {
				  // x := reflect.TypeOf(c).Field(i).Type
				  x := refValC.Field(i).MapIndex(iter.Key()).Type()
				  update(configMap[tagName], x, refValC.Field(i).MapIndex(iter.Key()).Elem())
				}
				continue
			}
			update(configMap[tagName], reflect.TypeOf(cfg).Field(i).Type, refValC.Field(i).Elem())
		  }
		}
	}
	fmt.Printf("\n Config AFTER:  \n %+v \n", cfg.Backends)
	c = &cfg
	fmt.Println("Updated config")
}

func update(nameMap map[string]string, f reflect.Type, v reflect.Value) {
	for i:=0; i < f.Elem().NumField(); i++ {
	  if tag, ok := f.Elem().Field(i).Tag.Lookup("toml"); ok {
		if newValue, ok := nameMap[tag]; ok {
		  // updateViaReflect(f.Elem().Field(i).Type, v.Field(i), newValue)
		  switch f.Elem().Field(i).Type.Kind() {
			case reflect.Int, reflect.Int64:
	  			n, _ := strconv.ParseInt(newValue, 0, 64)
	  			v.Field(i).SetInt(n)
			case reflect.String:
				v.Field(i).SetString(newValue)
			case reflect.Bool:
				n, _ := strconv.ParseBool(newValue)
				v.Field(i).SetBool(n)
			default:
	  			fmt.Println("Unhandled type")
		  }
		}
	  }
	}
}