package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Faller's Faithful — Creature — Human Wizard {2}{B}, 3/1 (Edge of
// Eternities, #4461):
//
//	"When this creature enters, destroy up to one other target creature.
//	 If that creature wasn't dealt damage this turn, its controller draws
//	 two cards."
//
// Tainted Treats' shape: the second sentence has no "if you do", so it
// doesn't depend on the destruction. An indestructible or regenerated
// target still draws its controller two cards if it was undamaged. Both
// the damage record (DealtDamageThisTurn, Covert Cutpurse's read of the
// turn's damage events) and the controller are read before the destroy,
// the last moment the creature is guaranteed findable (CR 608.2h). With
// no target chosen, or a target gone by resolution, nothing happens
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "df7500e8-76b4-49b3-a427-9a28ffa14637",
		Name:         "Faller's Faithful",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Faller's Faithful — destroy up to one other target creature; if it wasn't dealt damage this turn, its controller draws two",
				fallersFaithfulDestroy),
				Another(TargetCreature("up to one other target creature")).WithCount(0, 1)),
		},
	})
}

// fallersFaithfulDestroy destroys the target, then draws its controller
// two cards if it was undamaged this turn.
func fallersFaithfulDestroy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	undamaged := !DealtDamageThisTurn()(g, item.Controller, c)
	controller := c.Controller
	if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	if !undamaged {
		return nil
	}
	return DrawCards{Player: controller, N: 2}.Apply(ctx)
}
