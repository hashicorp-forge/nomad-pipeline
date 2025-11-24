package helper

import (
	"fmt"
	"net/url"
	"time"

	"github.com/ryanuber/columnize"

	"github.com/hashicorp-forge/nomad-pipeline/pkg/api/v1"
)

func FormatKV(in []string) string {
	columnConf := columnize.DefaultConfig()
	columnConf.Empty = "<none>"
	columnConf.Glue = " = "
	return columnize.Format(in, columnConf)
}

func FormatTime(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}
	return t.Format(time.RFC3339)
}

func FormatError(cliMsg string, err error) string {

	var code int

	switch e := err.(type) {
	case *api.ResponseError:
		code = e.ErrorBody.Code
	case *url.Error:
		code = 500
	default:
		code = 400
	}

	return FormatKV([]string{
		fmt.Sprintf("Description|%s", cliMsg),
		fmt.Sprintf("Error|%s", err),
		fmt.Sprintf("Code|%v", code),
	})
}
