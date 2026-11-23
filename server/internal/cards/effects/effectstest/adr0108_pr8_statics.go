package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0108_pr8_statics_test.go — ADR 0108 PR 8 (#1906): the prevention
// statics whose additional effect mills, draws, exiles or strikes back.
// Each test proves the prevention, the additional effect's unit (per
// recipient or per source, in one damage instance) and what CR 615.12
// does to it, as the card's rulings say.

const (
	Pr8GloomSurgeon     = "48b8ab2d-9f3f-42b2-9c5d-7497bceae43e"
	Pr8AngelOfSuffering = "92380c57-0a92-48eb-b562-2be149b5792a"
	Pr8ImmortalCoil     = "c85fc672-12d8-4ab0-b18d-ccd590ac063c"
	Pr8Swans            = "b187aeeb-5cf5-4b73-a3ef-f39188d2ba33"
	Pr8Vindicator       = "9b5cfbb7-21ed-491d-b77d-547e4d30a7da"
	Pr8Mindskinner      = "2e3ec514-fc9e-4964-a493-2a6cc9e63630"
)

// pr8StaticsCreature puts a catalogued creature with real power and
// toughness onto the battlefield, sick-free.
func Pr8StaticsCreature(g *game.Game, owner uuid.UUID, name, oracle string, power, toughness int) uuid.UUID {
	return ApaPush(g, owner, owner, game.Card{Name: name, TypeLine: "Creature — Test", OracleID: oracle,
		Power: power, Toughness: toughness})
}
