package handler

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/gomodule/redigo/redis"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net/http"
	"user/internal/cache"
	"user/internal/repository"
	"user/internal/service"
	"user/pkg/e"
)


func auth(token string) int {
	_, err := cache.Get(token)
	if err == redis.ErrNil {
		return http.StatusUnauthorized
	}
	if err != nil {
		return http.StatusInternalServerError
	}
	return 0
}

func AuthGrpc(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// todo: 后端服务间调用（非api-gateway）鉴权
	return handler(ctx, req)
}

func AuthHttp() gin.HandlerFunc {
	return func(c *gin.Context) {
		// todo: 后端服务间调用（非api-gateway）鉴权
		return
	}
}

type UserService struct {
}

func NewUserService() *UserService {
	return &UserService{}
}

func generateToken(username, passwordDigest string) (token string, err error) {
	b, err := bcrypt.GenerateFromPassword([]byte(username+passwordDigest), 12)
	if err != nil {
		return
	}
	return string(b), nil
}

func (*UserService) UserLogin(ctx context.Context,req *service.UserRequest) (resp *service.UserDetailResponse,err error) {
	var user repository.User
	resp = new(service.UserDetailResponse)
	resp.Code = e.SUCCESS
	err = user.ShowUserInfo(req)
	if err != nil {
		resp.Code = e.ERROR
		return resp, err
	}
	resp.UserDetail = repository.BuildUser(user)
	token, err := generateToken(user.UserName, user.PasswordDigest)
	if err != nil {
		resp.Code = e.ErrorAuthToken
		return resp, err
	}
	err = cache.Set(token, "", 24*3600)
	if err != nil {
		resp.Code = e.ErrorAuthToken
		return resp, err
	}
	resp.Token = token
	return resp, nil
}

func (*UserService) UserRegister(ctx context.Context, req *service.UserRequest) (resp *service.UserDetailResponse,err error) {
	var user repository.User
	resp = new(service.UserDetailResponse)
	resp.Code = e.SUCCESS
	err = user.Create(req)
	if err != nil {
		resp.Code = e.ERROR
		return resp,err
	}
	resp.UserDetail = repository.BuildUser(user)
	return resp,nil
}

func (*UserService) UserLogout (ctx context.Context, req *empty.Empty) (resp *service.UserDetailResponse,err error) {
	resp = new(service.UserDetailResponse)
	resp.Code = e.SUCCESS
	md, ok := metadata.FromIncomingContext(ctx); if ok {
		token := md.Get("Authorization")
		_, _ = cache.Del(token[0])
	}
	return resp, nil
}

func (*UserService) AuthCheck (ctx context.Context, req *empty.Empty) (resp *service.UserDetailResponse,err error) {
	resp = new(service.UserDetailResponse)
	resp.Code = e.ErrorAuthCheckTokenTimeout
	md, ok := metadata.FromIncomingContext(ctx); if ok {
		token := ""
		v := md.Get("Authorization")
		if v != nil {
			token = v[0]
		}
		code := auth(token)
		if code != 0 {
			return resp, nil
		}
		resp.Token = token
	}
	resp.Code = e.SUCCESS
	return resp, nil
}

func (*UserService) UserInfo (ctx context.Context, req *service.UserModel) (resp *service.UserDetailResponse,err error) {
	var user repository.User
	resp = new(service.UserDetailResponse)
	resp.Code = e.SUCCESS
	err = user.GetUserInfo(req.UserID)
	if err != nil {
		resp.Code = e.ERROR
		return resp, err
	}
	resp.UserDetail = repository.BuildUser(user)
	return resp, nil
}
