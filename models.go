package main

type Vote struct {
	ID       int
	Name     string `gorm:"column:name;type:varchar(255)"`
	SName    string `gorm:"column:sname;type:varchar(255)"`
	region   Region
	RegionID int `json:"region_id",gorm:"calumn:region_id;type:int(11)"`
	City     string `gorm:"calumn:city;type:varchar(255)"`
}

type Region struct {
	ID      int    `json:"id"`
	NameUkr string `json:"name_ukr",gorm:"column:name_ukr;type:varchar(255)"`
}
