# Uranus

名字来源于古希腊神话中的天王乌拉诺斯,通过《海贼王》古代三大兵器了解到的这个名称,同时也是天王星的名称.

起初开发 hackernel 辅助学习 Linux 内核相关的知识,在内核中做行为的审计和拦截.
然而其由 C/C++ 开发的,做功能扩展或者与用户交互都很困难,
因此仅实现核心功能并对外提供 Unix Domain Socket 接口.

Uranus 的存在就是为了解决 hackernel 遗留下的问题:
通过 Golang 丰富的第三方库完成 C/C++ 无法轻易完成的功能,
封装 hackernel 接口,对外提供更简单易操作的的用户界面.

更多功能参考[项目网站](https://hackernel.org/).

## 构建

本项目依赖 [hackernel](https://github.com/freshdom/hackernel)
提供的 Unix Domain Socket 接口,请先根据文档部署 hackernel.

```bash
# 设置使用国内镜像
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
