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
