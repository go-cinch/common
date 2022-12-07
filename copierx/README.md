# Copierx


object [copier](https://github.com/jinzhu/copier) with [carbon](https://github.com/golang-module/carbon).

- carbon.DateTime to string
- string to carbon.DateTime
- carbon.Date to string
- string to carbon.Date


## Usage


```bash
go get -u github.com/go-cinch/common/copierx
```


```go
import (
	"fmt"
	"github.com/go-cinch/common/copierx"
	"github.com/golang-module/carbon/v2"
	"github.com/jinzhu/copier"
)

type A struct {
	CreateTime carbon.DateTime
	CreateDate carbon.Date
	UpdateTime string
	UpdateDate string
}

type B struct {
	CreateTime string
	CreateDate string
	UpdateTime carbon.DateTime
	UpdateDate carbon.Date
}

func main() {
	now := carbon.Now()
	a := A{
		CreateTime: carbon.DateTime{
			Carbon: now,
		},
		CreateDate: carbon.Date{
			Carbon: now,
		},
		UpdateTime: "2022-12-07 00:00:00",
		UpdateDate: "2022-12-07",
	}
	var b B
	fmt.Println(b)

	// copier not support
	copier.Copy(&b, a)
	fmt.Println(b)

	copierx.Copy(&b, a)
	fmt.Println(b)
}
```
