/*
Copyright 2016 The Rook Authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Some of the code below came from https://github.com/coreos/etcd-operator
which also has the apache 2.0 license.
*/
package operator

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/rook/rook/pkg/operator/cluster"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwatch "k8s.io/apimachinery/pkg/watch"
)

type clusterEvent struct {
	Type   kwatch.EventType
	Object *cluster.Cluster
}

type poolEvent struct {
	Type   kwatch.EventType
	Object *cluster.Pool
}

type rawEvent struct {
	Type   kwatch.EventType
	Object json.RawMessage
}

func pollClusterEvent(decoder *json.Decoder) (*clusterEvent, *metav1.Status, error) {
	re, status, err := pollEvent(decoder)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to poll cluster event. %+v", err)
	}

	if status != nil {
		return nil, status, nil
	}

	ev := &clusterEvent{
		Type:   re.Type,
		Object: &cluster.Cluster{},
	}
	err = json.Unmarshal(re.Object, ev.Object)
	if err != nil {
		return nil, nil, fmt.Errorf("fail to unmarshal Cluster object from data (%s): %v", re.Object, err)
	}
	return ev, status, nil
}

func pollPoolEvent(decoder *json.Decoder) (*poolEvent, *metav1.Status, error) {
	re, status, err := pollEvent(decoder)
	if err != nil {
		return nil, status, fmt.Errorf("failed to poll pool event. %+v", err)
	}

	if status != nil {
		return nil, status, nil
	}

	ev := &poolEvent{
		Type:   re.Type,
		Object: &cluster.Pool{},
	}
	err = json.Unmarshal(re.Object, ev.Object)
	if err != nil {
		return nil, nil, fmt.Errorf("fail to unmarshal Pool object from data (%s): %v", re.Object, err)
	}
	return ev, nil, nil
}

func pollEvent(decoder *json.Decoder) (*rawEvent, *metav1.Status, error) {
	re := &rawEvent{}
	err := decoder.Decode(re)
	if err != nil {
		if err == io.EOF {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("fail to decode raw event from apiserver (%v)", err)
	}

	if re.Type == kwatch.Error {
		status := &metav1.Status{}
		err = json.Unmarshal(re.Object, status)
		if err != nil {
			return nil, nil, fmt.Errorf("fail to decode (%s) into metav1.Status (%v)", re.Object, err)
		}
		logger.Infof("returning pollEvent status %+v", status)
		return nil, status, nil
	}

	return re, nil, nil
}
