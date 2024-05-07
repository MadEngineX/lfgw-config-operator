/*
Copyright 2024.

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
	v1alpha1 "github.com/MadEngineX/lfgw-config-operator/api/v1alpha1"
	"github.com/MadEngineX/lfgw-config-operator/config"
	"github.com/sirupsen/logrus"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
)

const (
	appendItem = "append"
	removeItem = "remove"
)

var (
	defaultFinalizer = v1alpha1.GroupVersion.Group + "/finalyzer"
)

// ACLReconciler reconciles a ACL object
type ACLReconciler struct {
	client.Client
	log    *logrus.Entry
	Scheme *runtime.Scheme
	Mu     *sync.Mutex
}

//+kubebuilder:rbac:groups=controls.lfgw.io,resources=acls,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=controls.lfgw.io,resources=acls/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=controls.lfgw.io,resources=acls/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ACL object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *ACLReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	defer r.Mu.Unlock()
	r.Mu.Lock()

	var log = r.log.WithField("acl", req.NamespacedName)

	log.Infof("> Handling resource: %v", req.NamespacedName)

	// Get resource from k8s
	var desiredResource v1alpha1.ACL
	var err = r.Get(ctx, req.NamespacedName, &desiredResource)
	if err != nil {
		if kerrors.IsNotFound(err) {
			log.Debug("Resource have been already deleted")
			return ctrl.Result{}, nil
		}
		log.Errorf("< Error while reading CR ACL: %v", err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check deletion
	var deleted = desiredResource.GetDeletionTimestamp() != nil
	if deleted {
		err = r.processDefaultFinalization(log, ctx, &desiredResource, r.finalize)
		if err != nil {
			log.Errorf("Error while finalizing resource %v: %v", desiredResource.GetName(), err)
			r.updateStatus(log, ctx, &desiredResource, &v1alpha1.ACLStatus{Code: "ERROR", Message: err.Error()})
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer
	if !r.hasDefaultFinalizer(&desiredResource) {
		err = r.InjectDefaultFinalizer(ctx, &desiredResource)
		if err != nil {
			log.Errorf("Error while adding finalyzer to resource %v: %v", desiredResource.GetName(), err)
			r.updateStatus(log, ctx, &desiredResource, &v1alpha1.ACLStatus{Code: "ERROR", Message: err.Error()})
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		return ctrl.Result{}, nil
	}

	err = r.createOrUpdateACL(log, &desiredResource)
	if err != nil {
		log.Errorf("Error while syncing ACL to ConfigMap %v: %v", desiredResource.GetName(), err)
		r.updateStatus(log, ctx, &desiredResource, &v1alpha1.ACLStatus{Code: "ERROR", Message: err.Error()})
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	r.updateStatus(log, ctx, &desiredResource, &v1alpha1.ACLStatus{Code: "SYNCED", Message: "ACL added to ConfigMap"})

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ACLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = logrus.WithField("controller", "acl")

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ACL{}).
		Complete(r)
}

func (r *ACLReconciler) updateStatus(log *logrus.Entry, ctx context.Context, resource *v1alpha1.ACL, status *v1alpha1.ACLStatus) {
	if !reflect.DeepEqual(status, resource.Status) {
		resource.Status = *status
		var err = r.Client.Status().Update(ctx, resource)
		if err != nil {
			log.Errorf("  Error while updating resource status [%v.%v]: %v", resource.GetName(), resource.GetNamespace(), err)
		}
	}
}

func (r *ACLReconciler) processDefaultFinalization(
	log *logrus.Entry,
	ctx context.Context,
	resource *v1alpha1.ACL,
	finalizer func(resource *v1alpha1.ACL) error,
) error {
	if controllerutil.ContainsFinalizer(resource, defaultFinalizer) {
		// Launch finalization process
		if err := finalizer(resource); err != nil {
			return err
		}

		// Deletion finalizer and resource
		controllerutil.RemoveFinalizer(resource, defaultFinalizer)
		err := r.Client.Update(ctx, resource)
		if err != nil {
			return err
		}
		log.Infof("  Successfully deleted resource finalizer [%v.%v]", resource.GetName(), resource.GetNamespace())
	} else {
		log.Infof("  Successfully deleted resource [%v.%v]", resource.GetName(), resource.GetNamespace())
	}

	return nil
}

func (r *ACLReconciler) hasDefaultFinalizer(resource *v1alpha1.ACL) bool {
	return controllerutil.ContainsFinalizer(resource, defaultFinalizer)
}

func (r *ACLReconciler) InjectDefaultFinalizer(ctx context.Context, resource *v1alpha1.ACL) error {
	controllerutil.AddFinalizer(resource, defaultFinalizer)
	return r.Client.Update(ctx, resource)
}

func (r *ACLReconciler) finalize(resource *v1alpha1.ACL) error {

	// Check if this ConfigMap already exists
	found := &corev1.ConfigMap{}

	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: config.AppConfig.ConfigMapName, Namespace: config.AppConfig.ConfigMapNamespace}, found)
	// Create new ConfigMap if it doesn't exist
	if err != nil && kerrors.IsNotFound(err) {
		return fmt.Errorf("error, while finalizing ACL, unable to find ConfigMap: %s", err.Error())
	} else if err != nil {
		return fmt.Errorf("error, while trying to get existing ConfigMap: %s", err.Error())
	}

	configMap, updateNeeded := r.updateConfigMapObject(config.AppConfig.ConfigMapFilename, found, resource, removeItem)
	if updateNeeded {
		err = r.Client.Update(context.TODO(), configMap)
		if err != nil {
			return fmt.Errorf("error while updating existing ConfigMap: %s", err.Error())
		}
	}

	return nil
}

func (r *ACLReconciler) createOrUpdateACL(log *logrus.Entry, resource *v1alpha1.ACL) error {

	// Check if this ConfigMap already exists
	found := &corev1.ConfigMap{}

	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: config.AppConfig.ConfigMapName, Namespace: config.AppConfig.ConfigMapNamespace}, found)
	// Create new ConfigMap if it doesn't exist
	if err != nil && kerrors.IsNotFound(err) {

		log.Infof(" Creating a new ConfigMap, Namespace [%v], Name [%v], Filename [%v]", config.AppConfig.ConfigMapNamespace, config.AppConfig.ConfigMapName, config.AppConfig.ConfigMapFilename)
		// Define a new ConfigMap object
		configMap := r.prepareConfigMapObject(config.AppConfig.ConfigMapNamespace, config.AppConfig.ConfigMapName, config.AppConfig.ConfigMapFilename, resource)

		err = r.Client.Create(context.TODO(), configMap)
		if err != nil {
			return fmt.Errorf("error, while creating new ConfigMap: %s", err.Error())
		} else {
			return nil
		}

	} else if err != nil {
		return fmt.Errorf("error, while trying to get existing ConfigMap: %s", err.Error())
	}

	configMap, updateNeeded := r.updateConfigMapObject(config.AppConfig.ConfigMapFilename, found, resource, appendItem)
	if updateNeeded {
		err = r.Client.Update(context.TODO(), configMap)
		if err != nil {
			return fmt.Errorf("error while updating existing ConfigMap: %s", err.Error())
		}
	}

	return nil
}

func (r *ACLReconciler) prepareConfigMapObject(namespace, name, filename string, resource *v1alpha1.ACL) *corev1.ConfigMap {

	var data = make(map[string]string)
	var aclList string

	for _, rule := range resource.Spec.Rules {
		aclList += rule.RoleName + ": " + rule.NamespaceFilter + "\n"
	}

	data[filename] = aclList

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}

	return cm
}

func (r *ACLReconciler) updateConfigMapObject(filename string, configmap *corev1.ConfigMap, resource *v1alpha1.ACL, mode string) (*corev1.ConfigMap, bool) {

	var data = make(map[string]string)
	var aclList string
	var updateNeeded bool

	switch mode {
	case appendItem:
		for _, rule := range resource.Spec.Rules {
			if !strings.Contains(configmap.Data[filename], rule.RoleName+": "+rule.NamespaceFilter) {
				aclList += rule.RoleName + ": " + rule.NamespaceFilter + "\n"
				updateNeeded = true
			}
		}
		data[filename] = configmap.Data[filename] + aclList

	case removeItem:
		data[filename] = configmap.Data[filename]
		for _, rule := range resource.Spec.Rules {
			data[filename] = strings.Replace(data[filename], rule.RoleName+": "+rule.NamespaceFilter+"\n", "", 1)
		}
		updateNeeded = true
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configmap.Name,
			Namespace: configmap.Namespace,
		},
		Data: data,
	}

	return cm, updateNeeded
}
