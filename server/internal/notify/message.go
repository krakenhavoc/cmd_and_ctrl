package notify

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

// message.go is ADR 0162 §3: plain text, one line and a link, nothing
// from the game's hidden information. The table name has already passed
// the ADR 0075 §2.6 filter; it is flattened here so it cannot restyle or
// restructure the message, and SendDM's empty allowed_mentions means it
// cannot ping anyone either.

// maxNameLen caps the table name in a message, as the invite DM does.
const maxNameLen = 80

// turnMessage is "your turn". Before turn 1 (the opening roll and the
// mulligans) there is no turn number to give.
func turnMessage(table string, turn int, link string) string {
	if turn <= 0 {
		return fmt.Sprintf("The game at **%s** is waiting on you. %s", flatten(table), link)
	}
	return fmt.Sprintf("It's your turn at **%s** (turn %d). %s", flatten(table), turn, link)
}

// filledMessage is "table ready" or "table filled" for a fill.
func filledMessage(table, link string) string {
	return fmt.Sprintf("**%s** is full and ready to start. %s", flatten(table), link)
}

// startedMessage is "table ready" for a start.
func startedMessage(table, link string) string {
	return fmt.Sprintf("**%s** has started. %s", flatten(table), link)
}

// gameURL is the client's route for a table.
func gameURL(base string, game uuid.UUID) string {
	return fmt.Sprintf("%s/#/games/%s", strings.TrimRight(base, "/"), game)
}

// flatten strips the characters a name could use to restructure or
// style the message, collapses whitespace, and caps the length.
func flatten(s string) string {
	s = strings.Map(func(r rune) rune {
		switch r {
		case '\n', '\r', '`', '*', '_', '~', '|', '>', '@', '[', ']':
			return ' '
		}
		if r < 0x20 {
			return ' '
		}
		return r
	}, s)
	s = strings.Join(strings.Fields(s), " ")
	if utf8.RuneCountInString(s) > maxNameLen {
		s = strings.TrimSpace(string([]rune(s)[:maxNameLen]))
	}
	if s == "" {
		return "your table"
	}
	return s
}
