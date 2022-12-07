# Page


simple page with [gorm](https://gorm.io/gorm), find multiple pieces of data is helpful.


## Usage


```bash
go get -u github.com/go-cinch/common/page
```


### Field


- `Num` - current page
- `Size` - page per count
- `Total` - all data count 
- `Disable` - disable pagination, query all data
- `Count` - not use 'SELECT count(*) FROM ...' before 'SELECT * FROM ...'
- `Primary` - When there is a large amount of data, limit is optimized by specifying a field (the field is usually self incremented ID or indexed), which can improve the query efficiency (if it is not transmitted, it will not be optimized)


### Find


#### count and data


```go
func (ro roleRepo) Find(ctx context.Context, condition *biz.FindRole) (rp []biz.Role) {
	db := ro.data.DB(ctx)
	db = db.
		Model(&Role{}).
		Order("id DESC")
	rp = make([]biz.Role, 0)
	condition.Page.
		WithContext(ctx).
		Query(db).
		Find(&list)
	copierx.Copy(&rp, list)
	return
}
```

sql log:  
```mysql
SELECT count(*) FROM `role`;
SELECT `role`.`id`,`role`.`name`,`role`.`word`,`role`.`action` FROM `role` ORDER BY id DESC LIMIT 20;
```


#### only data


```go
func (ro roleRepo) Find(ctx context.Context, condition *biz.FindRole) (rp []biz.Role) {
	db := ro.data.DB(ctx)
	db = db.
		Model(&Role{}).
		Order("id DESC")
	rp = make([]biz.Role, 0)
	condition.Page.Count = true
	condition.Page.
		WithContext(ctx).
		Query(db).
		Find(&list)
	copierx.Copy(&rp, list)
	return
}
```

sql log:
```mysql
SELECT `role`.`id`,`role`.`name`,`role`.`word`,`role`.`action` FROM `role` ORDER BY id DESC LIMIT 20;
```


#### all data


```go
func (ro roleRepo) Find(ctx context.Context, condition *biz.FindRole) (rp []biz.Role) {
	db := ro.data.DB(ctx)
	db = db.
		Model(&Role{}).
		Order("id DESC")
	rp = make([]biz.Role, 0)
	condition.Page.Disable = true
	condition.Page.
		WithContext(ctx).
		Query(db).
		Find(&list)
	copierx.Copy(&rp, list)
	return
}
```

sql log:
```mysql
SELECT `role`.`id`,`role`.`name`,`role`.`word`,`role`.`action` FROM `role` ORDER BY id DESC;
```


#### limit optimize


```go
func (ro roleRepo) Find(ctx context.Context, condition *biz.FindRole) (rp []biz.Role) {
	db := ro.data.DB(ctx)
	db = db.
		Model(&Role{}).
		Order("id DESC")
	rp = make([]biz.Role, 0)
	condition.Page.Primary = "id"
	condition.Page.
		WithContext(ctx).
		Query(db).
		Find(&list)
	copierx.Copy(&rp, list)
	return
}
```

sql log:
```mysql
SELECT count(*) FROM `role`;  
SELECT `role`.`id`,`role`.`name`,`role`.`word`,`role`.`action` FROM `role` JOIN (SELECT `role`.`id` AS `OFFSET_KEY` FROM `role` ORDER BY id DESC LIMIT 1) AS `OFFSET_T` ON `role`.`id` = `OFFSET_T`.`OFFSET_KEY` ORDER BY id DESC;
```

example from [auth.Role.Find](https://github.com/go-cinch/auth/blob/dev/internal/data/role.go#L55)
