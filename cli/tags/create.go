package tags

import (
	"fmt"
	"github.com/configUpdate/cli/utils"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
	"os"
)

var createCli = &cobra.Command{
	Use: "create",
	Run: func(cmd *cobra.Command, args []string) {
		createTag(cmd)
	},
}

var (
	releaseName string
	projectName string
	branchName	string
)

func init() {

	createCli.Flags().StringVarP(&releaseName, "release", "r", "2021202", "该次升级产生的唯一标识")
	createCli.Flags().StringVarP(&projectName, "project", "p", "qa/config-repo", "配置仓库")
	createCli.Flags().StringVarP(&branchName, "branch", "b", "master", "仓库分支")

}

func createTag(cmd *cobra.Command) {
	client := utils.NewGitClient(cmd)
	msg := fmt.Sprintf("create a flag for update '%s'", releaseName)
	_, _, err := client.Tags.CreateTag(projectName, &gitlab.CreateTagOptions{
		TagName:            &releaseName,
		Ref:                &branchName,
		Message:            &msg,
	})
	if err != nil {
		fmt.Println("Fail to create tag, error: ", err)
		os.Exit(1)
	}
	fmt.Println("Create tag successfully")
}