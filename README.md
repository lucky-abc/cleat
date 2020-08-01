# cleat简介
采集Windows操作系统下的事件日志、文本文件和linux系统下的文本文件

**windows环境：**
- 采集windows事件：Application、System、Security、Setup
- 采集文本文件

**linux环境：**

- 采集linux系统中文本文件

**输出：**
- 支持udp方式将数据输出
- 支持tcp方式将数据输出
- 输出数据的字符编码为UTF-8

# 运行
**windows环境：**

执行命令

```shell
bin/cleat.exe
```

> windows环境中如果不使用管理员方式运行，Security类型事件日志会被拒绝访问

**linux环境：**

执行命令

```shell
bin/cleat
```

