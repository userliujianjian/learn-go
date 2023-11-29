## SQL优化之最左侧匹配原则

### 介绍
项目监控中，时常能看到数据库慢日志。通过这篇文章我来讲讲其中一个原因，表创建了索引，但是查询没有匹配到索引的原因。

- 清单1(数据准备)
```SQL
CREATE TABLE `test_user` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '名称',
  `age` int(3) NOT NULL DEFAULT '0' COMMENT '年龄',
  `sex` tinyint(1) NOT NULL DEFAULT '0' COMMENT '性别（0:未知 1:男 2:女）',
  `bio` varchar(255) NOT NULL DEFAULT '' COMMENT '个人简介',
  `mobile` varchar(20) NOT NULL DEFAULT '' COMMENT '手机号码',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态(1：正常 2:禁用 )',
  PRIMARY KEY (`id`),
  KEY `age_sex_status_index` (`age`,`sex`,`status`) USING BTREE COMMENT '手机年龄索引'
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4;

// 插入数据
INSERT INTO `zhiqun`.`test_user` (`id`, `name`, `age`, `sex`, `bio`, `mobile`, `status`) VALUES (1, 'jack', 20, 1, '', '18200000000', 1);
INSERT INTO `zhiqun`.`test_user` (`id`, `name`, `age`, `sex`, `bio`, `mobile`, `status`) VALUES (2, 'rose', 30, 2, '', '18300000000', 1);


```

看清单1，知道创建表test_user, 包含联合索引age_sex_status。那我们来看清单2中的SQL是否能命中索引？为什么？
### 分析：  
单单回答最做前缀匹配原则是很简单，但是没有亮点。亮点在最左侧匹配大概是如何运作的。之所以只需要回答大概如何运作，是因为详细回答太难了，面试官没读过源码也搞不清楚，犯不着。

- 清单2（判断SQL是否命中索引）
```SQL
1. EXPLAIN SELECT * FROM test_user WHERE age = 30 AND sex =1 and `status` = 1;
2. EXPLAIN SELECT age, sex, status FROM test_user WHERE age = 30 AND sex =1 and `status` = 1;
3. EXPLAIN SELECT age, sex, status FROM test_user WHERE age > 30 AND sex =1 and `status` = 1;
4. EXPLAIN SELECT age, sex, status FROM test_user WHERE age = 30 AND sex > 1 and `status` = 1;
5. EXPLAIN SELECT age, sex, status FROM test_user WHERE age = 30 AND sex = 1 and `status` > 1;
6. EXPLAIN SELECT age, sex, status FROM test_user WHERE sex in (1, 2) and `status` > 1;


output:
select_type table		type		key						key_len		ref					rows		extra
SIMPLE		test_user	ref			age_sex_status_index	6			const,const,const	1		null
SIMPLE		test_user	ref			age_sex_status_index	6			const,const,const	1		using index
SIMPLE		test_user	index		age_sex_status_index	6				null			2		Using where; Using index
SIMPLE		test_user	ref			age_sex_status_index	4			const				1		Using where; Using index
SIMPLE		test_user	ref			age_sex_status_index	5			const,const			1		Using where; Using index
SIMPLE		test_user	index		age_sex_status_index	6								2	Using where; Using index

```

- **分析**
通过上面我们知道了，以上所有的sql都使用了联合索引age_sex_status_index， **那么为什么使用`>`范围查找后依然能命中索引呢？，而且索引长度并不只是age？**
  
  - SQL1与SQL2相差在查询字段`*`上，SQL2明显只使用了索引，SQL1虽然使用了索引，但是extra中为空，是因为有回表操作。
  - 范围查找依然能命中索引，是因为B+树叶子结点组成`链表`数据结构
  - 条件age范围查找后sex字段依然能使用索引，是因为联合索引底部其实有(a, ab, abc)好几个索引组成的，mysql索引列其实是一层套一层循环


- **答案**
联合索引的顺序很重要，因为mysql在索引中按照创建索引时的顺序进行存储。
第3条SQL中age范围查找，索引(age,sex,status)的左侧前缀就是(age).   
第4条sex范围查找条件命中索引第二个字段，最左侧前缀就是(age, sex).  
第5条status范围查找条件命中第三个字段，最左侧前缀就是(age, sex, status)
第6条sex in查找命中第二个字段，最左侧前缀索引就是(age, sex)

- 总结：  
最左前缀匹配原则是指，MySQL会按照联合索引创建的顺序，从左至右开始匹配。例如创建了一个联合索引（A，B，C)，那么本质上来说，是创建了A，（A，B），（A，B，C）三个索引。之所以如此，因为MySQL在使用索引的时候，类似于多重循环，一个列就是一个循环。在这种原则下，我们会优先考虑把区分度最好的放在最左边，而区分度可以简单使用不同值的数量除以总行数来计算（distinct(a, b, c)/count(*)）。


- 扩展（切记记不清楚就不要往这块引）：  
- Explain 参数：
	- select_type(查询类型) 常见值包括：
		- SIMPLE: 没有子查询或者union的简单查询
		- PRIMARY：JOIN中最外层的SELECT查询
		- SUBQUERY：FROM子句中的子查询
		- DERIVED: FROM子句中派生表
		- UNION：UNION中第二个或后续SELECT语句
	- table:表明
	- type 链接类型：
		- system：标识只有一行，例如系统目录
		- const：表最多只有一行匹配，使用唯一索引读取
		- eq_ref: 链接使用索引的所有部分，索引是主键或唯一非空索引
		- ref： 仅使用索引的一个子集，索引不是唯一的
		- range： 使用索引范围
		- index：扫描整个索引
		- all：全表扫描
	- possible_keys: 可能使用的索引
	- key：实际使用的索引
	- key_len: 实际使用索引的长度
	- ref：参考,与在key列中命名的索引进行比较的列
	- rows: 必须检查的行数
	- extra:额外信息
		- using index
		- using where
		- using temporary
		- using filesort


