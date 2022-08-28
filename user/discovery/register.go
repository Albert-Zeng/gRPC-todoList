package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Register struct {
	EtcdAddrs   []string
	DialTimeout int

	closeCh     chan struct{}
	leasesID    clientv3.LeaseID
	keepAliveCh <-chan *clientv3.LeaseKeepAliveResponse

	srvInfo Server
	srvTTL  int64
	cli     *clientv3.Client
	logger  *logrus.Logger
}

// NewRegister create a register based on etcd
func NewRegister(etcdAddrs []string, logger *logrus.Logger) *Register {
	return &Register{
		EtcdAddrs:   etcdAddrs,
		DialTimeout: 3,
		logger:      logger,
	}
}

// Register a service
func (r *Register) Register(srvInfo Server, ttl int64) (chan<- struct{}, error) {
	var err error

	if strings.Split(srvInfo.Addr, ":")[0] == "" {
		return nil, errors.New("invalid ip address")
	}

	if r.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   r.EtcdAddrs,
		DialTimeout: time.Duration(r.DialTimeout) * time.Second,
	}); err != nil {
		return nil, err
	}

	r.srvInfo = srvInfo
	r.srvTTL = ttl

	if err = r.register(); err != nil {
		return nil, err
	}

	r.closeCh = make(chan struct{})

	go r.keepAlive()

	return r.closeCh, nil
}

func (r *Register) register() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.DialTimeout)*time.Second)
	defer cancel()

	leaseResp, err := r.cli.Grant(ctx, r.srvTTL)
	if err != nil {
		return err
	}

	r.leasesID = leaseResp.ID

	if r.keepAliveCh, err = r.cli.KeepAlive(context.Background(), r.leasesID); err != nil {
		return err
	}

	data, err := json.Marshal(r.srvInfo)
	if err != nil {
		return err
	}

	_, err = r.cli.Put(context.Background(), BuildRegisterPath(r.srvInfo), string(data), clientv3.WithLease(r.leasesID))

	return err
}

// Stop stop register
func (r *Register) Stop() {
	r.closeCh <- struct{}{}
}

// unregister 删除节点
func (r *Register) unregister() error {
	_, err := r.cli.Delete(context.Background(), BuildRegisterPath(r.srvInfo))
	return err
}

func (r *Register) keepAlive() {
	ticker := time.NewTicker(time.Duration(r.srvTTL) * time.Second)

	for {
		select {
		case <-r.closeCh:
			if err := r.unregister(); err != nil {
				r.logger.Error("unregister failed, error: ", err)
			}

			if _, err := r.cli.Revoke(context.Background(), r.leasesID); err != nil {
				r.logger.Error("revoke failed, error: ", err)
			}
		case res := <-r.keepAliveCh:
			if res == nil {
				if err := r.register(); err != nil {
					r.logger.Error("register failed, error: ", err)
				}
			}
		case <-ticker.C:
			if r.keepAliveCh == nil {
				if err := r.register(); err != nil {
					r.logger.Error("register failed, error: ", err)
				}
			}
		}
	}
}

func (r *Register) UpdateHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		weightstr := req.URL.Query().Get("weight")
		weight, err := strconv.Atoi(weightstr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		var update = func() error {
			r.srvInfo.Weight = int64(weight)
			data, err := json.Marshal(r.srvInfo)
			if err != nil {
				return err
			}

			_, err = r.cli.Put(context.Background(), BuildRegisterPath(r.srvInfo), string(data), clientv3.WithLease(r.leasesID))
			return err
		}

		if err := update(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		_, _ = w.Write([]byte("update server weight success"))
	})
}

func (r *Register) GetServerInfo() (Server, error) {
	resp, err := r.cli.Get(context.Background(), BuildRegisterPath(r.srvInfo))
	if err != nil {
		return r.srvInfo, err
	}

	server := Server{}
	if resp.Count >= 1 {
		if err := json.Unmarshal(resp.Kvs[0].Value, &server); err != nil {
			return server, err
		}
	}

	return server, err
}

func (r *Register) GetTargetServer(key string) (ser *Server, err error) {
	if r.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   r.EtcdAddrs,
		DialTimeout: time.Duration(r.DialTimeout) * time.Second,
	}); err != nil {
		return
	}

	ctx := context.Background()
	res, err := r.cli.Get(ctx, key, clientv3.WithPrefix()); if err != nil {
		return
	}
	for _, kv := range res.Kvs {
		ser = &Server{}
		err = json.Unmarshal(kv.Value, ser); if err != nil {
			r.logger.Warn("etcd response.Value Unmarshal error, err(%v)", err)
			continue
		}
		return
	}
	return
}

// 相当于 resolver 的简易实现
func (r *Register) WatchEtcdServerInfo(serverMap map[string]*Server, key string) (err error) {
	// serverMap version/name/protoc(http/grpc):*Server
	if r.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   r.EtcdAddrs,
		DialTimeout: time.Duration(r.DialTimeout) * time.Second,
	}); err != nil {
		return
	}

	ctx := context.Background()
	res, err := r.cli.Get(ctx, key, clientv3.WithPrefix()); if err != nil {
		return
	}
	for _, kv := range res.Kvs {
		v := &Server{}
		err = json.Unmarshal(kv.Value, v); if err != nil {
			r.logger.Warn("etcd response.Value Unmarshal error, err(%v)", err)
			continue
		}
		split := strings.Split(string(kv.Key), "/")
		k := strings.Join(split[:len(split)-1], "/")
		serverMap[k] = v
	}

	// 监听etcd变化
	go func() {
		watchStartRevision := res.Header.Revision + 1
		r.logger.Info("watch etcd keys beginning with revision: ", watchStartRevision)
		watchRespChan := r.cli.Watch(ctx, key, clientv3.WithPrefix(), clientv3.WithRev(watchStartRevision))
		// 处理kv变化事件
		for watchResp := range watchRespChan {
			for _, event := range watchResp.Events {
				kv := event.Kv
				v := &Server{}
				if event.Kv.Value != nil {
					err = json.Unmarshal(kv.Value, v); if err != nil {
						r.logger.Warn("etcd response.Value Unmarshal error, value(%s), err(%v)", kv.Value, err)
						continue
					}
				}
				split := strings.Split(string(kv.Key), "/")
				k := strings.Join(split[:len(split)-1], "/")

				switch event.Type {
				case mvccpb.PUT:
					serverMap[k] = v
				case mvccpb.DELETE:
					delete(serverMap, k)
				}
			}
		}
	}()

	return
}
