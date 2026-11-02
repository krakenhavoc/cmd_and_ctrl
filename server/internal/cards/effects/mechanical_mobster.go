package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mechanical Mobster — Artifact Creature — Human Robot Villain {3}, 2/1
// (Marvel's Spider-Man, #4460):
//
//	"When this creature enters, exile up to one target card from a
//	 graveyard. Target creature you control connives. (Draw a card, then
//	 discard a card. If you discarded a nonland card, put a +1/+1 counter
//	 on that creature.)"
//
// One trigger with two target clauses (Stunning Shot's slots): clause 0 is
// "up to one target card from a graveyard", clause 1 is "target creature you
// control" (Mobster herself is always a legal pick while she is out). Each
// clause is judged separately at resolution (CR 608.2b): a graveyard card
// that has gone does not stop the connive, and a creature that has left
// does not stop the exile. The exile comes first, as printed; the connive
// is ConniveEach for the chosen creature, the controller's draw and discard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b0672aa3-3207-4ff8-af8f-bdabacbb4d04",
		Name:         "Mechanical Mobster",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Mechanical Mobster — exile up to one target card from a graveyard; target creature you control connives",
					mechanicalMobsterETB),
				Clauses(
					TargetCardInGraveyard("up to one target card from a graveyard").WithCount(0, 1),
					TargetCreature("target creature you control", YouControl()),
				)),
		},
	})
}

// mechanicalMobsterETB exiles the first clause's card, then connives the
// second clause's creature. The connive is not an "if you do": it runs after
// the exile has settled, whether or not a card was exiled (the exile may stop
// to ask, CR 614), so it is the exile's keyed tail (ExileTarget.ThenTail).
func mechanicalMobsterETB(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if t, ok := ctx.ClauseTarget(0); ok {
		return ExileTarget{Target: t.ID, ThenTail: ctx.ItemTail(mechanicalMobsterConnive)}.Apply(ctx)
	}
	return mechanicalMobsterConniveTarget(ctx)
}

// mechanicalMobsterConnive is the clause after the exile. An on-disk key.
var mechanicalMobsterConnive = CardTail("mechanical-mobster/connive", func(ctx *Context, _ game.TailOutcome) error {
	return mechanicalMobsterConniveTarget(ctx)
})

// mechanicalMobsterConniveTarget connives the creature the second clause chose,
// if it is still a legal target (CR 608.2b).
func mechanicalMobsterConniveTarget(ctx *Context) error {
	if t, ok := ctx.ClauseTarget(1); ok {
		return ConniveEach{Controller: ctx.Controller(), Creatures: []uuid.UUID{t.ID}}.Apply(ctx)
	}
	return nil
}
