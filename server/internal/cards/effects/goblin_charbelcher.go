package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goblin Charbelcher — Artifact {4} (Mirrodin, #5606):
//
//	"{3}, {T}: Reveal cards from the top of your library until you reveal
//	 a land card. This artifact deals damage equal to the number of
//	 nonland cards revealed this way to any target. If the revealed land
//	 card was a Mountain, this artifact deals double that damage instead.
//	 Put the revealed cards on the bottom of your library in any order."
//
// The target is chosen as the ability is activated and checked again as
// it resolves (CR 608.2b); with no legal target left the ability does
// nothing, so nothing is revealed. The run is the reveal-until family's
// (revealUntil): the whole run is shown to the table, and a library with
// no land reveals itself entirely and counts every card. The damage is
// the nonland cards revealed, doubled when the land the run stopped on
// is a Mountain ("Mountain" read through the ability's text, CR 612.2),
// dealt by the Charbelcher. Then every revealed card, the land included,
// goes to the bottom in the order its controller chooses (the ordered
// bottom of ADR 0088).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c9b1557b-70b1-44f2-9fc2-7aa0b035b187",
		Name:         "Goblin Charbelcher",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}: Reveal cards from the top of your library until you reveal a land card. This artifact deals damage equal to the number of nonland cards revealed this way to any target. If the revealed land card was a Mountain, this artifact deals double that damage instead. Put the revealed cards on the bottom of your library in any order.",
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			Targets: TargetAny(),
			Effect:  goblinCharbelcherFire,
		}},
	})
}

// goblinCharbelcherFire is the ability's resolution: reveal until a land,
// deal the damage, then queue the ordered bottom.
func goblinCharbelcherFire(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	run, land := revealUntil(ctx, ctx.Controller(), func(c game.Card) bool { return c.IsLand() },
		"Goblin Charbelcher — reveal cards until a land card")
	damage := 0
	for _, id := range run {
		if c, ok := g.LookupCardForEffect(id); ok && id != land && !c.IsLand() {
			damage++
		}
	}
	if land != uuid.Nil {
		if c, ok := g.LookupCardForEffect(land); ok && c.HasSubtype(ctx.TextSubtype("Mountain")) {
			damage *= 2
		}
	}
	if err := (DealDamage{Source: item.SourceCardID, Target: targets[0].ID, Amount: damage}).Apply(ctx); err != nil {
		return err
	}
	return PutInLibraryInAnyOrder{
		Cards:     run,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceBottom,
		Label:     "Goblin Charbelcher — put the revealed cards on the bottom of your library in any order",
	}.Apply(ctx)
}
