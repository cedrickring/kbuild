package kaniko

import (
	"bufio"
	"fmt"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"sync"
	"sync/atomic"
	"time"
)

//code used from github.com/GoogleContainerTools/skaffold
func (b Build) streamLogs(clientset *kubernetes.Clientset, podName string) func() {
	pods := clientset.CoreV1().Pods("default")

	var wg sync.WaitGroup
	wg.Add(1)

	var retry int32 = 1
	go func() {
		defer wg.Done()

		for atomic.LoadInt32(&retry) == 1 {
			readCloser, err := pods.GetLogs(podName, &v1.PodLogOptions{
				Follow:    true,
				Container: "kaniko-build",
			}).Stream()

			if err != nil {
				time.Sleep(1 * time.Second) //pod is still initializing
				continue
			}

			scanner := bufio.NewScanner(readCloser)
			for scanner.Scan() {
				fmt.Println(scanner.Text())
			}

			return
		}
	}()

	return func() {
		atomic.StoreInt32(&retry, 0)
		wg.Wait()
	}
}
