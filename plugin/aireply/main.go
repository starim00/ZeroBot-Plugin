// Package aireply AI 回复
package aireply

import (
	sql "github.com/FloatTech/sqlite"
	"github.com/FloatTech/zbputils/ctxext"
	"os"
	"regexp"
	"strings"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type UserPrompt struct {
	UserId int64  `json:"user_id"`
	Prompt string `json:"prompt"`
}

var replmd = replymode([]string{"婧枫", "沫沫", "青云客", "小爱", "ChatGPT", "DeepSeek"})

var db = sql.Sqlite{}

var ttsmd = newttsmode()

func init() { // 插件主体
	ent := control.Register("tts", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: true,
		Brief:            "人工智能语音回复",
		Help: "- @Bot 任意文本(任意一句话回复)\n" +
			"- 设置语音模式[百度/TTSCN/桑帛云] 数字(百度/TTSCN说话人/桑帛云)\n" +
			"- 设置默认语音模式[百度/TTSCN/桑帛云] 数字(百度/TTSCN说话人/桑帛云)\n" +
			"- 恢复成默认语音模式\n" +
			"- 设置语音回复模式[沫沫|婧枫|青云客|小爱|ChatGPT]\n" +
			"- 设置百度语音 api id xxxxxx secret xxxxxx (请自行获得)\n" +
			"\n当前适用的TTSCN人物含有以下(以数字顺序代表): \n" + list(ttscnspeakers[:], 5),
		PrivateDataFolder: "tts",
	})

	enr := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "人工智能回复",
		Help:              "- @Bot 任意文本(任意一句话回复)\n- 设置文字回复模式[婧枫|沫沫|青云客|小爱|ChatGPT]\n- 设置 ChatGPT api key xxx",
		PrivateDataFolder: "aireply",
	})

	db = sql.New(enr.DataFolder() + "prompt.db")
	err := db.Open(time.Hour)
	if err != nil {
		panic(err)
	}
	err = db.Create("user_prompt", &UserPrompt{})
	if err != nil {
		panic(err)
	}

	enr.OnMessage(zero.OnlyToMe).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			aireply := replmd.getReplyMode(ctx)
			reply := message.ParseMessageFromString(aireply.Talk(ctx.Event.UserID, ctx.ExtractPlainText(), zero.BotConfig.NickName[0]))
			// 回复
			time.Sleep(time.Second * 1)
			reply = append(reply, message.Reply(ctx.Event.MessageID))
			ctx.Send(reply)
		})
	setReplyMode := func(ctx *zero.Ctx) {
		param := ctx.State["args"].(string)
		err := replmd.setReplyMode(ctx, param)
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("成功"))
	}
	enr.OnPrefix("设置文字回复模式", zero.AdminPermission).SetBlock(true).Handle(setReplyMode)
	enr.OnRegex(`^设置\s*ChatGPT\s*api\s*key\s*(.*)$`, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		err := ཆཏ.set(ctx.State["regex_matched"].([]string)[1])
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("设置成功"))
	})
	enr.OnRegex(`^设置Prompt\s*(.*)$`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		prompt := ctx.State["regex_matched"].([]string)[1]
		if len(prompt) > 1000 {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("prompt过长，请重新设置"))
			return
		}
		db.Insert("user_prompt", &UserPrompt{
			UserId: ctx.Event.UserID,
			Prompt: prompt,
		})
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("设置成功，Prompt为"+prompt))
	})

	endpre := regexp.MustCompile(`\pP$`)
	ttscachedir := ent.DataFolder() + "cache/"
	_ = os.RemoveAll(ttscachedir)
	err = os.MkdirAll(ttscachedir, 0755)
	if err != nil {
		panic(err)
	}
	ent.OnMessage(zero.OnlyToMe).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			msg := ctx.ExtractPlainText()
			// 获取回复模式
			r := replmd.getReplyMode(ctx)
			// 获取回复的文本
			reply := message.ParseMessageFromString(r.TalkPlain(ctx.Event.UserID, msg, zero.BotConfig.NickName[0]))
			// 过滤掉文字消息
			filterMsg := make([]message.Segment, 0, len(reply))
			sb := strings.Builder{}
			for _, v := range reply {
				if v.Type != "text" {
					filterMsg = append(filterMsg, v)
				} else {
					sb.WriteString(v.Data["text"])
				}
			}
			// 纯文本
			plainReply := sb.String()
			plainReply = strings.ReplaceAll(plainReply, "\n", "")
			// 获取语音
			speaker, err := ttsmd.getSoundMode(ctx)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			rec, err := speaker.Speak(ctx.Event.UserID, func() string {
				if !endpre.MatchString(plainReply) {
					return plainReply + "。"
				}
				return plainReply
			})
			// 发送前面的图片
			if len(filterMsg) != 0 {
				filterMsg = append(filterMsg, message.Reply(ctx.Event.MessageID))
				ctx.Send(filterMsg)
			}
			if err != nil {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(plainReply))
				return
			}
			// 发送语音
			if id := ctx.SendChain(message.Record(rec)); id.ID() == 0 {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(plainReply))
			}
		})
}
