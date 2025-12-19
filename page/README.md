# Page

Simple pagination with [gorm](https://gorm.io/gorm) using Scopes pattern.

## Usage

```bash
go get -u github.com/go-cinch/common/page/v2
```

### Fields

| Field   | Type   | Description                        |
|---------|--------|------------------------------------|
| Num     | uint64 | current page number                |
| Size    | uint64 | page size (records per page)       |
| Total   | int64  | total records count                |
| Disable | bool   | disable pagination, query all data |

### Example

```go
p := &page.Page{Num: 1, Size: 10}
db.Model(&User{}).Count(&p.Total)
db.Scopes(p.Scope()).Find(&users)
```
