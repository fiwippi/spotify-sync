package server

import (
	"encoding/json"
	"errors"
	bolt "go.etcd.io/bbolt"
)

// The database
var db *bolt.DB

// Entry in the db
type entry struct {
	Name     string `json:"name"`               // Current name of the entry
	NewName  string `json:"new_name,omitempty"` // New name of the entry if is going to be updated
	Password string `json:"password"`           // Password of the entry
	Token    string `json:"token"`              // oauth2 token of the entry
}

// Saves an entry to the database. If overwrite is false then an
// error is returned when attempting to save over the existing user
func dbSaveUser(e *entry, overwrite bool) error {
	if len(e.Name) == 0 {
		return errors.New("Entry must have a valid name")
	}

	return db.Update(func(tx *bolt.Tx) error {
		// Get the users bucket
		b := tx.Bucket([]byte("users"))

		// Exit if key already exists and overwrite set to false
		if !overwrite && b.Get([]byte(e.Name)) != nil {
			return errors.New("User already exists and overwrite set to false")
		}

		// Create the users struct
		v, err := json.Marshal(e)
		if err != nil {
			return err
		}

		// Write the struct to the bucket
		return b.Put([]byte(e.Name), v)
	})
}

// Deletes a user entry from the database
func dbDeleteUser(e *entry) error {
	return db.Update(func(tx *bolt.Tx) error {
		// Get the users bucket
		b := tx.Bucket([]byte("users"))

		return b.Delete([]byte(e.Name))
	})
}

// Updates a user's data within the database, this allows
// the changing of both username and password
func dbUpdateUser(newEntry *entry) error {
	return db.Update(func(tx *bolt.Tx) error {
		// Get the users bucket
		b := tx.Bucket([]byte("users"))

		// Deserialise the entry
		var oldEntry entry
		err := json.Unmarshal(b.Get([]byte(newEntry.Name)), &oldEntry)
		if err != nil {
			return err
		}

		// Keep copy of the old name for later
		oldName := oldEntry.Name

		// Edit the entry
		if newEntry.Password != oldEntry.Password && len(newEntry.Password) != 0 {
			oldEntry.Password = newEntry.Password
		}
		if newEntry.NewName != oldEntry.Name && len(newEntry.NewName) != 0 {
			oldEntry.NewName = newEntry.NewName
		}

		// Ensure new name is not taken
		if oldEntry.NewName != oldEntry.Name && len(oldEntry.NewName) != 0 {
			if b.Get([]byte(oldEntry.NewName)) == nil {
				oldEntry.Name = oldEntry.NewName

				// Delete the old key
				err = b.Delete([]byte(oldName))
				if err != nil {
					return err
				}
			} else {
				return errors.New("Cannot switch name because new name already exists")
			}
		}

		// Create the users struct
		v, err := json.Marshal(oldEntry)
		if err != nil {
			return err
		}

		// Write the struct to the bucket
		err = b.Put([]byte(oldEntry.Name), v)
		if err != nil {
			return err
		}

		return nil
	})
}

// Returns the deserialised entry for a user from the database
func dbViewUser(name string) (*entry, error) {
	var e entry

	err := db.View(func(tx *bolt.Tx) error {
		// Get the users bucket
		b := tx.Bucket([]byte("users"))

		// Deserialise the entry
		err := json.Unmarshal(b.Get([]byte(name)), &e)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &e, nil
}