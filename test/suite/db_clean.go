package suite

import (
	"database/sql"
	"fmt"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

// DeleteCreatedEntities records all created entities on the gorm.DB connection
// and returns a function which can be called on defer to delete created
// entities in reverse order on function exit.
//
func DeleteCreatedEntities(db *gorm.DB) func() {
	hookName := fmt.Sprintf("stopwatch:record:%s", uuid.NewV4().String())
	type entity struct {
		table   string
		keyname string
		key     interface{}
	}
	var entires []entity
	db.Callback().Create().Get(hookName)
	db.Callback().Create().After("gorm:create").Register(hookName, func(scope *gorm.Scope) {
		log.Debugf("Inserted entities from %s with %s=%v", scope.TableName(), scope.PrimaryKey(), scope.PrimaryKeyValue())
		entires = append(entires, entity{table: scope.TableName(), keyname: scope.PrimaryKey(), key: scope.PrimaryKeyValue()})
	})
	return func() {
		defer db.Callback().Create().Remove(hookName)
		// Find out if the current db object is already a transaction
		_, inTransaction := db.CommonDB().(*sql.Tx)
		tx := db
		if !inTransaction {
			tx = db.Begin()
		}
		for i := len(entires) - 1; i >= 0; i-- {
			entry := entires[i]
			log.Debugf("Deleting entities from '%s' table with key %v", entry.table, entry.key)
			tx.Table(entry.table).Where(entry.keyname+" = ?", entry.key).Delete("")
		}

		// Delete the work item cache as well
		// NOTE: Feel free to add more cache freeing calls here as needed.
		// workitem.ClearGlobalWorkItemTypeCache()
		// TODO: need a way to hook custom clean functions in here

		if !inTransaction {
			tx.Commit()
		}
	}
}
