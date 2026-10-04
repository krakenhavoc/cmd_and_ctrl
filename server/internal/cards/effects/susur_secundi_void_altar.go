package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Susur Secundi, Void Altar — Land — Planet:
//
//	"This land enters tapped.
//	 {T}: Add {B}.
//	 Station (Tap another creature you control: Put charge counters
//	 equal to its power on this Planet. Station only as a sorcery.)
//	 12+ | {1}{B}, {T}, Pay 2 life, Sacrifice a creature: Draw cards
//	 equal to the sacrificed creature's power. Activate only as a
//	 sorcery."
//
// Adagia's shape (a station Planet land with a charge-gated ability)
// with Greater Good's last-known-power read for the draw.
func init() {
	Register(Spec{
		OracleID:     "50d6cadc-07e4-479e-90f4-e3a20f769bab",
		Name:         "Susur Secundi, Void Altar",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Activated: []ActivatedAbility{
			Station(),
			{
				Label:        "{1}{B}, {T}, Pay 2 life, Sacrifice a creature: Draw cards equal to the sacrificed creature's power. Activate only as a sorcery.",
				Cost:         Plus(ManaCost("{1}{B}"), TapCost(), PayLife(2), SacrificeACreature()),
				SorcerySpeed: true,
				ActiveWhen:   AtChargeCounters(12),
				Effect: func(g *game.Game, item *game.StackItem) error {
					fed, ok := b17PermanentSacrificedToPay(g, item)
					if !ok {
						return nil
					}
					return DrawCards{Player: item.Controller, N: departedCreaturePower(g, fed)}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
