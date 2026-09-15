package setutime

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/AnimeAPI/pixiv"
	"github.com/tidwall/gjson"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// DefaultTransport 支持 HTTP_PROXY / HTTPS_PROXY，所有请求均有超时。
var pixivHTTP = &http.Client{Timeout: 30 * time.Second}

func pixivGet(link string, limit int64) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", "https://www.pixiv.net/")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := pixivHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Pixiv HTTP %d: %s", resp.StatusCode, link)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("Pixiv 响应超过 %d 字节", limit)
	}
	return data, nil
}

// 本插件只发送第零页，不推测多页作品的文件名或扩展名。
func pixivWorks(id int64) (*pixiv.Illust, error) {
	if id <= 0 {
		return nil, fmt.Errorf("无效 Pixiv ID: %d", id)
	}
	data, err := pixivGet("https://www.pixiv.net/ajax/illust/"+strconv.FormatInt(id, 10), 4<<20)
	if err != nil {
		return nil, err
	}
	if !gjson.ValidBytes(data) {
		return nil, fmt.Errorf("Pixiv 返回了无效 JSON")
	}
	result := gjson.ParseBytes(data)
	if result.Get("error").Bool() {
		return nil, fmt.Errorf("Pixiv: %s", result.Get("message").Str)
	}
	body := result.Get("body")
	original := body.Get("urls.original").Str
	if body.Get("illustId").Int() != id || original == "" {
		return nil, fmt.Errorf("Pixiv 作品 %d 不可用或未返回原图地址", id)
	}
	if _, err := imageURL(original); err != nil {
		return nil, err
	}
	illust := &pixiv.Illust{
		Pid: id, Title: body.Get("illustTitle").Str,
		Caption:   strings.ReplaceAll(body.Get("illustComment").Str, "<br />", "\n"),
		Tags:      fmt.Sprintln(body.Get("tags.tags.#.tag").Array()),
		ImageUrls: []string{original}, AgeLimit: "all-age",
		CreatedTime: body.Get("createDate").Str,
		UserID:      body.Get("userId").Int(), UserName: body.Get("userName").Str,
	}
	if body.Get("xRestrict").Int() != 0 {
		illust.AgeLimit = "r18"
	}
	return illust, nil
}

func imageURL(link string) (*url.URL, error) {
	u, err := url.Parse(link)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.Path == "" || strings.HasSuffix(u.Path, "/") {
		return nil, fmt.Errorf("无效 Pixiv 图片地址: %q", link)
	}
	return u, nil
}

func validImage(data []byte) bool {
	_, _, err := image.Decode(bytes.NewReader(data))
	return err == nil
}

func (p *imgpool) image(illust *pixiv.Illust) (message.Segment, error) {
	// 旧数据库可能缺少地址；已有地址下载失败时也重新查询一次。
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 || len(illust.ImageUrls) == 0 || illust.ImageUrls[0] == "" {
			fresh, err := pixivWorks(illust.Pid)
			if err != nil {
				return message.Segment{}, fmt.Errorf("刷新作品 %d 失败: %w (下载: %v)", illust.Pid, err, lastErr)
			}
			illust.ImageUrls = fresh.ImageUrls
		}
		data, err := p.cachedImage(illust.ImageUrls[0])
		if err == nil {
			// OneBot 与插件可以运行在不同机器，无需共享缓存目录。
			return message.ImageBytes(data), nil
		}
		lastErr = err
	}
	return message.Segment{}, fmt.Errorf("下载作品 %d 失败: %w", illust.Pid, lastErr)
}

func (p *imgpool) cachedImage(link string) ([]byte, error) {
	u, err := imageURL(link)
	if err != nil {
		return nil, err
	}
	name := filepath.Base(u.Path)
	if name == "." || name == ".." || strings.ContainsAny(name, `\:`) {
		return nil, fmt.Errorf("无效图片文件名: %q", name)
	}
	path := filepath.Join(p.path, name)
	if data, err := os.ReadFile(path); err == nil && validImage(data) {
		return data, nil
	}
	// 不依赖 HEAD、Content-Length 或服务端的 Range 支持。
	data, err := pixivGet(link, 32<<20)
	if err != nil {
		return nil, err
	}
	if !validImage(data) {
		return nil, fmt.Errorf("Pixiv 返回的内容不是有效图片: %s", link)
	}
	if err := os.MkdirAll(p.path, 0755); err != nil {
		return nil, err
	}
	// 完成下载后再原子替换，防止其他请求读到半张图片。
	f, err := os.CreateTemp(p.path, ".pixiv-*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())
	_, writeErr := f.Write(data)
	closeErr := f.Close()
	if writeErr != nil {
		return nil, writeErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if err := os.Rename(f.Name(), path); err != nil {
		return nil, err
	}
	return data, nil
}
