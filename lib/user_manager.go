package lib

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

func StoreFile(app *pocketbase.PocketBase, userId string, filepath string) (string, error) {
	collection, err := app.Dao().FindCollectionByNameOrId("files")
	if err != nil {
		return "", err
	}

	record := models.NewRecord(collection)
	record.Set("owner", userId)

	form := forms.NewRecordUpsert(app, record)

	f1, err := filesystem.NewFileFromPath(filepath)
	if err != nil {
		return "", err
	}

	form.AddFiles("file", f1)

	if err := form.Submit(); err != nil {
		return "", err
	}

	return record.Id, nil
}

// func updateUserRecord(app *pocketbase.PocketBase, user *models.Record, fileId string) error {
// 	filesId := user.GetStringSlice("user_files")
// 	form := forms.NewRecordUpsert(app, user)

// 	form.LoadData(map[string]any{
// 		"user_files": append(filesId, fileId),
// 	})

// 	if err := form.Submit(); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func storeUserFile(app *pocketbase.PocketBase, filepath string, user *models.Record) error {
// 	fileId, err := StoreFile(app, user.Id, filepath)
// 	if err != nil {
// 		return err
// 	}

// 	err = updateUserRecord(app, user, fileId)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
