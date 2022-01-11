// Package aifalse 暂时只有服务器监控
package aifalse

import (
	"fmt"
	"math"
	"os"
	"time"

	control "github.com/FloatTech/zbpctrl"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() { // 插件主体
	engine := control.Register("aifalse", &control.Options{
		DisableOnDefault: false,
		Help: "AIfalse\n" +
			"- 查询计算机当前活跃度: [检查身体|自检|启动自检|系统状态]",
	})
	engine.OnFullMatchGroup([]string{"检查身体", "自检", "启动自检", "系统状态"}, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(
				"* CPU占用: ", cpuPercent(), "%\n",
				"* RAM占用: ", memPercent(), "%\n",
				"* 硬盘使用: ", diskPercent(),
			),
			)
		})
	engine.OnFullMatch("清理缓存", zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			err := os.RemoveAll("data/cache/*")
			if err != nil {
				ctx.SendChain(message.Text("错误: ", err.Error()))
			} else {
				ctx.SendChain(message.Text("成功!"))
			}
		})
}

func cpuPercent() float64 {
	percent, _ := cpu.Percent(time.Second, false)
	return math.Round(percent[0])
}

func memPercent() float64 {
	memInfo, _ := mem.VirtualMemory()
	return math.Round(memInfo.UsedPercent)
}

func diskPercent() string {
	parts, _ := disk.Partitions(true)
	msg := ""
	for _, p := range parts {
		diskInfo, _ := disk.Usage(p.Mountpoint)
		pc := uint(math.Round(diskInfo.UsedPercent))
		if pc > 0 {
			msg += fmt.Sprintf("\n  - %s(%dM) %d%%", p.Mountpoint, diskInfo.Total/1024/1024, pc)
		}
	}
	return msg
}
