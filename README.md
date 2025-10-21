# Overview

此应用用于读取流感疫苗预登记码（条形码）并把条形码重新排版到 PDF 文件中。

# Usage

```
go run . -s [source_dir] -t [title]
```

- `[source_dir]`: 原始预登记码(jpg 文件)的存放位置
- `[title]`: 生成的 PDF 文件名

# Q&A

## 如何把当前目录下的多个 png 文件转换成 jpg 文件?

使用 GraphicsMagick 来转换：

```shell
for f in *.png; do gm convert "$f" "${f%.png}.jpg"; done
```

## 接龙管理中下载的 zip 压缩包解压出来的文件名都是乱码，要怎么处理?

使用 `unzip` 来解压并通过 `-O` 选项指定归档文件中的文件名编码

```shell
unzip -O gbk -d /tmp/source_dir/
```

# References

https://www.onlinebarcodereader.com/
