package controllers

import (
	"fmt"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/blinkbean/dingtalk"
	"time"
)

func Notice(resourceKind, resourceName *string) error {
	tokens := "f4de0ca30038d12d46d054dfc4ce3fea0624664a2611398a0f225adaa6b51d59"
	secret := "SEC687aefffd3bbda62b182fb214909325390be7fd3f966ce536843d587253236f9"
	d := dingtalk.InitDingTalkWithSecret(tokens, secret)
	msg := "yaml has edit"
	mdmsg := []string{
		"### Kubernetes 资源警告信息",
		fmt.Sprintf("- 时间：%v", time.Now().Format("2006-01-02 15:04")),
		fmt.Sprintf("- Kind：%s", tea.StringValue(resourceKind)),
		fmt.Sprintf("- Name：%s", tea.StringValue(resourceName)),
		fmt.Sprintf("- Message: %s", msg),
	}
	mobiles := []string{"刘向标"}
	err := d.SendMarkDownMessageBySlice("test1", mdmsg, dingtalk.WithAtAll(), dingtalk.WithAtMobiles(mobiles))
	if err != nil {
		return err
	}
	return nil
}
