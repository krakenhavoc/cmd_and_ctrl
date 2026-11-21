package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Weakstone's Subjugation — Enchantment — Aura {U} (The Brothers' War):
//
//	"Enchant artifact or creature
//	 When this Aura enters, you may pay {3}. If you do, tap enchanted
//	 permanent.
//	 Enchanted permanent doesn't untap during its controller's untap step."
//
// Bitter Chill's pay-then-act shape, on an artifact or a creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c3745ff-6e1e-4660-abe6-1a30611ec00c",
		Name:         "Weakstone's Subjugation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("enchant artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{
			{Label: "enchanted permanent doesn't untap", Restricts: effects.AttachedToSource},
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Weakstone's Subjugation — you may pay {3}; if you do, tap enchanted permanent",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{3}",
						Question: "Weakstone's Subjugation — pay {3} to tap enchanted permanent?",
						Then:     weakstonesPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var weakstonesPaid = effects.PayUnlessBranches("pay-unless/weakstones-subjugation-tap",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		aura, ok := ctx.TriggeringPermanent()
		if !ok || aura.AttachedTo.Kind != game.TargetCard {
			return nil
		}
		return effects.TapTarget{Target: aura.AttachedTo.ID}.Apply(ctx)
	}, nil)
