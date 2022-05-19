# Uranus

名字来源于古希腊神话中的天王乌拉诺斯,通过《海贼王》古代三大兵器了解到的这个名称,同时也是天王星的名称.

## 构建

```bash
# 设置使用国内镜像
export GOPROXY=https://mirrors.aliyun.com/goproxy/

# 安装依赖
make init

# 构建
make
```

## 运行

这个项目的运行依赖 [hackernel](https://github.com/lanthora/hackernel) 项目,根据其说明文档完成构建并启动服务.

编译后的二进制为 `/cmd/dirname/uranus-dirname`, 其中 `dirname` 为 `cmd` 的子目录名.

```bash
# 运行示例程序,将显示进程审计事件
./cmd/sample/uranus-sample
```

其他程序的运行可能需要配置文件,配置文件模板见 `configs` 目录.
