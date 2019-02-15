/*
   Copyright 2018 Cedric Kring

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package kaniko

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cedrickring/kbuild/pkg/constants"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

//code used from github.com/GoogleContainerTools/skaffold
func (b Build) streamLogs(ctx context.Context, clientset *kubernetes.Clientset, podName string) func() {
	pods := clientset.CoreV1().Pods(b.Namespace)

	var wg sync.WaitGroup
	wg.Add(1)

	var bytesRead int64
	var retry int32 = 1

	go func() {
		defer wg.Done()

		for atomic.LoadInt32(&retry) == 1 {
			readCloser, err := pods.GetLogs(podName, &v1.PodLogOptions{
				Follow:    true,
				Container: constants.KanikoContainerName,
			}).Context(ctx).Stream()
			if err != nil {
				time.Sleep(1 * time.Second) //pod is still initializing
				continue
			}

			written, _ := io.Copy(os.Stdout, readCloser)
			atomic.AddInt64(&bytesRead, written)
			return
		}
	}()

	return func() {
		atomic.StoreInt32(&retry, 0)
		wg.Wait()

		if atomic.LoadInt64(&bytesRead) == 0 {
			logs, err := pods.GetLogs(podName, &v1.PodLogOptions{
				Container: constants.KanikoContainerName,
			}).DoRaw()
			if err == nil {
				fmt.Println(string(logs))
			}
		}
	}
}
