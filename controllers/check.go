package controllers

import deployv1 "operator-develop/api/v1"

const ServiceClusterIP = "ClusterIP"

// 检查Service ServiceType
func checkService(appDeployer *deployv1.AppDeployer) bool {
	if appDeployer.Spec.ServiceType == ServiceClusterIP && appDeployer.Spec.Ports[0].NodePort != 0 {
		return false
	}

	return true
}
