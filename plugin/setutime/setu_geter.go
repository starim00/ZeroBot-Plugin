// Package setutime 来份涩图
package setutime

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/FloatTech/AnimeAPI/pixiv"
	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/math"
	"github.com/FloatTech/floatbox/process"
	sql "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// Pools 图片缓冲池
type imgpool struct {
	db      sql.Sqlite
	dbmu    sync.RWMutex
	path    string
	max     int
	pool    map[string][]*message.Segment
	poolmu  sync.Mutex
	filling sync.Map
}

func (p *imgpool) List() (l []string) {
	var err error
	p.dbmu.RLock()
	defer p.dbmu.RUnlock()
	l, err = p.db.ListTables()
	if err != nil {
		l = []string{"涩图", "二次元", "风景", "车万"}
	}
	return l
}

var pool = &imgpool{
	path: pixiv.CacheDir,
	max:  10,
	pool: make(map[string][]*message.Segment),
}

func init() { // 插件主体
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "涩图",
		Help: "- 来份[涩图/二次元/风景/车万]\n" +
			"- 添加[涩图/二次元/风景/车万][P站图片ID]\n" +
			"- 删除[涩图/二次元/风景/车万][P站图片ID]\n" +
			"- >setu status",
		PublicDataFolder: "SetuTime",
	})

	getdb := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		// 如果数据库不存在则下载
		pool.db = sql.New(engine.DataFolder() + "SetuTime.db")
		_, _ = engine.GetLazyData("SetuTime.db", false)
		err := pool.db.Open(time.Hour)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		for _, imgtype := range pool.List() {
			if err := pool.db.Create(imgtype, &pixiv.Illust{}); err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return false
			}
		}
		return true
	})

	engine.OnRegex(`^来份(.+)$`, getdb, fcext.ValueInList(func(ctx *zero.Ctx) string { return ctx.State["regex_matched"].([]string)[1] }, pool)).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			var imgtype = ctx.State["regex_matched"].([]string)[1]
			// 补充池子
			go pool.fill(ctx, imgtype)
			// 等待首张图片就绪，避免固定等待十秒后误报失败。
			if pool.size(imgtype) == 0 {
				ctx.SendChain(message.Text("INFO: 正在填充弹药......"))
			}
			deadline := time.NewTimer(65 * time.Second)
			defer deadline.Stop()
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()
			img := pool.pop(imgtype)
			for img == nil {
				select {
				case <-deadline.C:
					ctx.SendChain(message.Text("ERROR: 等待填充，请稍后再试......"))
					return
				case <-ticker.C:
					img = pool.pop(imgtype)
				}
			}
			// 从缓冲池里抽一张
			m := message.Message{ctxext.FakeSenderForwardNode(ctx, *img)}
			if id := ctx.Send(m).ID(); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})

	engine.OnRegex(`^添加\s*([^0-9\s]+)\s*(\d+)$`, zero.SuperUserPermission, getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var (
				imgtype = ctx.State["regex_matched"].([]string)[1]
				id, _   = strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
			)
			err := pool.add(ctx, imgtype, id)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("成功向分类", imgtype, "添加图片", id))
		})

	engine.OnRegex(`^删除\s*([^0-9\s]+)\s*(\d+)$`, getdb, fcext.ValueInList(func(ctx *zero.Ctx) string { return ctx.State["regex_matched"].([]string)[1] }, pool), zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var (
				imgtype = ctx.State["regex_matched"].([]string)[1]
				id, _   = strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
			)
			// 查询数据库
			if err := pool.remove(imgtype, id); err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("删除成功"))
		})

	// 查询数据库涩图数量
	engine.OnFullMatch(">setu status", getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			state := []string{"[SetuTime]"}
			for _, imgtype := range pool.List() {
				pool.dbmu.RLock()
				num, err := pool.db.Count(imgtype)
				pool.dbmu.RUnlock()
				if err != nil {
					num = 0
				}
				state = append(state, "\n")
				state = append(state, imgtype)
				state = append(state, ": ")
				state = append(state, fmt.Sprintf("%d", num))
			}
			ctx.SendChain(message.Text(state))
		})
}

// size 返回缓冲池指定类型的现有大小
func (p *imgpool) size(imgtype string) int {
	p.poolmu.Lock()
	defer p.poolmu.Unlock()
	return len(p.pool[imgtype])
}

func (p *imgpool) push(ctx *zero.Ctx, imgtype string, illust *pixiv.Illust) {
	msg, err := p.image(illust)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	p.poolmu.Lock()
	if len(p.pool[imgtype]) < p.max {
		p.pool[imgtype] = append(p.pool[imgtype], &msg)
	}
	p.poolmu.Unlock()
}

func (p *imgpool) pop(imgtype string) (msg *message.Segment) {
	p.poolmu.Lock()
	defer p.poolmu.Unlock()
	if len(p.pool[imgtype]) == 0 {
		return
	}
	msg = p.pool[imgtype][0]
	p.pool[imgtype] = p.pool[imgtype][1:]
	return
}

// fill 补充池子
func (p *imgpool) fill(ctx *zero.Ctx, imgtype string) {
	if _, loaded := p.filling.LoadOrStore(imgtype, true); loaded {
		return
	}
	defer p.filling.Delete(imgtype)
	times := math.Min(p.max-p.size(imgtype), 2)
	for i := 0; i < times; i++ {
		illust := &pixiv.Illust{}
		// 查询出一张图片
		p.dbmu.RLock()
		err := p.db.Pick(imgtype, illust)
		p.dbmu.RUnlock()
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			continue
		}
		// 向缓冲池添加一张图片
		p.push(ctx, imgtype, illust)
		process.SleepAbout1sTo2s()
	}
}

func (p *imgpool) add(ctx *zero.Ctx, imgtype string, id int64) error {
	ctx.SendChain(message.Text("少女祈祷中......"))
	// 查询P站插图信息
	illust, err := pixivWorks(id)
	if err != nil {
		return err
	}
	img, err := p.image(illust)
	if err != nil {
		return err
	}
	if ctx.SendChain(img).ID() == 0 {
		return fmt.Errorf("图片发送失败，未添加到分类")
	}
	// 添加插画到对应的数据库table
	p.dbmu.Lock()
	defer p.dbmu.Unlock()
	if err := p.db.Create(imgtype, &pixiv.Illust{}); err != nil {
		return err
	}
	return p.db.Insert(imgtype, illust)
}

func (p *imgpool) remove(imgtype string, id int64) error {
	p.dbmu.Lock()
	defer p.dbmu.Unlock()
	return p.db.Del(imgtype, "WHERE pid = ?", id)
}
