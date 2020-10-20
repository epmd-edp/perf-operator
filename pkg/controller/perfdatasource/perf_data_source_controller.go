package perfdatasource

import (
	"context"
	edpv1alpha1 "github.com/epmd-edp/perf-operator/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/perf-operator/pkg/controller/perfdatasource/chain"
	"github.com/epmd-edp/perf-operator/pkg/util"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

var (
	log = logf.Log.WithName("controller_perf_data_source")
)

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePerfDataSource{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
	}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("perfdatasource-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	if err = c.Watch(&source.Kind{Type: &edpv1alpha1.PerfDataSource{}}, &handler.EnqueueRequestForObject{}); err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcilePerfDataSource{}

type ReconcilePerfDataSource struct {
	client client.Client
	scheme *runtime.Scheme
}

func (r *ReconcilePerfDataSource) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	rl := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	rl.V(2).Info("Reconciling PerfDataSource")

	i := &edpv1alpha1.PerfDataSource{}
	if err := r.client.Get(context.TODO(), request.NamespacedName, i); err != nil {
		if k8serrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	ps, err := util.GetPerfServerCr(r.client, i.Spec.PerfName, i.Namespace)
	if err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "couldn't get %v PerfServer from cluster", i.Spec.PerfName)
	}

	if !ps.Status.Available {
		log.Info("Perf instance is unavailable. skip creating/updating data source in PERF", "name", ps.Name)
		return reconcile.Result{RequeueAfter: 1 * time.Minute}, nil
	}

	if err := chain.CreateDefChain(r.client, r.scheme).ServeRequest(i); err != nil {
		return reconcile.Result{}, err
	}

	rl.Info("Reconciling PerfDataSource has been finished")
	return reconcile.Result{}, nil
}
