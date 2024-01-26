### Redis集群中的纪元(epoch)

#### 纪元(epoch)
Redis Cluster使用了类似于Raft算法term（任期）的概念称为epoch（纪元）， 用来给事件增加版本号。Redis集群中的纪元主要是两种：currentEpoch和configEpoch。  

#### currentEpoch  
这是一个集群状态相关的概念，可以当作记录集群状态变更的递增版本号。每个集群节点，都会通过server.cluster->currentEpoch记录当前的currentEpoch。  

集群节点创建时，不管是master还是slave，都置currentEpoch为0.当前节点接收到来自其他节点都包时，如果发送着的currentEpoch（消息头部会包含发送者的currentEpoch）大于当前节点的currentEpoch，那么当前节点会更新currentEpoch为发送着的currentEpoch。因此集群中所有节点的currentEpoch最终会达成一致，相当于集群状态的认知达成了一致。  

#### CurrentEpoch作用  

currentEpoch作用在于，当集群的状态发生改变，某个节点为了执行一些动作需要寻求其他节点的同意时，就会增加currentEpoch的值。目前currentEpoch只用于slave的故障转移流程，这就跟哨兵中的[sentinel](https://so.csdn.net/so/search?q=sentinel&spm=1001.2101.3001.7020).current_epoch作用时一摸一样的。当slave A发现其所属的master下线时，就会试图发起鼓掌流程转移流程。首先就是增加currentEpoch的值，这个增加后的currentEpoch是所有集群节点中最大的。然后slave A向所有节点发起拉票请求，请求其他master投票给自己，使自己能成为新的master。其他节点收到包后，发现发送者的currentEpoch比自己大currentEpoch大，就会更新自己的currentEpoch，并在尚未投票的情况下，投票给slave A，表示同意使其成为新的master。  

#### configEpoch  
configEpoch主要用于解决不同的节点的配置发生冲突的情况。举个例子就明白了：节点A宣称负责slot1，其向外发送的包中，包含了自己的configEpoch和负责的slots信息。节点c收到A发来的包后，发现自己当前没有记录slot1的负责节点（也就是server.cluster->slots[1]为null），就会将A置为slot1的负责节点（server.cluster->slots[1] = A）,并记录节点A的configEpoch。后来，节点C又收到了B发来的包，它也宣称负责slot1，此时，如何判断slot1到底由谁负责呢？  

这就是configEpoch起作用的时候了，C在B发来的包中，发现它的configEpoch，要比A的大，说明B是更新的配置。因此，九江slot1的负责节点设置为B(server.cluster->slots[1] = B)。在slabe发起选举，获得足够多的选票后，成功当选时，也就是slave试图替代其已经下线的旧master，成为新的master时，会增加它自己的configEpoch，使其成为当前所有集群节点的configEpoch中的最大值。这样，该slave成为master后，就会像所有节点发送广播包，强制其他节点更新相关slots的负责节点为自己。  

### 参考资料：
- [原文](https://blog.csdn.net/chen_kkw/article/details/82724330)  
- [Redis Ccluster Specification](https://redis.io/docs/reference/cluster-spec/)  
- [Redis Cluster tutorial](https://redis.io/docs/management/scaling/)  
- [Redis系列九：redis集群高可用](https://www.cnblogs.com/leeSmall/p/8414687.html)  
- [Redis远吗解析：27集群（三）主从复制、故障转移](https://www.cnblogs.com/gqtcgq/p/7247042.html)  