//go:build integrationtest
// +build integrationtest

package specs

import (
	"context"
	"fmt"
	"os"
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
							Name: "pwxbuild",
						},
					},
					ServiceAccountName: "px-object-controller",
					Containers: []corev1.Container{
						{
							Name:            "px-object-controller",
							Image:           os.Getenv("PX_OBJECT_CONTROLLER_IMG"),
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env:             []corev1.EnvVar{},
						},
					},
				},
			},
		},
	}
)

// EnvConfig represents a config for setting up a test env
type EnvConfig struct {
	Namespace                  string
	S3AdminAccessKeyID         string
	S3AdminSecretAccessKey     string
	PureFBAdminAccessKeyID     string
	PureFBAdminSecretAccessKey string

	ImagePullSecretUsername string
	ImagePullSecretPassword string
}

// TestConfig represents a config for setting up a test
type TestConfig struct {
	Env          *EnvConfig
	Namespace    string
	RetainBucket bool
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
func GetPXObjectControllerDeployment(ec *EnvConfig) *v1.Deployment {
	deployment := &simpleDeploymentTemplate

	// Add AWS S3 secret references
	deployment = addDeploymentSecret(deployment, "S3_ADMIN_ACCESS_KEY_ID", "object-service-credentials", "S3AdminAccessKeyID")
	deployment = addDeploymentSecret(deployment, "S3_ADMIN_SECRET_ACCESS_KEY", "object-service-credentials", "S3AdminSecretAccessKey")
	deployment = addDeploymentSecret(deployment, "PURE_FB_ADMIN_ACCESS_KEY_ID", "object-service-credentials", "PureFBAdminAccessKeyID")
	deployment = addDeploymentSecret(deployment, "PURE_FB_ADMIN_SECRET_ACCESS_KEY", "object-service-credentials", "PureFBAdminSecretAccessKey")

	return deployment
}

// CreatePXObjectControllerDeployment creates the px-object deployment and any dependencies
func CreatePXObjectControllerDeployment(k8sClient *kubernetes.Clientset, ec *EnvConfig) error {
	deployment := GetPXObjectControllerDeployment(ec)
	if ec.Namespace == "" {
		ec.Namespace = "kube-system"
	}

	// Create Object service Credentials
	_, err := k8sClient.CoreV1().Secrets(ec.Namespace).Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "object-service-credentials",
			Namespace: ec.Namespace,
		},
		Data: map[string][]byte{
			"S3AdminAccessKeyID":         []byte(ec.S3AdminAccessKeyID),
			"S3AdminSecretAccessKey":     []byte(ec.S3AdminSecretAccessKey),
			"PureFBAdminAccessKeyID":     []byte(ec.PureFBAdminAccessKeyID),
			"PureFBAdminSecretAccessKey": []byte(ec.PureFBAdminSecretAccessKey),
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	// Create RBAC
	_, err = k8sClient.CoreV1().ServiceAccounts(ec.Namespace).Create(context.TODO(), &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "px-object-controller",
			Namespace: ec.Namespace,
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
				Verbs:     []string{"get", "list", "create", "delete", "update"},
			},
			{
				APIGroups: []string{"object.portworx.io"},
				Resources: []string{"pxbucketclaims", "pxbucketaccesses", "pxbucketclasses"},
				Verbs:     []string{"list", "watch", "create", "update", "patch", "get"},
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
				Namespace: ec.Namespace,
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

	_, err = k8sClient.RbacV1().Roles(ec.Namespace).Create(context.TODO(), &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "px-object-controller",
			Namespace: ec.Namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"coordination.k8s.io"},
				Resources: []string{"leases"},
				Verbs:     []string{"get", "watch", "list", "delete", "update", "create"},
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	_, err = k8sClient.RbacV1().RoleBindings(ec.Namespace).Create(context.TODO(), &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "px-object-controller",
			Namespace: ec.Namespace,
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
	_, err = k8sClient.AppsV1().Deployments(ec.Namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
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
		dep, err := k8sClient.AppsV1().Deployments(ec.Namespace).Get(context.TODO(), deployment.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if dep.Status.ReadyReplicas == *deployment.Spec.Replicas && dep.Status.UpdatedReplicas == *deployment.Spec.Replicas {
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
