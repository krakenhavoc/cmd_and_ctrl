package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	damage := sourceDealsDamageToEachLegalTarget(3)
	Register(Spec{
		OracleID:        "7ea71a36-8fa8-4ba3-9cb1-7fc6917c3ddd",
		Name:            "Firemane Avenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(On(game.EventAttack, battalionAttack, label,
				func(g *game.Game, item *game.StackItem) error {
					if err := damage(g, item); err != nil {
						return err
					}
					return GainLife{Amount: 3}.Apply(NewContext(g, item))
				}), TargetAny()), ForTargets(DamageToTarget(0, 3))),
		},
	})
}
