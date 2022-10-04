# Uranus

[Tellus](https://github.com/lanthora/tellus/wiki) 组件之一,
是包括 Web 后端在内的以 [hackernel](https://github.com/lanthora/hackernel) 为服务端的客户端程序.

## 构建

本项目依赖 [hackernel](https://github.com/freshdom/hackernel)
提供的 Unix Domain Socket 接口,请先根据文档部署 hackernel.

```bash
# 如果更新依赖时出现网络问题,可以设置使用国内镜像
export GOPROXY=https://goproxy.cn

# 安装依赖
make init

# 构建
make
```

## 运行

编译后的二进制为 `/cmd/dirname/uranus-dirname`, 其中 `dirname` 为 `cmd` 的子目录名.

```bash
# 运行示例程序,将显示进程审计事件
./cmd/sample/uranus-sample
```

其他程序的运行可能需要配置文件,配置文件模板见 `configs` 目录.

