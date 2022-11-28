package docker

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dc "github.com/docker/docker/client"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"swan-provider/logs"
	"swan-provider/models"
	"time"
)

var myCli *DockerCli

type DockerCli struct {
	Image    string
	Name     string
	DataPath string

	Client *dc.Client
	Ctx    context.Context

	Cid string
}

func GetMyCli(image string, name string, path string) *DockerCli {

	if myCli == nil {
		cli := &DockerCli{
			Image:    image,
			Name:     name,
			DataPath: path,
		}

		client, err := dc.NewClientWithOpts(dc.FromEnv)
		if err != nil {
			logs.GetLog().Error(err)
			return nil
		}
		cli.Client = client
		cli.Ctx = context.Background()

		clist, err := cli.Client.ContainerList(cli.Ctx, types.ContainerListOptions{All: true})
		if err != nil {
			logs.GetLog().Error(err)
			return nil
		}

		finded := false
		for _, container := range clist {
			if "/"+cli.Name == container.Names[0] {
				cli.Cid = container.ID
				finded = true
				break
			}
		}
		if !finded {
			//
		}

		myCli = cli
		return myCli
	}

	return myCli
}

func GetDockerCli(image string, name string, path string) *DockerCli {
	dockerCli := &DockerCli{
		Image:    image,
		Name:     name,
		DataPath: path,
	}

	client, err := dc.NewClientWithOpts(dc.FromEnv)
	if err != nil {
		logs.GetLog().Error(err)
		return nil
	}

	dockerCli.Client = client
	dockerCli.Ctx = context.Background()
	dockerCli.Cid = ""

	return dockerCli
}

func (cli *DockerCli) UpdateCid() (bool, error) {

	clist, err := cli.Client.ContainerList(cli.Ctx, types.ContainerListOptions{All: true})
	if err != nil {
		logs.GetLog().Error(err)
		return false, err
	}

	found := false
	for _, container := range clist {
		if "/"+cli.Name == container.Names[0] {
			cli.Cid = container.ID
			found = true
			break
		}
	}

	if !found {
		return false, errors.New("do not find container")
	}

	return true, nil
}

func (cli *DockerCli) PoktCtnCreate() bool {
	out, err := cli.Client.ImagePull(cli.Ctx, cli.Image, types.ImagePullOptions{})
	if err != nil {
		logs.GetLog().Error("Image Pull Error:", err)
		return false
	}
	defer out.Close()
	io.Copy(os.Stdout, out)

	body, err := cli.Client.ContainerCreate(
		cli.Ctx,
		&container.Config{Image: cli.Image, Tty: true},
		&container.HostConfig{NetworkMode: "host", Binds: []string{cli.DataPath + ":/home/app/.pocket"}},
		nil,
		nil,
		cli.Name)
	if err != nil {
		logs.GetLog().Error("Container Create Error:", err)
		return false
	}

	logs.GetLog().Debug("Container Create Id:", body.ID[:10])
	cli.Cid = body.ID
	return true
}

func (cli *DockerCli) PoktCtnPullAndCreate(cmd, env []string, autoRemove bool) bool {
	out, err := cli.Client.ImagePull(cli.Ctx, cli.Image, types.ImagePullOptions{})
	if err != nil {
		logs.GetLog().Error("Image Pull Error:", err)
		return false
	}
	defer out.Close()
	io.Copy(os.Stdout, out)

	logs.GetLog().Debug("Container Create Para DataPath=", cli.DataPath, " autoRemove=", autoRemove)
	body, err := cli.Client.ContainerCreate(
		cli.Ctx,
		&container.Config{
			Cmd:   cmd,
			Env:   env,
			Image: cli.Image,
			Tty:   true},
		&container.HostConfig{
			NetworkMode: "host",
			Binds:       []string{cli.DataPath + ":/home/app/.pocket"},
			AutoRemove:  autoRemove},
		nil,
		nil,
		cli.Name)
	if err != nil {
		logs.GetLog().Error("Container Create Error:", err)
		return false
	}

	logs.GetLog().Info("Account Container Create Id:", body.ID[:10])
	cli.Cid = body.ID
	return true
}

func (cli *DockerCli) PoktCtnCreateRun(cmd, env []string, autoRemove bool) bool {

	body, err := cli.Client.ContainerCreate(
		cli.Ctx,
		&container.Config{
			Cmd:   cmd,
			Env:   env,
			Image: cli.Image,
			Tty:   true},
		&container.HostConfig{
			NetworkMode: "host",
			Binds:       []string{cli.DataPath + ":/home/app/.pocket"},
			AutoRemove:  autoRemove},
		nil,
		nil,
		cli.Name)
	if err != nil {
		logs.GetLog().Error("Container Create Error:", err)
		return false
	}

	logs.GetLog().Debug("Container Create Id:", body.ID[:10])
	cli.Cid = body.ID
	return true
}

func (cli *DockerCli) PoktCtnExist() bool {
	clist, err := cli.Client.ContainerList(cli.Ctx, types.ContainerListOptions{All: true})
	if err != nil {
		logs.GetLog().Error(err)
		return false
	}

	for _, container := range clist {
		if "/"+cli.Name == container.Names[0] {
			cli.Cid = container.ID
			return true
		}
	}

	logs.GetLog().Debug("Can Not Find Container:", " Name=", cli.Name)
	return false
}

func (cli *DockerCli) PoktCtnList() bool {
	clist, err := cli.Client.ContainerList(cli.Ctx, types.ContainerListOptions{All: true})
	if err != nil {
		logs.GetLog().Error(err)
		return false
	}

	for _, container := range clist {
		logs.GetLog().Info("Container Create Name:", container.Names[0], " ID=", container.ID[:10])
	}
	return true
}

func (cli *DockerCli) PoktCtnStart() bool {

	containers, err := cli.Client.ContainerList(cli.Ctx, types.ContainerListOptions{All: true})
	if err != nil {
		logs.GetLog().Error(err)
		return false
	}

	for _, container := range containers {
		if "/"+cli.Name == container.Names[0] {
			cli.Cid = container.ID

			if !strings.Contains(container.Status, "Up") {
				err := cli.Client.ContainerStart(context.Background(), cli.Cid, types.ContainerStartOptions{})
				if err != nil {
					logs.GetLog().Error("Container Start Error:", err)
					return false
				}
				logs.GetLog().Debug("Container start:", " id=", cli.Cid, " name=", cli.Name)
				return true
			} else {

				logs.GetLog().Debug("Container Already Running:", " id=", cli.Cid[:10], " name=", cli.Name, " status=", container.Status)
				return true
			}

			break
		}
	}

	return false
}

func (cli *DockerCli) PoktCtnStop() bool {
	timeout := time.Second * 5
	err := cli.Client.ContainerStop(cli.Ctx, cli.Cid, &timeout)
	if err != nil {
		logs.GetLog().Error("Container Stop Error:", err)
		return false
	}

	logs.GetLog().Debug("Stop Container Id:", cli.Cid[:10])
	return true
}

func (cli *DockerCli) PoktCtnExec(cmd []string) (string, error) {

	rst, err := cli.Client.ContainerExecCreate(
		cli.Ctx, cli.Cid,
		types.ExecConfig{AttachStdout: true, AttachStderr: true, Cmd: cmd})
	if err != nil {
		logs.GetLog().Error("Container Exec Create Error:", err)
		return "", err
	}

	response, err := cli.Client.ContainerExecAttach(cli.Ctx, rst.ID, types.ExecStartCheck{})
	if err != nil {
		logs.GetLog().Error("Container Exec Attach Error:", err)
		return "", err
	}
	defer response.Close()

	data, _ := ioutil.ReadAll(response.Reader)
	logs.GetLog().Debug("Container Exec Response:", string(data))

	return string(data), nil
}

func (cli *DockerCli) PoktCtnExecVersion() (*models.VersionData, error) {
	strRes, err := cli.PoktCtnExec([]string{"pocket", "version"})
	if err != nil {
		logs.GetLog().Error("Exec Pocket Version Error:", err)
		return nil, err
	}

	index := strings.Index(strRes, ":")
	jOut := &models.VersionData{
		Version: strings.TrimSuffix(strRes[index+2:], "\n"),
	}
	logs.GetLog().Debug("pocket query version result:", jOut.Version)

	return jOut, nil
}

func (cli *DockerCli) PoktCtnExecNodeAddress() (string, error) {
	strRes, err := cli.PoktCtnExec([]string{"pocket", "accounts", "list"})
	if err != nil {
		logs.GetLog().Error("Exec Pocket Node Account Error:", err)
		return "", err
	}

	index := strings.Index(strRes, ")")
	index += 2
	jOut := strRes[index : index+40]
	logs.GetLog().Debug("pocket query node account result:", jOut)

	_, finded := os.LookupEnv("TEST_POCKET_MODE")
	if finded {
		//ONLY FOR TEST
		jOut = "ffad090789253ad0439c56b7b9c301f90424d5b7"
	}

	return jOut, nil
}

func (cli *DockerCli) PoktCtnExecSetAccount(address string) (string, error) {
	rsp, err := cli.PoktCtnExec([]string{"pocket", "accounts", "set-validator", address})
	if err != nil {
		logs.GetLog().Error("Exec Pocket Set Account Error:", err)
		return "", err
	}
	return rsp, nil
}

func (cli *DockerCli) PoktCtnExecHeight() (*models.HeightData, error) {
	strRes, err := cli.PoktCtnExec([]string{"pocket", "query", "height"})
	if err != nil {
		logs.GetLog().Error("Exec Pocket Height Error:", err)
		return nil, err
	}

	jOut := &models.HeightData{}
	index := strings.Index(strRes, "{")
	logs.GetLog().Debug("pocket query height result for json:", strRes[index:])
	err = json.Unmarshal([]byte(strRes[index:]), jOut)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return jOut, nil
}

func (cli *DockerCli) PoktCtnExecBalance(address string) (*models.BalanceData, error) {
	strRes, err := cli.PoktCtnExec([]string{"pocket", "query", "balance", address})
	if err != nil {
		logs.GetLog().Error("Exec Pocket Balance Error:", err)
		return nil, err
	}

	jOut := &models.BalanceData{}
	index := strings.Index(strRes, "{")
	logs.GetLog().Debug("pocket query balance result for json:", strRes[index:])
	err = json.Unmarshal([]byte(strRes[index:]), jOut)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return jOut, nil
}

func (cli *DockerCli) PoktCtnExecSignInfo(address string) ([]*models.SignInfo, error) {
	strRes, err := cli.PoktCtnExec([]string{"pocket", "query", "signing-info", address})
	if err != nil {
		logs.GetLog().Error("Exec Pocket Sign Info Error:", err)
		return nil, err
	}

	jOut := &models.SignInfoResponse{}
	index := strings.Index(strRes, "{")
	logs.GetLog().Debug("pocket query signing-info result for json:", strRes[index:])
	err = json.Unmarshal([]byte(strRes[index:]), jOut)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return jOut.Result, nil
}

func (cli *DockerCli) PoktCtnExecNode(address string) (*models.NodeData, error) {
	strRes, err := cli.PoktCtnExec([]string{"pocket", "query", "node", address})
	if err != nil {
		logs.GetLog().Error("Exec Pocket Node Error:", err)
		return nil, err
	}

	jOut := &models.NodeData{}
	index := strings.Index(strRes, "{")
	logs.GetLog().Debug("pocket query node result for json:", strRes[index:])
	err = json.Unmarshal([]byte(strRes[index:]), jOut)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return jOut, nil
}

func (cli *DockerCli) PoktCtnExecCustodial(address, amount, relayChainIDs, serviceURI, networkID, fee, isBefore string) (string, error) {
	rsp, err := cli.PoktCtnExec([]string{"expect", "~/.pocket/custodial.sh", address, amount, relayChainIDs, serviceURI, networkID, fee, isBefore})
	if err != nil {
		logs.GetLog().Error("Exec Pocket Custodial Error:", err)
		return "", err
	}
	logs.GetLog().Debug("Exec Pocket Custodial Result:", rsp)

	return rsp, nil
}

func (cli *DockerCli) PoktCtnExecNonCustodial(pubKey, outputAddr, amount, relayChainIDs, serviceURI, networkID, fee, isBefore string) (string, error) {
	rsp, err := cli.PoktCtnExec([]string{"expect", "~/.pocket/noncustodial.sh", pubKey, outputAddr, amount, relayChainIDs, serviceURI, networkID, fee, isBefore})
	if err != nil {
		logs.GetLog().Error("Exec Pocket NonCustodial Error:", err)
		return "", err
	}
	return rsp, nil
}
