package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shuriken — Artifact — Equipment {1} (Betrayers of Kamigawa, #5602):
//
//	"Equipped creature has "{T}, Unattach Shuriken: Shuriken deals 2 damage to
//	 target creature. That creature's controller gains control of Shuriken
//	 unless it was unattached from a Ninja."
//	 Equip {2}"
//
// Toralf's Hammer's shape: the equipped creature is granted an activated
// ability whose cost taps it and unattaches the Equipment (UnattachTheGrantor).
// The Shuriken itself is the damage source (the grantor, CR 120.3). The target
// creature's controller as the ability resolves gains control of the Shuriken
// for good (CR 611.2a, indefinite) unless the creature it was unattached from
// was a Ninja, read as that creature last existed (CR 608.2h): a changeling
// counts. If the target left before it resolves, the ability does nothing
// (CR 608.2b). If the Shuriken has left the battlefield, there is nothing to
// give.
//
// No simplification.
func init() {
	const grant = "shuriken/throw"
	effects.Register(effects.Spec{
		OracleID:     "c6e92e21-d3e7-4982-b9f3-af56d4014c07",
		Name:         "Shuriken",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}, Unattach Shuriken: Shuriken deals 2 damage to target creature. That creature's controller gains control of Shuriken unless it was unattached from a Ninja.",
				Purpose: game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 2)), Answers: game.AnswerRemove},
				Cost:    effects.Plus(effects.TapCost(), effects.UnattachTheGrantor()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  shurikenThrow,
			}},
			Text: "{T}, Unattach Shuriken: Shuriken deals 2 damage to target creature. That creature's controller gains control of Shuriken unless it was unattached from a Ninja.",
		}},
		Static:    []game.StaticAbility{effects.GrantAbilitiesToAttached(grant)},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

func shurikenThrow(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	shuriken := ctx.Grantor()
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	victim, ok := g.LookupCardForEffect(targets[0].ID)
	if !ok {
		return nil
	}
	if err := (effects.DealDamage{Source: shuriken, Target: victim.InstanceID, Amount: 2}).Apply(ctx); err != nil {
		return err
	}
	if shuriken == [16]byte{} || !effects.OnBattlefield(g, shuriken) {
		return nil
	}
	if info, ok := ctx.SourcePermanent(); ok {
		ch := info.Characteristic
		if ch.AllCreatureTypes {
			return nil
		}
		for _, st := range ch.Subtypes {
			if st == ctx.TextSubtype("Ninja") {
				return nil
			}
		}
	}
	// The creature's controller as the ability resolves (it may have died from
	// the damage: its last controller still gains the Shuriken).
	controller := victim.Controller
	if c, ok := g.LookupCardForEffect(victim.InstanceID); ok {
		controller = c.Controller
	}
	return effects.GainControl{Target: shuriken, Controller: controller, Duration: game.IndefiniteDuration(), Label: "Shuriken — gain control"}.Apply(ctx)
}
