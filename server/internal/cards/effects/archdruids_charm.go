package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Archdruid's Charm — Instant {G}{G}{G}:
//
//	"Choose one —
//	 • Search your library for a creature or land card and reveal it.
//	   Put it onto the battlefield tapped if it's a land card.
//	   Otherwise, put it into your hand. Then shuffle.
//	 • Put a +1/+1 counter on target creature you control. It deals
//	   damage equal to its power to target creature you don't control.
//	 • Exile target artifact or enchantment."
//
// Bullet one searches to the hand and, from the search's continuation,
// puts a found land onto the battlefield tapped: the search primitive
// has one destination per search, and a land that passes through the
// hand on the way is not observable (nothing happens between the two
// moves). The reveal is the printed one.
//
// Bullet two is a one-sided bite whose counter lands FIRST (the
// printed order, CR 608.2c), so the damage reads the grown power; the
// damage is the placement's continuation because a replacement (a
// Doubling Season beside a Hardened Scales) can pause the placement.
// Each target keeps its own clause, enforced at announce.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3c1ef404-e2c6-486d-a5a2-d5779c71d498",
		Name:         "Archdruid's Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Search your library for a creature or land card and reveal it. Put it onto the battlefield tapped if it's a land card. Otherwise, put it into your hand. Then shuffle."),
			Mode("Put a +1/+1 counter on target creature you control. It deals damage equal to its power to target creature you don't control.",
				Clauses(
					TargetCreature("target creature you control", YouControl()),
					TargetCreature("target creature you don't control", OpponentControls()),
				)),
			Mode("Exile target artifact or enchantment.",
				TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment()))),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			switch {
			case ctx.HasMode(0):
				controller := ctx.Controller()
				return SearchLibrary{
					Player:    controller,
					Predicate: func(c game.Card) bool { return c.IsCreature() || c.IsLand() },
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Archdruid's Charm — a creature or land card",
					// #4253 slice 4: "put it onto the battlefield tapped if it's
					// a land card" is a shared registered tail.
					ThenTail: game.KeyedTail{Ref: thenPutFoundLandsTapped, Player: controller, Source: ctx.Source()},
				}.Apply(ctx)
			case ctx.HasMode(1):
				occ := -1
				for o, m := range ctx.Modes() {
					if m == 1 {
						occ = o
					}
				}
				mine, ok := ModeClauseTarget(ctx, occ, 0)
				if !ok {
					return nil
				}
				theirs, ok := ModeClauseTarget(ctx, occ, 1)
				if !ok {
					return nil
				}
				then := ctx.ItemTail(archdruidsCharmBite)
				then.Carry = []uuid.UUID{mine.ID, theirs.ID}
				return ctx.Game.AddCounterByKeyedForEffect(uuid.Nil, mine.ID, game.CounterPlusOne, 1, then)
			case ctx.HasMode(2):
				if t, ok := ModeTarget(ctx, 0); ok {
					return ExileTarget{Target: t.ID}.Apply(ctx)
				}
			}
			return nil
		},
	})
}

// archdruidsCharmBite is "It deals damage equal to its power to target
// creature you don't control" once the counter has landed; Carry is the
// creature you control, then the other. An on-disk key: never renamed, never
// reused.
var archdruidsCharmBite = CardTail("archdruids-charm/bite", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) < 2 {
		return nil
	}
	mine, theirs := o.Tail.Carry[0], o.Tail.Carry[1]
	c, ok := ctx.Game.LookupCardForEffect(mine)
	if !ok {
		return nil
	}
	return DealDamage{Source: mine, Target: theirs, Amount: c.CurrentPower()}.Apply(ctx)
})
