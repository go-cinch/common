package copierx

import (
	"testing"

	"github.com/golang-module/carbon/v2"
	"github.com/jinzhu/copier"
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
