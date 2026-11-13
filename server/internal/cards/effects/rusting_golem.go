package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rusting Golem — Artifact Creature — Golem {4}, */*:
//
//	"Fading 5 (This creature enters with five fade counters on it. At the
//	 beginning of your upkeep, remove a fade counter from it. If you
//	 can't, sacrifice it.)
//	 Rusting Golem's power and toughness are each equal to the number of
//	 fade counters on it."
//
// Fading is Fading/FadingUpkeep (fading.go, CR 702.32a). The size is a
// layer 7a characteristic-defining ability (CR 604.3, 613.4a) counted as
// the layers run. The counters are on it as it enters (they are an entry
// replacement), so it is never a 0/0 on the battlefield while it has
// any; it shrinks by one each upkeep and, once the fifth upkeep has
// taken the last counter, it is a 0/0 the state-based actions put into
// the graveyard before the next upkeep can sacrifice it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "688c232f-385b-4bd5-aed0-e55059111fb3",
		Name:         "Rusting Golem",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{Fading(5)},
		Triggered:    []game.TriggeredAbility{FadingUpkeep()},
		Static: []game.StaticAbility{{
			Label:     "Rusting Golem's power and toughness are each equal to the number of fade counters on it.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				n := source.Counters[game.CounterFade]
				c.Power, c.Toughness = n, n
			},
		}},
	})
}
