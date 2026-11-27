package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skyfire Kirin — Legendary Creature — Kirin Spirit {2}{R}{R}, 3/3
// (Saviors of Kamigawa, #5601):
//
//	"Flying
//	 Whenever you cast a Spirit or Arcane spell, you may gain control of
//	 target creature with that spell's mana value until end of turn."
//
// The target clause is the cast event's Amount (the spell's mana value as
// cast, CR 202.3), read through ManaValueEqualsEventAmount as the
// trigger goes on the stack and again as it resolves (CR 608.2b). The
// "may" is asked as it resolves (CR 603.5). The control lasts until end
// of turn (CR 611.2b); nothing untaps it and it gains no haste.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "169c468e-0e17-450b-8954-df47d78ad8f8",
		Name:            "Skyfire Kirin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Optional(effects.WheneverYouCast(effects.ChkB1SpiritOrArcane(), "Skyfire Kirin — gain control of target creature with that spell's mana value until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id := effects.FirstLegalBattlefieldTarget(ctx)
					if id == uuid.Nil {
						return nil
					}
					return effects.GainControl{Target: id, Duration: effects.DurationUntilEndOfTurn(ctx), Label: "Skyfire Kirin — gain control until end of turn"}.Apply(ctx)
				}), "Skyfire Kirin — gain control of the target creature until end of turn?"),
				effects.ManaValueEqualsTheEventAmount(effects.TargetCreature("target creature with that spell's mana value"))),
		},
	})
}
