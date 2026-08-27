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
	// Fetch record from databse. If record doesn't exist, create new record
	result := r.DB.FirstOrCreate(&record, record)
	if err := result.Error; err != nil {
		panic(err)
	}
	// If record exists, update its fields
	if recordExists := result.RowsAffected == 0; recordExists {
		r.DB.Save(&record)
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

	// Cast DAO type
	// daoRecords := make([]DAO, len(records))
	// for i, record := range records {
	// 	daoRecords[i] = record
	// }

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
