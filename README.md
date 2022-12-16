## 自定义的消息队列协议
基于websocket

在连接后，客户端的第一个包应该为认证请求，payload为认证信息
认证请求type为0, topic为auth，message中为认证消息
### 除了认证请求外，有以下几种请求

#### 1.向特定topic发送消息
{
    "type" : 1,
    "topic" : "topic",
    "message" : 
}