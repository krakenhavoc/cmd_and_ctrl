package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// targets_test.go — S20 sub-PR 1 catalog coverage: the predicate
// constructors produce the right legal sets, announce rejects
// illegal picks, and the migrated cards keep working.

const DoomBladeOracle = "59e7f2ae-4535-4191-98be-3e65b6b2befa"

func PushTypedCard(g *game.Game, owner uuid.UUID, name, typeLine, manaCost string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
	return id
}

func LegalCards(g *game.Game, caster uuid.UUID, oracle string) map[uuid.UUID]bool {
	lt := g.LegalTargetsFor(game.SourceChooser(caster), oracle)
	out := map[uuid.UUID]bool{}
	if lt == nil {
		return out
	}
	for _, id := range lt.Cards {
		out[id] = true
	}
	return out
}
