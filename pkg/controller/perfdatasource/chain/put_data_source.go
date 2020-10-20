package chain

import (
	"github.com/epmd-edp/perf-operator/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/perf-operator/pkg/client/perf"
	"github.com/epmd-edp/perf-operator/pkg/controller/perfdatasource/chain/handler"
	"github.com/epmd-edp/perf-operator/pkg/util"
	"github.com/epmd-edp/perf-operator/pkg/util/consts"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PutDataSource struct {
	next   handler.PerfDataSourceHandler
	client client.Client
}

func (h PutDataSource) ServeRequest(ds *v1alpha1.PerfDataSource) error {
	log.Info("start creating/updating data source in PERF", "name", ds.Name)
	return nil
}

func (h PutDataSource) tryToPutDataSource(ds v1alpha1.PerfDataSource) error {
	ow := util.GetOwnerReference(consts.PerfServerKind, ds.GetOwnerReferences())
	ps, err := util.GetPerfServerCr(h.client, ow.Name, ds.Namespace)
	if err != nil {
		return errors.Wrapf(err, "couldn't get %v PerfServer from cluster", ds.Spec.PerfName)
	}

	_, err = h.newPerfRestClient(ps.Spec.ApiUrl, ps.Spec.CredentialName, ps.Namespace)
	if err != nil {
		return errors.Wrapf(err, "couldn't init PERF client", "url", ps.Spec.RootUrl)
	}

	/*reqDs, err := c.GetDataSource(ds.Name)
	if err != nil {
		return errors.Wrapf(err, "couldn't get data source from PERF.", "id", ds.Name)
	}*/
	return nil
}

func (h PutDataSource) newPerfRestClient(url, secretName, namespace string) (*perf.PerfClientAdapter, error) {
	s, err := util.GetSecret(h.client, secretName, namespace)
	if err != nil {
		return nil, err
	}

	c, err := perf.NewRestClient(url, string(s.Data["username"]), string(s.Data["password"]))
	if err != nil {
		return nil, err
	}
	return c, nil
}
