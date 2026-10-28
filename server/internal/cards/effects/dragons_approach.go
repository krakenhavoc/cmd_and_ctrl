package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dragon's Approach — Sorcery {2}{R}:
//
//	"Dragon's Approach deals 3 damage to each opponent.
//	 You may exile this spell and four cards named Dragon's Approach from
//	 your graveyard. If you do, search your library for a Dragon creature
//	 card, put it onto the battlefield, then shuffle.
//	 A deck can have any number of cards named Dragon's Approach."
//
// The deck rule is CR 113.6n's, read off the oracle text by internal/deck
// for every card, catalog or not (Shadowborn Apostle), so the card file
// writes nothing for it.
//
// The damage is dealt to each opponent in turn. The "may" is asked only
// when four other copies are in the caster's graveyard, since without them
// the exile cannot happen. The answer exiles four of them (they are
// interchangeable, so the engine picks), then this spell, which has been
// routed to the graveyard by then (so it is exiled from there), and the
// search runs once the spell is really in exile ("if you do"). The search
// has a stated quality, so the searcher may find nothing (CR 701.23b), and
// the library is shuffled either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "47e191f7-6314-4875-b5ee-57e5daf089c4",
		Name:         "Dragon's Approach",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				// One instance of damage (CR 615.8).
				return ctx.Game.DamageInstanceForEffect(func() error {
					for _, opp := range ctx.Opponents() {
						if err := (DealDamage{Source: ctx.Source(), Target: opp, Amount: 3}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				})
			}),
			Instruction(func(ctx *Context) error {
				if len(dragonsApproachOthers(ctx)) < 4 {
					return nil
				}
				return MayChoice{
					Question: "Dragon's Approach — exile it and four cards named Dragon's Approach from your graveyard to search for a Dragon?",
					YesLabel: "Exile them",
					NoLabel:  "Don't",
					Then:     dragonsApproachAnswer,
				}.Apply(ctx)
			}),
		),
	})
}

// dragonsApproachOthers are the cards named Dragon's Approach in the
// controller's graveyard other than the spell itself.
func dragonsApproachOthers(ctx *Context) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range graveyardCardsOfPlayer(ctx.Game, ctx.Controller()) {
		if id == ctx.Source() {
			continue
		}
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && game.HasName(c, "Dragon's Approach") {
			out = append(out, id)
		}
	}
	return out
}

// dragonsApproachAnswer is the "yes": four other copies, then this spell,
// leave for exile, and the search follows the spell. An on-disk key.
var dragonsApproachAnswer = MayYes("confirm/dragons-approach-exile", func(ctx *Context, _ game.ConfirmAnswer) error {
	others := dragonsApproachOthers(ctx)
	if len(others) < 4 {
		return nil
	}
	for _, id := range others[:4] {
		if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return ExileTarget{Target: ctx.Source(), ThenTail: ctx.Tail(dragonsApproachSearch)}.Apply(ctx)
})

// dragonsApproachSearch is "search your library for a Dragon creature
// card, put it onto the battlefield, then shuffle", told the spell when it
// reached exile. An on-disk key.
var dragonsApproachSearch = CardTail("then/dragons-approach-search", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return SearchLibrary{
		Player: ctx.Controller(),
		// SearchLibrary lends the source's text (CR 612.2, #3984).
		Predicate: func(c game.Card) bool { return c.IsCreature() && c.HasSubtypeWord("Dragon") },
		Dest:      game.ZoneBattlefield,
		Limit:     1,
		Shuffle:   true,
		Reason:    "Dragon's Approach — search for a Dragon creature card",
		Source:    ctx.Source(),
	}.Apply(ctx)
})
