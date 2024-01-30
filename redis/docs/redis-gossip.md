### Redis Cluster Gossip协议

今天来讲一下Redis Cluster的Gossip协议和集群操作，文章思维导图如下所示：  
![cluster-gossip](img/redis-cluster-gossip.png)

#### 集群模式和Gossip简介  
**对于数据存储领域，当数据量或者请求量到达一定程度后，就必然会引入分布式**。比如Redis，虽然其单机性能十分优秀，但是遇到下列原因时，也不得不引入集群。  
- 单机无法保证高可用，需要引入多实例来提供高可用性  
- 单机能够提供高达8W左右的QPS，再高的QPS则需要引入多实例  
- 单机能够支持的数据量有限，处理更多的数据需要引入多实例  
- 单机所处理的网络流量已经超过服务器网卡的上限值，需要引入多实例来分流  

有集群，集群往往需要维护一定的元数据，比如实例的ip地址，缓存分片的slots信息等，所以需要一套分布式机制来维护元数据的一致性。这类机制一般有两个模式：分散式和集中式  

分散式机制将元数据存储在部分或者所有节点上，不同节点之间进行不断的通信来维护元数据的变更和一致性。Redis Cluster， Consul等都是该模式。  


![redis-decentralized](img/redis-cluster-decentralized.png)  

而集中式是将集群元数据集中存储在外部节点或者中间件上，比如zookeeper。旧版本的kafka和storm等都是使用该模式。

![redis-decentralized](img/redis-cluster-centralized.png)  

两种模式个有优劣，具体如下表所示：  
| 模式 | 优点 | 缺点 |
|--------|--------|--------|
| 集中式 | 数据更新及时，时效好，元数据的更新和读取，时效性非常好，一旦元数据出现变更，立即就更新到集中式的外部节点中，其他节点读取的时候立即可以感知到 | 较大数据更新压力，更新压力全部集中在外部节点，作为单节点影响整个系统 |
| 分散式 | 数据更新压力分散，元数据的更新比较分散，不是集中某一个节点，更新请求比较分散，而且有不同的节点处理，有一定的延时，降低了并发压力 | 数据更新延迟，可能导致集群的感知有一定的滞后 |  

**【分散式】**的元数据模式有多种可选的算法进行元数据的同步，比如说`Paxos`、`Raft`和`Gossip`。`Paxos`和`Raft`等都需要全部节点或者大多数节点（超过一半）正常运行，整个集群才能稳定运行，而`Gossip`则不需要半数以上的节点运行。  

Gossip协议，顾名思义，就像流言蜚语一样，利用一种随机、带有传染性的方式，将信息传播到整个网络中，并在一定时间内，使得系统内的所有节点数据一致。对你来说，掌握这个协议不仅能很好地理解这种最常用的，实现最终一致性的算法，也能在后续工作中得心应手地实现数据的最终一致性。  

![redis-gossip](img/redis-cluster-gossip.gif)  

Gossip协议又称epidemic协议（epidemic protocol），是基于流行病传播方式的节点或者进程之间信息交换的协议，在P2P网络和分布式系统中应用广泛，他的方法论也特别简单：  
> 在一个处于游街网络的集群里，如果每个界定都随机与其他节点交换特定信息，经过足够长的时间后，集群各节点对该身份信息的认知终将收敛到一致。  

这里的“特定信息”一般就是指集群状态、各节点的状态以及其他元数据等。Gossip协议是完全符合BASE原则，可以用在任何要求最终一致性的领域，比如分布式存储和注册中心。另外他可以很方便地实现*弹性集群*，允许节点随时上下线，提供快捷的失败检测和动态负载均衡等。  

此外，Gossip协议最大的好处是，及时集群节点的数量增加，每个节点的负载也不会增加很多，几乎是恒定的。这就允许Redis cluster或者Consul集群管理的节点规模能横向扩展到数千个。  

#### Redis Cluster的Gossip通信机制。

Redis Cluster是在3.0版本引入集群功能。为了让集群中的每个实例都知道其他所有实例的状态信息，Redis集群规定各个势力之间按照Gossip协议来通信传递信息。  

![redis-cluster-node-gossip](img/redis-cluster-node-gossip.png)  


上图展示了主从架构的Redis Cluster示意图，其中实线表示节点间著丛夫之关系，而虚线表示各节点之间的Gossip通信。  

- Redis Cluster中**每个节点都维护一份自己视角下的当前整个集群的状态**，主要包括：  
	- 当前集群状态
	- 集群各节点所负责的slots信息，及其migrate状态。
	- 集群中各节点的master-slave状态
	- 集群中各节点的存活状态及怀疑Fail状态  

> *也就是说上面的信息，就是集群中Node相互八卦传播流言蜚语的内容主题，而且比较全面，既有自己的更有别人的，这么一来大家都相互传，最终信息就全面一致了。*  

- Redis cluster的节点之间会互相发送多种消息，较为重要的如下所示：  
	- MEET：通过[cluster meet ip port]命令，已有集群的节点会向新的节点发送邀请，加入现在已有的集群，然后新节点就会开始与其他节点进行通信；  
	- PING：节点按照配置的时间间隔向集群中其他节点发送ping消息，消息中带有自己的状态，还有自己维护的集群元数据，和部分其他节点的元数据；  
	- PONG：节点用于回应PING和MEET消息，结构和PING消息类似，也包含自己的状态和其他信息，也可以用于信息广播和更新；
	- FAIL：节点PING不通某个节点后，回想集群所有节点广播该节点挂掉的消息。其他节点收到消息后标记已下线。  

Redis的源码中cluster.h文件定义了全部的消息类型，代码为Redis4.0版本
```c
// 注意，PING、PONG和MEET实际上是同一种消息
// PONG是对PING的恢复，他的实际格式也为PING消息
// 而MEET 则是一种特殊的PING消息，用于强制消息的接受者将消息的发送着添加到集群中（如果节点尚未在节点列表中的话）。
#define CLUSTERMSG_TYPE_PING 0 /* ping消息 */
#define CLUSTERMSG_TYPE_PONG 1 /* pong消息 */
#define CLUSTERMSG_TYPE_MEET 2 /* Meet请求某个节点添加到集群中 */
#define CLUSTERMSG_TYPE_FAIL 3 /* Fail将某个节点标记为Fail */
#define CLUSTERMSG_TYPE_PUBLISH 4 /* 通过发布与订阅功能广播消息 */
#define CLUSTERMSG_TYPE_FAILOVER_AUTH_REQUEST 5 /* 请求进行故障转移操作，要求消息的接收者通过投票来支持消息的发送着 */
#define CLUSTERMSG_TYPE_FAILOVER_AUTH_ACK 6 /* 消息的接收者同意向消息的发送着投票 */
#define CLUSTERMSG_TYPE_UPDATE 7 /* slots已经发生变化，消息发送着要求消息接收者进行相似更新 */
#define CLUSTERMSG_TYPE_MFSTART 8 /* 为了进行手动故障转移，暂停各个客户端 */
#define CLUSTERMSG_TYPE_COUNT 9 /* 消息总数 */
```  
通过上述这些消息，集群中的每一个实例都能获得其他所有实例的状态信息。这样一来，即使有新节点加入、节点故障、slot变更等事件发生，实例间也可以通过PING、PONG消息的传递，完成集群状态在每个实例上的同步。下面，我们一次来看看集中常见的场景。  

#### 定时PING/PONG消息
Redis Cluster中的节点都会定时向其他节点发送PING消息，来交换各个几点状态信息，检查各个节点状态，包括在线状态、疑似下线状态PFAIL和已经下线状态FAIL。  

- Redis集群的定时PING/PONG的工作原理可以概括成两点：
	- 一是，每个实例之间会按照一定的频率，从集群中随机挑选一些实例，把PING消息发送给挑选出来的实例，用来检测这些实例是否在线，并交换彼此的状态信息。PING消息中封装了发送消息的实例自身的状态信息、部分其他示例的状态信息，以及slot映射表。
	- 二是，一个实例在接收到PING消息后，会给发送PING消息的实例，发送一个PONG消息。PONG消息包含的内容和PING消息一样。  

下图显示了两个实例间进行PING、PONG消息传递的情况，其中实例一为发送节点，实例二是接收节点:  
![ping-pong](img/redis-cluster-node-swap.png)  

#### 新节点上线  
Redis Cluster加入新节点时，客户端需要执行Cluster MEET命令，如图所示：  
![msg-type-meet](img/redis-cluster-meet.png)  

- 步骤如下：
	- 节点1在执行CLUSTER MEET命令时会首先为新节点创建一个clusterNode数据，并将其添加到自己维护的clusterState和nodes字典中。有关clusterState和clusterNode关系，我们在最后一节会有详尽的示意图和源码来讲解。  

	- 然后节点1会根据CLUSTER MEET命令中的IP地址和端口号，向新节点发送一条MEET消息。新节点接收到节点一发送的MEET消息后，新节点也会为节点一创建一个clusterNode结构，并将该结构添加到自己维护的clusterState的nodes字典中。  

	- 接着，新节点向节点1返回一条PONG消息。节点1收到新节点B返回的PONG消息后，得知新节点已经成功的接收到了自己发送的MEET消息。  

	- 最后，节点1还会向节点B发送一条PING消息。新节点接收到该条PING消息后，可以知道节点1已经成功收到了自己返回的PONG消息，从而完成了新节点接入的我收操作。

MEET操作成功之后，节点1会通过稍早时讲的定时PING机制将新节点的信息发送给集群中的其他节点，让其他节点也与新节点进行握手，最终，经过一段时间后，新节点会被集群中的所有节点认识。  

#### **节点疑似下线和真正下线**  

Redis Cluster中的节点会定期检查已经发送PING消息的接收方节点是否在规定时间（cluster-node-timeout）内返回PONG消息，如果没有则会将其标记为疑似下线状态，也就是PFAIL状态，如下图：  

![msg-type-pfail](img/redis-cluster-pfail.png)  

然后节点1会通过PING消息，将节点2处于疑似下线状态的信息传递给其他节点，例如节点3。节点3接收到节点1的PING消息得知节点2进入PFAIL状态后，会在自己维护的clusterState的nodes字典中找到节点2所对应的clusterNode结构，并将煮节点1的`下线报告`添加到clusterNode结构的fail_reports链表中。  


![msg-type-pfail2](img/redis-cluster-fail.png)  

随着时间的推移，如果节点10（举个例子）也因为PONG超时而认为节点2疑似下线了，并且发现自己维护的节点2的clusterNode的fail_reports中有**半数以上的主节点数量的未过时的节点2标记为PFAIL状态报告日志**，那么节点10将会把节点2标记为已下线FAIL状态，并且节点10会**立刻**向集群其他节点广播主节点2已下线的FAIL消息，所有收到FAIL消息的节点都会立即将节点2状态标记为已下线。如下图所示：  

![msg-type-fail](img/redis-cluster-fail-2.png)  

需要注意的是，报告疑似下线记录是有时效性的，如果超时cluster-node-timeout\*2的时间，这个报告就会被忽略掉，让节点2又恢复正常状态。  


#### Redis Cluster通信源码实现  

综上所述，我们了解了RedisCluster在定时PING/PONG、新节点上线、节点疑似下线和真正下线等环节的原理和操作流程，下面我们来真正看一下Redis在这些环节的源码实现和具体操作。  


#### **设计的数据结构体**  

首先，我们先来讲解一下其中设计的数据结构，也就是上文中提到的clusterNode等结构。  

**每个节点都会维护一个clusterState结构**，表示当前集群的状态，它的定义如下：  
```c
typedef struct clusterState{
	clusterNode *myself; /* 当前节点的clusterNode信息 */
	...
	dict *nodes; /* name到clusterNode的字典 */
	...
	clusterNode *slots[CLUSTER_SLOTS]; /* slot和节点的对应关系 */
	...
} clusterState;
```
- 它有三个比较关键的字段，具体示意图如下所示：  
	- myself字段，是一个clusterNode结构，用来记录自身的状态；  
	- nodes字典，记录一个name到clusterNode结构的映射，以此来记录其他节点的状态；  
	- slot数组，记录slot对应的节点clusterNode结构。  

![redis-cluster-state](img/redis-cluster-state.png)  


clusterNode结构**保存了一个节点的当前状态**，比如**节点的创建时间、节点的名字、当前配置纪元、节点的IP地址和端口号等等**。除此之外，clusterNode结构的link属性是一个clusterLink结构，该结构保存了连接节点需要的有关信息，**比如套接字描述符**，输入缓冲区和输出缓冲区。clusterNode还有一个fail_report的列表，用来记录意思下线报告。具体定义如下：  
```c
typedef struct clusterNode {
	mstime_t ctime; /* 创建节点的时间 */
	char name[CLUSTER_NAMELEN]; /* 节点名称 */
	int flags; /* 节点标识，标记节点角色或者状态，比如著丛节点丛节点或者在线和下线 */
	uint64_t configEpoch; /* 当前节点已知的集群统一epoch */
	unsigned char slots[CLUSTER_SLOTS/8]; /* slots handled by this node */
	int numslots;   /* Number of slots handled by this node */
    int numslaves;  /* Number of slave nodes, if this is a master */
    struct clusterNode **slaves; /* pointers to slave nodes */
    struct clusterNode *slaveof; /* pointer to the master node. Note that it
                                    may be NULL even if the node is a slave
                                    if we don\'t have the master node in our
                                    tables. */
    mstime_t ping_sent; /* 当前节点最后一次向该节点发送PING消息的时间 */
    mstime_t pong_received; /* 当前节点最后一次收到该节点PONG消息的时间 */
    mstime_t fail_time; /* FAIL 标志位被设置的时间 */
    mstime_t voted_time;     /* Last time we voted for a slave of this master */
    mstime_t repl_offset_time;  /* Unix time we received offset for this node */
    mstime_t orphaned_time;     /* Starting time of orphaned master condition */
	long long repl_offset; /* 当前节点的repl偏移 */
	char ip[NET_IP_STR_LEN]; /* 当前节点的IP地址 */
	int port; /* 端口 */
	int cport; /* 通信端口，一般是端口+1000 */
	clusterLink *link; /* 和该节点的TCP链接 */
	list *fail_reoprts; /* 下线记录列表 */
}clusterNode;
```  

clusterNodeFailReport 是记录节点下线报告的结构体， node 是报告节点的信息，而 time 则代表着报告时间。

```c
typedef struct clusterNodeFailReport{
	struct clusterNode *node /* 报告当前节点已下线的节点 */
	mstime_t time; /* 报告时间 */
}clusterNodeFailReport;
```

#### 消息结构体  

了解了Redis节点维护的数据结构题后，我们再来看看节点进行通信的消息结构体。通信消息最外侧的结构体未clusterMsg，它包括了很多消息记录信息，包括RCmb标识位，消息总长度，消息协议版本，消息类型；它还包含了发送该消息节点的记录信息，比如节点名称，节点负责的slot信息，节点ip和端口灯；最后它还包含了一个clusterMsgData来携带具体类型的消息。  
```c
typedef struct {
    char sig[4];        /* 标志位，"RCmb" (Redis Cluster message bus). */
    uint32_t totlen;    /* 消息总长度 */
    uint16_t ver;       /* 消息协议版本 */
    uint16_t port;      /* 端口 */
    uint16_t type;      /* 消息类型 */
    uint16_t count;     /*  */
    uint64_t currentEpoch;  /* 表示本节点当前记录的整个集群的统一的epoch，用来决策选举投票等，与configEpoch不同的是：configEpoch表示的是master节点的唯一标志，currentEpoch是集群的唯一标志。 */
    uint64_t configEpoch;   /* 每个master节点都有一个唯一的configEpoch做标志，如果和其他master节点冲突，会强制自增使本节点在集群中唯一 */
    uint64_t offset;    /* 主从复制偏移相关信息，主节点和从节点含义不同 */
    char sender[CLUSTER_NAMELEN]; /* 发送节点的名称 */
    unsigned char myslots[CLUSTER_SLOTS/8]; /* 本节点负责的slots信息,16384/8个char数组，一共为16384bit */
    char slaveof[CLUSTER_NAMELEN]; /* master信息，假如本节点是slave节点的话，协议带有master信息 */
    char myip[NET_IP_STR_LEN];    /* IP */
    char notused1[34];  /* 保留字段 */
    uint16_t cport;      /* 集群的通信端口 */
    uint16_t flags;      /* 本节点当前的状态，比如 CLUSTER_NODE_HANDSHAKE、CLUSTER_NODE_MEET */
    unsigned char state; /* Cluster state from the POV of the sender */
    unsigned char mflags[3]; /* 本条消息的类型，目前只有两类：CLUSTERMSG_FLAG0_PAUSED、CLUSTERMSG_FLAG0_FORCEACK */
    union clusterMsgData data;
} clusterMsg;
```

clusterMsgData 是一个 union 结构体，它可以为 PING，MEET，PONG 或者 FAIL 等消息体。其中当消息为 PING、MEET 和 PONG 类型时，ping 字段是被赋值的，而是 FAIL 类型时，fail 字段是被赋值的。  


```c
// 注意这是 union 关键字
union clusterMsgData {
    /* PING, MEET 或者 PONG 消息时，ping 字段被赋值 */
    struct {
        /* Array of N clusterMsgDataGossip structures */
        clusterMsgDataGossip gossip[1];
    } ping;
    /*  FAIL 消息时，fail 被赋值 */
    struct {
        clusterMsgDataFail about;
    } fail;
    // .... 省略 publish 和 update 消息的字段
};
```
clusterMsgDataGossip 是 PING、PONG 和 MEET 消息的结构体，它会包括发送消息节点维护的其他节点信息，也就是上文中 clusterState 中 nodes 字段包含的信息，具体代码如下所示，你也会发现二者的字段是类似的。

```c
typedef struct {
    /* 节点的名字，默认是随机的，MEET消息发送并得到回复后，集群会为该节点设置正式的名称*/
    char nodename[CLUSTER_NAMELEN]; 
    uint32_t ping_sent; /* 发送节点最后一次给接收节点发送 PING 消息的时间戳，收到对应 PONG 回复后会被赋值为0 */
    uint32_t pong_received; /* 发送节点最后一次收到接收节点发送 PONG 消息的时间戳 */
    char ip[NET_IP_STR_LEN];  /* IP address last time it was seen */
    uint16_t port;       /* IP*/       
    uint16_t cport;      /* 端口*/  
    uint16_t flags;      /* 标识*/ 
    uint32_t notused1;   /* 对齐字符*/
} clusterMsgDataGossip;

typedef struct {
    char nodename[CLUSTER_NAMELEN]; /* 下线节点的名字 */
} clusterMsgDataFail;
```  

看完了节点维护的数据结构体和发送的消息结构体后，我们就来看看 Redis 的具体行为源码了。  

#### 随机周期性发送PING消息

Redis 的CclusterCron函数会被定时调用，每被执行10次，就会准备向随机的一个节点发送PING消息。  

它会先随机的选出5个节点，然后从中选择最久没有与之通信的节点，调用clusterSendPing函数发送类行为CLUSTERMSG_TYPE_PING的消息。  

```c
// cluster.c 文件
// clusterCron()没执行10次（至少间隔一秒钟），就像一个随机节点发送gossip信息
if(!(iteration % 10)){
	int j;

	/* 随机5个节点，选出其中一个 */
	for(j = 0; j < 5; j++) {
		de = dictGetRandomKey(server.cluster->nodes);
		clusterNode *this = dictGetVal(de);

		/* 不要PING链接断开的节点，也不要PING最近已经PING过的节点 */
		if(this->link == NULL || this->ping_send != 0) continue;
		if(this->flags & (CLUSTER_NODE_MYSELF|CLUSTER_NODE_HANDSHAKE))
			continue;
		/* 对比pong_received字段，选出更长时间未收到其PONG消息的节点（表示好久没有接收到该节点的PONG消息了） */
		if(min_pong_node == NULL || min_pong > this->pong_received) {
			min_pong_node = this;
			min_pong = this->pong_received;
		}
	}
	/* 向最久没有收到PONG回复的节点发送PING命令 */
	if (min_pong_node) {
		serverLog(LL_DEBUG, "Pinging node %.40s", min_pong_node->name)
		clusterSendPing(min_pong_node->link, CCLUSTERMSG_TYPE_PING);
	}

}
```  
clusterSendPing函数的具体行为我们后续再了解，因为该函数在其他环节也会经常用到。  

### 节点加入集群  
当节点执行CLUSTER MEET命令后，会在自身给新节点维护一个clusterNode结构体，该结构体的link也就是TCP链接字段是null，表示是新节点尚未建立连接。  

clusterCron函数也会处理这些为建立连接的新节点，调用createClusterLink创立连接，然后调用clusterSendPing函数来发送MEET消息。  

```c
/* cluster.c clusterCron 函数部分，为未创建连接的节点创建连接 */
if(node->link == NULL){
	int fd;
	mstime_t old_ping_sent;
	clusterLink *link;
	/* 和该节点建立连接 */
	fd = anetTcpNonBlockBindConnect(server.neteer, node->ip, node->cport, NET_FIRST_BING_ADDR);

	/* ...fd为-1时的异常处理 */
	/* 建立link */
	link = createClusterLink(node);
	link->fd = fd;
	node->link = link;
	aeCreateFileEvent(server.el, link->fd, AE_READABLE, clusterReadHandler, link);
	/* 向新连接的节点发送PING命令，防止节点被标识进入下线 */
	/* 如果节点被标记为MEET，那么发送MEET命令，否则发送PING命令 */
	old_ping_sent = node->ping_sent;
	clusterSendPing(link, node->flags & CLUSTER_NODE_MEET ? CLUSTER_TYPE_MEET:CLUSTERMSG_TYPE_PING);
	/* .... */
    /* 如果当前节点（发送者）没能收到 MEET 信息的回复，那么它将不再向目标节点发送命令。*/
    /* 如果接收到回复的话，那么节点将不再处于 HANDSHAKE 状态，并继续向目标节点发送普通 PING 命令*/
    node->flags &= ~CLUSTER_NODE_MEET;
}
```

##### 防止节点假超时及状态过期  

防止节点假超时和标记疑似下线标记也是在 clusterCron 函数中，具体如下所示。它会检查当前所有的 nodes 节点列表，如果发现某个节点与自己的最后一个 PONG 通信时间超过了预定的阈值的一半时，为了防止节点是假超时，会主动释放掉与之的 link 连接，然后会主动向它发送一个 PING 消息。

```c
/* cluster.c clusterCron 函数部分，遍历节点来检查 fail 的节点*/
while((de = dictNext(di)) != NULL) {
    clusterNode *node = dictGetVal(de);
    now = mstime(); /* Use an updated time at every iteration. */
    mstime_t delay;

    /* 如果等到 PONG 到达的时间超过了 node timeout 一半的连接 */
    /* 因为尽管节点依然正常，但连接可能已经出问题了 */
    if (node->link && /* is connected */
        now - node->link->ctime >
        server.cluster_node_timeout && /* 还未重连 */
        node->ping_sent && /* 已经发过ping消息 */
        node->pong_received < node->ping_sent && /* 还在等待pong消息 */
        /* 等待pong消息超过了 timeout/2 */
        now - node->ping_sent > server.cluster_node_timeout/2)
    {
        /* 释放连接，下次 clusterCron() 会自动重连 */
        freeClusterLink(node->link);
    }

    /* 如果目前没有在 PING 节点*/
    /* 并且已经有 node timeout 一半的时间没有从节点那里收到 PONG 回复 */
    /* 那么向节点发送一个 PING ，确保节点的信息不会太旧，有可能一直没有随机中 */
    if (node->link &&
        node->ping_sent == 0 &&
        (now - node->pong_received) > server.cluster_node_timeout/2)
    {
        clusterSendPing(node->link, CLUSTERMSG_TYPE_PING);
        continue;
    }
    /* .... 处理failover和标记遗失下线 */
}
```  

##### 处理failover和标记疑似下线  

如果防止节点假超时处理后，节点依旧未收到目标节点的 PONG 消息，并且时间已经超过了 cluster_node_timeout，那么就将该节点标记为疑似下线状态。
```c
/* 如果这是一个主节点，并且有一个从服务器请求进行手动故障转移,那么向从服务器发送 PING*/
if (server.cluster->mf_end &&
    nodeIsMaster(myself) &&
    server.cluster->mf_slave == node &&
    node->link)
{
    clusterSendPing(node->link, CLUSTERMSG_TYPE_PING);
    continue;
}

/* 后续代码只在节点发送了 PING 命令的情况下执行*/
if (node->ping_sent == 0) continue;

/* 计算等待 PONG 回复的时长 */ 
delay = now - node->ping_sent;
/* 等待 PONG 回复的时长超过了限制值，将目标节点标记为 PFAIL （疑似下线)*/
if (delay > server.cluster_node_timeout) {
    /* 超时了，标记为疑似下线 */
    if (!(node->flags & (REDIS_NODE_PFAIL|REDIS_NODE_FAIL))) {
        redisLog(REDIS_DEBUG,"*** NODE %.40s possibly failing",
            node->name);
        // 打开疑似下线标记
        node->flags |= REDIS_NODE_PFAIL;
        update_state = 1;
    }
}


```  

#### 实际发送Gossip消息  
以下是前方多次调用过的clusterSendPing()方法的源码，代码中有详细的注释，大家可以自行阅读。主要的操作就是将节点自身维护的 clusterState 转换为对应的消息结构体。  

```c
/* 向指定节点发送一条 MEET 、 PING 或者 PONG 消息 */
void clusterSendPing(clusterLink *link, int type) {
    unsigned char *buf;
    clusterMsg *hdr;
    int gossipcount = 0; /* Number of gossip sections added so far. */
    int wanted; /* Number of gossip sections we want to append if possible. */
    int totlen; /* Total packet length. */
    // freshnodes 是用于发送 gossip 信息的计数器
    // 每次发送一条信息时，程序将 freshnodes 的值减一
    // 当 freshnodes 的数值小于等于 0 时，程序停止发送 gossip 信息
    // freshnodes 的数量是节点目前的 nodes 表中的节点数量减去 2 
    // 这里的 2 指两个节点，一个是 myself 节点（也即是发送信息的这个节点）
    // 另一个是接受 gossip 信息的节点
    int freshnodes = dictSize(server.cluster->nodes)-2;


    /* 计算要携带多少节点的信息，最少3个，最多 1/10 集群总节点数量*/
    wanted = floor(dictSize(server.cluster->nodes)/10);
    if (wanted < 3) wanted = 3;
    if (wanted > freshnodes) wanted = freshnodes;

    /* .... 省略 totlen 的计算等*/

    /* 如果发送的信息是 PING ，那么更新最后一次发送 PING 命令的时间戳 */
    if (link->node && type == CLUSTERMSG_TYPE_PING)
        link->node->ping_sent = mstime();
    /* 将当前节点的信息（比如名字、地址、端口号、负责处理的槽）记录到消息里面 */
    clusterBuildMessageHdr(hdr,type);

    /* Populate the gossip fields */
    int maxiterations = wanted*3;
    /* 每个节点有 freshnodes 次发送 gossip 信息的机会
       每次向目标节点发送 2 个被选中节点的 gossip 信息（gossipcount 计数） */
    while(freshnodes > 0 && gossipcount < wanted && maxiterations--) {
        /* 从 nodes 字典中随机选出一个节点（被选中节点） */
        dictEntry *de = dictGetRandomKey(server.cluster->nodes);
        clusterNode *this = dictGetVal(de);

        /* 以下节点不能作为被选中节点：
         * Myself:节点本身。
         * PFAIL状态的节点
         * 处于 HANDSHAKE 状态的节点。
         * 带有 NOADDR 标识的节点
         * 因为不处理任何 Slot 而被断开连接的节点 
         */
        if (this == myself) continue;
        if (this->flags & CLUSTER_NODE_PFAIL) continue;
        if (this->flags & (CLUSTER_NODE_HANDSHAKE|CLUSTER_NODE_NOADDR) ||
            (this->link == NULL && this->numslots == 0))
        {
            freshnodes--; /* Tecnically not correct, but saves CPU. */
            continue;
        }

        // 检查被选中节点是否已经在 hdr->data.ping.gossip 数组里面
        // 如果是的话说明这个节点之前已经被选中了
        // 不要再选中它（否则就会出现重复）
        if (clusterNodeIsInGossipSection(hdr,gossipcount,this)) continue;

        /* 这个被选中节点有效，计数器减一 */
        clusterSetGossipEntry(hdr,gossipcount,this);
        freshnodes--;
        gossipcount++;
    }

    /* .... 如果有 PFAIL 节点，最后添加 */


    /* 计算信息长度 */
    totlen = sizeof(clusterMsg)-sizeof(union clusterMsgData);
    totlen += (sizeof(clusterMsgDataGossip)*gossipcount);
    /* 将被选中节点的数量（gossip 信息中包含了多少个节点的信息）记录在 count 属性里面*/
    hdr->count = htons(gossipcount);
    /* 将信息的长度记录到信息里面 */
    hdr->totlen = htonl(totlen);
    /* 发送网络请求 */
    clusterSendMessage(link,buf,totlen);
    zfree(buf);
}


void clusterSetGossipEntry(clusterMsg *hdr, int i, clusterNode *n) {
    clusterMsgDataGossip *gossip;
    /* 指向 gossip 信息结构 */
    gossip = &(hdr->data.ping.gossip[i]);
    /* 将被选中节点的名字记录到 gossip 信息 */   
    memcpy(gossip->nodename,n->name,CLUSTER_NAMELEN);
    /* 将被选中节点的 PING 命令发送时间戳记录到 gossip 信息 */
    gossip->ping_sent = htonl(n->ping_sent/1000);
    /* 将被选中节点的 PONG 命令回复的时间戳记录到 gossip 信息 */
    gossip->pong_received = htonl(n->pong_received/1000);
    /* 将被选中节点的 IP 记录到 gossip 信息 */
    memcpy(gossip->ip,n->ip,sizeof(n->ip));
    /* 将被选中节点的端口号记录到 gossip 信息 */
    gossip->port = htons(n->port);
    gossip->cport = htons(n->cport);
    /* 将被选中节点的标识值记录到 gossip 信息 */
    gossip->flags = htons(n->flags);
    gossip->notused1 = 0;
}
```


下面是 clusterBuildMessageHdr 函数，它主要负责填充消息结构体中的基础信息和当前节点的状态信息。  

```c
/* 构建消息的 header */
void clusterBuildMessageHdr(clusterMsg *hdr, int type) {
    int totlen = 0;
    uint64_t offset;
    clusterNode *master;

    /* 如果当前节点是salve，则master为其主节点，如果当前节点是master节点，则master就是当前节点 */
    master = (nodeIsSlave(myself) && myself->slaveof) ?
              myself->slaveof : myself;

    memset(hdr,0,sizeof(*hdr));
    /* 初始化协议版本、标识、及类型， */
    hdr->ver = htons(CLUSTER_PROTO_VER);
    hdr->sig[0] = 'R';
    hdr->sig[1] = 'C';
    hdr->sig[2] = 'm';
    hdr->sig[3] = 'b';
    hdr->type = htons(type);
    /* 消息头设置当前节点id */
    memcpy(hdr->sender,myself->name,CLUSTER_NAMELEN);

    /* 消息头设置当前节点ip */
    memset(hdr->myip,0,NET_IP_STR_LEN);
    if (server.cluster_announce_ip) {
        strncpy(hdr->myip,server.cluster_announce_ip,NET_IP_STR_LEN);
        hdr->myip[NET_IP_STR_LEN-1] = '\0';
    }

    /* 基础端口及集群内节点通信端口 */
    int announced_port = server.cluster_announce_port ?
                         server.cluster_announce_port : server.port;
    int announced_cport = server.cluster_announce_bus_port ?
                          server.cluster_announce_bus_port :
                          (server.port + CLUSTER_PORT_INCR);
    /* 设置当前节点的槽信息 */
    memcpy(hdr->myslots,master->slots,sizeof(hdr->myslots));
    memset(hdr->slaveof,0,CLUSTER_NAMELEN);
    if (myself->slaveof != NULL)
        memcpy(hdr->slaveof,myself->slaveof->name, CLUSTER_NAMELEN);
    hdr->port = htons(announced_port);
    hdr->cport = htons(announced_cport);
    hdr->flags = htons(myself->flags);
    hdr->state = server.cluster->state;

    /* 设置 currentEpoch and configEpochs. */
    hdr->currentEpoch = htonu64(server.cluster->currentEpoch);
    hdr->configEpoch = htonu64(master->configEpoch);

    /* 设置复制偏移量 */
    if (nodeIsSlave(myself))
        offset = replicationGetSlaveOffset();
    else
        offset = server.master_repl_offset;
    hdr->offset = htonu64(offset);

    /* Set the message flags. */
    if (nodeIsMaster(myself) && server.cluster->mf_end)
        hdr->mflags[0] |= CLUSTERMSG_FLAG0_PAUSED;

    /* 计算并设置消息的总长度 */
    if (type == CLUSTERMSG_TYPE_FAIL) {
        totlen = sizeof(clusterMsg)-sizeof(union clusterMsgData);
        totlen += sizeof(clusterMsgDataFail);
    } else if (type == CLUSTERMSG_TYPE_UPDATE) {
        totlen = sizeof(clusterMsg)-sizeof(union clusterMsgData);
        totlen += sizeof(clusterMsgDataUpdate);
    }
    hdr->totlen = htonl(totlen);
}
```  

- 原文：https://zhuanlan.zhihu.com/p/328728595









