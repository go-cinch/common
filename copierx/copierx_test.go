package copierx

import (
	"testing"

	"github.com/golang-module/carbon/v2"
	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
)

func TestCarbonToString(t *testing.T) {
	type Src struct {
		CreateTime carbon.DateTime
	}
	type Dst struct {
		CreateTime string
	}

	now := carbon.Now()
	src := Src{CreateTime: carbon.DateTime{Carbon: now}}
	var dst Dst

	err := Copy(&dst, src)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
	if dst.CreateTime != now.ToDateTimeString() {
		t.Errorf("expected %s, got %s", now.ToDateTimeString(), dst.CreateTime)
	}
}

func TestStringToCarbon(t *testing.T) {
	type Src struct {
		CreateTime string
	}
	type Dst struct {
		CreateTime carbon.DateTime
	}

	src := Src{CreateTime: "2024-01-15 10:30:00"}
	var dst Dst

	err := Copy(&dst, src)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
	if dst.CreateTime.ToDateTimeString() != "2024-01-15 10:30:00" {
		t.Errorf("expected 2024-01-15 10:30:00, got %s", dst.CreateTime.ToDateTimeString())
	}
}

func TestPtrCarbonToPtrString(t *testing.T) {
	type Src struct {
		CreateTime *carbon.DateTime
	}
	type Dst struct {
		CreateTime *string
	}

	now := carbon.Now()
	dt := carbon.DateTime{Carbon: now}
	src := Src{CreateTime: &dt}
	var dst Dst

	err := Copy(&dst, src)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
	if dst.CreateTime == nil {
		t.Fatal("expected non-nil CreateTime")
	}
	if *dst.CreateTime != now.ToDateTimeString() {
		t.Errorf("expected %s, got %s", now.ToDateTimeString(), *dst.CreateTime)
	}
}

func TestPtrStringToPtrCarbon(t *testing.T) {
	type Src struct {
		CreateTime *string
	}
	type Dst struct {
		CreateTime *carbon.DateTime
	}

	s := "2024-01-15 10:30:00"
	src := Src{CreateTime: &s}
	var dst Dst

	err := Copy(&dst, src)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
	if dst.CreateTime == nil {
		t.Fatal("expected non-nil CreateTime")
	}
	if dst.CreateTime.ToDateTimeString() != "2024-01-15 10:30:00" {
		t.Errorf("expected 2024-01-15 10:30:00, got %s", dst.CreateTime.ToDateTimeString())
	}
}

func TestPtrCarbonToString(t *testing.T) {
	type Src struct {
		CreateTime *carbon.DateTime
	}
	type Dst struct {
		CreateTime string
	}

	now := carbon.Now()
	dt := carbon.DateTime{Carbon: now}
	src := Src{CreateTime: &dt}
	var dst Dst

	err := Copy(&dst, src)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
	if dst.CreateTime != now.ToDateTimeString() {
		t.Errorf("expected %s, got %s", now.ToDateTimeString(), dst.CreateTime)
	}
}

func TestCarbonToPtrString(t *testing.T) {
	type Src struct {
		CreateTime carbon.DateTime
	}
	type Dst struct {
		CreateTime *string
	}

	now := carbon.Now()
	src := Src{CreateTime: carbon.DateTime{Carbon: now}}
	var dst Dst

	err := Copy(&dst, src)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
	if dst.CreateTime == nil {
		t.Fatal("expected non-nil CreateTime")
	}
	if *dst.CreateTime != now.ToDateTimeString() {
		t.Errorf("expected %s, got %s", now.ToDateTimeString(), *dst.CreateTime)
	}
}

func TestPtrStringToCarbon(t *testing.T) {
	type Src struct {
		CreateTime *string
	}
	type Dst struct {
		CreateTime carbon.DateTime
	}

	s := "2024-01-15 10:30:00"
	src := Src{CreateTime: &s}
	var dst Dst

	err := Copy(&dst, src)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
	if dst.CreateTime.ToDateTimeString() != "2024-01-15 10:30:00" {
		t.Errorf("expected 2024-01-15 10:30:00, got %s", dst.CreateTime.ToDateTimeString())
	}
}

func TestStringToPtrCarbon(t *testing.T) {
	type Src struct {
		CreateTime string
	}
	type Dst struct {
		CreateTime *carbon.DateTime
	}

	src := Src{CreateTime: "2024-01-15 10:30:00"}
	var dst Dst

	err := Copy(&dst, src)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
	if dst.CreateTime == nil {
		t.Fatal("expected non-nil CreateTime")
	}
	if dst.CreateTime.ToDateTimeString() != "2024-01-15 10:30:00" {
		t.Errorf("expected 2024-01-15 10:30:00, got %s", dst.CreateTime.ToDateTimeString())
	}
}

func TestNilPtrCarbonToPtrString(t *testing.T) {
	type Src struct {
		CreateTime *carbon.DateTime
	}
	type Dst struct {
		CreateTime *string
	}

	src := Src{CreateTime: nil}
	var dst Dst

	err := Copy(&dst, src)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
	if dst.CreateTime != nil {
		t.Errorf("expected nil, got %v", dst.CreateTime)
	}
}

func TestNilPtrStringToPtrCarbon(t *testing.T) {
	type Src struct {
		CreateTime *string
	}
	type Dst struct {
		CreateTime *carbon.DateTime
	}

	src := Src{CreateTime: nil}
	var dst Dst

	err := Copy(&dst, src)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
	if dst.CreateTime != nil {
		t.Errorf("expected nil, got %v", dst.CreateTime)
	}
}

func TestCopyWithOptionPreservesExternalConverters(t *testing.T) {
	type Src struct {
		CreateTime carbon.DateTime
		Value      int
	}
	type Dst struct {
		CreateTime string
		Value      string
	}

	customConverter := copier.TypeConverter{
		SrcType: int(0),
		DstType: copier.String,
		Fn: func(src interface{}) (interface{}, error) {
			if v, ok := src.(int); ok {
				return "custom:" + string(rune('0'+v)), nil
			}
			return "", nil
		},
	}

	now := carbon.Now()
	src := Src{CreateTime: carbon.DateTime{Carbon: now}, Value: 5}
	var dst Dst

	err := CopyWithOption(&dst, src, copier.Option{
		Converters: []copier.TypeConverter{customConverter},
	})
	if err != nil {
		t.Fatalf("CopyWithOption failed: %v", err)
	}
	if dst.CreateTime != now.ToDateTimeString() {
		t.Errorf("expected %s, got %s", now.ToDateTimeString(), dst.CreateTime)
	}
}

// Decimal tests
func TestDecimalToString(t *testing.T) {
	type Src struct {
		Amount decimal.Decimal
	}
	type Dst struct {
		Amount string
	}

	src := Src{Amount: decimal.NewFromFloat(123.45)}
	var dst Dst

	err := Copy(&dst, src)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
	if dst.Amount != "123.45" {
		t.Errorf("expected 123.45, got %s", dst.Amount)
	}
}

func TestStringToDecimal(t *testing.T) {
	type Src struct {
		Amount string
	}
	type Dst struct {
		Amount decimal.Decimal
	}

	src := Src{Amount: "123.45"}
	var dst Dst

	err := Copy(&dst, src)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
	if !dst.Amount.Equal(decimal.NewFromFloat(123.45)) {
		t.Errorf("expected 123.45, got %s", dst.Amount.String())
	}
}

func TestPtrDecimalToPtrString(t *testing.T) {
	type Src struct {
		Amount *decimal.Decimal
	}
	type Dst struct {
		Amount *string
	}

	d := decimal.NewFromFloat(123.45)
	src := Src{Amount: &d}
	var dst Dst

	err := Copy(&dst, src)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
	if dst.Amount == nil {
		t.Fatal("expected non-nil Amount")
	}
	if *dst.Amount != "123.45" {
		t.Errorf("expected 123.45, got %s", *dst.Amount)
	}
}

func TestPtrStringToPtrDecimal(t *testing.T) {
	type Src struct {
		Amount *string
	}
	type Dst struct {
		Amount *decimal.Decimal
	}

	s := "123.45"
	src := Src{Amount: &s}
	var dst Dst

	err := Copy(&dst, src)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
	if dst.Amount == nil {
		t.Fatal("expected non-nil Amount")
	}
	if !dst.Amount.Equal(decimal.NewFromFloat(123.45)) {
		t.Errorf("expected 123.45, got %s", dst.Amount.String())
	}
}

func TestPtrDecimalToString(t *testing.T) {
	type Src struct {
		Amount *decimal.Decimal
	}
	type Dst struct {
		Amount string
	}

	d := decimal.NewFromFloat(123.45)
	src := Src{Amount: &d}
	var dst Dst

	err := Copy(&dst, src)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
	if dst.Amount != "123.45" {
		t.Errorf("expected 123.45, got %s", dst.Amount)
	}
}

func TestDecimalToPtrString(t *testing.T) {
	type Src struct {
		Amount decimal.Decimal
	}
	type Dst struct {
		Amount *string
	}

	src := Src{Amount: decimal.NewFromFloat(123.45)}
	var dst Dst

	err := Copy(&dst, src)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
	if dst.Amount == nil {
		t.Fatal("expected non-nil Amount")
	}
	if *dst.Amount != "123.45" {
		t.Errorf("expected 123.45, got %s", *dst.Amount)
	}
}

func TestPtrStringToDecimal(t *testing.T) {
	type Src struct {
		Amount *string
	}
	type Dst struct {
		Amount decimal.Decimal
	}

	s := "123.45"
	src := Src{Amount: &s}
	var dst Dst

	err := Copy(&dst, src)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
	if !dst.Amount.Equal(decimal.NewFromFloat(123.45)) {
		t.Errorf("expected 123.45, got %s", dst.Amount.String())
	}
}

func TestStringToPtrDecimal(t *testing.T) {
	type Src struct {
		Amount string
	}
	type Dst struct {
		Amount *decimal.Decimal
	}

	src := Src{Amount: "123.45"}
	var dst Dst

	err := Copy(&dst, src)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
	if dst.Amount == nil {
		t.Fatal("expected non-nil Amount")
	}
	if !dst.Amount.Equal(decimal.NewFromFloat(123.45)) {
		t.Errorf("expected 123.45, got %s", dst.Amount.String())
	}
}

func TestNilPtrDecimalToPtrString(t *testing.T) {
	type Src struct {
		Amount *decimal.Decimal
	}
	type Dst struct {
		Amount *string
	}

	src := Src{Amount: nil}
	var dst Dst

	err := Copy(&dst, src)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
	if dst.Amount != nil {
		t.Errorf("expected nil, got %v", dst.Amount)
	}
}

func TestNilPtrStringToPtrDecimal(t *testing.T) {
	type Src struct {
		Amount *string
	}
	type Dst struct {
		Amount *decimal.Decimal
	}

	src := Src{Amount: nil}
	var dst Dst

	err := Copy(&dst, src)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
	if dst.Amount != nil {
		t.Errorf("expected nil, got %v", dst.Amount)
	}
}
