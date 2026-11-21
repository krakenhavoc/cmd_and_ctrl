package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Firemane Avenger — Creature — Angel {2}{R}{W}, 3/3 (Ravnica: Clue
// Edition):
//
//	"Flying
//	 Battalion — Whenever this creature and at least two other creatures
//	 attack, this creature deals 3 damage to any target and you gain 3
//	 life."
//
// Battalion is Boros Strike-Captain's group test. The damage reads the
// Avenger by last-known information if it has left (CR 608.2h); the life
// gain happens whenever the ability resolves with a legal target.
//
// No simplification.
func init() {
	const label = "Firemane Avenger — 3 damage to any target and you gain 3 life (battalion)"
	damage := effects.SourceDealsDamageToEachLegalTarget(3)
	effects.Register(effects.Spec{
		OracleID:        "7ea71a36-8fa8-4ba3-9cb1-7fc6917c3ddd",
		Name:            "Firemane Avenger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.On(game.EventAttack, effects.BattalionAttack, label,
				func(g *game.Game, item *game.StackItem) error {
					if err := damage(g, item); err != nil {
						return err
					}
					return effects.GainLife{Amount: 3}.Apply(effects.NewContext(g, item))
				}), effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 3))),
		},
	})
}
