### HTTP篇


![http协议](img/http-https-diff.png)
#### **HTTP和HTTPS有哪些区别？**
- 分析：基础在于**超稳本传输协议、端口号，CA证书-服务器身份**，亮点在于点出基于**建立连接、断开连接握手次数**（TCP连接时 三次握手SYN(客户端-服务器端)、SYN+ACK（服务器端-客户端）、ACK（客户端-服务器端）），也可以引导面试官问TCP三四次握手
- 答案：
	- 安全传输：HTTP是超文本传输协议，信息是铭文传输，存在安全风险问题。HTTPS则解决HTTP不安全的缺陷，在TCP和HTTP网络层之间加入了SSL/TLS安全协议，使得报文能够加密传输。
	- 是否加密：HTTP连接建立相对简单，TCP三次握手之后便可进行HTTP的报文传输。而HTTPS在TCP三次握手之后，还需要进行SSL/TLS的握手过程，才可以加密报文传输。
	- 默认端口：两者的默认端口不同，HTTP默认80端口，HTTPS默认443端口
	- CA证书：HTTPS协议需要想CA（证书权威机构）申请数字证书，来保证服务器的身份时可信的。




参考文章：https://www.xiaolincoding.com/network/2_http/http_interview.html
