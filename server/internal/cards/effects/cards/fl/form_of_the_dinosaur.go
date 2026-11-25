package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Form of the Dinosaur — Enchantment {4}{R}{R} (Rivals of Ixalan, S82 OLD batch
// 217, #5531):
//
//	"When this enchantment enters, your life total becomes 15.
//	 At the beginning of your upkeep, this enchantment deals 15 damage to target
//	 creature an opponent controls and that creature deals damage equal to its
//	 power to you."
//
// The enters trigger sets your life total by gaining or losing the difference
// (CR 119.5, Oketra's Last Mercy's B31LifeBecomes), so a prevention or a
// "can't gain life" applies. The upkeep trigger's target is re-checked as it
// resolves (CR 608.2b). The creature's power is read before the 15 damage, as
// the creature last existed on the battlefield (CR 608.2h), and it deals that
// damage to you as the object it was even when the 15 damage killed it
// (damage doesn't change power); a creature that was exiled in response is not
// a target any longer and the ability does nothing.
//
// No simplification.
func init() {
	const upkeep = "Form of the Dinosaur — deal 15 damage to target creature an opponent controls; it deals damage equal to its power to you"
	effects.Register(effects.Spec{
		OracleID:     "68af06fd-a715-47c6-b439-c9fcb1b01a1d",
		Name:         "Form of the Dinosaur",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Form of the Dinosaur — your life total becomes 15", func(g *game.Game, item *game.StackItem) error {
				return effects.B31LifeBecomes(effects.NewContext(g, item), item.Controller, 15)
			}),
			effects.Targeting(effects.AtYourUpkeep(upkeep, formOfTheDinosaurStrike),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}

// formOfTheDinosaurStrike is the upkeep body: 15 damage to the target, then the
// target deals damage equal to its power to the Form's controller.
func formOfTheDinosaurStrike(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var target uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			target = t.ID
		}
	}
	if target == uuid.Nil {
		return nil
	}
	if _, ok := g.LookupCardForEffect(target); !ok {
		return nil
	}
	g.RecomputeLayersIfStaleLocked()
	c, ok := g.LookupCardForEffect(target)
	if !ok {
		return nil
	}
	power := c.CurrentPower()
	ref := effects.DamagedCreatureRef(g, target)
	if err := (effects.DealDamage{Source: item.SourceCardID, Target: target, Amount: 15}).Apply(ctx); err != nil {
		return err
	}
	if power <= 0 {
		return nil
	}
	return effects.DealDamage{SourceObject: &ref, Target: item.Controller, Amount: power}.Apply(ctx)
}
