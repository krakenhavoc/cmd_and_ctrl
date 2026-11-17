package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prava of the Steel Legion — Legendary Creature — Cat Soldier {2}{W}, 1/4
// (Commander Legends, #5507):
//
//	"During your turn, creature tokens you control get +1/+4.
//	 {3}{W}: Create a 1/1 white Soldier creature token.
//	 Partner (You can have two commanders if both have partner.)"
//
// The pump is a layer-7c static whose window is Anara's: on every
// recompute it asks whether Prava's controller is the active player, so it
// is live from your untap step through your cleanup (the damage removed
// there is still checked against the pumped toughness, as the ruling
// says) and absent on every other turn, and it stops the moment Prava
// leaves. The token row is Ironroot Warlord's. Partner is the deck rule
// internal/deck reads off the oracle text (ADR 0144).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4daa7b7d-e410-4bcf-b3e2-2751774bef47",
		Name:         "Prava of the Steel Legion",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "During your turn, creature tokens you control get +1/+4.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && IsToken(*target) &&
					target.Controller == source.Controller &&
					isActivePlayer(g, source.Controller)
			},
			Apply: pumpBy(1, 4),
		}},
		Activated: []ActivatedAbility{{
			Label:   "{3}{W}: Create a 1/1 white Soldier creature token.",
			Cost:    ManaCost("{3}{W}"),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1, CreatureTokens: 1},
			Effect:  Do(CreateToken{Template: TokenCard("1/1 white Soldier"), N: 1}),
		}},
	})
}
