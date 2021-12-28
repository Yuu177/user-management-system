package rpc

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
)

// Call represents an active RPC.
type Call struct {
	ServiceMethod string      // format "<service>.<method>"
	Arg           interface{} // arguments to the function
	Reply         interface{} // reply from the function
	Error         error       // if error occurs, it will be set
}

// RPCClient rpc客户端 包含一个连接池(连接 rpc 服务器).
type RPCClient struct {
	pool chan net.Conn
}

// Client 创建 connections 个 tcp 连接, 连接到 address 中，并且将连接保存到连接池作为返回值返回
func Client(connections int, address string) (RPCClient, error) {
	// tcpAddr, err := net.ResolveTCPAddr("tcp4", address)
	// if err != nil {
	// 	return RPCClient{}, err
	// }
	log.Println("tpy client start...")
	//创建 connections 个连接，并将其保存到 pool 连接池中
	pool := make(chan net.Conn, connections)
	for i := 0; i < connections; i++ {
		// laddr 本地地址默认.
		// conn, err := net.DialTCP("tcp4", nil, tcpAddr)
		conn, err := net.Dial("tcp", address)
		if err != nil {
			return RPCClient{}, errors.New("rpc: init client failed")
		}
		log.Println("tpy: ", conn)
		pool <- conn
	}

	return RPCClient{pool: pool}, nil
}

// call 真正 rpc 调用逻辑，使用rpc调用函数 serviceMethod(arg, &reply)
func (r *RPCClient) call(serviceMethod string, arg interface{}, reply interface{}) (err error) {
	//从连接池获取一个空闲连接.
	conn := r.getConn()
	defer r.releaseConn(conn)

	cc := NewGobCodec(conn)
	var h = &Header{
		ServiceMethod: serviceMethod,
		Error:         "",
	}
	// var body interface{}

	// 将数据编码并发送到rpc服务器，然后等待接收
	if err = cc.Write(h, arg); err != nil {
		return err
	}

	log.Println("tpy read header")
	// 阻塞接收
	if err = cc.ReadHeader(h); err != nil {
		return err
	}
	if h.Error != "" {
		return errors.New(h.Error)
	}

	log.Println("tpy read body")

	err = cc.ReadBody(reply)
	if err != nil && err != io.EOF {
		return err
	}
	log.Println("call end")
	return nil
}

func receive(call *Call, cc Codec) error {
	var err error
	var h Header
	if err = cc.ReadHeader(&h); err != nil {
		return err
	}

	switch {
	case h.Error != "":
		call.Error = fmt.Errorf(h.Error)
		err = cc.ReadBody(nil)
	default:
		err = cc.ReadBody(call.Reply)
		if err != nil {
			call.Error = errors.New("reading body " + err.Error())
		}
	}
	return err

}

func send(call *Call, cc Codec) error {
	// prepare request header
	var h = &Header{
		ServiceMethod: call.ServiceMethod,
		Error:         "",
	}

	// encode and send the request
	if err := cc.Write(h, call.Arg); err != nil {
		return err
	}
	return nil
}

// func (r *RPCClient) Call(serviceMethod string, arg, reply interface{}) error {
// 	conn := r.getConn()
// 	defer r.releaseConn(conn)
// 	cc := NewGobCodec(conn)
// 	call := &Call{
// 		ServiceMethod: serviceMethod,
// 		Arg:           arg,
// 		Reply:         reply,
// 	}
// 	send(call, cc)
// 	receive(call, cc)
// 	return nil
// }

// getConn 从 RPCClient 连接池中随机取出一个空闲连接，如果没有，那将会阻塞。
func (r *RPCClient) getConn() (conn net.Conn) {
	log.Println(len(r.pool))

	select {
	case conn := <-r.pool:
		// log.Println("get 2")
		return conn
	}
}

// releaseConn 将连接 conn 重新写入到连接池中。
func (r *RPCClient) releaseConn(conn net.Conn) {
	r.pool <- conn
	// log.Println("release 1")
	// select {
	// case r.pool <- conn:
	// 	log.Println("release 1")

	// 	// TODO 返回空值怎么可行？
	// 	return
	// }
}

// Call 对外提供方法，调用服务端方法，reply 必须为指针类型，保存返回结果数据.
func (r *RPCClient) Call(serviceMethod string, arg interface{}, reply interface{}) error {
	return r.call(serviceMethod, arg, reply)
}
