# TIFF图片转换器

将指定目录下所有TIFF格式文件批量转换为其他格式图片，主要服务器于上百万图片转换。

# 安装
```
go install github.com/ysqi/tiff2
```

# 运行
```shell
$ ./tiff2 -h
```

# Help 
```
usage: tiff2 [flags] [path ...]
  -o string
        转换后文件存放目录 (default "./output")
  -r    转换存储时是否覆盖已存在的文件
  -t string
        需要转换的目标格式 (default "jpg")
```