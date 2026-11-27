package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sylvan Primordial — Creature — Avatar {5}{G}{G}, 6/8 (Gatecrash, #5565):
//
//	"Reach
//	 When this creature enters, for each opponent, destroy target noncreature
//	 permanent that player controls. Search your library for up to X Forest
//	 cards, where X is the number of permanents destroyed this way, put those
//	 cards onto the battlefield tapped, then shuffle."
//
// One required target clause per opponent, each bound to that opponent
// (PerOpponentRequiredClauses, Blatant Thievery's shape); an opponent with
// no noncreature permanent has no target and is skipped. The permanents are
// destroyed as one batch (DestroyPermanentsKeyedForEffect) whose tail is
// told the ones actually destroyed — an indestructible permanent is not —
// and X is that count. Forest cards are any card with the Forest land type
// (Binding the Old Gods), not only basics. The search is "up to X" with a
// quality, and the library is shuffled even when X is 0 (the search still
// happens).
//
// No simplification.
func init() {
	trigger := effects.WhenThisEnters("Sylvan Primordial — for each opponent, destroy target noncreature permanent that player controls; search for up to X Forest cards",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			if len(ids) == 0 {
				return sylvanPrimordialSearch(ctx, 0)
			}
			return g.DestroyPermanentsKeyedForEffect(ids, ctx.ItemTail(sylvanPrimordialDestroyed))
		})
	trigger.Targets = effects.PerOpponentRequiredClauses("noncreature permanent", effects.Not(effects.Creature()))
	effects.Register(effects.Spec{
		OracleID:        "52b88034-bf0f-4d29-b776-40ce421f1107",
		Name:            "Sylvan Primordial",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}

// sylvanPrimordialDestroyed is the search, sized by what was destroyed. An
// on-disk key: never renamed, never reused.
var sylvanPrimordialDestroyed = effects.CardTail("destroy-then/sylvan-primordial-forests", func(ctx *effects.Context, o game.TailOutcome) error {
	return sylvanPrimordialSearch(ctx, len(o.IDs))
})

func sylvanPrimordialSearch(ctx *effects.Context, x int) error {
	if x <= 0 {
		return ctx.Game.ShuffleLibraryForEffect(ctx.Controller())
	}
	return effects.SearchLibrary{
		Player:        ctx.Controller(),
		Predicate:     effects.IsLandWithSubtype("Forest"),
		Dest:          game.ZoneBattlefield,
		Limit:         x,
		Shuffle:       true,
		TappedOnEntry: true,
		Reason:        "Sylvan Primordial — search for up to X Forest cards",
	}.Apply(ctx)
}
