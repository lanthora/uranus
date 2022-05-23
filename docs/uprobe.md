# uprobe
 
uprobe需要向内核传递一个函数地址,在这个地址的函数被调用时触发事件.
编译一个简单的 Go 程序,通过uprobe追踪其中的某个函数的调用.
下面的例子追踪的是 Divide 函数,根据被除数除数计算商和模.

```go
package main

import (
	"fmt"
)

//go:noinline
func Divide(dividend int, divisor int) (quotient int, remainder int) {
	quotient = dividend / divisor
	remainder = dividend % divisor
	return
}

func main() {
	dividend, divisor := 5, 3
	quotient, remainder := Divide(dividend, divisor)
	fmt.Printf("dividend=%d divisor=%d quotient=%d remainder=%d\n", dividend, divisor, quotient, remainder)
}
```

通过 `//go:noinline` 禁止 `Divide` 函数内联.

```bash
go build -o uranus-uprobe
```

从符号表中找到函数 `Divide` 的地址为 `0x47f500`,

```bash
readelf --syms uranus-uprobe | grep Divide
#1447: 000000000047f500    53 FUNC    GLOBAL DEFAULT    1 main.Divide
```

在 libbpf-bootstrap 的 uprobe 示例中有这样一段注释

> If we were to parse ELF to calculate this function, we'd need 
> to add .text section offset and function's offset within .text
> ELF section.


所以需要两个偏移量:

* 函数在.text段中的偏移量 (0x47f500 - 0x401000)
* .text本身的偏移量 (0x1000)

最终结果为 0x7f500 (0x47f500 - 0x401000 + 0x1000)

```bash
readelf --section-headers uranus-uprobe | grep .text
# [ 1] .text             PROGBITS         0000000000401000  00001000
```

参考内核文档进行设置.

> Similar to the kprobe-event tracer, this doesn’t need to be activated via current_tracer. Instead of that, add probe points via /sys/kernel/debug/tracing/uprobe_events, and enable it via /sys/kernel/debug/tracing/events/uprobes/<EVENT>/enable.

```bash
# 注册一个新的uprobe
echo 'p /path/to/uranus-uprobe:0x7f500' >> /sys/kernel/debug/tracing/uprobe_events

# 如果偏移量计算错误,这里会出现写入错误
echo 1 > /sys/kernel/debug/tracing/events/uprobes/enable
```

```bash
# 运行测试程序
/path/to/uranus-uprobe

# 查看日志
cat /sys/kernel/debug/tracing/trace
```

```txt
# tracer: nop
#
# entries-in-buffer/entries-written: 1/1   #P:16
#
#                                _-----=> irqs-off/BH-disabled
#                               / _----=> need-resched
#                              | / _---=> hardirq/softirq
#                              || / _--=> preempt-depth
#                              ||| / _-=> migrate-disable
#                              |||| /     delay
#           TASK-PID     CPU#  |||||  TIMESTAMP  FUNCTION
#              | |         |   |||||     |         |
   uranus-uprobe-14031   [007] DNZff  6798.683418: p_uranus_0x7f500: (0x47f500)
```

接下来的目标就是从 Go 二进制中(`.gopclntab`段)计算出 `0x7f500`.

`strip` 去除符号表后,通过[GoReSym](https://github.com/mandiant/GoReSym)可以拿到函数地址`0x47f500`,
与符号表中的记录地址一致.
