package copierx

import (
	"github.com/golang-module/carbon/v2"
	"github.com/jinzhu/copier"
)

var (
	CarbonToString = []copier.TypeConverter{
		{
			SrcType: carbon.DateTime{},
			DstType: copier.String,
			Fn:      carbonToString,
		},
		{
			SrcType: carbon.Date{},
			DstType: copier.String,
			Fn:      carbonToString,
		},
	}
	StringToCarbon = []copier.TypeConverter{
		{
			SrcType: copier.String,
			DstType: carbon.DateTime{},
			Fn:      stringToCarbonTime,
		},
		{
			SrcType: copier.String,
			DstType: carbon.Date{},
			Fn:      stringToCarbonDate,
		},
	}
)

func carbonToString(src interface{}) (rp interface{}, err error) {
	rp = ""
	if v, ok := src.(carbon.DateTime); ok {
		if !v.IsZero() {
			rp = v.ToDateTimeString()
		}
	}
	if v, ok := src.(carbon.Date); ok {
		if !v.IsZero() {
			rp = v.ToDateString()
		}
	}
	return
}

func stringToCarbonTime(src interface{}) (rp interface{}, err error) {
	if v, ok := src.(string); ok {
		rp = carbon.DateTime{
			Carbon: carbon.Parse(v),
		}
	}
	return
}

func stringToCarbonDate(src interface{}) (rp interface{}, err error) {
	if v, ok := src.(string); ok {
		rp = carbon.Date{
			Carbon: carbon.Parse(v),
		}
	}
	return
}

func Copy(to interface{}, from interface{}) (err error) {
	return copier.CopyWithOption(to, from, copier.Option{Converters: append(CarbonToString, StringToCarbon...)})
}

func CopyWithOption(to interface{}, from interface{}, opt copier.Option) (err error) {
	opt.Converters = append(CarbonToString, StringToCarbon...)
	return copier.CopyWithOption(to, from, opt)
}
