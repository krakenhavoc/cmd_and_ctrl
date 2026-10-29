package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crossbones, Malicious Mercenary — Legendary Creature — Human Mercenary
// Villain {3}{B}, 3/3 (Marvel Super Heroes, #4455):
//
//	"Deathtouch
//	 Whenever another Villain you control enters, put a +1/+1 counter on
//	 Crossbones. He deals 2 damage to each opponent. This ability
//	 triggers only once each turn."
//
// "Only once each turn" is the per-turn trigger record (b11TriggeredThisTurn):
// once the ability has been put on the stack this turn no later Villain
// triggers it, even if the first was countered. The counter goes on
// Crossbones and the 2 damage comes from him, to every opponent at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ac950d10-e373-42d3-907b-ee56e45a2084",
		Name:            "Crossbones, Malicious Mercenary",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return mshB2AnotherVillainEntered(ev, source, g) && !b11TriggeredThisTurn(g, source.InstanceID, crossbonesLabel)
			}, crossbonesLabel, func(g *game.Game, item *game.StackItem) error {
				if err := plusOneCountersOnThis(1)(g, item); err != nil {
					return err
				}
				return damageToEachOpponent(g, item, 2)
			}),
		},
	})
}

const crossbonesLabel = "Crossbones, Malicious Mercenary — a +1/+1 counter and 2 damage to each opponent"
