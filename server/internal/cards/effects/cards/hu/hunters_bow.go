package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hunter's Bow — Artifact — Equipment {1}{G} (Assassin's Creed, #5948):
//
//	"When this Equipment enters, attach it to target creature you control.
//	 That creature deals damage equal to its power to up to one target
//	 creature you don't control.
//	 Equipped creature has reach and ward {2}.
//	 Equip {1}"
//
// One entry trigger with two target clauses (#764): slot 0 the creature you
// control, which the Bow attaches to and which then bites; slot 1 an "up to
// one" creature you don't control. The attach goes to slot 0 only, never to a
// leftover slot-1 pick. The damage is the biter's power as the trigger
// resolves, and nothing is dealt when either pick is gone or none was made
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "810592fe-65f5-4c95-9d75-d6806e004903",
		Name:         "Hunter's Bow",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.GrantToAttached("reach")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Hunter's Bow — attach it to target creature you control; it deals damage equal to its power to up to one target creature you don't control", huntersBowBite),
				effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.UpToOneTargetCreature("up to one target creature you don't control", effects.OpponentControls()),
				),
			),
			effects.WardAttached(effects.WardMana("{2}"), "Hunter's Bow — equipped creature's ward {2}"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}

// huntersBowBite attaches the Bow to slot 0's creature, then has that creature
// deal damage equal to its power to slot 1's creature.
func huntersBowBite(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	mine, ok := ctx.ClauseTarget(0)
	if !ok || mine.Kind != game.TargetCard {
		return nil
	}
	if err := g.AttachSourceForEffect(item, mine); err != nil {
		return err
	}
	victim, ok := ctx.ClauseTarget(1)
	if !ok || victim.Kind != game.TargetCard {
		return nil
	}
	c, ok := g.LookupCardForEffect(mine.ID)
	if !ok {
		return nil
	}
	return effects.DealDamage{Source: mine.ID, Target: victim.ID, Amount: c.CurrentPower()}.Apply(ctx)
}
