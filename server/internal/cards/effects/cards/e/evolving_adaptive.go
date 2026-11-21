package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Evolving Adaptive — Creature — Phyrexian Warrior {G}, 0/0 (Phyrexia:
// All Will Be One):
//
//	"This creature enters with an oil counter on it.
//	 This creature gets +1/+1 for each oil counter on it.
//	 Whenever another creature you control enters, if that creature has
//	 greater power or toughness than this creature, put an oil counter on
//	 this creature."
//
// Evolve's comparison (CR 702.100a) with an oil counter instead of a
// +1/+1 counter, so it is written out rather than a keyword. It is read
// twice, as the creature enters and again as the ability resolves
// (CR 603.4), both sides with every counter and effect they have; the
// counter goes on only while this is still the permanent that triggered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e350c2d5-0ca2-4b9c-b01f-a3c31cbf8d10",
		Name:         "Evolving Adaptive",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(effects.S82OneOil, 1, "Evolving Adaptive: enters with an oil counter"),
		},
		Static: []game.StaticAbility{
			effects.S82OilPump(1, 1, "This creature gets +1/+1 for each oil counter on it."),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				entered, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				return ok && s82EvolvingAdaptiveGreater(entered, source)
			}, "Evolving Adaptive — put an oil counter on it", func(g *game.Game, item *game.StackItem) error {
				// CR 603.4: the comparison is made again on resolution.
				if item.Trigger == nil || !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
					return nil
				}
				entered, ok := g.LookupCardForEffect(item.Trigger.Event.CardID)
				self, sok := g.LookupCardForEffect(item.SourceCardID)
				if !ok || !sok || !s82EvolvingAdaptiveGreater(entered, &self) {
					return nil
				}
				return effects.AddCounter{Target: item.SourceCardID, Kind: effects.S82OneOil, N: 1}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

// s82EvolvingAdaptiveGreater is CR 702.100a's comparison: the entered
// creature's power or its toughness is greater than this creature's.
func s82EvolvingAdaptiveGreater(entered game.Card, self *game.Card) bool {
	return entered.IsCreature() && self.IsCreature() &&
		(entered.PowerForComparison() > self.PowerForComparison() ||
			entered.CurrentToughness() > self.CurrentToughness())
}
