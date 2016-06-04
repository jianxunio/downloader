package rabbitmq

import (
	"time"
	// "fmt"
	"downloader/config"
)

var nowFunc = time.Now // for testing

type ConnPool struct {
	// Dial is an application supplied function for creating and configuring a
	// connection
	Dial func(config.Config) (interface{}, error)
	//Dial func() (*autorc.Conn, error)

	// Maximum number of idle connections in the pool.
	MaxIdle int

	// Maximum number of connections allocated by the pool at a given time.
	// When zero, there is no limit on the number of connections in the pool.
	MaxActive int

	Cfg config.Config

	size   int
	closed bool
	active int

	idle chan interface{}
}

type idleConn struct {
	c interface{}
	t time.Time
}

// 批量生成连接，并把连接放到连接池channel里面
func (this *ConnPool) InitPool() error {
	this.idle = make(chan interface{}, this.MaxActive)
	for x := 0; x < this.MaxActive; x++ {
		//conn, err := this.Dial()
		// 这里返回DB类，而不是返回mysql.Conn，否则DB类的insert，update 这些Active Record类方法没法使用
		db, err := this.Dial(this.Cfg)
		//fmt.Println(" --- reflect --- ", reflect.TypeOf(db))
		if err != nil {
			return err
		}
		//this.idle <-conn
		this.idle <- idleConn{t: nowFunc(), c: db}
	}
	this.size = this.MaxActive
	return nil
}

// 从连接池里取出连接
func (this *ConnPool) Get() interface{} {
	// 如果空闲连接为空，初始化连接池
	if this.size == 0 {
		this.InitPool()
	}

	// 赋值一下好给下面回收和返回
	//conn := <-this.idle
	//idleConn
	ic := <-this.idle
	this.size -= 1
	// 这里要用 (idleConn) 把interface{} 类型转化为 idleConn 类型的，否则拿不到里面的属性t、c
	conn := ic.(idleConn).c
	//fmt.Println(conn.(*DB).conn)
	//fmt.Println(" --- reflect --- ", reflect.TypeOf(conn))
	// fmt.Println("connection time", ic.(idleConn).t, "size: ", this.size)

	// 使用完把连接回收到连接池里
	// defer this.Release(conn)

	// 因为channel是有锁的，所以就没必要借助sync.Mutex来进行读写锁定
	// container/list就需要锁住，不然并发就互抢出问题了
	return conn
}

// 回收连接到连接池
func (this *ConnPool) Release(conn interface{}) {
	//this.idle <-conn
	ic := idleConn{t: nowFunc(), c: conn}
	this.idle <- ic
	this.size += 1
	// fmt.Println("add conn: ", ic.t, "size: ", this.size)
}
