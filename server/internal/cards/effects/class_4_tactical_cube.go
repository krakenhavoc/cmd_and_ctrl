package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Class-4 Tactical Cube — Artifact — Spacecraft {4}, 4/4 (Star Trek
// Commander, #4974):
//
//	"When this Spacecraft enters, create two 1/1 colorless Borg artifact
//	 creature tokens.
//	 Station (Tap another creature you control: Put charge counters equal
//	 to its power on this Spacecraft. Station only as a sorcery. It's an
//	 artifact creature at 7+.)
//	 7+ | Flying
//	 Artifact creatures you control get +1/+1."
//
// Station is the ordinary keyword (station.go): at seven charge counters
// it is a 4/4 artifact creature with flying. The 7+ anthem is a layer-7c
// static under the same charge-counter gate (CR 721.2a), read live, so it
// switches on and off with the counters; it covers every artifact
// creature you control, the Cube itself and the Borg tokens included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9c14e21b-a733-4882-8e09-6dbfc85ac3aa",
		Name:         "Class-4 Tactical Cube",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Station()},
		Static: append(append(
			SpacecraftAt(7, 4, 4),
			ThresholdKeywords(7, "flying")),
			game.StaticAbility{
				Label:      "Artifact creatures you control get +1/+1.",
				Layer:      game.Layer7PT,
				SubLayer:   game.SubLayer7C_Modify,
				ActiveWhen: AtChargeCounters(7),
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.IsArtifact() && target.Controller == source.Controller
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
					c.Toughness++
				},
			}),
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Class-4 Tactical Cube — create two 1/1 colorless Borg artifact creature tokens",
				Do(CreateToken{Template: TokenCard("1/1 colorless Borg artifact"), N: 2})),
		},
	})
}
