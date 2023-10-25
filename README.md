# GoFileService

基于Golang编写的简易文件服务。

- [x] 文件上传

- [x] 文件下载

- [x] 限制IP访问

- [x] 使用授权访

### 风险提示：

个人用于可信服务器间文件传输，作为公开地址使用可能有安全性风险，需要自己考量。

scp和sftp可能是更好选择

### 编译流程

````
go mod tidy

go build .
````

### .env配置文件说明

```sh
# 开启服务的端口
ADDRESS=":5050"
# 工作目录，用于存放上传和下载的文件
WORK_PATH="D:/"
# 是否启用HTTPS(true/false)
HTTPS="false"
# TLS证书路径
HTTPS_CERT_FILE=""
# TLS密钥路径
HTTPS_KEY_FILE=""
# 授权文件路径
AUTH_CONFIG_FILE="./auth.yaml"
# 下载路径配置
# 如果 DOWNLOAD_RELATIVE_PATH 配置为 "/download", 那么
# 就可以通过URL(未启用 HTTPS的情况下): GET:http://IP:ADDRESS/download/xxx
# 下载D:/xxx的文件
DOWNLOAD_RELATIVE_PATH="/download"

# 上传路径配置
# 如果 UPLOAD_RELATIVE_PATH 配置为 "/upload", 那么
# 就可以通过URL(未启用 HTTPS的情况下): POST:http://IP:ADDRESS/upload
# 进行文件上传
# POST Body 参数(form-data格式):
#  file: 要上传的文件的字段
#  to: 要保存的路径字段
#
# 如果参数中的to是"/test/"那么文件会上传到WORK_PATH/test/filename
UPLOAD_RELATIVE_PATH="/upload"
```

### auth.yaml配置文件说明

```
# 授权可访问的Origin，如果地址在这里配置了
# 那么就不会使用AuthorizationHeader进行身份验证，而直接允许访问
AuthorizedIPs:
 - 127.0.0.1

# 身份验证代码，在没有授权地址的情况下，会使用这里的参数进行身份验证
# 如果该值为空，则相当于取消了身份验证。任何IP都可以直接访问
AuthorizationHeader : Authorization
```

