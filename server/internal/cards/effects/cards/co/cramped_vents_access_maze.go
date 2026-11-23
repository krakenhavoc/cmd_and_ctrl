package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
// and prevention is already accounted for. Access Maze is a granted
// alternative cost behind its door (#3174, ADR 0118 §3): life equal to
// the spell's mana value, from hand, once during each of your turns per
// Access Maze. A spell with {X} is cast with X = 0 (CR 107.3b), and a
// spell with no mana cost costs 0 life (CR 118.6).
func init() {
	effects.Register(effects.Room(effects.RoomSpec{
		OracleID:     "1e847d69-e527-4bb9-a624-17c25ac4fed4",
		Name:         "Cramped Vents // Access Maze",
		Completeness: effects.CompletenessFull,
		Left: effects.Door{Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WhenYouUnlockThisDoor(game.DoorLeft, "Cramped Vents — 6 damage to target creature an opponent controls; gain life equal to the excess",
					crampedVentsDamage),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
				effects.ForTargets(effects.DamageToTarget(0, 6))),
		}},
		Right: effects.Door{GrantedAlternativeCosts: []game.GrantedAlternativeCost{effects.PayLifeForASpellFromHandOnceEachOfYourTurns()}},
	}))
}

func crampedVentsDamage(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
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
	if err := (effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 6}).Apply(ctx); err != nil {
		return err
	}
	if excess := pastLethal() - before; excess > 0 {
		return effects.GainLife{Player: item.Controller, Amount: excess}.Apply(ctx)
	}
	return nil
}
