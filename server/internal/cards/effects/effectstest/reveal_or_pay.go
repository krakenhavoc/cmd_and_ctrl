package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reveal_or_pay_test.go — ADR 0100 amendment 2026-10-07: the reveal and
// behold branch components, and the fifteen cards that carry them. The
// component rules are tested on Wren's Run Vanquisher (reveal) and
// Silvergill Mentor (behold); the rest get one test for what is theirs.

const (
	VanquisherOracle     = "fe9cdd15-a390-4a1e-bae9-474ce8d355b8"
	MentorOracle         = "dcf07f38-422f-46c3-aee9-50540b9c9115"
	SkymarcherOracle     = "9e1fe5ec-1669-4587-8e42-b189fd6d39e0"
	AdeptOracle          = "cb66c5ec-e7a4-4bd2-b825-6c9b846ba40d"
	PieSneakOracle       = "7243b935-73fc-47d2-ac65-ab390ed94f3f"
	ThunderherdOracle    = "006bbd8b-2007-419d-8b84-9ff73f2f95b7"
	CausticExhaleOracle  = "c4023e1f-26e3-4225-98d0-3d33f6654616"
	AspirantOracle       = "845d80ee-1cd5-437a-90a5-40dcc657d3aa"
	LysAlanaOracle       = "ce7582a2-e014-41fb-8753-00d29cffc81a"
	CursetosserOracle    = "b63e5518-61f3-4cde-a799-5f3ee6aba70b"
	SoulbrightOracle     = "762679dd-4345-43c3-929b-15959443aeb0"
	SurtlandOracle       = "b1a3b2ab-d9f8-4a0a-b2e2-aaa9989b2e41"
	RevealPlainOracle    = "1e6f2997-a84b-4302-9d8b-ca4928961d24" // Daring Buccaneer
	BladewhirlOracle     = "c37770aa-4770-4198-b6a6-f9ff86dbb06b"
	GoldmeadowOracle     = "1aff8653-12c5-4549-915f-ac2fd6dfd43b"
	RevealBranchIdx      = 0
	ManaBranchIdx        = 1
	VanquisherManaCost   = "{1}{G}"
	VanquisherTypeLine   = "Creature — Elf Warrior"
	MentorManaCost       = "{1}{U}"
	MentorTypeLine       = "Creature — Merfolk Wizard"
	RevealTestHandName   = "Elf Friend"
	RevealTestHandType   = "Creature — Elf Druid"
	RevealTestOtherType  = "Creature — Goblin"
	RevealTestMerfolkTyp = "Creature — Merfolk"
)

// revealedEvents counts the EventRevealCards the table has seen for id.
func RevealedEvents(g *game.Game, id uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventRevealCards && ev.CardID == id {
			n++
		}
	}
	return n
}
