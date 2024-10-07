package service

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/aerokube/selenoid/session"
	"github.com/emicklei/go-restful/v3/log"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Kubernetes struct {
	ServiceBase
	Environment
	Client *rest.Config
}

func (k *Kubernetes) StartWithCancel() (*StartedService, error) {
	clientset, err := kubernetes.NewForConfig(k.Client)
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("selenoid-browser-%d", k.RequestId)
	podClient := clientset.CoreV1().Pods(k.Environment.KubernetesNamespace)
	reqID := fmt.Sprintf("%d", k.RequestId)
	pod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"selenoid-request-id": reqID,
			},
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Name:  "browser",
					Image: k.Service.Image.(string),
					Ports: []apiv1.ContainerPort{
						{
							Name:          "browser",
							Protocol:      apiv1.ProtocolTCP,
							ContainerPort: 4444,
						},
						{
							Name:          "vnc",
							Protocol:      "TCP",
							ContainerPort: 5900,
						},
						{
							Name:          "devtools",
							Protocol:      "TCP",
							ContainerPort: 7070,
						},
						{
							Name:          "fileserver",
							Protocol:      "TCP",
							ContainerPort: 8080,
						},
						{
							Name:          "clipboard",
							Protocol:      "TCP",
							ContainerPort: 9090,
						},
					},
					Args: []string{
						"-session-attempt-timeout",
						"240s",
						"-service-startup-timeout",
						"240s",
					},
					LivenessProbe: &apiv1.Probe{
						ProbeHandler: apiv1.ProbeHandler{
							HTTPGet: &apiv1.HTTPGetAction{
								Port: intstr.FromString("browser"),
								Path: k.Service.Path,
							},
						},
					},
					ReadinessProbe: &apiv1.Probe{
						ProbeHandler: apiv1.ProbeHandler{
							HTTPGet: &apiv1.HTTPGetAction{
								Port: intstr.FromString("browser"),
								Path: k.Service.Path,
							},
						},
					},
				},
			},
		},
	}
	pod, err = podClient.Create(context.Background(), pod, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	ready := false
	for !ready {
		log.Printf("[KUBERNETES_BACKEND] Waiting for the pod to be ready")
		time.Sleep(10 * time.Second)
		ready = true
		statuses := pod.Status.ContainerStatuses
		for _, status := range statuses {
			if !status.Ready {
				ready = false
			}
		}
	}
	log.Printf("[KUBERNETES_BACKEND] Pod is ready")

	svcClient := clientset.CoreV1().Services(k.Environment.KubernetesNamespace)
	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: apiv1.ServiceSpec{
			Selector: map[string]string{
				"selenoid-request-id": reqID,
			},
			Ports: []apiv1.ServicePort{
				{
					Name:     "browser",
					Protocol: apiv1.ProtocolTCP,
					Port:     4444,
				},
				{
					Name:     "vnc",
					Protocol: "TCP",
					Port:     5900,
				},
				{
					Name:     "devtools",
					Protocol: "TCP",
					Port:     7070,
				},
				{
					Name:     "fileserver",
					Protocol: "TCP",
					Port:     8080,
				},
				{
					Name:     "clipboard",
					Protocol: "TCP",
					Port:     9090,
				},
			},
		},
	}
	_, err = svcClient.Create(context.Background(), service, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	// Wait until pod is running
	podUpdated, err := clientset.CoreV1().Pods(k.Environment.KubernetesNamespace).Get(context.Background(), name, metav1.GetOptions{})
	svcUpdated, err := clientset.CoreV1().Services(k.Environment.KubernetesNamespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	// Create a pod
	// Create a service
	// Wait until pod is ready
	// Define the StartedService
	hp := session.HostPort{
		Selenium: net.JoinHostPort(fmt.Sprintf("%s.selenoid.svc.cluster.local", name), "4444"),
	}
	u := &url.URL{Scheme: "http", Host: hp.Selenium, Path: k.Service.Path}
	s := StartedService{
		Url:    u,
		Origin: net.JoinHostPort(fmt.Sprintf("%s.selenoid.svc.cluster.local", name), "4444"),
		Container: &session.Container{
			ID:        string(podUpdated.ObjectMeta.GetUID()),
			IPAddress: svcUpdated.Spec.ClusterIP,
			Ports:     map[string]string{"4444": "4444"},
		},
		HostPort: hp,
		Cancel: func() {
			if err := k.Cancel(context.Background(), k.RequestId, podUpdated.Name, svcUpdated.Name); err != nil {
				log.Printf("[KUBERNETES_ERROR] %s", err)
			}
		},
	}
	return &s, nil
}

func (k *Kubernetes) Cancel(ctx context.Context, requestID uint64, podName, serviceName string) error {

	clientset, err := kubernetes.NewForConfig(k.Client)
	if err != nil {
		return err
	}
	podClient := clientset.CoreV1().Pods(k.KubernetesNamespace)
	if err := podClient.Delete(ctx, podName, *metav1.NewDeleteOptions(60)); err != nil {
		return err
	}
	svcClient := clientset.CoreV1().Services(k.KubernetesNamespace)
	if err := svcClient.Delete(ctx, serviceName, *metav1.NewDeleteOptions(60)); err != nil {
		return err
	}
	return nil
}
