# Go_Auto_Proxy

自动设置Windows 本地 IE代理

快速上手：
```
proxy.exe -u http://127.0.0.1:5010/get_all/ -t 10
```

帮助：
```
C:\Users\xxx\proxy.exe -h
Usage of proxy.exe:
  -c string
        -c cls 重置代理设置为自动代理
  -l int
        循环次数 (default 1)
  -t int
        自动切换代理时间间隔 (default 30)
  -u string
        代理 Url，例如 http://127.0.0.1:5010/get_all

```

目前比较遗憾的是获取到的IP不一定是HTTPS，免费的代理自然比不上收费的质量好，如果师傅们有资源可以自行修改下 getproxy 函数和 Proxy_Pool 结构体

当然如果有好的代理池推荐可以发下Issues，不胜感激
