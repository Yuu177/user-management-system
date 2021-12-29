package rpc

import (
	"errors"
	"log"
	"net"
	"reflect"
	"strings"
)

// 保存多个注册了的结构体
type Server struct {
	serviceMap map[string]*service
}

// 读取客户端请求，解析并填充 request
// client.Call(Foo.sum, arg, reply)
type request struct {
	h     *Header // Foo.sum 连接的名称
	arg   reflect.Value
	reply reflect.Value
	mtype *methodType // 对应的 sum 方法
	svc   *service    // 对应的 Foo
}

func NewServer() *Server {
	return &Server{make(map[string]*service)}
}

func (s *Server) listen(addr string) (net.Listener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("network error: ", err)
	}
	return l, nil
}

func (s *Server) findService(serviceMethod string) (svc *service, mType *methodType, err error) {
	// 把 Foo.sum 切分为 Foo 和 sum
	pos := strings.LastIndex(serviceMethod, ".")
	if pos < 0 {
		err = errors.New("rpc server: service/method request err : " + serviceMethod)
		return
	}
	serviceName, methodName := serviceMethod[:pos], serviceMethod[pos+1:]
	svc, ok := s.serviceMap[serviceName]
	if !ok {
		err = errors.New("rpc server: can't find service " + serviceName)
		return
	}
	mType = svc.methods[methodName]
	return
}

// 读取 header 和 body
func (s *Server) readRequest(cc Codec) (*request, error) {
	var h Header
	err := cc.ReadHeader(&h)
	if err != nil {
		return nil, err
	}

	// 填充 request
	req := &request{h: &h}
	// 查询是否注册有该方法
	req.svc, req.mtype, err = s.findService(h.ServiceMethod)
	if err != nil {
		return nil, err
	}

	req.arg = req.mtype.newArg()     // 根据找到的方法的入参的类型 new 一个入参出来
	req.reply = req.mtype.newReply() // 同上

	// 不能直接修改反射的值，必须把它变成一个接口。
	argvi := req.arg.Interface()
	if req.arg.Type().Kind() != reflect.Ptr {
		argvi = req.arg.Addr().Interface()
	}

	// ReadBody 接受的是一个指针
	if err = cc.ReadBody(argvi); err != nil {
		log.Println("rpc server: read body err:", err)
		return req, err
	}
	return req, nil
}

var invalidRequest = struct{}{}

func (s *Server) handleRequest(cc Codec, req *request) error {
	// 通过request 我们知道了：结构体，方法，入参，出参
	err := req.svc.call(req.mtype, req.arg, req.reply)
	return err
}

func (s *Server) sendResponse(cc Codec, h *Header, body interface{}) {

	if err := cc.Write(h, body); err != nil {
		log.Println("rpc server: write response error:", err)
	}
}

func (s *Server) serveConn(conn net.Conn) {
	if conn == nil {
		log.Println("conn is nil")
		return
	}
	for {
		// 每次来新的连接就new 一个 codec，因为 codec 的 conn 不一样
		cc := NewGobCodec(conn)
		req, err := s.readRequest(cc)
		if err != nil {
			if req == nil {
				// cc.Close() // it's not possible to recover, so close the connection
				return
			}
			req.h.Error = err.Error()
			s.sendResponse(cc, req.h, invalidRequest)
			return
		}
		err = s.handleRequest(cc, req)
		if err != nil {
			// handle 返回失败的值
			req.h.Error = err.Error()
			s.sendResponse(cc, req.h, invalidRequest)
			return
		}
		s.sendResponse(cc, req.h, req.reply.Interface())
	}
}

func (s *Server) acceptAndServeConn(listener net.Listener) error {
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		defer conn.Close()
		go s.serveConn(conn)
	}
}

func (s *Server) ListenAndServe(addr string) error {
	// 监听
	listener, err := s.listen(addr)
	if err != nil {
		return err
	}
	// 阻塞接收连接，并处理请求
	if err = s.acceptAndServeConn(listener); err != nil {
		return err
	}
	return nil
}

func (s *Server) Register(rcvr interface{}) {
	svc := newService(rcvr)
	s.serviceMap[svc.name] = svc
}
