package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vigilante Justice — Enchantment {3}{R} (Commander 2020, S82 OLD, #5513):
//
//	"Whenever a Human you control enters, this enchantment deals 1 damage to
//	 any target."
//
// Bat Colony's entry trigger over a Human (a changeling counts, CR 702.73a),
// one trigger per Human that enters under your control. The damage is the
// enchantment's, dealt to the chosen target if it is still legal as the
// trigger resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9610176a-dcd6-4117-89db-45f56b16cdd6",
		Name:         "Vigilante Justice",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
					return ok && c.HasSubtype(source.TextSubtype("Human"))
				}, "Vigilante Justice — this enchantment deals 1 damage to any target",
					effects.SourceDealsDamageToEachLegalTarget(1)),
				effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 1))),
		},
	})
}
