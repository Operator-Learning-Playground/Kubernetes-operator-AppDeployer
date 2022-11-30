```bigquery

// 不好用，废弃
func (r *AppDeployerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var appDeploy deployv1.AppDeployer
	err := r.Get(ctx, req.NamespacedName, &appDeploy)
	if err != nil {
		// 找不到资源对象，直接返回，不再回到queue中。
		// 删除一个不存在的对象，可能会报not-found错误
		// 这种情况不需要返回queue(requeue)
		if client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	// 当前资源对象已经被标记删除
	if appDeploy.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	// 1. 如果不存在的资源，需要判断是否创建
	// 2. 如果存在的资源，需要判断是否更新
	// 3. 如果需要更新，则直接更新
	// 4. 如果不需要更新，则正常返回

	// 没找到，需要直接创建
	deployment := &appsv1.Deployment{}
	if err := r.Client.Get(ctx, req.NamespacedName, deployment); err != nil && errors.IsNotFound(err) {
		// 关联Annotation
		data, err := json.Marshal(appDeploy.Spec)
		if err != nil {
			return ctrl.Result{}, err
		}

		if appDeploy.Annotations != nil {
			// annotation不为空，直接加入
			appDeploy.Annotations[oldSpecAnnotation] = string(data)
		} else {
			appDeploy.Annotations = map[string]string{oldSpecAnnotation: string(data)}
		}

		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err := r.Client.Update(ctx, &appDeploy); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return ctrl.Result{}, err
		}

		// 创建Deployment
		newDeployment := NewDeployment(&appDeploy)
		if err := r.Client.Create(ctx, newDeployment); err != nil {
			return ctrl.Result{}, err
		}

		// 创建Service
		newService := NewService(&appDeploy)
		if err := r.Client.Create(ctx, newService); err != nil {
			return ctrl.Result{}, err
		}
		// create成功，返回
		return ctrl.Result{}, nil
	}

	// 如果找到，需要更新，查看yaml文件是否变化
	oldSpec := deployv1.AppDeployerSpec{}
	if err := json.Unmarshal([]byte(appDeploy.Annotations[oldSpecAnnotation]), &oldSpec); err != nil {
		return ctrl.Result{}, err
	}

	// 方法一：不好用！
	// 比较新旧对象
	if !reflect.DeepEqual(appDeploy.Spec, oldSpec) {
		// Deployment
		newDeployment := NewDeployment(&appDeploy)
		oldDeployment := &appsv1.Deployment{}
		if err := r.Client.Get(ctx, req.NamespacedName, oldDeployment); err != nil {
			return ctrl.Result{}, err
		}
		oldDeployment.Spec = newDeployment.Spec
		// 能够更新oldDeployment
		// 一般情况不会直接 Update更新 如：r.Client.Update(ctx, oldDeployment)

		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err := r.Client.Update(ctx, oldDeployment); err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			return ctrl.Result{}, err
		}

		// Service
		newService := NewService(&appDeploy)
		oldService := &corev1.Service{}
		if err := r.Client.Get(ctx, req.NamespacedName, oldService); err != nil {
			return ctrl.Result{}, err
		}
		// 指定cluster ip是之前的
		newService.Spec.ClusterIPs = oldService.Spec.ClusterIPs
		oldService.Spec = newService.Spec
		// 一般情况不会直接 Update更新 如：r.Client.Update(ctx, oldService)

		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err := r.Client.Update(ctx, oldService); err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil

	}

	return ctrl.Result{}, nil
}
```