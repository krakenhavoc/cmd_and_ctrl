package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Anje Falkenrath — Legendary Creature — Vampire {1}{B}{R}, 1/3 (Secret Lair
// Drop):
//
//	"Haste
//	 {T}, Discard a card: Draw a card.
//	 Whenever you discard a card, if it has madness, untap Anje Falkenrath."
//
// The loot is an ordinary activated ability whose cost discards (so a madness
// card discarded to it can be cast, as the madness trigger resolves). The
// untap is an intervening-"if" trigger (CR 603.4) on a card YOU discarded
// that has madness, read wherever the card now is (the madness replacement
// has usually exiled it): checked as it triggers and again as it resolves. A
// card has madness when the catalog declares a madness cost for it or it
// carries the keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4dab6a96-4376-4aea-983d-406167993214",
		Name:            "Anje Falkenrath",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDiscardCard, effects.YouDiscardedA(discardedCardHasMadness),
				"Anje Falkenrath — untap Anje Falkenrath",
				anjeUntap),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Discard a card: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.TapCost(), effects.DiscardACard()),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}

// discardedCardHasMadness is "it has madness": the discarded card's printed
// keywords, or the catalog's declared madness cost for its oracle ID.
func discardedCardHasMadness(c game.Card) bool {
	if game.HasKeyword(&c, "madness") {
		return true
	}
	spec, ok := effects.Lookup(c.OracleID)
	return ok && spec.Madness != ""
}

// anjeUntap untaps Anje, if the card that was discarded still has madness as
// the trigger resolves (CR 603.4) and Anje is still the permanent that
// triggered.
func anjeUntap(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	if c, ok := g.LookupCardForEffect(item.Trigger.Event.CardID); !ok || !discardedCardHasMadness(c) {
		return nil
	}
	return effects.UntapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
}
