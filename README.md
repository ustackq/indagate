#  Indagate 问答系统简介



---

Indagate 问答系统是一套开源的社交化问答软件系统。作为国内首个推出基于 Goalng微服务化的社交化问答系统，Indagate 期望能够给更多的站长或者企业提供一套完整的社交问答系统，帮助社区或者企业搭建相关的知识库建设。


### Indagate 问答系统的下载

您可以随时从我们的官方下载站下载到最新版本，以及各种补丁

[http://github.com/ustackq/indagate/release][1]

### Indagate 问答系统的环境需求

 1. 安装docker
 2. 可用的 www 服务器，如 Apache、IIS、nginx, 推荐使用性能高效的 Apache 或 nginx.
 3. MySQL 5.0 及以上, 服务器需要支持 MySQLi 或 PDO_MySQL
 

### Indagate 问答系统的安装

 1. 上传 upload 目录中的文件到服务器
 2. 设置目录属性（windows 服务器可忽略这一步）
以下这些目录需要可读写权限
> ./
./system
./system/config 含子目录

 3. 访问站点开始安装
 4. 参照页面提示，进行安装，直至安装完毕


### Indagate 问答系统的升级

升级过程非常简单, 覆盖所有文件之后运行 http://您的域名/upgrade/ 按照提示操作即可


### Indagate 软件的技术支持

当您在安装、升级、日常使用当中遇到疑难，请您到以下站点获取技术支持。

 - 支持：http://www.ustack.io/support/

[1]: http://github.com/ustackq/indagate/release
