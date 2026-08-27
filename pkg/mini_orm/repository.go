package mini_orm

import (
	"gorm.io/gorm"
)

// Repository Struct
type Repository[D DAO] struct {
	DB *gorm.DB
}

// Checks if the ID of the crudModel already exists in the database. If it does, update
// the fields of the entry based on the fields of the crudModel, if not, create a new
// entry.
func (r Repository[D]) Save(record D) {
	// Creates record if record not in database. Updates record
	// if present in the database
	result := r.DB.Save(&record)
	if err := result.Error; err != nil {
		panic(err)
	}
}

// Returns all records (of type DAO) from database
func (r Repository[D]) Load() []D {
	var records []D

	// Fetch all records
	result := r.DB.Find(&records)
	if err := result.Error; err != nil {
		panic(err)
	}

	return records
}

// Deletes record by primary key
func (r Repository[D]) Delete(ID uint) {
	// Fetch record
	var record D
	r.DB.Find(&record, ID)
	// Delete record
	result := r.DB.Delete(&record)
	if err := result.Error; err != nil {
		panic(err)
	}
}
