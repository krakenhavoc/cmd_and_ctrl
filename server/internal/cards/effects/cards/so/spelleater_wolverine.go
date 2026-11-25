package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spelleater Wolverine — Creature — Wolverine {2}{R}, 3/2 (Ikoria: Lair
// of Behemoths, #5512):
//
//	"This creature has double strike as long as there are three or more
//	 instant and/or sorcery cards in your graveyard."
//
// Ghitu Lavarunner's shape: a layer 6 keyword grant to itself, read live
// from the controller's graveyard, so it turns off again when the cards
// leave.
//
// No simplification.
func init() {
	threeSpells := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID &&
			effects.GraveyardAtLeast(3, effects.MatchInstantOrSorcery)(g, source.Controller, source.InstanceID)
	}
	effects.Register(effects.Spec{
		OracleID:     "a3b226b9-f149-4578-882b-77447608380a",
		Name:         "Spelleater Wolverine",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.B16GrantKeywords(threeSpells, "double strike")},
	})
}
