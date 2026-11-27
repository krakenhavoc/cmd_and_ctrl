package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mogg Bombers — Creature — Goblin {3}{R}, 3/4 (Stronghold, S82 OLD batch 747,
// #5627):
//
//	"When another creature enters, sacrifice this creature and it deals 3
//	 damage to target player or planeswalker."
//
// Soul Warden's "another creature enters" (any creature under any player's
// control, one trigger each) with the target chosen as the trigger goes on the
// stack. As it resolves the Bombers is sacrificed while it is still the same
// permanent (CR 400.7), and deals the damage as it last existed (CR 608.2h)
// whether or not it could be sacrificed; the target is read again then
// (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Mogg Bombers — sacrifice it and it deals 3 damage to target player or planeswalker"
	effects.Register(effects.Spec{
		OracleID:     "69812cd2-9885-45ed-9d79-658597bfeb8b",
		Name:         "Mogg Bombers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.On(game.EventETB, effects.AnotherCreatureEntered, label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := effects.SacrificeThisIfStillHere(ctx); err != nil {
					return err
				}
				for _, t := range ctx.LegalTargets() {
					return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 3}.Apply(ctx)
				}
				return nil
			}), effects.TargetPlayerOrPlaneswalker()), effects.ForTargets(effects.DamageToTarget(0, 3))),
		},
	})
}
