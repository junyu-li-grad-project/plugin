package internal

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/victor-leee/earth"
	earth_gen "github.com/victor-leee/earth/github.com/victor-leee/earth"
	"github.com/victor-leee/plugin/github.com/victor-leee/plugin"
	side_car "github.com/victor-leee/plugin/github.com/victor-leee/side-car"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	network     = "unix"
	path        = "/tmp/sc.sock"
	concurrency = 10
)

func SetServerConnections() {
	connList, err := Init(network, path, concurrency)
	if err != nil {
		panic(err)
	}
	for _, serverConn := range connList {
		go waitMsg(serverConn)
	}

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	logrus.Info("start cleaning up")
	logrus.Info("cleaning up finished")
}

func waitMsg(conn net.Conn) {
	for {
		req, err := earth.FromReader(conn, blockRead)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}
			logrus.Errorf("[waitMsg] read failed: %v", err)
			continue
		}
		logrus.Infof("[waitMsg] got request: %+v", req)
		rpcReq := &plugin.UnaryRPCRequest{}
		if err = proto.Unmarshal(req.Body, rpcReq); err != nil {
			logrus.Errorf("[waitMsg] unmarshal message failed: %v", err)
			continue
		}
	}
}

func Init(network, path string, respListSize int) ([]net.Conn, error) {
	// there are some connections that are used to receive responses
	// such connections should be established first
	responseConnList := make([]net.Conn, respListSize)
	for i := 0; i < respListSize; i++ {
		respConn, err := net.Dial(network, path)
		if err != nil {
			return nil, err
		}
		responseConnList[i] = respConn
	}

	wg := sync.WaitGroup{}

	for _, conn := range responseConnList {
		wg.Add(1)
		go func() {
			if _, err := earth.FromProtoMessage(nil, &earth_gen.Header{
				MessageType: earth_gen.Header_SET_USAGE,
			}).Write(conn); err != nil {
				logrus.Errorf("[Init] concurrently init failed: %v", err)
				return
			}
			resp, err := earth.FromReader(conn, blockRead)
			if err != nil {
				logrus.Errorf("[Init] concurrently respond failed: %v", err)
				return
			}

			baseResponse := &side_car.BaseResponse{}
			if err = proto.Unmarshal(resp.Body, baseResponse); err != nil {
				logrus.Errorf("[Init] concurrently unmarshal failed: %v", err)
				return
			}
			if baseResponse.Code != side_car.BaseResponse_CODE_SUCCESS {
				logrus.Errorf("[Init] concurrently init result failed: %v", baseResponse.Code)
				return
			}
			logrus.Infof("[Init] concurrently init one connection success")
			wg.Done()
		}()

		done := make(chan struct{})
		go func() {
			wg.Wait()
			<-done
		}()

		select {
		case <-done:
			logrus.Info("[Init] ok")
		case <-time.After(time.Second * 3):
			logrus.Errorf("[Init] time limit exceeded")
		}
	}

	return responseConnList, nil
}

func blockRead(reader io.Reader, size uint64) ([]byte, error) {
	b := make([]byte, size)
	already := 0
	inc := 0
	var err error
	for uint64(already) < size {
		if inc, err = reader.Read(b[already:]); err != nil {
			return nil, err
		}
		already += inc
	}

	return b, nil
}
