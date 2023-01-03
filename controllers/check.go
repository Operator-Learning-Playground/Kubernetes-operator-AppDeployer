package controllers

import (
	v1 "k8s.io/api/core/v1"
	deployv1 "operator-develop/api/v1"
	"strconv"
)

const ServiceClusterIP = "ClusterIP"

// 检查Service ServiceType
func checkService(appDeployer *deployv1.AppDeployer) bool {
	if appDeployer.Spec.ServiceType == ServiceClusterIP && appDeployer.Spec.Ports[0].NodePort != 0 {
		return false
	}
	return true
}

func setVolumes(appDeployer *deployv1.AppDeployer) []v1.Volume {
	var volumes []v1.Volume
	a := v1.LocalObjectReference{
		Name: appDeployer.Name,
	}
	if appDeployer.Spec.Configmap {
		volumes = []v1.Volume{{
			Name: appDeployer.Name,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: a,
				},
			},
		},
	}
	} else {
		volumes = []v1.Volume{}
	}



	return volumes

}

func StringToInt(a string) int {

	res, _ := strconv.Atoi(a)

	return res



}
