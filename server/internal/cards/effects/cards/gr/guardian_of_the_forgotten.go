package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guardian of the Forgotten — Creature — Elephant Warrior {3}{W}, 4/4
// (Modern Horizons 3):
//
//	"Vigilance
//	 Whenever a modified creature you control dies, manifest the top card
//	 of your library. (Equipment, Auras you control, and counters are
//	 modifications. To manifest a card, put it onto the battlefield face
//	 down as a 2/2 creature. Turn it face up any time for its mana cost if
//	 it's a creature card.)"
//
// Costume Closet's "modified creature you control dies" (the creature is
// read as it last existed on the battlefield, CR 700.9 and 603.10a, so a
// creature that dies with a counter or attachment counts, the Guardian
// itself included) with Cryptic Pursuit's manifest of the top card of the
// library.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "89b88a6c-92be-43eb-9de4-f79cc75c6d90",
		Name:            "Guardian of the Forgotten",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.ACreatureYouControlDied(ev, source, lki, g) && effects.Msc17WasModified(g, ev.CardID)
			}, "Guardian of the Forgotten — manifest the top card of your library", effects.Do(effects.Manifest{N: 1})),
		},
	})
}
