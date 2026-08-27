package repository

// Data Access Object interface. Empty interface. Declared for code readability.
// Any struct implements DAO by default, therefore allowing for any struct that
// extends gorm.Model to be saved into the database through a Repository
type DAO interface{}
