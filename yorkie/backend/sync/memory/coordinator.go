/*
 * Copyright 2021 The Yorkie Authors. All rights reserved.
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

package memory

import (
	"context"

	"github.com/moby/locker"

	"github.com/yorkie-team/yorkie/pkg/document/key"
	"github.com/yorkie-team/yorkie/pkg/document/time"
	"github.com/yorkie-team/yorkie/pkg/types"
	"github.com/yorkie-team/yorkie/yorkie/backend/sync"
)

// Coordinator is a memory-based implementation of sync.Coordinator.
type Coordinator struct {
	agentInfo *sync.AgentInfo

	locks  *locker.Locker
	pubSub *PubSub
}

// NewCoordinator creates an instance of Coordinator.
func NewCoordinator(agentInfo *sync.AgentInfo) *Coordinator {
	return &Coordinator{
		agentInfo: agentInfo,
		locks:     locker.New(),
		pubSub:    NewPubSub(),
	}
}

// NewLocker creates locker of the given key.
func (m *Coordinator) NewLocker(
	ctx context.Context,
	key sync.Key,
) (sync.Locker, error) {
	return &internalLocker{
		key.String(),
		m.locks,
	}, nil
}

// Subscribe subscribes to the given documents.
func (m *Coordinator) Subscribe(
	_ context.Context,
	subscriber types.Client,
	keys []*key.Key,
) (*sync.Subscription, map[string][]types.Client, error) {
	sub, err := m.pubSub.Subscribe(subscriber, keys)
	if err != nil {
		return nil, nil, err
	}

	peersMap := m.pubSub.BuildPeersMap(keys)
	return sub, peersMap, nil
}

// Unsubscribe unsubscribes the given documents.
func (m *Coordinator) Unsubscribe(
	_ context.Context,
	keys []*key.Key,
	sub *sync.Subscription,
) error {
	m.pubSub.Unsubscribe(keys, sub)
	return nil
}

// Publish publishes the given event.
func (m *Coordinator) Publish(
	_ context.Context,
	publisherID *time.ActorID,
	event sync.DocEvent,
) {
	m.pubSub.Publish(publisherID, event)
}

// PublishToLocal publishes the given event.
func (m *Coordinator) PublishToLocal(
	_ context.Context,
	publisherID *time.ActorID,
	event sync.DocEvent,
) {
	m.pubSub.Publish(publisherID, event)
}

// UpdateMetadata updates the metadata of the given client.
func (m *Coordinator) UpdateMetadata(
	_ context.Context,
	publisher *types.Client,
	keys []*key.Key,
) (*sync.DocEvent, error) {
	m.pubSub.UpdateMetadata(publisher, keys)

	return &sync.DocEvent{
		Type:         types.MetadataChangedEvent,
		Publisher:    *publisher,
		DocumentKeys: keys,
	}, nil
}

// Members returns the members of this cluster.
func (m *Coordinator) Members() map[string]*sync.AgentInfo {
	members := make(map[string]*sync.AgentInfo)
	members[m.agentInfo.ID] = m.agentInfo
	return members
}

// Close closes all resources of this Coordinator.
func (m *Coordinator) Close() error {
	return nil
}
