package watch

import (
	"bytes"
	"context"
	"io"

	"github.com/rs/zerolog/log"
	v1core "k8s.io/api/core/v1"
)

type Pod struct {
	v1core.Pod
}

func (p *Pod) FailedLogs(container string) string {
	opts := v1core.PodLogOptions{
		Previous: true,
	}

	req := client.CoreV1().Pods(p.Namespace).GetLogs(p.Name, &opts)
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