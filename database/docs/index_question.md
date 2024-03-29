### 索引相关问题

### 数据库支持哈希索引嘛？

- 分析 
很少有面试官会在数据库面试里边了哈希索引，因为这个东西很罕见，用法也比较诡异。不过这也是一个优雅的装逼点  

- 答案
哈希索引是利用哈希表来实现的，适用于等值查询，如等于，不等于，in等，对范围查询是不支持的。我们惯用的innodb引擎是不支持用户自定义哈希索引的，但是innodb有一个优化会建立自适应哈希索引。所谓的自适应哈希索引，是指innodb引擎，如果发现二级索引（除了主键以外的别的索引）被经常食用，那么innodb会给这个索引建立一个哈希索引，加快查询。所以从本质上说innodb的自适应还索引是一个对索引的哈希索引。 

- 关键：等值查询，对索引的哈希索引。

- 如何引导：  
在前面回答了哈希索引之后直接调过来这里，例如“哈希索引在KV数据库上比较常见，不过innodb引擎支持自适应哈希索引，它是...”


#### 聚簇索引和非聚簇索引的区别  
- 分析： 
这其实是一个很简单的问题，但也是一个很能装逼的问题。聚簇索引和非聚簇索引的区别，只需要回答，他们叶子节点是否存储了数据。但是要打出两点，就要多回答两个点：第一：Mysql的非聚簇索引存储了主键；第二覆盖索引不需要回表。  
- 答案：
聚簇索引是指叶子结点存储了数据的索引。MYSQL整张表可以看作是一个聚簇索引。因为非聚簇索引没有存储数据，所以一般是存储了主键。于是会导致一个回表的问题。即如果我们查询的列包含不再索引上的列，这会引起数据库先根据非聚簇索引找到主键，然后拿着主键去聚簇索引里面捞出来数据。而根据主键找到数据会引起磁盘IO，性能大幅度下降。这就是我们推荐使用覆盖索引的原因。  
- 关键点：聚簇索引存储了数据，非聚簇索引要回表。

- **如何引导过这里？**。
	- 聊到了覆盖索引与回表问题，话术可以是“一般用覆盖索引，在不使用覆盖索引的时候，会引起回表查询，这是因为MySQL的非聚簇索引。。。”
	- 聊到如何计算一次查询的开销，这个比阿吉哦啊少见，因为一般面试官也讲不清楚一次Mysql查询时间开销会在哪里；
	- 前面基本回答，回答了聚簇索引之后直接回答这部分
	- 聊到B+树的叶子结点可以存放什么，或者聊到索引的叶子节点可以存放什么
	- 是不是查询一定会引起回表？这其实是考察覆盖索引，所以在谈及覆盖索引之后可以聊聊聚簇索引和非聚簇索引点

#### MySQL为什么使用B+树索引
- 分析
实际上是为了考察数据结构，B+树的特征，而且能根据B+树的特征，理解Mysql选择B+树的原因。面试官可能同时希望你能够横向比较B+树、B树、平衡二叉树，红黑树和跳表，志杰背这几种树的基本特征是比较难得，索引我们可以只回答关键点。关键点有三个：**和二叉树比起来，B+树是多叉的，高度低。和B树比起来，它的叶子节点组成一个链表。第三点是一个角度很清奇低点，就是查询时间稳定性，可以测性。和跳表的比较比较诡异，要从Mysql组织B+树的角度出发**。
- 答案：Mysql使用B+树就考虑三个角度：  
	- 和二叉树，如平衡术，红黑树比起来，B+树是多叉树，比如Mysql默认是1200叉树，同样的数据量，高度要比二叉树低；
	- 和B树相比，B+树的叶子节点被连接起来形成一个链表，这意味着，当我们执行范围查询的时候，Mysql可以利用这个特性，沿着叶子节点前进。而之索引Nosql数据库会使用B树作为索引，是因为他们不像关系型数据库那样大量查询都是范围查询。
	- B+树之灾叶子节点存放数据，和B树比起来，查询时间稳定可预测。（注：这是一个高级观点，就是在工程实践中，我们可能倾向于追求一种稳定可预测，而不是某些数据贼快，某些数据唰一下贼慢）
	- B+树和跳表比起来，Mysql将B+树节点设置为磁盘大小，这样可以充分利用Mysql的预加载机制，减少磁盘IO

- 关键点：
高度低，叶子节点是链表，查询时间可预测，节点大小等于磁盘页大小。

- 如何引导过来？  
	- 面试官直接问起来
	- 你们聊起数据结构，聊到B树和B+树，话术一般是“因为B+树和B树比起来，有。。。优点，Mysql索引主要是使用B+树的”
	- 聊到反胃查询或者全表扫描，你可以从B+树的角度来说，这种扫描利用到了B+树叶子节点是链表特征

#### 为什么使用自增主键
- 分析：
这是一个常考的点，从根源上来说，是为了考察你对数据库如何组织数据的理解。问题在于，数据库如何组织数据其实是一个很难得问题，所以一般情况下，不需要回答道非常底层的地步。  
- 答案：
Mysql的innodb的主键是一个聚簇索引，即它的叶子节点存放了数据。在使用自增主键的情况下，会保证树的分裂是按照单向分裂的，这会大概率导致物理页的分裂也是朝着单方向进行的，即连续的。在不使用自增主键的情况下，如果已经满的页里插入，会导致MYSQL页分裂，虽然逻辑上页依旧是连续的，但是物理页已经不连续了。如果在使用机械硬盘的情况下，会导致范围查找经常导致机械硬盘重新定位，性能差。

- 关键点：
	单方向增长，物理页连续。

- 如何引导： 
	- 面试官直接问你
	- 在基础回答聚簇索引的时候主动说起，为什么我们要使用自增索引。话术可能是“MYSQL的主键索引是聚簇索引，每张表一个，所以我们一般推荐使用自增主键，因为自增主键会保证树单向分裂”
	- 聊起数据库表结构设计，你们公司推荐使用自增主键，可以主动说，我们公司强制要求使用自增主键，因为。。。
	- 聊起数据库表结构设计，你有一些特殊的表，没有使用自增主键，你可以说，“我们大多数表都是使用自增主键的，因为。。。。但是这几张表我们没有使用，因为xxx结合你们的业务特征回答”慎用
	- 聊到树结构的特征。比如面试官其实面你的数据结构，而不是数据库，但是你们聊到了树，就可以主动提起。因为大部分树，比如说红黑树，二叉平衡树，B树，B+树都有一个调整树结构的过程，所以可以强行引过来
	- 聊起分库分表设计，主键生成的时候，可以提起生成主键为什么最好是单调递。这个问题其实和为什么使用自增主键其实是同一个问题
	- 评价为什么使用uuid作为主键生成策略会很糟糕


#### 索引有什么缺点
- 分析：
	面试官是为了吓你，出其不意攻其不备。又或者面试官问为啥不在所有列上建立索引，或者问为什么不建立多一点索引
- 答案：
	索引的维护是有开销的。在增改数据的时候，数据库都要对应修改索引；如果索引过多，以至于内存没办法装下全部索引，那么会导致索引本身都会触发IO。索引索引不是越多越好。比如为了避免数据量过大，某些时候我们会使用前缀索引。 


#### 什么是索引下推
- 分析：
	这个题更加无聊，因为它对于你的实际工作帮助可以说没有了。前面那些点理解清楚，还可以说有助于自己设计索引，这个就可以说，完全没用。回答这个问题的关键点在于，要和联合索引、覆盖索引一起讨论。因为他们体现的都是一个东西：即尽量利用索引数据，避免回表。

- 答案：
	索引下推是指将于索引有关的条件由MySQL服务器下推到引擎。例如按照名字存取姓张的，like "张%"。在原来没有索引下推的时候，即便在用户名字上建立了索引，但是还是不能利用这个索引。而在支持索引下推的引擎上，引擎就可以利用名字索引，将数据提前过滤，避免回表。目前innodb引擎和MyISAM都支持索引下推。索引下推和覆盖索引的理念都是一致的，尽量避免回表。

#### **使用索引了为什么还是很慢**
- 分析：
	这又是属于违背直觉的问题，本质上考察的是你对索引，和mysql执行过程的理解。记住一个核心点，**索引只能帮你快速定位数据**，而定位到数据之后的事情，比如说把读数据，写数据 这都是需要时间的。尤其是要考虑食物机制，竞争锁的问题。

- 答案：
	索引只能帮助定位数据，但是**从索引定位到数据，到返回结果，或者更新数据都是需要时间的**。尤其是在事物中，索引定位到数据之后，可能一直在等待锁。如果别的事物执行时间缓慢，那么即使你用了索引，这一次的查询还是很慢。本质上是因为，mysql的执行速度是收到很多因素影响的，准确来说，索引只是大概率能够加速这个过程而已。

	另外要考虑，数据库是否使用错了索引。如果我们的表上创建了多个索引，那么就会导致mysql选择使用了不那么恰当的索引。在这种时候我们可以通过数据库的Hint机制提示数据库走某个索引。 

- 关键字： 锁竞争
- 类似问题：
	- 为什么我定义了索引，查询还是很慢？
	- 这个问题有一个陷阱，即他没有说我用到了索引，也就是说，你定义了索引，但可能是mysql没用，也可能用了，但是卡在锁竞争哪里

- 如何引导：
	在前面聊到使用索引来优化的时候，可以提一嘴这个，即并不是说使用了索引就肯定很快

#### 什么时候索引会失效？
- 分析：
	索引失效这个说法有点误导人，准确的说法是，为什么我明明定义了索引，但是mysql却没有使用索引？关键点是权衡，即下面的第二个理由

- 答案：没有使用索引主要有两大类原因，一个是sql没写好
	- 索引在列上做了计算
	- like关键字使用了前缀匹配，例如“%ab”，注意的是，后缀撇配是可以使用索引
	- 字符串没有引号导致类型转换
	另一种，则是mysql判断使用索引的代价很高，比如说要全索引扫描并且回表，那么就会退化成全表扫描，数据库数据量的大小和数据分布，会影响到mysql决策。
- 类似问题：
	- 为什么我定义了索引查询还是很慢？没用or锁竞争
	- 为什么我定义了索引，mysql却不用
	



