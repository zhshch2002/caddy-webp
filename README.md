# Caddy Webp
Caddy2插件——无感的将图片转换为WebP格式。

# Get Start
[release](https://github.com/zhshch2002/caddy-webp/releases)里有我编译的linux-amd64的Caddy可执行文件，其他的平台如下：
1. 克隆本仓库
2. `go build ./caddy/main.go`
3. 将编译得到的新文件替换Caddy原始`/usr/bin/caddy`
4. 修改Caddyfile

```
{ # 全局块
    order webp before file_server # 设置Handler激活顺序，缺少将无法启动caddy
}

localhost {
    root * /var/www/
    encode gzip
    webp # <= webp插件在这里启动了！
    file_server
}
```

![](./screenshot.png)

Just for fun!