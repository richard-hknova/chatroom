# 项目实战:
开发一个简单的聊天工具后台系统
1. 采取微服务架构，包括但不限于用户服务、消息服务、好友服务， 仅需实现用户注册、登录、添加好友、发送消息等基本功能就行
2. 需要采用的技术栈
   1. Golang
   2. Gin作为Web框架
   3. Gorm作为ORM框架
   4. Go-Zero作为微服务框架
   5. Redis作为缓存中间件
   6. Mysql作为数据库
3. 依赖组件如mysql、redis、etcd等可以采用docker容器化部署

## 项目结构：
1. database: 用Mysql作为数据库，用Redis作为缓存。包含数据：
    * 用户(登入，注册，验证，查找)
    * 好友(好友请求，接受请求，删除/拒绝好友)
    * 对话(在对话对象未上线时存入数据库并在对象登入时取出，备份对话)
2. handler: 处理websocket和http请求。
    * /ws：验证身份并开启Websocket通信
    * /user/signin：用户登入并获取好友，好友请求，未接受的信息等数据
    * /user/signup：注册新用户
    * /user/search：搜索用户
    * /friend/request：发送好友请求
    * /friend/accept：接受好友请求
    * /friend/delete：删除/拒绝好友