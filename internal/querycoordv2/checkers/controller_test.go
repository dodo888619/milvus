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

package checkers

import (
	"context"
	"testing"
	"time"

	"github.com/milvus-io/milvus/internal/kv"
	etcdkv "github.com/milvus-io/milvus/internal/kv/etcd"
	"github.com/milvus-io/milvus/internal/metastore/kv/querycoord"
	"github.com/milvus-io/milvus/internal/proto/datapb"
	"github.com/milvus-io/milvus/internal/querycoordv2/balance"
	"github.com/milvus-io/milvus/internal/querycoordv2/meta"
	. "github.com/milvus-io/milvus/internal/querycoordv2/params"
	"github.com/milvus-io/milvus/internal/querycoordv2/session"
	"github.com/milvus-io/milvus/internal/querycoordv2/task"
	"github.com/milvus-io/milvus/internal/querycoordv2/utils"
	"github.com/milvus-io/milvus/pkg/util/etcd"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/atomic"
)

type CheckerControllerSuite struct {
	suite.Suite
	kv            kv.MetaKv
	meta          *meta.Meta
	broker        *meta.MockBroker
	nodeMgr       *session.NodeManager
	dist          *meta.DistributionManager
	targetManager *meta.TargetManager
	scheduler     *task.MockScheduler
	balancer      *balance.MockBalancer

	controller *CheckerController
}

func (suite *CheckerControllerSuite) SetupSuite() {
	Params.Init()
}

func (suite *CheckerControllerSuite) SetupTest() {
	var err error
	config := GenerateEtcdConfig()
	cli, err := etcd.GetEtcdClient(
		config.UseEmbedEtcd.GetAsBool(),
		config.EtcdUseSSL.GetAsBool(),
		config.Endpoints.GetAsStrings(),
		config.EtcdTLSCert.GetValue(),
		config.EtcdTLSKey.GetValue(),
		config.EtcdTLSCACert.GetValue(),
		config.EtcdTLSMinVersion.GetValue())
	suite.Require().NoError(err)
	suite.kv = etcdkv.NewEtcdKV(cli, config.MetaRootPath.GetValue())

	// meta
	store := querycoord.NewCatalog(suite.kv)
	idAllocator := RandomIncrementIDAllocator()
	suite.nodeMgr = session.NewNodeManager()
	suite.meta = meta.NewMeta(idAllocator, store, suite.nodeMgr)
	suite.dist = meta.NewDistributionManager()
	suite.broker = meta.NewMockBroker(suite.T())
	suite.targetManager = meta.NewTargetManager(suite.broker, suite.meta)

	suite.balancer = balance.NewMockBalancer(suite.T())
	suite.scheduler = task.NewMockScheduler(suite.T())
	suite.controller = NewCheckerController(suite.meta, suite.dist, suite.targetManager, suite.balancer, suite.nodeMgr, suite.scheduler, suite.broker)
}

func (suite *CheckerControllerSuite) TestBasic() {

	// set meta
	suite.meta.CollectionManager.PutCollection(utils.CreateTestCollection(1, 1))
	suite.meta.ReplicaManager.Put(utils.CreateTestReplica(1, 1, []int64{1, 2}))
	suite.nodeMgr.Add(session.NewNodeInfo(1, "localhost"))
	suite.nodeMgr.Add(session.NewNodeInfo(2, "localhost"))
	suite.meta.ResourceManager.AssignNode(meta.DefaultResourceGroupName, 1)
	suite.meta.ResourceManager.AssignNode(meta.DefaultResourceGroupName, 2)

	// set target
	segments := []*datapb.SegmentInfo{
		{
			ID:            1,
			PartitionID:   1,
			InsertChannel: "test-insert-channel",
		},
	}
	suite.broker.EXPECT().GetRecoveryInfoV2(mock.Anything, int64(1)).Return(
		nil, segments, nil)
	suite.targetManager.UpdateCollectionNextTargetWithPartitions(int64(1), int64(1))

	// set dist
	suite.dist.ChannelDistManager.Update(2, utils.CreateTestChannel(1, 2, 1, "test-insert-channel"))
	suite.dist.LeaderViewManager.Update(2, utils.CreateTestLeaderView(2, 1, "test-insert-channel", map[int64]int64{}, map[int64]*meta.Segment{}))

	counter := atomic.NewInt64(0)
	suite.scheduler.EXPECT().Add(mock.Anything).Run(func(task task.Task) {
		counter.Inc()
	}).Return(nil)
	suite.scheduler.EXPECT().GetSegmentTaskNum().Return(0).Maybe()
	suite.scheduler.EXPECT().GetChannelTaskNum().Return(0).Maybe()

	suite.balancer.EXPECT().AssignSegment(mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.balancer.EXPECT().AssignChannel(mock.Anything, mock.Anything).Return(nil)
	ctx := context.Background()
	suite.controller.Start(ctx)
	defer suite.controller.Stop()

	suite.Eventually(func() bool {
		suite.controller.Check()
		return counter.Load() > 0
	}, 5*time.Second, 1*time.Second)
}

func TestCheckControllerSuite(t *testing.T) {
	suite.Run(t, new(CheckerControllerSuite))
}
