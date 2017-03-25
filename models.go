package main

type Vote struct {
	ID       int
	Name     string `gorm:"column:name;type:varchar(255)"`
	SName    string `gorm:"column:sname;type:varchar(255)"`
	Region   Region
	RegionID int `json:"region_id",gorm:"calumn:region_id;type:int(11)"`
	City     string `gorm:"calumn:city;type:varchar(255)"`
}

type Region struct {
	ID      int    `json:"id"`
	NameUkr string `json:"name_ukr",gorm:"column:name_ukr;type:varchar(255)"`
}

//func SetField(obj interface{}, name string, value interface{}) error {
//	structValue := reflect.ValueOf(obj).Elem()
//	structFieldValue := structValue.FieldByName(name)
//
//	if !structFieldValue.IsValid() {
//		return fmt.Errorf("No such field: %s in obj", name)
//	}
//
//	if !structFieldValue.CanSet() {
//		return fmt.Errorf("Cannot set %s field value", name)
//	}
//
//	structFieldType := structFieldValue.Type()
//	val := reflect.ValueOf(value)
//	if structFieldType != val.Type() {
//		return errors.New("Provided value type didn't match obj field type")
//	}
//
//	structFieldValue.Set(val)
//	return nil
//}
