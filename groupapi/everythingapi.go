package everything

import (
	"context"
	"net"
)

type IConnection interface {
	Start()                   //启动连接，让当前连接开始工作
	Stop()                    //停止连接，结束当前连接状态M
	Context() context.Context //返回ctx，用于用户自定义的go程获取连接退出状态

	GetTCPConnection() *net.TCPConn //从当前连接获取原始的socket TCPConn
	GetConnID() uint32              //获取当前连接ID
	RemoteAddr() net.Addr           //获取远程客户端地址信息

	SendMsg(msgID uint32, data []byte) error     //直接将Message数据发送数据给远程的TCP客户端(无缓冲)
	SendBuffMsg(msgID uint32, data []byte) error //直接将Message数据发送给远程的TCP客户端(有缓冲)

	SetProperty(key string, value interface{})   //设置链接属性
	GetProperty(key string) (interface{}, error) //获取链接属性
	RemoveProperty(key string)                   //移除链接属性
}

type IConnManager interface {
	Add(conn IConnection)                   //添加链接
	Remove(conn IConnection)                //删除连接
	Get(connID uint32) (IConnection, error) //利用ConnID获取链接
	Len() int                               //获取当前连接
	ClearConn()                             //删除并停止所有链接
}

type IDataPack interface {
	GetHeadLen() uint32                //获取包头长度方法
	Pack(msg IMessage) ([]byte, error) //封包方法
	Unpack([]byte) (IMessage, error)   //拆包方法
}

type IMessage interface {
	GetDataLen() uint32 //获取消息数据段长度
	GetMsgID() uint32   //获取消息ID
	GetData() []byte    //获取消息内容

	SetMsgID(uint32)   //设计消息ID
	SetData([]byte)    //设计消息内容
	SetDataLen(uint32) //设置消息数据段长度
}

type IMsgHandle interface {
	DoMsgHandler(request IRequest)          //马上以非阻塞方式处理消息
	AddRouter(msgID uint32, router IRouter) //为消息添加具体的处理逻辑
	StartWorkerPool()                       //启动worker工作池
	SendMsgToTaskQueue(request IRequest)    //将消息交给TaskQueue,由worker进行处理
}

type Packet interface {
	Unpack(binaryData []byte) (IMessage, error)
	Pack(msg IMessage) ([]byte, error)
	GetHeadLen() uint32
}

type IRequest interface {
	GetConnection() IConnection //获取请求连接信息
	GetData() []byte            //获取请求消息的数据
	GetMsgID() uint32           //获取请求的消息ID
}

type IRouter interface {
	PreHandle(request IRequest)  //在处理conn业务之前的钩子方法
	Handle(request IRequest)     //处理conn业务的方法
	PostHandle(request IRequest) //处理conn业务之后的钩子方法
}

type IServer interface {
	Start()                                 //启动服务器方法
	Stop()                                  //停止服务器方法
	Serve()                                 //开启业务服务方法
	AddRouter(msgID uint32, router IRouter) //路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用
	GetConnMgr() IConnManager               //得到链接管理
	SetOnConnStart(func(IConnection))       //设置该Server的连接创建时Hook函数
	SetOnConnStop(func(IConnection))        //设置该Server的连接断开时的Hook函数
	CallOnConnStart(conn IConnection)       //调用连接OnConnStart Hook函数
	CallOnConnStop(conn IConnection)        //调用连接OnConnStop Hook函数
	Packet() Packet
}
