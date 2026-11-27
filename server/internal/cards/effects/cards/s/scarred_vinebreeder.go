package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scarred Vinebreeder — Creature — Elf Shaman {1}{B}, 1/1 (Lorwyn, S82 OLD,
// #5593):
//
//	"{2}{B}, Exile an Elf card from your graveyard: This creature gets +3/+3
//	 until end of turn."
//
// Scarblade Elite's graveyard exile cost (the card named at announce, read by
// its subtypes so a changeling pays it) with mana beside it, and a self pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "759e454e-ee8c-4b0e-9503-26e4c0a2a8f9",
		Name:         "Scarred Vinebreeder",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{2}{B}, Exile an Elf card from your graveyard: This creature gets +3/+3 until end of turn.",
			Purpose: game.Purpose{
				Answers: game.AnswerPump,
				Pump:    &game.Pump{Power: 3, Toughness: 3},
			},
			Cost: effects.Plus(effects.ManaCost("{2}{B}"), effects.ExileFromGraveyard(1, "an Elf card", func(c game.Card) bool {
				return c.HasSubtypeWord("Elf")
			})),
			Effect: effects.ThisGetsUntilEndOfTurn(3, 3, "Scarred Vinebreeder — +3/+3"),
		}},
	})
}
