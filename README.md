

<h1 align="center">
  GoFileService是基于Golang编写的简易文件服务
</h1>

<p align="center">
  <img src="https://img.shields.io/github/v/release/GenesisAN/GoFileService?label=version">
  <img src="https://img.shields.io/github/actions/workflow/status/GenesisAN/GoFileService/build.yml">
</p>



- [x] 文件上传

- [x] 文件下载

- [x] 限制IP访问

- [x] 使用授权访

### 风险提示：

个人用于可信服务器间文件传输，作为公开地址使用可能有安全性风险，需要自己考量。

scp和sftp可能是更好选择

### 编译流程

```sh
go mod tidy
go build .
# 或者直接运行build.sh脚本
```

### 使用方法

```yaml
# 发起下载请求
# 如果WORK_PATH是D:/, 文件路径是test/1.txt
# 那么通过 HTTP GET http://IP:ADDRESS/DOWNLOAD_RELATIVE_PATH/test/1.txt
# 就可以下载到D:/test/1.txt
GET: 
   - 请求URL: http://IP:ADDRESS/DOWNLOAD_RELATIVE_PATH/文件路径
   - HEAD:
       - Authorization: 身份验证代码
       - Origin: 请求来源地址

# 发起上传请求
# 如果WORK_PATH是D:/, 上传路径是test/，文件名称是1.txt
# 那么通过 HTTP POST http://IP:ADDRESS/UPLOAD_RELATIVE_PATH
# BODY: form-data
#  to参数的值是/test/，
#  file参数的值是1.txt
# 就会上传到D:/test/1.txt
POST:
    - 请求URL: http://IP:ADDRESS/UPLOAD_RELATIVE_PATH
    - HEAD:
         - Authorization: 身份验证代码
         - Origin: 请求来源地址
    - BODY: # form-data格式
         - file: 要上传的文件
         - to: 上传路径

# ADDRESS,DOWNLOAD_RELATIVE_PATH，UPLOAD_RELATIVE_PATH在.env文件中配置
# Authorization和Origin需要填的值在auth.yaml文件中配置
```
### .env配置文件说明

```dotenv
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

```yaml
# 授权可访问的Origin，如果地址在这里配置了
# 那么就不会使用AuthorizationHeader进行身份验证，而直接允许访问，可以是网址或者IP
AuthorizedIPs:
 - 127.0.0.1

# 身份验证代码，在没有授权地址的情况下，会使用这里的参数进行身份验证
# 如果该值为空，则相当于取消了身份验证。任何IP都可以直接访问
AuthorizationHeader : Authorization
```

