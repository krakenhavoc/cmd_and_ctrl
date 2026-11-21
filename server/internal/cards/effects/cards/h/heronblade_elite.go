package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Heronblade Elite — Creature — Human Warrior {2}{G}, 1/1 (Midnight
// Hunt Commander, #5497):
//
//	"Vigilance
//	 Whenever another Human you control enters, put a +1/+1 counter on
//	 this creature.
//	 {T}: Add X mana of any one color, where X is this creature's power."
//
// Avatar Enthusiasts' trigger for Humans (effective subtypes, so a
// changeling counts) and Arbor Adherent's one-colour mana ability. X is
// the Elite's power through the layers and its counters, read as the
// mana ability resolves (CR 605.3b); a power of zero or less adds
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e3e255e0-ab04-4488-bdc1-bb1542f51e16",
		Name:            "Heronblade Elite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Human"))
			}, "Heronblade Elite — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:           effects.ManaAbilityCost{Tap: true},
			ProducedAmount: effects.ProducedOneColor(HeronbladeEliteX),
			Label:          "{T}: Add X mana of any one color, where X is this creature's power",
		}},
	})
}

// heronbladeEliteX is the ability source's power.
func HeronbladeEliteX(g *game.Game, _, source uuid.UUID) int {
	c, ok := g.LookupCardForEffect(source)
	if !ok {
		return 0
	}
	return max(c.PowerForComparison(), 0)
}
