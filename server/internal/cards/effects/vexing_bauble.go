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
//   - A cast the engine did not charge — permissive mode, the human
//     default, or a strict-mode override — records "unknown"
//     (PaidCost.OnPaper). The Bauble does NOT fire. It is the weaker
//     answer, and it is the only sane one: a Bauble that countered
//     every spell at a permissive table would end the game.
//
// The intervening-if is checked when the trigger would go on the
// stack and again on resolution (CR 603.4) — the trigger's AppliesTo
// and the effect's own re-read — so a spell whose payment somehow
// changed in between does not get countered on a stale reading.
//
// One caveat for the player, because the difference is invisible from
// the table: with strict mana off, the Bauble never counters
// anything.
func init() {
	Register(Spec{
		OracleID:     "4514777d-0589-4631-978b-ff244167c176",
		Name:         "Vexing Bauble",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"With strict mana off, the game doesn't track what you spent, so the Bauble never counters a spell — turn strict mana on for it to work."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			Key:     "Vexing Bauble — counter a spell cast for no mana",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return source != nil && NoManaWasSpentToCast(g, ev.CardID)
			},
			// A fill-in Build (ADR 0041 P9): the item's own label
			// differs from the row's Key, exactly as before.
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Vexing Bauble — counter that spell", nil)
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				spell := item.Trigger.Event.CardID
				// CR 603.4: the intervening-if is checked
				// again on resolution. A spell that is no
				// longer on the stack answers "known
				// nothing" — but CounterTarget on a missing
				// item is already a no-op, so the re-read is
				// the honest check and not the guard.
				if !NoManaWasSpentToCast(g, spell) {
					return nil
				}
				return CounterTarget{StackID: spell}.Apply(NewContext(g, item))
			},
		}},
		Activated: []ActivatedAbility{{
			Label: "{1}, {T}, Sacrifice this artifact: Draw a card.",
			Cost:  Plus(ManaCost("{1}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
