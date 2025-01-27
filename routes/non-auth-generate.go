package routes

import (
	"fmt"
	"immodi/startup/lib"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
)

func NonAuthedGenerate(c echo.Context, app *pocketbase.PocketBase, request *MessageRequest, javascript string) error {
	nonAuthUser, err := app.Dao().FindFirstRecordByData("no_auth_users", "user_ip", c.RealIP())
	if err != nil {
		collection, err := app.Dao().FindCollectionByNameOrId("no_auth_users")
		if err != nil {
			return err
		}

		nonAuthUser = models.NewRecord(collection)

		nonAuthUser.Set("user_ip", c.RealIP())
		nonAuthUser.Set("generations", 10)
		err = app.Dao().Save(nonAuthUser)
		if err != nil {
			return err
		}
	}

	if nonAuthUser != nil && nonAuthUser.GetInt("generations") <= 0 {
		return fmt.Errorf("please login to create more files")
	}

	if request.Data != nil {
		javascript = jsInjectionScript(&request.Data)
	}

	if request.Topic == "" {
		return fmt.Errorf("missing required fields, 'topic' or 'template'")
	}

	message, styleTag := MessageBuilder(c, app, request.Topic, request.Template, request.Level)

	htmlData := lib.HtmlFileData{
		Username:       c.RealIP(),
		UserBlobName:   lib.GenerateUniqueString(c.RealIP()),
		TemplateName:   request.Template,
		HtmlData:       "",
		StyleTag:       styleTag,
		InsertStyleTag: InsertStyleTag,
	}

	htmlFilePath, err := getAIResponseAndWriteHTMl(c, app, &htmlData, message, 3)
	if err != nil {
		go os.RemoveAll(htmlData.UserBlobName)
		return fmt.Errorf("error writing the html")
	}

	filepath, err := lib.ParsePdfFile(c, app, lib.HtmlParserConfig{
		TemplateName:    htmlFilePath,
		JavascriptToRun: javascript,
		OutputFileName:  htmlData.UserBlobName,
		Username:        c.RealIP(),
	})

	if err != nil {
		go os.RemoveAll(htmlData.UserBlobName)
		return fmt.Errorf("error parsing the pdf file")
	}

	go os.RemoveAll(fmt.Sprintf("user_data/%s", c.RealIP()))

	err = c.Attachment(filepath, "file.pdf")
	if err != nil {
		return fmt.Errorf("error sending the response")
	}

	return nil
}

func DecrementNonAuthedUserLimit(app *pocketbase.PocketBase, userIp string) error {
	record, err := app.Dao().FindFirstRecordByData("no_auth_users", "user_ip", userIp)
	if err != nil {
		CreateNonAuthedUser(app, userIp, record)
	}

	record.Set("generations", record.GetInt("generations")-1)

	err = app.Dao().Save(record)
	if err != nil {
		return err
	}

	return nil
}

func CreateNonAuthedUser(app *pocketbase.PocketBase, userIp string, record *models.Record) error {
	if record == nil {
		collection, err := app.Dao().FindCollectionByNameOrId("no_auth_users")
		if err != nil {
			return err
		}

		record = models.NewRecord(collection)

		record.Set("user_ip", userIp)
		record.Set("generations", 10)
	}

	return nil
}
