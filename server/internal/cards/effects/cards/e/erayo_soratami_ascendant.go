package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Erayo, Soratami Ascendant // Erayo's Essence — {1}{U} Legendary Creature
// — Moonfolk Monk 1/1 // Legendary Enchantment, a flip card (#3216, ADR
// 0151):
//
//	Top:     "Flying
//	          Whenever the fourth spell of a turn is cast, flip Erayo."
//	Flipped: "Whenever an opponent casts their first spell each turn,
//	          counter that spell."
//
// The count is the table's: the spells every player has cast this turn, the
// one being cast included (the tally is bumped before the cast event, so a
// spell that is the fourth is seen as it is cast). Only the fourth triggers it,
// never a later one. The flip is not conditional on anything after it triggers.
// The Essence is the Jace, Unraveler of Secrets emblem's trigger: an opponent's
// first spell of a turn is countered as the trigger resolves; the Essence is a
// blue permanent with no mana cost of its own (CR 710.1c).
//
// No simplification.
const ErayoOracleID = "d83bb32d-f409-421a-9d93-bd7ea3240b47"

// ErayoSpellsCastThisTurn is the number of spells cast by all players this turn.
func ErayoSpellsCastThisTurn(g *game.Game) int {
	n := 0
	for _, s := range g.Seats {
		if s != nil {
			n += g.CastTallyFor(s.ID).Total
		}
	}
	return n
}

func init() {
	effects.Register(effects.Spec{
		OracleID:        ErayoOracleID,
		Name:            "Erayo, Soratami Ascendant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventCast && ErayoSpellsCastThisTurn(g) == 4
			}, "Erayo, Soratami Ascendant — the fourth spell of a turn was cast: flip it", func(g *game.Game, item *game.StackItem) error {
				if effects.SourceIsNewObject(g, item) {
					return nil
				}
				return effects.FlipThis{}.Apply(effects.NewContext(g, item))
			}),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     ErayoOracleID + "#1",
		Name:         "Erayo's Essence",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AllOf(effects.AnOpponentCast(nil), func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return g.CastTallyFor(ev.Actor).Total == 1
			}), "Erayo's Essence — counter that spell", func(g *game.Game, item *game.StackItem) error {
				return effects.CounterTarget{StackID: item.Trigger.Event.CardID}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
