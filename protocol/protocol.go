package protocol

const (
	Success                 int = 0
	UserNameOrPasswordNull  int = 1
	UserCreateFailed        int = 2
	UserNameOrPasswordError int = 3
	LoginFailed             int = 4
	TokenCheckFailed        int = 5
	DataIsNil               int = 6
	GetProfileFailed        int = 7
	UserNotExist            int = 8
	UpdateFailed            int = 9
)

// 注册请求.
type ReqSignUp struct {
	UserName string // 用户名, 不为空
	Password string // 密码, 不为空
	NickName string // 昵称
}

// 注册返回.
type RespSignUp struct {
	Ret int // 结果码 0:成功 1:用户名或密码为空 2:用户名重复或创建失败
}

// 登录请求.
type ReqLogin struct {
	UserName string // 用户名, 不为空
	Password string // 密码, 不为空
}

// 登录返回.
type RespLogin struct {
	Ret   int    // 结果码 0:成功 1:用户名或密码错误 2:登录失败
	Token string // token
}

// 获取信息请求.
type ReqGetProfile struct {
	UserName string
	Token    string
}

// 获取信息返回.
type RespGetProfile struct {
	Ret      int    // 结果码 0:成功 1:token校验失败 2:数据为空 3:获取失败
	UserName string // 用户名，不为空
	NickName string // 昵称
	PicName  string // 头像(路径信息)
}

// 更新用户头像请求.
type ReqUpdateProfilePic struct {
	UserName string // 用户名, 不为空
	FileName string // 头像文件名
	Token    string // token
}

// 更新用户头像返回.
type RespUpdateProfilePic struct {
	Ret int // 结果码 0:成功 1:token校验失败 2:用户不存在 3:更新失败
}

// 更新用户昵称请求.
type ReqUpdateNickName struct {
	UserName string // 用户名, 不为空
	NickName string // 昵称
	Token    string // token
}

// 更新用户头像返回.
type RespUpdateNickName struct {
	Ret int // 结果码 0:成功 1:token校验失败 2:用户不存在 3:更新失败
}

// 用户表，用来保存账号密码
type User struct {
	UserName string `gorm:"column:user_name;primary_key;type:varchar(255);"` // 设置自动生成表的表名
	Password string `gorm:"column:password;not null;type:char(32);"`
}

func (User) TableName() string {
	return "user"
}

// 用户信息表，用来保存用户除了密码的其他信息
type UserProfile struct {
	UserName string `gorm:"column:user_name;primary_key;type:varchar(255);"` // 设置自动生成表的表名
	NickName string `gorm:"column:nick_name;type:varchar(255);"`
	PicName  string `gorm:"column:pic_name;type:varchar(255);"` // 用户头像文件名名称
}

// gorm 设置表名。如果不重写这个接口就是默认在后面加 s
func (UserProfile) TableName() string {
	return "user_profile"
}
