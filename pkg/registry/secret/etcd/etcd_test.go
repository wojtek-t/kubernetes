/*
Copyright 2015 Google Inc. All rights reserved.

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

package etcd

import (
	"testing"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/rest/resttest"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/testapi"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/tools"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/tools/etcdtest"
)

func newHelper(t *testing.T) (*tools.FakeEtcdClient, tools.EtcdHelper) {
	fakeEtcdClient := tools.NewFakeEtcdClient(t)
	fakeEtcdClient.TestIndex = true
	helper := tools.NewEtcdHelper(fakeEtcdClient, testapi.Codec(), etcdtest.PathPrefix())
	return fakeEtcdClient, helper
}

func validNewSecret(name string) *api.Secret {
	return &api.Secret{
		ObjectMeta: api.ObjectMeta{
			Name:      name,
			Namespace: api.NamespaceDefault,
		},
		Data: map[string][]byte{
			"test": []byte("data"),
		},
	}
}

func TestCreate(t *testing.T) {
	fakeEtcdClient, helper := newHelper(t)
	storage := NewStorage(helper)
	test := resttest.New(t, storage, fakeEtcdClient.SetError)
	secret := validNewSecret("foo")
	secret.Name = ""
	secret.GenerateName = "foo-"
	test.TestCreate(
		// valid
		secret,
		// invalid
		&api.Secret{},
		&api.Secret{
			ObjectMeta: api.ObjectMeta{Name: "name"},
			Data:       map[string][]byte{"name with spaces": []byte("")},
		},
		&api.Secret{
			ObjectMeta: api.ObjectMeta{Name: "name"},
			Data:       map[string][]byte{".dotfile": []byte("")},
		},
	)
}

func TestUpdate(t *testing.T) {
	fakeEtcdClient, helper := newHelper(t)
	storage := NewStorage(helper)
	test := resttest.New(t, storage, fakeEtcdClient.SetError)
	key := etcdtest.AddPrefix("secrets/default/foo")

	fakeEtcdClient.ExpectNotFoundGet(key)
	fakeEtcdClient.ChangeIndex = 2
	secret := validNewSecret("foo")
	existing := validNewSecret("exists")
	obj, err := storage.Create(api.NewDefaultContext(), existing)
	if err != nil {
		t.Fatalf("unable to create object: %v", err)
	}
	older := obj.(*api.Secret)
	older.ResourceVersion = "1"

	test.TestUpdate(
		secret,
		existing,
		older,
	)
}
