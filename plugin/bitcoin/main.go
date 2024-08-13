package bitcoin

import (
	"encoding/json"
	"fmt"
	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/file"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"net/http"
	"os"
	"time"
)

type Data struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	Quote  struct {
		CNY struct {
			Price       float64   `json:"price"`
			LastUpdated time.Time `json:"last_updated"`
		} `json:"CNY"`
	} `json:"quote"`
}

type Response struct {
	Status struct {
		Timestamp    time.Time   `json:"timestamp"`
		ErrorCode    int         `json:"error_code"`
		ErrorMessage interface{} `json:"error_message"`
		Elapsed      int         `json:"elapsed"`
		CreditCount  int         `json:"credit_count"`
		Notice       interface{} `json:"notice"`
	} `json:"status"`
	Data struct {
		BTC Data `json:"1"`
		ETH Data `json:"1027"`
	} `json:"data"`
}

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Help:              "- 查询比特币\n- 设置查询KEY xxxx",
		Brief:             "查询比特币和以太币价格",
		PrivateDataFolder: "Bitcoin",
	}).ApplySingle(ctxext.DefaultSingle)

	engine.OnRegex(`^设置查询KEY\s*(.*)$`, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		err := key.set(ctx.State["regex_matched"].([]string)[1])
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("设置成功"))
	})

	engine.OnFullMatch("查询比特币").SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		url := "https://pro-api.coinmarketcap.com/v2/cryptocurrency/quotes/latest?convert=CNY&id=1%2C1027"
		method := "GET"
		client := &http.Client{}
		req, err := http.NewRequest(method, url, nil)

		if err != nil {
			fmt.Println(err)
			return
		}
		req.Header.Add("X-CMC_PRO_API_KEY", key.k)
		res, err := client.Do(req)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		defer res.Body.Close()
		var response Response
		err = json.NewDecoder(res.Body).Decode(&response)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		btcPrice := response.Data.BTC.Quote.CNY.Price
		btcDate := response.Data.BTC.Quote.CNY.LastUpdated.Format("2006-01-02 15:04:05")
		ethPrice := response.Data.ETH.Quote.CNY.Price
		ethDate := response.Data.ETH.Quote.CNY.LastUpdated.Format("2006-01-02 15:04:05")
		ctx.SendChain(message.Text(fmt.Sprintf("BTC价格：%.10f,更新时间：%s\nETH价格：%.10f,更新时间%s", btcPrice, btcDate, ethPrice, ethDate)))
	})
}

var key = newApiKeyStore("./data/Bitcoin/bitcoin.key")

type apiKeyStore struct {
	k string
	p string
}

func newApiKeyStore(p string) (s apiKeyStore) {
	s.p = p
	if file.IsExist(p) {
		data, err := os.ReadFile(p)
		if err == nil {
			s.k = binary.BytesToString(data)
		}
	}
	return
}

func (s *apiKeyStore) set(k string) error {
	s.k = k
	return os.WriteFile(s.p, binary.StringToBytes(k), 0644)
}
