package watch

import (
	"bytes"
	"context"
	"io"

	"github.com/rs/zerolog/log"
	v1core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func LogsForPod(pod *v1core.Pod, container string, since *metav1.Time) string {
	opts := v1core.PodLogOptions{
		Container: container,
		SinceTime: since,
	}

	req := client.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &opts)
	logs, err := req.Stream(context.TODO())
	defer logs.Close()
	if err != nil {
		log.Error().Err(err).Msg("Error opening container log stream")
		return "error opening stream"
	}

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, logs)
	if err != nil {
		return "error in copy information from podLogs to buf"
	}
	str := buf.String()

	return str
}
