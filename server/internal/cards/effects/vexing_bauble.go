package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vexing Bauble — Artifact {1} (EDHREC rank 2966):
//
//	"Whenever a player casts a spell, if no mana was spent to cast
//	 it, counter that spell.
//	 {1}, {T}, Sacrifice this artifact: Draw a card."
//
// The free-spell hoser, and the card that most needed #761 to get
// "no mana was spent" right rather than merely implemented.
//
// The trigger watches EventCast from EVERY player — "a player", not
// "an opponent" — and reads the record on the spell's own stack item.
// For a spell the item's ID is the card's instance ID, so
// EventCast.CardID is the key; NoManaWasSpentToCast does the lookup.
//
// The whole card turns on the difference between two kinds of
// nothing, and the engine keeps them apart:
//
//   - A cascade or "without paying its mana cost" free cast, a {0}
//     alternative cost, and a copy of a spell (CR 707.10 — mana is
//     not an object) record a KNOWN nothing. The Bauble counters
//     them, which is what it is for.
//   - A cast the engine did not charge — an insufficient pool in
//     permissive mode, or Cast anyway — asks the caster what they
//     paid outside the pool, or whether they cast it without paying
//     its mana cost (#3855: the Bauble declares ReadsEveryCast). A
//     spell declared free with nothing from the pool is countered; a
//     declared payment is not. The declaration is in the game log.
//
// The intervening-if is checked when the trigger would go on the
// stack and again on resolution (CR 603.4) — the trigger's AppliesTo
// and the effect's own re-read — so a spell whose payment somehow
// changed in between does not get countered on a stale reading.
//
// A cast the engine already knows is free (cascade, a {0} alternative
// cost, a copy) asks nothing: nothing was waived.
func init() {
	Register(Spec{
		OracleID:       "4514777d-0589-4631-978b-ff244167c176",
		Name:           "Vexing Bauble",
		ReadsManaSpent: game.ReadsEveryCast,
		Completeness:   CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			Key:     "Vexing Bauble — counter a spell cast for no mana",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return source != nil && NoManaWasSpentToCast(g, ev.CardID)
			},
			// A fill-in Build (ADR 0041 P9): the item's own label
			// differs from the row's Key, exactly as before.
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Vexing Bauble — counter that spell")
			},
			Effect: counterTheSpellIfNoManaWasSpent,
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
