package lib

import (
	"html/template"
	"os"
)

type Tag struct {
	Index     int
	Type      bool //`opening:1 - closing:0`
	TagLength int
}

func WriteResponseHTML(htmlData string) error {
	htmlData = SanatizeHtml(htmlData)

	templateHtml, err := ReadHtmlFileData("template.html")
	if err != nil {
		return err
	}

	// Create or open a file to write the output
	file, err := os.Create("data.html")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Create a new template and parse the template string
	tmpl, err := template.New("page").Parse(templateHtml)
	if err != nil {
		panic(err)
	}

	// Write the template output to the file
	err = tmpl.Execute(file, struct {
		Content template.HTML
	}{
		Content: template.HTML(htmlData),
	})
	if err != nil {
		panic(err)
	}

	return nil
}

func SanatizeHtml(badHtml string) string {
	// badHtml := badHtml() + " "
	badIndex := make([]Tag, 0)

	for index, char := range badHtml {
		if char == '<' {
			if badHtml[index+1] != '/' {
				_, supposedClosingIndex, tagLength := getTag(&badHtml, index, "")
				if badHtml[supposedClosingIndex] != '>' {
					badIndex = append(badIndex, Tag{
						Index:     index + 2,
						Type:      true,
						TagLength: tagLength,
					})
				}
			} else {
				char := rune(badHtml[index+2])
				tagLength := getTagLength(char)
				// println(string(badHtml[index+2+tagLength]))
				if badHtml[index+2+tagLength] != '>' {
					badIndex = append(badIndex, Tag{
						Index:     index + tagLength + 1,
						Type:      false,
						TagLength: tagLength,
					})
				}
			}
		}
	}

	for shiftIndex, tag := range badIndex {
		badHtml = insertAtIndex(badHtml, tag.Index+1+shiftIndex, '>')
	}

	return badHtml
}

func badHtml() string {
	return `
	<div>
		<h1>Dark Souls: A Report</h1>
		<p>Developed by FromSoftware, Dark Souls is an action role-playing game that has garnered a cult following since its release in 2011. The game is known for its challenging gameplay, Gothic atmosphere, and interconnected world design.</p>
		<h2>Gameplay</h2>
		<ul>
			<li>Explore a vast, interconnected world featuring diverse environments, from dark forests to cursed cities.</li>
			<li>Fight against a variety of fearsome enemies, from giant spiders to undead warriors.</li>
			<li>Master a deep combat system that rewards strategy and skill.</li>
			<li>Level up and upgrade your character's abilities and equipment to overcome the challenges that lie ahead.</li>
		</ul>
		<h2 Story</h2>
		<p>In Dark Souls, you play as a cursed undead character who is tasked with reversing the spread of darkness in the world. The game's story is told through subtle hints and clues, rather than explicit narrative.</p>
		<ul>
			<li>Uncover the mysteries of the world by exploring hidden areas and speaking with NPCs.</li>
			<li>Make choices that impact the world and its inhabitants.</li>
			<li>Experience a story that is as much about the characters' own struggles as it is about the fate of the world.</li>
		</ul>
		<h2 Reception</h2>
		<p>Dark Souls has received widespread critical acclaim for its challenging gameplay, atmospheric world, and deep storytelling.</p>
		<ul>
			<li>The game has a Metacritic score of 89% on PC, 88% on PlayStation 3, and 87% on Xbox 360.</li>
			<li>Dark Souls has won numerous awards, including Game of the Year at the 2011 Golden Joystick Awards.</li>
			<li>A community of dedicated fans has sprung up around the game, with many creating their own mods, art, and fiction inspired by the game.</li>
		</ul>
	</div`
}

func getTag(refString *string, index int, tag string) (string, int, int) {
	t := tag
	stringArray := []byte(*refString)

	if stringArray[index+1] != '>' {
		t += string(stringArray[index+1])
		return getTag(refString, index+1, t)
	}

	tagCharLength := getTagLength(rune(tag[0]))
	if tagCharLength == 0 {
		tagCharLength = 2
	}

	return tag, index - (len(tag) - 1) + tagCharLength, tagCharLength
}

func getTagLength(tagRune rune) int {
	// Define a map to associate constant values with their names
	constNames := map[rune]int{
		'h': 2,
		'u': 2,
		'p': 1,
		'l': 2,
		'd': 3,
	}

	// Lookup the value in the map
	if len, exists := constNames[tagRune]; exists {
		return len
	}
	return 0
}

func insertAtIndex(original string, index int, newContent rune) string {
	// Ensure the index is within bounds
	if index < 0 || index > len(original) {
		return original // Index out of bounds
	}

	// Split the original string into two parts
	before := original[:index]
	after := original[index:]

	// Concatenate the parts with the new content in between
	return before + string(newContent) + after
}

func replaceChar(s string, index int, newChar rune) string {
	if index < 0 || index >= len(s) {
		return s // index out of bounds
	}

	// Convert the string to a slice of runes
	runes := []rune(s)
	// Replace the character at the given index
	runes[index] = newChar
	// Convert the rune slice back to a string
	return string(runes)
}
