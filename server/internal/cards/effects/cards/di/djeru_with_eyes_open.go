package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Djeru, With Eyes Open — Legendary Creature — Human Warrior {3}{W}{W}, 4/3
// (Hour of Devastation, #5534):
//
//	"Vigilance
//	 When Djeru enters, you may search your library for a planeswalker card,
//	 reveal it, put it into your hand, then shuffle.
//	 If a source would deal damage to a planeswalker you control, prevent 1 of
//	 that damage."
//
// The search is Runeforge Champion's optional fetch over the library alone.
// The prevention is Heart-Shaped Herb's, aimed at your planeswalkers instead
// of you (CR 615.1a): every source, yours included, and all damage, combat or
// not; a one-point hit is prevented entirely. The planeswalker is judged by
// its type as the damage is dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5d520dd4-7986-41fa-ab94-9226da751668",
		Name:            "Djeru, With Eyes Open",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Purpose:         game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Djeru — you may search your library for a planeswalker card, reveal it, put it into your hand, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.IsPlaneswalker() },
						Dest:      game.ZoneHand,
						Limit:     1,
						Optional:  true,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Djeru, With Eyes Open — a planeswalker card",
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Replacements: []game.ReplacementEffect{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Prevention: true, // CR 615.1a
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageTarget == uuid.Nil {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.DamageTarget)
				return ok && c.IsPlaneswalker() && c.Controller == src.Controller
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount--
				if ev.DamageAmount <= 0 {
					ev.Cancel()
				}
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
			Label:      "Djeru — prevent 1 of that damage",
			RowLabel:   "Prevent 1 of the damage a source would deal to a planeswalker you control.",
		}},
	})
}
