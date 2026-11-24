package cardsst

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Surtland Flinger — Creature — Giant Berserker {3}{R}{R}, 4/6 (Kaldheim):
//
//	"Whenever this creature attacks, you may sacrifice another creature. When
//	 you do, this creature deals damage equal to the sacrificed creature's
//	 power to any target. If the sacrificed creature was a Giant, this creature
//	 deals twice that much damage instead."
//
// Bolg of the North's shape: the attack trigger asks "you may sacrifice another
// creature" as it resolves (CR 608.2c), and a creature that really went raises
// the printed "when you do" as a CR 603.12 reflexive trigger, which chooses its
// own "any target" as it goes on the stack and carries the sacrificed creature
// as its payload. The damage is the Flinger's, equal to the sacrificed
// creature's power as it last stood (CR 608.2h), doubled when it was a Giant
// then (or a changeling, every creature type). A target that is gone leaves
// nothing dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f701ada1-e9e1-42ce-9a62-c11bcec03da7",
		Name:         "Surtland Flinger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Surtland Flinger — you may sacrifice another creature",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Surtland Flinger — sacrifice another creature to deal damage equal to its power to any target?", effects.Creature(), true, surtlandFlingerSacrificed)
				}),
		},
	})
}

// surtlandFlingerSacrificed raises the reflexive trigger once a creature has
// really gone. An on-disk key.
var surtlandFlingerSacrificed = effects.RegisterSacrificedOne("prompt-run/s82-old-b064-surtland-flinger-sacrifice", func(ctx *effects.Context, sacrificed uuid.UUID) error {
	return effects.ReflexiveTrigger{
		Label: "Surtland Flinger — damage equal to the sacrificed creature's power to any target (twice that if it was a Giant)",
		Cards: []uuid.UUID{sacrificed},
		Body:  surtlandFlingerDamageBody,
	}.Apply(ctx)
})

// surtlandFlingerDamageBody is the reflexive trigger's registration: the
// sacrificed creature is the payload, the victim the trigger's own any-target
// clause.
var surtlandFlingerDamageBody = game.ReflexiveBody("s82-old-b064/surtland-flinger-damage", effects.SimpleBody(surtlandFlingerDamage),
	func(uuid.UUID, game.EffectParams) *game.TargetSpec { return effects.TargetAny() })

func surtlandFlingerDamage(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	sacrificed := ctx.PayloadCards()
	targets := ctx.LegalTargets()
	if len(sacrificed) == 0 || len(targets) == 0 {
		return nil
	}
	amount := effects.B29PowerAsItLastStood(g, sacrificed[0])
	if info, ok := g.LastKnownPermanentForEffect(sacrificed[0]); ok &&
		(info.Characteristic.AllCreatureTypes || slices.Contains(info.Characteristic.Subtypes, ctx.TextSubtype("Giant"))) {
		amount *= 2
	}
	return effects.DealDamage{Source: item.SourceCardID, Target: targets[0].ID, Amount: amount}.Apply(ctx)
}
