package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Geistcatcher's Rig — Artifact Creature — Construct {6}, 4/5 (Innistrad
// Remastered, #5948):
//
//	"When this creature enters, you may have it deal 4 damage to target creature
//	 with flying."
//
// An enters trigger with a target (chosen as it goes on the stack, CR 603.3d)
// and a "you may" asked as it resolves. The damage is dealt by the Rig itself,
// or by its last-known self if it has left (CR 608.2h), to a creature with flying
// that is still legal.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1bfdbb96-cc18-4609-ba17-d169967df8d0",
		Name:         "Geistcatcher's Rig",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(
				Targeting(
					WhenThisEnters("Geistcatcher's Rig — it deals 4 damage to target creature with flying",
						func(g *game.Game, item *game.StackItem) error {
							ctx := NewContext(g, item)
							t, ok := firstLegalTarget(ctx)
							if !ok {
								return nil
							}
							return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 4}.Apply(ctx)
						}),
					TargetCreature("target creature with flying", HasKeyword("flying"))),
				"Geistcatcher's Rig — have it deal 4 damage to target creature with flying?"),
		},
	})
}
