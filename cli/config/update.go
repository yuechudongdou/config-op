package config

import (
	"encoding/base64"
	"fmt"
	"github.com/configUpdate/cli/utils"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
	"io/ioutil"
	"math"
	"os"
	"path"
	"strings"
	"time"
)

var updateCli = &cobra.Command{
	Use: "update",
	Run: func(cmd *cobra.Command, args []string) {
		updateConfigFiles(cmd)
	},
}

var (
	serverName  string
	releaseName string
	profiles    string
	projectName string
	branchName	string
	configPath string
	createBranch bool
)

func init() {
	updateCli.Flags().StringVarP(&serverName, "server", "s", "service-core", "服务名字")
	updateCli.Flags().StringVarP(&releaseName, "release", "r", "2021202", "该次升级产生的唯一标识")
	updateCli.Flags().StringVarP(&profiles, "profiles", "f", "drm,cus", "active profiles")
	updateCli.Flags().StringVarP(&projectName, "project", "p", "qa/config-repo", "配置仓库")
	updateCli.Flags().StringVarP(&branchName, "branch", "b", "test4", "仓库分支")
	updateCli.Flags().StringVarP(&configPath, "dir", "d", "/Users/maxq/test", "配置文件绝对路径")
	updateCli.Flags().BoolVarP(&createBranch, "create-branch", "c", false, "创建分支")

}

func getConfigBaseName(configPath string) string {
	files, err := ioutil.ReadDir(configPath)
	if err != nil {
		fmt.Println("读取文件列表错误; ", "filePath: ", configPath, "error: ", err)
		os.Exit(0)
	}
	fileLength := math.MaxInt32
	configName := ""
	for _, fileInfo := range files {
		fileName := fileInfo.Name()
		tmpLength := len(fileName)
		if tmpLength < fileLength {
			fileLength = tmpLength
			configName = strings.TrimSuffix(fileName, path.Ext(fileName))
		}
	}
	return configName
}

func updateConfigFiles(cmd *cobra.Command) {
	client := utils.NewGitClient(cmd)
	configName := getConfigBaseName(configPath)
	fmt.Println(configName)
	configFiles := []string{path.Join(configPath, fmt.Sprintf("%s.yml", configName))}
	for _, profile := range strings.Split(profiles, ",") {
		if profile != "cus" {
			configFiles = append(configFiles, path.Join(configPath, fmt.Sprintf("%s-%s.yml", configName, profile)))
		}
	}
	fmt.Println(configFiles)
	_, resp, _ := client.Branches.GetBranch(projectName, branchName)
	if resp.StatusCode == 404 {
		if createBranch {
			defaultBranch := "master"
			_, _, err := client.Branches.CreateBranch(projectName, &gitlab.CreateBranchOptions{
				Branch: &branchName,
				Ref:    &defaultBranch,
			})
			if err != nil {
				fmt.Println("create branch err: ", err)
				os.Exit(1)
			}
		} else {
			timer := time.NewTicker(time.Second)
			for {
				select {
				case <- timer.C:
					fmt.Println("wait config server create the branch")
					_, resp, _ := client.Branches.GetBranch(projectName, branchName)
					if resp.StatusCode != 404 {
						timer.Stop()
						goto updateRepo
					}
				}
			}
		}
	}
updateRepo:
	var fileActions []*gitlab.CommitActionOptions
	for _, configFile := range configFiles {
		content, err := ioutil.ReadFile(configFile)
		gitFile := path.Base(configFile)
		if err != nil {
			continue
		}
		fileContent := string(content)
		if err != nil {
			fmt.Println("read file error")
			os.Exit(1)
		}
		configFile = path.Base(configFile)
		destContent, resp, err := client.RepositoryFiles.GetFile(projectName, configFile, &gitlab.GetFileOptions{Ref: &branchName})
		if resp.StatusCode == 404 {
			action := gitlab.FileCreate
			fileActions = append(fileActions, &gitlab.CommitActionOptions{
				Action:          &action,
				FilePath:        &gitFile,
				Content:         &fileContent,
			})
		} else {
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			} else {
				fileBytes, err := base64.StdEncoding.DecodeString(destContent.Content)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				if fileContent != string(fileBytes) {
					action := gitlab.FileUpdate
					fileActions = append(fileActions, &gitlab.CommitActionOptions{
						Action:          &action,
						FilePath:        &gitFile,
						Content:         &fileContent,
					})
				}
			}
		}
	}
	//defaultBranch := "master"
	if len(fileActions) == 0 {
		os.Exit(0)
	}
	fmt.Printf("file actiions is %v", fileActions)
	commitMsg := fmt.Sprintf("update %s config file for '%s' update", serverName, releaseName)
	commit, _, err := client.Commits.CreateCommit(projectName, &gitlab.CreateCommitOptions{
		// StartBranch: &defaultBranch,
		Actions: fileActions,
		Branch: &branchName,
		CommitMessage: &commitMsg,
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Sprintf("%v", commit)
}