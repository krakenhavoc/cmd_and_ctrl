package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Can't Stay Away — Sorcery {W}{B} (Innistrad: Midnight Hunt, #5497):
//
//	"Return target creature card with mana value 3 or less from your
//	 graveyard to the battlefield. It gains "If this creature would die,
//	 exile it instead."
//	 Flashback {3}{W}{B}"
//
// The creature comes back under your control (it is your graveyard's
// card, so you own it). The granted replacement effect has no duration:
// it is a record pinned to that object, lasting while it stays on the
// battlefield (CR 400.7), past the turn, and it exiles the creature
// instead of letting it die (CR 614.1a, 700.4, so no dies trigger); a
// creature that leaves any other way (exile, bounce) is not affected. A
// target no longer in the graveyard fizzles the spell (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "69ae219e-bf97-4626-9b16-7901f51f0343",
		Name:         "Can't Stay Away",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("target creature card with mana value 3 or less in your graveyard",
			YouOwn(), Creature(), ManaValueLE(3)),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{W}{B}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			id := item.Targets[0].ID
			if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
				return nil
			}
			return ctx.Game.ReturnFromGraveyardWithCountersKeyedForEffect(id, item.Controller, false, nil,
				ctx.ItemTail(s82Mid10ReturnedExileIfItWouldDie))
		},
	})
}
