// Licensed to the LF AI & Data foundation under one
// or more contributor license agreements. See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership. The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package nmq

import (
	"sync"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"go.uber.org/zap"

	"github.com/milvus-io/milvus/pkg/log"
	"github.com/milvus-io/milvus/pkg/util/paramtable"
)

// Nmq is global natsmq instance that will be initialized only once
var Nmq *server.Server

// once is used to init global natsmq
var once sync.Once

// NatsMQConfig is used to initialize NatsMQ.
type NatsMQConfig struct {
	Opts              server.Options
	InitializeTimeout time.Duration
}

// MustInitNatsMQ init global local natsmq instance.
// Panic if initailizing operation failed.
func MustInitNatsMQ(cfg *NatsMQConfig) {
	once.Do(func() {
		log.Info("try to initialize global nmq", zap.Any("config", cfg))
		var err error
		Nmq, err = server.NewServer(&cfg.Opts)
		if err != nil {
			log.Fatal("fail to initailize nmq", zap.Error(err))
		}

		// Start Nmq in background and wait until it's ready for connection.
		go Nmq.Start()
		// Wait for server to be ready for connections
		if !Nmq.ReadyForConnections(cfg.InitializeTimeout) {
			log.Fatal("nmq is not ready within timeout")
		}
		log.Info("initialize nmq finished", zap.String("client-url", Nmq.ClientURL()), zap.Error(err))
	})
}

// ParseServerOption get nats server option from paramstable.
func ParseServerOption(params *paramtable.ComponentParam) *NatsMQConfig {
	return &NatsMQConfig{
		Opts: server.Options{
			Host:              "127.0.0.1", // Force to use loopback address.
			Port:              params.NatsmqCfg.ServerPort.GetAsInt(),
			MaxPayload:        params.NatsmqCfg.ServerMaxPayload.GetAsInt32(),
			MaxPending:        params.NatsmqCfg.ServerMaxPending.GetAsInt64(),
			JetStream:         true,
			JetStreamMaxStore: params.NatsmqCfg.ServerMaxFileStore.GetAsInt64(),
			StoreDir:          params.NatsmqCfg.ServerStoreDir.GetValue(),
			Debug:             params.NatsmqCfg.ServerMonitorDebug.GetAsBool(),
			Logtime:           params.NatsmqCfg.ServerMonitorLogTime.GetAsBool(),
			LogFile:           params.NatsmqCfg.ServerMonitorLogFile.GetValue(),
			LogSizeLimit:      params.NatsmqCfg.ServerMonitorLogSizeLimit.GetAsInt64(),
		},
		InitializeTimeout: time.Duration(params.NatsmqCfg.ServerInitializeTimeout.GetAsInt()) * time.Millisecond,
	}
}

// CloseNatsMQ is used to close global natsmq
func CloseNatsMQ() {
	log.Debug("Closing Natsmq!")
	if Nmq != nil {
		// Shut down the server.
		Nmq.Shutdown()
		// Wait for server shutdown.
		Nmq.WaitForShutdown()
		Nmq = nil
	}
}
