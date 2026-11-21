package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Urza's Saga — Enchantment Land — Urza's Saga:
//
//	"(As this Saga enters and after your draw step, add a lore
//	 counter. Sacrifice after III.)
//	 I — This Saga gains "{T}: Add {C}."
//	 II — This Saga gains "{2}, {T}: Create a 0/0 colorless Construct
//	      artifact creature token with 'This token gets +1/+1 for each
//	      artifact you control.'"
//	 III — Search your library for an artifact card with mana cost {0}
//	       or {1}, put it onto the battlefield, then shuffle."
//
// The first two chapters are duration grants with NO stated duration
// (ADR 0093 PR 4, #1584): CR 611.2a makes such an effect last for the
// rest of the game, and it is pinned to the Saga, so it ends when the
// Saga leaves the battlefield and a Saga that comes back is a new
// object without either ability (CR 400.7). Both are ScopedEffect
// records — data — so the table stays a restore point for the three
// turns the Saga lives.
//
// The abilities are the SAGA's own once granted: its controller taps
// it for {C} (the auto-tapper plans it like any land's mana), and a
// later "loses all abilities" takes them (CR 613.6).
//
// Chapter III's "mana cost {0} or {1}" is the printed MANA COST, not
// the mana value: a card with no mana cost, or an {X} cost, is not
// found (the card's 2021-06-18 rulings).
//
// The Construct is Urza, Lord High Artificer's token template: the
// same printed token.
//
// No simplification.
const (
	urzasSagaColorless = "urzas-saga/colorless"
	urzasSagaConstruct = "urzas-saga/construct"
)

func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c6a0c30-b547-4eff-8ff4-0ca25803c076",
		Name:         "Urza's Saga",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			effects.TapForManaGrant(urzasSagaColorless, "{C}", "Add {C}", "{T}: Add {C}."),
			{
				Key: urzasSagaConstruct,
				Activated: []effects.ActivatedAbility{{
					Label:   "{2}, {T}: Create a 0/0 Construct with +1/+1 for each artifact you control",
					Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerMakesBlocker},
					Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
					Effect: func(g *game.Game, item *game.StackItem) error {
						return effects.CreateToken{Template: effects.UrzaConstructToken(), N: 1}.Apply(effects.NewContext(g, item))
					},
				}},
				Text: "{2}, {T}: Create a 0/0 colorless Construct artifact creature token with \"This token gets +1/+1 for each artifact you control.\"",
			},
		},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, "Urza's Saga — this Saga gains \"{T}: Add {C}.\"", urzasSagaGains(urzasSagaColorless)),
			effects.ChapterTrigger(2, "Urza's Saga — this Saga gains the Construct ability", urzasSagaGains(urzasSagaConstruct)),
			effects.ChapterTrigger(3, "Urza's Saga — search for an artifact with mana cost {0} or {1}", urzasSagaSearch),
		},
	})
}

// urzasSagaGains is chapters I and II: the Saga gives ITSELF a bundle
// for the rest of the game, pinned to this object.
func urzasSagaGains(key string) effects.Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if effects.SourceIsNewObject(g, item) {
			return nil // "this Saga" left and came back: a new object
		}
		saga := item.SourceCardID
		return effects.GrantAbilitiesFor{
			Target:   saga,
			Keys:     []string{key},
			Duration: g.PinnedTo(game.IndefiniteDuration(), saga),
			Label:    "Urza's Saga — gains " + key,
		}.Apply(effects.NewContext(g, item))
	}
}

// urzasSagaSearch is chapter III.
func urzasSagaSearch(g *game.Game, item *game.StackItem) error {
	return effects.SearchLibrary{
		Player:    item.Controller,
		Predicate: urzasSagaFindable,
		Dest:      game.ZoneBattlefield,
		Limit:     1,
		Shuffle:   true,
		Reason:    "Urza's Saga — an artifact card with mana cost {0} or {1}",
		Source:    item.SourceCardID,
	}.Apply(effects.NewContext(g, item))
}

// urzasSagaFindable is "an artifact card with mana cost {0} or {1}".
func urzasSagaFindable(c game.Card) bool {
	return c.IsArtifact() && (c.ManaCost == "{0}" || c.ManaCost == "{1}")
}
