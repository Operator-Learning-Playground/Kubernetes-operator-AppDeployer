package controllers

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime/schema"
	deployv1 "operator-develop/api/v1"
)

func MutateDeployment(appDeployer *deployv1.AppDeployer, deployment *appsv1.Deployment) {
	labels := map[string]string{
		"appDeployer": appDeployer.Name,
	}

	selector := metav1.LabelSelector{
		MatchLabels: labels,
	}

	deployment.Spec = appsv1.DeploymentSpec{
		Replicas: &appDeployer.Spec.Size,
		Selector: &selector,
		Template: corev1.PodTemplateSpec{ // Pod Template
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				Containers: newContainers(appDeployer),
			},
		},
	}

}

func MutateService(appDeployer *deployv1.AppDeployer, service *corev1.Service) {
	service.Spec = corev1.ServiceSpec{
		Type:  corev1.ServiceTypeNodePort,
		Ports: appDeployer.Spec.Ports,
		Selector: map[string]string{
			"appDeployer": appDeployer.Name,
		},
	}
}

// NewDeployment 创建deployment
func NewDeployment(appDeployer *deployv1.AppDeployer) *appsv1.Deployment {

	labels := map[string]string{
		"appDeployer": appDeployer.Name,
	}

	selector := metav1.LabelSelector{
		MatchLabels: labels,
	}

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "",
			Kind:       "",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appDeployer.Name,
			Namespace: appDeployer.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(appDeployer, schema.GroupVersionKind{
					Kind:    deployv1.Kind,
					Group:   deployv1.GroupVersion.Group,
					Version: deployv1.GroupVersion.Version,
				}),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &appDeployer.Spec.Size,
			Selector: &selector,
			Template: corev1.PodTemplateSpec{ // Pod Template
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: newContainers(appDeployer),
				},
			},
		},
	}

}

func newContainers(appDeploy *deployv1.AppDeployer) []corev1.Container {

	containerPorts := []corev1.ContainerPort{}

	for _, p := range appDeploy.Spec.Ports {
		containerPorts = append(containerPorts, corev1.ContainerPort{
			ContainerPort: p.TargetPort.IntVal,
		})
	}

	return []corev1.Container{
		{
			Name:      appDeploy.Name,
			Image:     appDeploy.Spec.Image,
			Ports:     containerPorts,
			Env:       appDeploy.Spec.Envs,
			Resources: appDeploy.Spec.Resources,
		},
	}
}

// NewService 创建service
func NewService(appDeployer *deployv1.AppDeployer) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appDeployer.Name,
			Namespace: appDeployer.Namespace,
			// 需要关联删除
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(appDeployer, schema.GroupVersionKind{
					Group:   deployv1.GroupVersion.Group,
					Version: deployv1.GroupVersion.Version,
					Kind:    deployv1.Kind,
				}),
			},
		},
		Spec: corev1.ServiceSpec{
			Type:  corev1.ServiceTypeNodePort,
			Ports: appDeployer.Spec.Ports,
			Selector: map[string]string{
				"appDeployer": appDeployer.Name,
			},
		},
	}
}
