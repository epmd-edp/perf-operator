package chain

import (
	"errors"
	"fmt"
	"github.com/epmd-edp/perf-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/perf-operator/v2/pkg/client/perf/mock"
	"github.com/epmd-edp/perf-operator/v2/pkg/model/command"
	"github.com/epmd-edp/perf-operator/v2/pkg/model/dto"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"strings"
	"testing"
)

const (
	sonarDsType      = "SONAR"
	fakeCodebaseName = "stub-val"
)

func TestPutDataSource_ShouldUpdateSonarDataSourceWithoutActivating(t *testing.T) {
	pds := &v1alpha1.PerfDataSourceSonar{
		ObjectMeta: v1.ObjectMeta{
			Namespace: fakeNamespace,
			OwnerReferences: []v1.OwnerReference{
				{
					Kind: "PerfServer",
					Name: fakeName,
				},
			},
		},
		Spec: v1alpha1.PerfDataSourceSonarSpec{
			Type: sonarDsType,
			Config: v1alpha1.DataSourceSonarConfig{
				ProjectKeys: []string{fmt.Sprintf("/%v/%v-Build-%v", fakeName, strings.ToUpper(fakeName), fakeName),
					fmt.Sprintf("/%v/%v-Build-%v", fakeCodebaseName, strings.ToUpper(fakeName), fakeCodebaseName)},
				Url: fakeName,
			},
		},
	}

	ps := &v1alpha1.PerfServer{
		ObjectMeta: v1.ObjectMeta{
			Name:      fakeName,
			Namespace: fakeNamespace,
		},
		Spec: v1alpha1.PerfServerSpec{
			ProjectName: fakeName,
		},
	}

	sec := &coreV1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      sonarDataSourceSecretName,
			Namespace: fakeNamespace,
		},
		Data: map[string][]byte{
			"username": []byte("fake"),
			"password": []byte("fake"),
		},
	}

	objs := []runtime.Object{
		pds, ps, sec,
	}

	s := scheme.Scheme
	s.AddKnownTypes(v1.SchemeGroupVersion, pds, ps)

	mPerfCl := new(mock.MockPerfClient)
	ch := PutDataSource{
		client:     fake.NewFakeClient(objs...),
		perfClient: mPerfCl,
	}

	mPerfCl.On("GetProjectDataSource", fakeName, sonarDsType).
		Return(&dto.DataSource{
			Active: true,
			Type:   sonarDsType,
			Config: map[string]interface{}{
				"projectKeys": []interface{}{fmt.Sprintf("/%v/%v-Build-%v", fakeName, strings.ToUpper(fakeName), fakeName)},
			},
		}, nil)

	mPerfCl.On("UpdateDataSource", command.DataSourceCommand{
		Type: sonarDsType,
		Config: command.DataSourceSonarConfig{
			ProjectKeys: []string{fmt.Sprintf("/%v/%v-Build-%v", fakeName, strings.ToUpper(fakeName), fakeName),
				fmt.Sprintf("/%v/%v-Build-%v", fakeCodebaseName, strings.ToUpper(fakeName), fakeCodebaseName)},
			Url:      fakeName,
			Username: "fake",
			Password: "fake",
		},
	}).Return(nil)

	assert.NoError(t, ch.ServeRequest(pds))
	assert.Equal(t, "created", pds.Status.Status)
}

func TestPutDataSource_ShouldUpdateSonarDataSourceWithActivating(t *testing.T) {
	pds := &v1alpha1.PerfDataSourceSonar{
		ObjectMeta: v1.ObjectMeta{
			Namespace: fakeNamespace,
			OwnerReferences: []v1.OwnerReference{
				{
					Kind: "PerfServer",
					Name: fakeName,
				},
			},
		},
		Spec: v1alpha1.PerfDataSourceSonarSpec{
			Type: sonarDsType,
			Config: v1alpha1.DataSourceSonarConfig{
				ProjectKeys: []string{fmt.Sprintf("/%v/%v-Build-%v", fakeName, strings.ToUpper(fakeName), fakeName),
					fmt.Sprintf("/%v/%v-Build-%v", fakeCodebaseName, strings.ToUpper(fakeName), fakeCodebaseName)},
				Url: fakeName,
			},
		},
	}

	ps := &v1alpha1.PerfServer{
		ObjectMeta: v1.ObjectMeta{
			Name:      fakeName,
			Namespace: fakeNamespace,
		},
		Spec: v1alpha1.PerfServerSpec{
			ProjectName: fakeName,
		},
	}

	sec := &coreV1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      sonarDataSourceSecretName,
			Namespace: fakeNamespace,
		},
		Data: map[string][]byte{
			"username": []byte("fake"),
			"password": []byte("fake"),
		},
	}

	objs := []runtime.Object{
		pds, ps, sec,
	}

	s := scheme.Scheme
	s.AddKnownTypes(v1.SchemeGroupVersion, pds, ps)

	mPerfCl := new(mock.MockPerfClient)
	ch := PutDataSource{
		client:     fake.NewFakeClient(objs...),
		perfClient: mPerfCl,
	}

	mPerfCl.On("GetProjectDataSource", fakeName, sonarDsType).
		Return(&dto.DataSource{
			Active: false,
			Type:   sonarDsType,
			Config: map[string]interface{}{
				"projectKeys": []interface{}{fmt.Sprintf("/%v/%v-Build-%v", fakeName, strings.ToUpper(fakeName), fakeName)},
			},
		}, nil)

	mPerfCl.On("UpdateDataSource", command.DataSourceCommand{
		Type: sonarDsType,
		Config: command.DataSourceSonarConfig{
			ProjectKeys: []string{fmt.Sprintf("/%v/%v-Build-%v", fakeName, strings.ToUpper(fakeName), fakeName),
				fmt.Sprintf("/%v/%v-Build-%v", fakeCodebaseName, strings.ToUpper(fakeName), fakeCodebaseName)},
			Url:      fakeName,
			Username: "fake",
			Password: "fake",
		},
	}).Return(nil)

	mPerfCl.On("ActivateDataSource", fakeName, 0).Return(nil)

	assert.NoError(t, ch.ServeRequest(pds))
	assert.Equal(t, "created", pds.Status.Status)
}

func TestPutDataSource_ShouldCreateSonarDataSource(t *testing.T) {
	pds := &v1alpha1.PerfDataSourceSonar{
		ObjectMeta: v1.ObjectMeta{
			Namespace: fakeNamespace,
			OwnerReferences: []v1.OwnerReference{
				{
					Kind: "PerfServer",
					Name: fakeName,
				},
			},
		},
		Spec: v1alpha1.PerfDataSourceSonarSpec{
			Type: sonarDsType,
			Config: v1alpha1.DataSourceSonarConfig{
				ProjectKeys: []string{fmt.Sprintf("/%v/%v-Build-%v", fakeName, strings.ToUpper(fakeName), fakeName),
					fmt.Sprintf("/%v/%v-Build-%v", fakeCodebaseName, strings.ToUpper(fakeName), fakeCodebaseName)},
				Url: fakeName,
			},
		},
	}

	ps := &v1alpha1.PerfServer{
		ObjectMeta: v1.ObjectMeta{
			Name:      fakeName,
			Namespace: fakeNamespace,
		},
		Spec: v1alpha1.PerfServerSpec{
			ProjectName: fakeName,
		},
	}

	sec := &coreV1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      sonarDataSourceSecretName,
			Namespace: fakeNamespace,
		},
		Data: map[string][]byte{
			"username": []byte("fake"),
			"password": []byte("fake"),
		},
	}

	objs := []runtime.Object{
		pds, ps, sec,
	}

	s := scheme.Scheme
	s.AddKnownTypes(v1.SchemeGroupVersion, pds, ps)

	mPerfCl := new(mock.MockPerfClient)
	ch := PutDataSource{
		client:     fake.NewFakeClient(objs...),
		perfClient: mPerfCl,
	}

	mPerfCl.On("GetProjectDataSource", fakeName, sonarDsType).Return(nil, nil)

	mPerfCl.On("CreateDataSource", fakeName, command.DataSourceCommand{
		Type: sonarDsType,
		Config: command.DataSourceSonarConfig{
			ProjectKeys: []string{fmt.Sprintf("/%v/%v-Build-%v", fakeName, strings.ToUpper(fakeName), fakeName),
				fmt.Sprintf("/%v/%v-Build-%v", fakeCodebaseName, strings.ToUpper(fakeName), fakeCodebaseName)},
			Url:      fakeName,
			Username: "fake",
			Password: "fake",
		},
	}).Return(nil)

	assert.NoError(t, ch.ServeRequest(pds))
	assert.Equal(t, "created", pds.Status.Status)
}

func TestPutDataSource_ShouldNotFindDataSourceInPERF(t *testing.T) {
	ps := &v1alpha1.PerfServer{
		ObjectMeta: v1.ObjectMeta{
			Name:      fakeName,
			Namespace: fakeNamespace,
		},
		Spec: v1alpha1.PerfServerSpec{
			ProjectName: fakeName,
		},
	}

	objs := []runtime.Object{
		ps,
	}

	s := scheme.Scheme
	s.AddKnownTypes(v1.SchemeGroupVersion, ps)

	mPerfCl := new(mock.MockPerfClient)
	ch := PutDataSource{
		client:     fake.NewFakeClient(objs...),
		perfClient: mPerfCl,
	}

	mPerfCl.On("GetProjectDataSource", fakeName, "").Return(nil, errors.New("failed"))

	pds := &v1alpha1.PerfDataSourceSonar{
		ObjectMeta: v1.ObjectMeta{
			Namespace: fakeNamespace,
			OwnerReferences: []v1.OwnerReference{
				{
					Kind: "PerfServer",
					Name: fakeName,
				},
			},
		},
	}

	assert.Error(t, ch.ServeRequest(pds))
	assert.Equal(t, "error", pds.Status.Status)
}

func TestPutDataSource_ShouldNotActivateDataSource(t *testing.T) {
	pds := &v1alpha1.PerfDataSourceSonar{
		ObjectMeta: v1.ObjectMeta{
			Namespace: fakeNamespace,
			OwnerReferences: []v1.OwnerReference{
				{
					Kind: "PerfServer",
					Name: fakeName,
				},
			},
		},
		Spec: v1alpha1.PerfDataSourceSonarSpec{
			Type: sonarDsType,
			Config: v1alpha1.DataSourceSonarConfig{
				ProjectKeys: []string{fmt.Sprintf("/%v/%v-Build-%v", fakeName, strings.ToUpper(fakeName), fakeName),
					fmt.Sprintf("/%v/%v-Build-%v", fakeCodebaseName, strings.ToUpper(fakeName), fakeCodebaseName)},
				Url: fakeName,
			},
		},
	}

	ps := &v1alpha1.PerfServer{
		ObjectMeta: v1.ObjectMeta{
			Name:      fakeName,
			Namespace: fakeNamespace,
		},
		Spec: v1alpha1.PerfServerSpec{
			ProjectName: fakeName,
		},
	}

	sec := &coreV1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      sonarDataSourceSecretName,
			Namespace: fakeNamespace,
		},
		Data: map[string][]byte{
			"username": []byte("fake"),
			"password": []byte("fake"),
		},
	}

	objs := []runtime.Object{
		pds, ps, sec,
	}

	s := scheme.Scheme
	s.AddKnownTypes(v1.SchemeGroupVersion, pds, ps)

	mPerfCl := new(mock.MockPerfClient)
	ch := PutDataSource{
		client:     fake.NewFakeClient(objs...),
		perfClient: mPerfCl,
	}

	mPerfCl.On("GetProjectDataSource", fakeName, sonarDsType).
		Return(&dto.DataSource{
			Active: false,
			Type:   sonarDsType,
			Config: map[string]interface{}{
				"projectKeys": []interface{}{fmt.Sprintf("/%v/%v-Build-%v", fakeName, strings.ToUpper(fakeName), fakeName)},
			},
		}, nil)

	mPerfCl.On("UpdateDataSource", command.DataSourceCommand{
		Type: sonarDsType,
		Config: command.DataSourceSonarConfig{
			ProjectKeys: []string{fmt.Sprintf("/%v/%v-Build-%v", fakeName, strings.ToUpper(fakeName), fakeName),
				fmt.Sprintf("/%v/%v-Build-%v", fakeCodebaseName, strings.ToUpper(fakeName), fakeCodebaseName)},
			Url:      fakeName,
			Username: "fake",
			Password: "fake",
		},
	}).Return(nil)

	mPerfCl.On("ActivateDataSource", fakeName, 0).Return(errors.New("failed"))

	assert.Error(t, ch.ServeRequest(pds))
	assert.Equal(t, "error", pds.Status.Status)
}

func TestPutDataSource_ShouldNotUpdateDataSourceBecauseOfMissingNewParameters(t *testing.T) {
	pds := &v1alpha1.PerfDataSourceSonar{
		ObjectMeta: v1.ObjectMeta{
			Namespace: fakeNamespace,
			OwnerReferences: []v1.OwnerReference{
				{
					Kind: "PerfServer",
					Name: fakeName,
				},
			},
		},
		Spec: v1alpha1.PerfDataSourceSonarSpec{
			Type: sonarDsType,
			Config: v1alpha1.DataSourceSonarConfig{
				ProjectKeys: []string{fmt.Sprintf("/%v/%v-Build-%v", fakeName, strings.ToUpper(fakeName), fakeName)},
				Url:         fakeName,
			},
		},
	}

	ps := &v1alpha1.PerfServer{
		ObjectMeta: v1.ObjectMeta{
			Name:      fakeName,
			Namespace: fakeNamespace,
		},
		Spec: v1alpha1.PerfServerSpec{
			ProjectName: fakeName,
		},
	}

	sec := &coreV1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      sonarDataSourceSecretName,
			Namespace: fakeNamespace,
		},
		Data: map[string][]byte{
			"username": []byte("fake"),
			"password": []byte("fake"),
		},
	}

	objs := []runtime.Object{
		pds, ps, sec,
	}

	s := scheme.Scheme
	s.AddKnownTypes(v1.SchemeGroupVersion, pds, ps)

	mPerfCl := new(mock.MockPerfClient)
	ch := PutDataSource{
		client:     fake.NewFakeClient(objs...),
		perfClient: mPerfCl,
	}

	mPerfCl.On("GetProjectDataSource", fakeName, sonarDsType).
		Return(&dto.DataSource{
			Active: true,
			Type:   sonarDsType,
			Config: map[string]interface{}{
				"projectKeys": []interface{}{fmt.Sprintf("/%v/%v-Build-%v", fakeName, strings.ToUpper(fakeName), fakeName)},
			},
		}, nil)

	mPerfCl.On("UpdateDataSource", command.DataSourceCommand{
		Type: sonarDsType,
		Config: command.DataSourceSonarConfig{
			ProjectKeys: []string{fmt.Sprintf("/%v/%v-Build-%v", fakeName, strings.ToUpper(fakeName), fakeName),
				fmt.Sprintf("/%v/%v-Build-%v", fakeCodebaseName, strings.ToUpper(fakeName), fakeCodebaseName)},
			Url:      fakeName,
			Username: "fake",
			Password: "fake",
		},
	}).Return(nil)

	assert.NoError(t, ch.ServeRequest(pds))
}
