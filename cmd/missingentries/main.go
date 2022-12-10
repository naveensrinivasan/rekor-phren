package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/naveensrinivasan/rekor-phren/pkg"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	// auth provider gcp
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {
	// This checks for missing entries and creates jobs to update them.
	// These items could be missing because of a failure in the previous run.
	// This creates a k8s job for each missing entry.
	args := os.Args
	dataset := "phren"
	tableName := "phren"
	if len(args) > 3 {
		dataset = args[1]
		tableName = args[2]
	}
	for {
		missing, err := pkg.GetMissingEntries(dataset, tableName)
		if err != nil {
			panic(err)
		}
		if len(missing) == 0 {
			fmt.Println("all entries are present")
			return
		}
		for _, id := range missing {
			createJob(int(id))
		}
		time.Sleep(5 * time.Second)
	}
}
func buildConfig(kubeconfig string) (*rest.Config, error) {
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		return rest.InClusterConfig()
	}
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
}
func createJob(id int) {
	config, err := buildConfig("")
	if err != nil {
		panic(err.Error())
	}
	deleteJobTime := int32(60)
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	jobsClient := clientset.BatchV1().Jobs("default")
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("job-%d", id),
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &deleteJobTime,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "test",
							Image:   "gcr.io/openssf/rekor-phren-c5fc4a6e85fec69cce84b35fd28b14cc@sha256:4a52ce50e4e240b84d69e04f87d3df684f03a21640cb9229bd4fe8f63b1afc43",
							Command: []string{"rekor-phren", "--bigquery-dataset", "phren", "--bigquery-table-name", "rekor", "--rekor-url", "http://10.117.1.69", "--start-index", fmt.Sprintf("%d", id), "--end-index", fmt.Sprintf("%d", id+1), "update"},
						},
					},
					RestartPolicy:      corev1.RestartPolicyNever,
					ServiceAccountName: "phren",
				},
			},
		},
	}
	result, err := jobsClient.Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created job %q. \n", result.GetObjectMeta().GetName())
}
