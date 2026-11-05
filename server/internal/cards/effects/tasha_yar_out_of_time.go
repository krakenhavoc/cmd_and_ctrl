package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tasha Yar, Out of Time — Legendary Creature — Human Officer {1}{R},
// 2/1 (Star Trek, #4973):
//
//	"Haste
//	 When Tasha Yar dies, exile the top card of your library. Until the
//	 end of your next turn, you may play that card.
//	 {2}{R}, Exile this card from your graveyard: Create a token that's a
//	 copy of this card. Sacrifice it at the beginning of the next end
//	 step. Activate only as a sorcery."
//
// Haste is the printed keyword. The dies trigger is Atsushi's first mode
// with one card (ExileTopNUntilYourNextTurn): its controller is the one
// Tasha had as she died (last-known information, CR 608.2h), and the play
// window runs through the end of that player's next turn.
//
// The graveyard ability is encore's shape (encore.go) for a single token:
// an activated ability that works from the graveyard, whose cost exiles
// the card, at sorcery speed. The token is a copy of the card as it is in
// exile (CR 707.2), so it is a legendary 2/1 with haste and these same
// abilities — a token copy that dies triggers the dies ability again — and
// one delayed trigger sacrifices it at the beginning of the next end step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1a9607a2-58cb-4bb8-b454-4813528221f8",
		Name:            "Tasha Yar, Out of Time",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Tasha Yar, Out of Time — exile the top card; you may play it until the end of your next turn",
				Do(Instruction(func(ctx *Context) error { return ExileTopNUntilYourNextTurn(ctx, 1) }))),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}{R}, Exile this card from your graveyard: Create a token that's a copy of this card. " +
				"Sacrifice it at the beginning of the next end step. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{2}{R}"), ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect:       trekB18TokenCopyOfThisSacrificedAtNextEndStep("Tasha Yar, Out of Time — sacrifice the token"),
		}},
	})
}
