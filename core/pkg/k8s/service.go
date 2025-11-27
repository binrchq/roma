package k8s

import (
	"context"
	"fmt"
	"os"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ServiceInfo struct {
	Host string
	Port string
}

func GetLoadBalancerAddress(serviceName, namespace string) (*ServiceInfo, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("无法获取 Kubernetes 配置: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建 Kubernetes 客户端失败: %v", err)
	}

	service, err := clientset.CoreV1().Services(namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("查询 Service 失败: %v", err)
	}

	if service.Spec.Type != corev1.ServiceTypeLoadBalancer {
		return nil, fmt.Errorf("Service %s 不是 LoadBalancer 类型", serviceName)
	}

	if len(service.Status.LoadBalancer.Ingress) == 0 {
		return nil, fmt.Errorf("LoadBalancer Service %s 还没有分配外部地址", serviceName)
	}

	ingress := service.Status.LoadBalancer.Ingress[0]
	var host string
	if ingress.Hostname != "" {
		host = ingress.Hostname
	} else if ingress.IP != "" {
		host = ingress.IP
	} else {
		return nil, fmt.Errorf("LoadBalancer Service %s 没有有效的外部地址", serviceName)
	}

	var port string
	if len(service.Spec.Ports) > 0 {
		// 对于 LoadBalancer 类型，优先使用 Service Port（LoadBalancer 的外部端口）
		// 但如果需要通过节点 IP 访问，可能需要使用 NodePort
		// 这里使用 Service Port，因为 LoadBalancer 的外部访问应该使用 Service Port
		// 如果 LoadBalancer 没有分配外部 IP，可以通过 NodePort 访问（但这种情况较少）
		port = strconv.Itoa(int(service.Spec.Ports[0].Port))
	} else {
		port = "2200"
	}

	return &ServiceInfo{
		Host: host,
		Port: port,
	}, nil
}

func GetLoadBalancerAddressFromEnv() (*ServiceInfo, error) {
	serviceName := os.Getenv("ROMA_SSH_SERVICE_NAME")
	namespace := os.Getenv("ROMA_SSH_SERVICE_NAMESPACE")

	if serviceName == "" || namespace == "" {
		return nil, fmt.Errorf("环境变量 ROMA_SSH_SERVICE_NAME 或 ROMA_SSH_SERVICE_NAMESPACE 未设置")
	}

	return GetLoadBalancerAddress(serviceName, namespace)
}

// GetNodePort 获取 LoadBalancer Service 的 NodePort
func GetNodePort(serviceName, namespace string) (string, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return "", fmt.Errorf("无法获取 Kubernetes 配置: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("创建 Kubernetes 客户端失败: %v", err)
	}

	service, err := clientset.CoreV1().Services(namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("查询 Service 失败: %v", err)
	}

	if service.Spec.Type != corev1.ServiceTypeLoadBalancer {
		return "", fmt.Errorf("Service %s 不是 LoadBalancer 类型", serviceName)
	}

	if len(service.Spec.Ports) == 0 {
		return "", fmt.Errorf("Service %s 没有配置端口", serviceName)
	}

	// 优先使用 NodePort，如果没有则使用 Service Port
	port := service.Spec.Ports[0]
	if port.NodePort != 0 {
		return strconv.Itoa(int(port.NodePort)), nil
	}

	// 如果没有 NodePort，使用 Service Port
	return strconv.Itoa(int(port.Port)), nil
}

// GetNodePortFromEnv 从环境变量获取 NodePort
func GetNodePortFromEnv() (string, error) {
	serviceName := os.Getenv("ROMA_SSH_SERVICE_NAME")
	namespace := os.Getenv("ROMA_SSH_SERVICE_NAMESPACE")

	if serviceName == "" || namespace == "" {
		return "", fmt.Errorf("环境变量 ROMA_SSH_SERVICE_NAME 或 ROMA_SSH_SERVICE_NAMESPACE 未设置")
	}

	return GetNodePort(serviceName, namespace)
}
