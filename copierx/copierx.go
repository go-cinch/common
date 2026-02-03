package copierx

import (
	"github.com/golang-module/carbon/v2"
	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
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
	// Decimal converters
	DecimalToString = []copier.TypeConverter{
		{
			SrcType: decimal.Decimal{},
			DstType: copier.String,
			Fn:      decimalToString,
		},
	}
	StringToDecimal = []copier.TypeConverter{
		{
			SrcType: copier.String,
			DstType: decimal.Decimal{},
			Fn:      stringToDecimal,
		},
	}
	PtrDecimalToPtrString = []copier.TypeConverter{
		{
			SrcType: &decimal.Decimal{},
			DstType: new(string),
			Fn:      ptrDecimalToPtrString,
		},
	}
	PtrStringToPtrDecimal = []copier.TypeConverter{
		{
			SrcType: new(string),
			DstType: &decimal.Decimal{},
			Fn:      ptrStringToPtrDecimal,
		},
	}
	PtrDecimalToString = []copier.TypeConverter{
		{
			SrcType: &decimal.Decimal{},
			DstType: copier.String,
			Fn:      ptrDecimalToStringFn,
		},
	}
	DecimalToPtrString = []copier.TypeConverter{
		{
			SrcType: decimal.Decimal{},
			DstType: new(string),
			Fn:      decimalToPtrString,
		},
	}
	PtrStringToDecimal = []copier.TypeConverter{
		{
			SrcType: new(string),
			DstType: decimal.Decimal{},
			Fn:      ptrStringToDecimalFn,
		},
	}
	StringToPtrDecimal = []copier.TypeConverter{
		{
			SrcType: copier.String,
			DstType: &decimal.Decimal{},
			Fn:      stringToPtrDecimal,
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

// Decimal conversion functions
func decimalToString(src interface{}) (rp interface{}, err error) {
	rp = ""
	if v, ok := src.(decimal.Decimal); ok {
		rp = v.String()
	}
	return
}

func stringToDecimal(src interface{}) (rp interface{}, err error) {
	rp = decimal.Decimal{}
	if v, ok := src.(string); ok {
		if v != "" {
			rp, err = decimal.NewFromString(v)
		}
	}
	return
}

func ptrDecimalToPtrString(src interface{}) (rp interface{}, err error) {
	if v, ok := src.(*decimal.Decimal); ok {
		if v != nil {
			s := v.String()
			rp = &s
		}
	}
	return
}

func ptrStringToPtrDecimal(src interface{}) (rp interface{}, err error) {
	if v, ok := src.(*string); ok {
		if v != nil && *v != "" {
			d, e := decimal.NewFromString(*v)
			if e == nil {
				rp = &d
			}
			err = e
		}
	}
	return
}

func ptrDecimalToStringFn(src interface{}) (rp interface{}, err error) {
	rp = ""
	if v, ok := src.(*decimal.Decimal); ok {
		if v != nil {
			rp = v.String()
		}
	}
	return
}

func decimalToPtrString(src interface{}) (rp interface{}, err error) {
	if v, ok := src.(decimal.Decimal); ok {
		s := v.String()
		rp = &s
	}
	return
}

func ptrStringToDecimalFn(src interface{}) (rp interface{}, err error) {
	rp = decimal.Decimal{}
	if v, ok := src.(*string); ok {
		if v != nil && *v != "" {
			rp, err = decimal.NewFromString(*v)
		}
	}
	return
}

func stringToPtrDecimal(src interface{}) (rp interface{}, err error) {
	if v, ok := src.(string); ok {
		if v != "" {
			d, e := decimal.NewFromString(v)
			if e == nil {
				rp = &d
			}
			err = e
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
	converters = append(converters, DecimalToString...)
	converters = append(converters, StringToDecimal...)
	converters = append(converters, PtrDecimalToPtrString...)
	converters = append(converters, PtrStringToPtrDecimal...)
	converters = append(converters, PtrDecimalToString...)
	converters = append(converters, DecimalToPtrString...)
	converters = append(converters, PtrStringToDecimal...)
	converters = append(converters, StringToPtrDecimal...)
	return converters
}

func Copy(to interface{}, from interface{}) (err error) {
	return copier.CopyWithOption(to, from, copier.Option{Converters: allConverters()})
}

func CopyWithOption(to interface{}, from interface{}, opt copier.Option) (err error) {
	opt.Converters = append(opt.Converters, allConverters()...)
	return copier.CopyWithOption(to, from, opt)
}
