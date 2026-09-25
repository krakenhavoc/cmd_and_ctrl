package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pitiless Plunderer — Creature — Human Pirate {3}{B}, 1/4:
//
//	"Whenever another creature you control dies, create a Treasure
//	token."
//
// The card that turns an aristocrats board into a mana engine: every
// sacrifice refunds a mana, so a sacrifice outlet with a mana cost
// becomes free and the loop keeps going. Treasure already cracks for
// mana through S21 sub-PR 1's sacrifice-cost mana abilities, so this
// needs nothing new.
//
// "ANOTHER creature you control" — the Plunderer's own death does not
// trigger it, unlike Zulaport Cutthroat's "this creature or another".
//
// Audited for #1565: the printed text is one unconditional,
// non-optional trigger and that is the whole card — no additional
// clause is left unimplemented. See staples3_test.go for the "own
// death doesn't count" and "an opponent's death doesn't count" cases
// alongside the main line.
func init() {
	Register(Spec{
		OracleID:     "a784481f-eccb-4112-bb38-04a659319660",
		Name:         "Pitiless Plunderer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return anotherCreatureYouControlDied(ev, source, g)
			}, "Pitiless Plunderer — create a Treasure", Do(CreateToken{
				Template: TreasureToken(),
				N:        1,
			})),
		},
	})
}
