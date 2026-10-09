package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cramped Vents // Access Maze — Enchantment — Room (ADR 0103):
//
//	Cramped Vents {3}{B}: "When you unlock this door, this Room deals 6
//	 damage to target creature an opponent controls. You gain life equal
//	 to the excess damage dealt this way."
//	Access Maze {5}{B}{B}: "Once during each of your turns, you may cast
//	 a spell from your hand by paying life equal to its mana value
//	 rather than paying its mana cost."
//
// Cramped Vents reads the creature's damage before and after the hit, so
// the excess (CR 120.10) is how far past lethal THIS damage pushed it
// and prevention is already accounted for. Access Maze is NOT
// implemented: a standing "pay life equal to mana value rather than the
// mana cost, once each of your turns" has no primitive (ADR 0103's
// table). The door is empty, which is weaker than printed.
func init() {
	Register(Room(RoomSpec{
		OracleID:     "1e847d69-e527-4bb9-a624-17c25ac4fed4",
		Name:         "Cramped Vents // Access Maze",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Access Maze's pay-life casting isn't implemented, so that half does nothing."},
		Left: Door{Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				WhenYouUnlockThisDoor(game.DoorLeft, "Cramped Vents — 6 damage to target creature an opponent controls; gain life equal to the excess",
					crampedVentsDamage),
				TargetCreature("target creature an opponent controls", OpponentControls())),
				ForTargets(DamageToTarget(0, 6))),
		}},
	}))
}

func crampedVentsDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	pastLethal := func() int {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			return 0
		}
		return max(0, c.DamageMarked-c.CurrentToughness())
	}
	before := pastLethal()
	if err := (DealDamage{Source: ctx.Source(), Target: id, Amount: 6}).Apply(ctx); err != nil {
		return err
	}
	if excess := pastLethal() - before; excess > 0 {
		return GainLife{Player: item.Controller, Amount: excess}.Apply(ctx)
	}
	return nil
}
