package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emergent Haunting — Enchantment {1}{U} (Outlaws of Thunder Junction,
// #4468):
//
//	"At the beginning of your end step, if you haven't cast a spell from
//	 your hand this turn and this enchantment isn't a creature, it
//	 becomes a 3/3 Spirit creature with flying in addition to its other
//	 types.
//	 {2}{U}: Surveil 1."
//
// Both halves of the "if" are an intervening if (CR 603.4): asked as the
// end step begins and again as the ability resolves
// (AtYourEndStepIfNoSpellFromHand, #5235, with "isn't a creature" beside
// it), so a spell cast from the hand in response, or the enchantment
// becoming a creature some other way, turns it off. "In addition to its
// other types" keeps it an enchantment (CR 205.1b); the change has no
// duration, so it lasts as long as the permanent is this object (CR
// 611.2c, 400.7), and once it is a creature the ability stops triggering.
// The colour is untouched: it stays blue.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9f4d0133-0ecc-4498-8a47-4906321faae2",
		Name:         "Emergent Haunting",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStepIfNoSpellFromHand("Emergent Haunting — becomes a 3/3 Spirit creature with flying",
				effects.Do(effects.If{Cond: condThisIsNotACreature{}, Then: []effects.Applier{effects.Instruction(emergentHauntingBecomesASpirit)}}),
				thisIsNotACreature),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{U}"),
			Effect:  effects.Do(effects.Surveil{N: 1}),
		}},
	})
}

// thisIsNotACreature is the When for "this permanent isn't a creature"
// as the trigger fires.
func thisIsNotACreature(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return !source.IsCreature()
}

// condThisIsNotACreature is "this permanent isn't a creature" re-checked
// as the item resolves: false once the source has left the battlefield
// or come back as a new object. It declares no footprint: whether the
// source is a creature is read through the layers, which the
// trigger-independence check does not model for a condition, so a row
// that carries it is simply not reordered.
type condThisIsNotACreature struct{}

func (condThisIsNotACreature) Holds(ctx *effects.Context) bool {
	if !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return false
	}
	c, ok := ctx.Game.LookupCardForEffect(ctx.Item.SourceCardID)
	return ok && !c.IsCreature()
}

func (condThisIsNotACreature) Condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}

// emergentHauntingBecomesASpirit is "it becomes a 3/3 Spirit creature
// with flying in addition to its other types", pinned to this object.
func emergentHauntingBecomesASpirit(ctx *effects.Context) error {
	self := ctx.Item.SourceCardID
	return effects.ScopedEffectFor{
		Target: self,
		Mods: []game.Mod{
			game.AddTypesMod("Creature"),
			game.AddSubtypesMod(ctx.TextSubtypes("Spirit")...),
			game.SetBasePowerMod(3),
			game.SetBaseToughnessMod(3),
			game.AddKeywordsMod("flying"),
		},
		Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), self),
		Label:    "Emergent Haunting — becomes a 3/3 Spirit creature with flying",
	}.Apply(ctx)
}
