package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dire-Strain Rampage — Sorcery {1}{R}{G} (Innistrad: Midnight Hunt,
// #5497):
//
//	"Destroy target artifact, enchantment, or land. If a land was
//	 destroyed this way, its controller may search their library for up
//	 to two basic land cards, put them onto the battlefield tapped, then
//	 shuffle. Otherwise, its controller may search their library for a
//	 basic land card, put it onto the battlefield tapped, then shuffle.
//	 Flashback {3}{R}{G}"
//
// The controller is read BEFORE the destroy (the card is in a graveyard
// afterwards). "A land was destroyed this way" is decided by what
// happened: a land that is indestructible, or that a regeneration shield
// saved, is still on the battlefield, so the "otherwise" branch applies
// and its controller may fetch only one. A target that is no longer
// legal on resolution fizzles the whole spell (CR 608.2b), so nobody
// searches. The search is a real "may": it is Optional, so the
// controller can decline both the lands and the shuffle.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "55281479-9d30-4f3e-afd0-66b0df5de68c",
		Name:         "Dire-Strain Rampage",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact, enchantment, or land", Or(Artifact(), Enchantment(), Land())),
		// #2679: what the target's controller is given back.
		Purpose:          ForTargets(RemovalReturning(0, game.TargetReturn{Lands: 1})),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{R}{G}")},
		OnResolve:        direStrainRampageResolve,
	})
}

func direStrainRampageResolve(item *game.StackItem, ctx *Context) error {
	if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
		return nil
	}
	target := item.Targets[0].ID
	c, ok := ctx.Game.LookupCardForEffect(target)
	if !ok {
		return nil
	}
	wasLand := 0
	if c.IsLand() {
		wasLand = 1
	}
	// The search rides the destruction's continuation (a commander pauses
	// on CR 903.9's window first), and "a land was destroyed this way" is
	// read off the list of what was really destroyed.
	return ctx.Game.DestroyPermanentsKeyedForEffect([]uuid.UUID{target},
		withN(withCarry(ctx.ItemTail(direStrainRampageSearch), c.Controller), wasLand))
}

// direStrainRampageSearch is the clause after the destroy: its victim may
// fetch up to two basics if a land was destroyed, otherwise one. An
// on-disk key: never renamed, never reused.
var direStrainRampageSearch = CardTail("dire-strain-rampage/search", func(ctx *Context, o game.TailOutcome) error {
	n := 1
	if o.Tail.N == 1 && len(o.IDs) > 0 {
		n = 2
	}
	return SearchLibrary{
		Player:        tailCarry0(o),
		Predicate:     IsBasicLand,
		Dest:          game.ZoneBattlefield,
		Limit:         n,
		Reveal:        true,
		Shuffle:       true,
		TappedOnEntry: true,
		Optional:      true,
		Reason:        "Dire-Strain Rampage — you may search for basic land cards",
	}.Apply(ctx)
})
