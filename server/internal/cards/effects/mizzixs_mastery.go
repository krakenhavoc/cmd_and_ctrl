package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mizzix's Mastery — Sorcery {3}{R}:
//
//	"Exile target card that's an instant or sorcery from your
//	 graveyard. For each card exiled this way, copy it, and you may
//	 cast the copy without paying its mana cost. Exile Mizzix's
//	 Mastery.
//	 Overload {5}{R}{R}{R} (You may cast this spell for its overload
//	 cost. If you do, change "target" in its text to "each.")"
//
// The first card to use CopyAndCast (#3258, ADR 0066's 2026-10-13
// card-copy amendment): the exile goes through the CR 614 window and
// hands its continuation the cards that REALLY reached exile, so a card
// a replacement sent somewhere else is not copied. Each of those gets a
// copy (CR 707.12) in exile and a free-cast offer, all at once and with
// no cap, so an overloaded Mastery over a graveyard of five spells is
// five offers the caster takes in whatever order they like — the last
// copy cast resolves first. The copies are not cards: one not cast by
// the caster's next priority pass ceases to exist, and one that is cast
// ceases to exist as it resolves (CR 707.10a).
//
// Overload swaps "target" for "each" (CR 702.96a), which here widens the
// set from one chosen card to EVERY instant and sorcery card in the
// caster's own graveyard — the printed "from your graveyard" stays. The
// alternative cost clears the target clause, so an overloaded cast picks
// nothing.
//
// "Exile Mizzix's Mastery" is the last instruction: the spell leaves the
// stack to exile rather than to the graveyard, and it does so whether or
// not any card was exiled before it. A cast whose only target left in
// response fizzles instead (CR 608.2b) and the spell goes to the
// graveyard as for any fizzled spell.
//
// The free cast is a grant taken with an ordinary cast_spell after the
// resolution rather than an inline cast inside it (the posture every
// "you may cast it" a resolution offers takes, ADR 0066); it ignores
// timing, as CR 608.2g's cast-during-resolution does, so a copied
// sorcery is castable with Mastery's own resolution still behind it.
func init() {
	Register(Spec{
		OracleID:     "40362fe0-a1a9-4d76-8c35-eac474b91af5",
		Name:         "Mizzix's Mastery",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("target instant or sorcery card from your graveyard",
			Or(Instant(), Sorcery()), YouOwn()),
		AlternativeCosts: []game.AlternativeCost{Overload("{5}{R}{R}{R}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			if ctx.PaidAltCost("overload") {
				ids = graveyardIDs(ctx.Game, ctx.Controller(), func(c game.Card) bool {
					return c.IsInstant() || c.IsSorcery()
				})
			} else {
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						ids = append(ids, t.ID)
					}
				}
			}
			return ctx.Game.ExileCardsThenForEffect(ids, func(g *game.Game, exiled []uuid.UUID) error {
				ctx := NewContext(g, item)
				if err := (CopyAndCast{Cards: exiled, Free: true}).Apply(ctx); err != nil {
					return err
				}
				return ExileTarget{Target: item.SourceCardID}.Apply(ctx)
			})
		},
	})
}
