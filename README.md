tgState
==

一款以Telegram作为储存的文件外链系统

本fork版: `已修复原作者分包下载异常报错,报错原因: Request Entity Too Large` 及其他部分优化和改进.

<del>不限制文件大小和格式.</del>

虽然说不限制,但实际上,(为了能够分包下载,因为[官方限制:下载文件大小不能超过20MB](https://core.telegram.org/bots/faq#handling-media))超过20MB文件会分割上传,且受限于链路中各个环节对POST请求体大小的限制,故实际可上传大小请结合实际情况.

可以作为telegram图床,亦可作为telegram网盘使用

支持web上传文件和telegram直接上传

# 参数说明

必填参数

 - target
 - token

可选参数

 - pass
 - mode
 - url
 - port

## target

目标可为频道、群组、个人

当目标为频道时，需要将Bot拉入频道作为管理员，公开频道并自定义频道Link，target值填写Link，如@xxxx

当目标为群组时，需要将Bot拉入群组，公开群组并自定义群组Link，target值填写Link，如@xxxx

当目标为个人时，则为telegram id(@getmyid_bot获取)

## token

填写你的bot token

## pass

填写访问密码，如不需要，直接填写```none```即可

## mode

 - ```p``` 代表网盘模式运行，不限制上传后缀
 - ```m``` 在p模式的基础上关闭网页上传，可私聊进行上传（如果target是个人，则只支持指定用户进行私聊上传

## url

bot获取FileID的前置域名地址自动补充

## port

自定义运行端口

# 管理

## 获取FIleID

对bot聊天中的文件引用并回复```get```可以获取FileID，搭建地址+获取的path即可访问资源

如果配置了url参数，会直接返回完整的地址

![image](https://github.com/csznet/tgState/assets/127601663/5b1fd6c0-652c-41de-bb63-e2f20b257022)

# 部署

## 编译(推荐)

确保已经安装`go`

```
curl https://codeload.github.com/LanYunDev/tgState/zip/refs/heads/main -
-output main.zip
unzip main.zip
cd tgState-main
go build -ldflags "-w -s"
# Debug模式: go build -gcflags "all=-N -l"
```

## 二进制

Linux amd64下载

```
wget https://github.com/LanYunDev/tgState/releases/latest/download/tgstate.zip && unzip tgstate.zip && rm tgstate.zip
```

**使用方法**

```
 ./tgstate 参数
```

**例子**
```
 ./tgstate -token xxxx -target @xxxx
```

**后台运行**

```
nohup ./tgstate 参数 &
```

## Docker

参考[源项目](https://github.com/csznet/tgState)

之后用命令 `docker cp <本地二进制路径> <容器ID或容器名称>:/app/tgState` 替换二进制文件.

# API说明

POST方法路径: `/api`

表单传输，字段名为image，内容为二进制数据

文件下载路径: `/api/download/`
