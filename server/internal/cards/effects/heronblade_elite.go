package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:        "e3e255e0-ab04-4488-bdc1-bb1542f51e16",
		Name:            "Heronblade Elite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Human"))
			}, "Heronblade Elite — put a +1/+1 counter on it", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		ManaAbilities: []ManaAbility{{
			Cost:           ManaAbilityCost{Tap: true},
			ProducedAmount: ProducedOneColor(heronbladeEliteX),
			Label:          "{T}: Add X mana of any one color, where X is this creature's power",
		}},
	})
}

// heronbladeEliteX is the ability source's power.
func heronbladeEliteX(g *game.Game, _, source uuid.UUID) int {
	c, ok := g.LookupCardForEffect(source)
	if !ok {
		return 0
	}
	return max(c.PowerForComparison(), 0)
}
