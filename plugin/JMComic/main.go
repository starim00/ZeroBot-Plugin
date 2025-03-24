package JMComic

import (
	"encoding/json"
	"github.com/FloatTech/floatbox/file"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"net/http"
)

const (
	api = "http://127.0.0.1:8005"
)

type jmComic struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Url     string `json:"url"`
}

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "禁漫下载pdf",
		Help:             "- JM jmid",
		PublicDataFolder: "JMComic",
	}).ApplySingle(ctxext.DefaultSingle)

	engine.OnRegex(`^JM\s*(\d*)`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		jmid := ctx.State["regex_matched"].([]string)[1]
		req, err := http.NewRequest("GET", api+"/download/"+jmid, nil)
		client := &http.Client{}
		response, err := client.Do(req)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		defer response.Body.Close()
		var jmComic jmComic
		err = json.NewDecoder(response.Body).Decode(&jmComic)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		filePath := file.BOTPATH + "/" + engine.DataFolder() + jmid + ".pdf"
		err = file.DownloadTo(jmComic.Url, filePath)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.File("file:///"+filePath, jmid+".pdf"))
	})
}
