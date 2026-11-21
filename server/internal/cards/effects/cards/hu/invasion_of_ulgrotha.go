package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Invasion of Ulgrotha // Grandmother Ravi Sengir — March of the Machine (#5477).
//
// Front, Battle — Siege 5, {4}{B}: "When this Siege enters, it deals 3 damage
// to any other target and you gain 3 life." Back, Grandmother Ravi Sengir,
// Legendary Creature — Human Wizard 3/3: "Flying. Whenever a creature an
// opponent controls dies, put a +1/+1 counter on Grandmother Ravi Sengir and
// you gain 1 life."
//
// "Any other target": the Siege chosen as its own target is skipped at
// resolution (see tarkirAnyOtherTarget). The life is gained even when the
// damage cannot be dealt, as the ability still resolves for a legal target.
//
// No simplification.
const InvasionOfUlgrothaOracleID = "f1acb469-9d02-44e6-8c09-8b4157551b2c"

func init() {
	effects.Register(effects.Spec{
		OracleID:     InvasionOfUlgrothaOracleID,
		Name:         "Invasion of Ulgrotha",
		Completeness: effects.CompletenessFull,
		Battle:       &effects.BattleSpec{Defense: 5, Subtype: effects.BattleSubtypeSiege},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Invasion of Ulgrotha — 3 damage to any other target and you gain 3 life",
				func(g *game.Game, item *game.StackItem) error {
					if err := effects.S82MomB17DamageAnyOtherTarget(3)(g, item); err != nil {
						return err
					}
					return effects.Do(effects.GainLife{Amount: 3})(g, item)
				}), effects.TarkirAnyOtherTarget()),
			effects.DefeatedTrigger("Invasion of Ulgrotha — defeated: exile it, then cast Grandmother Ravi Sengir", effects.SiegeDefeated()),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        InvasionOfUlgrothaOracleID + "#1",
		Name:            "Grandmother Ravi Sengir",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B18OpponentsCreatureDied(ev, source, g)
			}, "Grandmother Ravi Sengir — a +1/+1 counter on Grandmother Ravi Sengir and you gain 1 life",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}, effects.GainLife{Amount: 1})),
		},
	})
}
