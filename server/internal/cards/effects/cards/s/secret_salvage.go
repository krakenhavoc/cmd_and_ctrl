package cardss

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Secret Salvage — Sorcery {3}{B}{B} (Aether Revolt, S82 older sets, #5536):
//
//	"Exile target nonland card from your graveyard. Search your library for
//	 any number of cards with the same name as that card, reveal them, put
//	 them into your hand, then shuffle."
//
// The name is read off the card before it is exiled (a face-down card has
// none, and a card with two names is found by either). The search happens
// whether or not the card reached exile (a replacement may send it elsewhere);
// a target that is no longer legal makes the spell do nothing (CR 608.2b).
// "Any number" is up to the number in the library, so the player may take
// none, some or all of them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f42fa525-dd17-448a-b0c3-29b90db2e30e",
		Name:         "Secret Salvage",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Targets:      effects.TargetCardInGraveyard("target nonland card from your graveyard", effects.YouOwn(), effects.Nonland()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(target)
			if !ok {
				return nil
			}
			names := effects.ExtractionNames(c)
			return effects.ExileTarget{
				Target:   target,
				ThenTail: effects.WithWord(ctx.ItemTail(secretSalvageExiled), strings.Join(names, "\n")),
			}.Apply(ctx)
		},
	})
}

// secretSalvageExiled is the search after the exile, a keyed tail (#4253).
// An on-disk key.
var secretSalvageExiled = effects.CardTail("secret-salvage/after-exile", func(ctx *effects.Context, o game.TailOutcome) error {
	names := strings.Split(o.Tail.Word, "\n")
	return effects.SearchLibrary{
		Player:    ctx.Controller(),
		Predicate: effects.SharesANameWith(names),
		Dest:      game.ZoneHand,
		Unbounded: true,
		Reveal:    true,
		Shuffle:   true,
		Reason:    "Secret Salvage — any number of cards with the same name",
	}.Apply(ctx)
})
