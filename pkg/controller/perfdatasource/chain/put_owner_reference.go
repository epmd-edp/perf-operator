package chain

import (
	"context"
	"github.com/epmd-edp/perf-operator/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/perf-operator/pkg/controller/perfdatasource/chain/handler"
	"github.com/epmd-edp/perf-operator/pkg/util"
	"github.com/epmd-edp/perf-operator/pkg/util/consts"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type PutOwnerReference struct {
	client client.Client
	scheme *runtime.Scheme
	next   handler.PerfDataSourceHandler
}

func (h PutOwnerReference) ServeRequest(ds *v1alpha1.PerfDataSource) error {
	log.Info("put owner reference for data source", "name", ds.Name)
	if err := h.setPerfOwnerRef(ds); err != nil {
		return err
	}
	log.Info("owner ref for perf data source has been added", "name", ds.Name)
	return nextServeOrNil(h.next, ds)
}

func (h PutOwnerReference) setPerfOwnerRef(ds *v1alpha1.PerfDataSource) error {
	log.Info("try to set owner ref for perf data source", "name", ds.Name)
	if ow := util.GetOwnerReference(consts.PerfServerKind, ds.GetOwnerReferences()); ow != nil {
		log.Info("PerfDataSource already has owner ref",
			"data source", ds.Name, "owner name", ow.Name)
		return nil
	}

	ps, err := util.GetPerfServerCr(h.client, ds.Spec.PerfName, ds.Namespace)
	if err != nil {
		return errors.Wrapf(err, "couldn't get %v PerfServer from cluster", ds.Spec.PerfName)
	}

	if err := controllerutil.SetControllerReference(ps, ds, h.scheme); err != nil {
		return errors.Wrapf(err, "couldn't set owner ref for %v PerfDataSource", ds.Name)
	}

	if err := h.client.Update(context.TODO(), ds); err != nil {
		return errors.Wrapf(err, "an error has been occurred while updating perf data source's owner %v", ds.Name)
	}
	return nil
}
