package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Genasi Enforcers — Creature — Elemental Shaman {1}{R}, 1/3 (Commander Legends:
// Battle for Baldur's Gate):
//
//	"Myriad (Whenever this creature attacks, for each opponent other than defending
//	 player, you may create a token copy that's tapped and attacking that player or a
//	 planeswalker they control. Exile the tokens at end of combat.)
//	 {1}{R}: Creatures you control named Genasi Enforcers get +1/+0 until end of turn."
//
// Myriad is the shared trigger (CR 702.116); the token copies keep the name, so the
// ability pumps them too. The set affected is the creatures you control named Genasi
// Enforcers as the ability resolves, locked then (CR 611.2c), compared on the
// effective name (CR 201.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b3601821-a0f5-4aac-b24e-25498a82d821",
		Name:         "Genasi Enforcers",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.GrantedMyriad("Genasi Enforcers")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: Creatures you control named Genasi Enforcers get +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl(), effects.GapB219NamedCard("Genasi Enforcers")),
					Power: 1,
					Label: "Genasi Enforcers — +1/+0 until end of turn",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
