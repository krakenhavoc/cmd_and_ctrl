package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Diabolic Intent — Sorcery for {1}{B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature."
//	"Search your library for a card, put that card into your hand,
//	 then shuffle."
//
// Demonic Tutor for two mana and a body. In Commander the creature
// is usually one that wanted to die anyway — a token, a Sakura-Tribe
// Elder, something already earmarked for the graveyard — so the
// "drawback" is frequently upside, which is why this is a staple and
// Demonic Tutor's third-best imitator is not.
//
// The batch-02 triage (#295) filed it under "cost modification".
// As with Harrow, that is a misfiling: this is CR 601.2f, an
// ADDITIONAL cost, and Spec.AdditionalCost's sacrifice clause
// (SacrificeCost) has existed since the S21 pass. Nothing modifies
// the mana cost.
//
// Paying with the spell already on the stack is load-bearing here in
// a way it is not on most sacrifice-cost cards: an aristocrats board
// (Blood Artist, Zulaport Cutthroat, Midnight Reaper) triggers ABOVE
// Diabolic Intent and drains before the tutor resolves. That is the
// printed sequence, and it is why the cost is not paid in OnResolve.
//
// "For a card" — no predicate, exactly as printed. Any card in the
// library is a legal find, and the searcher picks through the S22
// search chooser.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:       "038519b9-bca8-4b27-b5ac-2409595469d0",
		Name:           "Diabolic Intent",
		AdditionalCost: SacrificeCost("a creature", Creature()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:  item.Controller,
				Dest:    game.ZoneHand,
				Limit:   1,
				Shuffle: true,
				Reason:  "Diabolic Intent — search your library for a card",
			}.Apply(ctx)
		},
	})
}
