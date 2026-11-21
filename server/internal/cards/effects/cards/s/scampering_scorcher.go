package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scampering Scorcher — Creature — Elemental {3}{R}, 1/1 (Core Set 2020, #5518):
//
//	"When this creature enters, create two 1/1 red Elemental creature tokens.
//	 Elementals you control gain haste until end of turn. (They can attack
//	 and {T} this turn.)"
//
// The tokens are made first, so they are among the Elementals that gain
// haste; the set is fixed as the trigger resolves (CR 611.2c), so an
// Elemental that arrives later this turn does not have it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4899595e-e626-432c-9c9f-f9d5c243a64b",
		Name:         "Scampering Scorcher",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Scampering Scorcher — create two 1/1 red Elemental tokens; Elementals you control gain haste",
				effects.Do(
					effects.CreateToken{Template: effects.TokenCard("1/1 red Elemental"), N: 2},
					effects.GrantKeywordUntilEOT{Match: effects.And(effects.Subtype("Elemental"), effects.YouControl()), Keywords: []string{"haste"},
						Label: "Scampering Scorcher — haste"},
				)),
		},
	})
}
