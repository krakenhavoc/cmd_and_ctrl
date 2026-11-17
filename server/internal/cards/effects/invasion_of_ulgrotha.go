package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
const invasionOfUlgrothaOracleID = "f1acb469-9d02-44e6-8c09-8b4157551b2c"

func init() {
	Register(Spec{
		OracleID:     invasionOfUlgrothaOracleID,
		Name:         "Invasion of Ulgrotha",
		Completeness: CompletenessFull,
		Battle:       &BattleSpec{Defense: 5, Subtype: BattleSubtypeSiege},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Invasion of Ulgrotha — 3 damage to any other target and you gain 3 life",
				func(g *game.Game, item *game.StackItem) error {
					if err := s82MomB17DamageAnyOtherTarget(3)(g, item); err != nil {
						return err
					}
					return Do(GainLife{Amount: 3})(g, item)
				}), tarkirAnyOtherTarget()),
			DefeatedTrigger("Invasion of Ulgrotha — defeated: exile it, then cast Grandmother Ravi Sengir", SiegeDefeated()),
		},
	})
	Register(Spec{
		OracleID:        invasionOfUlgrothaOracleID + "#1",
		Name:            "Grandmother Ravi Sengir",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b18OpponentsCreatureDied(ev, source, g)
			}, "Grandmother Ravi Sengir — a +1/+1 counter on Grandmother Ravi Sengir and you gain 1 life",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}, GainLife{Amount: 1})),
		},
	})
}
