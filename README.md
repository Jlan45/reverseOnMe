# ReverseOnMe
WebSocket直接访问/wstotcp即可获取端口<br>
```bash
#设置监听端口范围，不设置默认20000-60000
HIGH="50000"
LOW="49990"
go build -o ReverseOnMe
./ReverseOnMe
```
访问8081端口即可

也可以用docker，镜像打包好了明天传

```bash
docker run -itd -e HIGH="50000" -e LOW="49990" -p 49990-50000:49990-50000 -P reverseonme
```
API
```
/create 生成一个随机监听端口
{
    "ID": "1yid1tjh",
    "port": 51313
}

/wstotcp/:id 通过获取ID进入对应的监听环境

```
