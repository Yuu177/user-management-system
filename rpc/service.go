package rpc

import (
	"go/ast"
	"log"
	"reflect"
)

// method：方法本身
// ArgType：第一个参数的类型
// ReplyType：第二个参数的类型
type methodType struct {
	method    reflect.Method
	argType   reflect.Type
	replyType reflect.Type
}

// 该 Service 就是结构体和该结构体下所有方法的集合
// name 即映射的结构体的名称，比如 T，比如 WaitGroup；
// typ 是结构体的类型；
// rcvr 即结构体的实例本身，保留 rcvr 是因为在调用时需要 rcvr 作为第 0 个参数；
// methods 是 map 类型，存储映射的结构体的所有符合条件的方法。
type service struct {
	name string
	typ  reflect.Type
	rcvr reflect.Value
	// 上面是结构体的变量，下面是方法体。
	methods map[string]*methodType // 这里只是一个指针，所以需要初始化一块内存才能存取数据
}

// new 一个methodType 中的 Arg 类型（入参）的反射的值
func (m *methodType) newArg() reflect.Value {
	var arg reflect.Value
	// arg may be a pointer type, or a value type
	if m.argType.Kind() == reflect.Ptr {
		arg = reflect.New(m.argType.Elem())
	} else {
		arg = reflect.New(m.argType).Elem()
	}
	return arg
}

// new 一个methodType 中的 reply 类型（出参）的反射的值
func (m *methodType) newReply() reflect.Value {
	// reply must be a pointer type
	reply := reflect.New(m.replyType.Elem())
	switch m.replyType.Elem().Kind() {
	case reflect.Map:
		reply.Elem().Set(reflect.MakeMap(m.replyType.Elem()))
	case reflect.Slice:
		reply.Elem().Set(reflect.MakeSlice(m.replyType.Elem(), 0, 0))
	}
	return reply
}

func (s *service) call(m *methodType, arg reflect.Value, reply reflect.Value) error {
	f := m.method.Func
	returnValues := f.Call([]reflect.Value{s.rcvr, arg, reply})
	// 判断调用的函数的返回值 err 是否为空
	if errInter := returnValues[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}

func (s *service) registerMethods() {
	s.methods = make(map[string]*methodType)
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		mType := method.Type
		// 这里入参必须为 3，返回值必须为 1
		// 方法反射时为 3 个，第 0 个是自身
		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			continue
		}

		// if mType.NumIn() != 3 {
		// 	continue
		// }
		// 返回值必须只有一个 error
		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		argType, replyType := mType.In(1), mType.In(2)
		s.methods[method.Name] = &methodType{
			method:    method,
			argType:   argType,
			replyType: replyType,
		}
		log.Printf("rpc server: register %s.%s\n", s.name, method.Name)
	}
}

// 把结构体传进来解析该结构体的[方法类型]和[名称]以及它们的[入参]、[出参]
func newService(rcvr interface{}) *service {
	s := new(service)
	s.rcvr = reflect.ValueOf(rcvr)
	s.name = reflect.Indirect(s.rcvr).Type().Name()
	s.typ = reflect.TypeOf(rcvr)

	// 小写相当于private，是不能通过反射修改的，会报异常
	if !ast.IsExported(s.name) { // 判断是否结构体头是否为大写
		log.Fatalf("rpc server: %s is not a valid service name", s.name)
	}
	log.Println(s.name, s.rcvr, s.typ)
	s.registerMethods()
	return s
}
