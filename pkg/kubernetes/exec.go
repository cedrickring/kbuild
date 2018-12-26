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

package kubernetes

import (
	"github.com/pkg/errors"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/remotecommand"
	"os"
)

//Exec contains all required information to execute a command in a container
type Exec struct {
	Namespace string
	PodName   string
	Container string

	Command []string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
}

//Exec executes a command in the specified container and opens a websocket stream if needed for Stdin
func (e Exec) Exec(client *kubernetes.Clientset) error {
	config, err := GetRestConfig()
	if err != nil {
		return errors.Wrap(err, "getting rest config")
	}

	req := client.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(e.PodName).
		Namespace(e.Namespace).
		SubResource("exec")

	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return errors.Wrap(err, "adding core/v1 to to new scheme")
	}

	parameterCodec := runtime.NewParameterCodec(scheme)
	req.VersionedParams(&corev1.PodExecOptions{
		Command:   e.Command,
		Container: e.Container,
		Stdin:     e.Stdin != nil,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, parameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return errors.Wrap(err, "couldn't create SPDY executor")
	}

	err = exec.Stream(e.getStreamOptions())
	if err != nil {
		return errors.Wrap(err, "error while streaming stdin")
	}

	return nil
}

func (e Exec) getStreamOptions() remotecommand.StreamOptions {
	options := remotecommand.StreamOptions{
		Tty: false,
	}

	if e.Stdin != nil {
		options.Stdin = e.Stdin
	}

	if e.Stdout == nil {
		options.Stdout = os.Stdout
	} else {
		options.Stdout = e.Stdout
	}

	if e.Stderr == nil {
		options.Stderr = os.Stderr
	} else {
		options.Stderr = e.Stderr
	}

	return options
}
