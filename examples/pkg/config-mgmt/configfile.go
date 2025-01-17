/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package configMgmt

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"

	toml "github.com/pelletier/go-toml"
	"google.golang.org/grpc/credentials"
)

type Tls struct {
	CaCert     string `toml:"ca_cert"`
	Cert       string `toml:"cert"`
	Key        string `toml:"key"`
	ClientCert string `toml:"client_cert"`
	ClientKey  string `toml:"client_key"`
}

type Endpoint struct {
	Port uint16 `toml:"port"`
}

type Grpc struct {
	Endpoint Endpoint `toml:"endpoint"`
}

type ConfigFileData struct {
	Tls  Tls  `toml:"tls"`
	Grpc Grpc `toml:"grpc"`
}

const (
	DefaultRootCaPath     = "/etc/bpfd/certs/ca/ca.pem"
	DefaultCertPath       = "/etc/bpfd/certs/bpfd/tls.crt"
	DefaultKeyPath        = "/etc/bpfd/certs/bpfd/tls.key"
	DefaultClientCertPath = "/etc/bpfd/certs/bpfd-client/bpfd-client.pem"
	DefaultClientKeyPath  = "/etc/bpfd/certs/bpfd-client/bpfd-client.key"
	DefaultPort           = 50051
)

func LoadConfig(configFilePath string) ConfigFileData {
	config := ConfigFileData{
		Tls: Tls{
			CaCert:     DefaultRootCaPath,
			Cert:       DefaultCertPath,
			Key:        DefaultKeyPath,
			ClientCert: DefaultClientCertPath,
			ClientKey:  DefaultClientKeyPath,
		},
		Grpc: Grpc{
			Endpoint: Endpoint{
				Port: DefaultPort,
			},
		},
	}

	log.Printf("Reading %s ...\n", configFilePath)
	file, err := os.ReadFile(configFilePath)
	if err == nil {
		err = toml.Unmarshal(file, &config)
		if err != nil {
			log.Printf("Unmarshal failed: err %+v\n", err)
		}
	} else {
		log.Printf("Read %s failed: err %+v\n", configFilePath, err)
	}

	return config
}

func LoadTLSCredentials(tlsFiles Tls) (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed server's certificate
	pemServerCA, err := os.ReadFile(tlsFiles.CaCert)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	// Load client's certificate and private key
	clientCert, err := tls.LoadX509KeyPair(tlsFiles.ClientCert, tlsFiles.ClientKey)
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	return credentials.NewTLS(config), nil
}
