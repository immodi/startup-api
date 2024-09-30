package lib

import "fmt"

func ReportTemplateInjectedJs(name string, affiliation string) string {
	if affiliation != "" && name != "" {
		return format(name, affiliation)
	} else {
		return format(
			"PLEASE INSERT YOUR NAME HERE",
			"PLEASE INSERT YOUR AFFILIATION",
		)
	}
}

func format(name string, affiliation string) string {
	return fmt.Sprintf(`() => {
			document.querySelector(".name").innerHTML = "%s"
			document.querySelector(".affiliation").innerHTML = "%s"
		}
	`, name, affiliation)
}
