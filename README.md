# Overview

此应用用于读取流感疫苗预登记码（条形码）并把条形码重新排版到 PDF 文件中。

# Usage

```
go run . -s [source_dir] -t [title]
```

- `[source_dir]`: 原始预登记码(jpg/jpeg/png 文件)的存放位置。目录中不能包含其他格式的文件，也不能有同名(仅扩展名不同)的文件
- `[title]`: 生成的 PDF 文件名

# Q&A

## 接龙管理中下载的 zip 压缩包解压出来的文件名都是乱码，要怎么处理?

使用 `unzip` 来解压并通过 `-O` 选项指定归档文件中的文件名编码

```shell
unzip -O gbk -d /tmp/source_dir/
```

# References

https://www.onlinebarcodereader.com/
