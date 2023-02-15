# kokomi

kokomi是一个ZeroBot-Plugin的扩展插件，提供包括原神角色面板绘制,快速返回图片等升级功能。
相比与喵喵菜单,本插件不依赖浏览器渲染,可以再树莓派等机器上运行(按理来说是可以的),占用内存较低

具体功能可在安装插件后 通过`kokomi菜单`,`/用法kokomi`,`/用法kokomi_wiki`进行查看。
# 安装与更新

请先安装ZeroBot-Plugin,仓库地址[ZeroBot-Plugin](https://github.com/FloatTech/ZeroBot-Plugin)

在配置成功上述程序源码后,再安装本插件,否则无法生效


推荐使用git进行安装，以方便后续升级。在ZeroBot-Plugin根目录夹打开终端，运行

//使用Github

    git clone https://github.com/lianhong2758/kokomi-plugin.git ./plugin/kokomi/
    
//使用Gitte

    git clone https://gitee.com/lianhong2758/kokomi-plugin.git ./plugin/kokomi/

进行下载插件。 

然后在main.go中导入包	`_ "github.com/FloatTech/ZeroBot-Plugin/plugin/kokomi"  // kokomi原神面板`

//在群文件下载

如果是手动下载的zip压缩包，请将解压后的kokomi文件夹放置在ZeroBot-Plugin目录下的plugin文件夹内。

然后在main.go中导入包	`_ "github.com/FloatTech/ZeroBot-Plugin/plugin/kokomi"  // kokomi原神面板`

# 功能说明(部分移植喵喵菜单)
- 原神面板功能:
- kokomi菜单
- 绑定xxx(uid)
- 删除账号
- 更新面板
- 全部面板
- 雷神面板
管理员专属
- 上传第(1|2)立绘XX
- 删除第(1|2)立绘XX
- 切换api[数字]

原神wiki功能:
- #材料/培养xxx [角色培养材料查询]
- #特产/位置xxx [特产位置查询] 
- #武器/图鉴xxx [武器图鉴查询]
- #收益xxx [角色收益曲线查询]
- #参考xxx [角色参考面板查询]
- #查卡xxx [七圣召唤查卡]
- #攻略xxx [角色攻略查询]
- #原魔xxx [原魔图鉴查询]
- #xxx [角色图鉴查询]

# 服务依赖

#更新面板 依赖于面板查询API，面板服务由 `http://enka.network/`提供
如果可以的话，也请在Patreon上支持Enka，或提供闲置的原神账户，具体可在Enka官网 Discord联系

国内网络如Enka服务访问不稳定，可尝试更换 @MiniGrayGay 大佬提供的中转服务
方法:发送切换api即可(需要权限)，未来可能适配喵喵api

    【链接1】：https://enka.microgg.cn/
    【链接2】：https://enka.minigg.cn/
# 未来可期 (以后将适配的功能)
#雷神伤害
#本地计算

免责声明

    功能仅限交流技术使用，请勿将kokomi用于以盈利为目的的场景
    图片与其他素材均来借用喵喵菜单，在此感谢大佬，如有侵权请联系，会立即删除

其他

    喵喵插件[Miao-Plugin]:感谢喵喵菜单提供模板学习
    Enka: 感谢Enka提供的面板服务
    genshin-atlas:感谢西风驿站提供wiki查询功能
    LittlePaimon:感谢小派蒙提供json数据库
    
# 获得帮助
    欢迎进行提问,如果可能,我将尽快回答你的问题,或者fix该问题
    若有开发建议,欢迎来开发群内进行讨论
    QQ群
       ZeroBot-Plugin官方二群609640932
       kokomi开发/维护/交流678586912
