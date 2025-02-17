package lib

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

var PlanFileLimits map[string]int = map[string]int{
	"free":    10,
	"monthly": 50,
}

func StoreFile(app *pocketbase.PocketBase, userId string, filepath string) (string, error) {
	collection, err := app.Dao().FindCollectionByNameOrId("files")
	if err != nil {
		return "", err
	}

	user, err := app.Dao().FindRecordById("users", userId)
	if err != nil {
		return "", err
	}

	userPlanId := user.GetString("current_plan")
	plan, err := app.Dao().FindRecordById("plans", userPlanId)
	if err != nil {
		return "", err
	}
	userPlanName := plan.GetString("name")

	err = popFirstFileBasedOnPlan(app, user.Id, userPlanName)
	if err != nil {
		return "", err
	}

	fileRecordId, err := addFileToFilesRecord(app, collection, userId, filepath)
	if err != nil {
		return "", err
	}

	user.RefreshUpdated()
	addFilesToUserRecord(app, user, append(user.GetStringSlice("user_files"), fileRecordId))

	if userPlanName == "free" {
		decrementUserToken(app, user)
	}

	return fileRecordId, nil

}

func addFilesToUserRecord(app *pocketbase.PocketBase, user *models.Record, userFiles []string) error {
	userFiles = getLastSixElements(userFiles)
	user.Set("user_files", userFiles)

	err := app.Dao().Save(user)
	if err != nil {
		return err
	}
	return nil
}

func addFileToFilesRecord(app *pocketbase.PocketBase, collection *models.Collection, userId string, filepath string) (string, error) {
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

func popFirstFileBasedOnPlan(app *pocketbase.PocketBase, userId string, userPlanName string) error {
	files, err := app.Dao().FindRecordsByExpr("files", dbx.HashExp{"owner": userId})
	if err != nil {
		return err
	}

	if len(files) >= PlanFileLimits[userPlanName] && len(files) > 0 {
		oldestFile := files[:1][0]

		app.Dao().DeleteRecord(oldestFile)
	}

	return nil
}

func decrementUserToken(app *pocketbase.PocketBase, user *models.Record) error {
	form := forms.NewRecordUpsert(app, user)
	tokens := user.GetInt("tokens")
	// Create the map with the updated value
	data := map[string]any{
		"tokens": tokens - 1,
	}

	// Simulate form.LoadData
	form.LoadData(data)

	if err := form.Submit(); err != nil {
		return err
	}

	return nil
}

func getLastSixElements(slice []string) []string {
	indexer := 6

	if len(slice) <= indexer {
		return slice
	}

	return slice[len(slice)-indexer:]
}
