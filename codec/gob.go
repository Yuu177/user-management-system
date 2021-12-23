package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	conn io.ReadWriteCloser // socket 连接句柄。不带缓存
	buf  *bufio.Writer      // 缓冲区，enc 序列化会把数据存到该缓冲区（包装 conn）
	dec  *gob.Decoder
	enc  *gob.Encoder
}

// GobCodec 构造函数
func NewGobCodec(conn io.ReadWriteCloser) Codec {
	// 创基 conn 的缓冲区，我们最终的目的是忘 conn 里写入数据，但何时发送缓冲区的数据给客户端由底层的 socket 决定
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn), // 从 conn 中读取数据解码
		enc:  gob.NewEncoder(buf),  // 把数据编码后写到 buf 中
	}
}

var _ Codec = (*GobCodec)(nil)

func (c *GobCodec) Close() error {
	return c.conn.Close()
}

func (c *GobCodec) ReadHeader(h *Header) error {
	return c.dec.Decode(h)
}

func (c *GobCodec) ReadBody(body interface{}) error { // 从 conn 解码数据到 body 中
	return c.dec.Decode(body)
}

// 编码后发给连接的客户端。
// 这里只是往客户端 conn 写数据，何时发送数据给客户端由底层的 socket 决定。
func (c *GobCodec) Write(h *Header, body interface{}) (err error) {
	defer func() {
		_ = c.buf.Flush() // 把缓冲区的数据全部写入到 conn
		if err != nil {
			_ = c.Close()
		}
	}()

	if err := c.enc.Encode(h); err != nil {
		log.Println("rpc codec: gob error encoding header:", err)
		return err
	}

	if err := c.enc.Encode(body); err != nil {
		log.Println("rpc codec: gob error encoding body:", err)
		return err
	}

	return nil
}
