# Ant-Style
默认实现是AntPathMatcher，支持* Ant-Style的模式语法。 该描述来自java spring框架。


# ant path 通配规则

> Table Ant Wildcard Characters

| Wildcard | Description |
| :-- | :--|
|?|匹配任何单字符|
|*|匹配0或者任意数量的字符|
|**|匹配0或者更多的目录|

> Table Example Ant-Style Path Patterns

| Path | Description |
| :-- | :-- |
| /app/*.x | 匹配(Matches)所有在app路径下的.x文件 |
| /app/p?ttern | 匹配(Matches) /app/pattern 和 /app/pXttern,但是不包括/app/pttern |
| /**/example | 匹配(Matches) /app/example, /app/foo/example, 和 /example |
| /app/**/dir/file. | 匹配(Matches) /app/dir/file.jsp, /app/foo/dir/file.html,/app/foo/bar/dir/file.pdf, 和 /app/dir/file.java	 |
| /**/*.jsp | 匹配(Matches)任何的.jsp 文件 |

# 基本使用PathMatcher接口

> 使用demo
```go
package main

import "fmt"
import "github.com/aluka-7/antstyle"

func main(){
	fmt.Println(antstyle.Match("test","test"))
	fmt.Println(antstyle.Match("test*aaa", "testblaaaa"))
	fmt.Println(antstyle.Match("t?st", "test"))
	fmt.Println(antstyle.Match("/{bla}.*", "/testing.html"))
}
```
