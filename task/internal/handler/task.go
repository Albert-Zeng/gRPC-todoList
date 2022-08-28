package handler

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"task/discovery"
	"task/internal/dependency"
	"task/internal/repository"
	"task/internal/service"
	"task/pkg/e"
)

type TaskService struct {
}

func NewTaskService() *TaskService {
	return &TaskService{}
}

func (*TaskService) TaskCreate(ctx context.Context, req *service.TaskRequest) (resp *service.CommonResponse, err error) {
	var task repository.Task
	resp = new(service.CommonResponse)
	resp.Code = e.SUCCESS
	err = task.Create(req)
	if err != nil {
		resp.Code = e.ERROR
		resp.Msg = e.GetMsg(e.ERROR)
		resp.Data = err.Error()
		return resp, err
	}
	resp.Msg = e.GetMsg(uint(resp.Code))
	return resp, nil
}

func (*TaskService) TaskShow(ctx context.Context, req *service.TaskRequest) (resp *service.TasksDetailResponse, err error) {
	var t repository.Task
	resp = new(service.TasksDetailResponse)
	tRep, err := t.Show(req)
	resp.Code = e.SUCCESS
	if err != nil {
		resp.Code = e.ERROR
		return resp, err
	}
	resp.TaskDetail = repository.BuildTasks(tRep)

	// grpc 调用user模块，获取用户信息
	etcdAddress := []string{viper.GetString("etcd.address")}
	etcdRegister := discovery.NewRegister(etcdAddress, logrus.New())
	targetServer, err := etcdRegister.GetTargetServer("/v1/user/grpc")
	if err != nil {
		resp.Code = e.ERROR
		return resp, err
	}
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}
	userConn, _ := grpc.Dial(targetServer.Addr, opts...)
	user := dependency.NewUserServiceClient(userConn)
	userResp, err := user.UserInfo(context.Background(), &dependency.UserModel{
		UserID: req.UserID,
	})
	if err != nil {
		userResp = &dependency.UserDetailResponse{}
	}
	resp.UserName = userResp.GetUserDetail().UserName
	resp.NickName = userResp.GetUserDetail().NickName
	return resp, nil
}

func (*TaskService) TaskUpdate(ctx context.Context, req *service.TaskRequest) (resp *service.CommonResponse, err error) {
	var task repository.Task
	resp = new(service.CommonResponse)
	resp.Code = e.SUCCESS
	err = task.Update(req)
	if err != nil {
		resp.Code = e.ERROR
		resp.Msg = e.GetMsg(e.ERROR)
		resp.Data = err.Error()
		return resp, err
	}
	resp.Msg = e.GetMsg(uint(resp.Code))
	return resp, nil
}

func (*TaskService) TaskDelete(ctx context.Context, req *service.TaskRequest) (resp *service.CommonResponse, err error) {
	var task repository.Task
	resp = new(service.CommonResponse)
	resp.Code = e.SUCCESS
	err = task.Delete(req)
	if err != nil {
		resp.Code = e.ERROR
		resp.Msg = e.GetMsg(e.ERROR)
		resp.Data = err.Error()
		return resp, err
	}
	resp.Msg = e.GetMsg(uint(resp.Code))
	return resp, nil
}
