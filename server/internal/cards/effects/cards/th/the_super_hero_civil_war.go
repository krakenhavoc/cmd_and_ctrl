package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Super Hero Civil War — {3}{R}{W} Enchantment — Saga (Marvel Super
// Heroes, #4455, #4616):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Gain control of up to two target creatures with total mana
//	     value 6 or less for as long as this Saga remains on the
//	     battlefield.
//	 II — Creatures you control get +1/+1 and gain vigilance until end
//	      of turn.
//	 III — Target creature you control fights up to one other target
//	       creature."
//
// Chapter I is The Akroan War's theft, twice, under the total set rule
// (TotalManaValueAtMost, #4616): the two creatures' mana values summed
// may not exceed 6 when the chapter goes on the stack (CR 601.2c via
// CR 603.3d), and the creatures still legal are summed again as it
// resolves (CR 608.2b). If the total has grown past 6 by then (a copy
// effect, a face turned), none of them is taken: the printed rulings
// on this family say the entire selection of targets is illegal. Each
// theft is keyed to the Saga as the object it is now (CR 400.7), so
// both creatures go home when the Saga is sacrificed after chapter III,
// or earlier if it leaves.
//
// Chapter II is The Argent Etchings' chapter II with vigilance.
//
// Chapter III is two clauses, the second "up to one" and Distinct from
// the first ("other"). With no second target, or one that has become
// illegal, nothing fights (CR 701.14b).
//
// No simplifications.
func init() {
	const name = "The Super Hero Civil War"
	effects.Register(effects.Spec{
		OracleID:     "11d74dc2-8c11-4d4b-b727-8216d2eb6b4e",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "gain control of up to two target creatures with total mana value 6 or less for as long as this Saga remains on the battlefield"),
				effects.TargetCreature("up to two target creatures with total mana value 6 or less").
					WithCount(0, 2).WithTotal(effects.TotalManaValueAtMost(6)),
				superHeroCivilWarSteal),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "creatures you control get +1/+1 and gain vigilance until end of turn"),
				superHeroCivilWarRally),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "target creature you control fights up to one other target creature"),
				effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.Distinct(effects.TargetCreature("up to one other target creature").WithCount(0, 1)),
				),
				effects.SlotZeroFightsSlotOne),
		},
	})
}

// superHeroCivilWarSteal is chapter I: control of each target still
// legal, for as long as the Saga remains.
func superHeroCivilWarSteal(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	d, ok := effects.DurationWhileSourceRemains(ctx, item.SourceCardID)
	if !ok {
		// CR 611.2b: the Saga is already gone, so the effect never
		// begins.
		return nil
	}
	for _, t := range targets {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (effects.GainControl{
			Target:     t.ID,
			Controller: item.Controller,
			Duration:   d,
			Label:      "The Super Hero Civil War — control for as long as the Saga remains",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// superHeroCivilWarRally is chapter II.
func superHeroCivilWarRally(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	yours := effects.And(effects.Creature(), effects.YouControl())
	if err := (effects.BoostUntilEOT{
		Match:     yours,
		Power:     1,
		Toughness: 1,
		Label:     "The Super Hero Civil War — +1/+1",
	}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{
		Match:    yours,
		Keywords: []string{"vigilance"},
		Label:    "The Super Hero Civil War — vigilance",
	}.Apply(ctx)
}
