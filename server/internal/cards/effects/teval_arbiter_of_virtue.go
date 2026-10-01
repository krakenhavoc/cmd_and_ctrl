package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Teval, Arbiter of Virtue — Legendary Creature — Spirit Dragon 6/6
// {2}{B}{G}{U}:
//
//	"Flying, lifelink
//	 Spells you cast have delve.
//	 Whenever you cast a spell, you lose life equal to its mana value."
//
// ADR 0100 sub-PR 2. "Spells you cast have delve" is
// Spec.SpellsYouCastHaveDelve, read by game.Game.DelveForLocked — the
// one door the pricer, the validator, the view and the bot enumerator
// already ask — so every spell its controller casts while Teval is on
// the battlefield may exile cards from their graveyard for its generic
// mana, and none may once Teval has gone or lost its abilities. The
// delve is a way to pay and nothing else (CR 702.66b): the mana value
// is untouched, and a second source of delve adds nothing (CR 702.66c).
//
// The trigger reads the spell's mana value as it triggers, with X at
// the value announced for it (CR 202.3e; the 2025-04-04 ruling), and
// carries it on the item, because the spell may have resolved or been
// countered by the time the trigger resolves. Delving does not lower it
// (CR 202.3; ADR 0100 §1): a Treasure Cruise that delved seven still
// costs its caster 8 life.
//
// No simplification.

const tevalArbiterOfVirtueLabel = "Teval, Arbiter of Virtue — lose life equal to that spell's mana value"

func init() {
	Register(Spec{
		OracleID:               "f6cf359a-91cb-4b3c-8837-be53e4ed9c93",
		Name:                   "Teval, Arbiter of Virtue",
		Completeness:           CompletenessFull,
		PrintedKeywords:        []string{"flying", "lifelink"},
		SpellsYouCastHaveDelve: true,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventCast},
			AppliesTo: YouCast(nil),
			Key:       tevalArbiterOfVirtueLabel,
			// A fill-in Build (ADR 0041 P9): the mana value is a fact of
			// the cast; the effect is the row's.
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				mv := 0
				if spell, ok := g.LookupCardForEffect(ev.CardID); ok {
					mv, _ = g.ManaValueForEffect(spell)
				}
				item := game.NewTriggeredItem(source, tevalArbiterOfVirtueLabel)
				item.Params.Amount = mv
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Params.Amount <= 0 {
					return nil
				}
				return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -item.Params.Amount)
			},
		}},
	})
}
