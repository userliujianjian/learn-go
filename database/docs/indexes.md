### 数据库索引

#### 数据库索引有哪些？
- 分析：
	考察数据库索引基础，注意优先保证回答的完整性，再引导出其他问题。 索引可以从四个方面说（结构、形态、聚簇索引、索引与锁）
- 基本回答：
	- 从数据结构出发：
		- 树结构索引（B树、B+树、二叉树）
		- 哈希索引（innodb自适应哈希索引）
		- 位图
	- 从形态上分为（6中）：
		- 覆盖索引：指查询的列全部命中了索引。
		- 联合索引：包含多个表字段，识别度越高越靠前：distinct(a,b,c)/ count(id) 
		- 全文索引：现在比较少用推荐使用中间件
		- 前缀索引：指利用数据前几个字符的索引，如果前几个字符区分度不好的话不建议使用
		- 唯一索引：是指数据表里要求该索引值必须要唯一，常用于保障业务唯一性
		- 主键索引：比较特殊，它的叶子节点要么存储了数据，要么存储了指向数据的指针
	- 聚簇索引角度：
		- 聚簇索引：Mysql主键就是聚簇索引
		- 非聚簇索引：本质上存储的是主键
	而对于Mysql的innodb引擎来说，它的行锁是利用索引来实现的，如果查询的时候没有索引，那么会导致表锁（这一句可能引导面试官问锁和事物的问题，如果不熟悉锁和事物，请不要回答）

