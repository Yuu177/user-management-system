package rpc

import (
	"errors"
	"io"
	"log"
	"net"
)

// rpc客户端 包含一个连接池(连接 rpc 服务器)
type RPCClient struct {
	pool chan net.Conn
}

// 创建 connections 个 tcp 连接, 连接到 address 中，并且将连接保存到连接池作为返回值返回
func Client(connections int, address string) (RPCClient, error) {
	log.Println("client conn pool init, please wait")
	// 创建 connections 个连接，并将其保存到 pool 连接池中
	pool := make(chan net.Conn, connections)
	for i := 0; i < connections; i++ {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			return RPCClient{}, errors.New("rpc: init client failed")
		}

		pool <- conn
	}
	log.Println("init rpc conn pool finish")
	return RPCClient{pool: pool}, nil
}

// 真正 rpc 调用逻辑，使用rpc调用函数 serviceMethod(arg, &reply)
func (r *RPCClient) call(serviceMethod string, arg interface{}, reply interface{}) (err error) {
	// 从连接池获取一个空闲连接.
	conn := r.getConn()
	defer r.releaseConn(conn) // TODO 服务端关闭连接后，客户端这个连接还能用吗？

	cc := NewGobCodec(conn)
	var h = &Header{
		ServiceMethod: serviceMethod,
		Error:         "",
	}

	// 将数据编码并发送到 rpc 服务器，然后等待接收
	if err = cc.Write(h, arg); err != nil {
		return err
	}

	// 接收
	if err = cc.ReadHeader(h); err != nil {
		return err
	}
	if h.Error != "" {
		return errors.New(h.Error)
	}

	err = cc.ReadBody(reply)
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

// 从 RPCClient 连接池中随机取出一个空闲连接，如果没有，那将会阻塞。
func (r *RPCClient) getConn() (conn net.Conn) {
	log.Println(len(r.pool))
	return <-r.pool
}

// 将连接 conn 重新写入到连接池中。
func (r *RPCClient) releaseConn(conn net.Conn) {
	r.pool <- conn
}

// 对外提供方法，调用服务端方法，reply 必须为指针类型，保存返回结果数据.
func (r *RPCClient) Call(serviceMethod string, arg interface{}, reply interface{}) error {
	return r.call(serviceMethod, arg, reply)
}
