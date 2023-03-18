package utils

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
	"context"
	"net/http"
	"os"
)


func NewGitClient(cmd *cobra.Command) *gitlab.Client {
	gitUrl, _ := cmd.Flags().GetString("address")
	gitToken, _ := cmd.Flags().GetString("token")
	client, err := gitlab.NewClient(gitToken, gitlab.WithBaseURL(gitUrl), gitlab.WithCustomRetry(func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		if err != nil {
			return false, err
		}
		if (resp.StatusCode == 400 || resp.StatusCode == 429 || resp.StatusCode >= 500) {
			return true, nil
		}
		return false, nil
	}))
	if err != nil {
		fmt.Println("init git client error:", err)
		os.Exit(1)
	}
	return client
}