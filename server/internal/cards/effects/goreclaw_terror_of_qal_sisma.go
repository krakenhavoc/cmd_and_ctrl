package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goreclaw, Terror of Qal Sisma — Legendary Creature — Bear {3}{G},
// 4/3:
//
//	"Creature spells you cast with power 4 or greater cost {2} less to
//	 cast.
//	 Whenever Goreclaw attacks, each creature you control with power 4
//	 or greater gets +1/+1 and gains trample until end of turn."
//
// The cost half reads the spell's printed power off the stack (CR
// 601.2f prices before the spell resolves, so a CDA that only becomes
// 4-power on the battlefield doesn't discount here — there is no
// printed CDA creature in the pool this matters for today). The
// combat half is a CR 611.2c one-shot: the affected set — creatures
// you control with power 4+ — is locked at the moment the trigger
// resolves, via BoostUntilEOT and GrantKeywordUntilEOT sharing one
// Match predicate, so a creature pumped to 4 power BY this trigger
// does not retroactively qualify for it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "befb211f-37ca-4083-98d4-9ff1f28be3f2",
		Name:         "Goreclaw, Terror of Qal Sisma",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(2, "Creature spells you cast with power 4 or greater cost {2} less to cast.",
				YourSpell(), CreatureSpell(), goreclawSpellPowerAtLeast4()),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Goreclaw, Terror of Qal Sisma — pump and trample for power-4 creatures",
				goreclawAttackEffect),
		},
	})
}

// goreclawSpellPowerAtLeast4 reads the spell's printed power — CR
// 601.2f prices the cast, not the battlefield object it will become.
func goreclawSpellPowerAtLeast4() CostPredicate {
	return func(q game.CostQuery) bool { return q.Card.Power >= 4 }
}

// goreclawPowerAtLeast4 is the battlefield-side twin: "each creature
// you control with power 4 or greater" (CR 611.2c, checked once as
// the trigger resolves).
func goreclawPowerAtLeast4() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.CurrentPower() >= 4 }
}

func goreclawAttackEffect(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	match := And(Creature(), YouControl(), goreclawPowerAtLeast4())
	if err := (BoostUntilEOT{
		Match: match,
		Power: 1, Toughness: 1,
		Label: "Goreclaw, Terror of Qal Sisma — +1/+1",
	}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{
		Match:    match,
		Keywords: []string{"trample"},
		Label:    "Goreclaw, Terror of Qal Sisma — trample",
	}.Apply(ctx)
}
