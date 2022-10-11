package auth

import (
	bm "github.com/mapgoo-lab/atreus/pkg/net/http/blademaster"
)

type UserTitle struct {
	//用户名
	UserName string `json:"user_name"`
	//昵称
	NickName string `json:"nick_name"`
	//openId
	OpenId string `json:"open_id"`
	//UnionID
	UnionId string `json:"union_id"`
	//头像url
	AvatarUrl string `json:"avatar_url"`
	//性别 0：未知、1：男、2：女
	Gender int32 `json:"gender"`
	//年龄
	Age int32 `json:"age"`
	//电话号码
	Mobile string `json:"mobile"`
	//邮箱
	Email string `json:"email"`
	//国家
	Country string `json:"country"`
	//省份
	Province string `json:"province"`
	//城市
	City string `json:"city"`
	//用户身份 0-未知 1-消费者 2-经销商
	UserType int32 `json:"user_type"`
}

type Auth interface {
	//用户校验
	Auth(c *bm.Context)

	//获取token
	GetToken(userTitle *UserTitle) (string, error)
}
