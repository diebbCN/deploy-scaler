/*
Copyright 2025.

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

package controller

import (
	"context"

	"fmt"
	apiv1alpha1 "github.com/jack410/deploy-scaler/api/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	//"k8s.io/apimachinery/pkg/util/json"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

var logger = log.Log.WithName("scaler_controller")

// ScalerReconciler reconciles a Scaler object
type ScalerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=api.scaler.com,resources=scalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=api.scaler.com,resources=scalers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=api.scaler.com,resources=scalers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Scaler object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.0/pkg/reconcile
func (r *ScalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	log := logger.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	log.Info("Reconcile called")
	//创建一个scaler实例
	scaler := &apiv1alpha1.Scaler{}
	err := r.Get(ctx, req.NamespacedName, scaler)
	if err != nil {
		//如果没有这个scaler实例，我们就用client.IgnoreNotFound(err)来忽略错误，让进程不中断
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	startTime := scaler.Spec.Start
	endTime := scaler.Spec.End
	replicas := scaler.Spec.Replicas

	currentHour := time.Now().Local().Hour()
	fmt.Println(currentHour, 666666666)
	log.Info(fmt.Sprintf("currentTime: %d", currentHour))

	//从startTime开始，endtime为止
	if currentHour >= startTime && currentHour < endTime {
		//加个日志便于调试
		log.Info("starting to call scaleDeployment func")
		// 从scaller实例中遍历deployments
		err := scalerDeployment(scaler, r, ctx, replicas)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{RequeueAfter: time.Duration(10 * time.Second)}, nil
}

func scalerDeployment(scaler *apiv1alpha1.Scaler, r *ScalerReconciler, ctx context.Context, replicas int32) error {
	for _, deploy := range scaler.Spec.Deployments {
		// 创建一个新的deployment实例
		deployment := &v1.Deployment{}
		err := r.Get(ctx, types.NamespacedName{
			Name:      deploy.Name,
			Namespace: deploy.Namespace,
		}, deployment)
		if err != nil {
			return err
		}

		//判断当前k8s集群中的deployment副本数是否等于scaler中的副本数
		if deployment.Spec.Replicas != &replicas {
			deployment.Spec.Replicas = &replicas
			r.Update(ctx, deployment)
			if err != nil {
				//如果更新deployment失败，状态为failed
				scaler.Status.Status = apiv1alpha1.FAILED
				r.Status().Update(ctx, scaler)
				return err
			}

			scaler.Status.Status = apiv1alpha1.SUCCESS
			r.Status().Update(ctx, scaler)
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1alpha1.Scaler{}).
		Named("scaler").
		Complete(r)
}
