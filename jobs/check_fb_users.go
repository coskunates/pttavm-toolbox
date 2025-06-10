package jobs

import (
	"encoding/csv"
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"my_toolbox/entities"
	"my_toolbox/helpers"
	"my_toolbox/models"
	"os"
)

type CheckFBUsers struct {
	Job
}

func (c *CheckFBUsers) Run() {
	users := c.getRecords()

	for _, user := range users {

		var fbUsers []entities.EUser
		c.DB.Model(entities.EUser{}).
			Where("fb_userid", user.FBUserId).
			Find(&fbUsers)

		userNames := make(map[string]string)
		for _, fbUser := range fbUsers {
			userNames[fmt.Sprintf(
				"%s:%s",
				cases.Lower(language.Turkish).String(fbUser.UserNome),
				cases.Lower(language.Turkish).String(fbUser.UserCognome),
			)] = fmt.Sprintf("%v", fbUser.FbUserid)
		}

		if len(userNames) > 1 {

			var data [][]string
			var names []string
			fbUserId := ""
			for name, id := range userNames {
				fbUserId = id
				names = append(names, name)
			}

			names = append(names, fbUserId)

			data = append(data, names)
			fmt.Println(data)
			helpers.Write("assets/user/different_names.csv", data)
		}
	}

}

func (c *CheckFBUsers) getRecords() []models.SameUserRecord {
	f, err := os.Open("assets/user/same_fb_user_id.csv")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		fmt.Println(err.Error())
	}

	var users []models.SameUserRecord
	for _, record := range records {
		users = append(users, models.SameUserRecord{
			FBUserId: record[0],
		})
	}

	return users
}
