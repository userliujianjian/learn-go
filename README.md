# learn-go
go初学者，练练手

# 第二周作业 Q1
> Q: 我们在数据库操作的时候，比如 dao 层中当遇到一个 sql.ErrNoRows 的时候，是否应该 Wrap 这个 error，抛给上层。为什么，应该怎么做请写出代码？

> A: 不需要把ErrNoRows返回给上一层  
> 1.ErrNoRows 查询结果为空，将业务数据置为空即可
> 2.dao层返回的Error与业务nil 彻底分开，返回的error更明确 容易判断

示例代码：
```
func (m *StaffRemarkUserModel) GetUserRemark(staffId int64, userId int64) (*po.StaffRemarkUser, error) {
	var detail po.StaffRemarkUser
	sq := m.Model.DB.Debug()
	sq = sq.Where("userId = ?", userId).Where("staffId = ?", staffId)
	sq = sq.First(&detail)
	if sq.Error != nil && sq.Error != gorm.ErrRecordNotFound {
		return nil, sq.Error
	}
	return &detail, nil
}
```
