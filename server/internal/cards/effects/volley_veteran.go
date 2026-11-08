package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Volley Veteran — Creature — Goblin Warrior {3}{R}, 4/2 (Foundations,
// #4465):
//
//	"When this creature enters, it deals damage to target creature an
//	 opponent controls equal to the number of Goblins you control."
//
// Warmaker Gunship's enters damage with a Goblin count: the permanents
// you control that are Goblins, read as the trigger resolves (CR
// 608.2h), so the Veteran counts itself while it is still on the
// battlefield and a changeling counts. "Goblins" is read through the
// Veteran's text (CR 612.2). The damage is the Veteran's. With no
// creature an opponent controls, the trigger is removed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bf4268e6-170a-445b-b215-718f7494ae28",
		Name:         "Volley Veteran",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Volley Veteran — damage equal to the number of Goblins you control to target creature an opponent controls",
					volleyVeteranDamage),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}

// volleyVeteranDamage deals the controller's Goblin count, read now, to
// the announced creature if it is still legal.
func volleyVeteranDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	goblin := ctx.TextSubtype("Goblin")
	n := countControlled(g, item.Controller, func(c game.Card) bool { return c.HasSubtype(goblin) })
	id, ok := b16FirstLegalTargetCard(ctx)
	if n <= 0 || !ok {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: id, Amount: n}.Apply(ctx)
}
