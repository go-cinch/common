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
	PtrCarbonToPtrString = []copier.TypeConverter{
		{
			SrcType: &carbon.DateTime{},
			DstType: new(string),
			Fn:      ptrCarbonDateTimeToPtrString,
		},
	}
	PtrStringToPtrCarbon = []copier.TypeConverter{
		{
			SrcType: new(string),
			DstType: &carbon.DateTime{},
			Fn:      ptrStringToPtrCarbonDateTime,
		},
	}
	PtrCarbonToString = []copier.TypeConverter{
		{
			SrcType: &carbon.DateTime{},
			DstType: copier.String,
			Fn:      ptrCarbonDateTimeToString,
		},
	}
	CarbonToPtrString = []copier.TypeConverter{
		{
			SrcType: carbon.DateTime{},
			DstType: new(string),
			Fn:      carbonDateTimeToPtrString,
		},
	}
	PtrStringToCarbon = []copier.TypeConverter{
		{
			SrcType: new(string),
			DstType: carbon.DateTime{},
			Fn:      ptrStringToCarbonDateTime,
		},
	}
	StringToPtrCarbon = []copier.TypeConverter{
		{
			SrcType: copier.String,
			DstType: &carbon.DateTime{},
			Fn:      stringToPtrCarbonDateTime,
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

func ptrCarbonDateTimeToPtrString(src interface{}) (rp interface{}, err error) {
	if v, ok := src.(*carbon.DateTime); ok {
		if v != nil && !v.IsZero() {
			s := v.ToDateTimeString()
			rp = &s
		}
	}
	return
}

func ptrStringToPtrCarbonDateTime(src interface{}) (rp interface{}, err error) {
	if v, ok := src.(*string); ok {
		if v != nil && *v != "" {
			rp = &carbon.DateTime{
				Carbon: carbon.Parse(*v),
			}
		}
	}
	return
}

func ptrCarbonDateTimeToString(src interface{}) (rp interface{}, err error) {
	rp = ""
	if v, ok := src.(*carbon.DateTime); ok {
		if v != nil && !v.IsZero() {
			rp = v.ToDateTimeString()
		}
	}
	return
}

func carbonDateTimeToPtrString(src interface{}) (rp interface{}, err error) {
	if v, ok := src.(carbon.DateTime); ok {
		if !v.IsZero() {
			s := v.ToDateTimeString()
			rp = &s
		}
	}
	return
}

func ptrStringToCarbonDateTime(src interface{}) (rp interface{}, err error) {
	if v, ok := src.(*string); ok {
		if v != nil && *v != "" {
			rp = carbon.DateTime{
				Carbon: carbon.Parse(*v),
			}
		}
	}
	return
}

func stringToPtrCarbonDateTime(src interface{}) (rp interface{}, err error) {
	if v, ok := src.(string); ok {
		if v != "" {
			rp = &carbon.DateTime{
				Carbon: carbon.Parse(v),
			}
		}
	}
	return
}

func allConverters() []copier.TypeConverter {
	converters := make([]copier.TypeConverter, 0)
	converters = append(converters, CarbonToString...)
	converters = append(converters, StringToCarbon...)
	converters = append(converters, PtrCarbonToPtrString...)
	converters = append(converters, PtrStringToPtrCarbon...)
	converters = append(converters, PtrCarbonToString...)
	converters = append(converters, CarbonToPtrString...)
	converters = append(converters, PtrStringToCarbon...)
	converters = append(converters, StringToPtrCarbon...)
	return converters
}

func Copy(to interface{}, from interface{}) (err error) {
	return copier.CopyWithOption(to, from, copier.Option{Converters: allConverters()})
}

func CopyWithOption(to interface{}, from interface{}, opt copier.Option) (err error) {
	opt.Converters = append(opt.Converters, allConverters()...)
	return copier.CopyWithOption(to, from, opt)
}
