package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bearer of Silence — {1}{B} Creature — Eldrazi 2/1 (Oath of the Gatewatch,
// S82 OLD b296):
//
//	"Devoid (This card has no color.)
//	 When you cast this spell, you may pay {1}{C}. If you do, target opponent
//	 sacrifices a creature.
//	 Flying
//	 This creature can't block."
//
// A cast trigger (CR 601.2i) that resolves above the spell. The opponent is
// aimed as it goes on the stack (CR 603.3d); the {1}{C} is asked as it
// resolves, and only colorless mana pays the {C}. Paid, the opponent chooses a
// creature of theirs to sacrifice (CR 701.21).
//
// No simplification.
func init() {
	cast := effects.WhenYouCastThisSpell("Bearer of Silence — you may pay {1}{C}; if you do, target opponent sacrifices a creature",
		func(g *game.Game, item *game.StackItem) error {
			return effects.MayPay{
				Chooser:  item.Controller,
				Cost:     "{1}{C}",
				Question: "Bearer of Silence — pay {1}{C} to make target opponent sacrifice a creature?",
				Then:     bearerOfSilencePaid,
			}.Apply(effects.NewContext(g, item))
		})
	cast.Targets = effects.TargetPlayer("target opponent", effects.Opponent())
	effects.Register(effects.Spec{
		OracleID:        "d69c8c55-6433-4cc5-bdc7-9722f121b8e2",
		Name:            "Bearer of Silence",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "flying"},
		Static:          []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Triggered:       []game.TriggeredAbility{cast},
	})
}

// bearerOfSilencePaid is the paid branch. An on-disk key: never renamed, never
// reused.
var bearerOfSilencePaid = effects.PayUnlessBranches("pay-unless/bearer-of-silence-sacrifice",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetPlayer && t.ID != uuid.Nil {
				ctx.Game.PlayerSacrificesForEffect(ctx.Item.SourceCardID, t.ID, effects.SacrificeSpec("a creature", effects.Creature()),
					"Bearer of Silence — sacrifice a creature")
				return nil
			}
		}
		return nil
	}, nil)
