# uller_log
  
## 高性能日志服务  
它可以把一定时间段内的相同日志信息在归为单条的日志信息并加以计数，支持C/S结构（tcp长链），以用户指定的存储形式（文件、数据库）存储。举例：30秒内A机器的a1程序包了100条日志信息，那么它的相应错误日志为：192.168.1.1 a1日志信息 100次。
  
## 目录说明  
comm                程序目录  
log_client          客户端调用示例  
    |-config.ini    配置文件  
log_server          服务器端调用示例  
    |-contfig.ini   配置文件  
    
##客户端配置文件说明
[devLogDB]  
数据库配置 

[devLog]  
日志基础配置  
storageType：日志存储类型，枚举file：本地文件存储，db：数据库存储，remotServer：远程服务器存储  
logFilePath：storageType=file，本地存储文件路径  
pingInterval：storageType=remotServer时，客户端发送心跳包时间间隔，单位秒  
encrypSecret：storageType=remotServer时，客户端想服务器发送数据加密秘钥，为空则不加密  
remotServerIP：storageType=remotServer时，远程服务器ip  
remotServerPort：storageType=remotServer时，远程服务器端口  
sendInterval：日志存储时间间隔，单位秒  
sendCount：日志缓存数量，开启缓存后当日志数量到达一定值后忽略SendInterval值将日志存储  
localCache：本地是否开启缓存  
  
##服务器端配置文件说明  
[devLogDB]  
数据库配置  

[devLog]  
日志基础配置  
localIp：本地服务器ip  
localPort：本地服务器端口  
storageType：日志存储类型，枚举file：本地文件存储，db：数据库存储，remotServer：远程服务器存储  
logFilePath：storageType=file，本地存储文件路径  
sendInterval：日志存储时间间隔，单位秒  
sendCount：日志缓存数量，开启缓存后当日志数量到达一定值后忽略SendInterval值将日志存储  
localCache：本地是否开启缓存  
whiteList：接收客户端ip白名单，多个以逗号分隔  
blackList：接收客户端ip黑名单，多个以逗号分隔  
  
重复造轮子（https://gitee.com/Uller/GLLog）是学习一门语言的最好方法，看着丑陋的golang，写起来非常舒服。