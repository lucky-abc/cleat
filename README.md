# cleat简介
采集Windows操作系统下的事件、文本文件和linux系统下的文本文件

**windows环境：**
- 采集windows事件：Application、System、Security、Setup
- 采集文本文件

**linux环境：**

- 采集linux系统中的文本文件

**转发：**
- 使用udp将采集的数据转发出去
- 转发的字符编码为UTF-8

# 运行
**windows环境：**

执行命令

```shell
bin/cleat.exe
```

> windows环境中如果不使用管理员方式运行，Security类型的日志会拒绝访问

**linux环境：**

执行命令

```shell
bin/cleat
```

