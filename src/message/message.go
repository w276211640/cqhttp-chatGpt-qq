// Package message
// @Description
// @Author root_wang
// @Date 2022/12/10 17:58
package message

import (
	"cqhttp-client/src/log"
	"cqhttp-client/src/utils"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// CQTYPE CQCode中CQ的类型
type CQTYPE string

func (C CQTYPE) String() string {
	return string(C)
}

const (
	AT    CQTYPE = "at"
	REPLY CQTYPE = "reply"
	IMAGE CQTYPE = "image"
)

type CQKEY string

func (C CQKEY) String() string {
	return string(C)
}

const (
	CQ CQKEY = "CQ"
	// QQ reply
	QQ CQKEY = "qq"
	ID CQKEY = "id"
	// FILE image
	FILE CQKEY = "file"
)

// CQCode 包含一个CQType和一系列键值对 不能确定有哪些键值对采取懒加载
type CQCode struct {
	rawMessage string
	keyValue   map[CQKEY]string
	cqtype     CQTYPE
}

func NewCQCode(rawMessage string, cqtype CQTYPE) *CQCode {
	return &CQCode{
		rawMessage: rawMessage,
		keyValue:   make(map[CQKEY]string),
		cqtype:     cqtype,
	}
}

func (c *CQCode) String() string {
	start := fmt.Sprintf("[CQ:%s,", c.cqtype)
	end := fmt.Sprintf("]%s", c.rawMessage)

	var data strings.Builder
	for key, value := range c.keyValue {
		if key == CQ {
			continue
		}
		str := utils.Any2string(value)
		data.WriteString(fmt.Sprintf("%s=%s,", key, str))
	}
	str := strings.TrimRight(data.String(), ",")
	return fmt.Sprintf("%s%s%s", start, str, end)
}

func (c *CQCode) ParseKey(keys ...CQKEY) {
	for _, key := range keys {
		if key == CQ {
			typeReg := regexp.MustCompile(`\[CQ:(\w+),`)
			matches := typeReg.FindStringSubmatch(c.rawMessage)
			if matches != nil {
				c.cqtype = CQTYPE(matches[1])
			}
			continue
		}
		keyStr := fmt.Sprintf("%s", key)
		reg := keyStr + `=(.*)[\],]{1}`
		keyReg := regexp.MustCompile(reg)
		matches := keyReg.FindStringSubmatch(c.rawMessage)
		if matches != nil {
			c.keyValue[key] = matches[1]
		}
	}
}

func (c *CQCode) ValueByKey(key CQKEY) string {
	if value, ok := c.keyValue[key]; ok {
		return value
	}
	return ""
}

func (c *CQCode) SetKeyValue(keys []CQKEY, values ...interface{}) {
	if len(keys) != len(values) {
		panic(fmt.Sprintf("must set the same numbers of key and value "))
	}

	defer func() {
		if err := recover(); err != nil {
			log.Errorf("set cqcode failed: %v", err)
		}
	}()
	for index, key := range keys {
		c.keyValue[key] = utils.Any2string(values[index])
	}
}

func (c *CQCode) SetType(t CQTYPE) {
	c.cqtype = t
}

func (c *CQCode) IsAt() bool {
	if c.ValueByKey(QQ) == "" {
		c.ParseKey(QQ)
	}
	if c.cqtype == AT && c.ValueByKey(QQ) == strconv.FormatInt(BotQQ, 10) {
		return true
	}
	return false
}

// CQMessage 包含了CQCode和消息
type CQMessage struct {
	cqCode *CQCode
	msg    string
}

func (c *CQMessage) CQCode() *CQCode {
	return c.cqCode
}

func (c *CQMessage) Message() string {
	return c.msg
}

func (c *CQMessage) IsEmpty() bool {
	return c.msg == "" || strings.Trim(c.msg, " ") == ""
}

// RawMessage 原始信息主要包含了CQCode
type RawMessage string

func (m RawMessage) String() string {
	return string(m)
}

func (m RawMessage) IsPlainMessage() bool {
	if strings.HasPrefix(fmt.Sprintf("%s", m), "[CQ:at") {
		return false
	}
	return true
}

func (m RawMessage) IsEmpty() bool {
	return utils.StringEqual("", m)
}

func (m RawMessage) ToCQCode() (cqMsg *CQMessage, err error) {
	cqMsg = new(CQMessage)
	reg := `(\[.*\])\s(.+)`
	cqReg := regexp.MustCompile(reg)
	matches := cqReg.FindStringSubmatch(string(m))
	if matches == nil {
		return nil, log.ErrorInside("未能捕获到消息")
	}

	cqMsg.msg = matches[2]
	if cqMsg.IsEmpty() {
		return nil, log.ErrorInside("不能发生空的消息")
	}

	cqMsg.cqCode = &CQCode{
		rawMessage: matches[1],
		keyValue:   make(map[CQKEY]string),
	}

	// 初始化CQMsg必须指明CQ类型
	cqMsg.cqCode.ParseKey(CQ)
	return
}

// Sender 发送消息的发送者
type Sender struct {
	Age      int64  `json:"age"`
	Area     string `json:"area"`
	Card     string `json:"card"`
	Level    string `json:"level"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
	Sex      string `json:"sex"`
	Title    string `json:"title"`
	UserId   int64  `json:"user_id"`
}

type ReceiveMessage struct {
	PostType    string     `json:"post_type"`
	MessageType string     `json:"message_type"`
	Time        int64      `json:"time"`
	SelfId      int64      `json:"self_id"`
	SubType     string     `json:"sub_type"`
	UserId      int64      `json:"user_id"`
	MessageId   int64      `json:"message_id"`
	Font        int64      `json:"font"`
	GroupId     int64      `json:"group_id"`
	MessageSeq  int64      `json:"message_seq"`
	RawMessage  RawMessage `json:"raw_message"`
	Message     string     `json:"message"`
	Sender      Sender     `json:"sender"`
}

func IsRequireImage(msg string) (prompt string, ok bool) {
	prompt = msg
	ok = false
	// p&amp;
	imageReg := regexp.MustCompile(`p&amp;:(.+)`)
	matches := imageReg.FindStringSubmatch(msg)
	if matches != nil {
		prompt = matches[1]
		ok = true
	}
	return
}
