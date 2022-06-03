//go:build integrationtest
// +build integrationtest

package specs

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

var (
	intStrZero = intstr.FromInt(0)
	intStrOne  = intstr.FromInt(1)

	simpleDeploymentTemplate = v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "px-object-controller",
			Namespace: "kube-system",
		},
		Spec: v1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "px-object-controller",
				},
			},
			MinReadySeconds: 15,
			Strategy: v1.DeploymentStrategy{
				RollingUpdate: &v1.RollingUpdateDeployment{
					MaxSurge:       &intStrZero,
					MaxUnavailable: &intStrOne,
				},
				Type: v1.RollingUpdateDeploymentStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "px-object-controller",
					},
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets: []corev1.LocalObjectReference{
						{
							Name: "pxwbuild",
						},
					},
					ServiceAccountName: "px-object-controller",
					Containers: []corev1.Container{
						{
							Name:  "px-object-controller",
							Image: "portworx/px-object-controller:latest",
							Args: []string{
								"--leader-election=true",
								"--usage-interval=1m",
								"--log-level=trace",
								"--collector-source=pds",
								"--pds-api-endpoint=https://staging.pds-dev.io",
								"--pds-token-endpoint=http://release-staging-api.portworx.dev",
							},
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{
									Name:  "ZUORA_ENDPOINT",
									Value: "https://rest.apisandbox.zuora.com",
								},
							},
						},
					},
				},
			},
		},
	}
)

// TestConfig represents a config for setting up a test
type TestConfig struct {
	Namespace         string
	PdsUsername       string
	PdsPassword       string
	PdsClientID       string
	PdsClientSecret   string
	ZuoraClientID     string
	ZuoraClientSecret string
}

func addDeploymentSecret(deployment *v1.Deployment, envName, secretName, secretKey string) *v1.Deployment {
	deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name: envName,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secretName,
				},
				Key: secretKey,
			},
		},
	})

	return deployment
}

// GetPXObjectControllerDeployment returns a px-object deployment object
func GetPXObjectControllerDeployment(mc *TestConfig) *v1.Deployment {
	deployment := &simpleDeploymentTemplate

	// Add PDS secret references
	deployment = addDeploymentSecret(deployment, "PDS_USERNAME", "pds-credentials", "username")
	deployment = addDeploymentSecret(deployment, "PDS_PASSWORD", "pds-credentials", "password")
	deployment = addDeploymentSecret(deployment, "PDS_CLIENT_ID", "pds-credentials", "client-id")
	deployment = addDeploymentSecret(deployment, "PDS_CLIENT_SECRET", "pds-credentials", "client-secret")

	// Add zuora secret refs
	deployment = addDeploymentSecret(deployment, "ZUORA_CLIENT_ID", "zuora-credentials", "client-id")
	deployment = addDeploymentSecret(deployment, "ZUORA_CLIENT_SECRET", "zuora-credentials", "client-secret")

	return deployment
}

// CreatePXObjectControllerDeployment creates the px-object deployment and any dependencies
func CreatePXObjectControllerDeployment(k8sClient *kubernetes.Clientset, mc *TestConfig) error {
	deployment := GetPXObjectControllerDeployment(mc)
	if mc.Namespace == "" {
		mc.Namespace = "kube-system"
	}

	// Create PDS Credentials
	_, err := k8sClient.CoreV1().Secrets(mc.Namespace).Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pds-credentials",
			Namespace: mc.Namespace,
		},
		Data: map[string][]byte{
			"username":      []byte(mc.PdsUsername),
			"password":      []byte(mc.PdsPassword),
			"client-id":     []byte(mc.PdsClientID),
			"client-secret": []byte(mc.PdsClientSecret),
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	// Create Zuora Credentials
	_, err = k8sClient.CoreV1().Secrets(mc.Namespace).Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "zuora-credentials",
			Namespace: mc.Namespace,
		},
		Data: map[string][]byte{
			"client-id":     []byte(mc.ZuoraClientID),
			"client-secret": []byte(mc.ZuoraClientSecret),
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	// Create RBAC
	_, err = k8sClient.CoreV1().ServiceAccounts(mc.Namespace).Create(context.TODO(), &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "px-object-controller",
			Namespace: mc.Namespace,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	_, err = k8sClient.RbacV1().ClusterRoles().Create(context.TODO(), &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "px-object-controller-runner",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list"},
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	_, err = k8sClient.RbacV1().ClusterRoleBindings().Create(context.TODO(), &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "px-object-controller-role",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "px-object-controller",
				Namespace: mc.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "px-object-controller-runner",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	_, err = k8sClient.RbacV1().Roles(mc.Namespace).Create(context.TODO(), &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "px-object-controller",
			Namespace: mc.Namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"coordination.k8s.io"},
				Resources: []string{"leases"},
				Verbs:     []string{"get", "watch", "list", "delete", "update", "create"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"list", "watch", "create", "update", "patch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps"},
				Verbs:     []string{"create", "update", "get"},
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	_, err = k8sClient.RbacV1().RoleBindings(mc.Namespace).Create(context.TODO(), &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "px-object-controller",
			Namespace: mc.Namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				Name: "px-object-controller",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     "px-object-controller",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	// Create deployment
	_, err = k8sClient.AppsV1().Deployments(mc.Namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	// Wait for deployment to be ready
	pxObjectControllerBackoff := wait.Backoff{
		Steps:    25,
		Duration: 5 * time.Second,
		Factor:   1,
		Jitter:   0,
	}
	if err := wait.ExponentialBackoff(pxObjectControllerBackoff, func() (bool, error) {
		dep, err := k8sClient.AppsV1().Deployments(mc.Namespace).Get(context.TODO(), deployment.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if dep.Status.Replicas == *deployment.Spec.Replicas {
			return true, nil
		}

		fmt.Printf("Waiting %v for deployment %s to be ready: %v\n", pxObjectControllerBackoff.Duration, dep.Name, dep.Status)
		return false, nil
	}); err != nil {
		return err
	}

	return nil
}

func int32Ptr(i int32) *int32 { return &i }
