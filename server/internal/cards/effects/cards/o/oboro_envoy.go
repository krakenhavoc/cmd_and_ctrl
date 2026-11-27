package cardso

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oboro Envoy — Creature — Moonfolk Wizard {3}{U}, 1/3 (Saviors of
// Kamigawa, #5601):
//
//	"Flying
//	 {2}, Return a land you control to its owner's hand: Target creature
//	 gets -X/-0 until end of turn, where X is the number of cards in your
//	 hand."
//
// The land is returned as a cost, named at announce (CR 602.2b), and goes
// to hand before X is counted, so the land itself counts. X is read as
// the ability resolves (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "be70c6e8-6f9f-49fb-ab40-c6ce0ec2077c",
		Name:            "Oboro Envoy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Return a land you control to its owner's hand: Target creature gets -X/-0 until end of turn, where X is the number of cards in your hand.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Targets: game.ForTargets(effects.HarmUntilEOT(0, 0, 0))},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.ReturnAPermanentToHand("a land you control", effects.Land())),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id := effects.FirstLegalBattlefieldTarget(ctx)
				if id == uuid.Nil {
					return nil
				}
				return effects.BoostUntilEOT{Target: id, Power: -effects.SokB1CardsInYourHand(ctx), Label: "Oboro Envoy — -X/-0"}.Apply(ctx)
			},
		}},
	})
}
