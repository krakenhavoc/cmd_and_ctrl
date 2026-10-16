package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cloudspire Captain — Creature — Human Pilot {2}{W}, 2/3:
//
//	"Mounts and Vehicles you control get +1/+1.
//	 This creature saddles Mounts and crews Vehicles as though its power
//	 were 2 greater."
//
// The anthem is a layer 7c modifier on every Mount and Vehicle its
// controller controls, creature or not: a Vehicle that has not been
// crewed carries the +1/+1 too, so it is already a 5/5 when the crew
// ability animates a 4/4. The Captain is neither a Mount nor a Vehicle
// and does not buff itself. TribeFilter is not used because it matches
// creatures only, and a Vehicle is a creature only while crewed.
//
// The second line is CrewsAsThoughPowerGreater (ADR 0186 Decision 4):
// the Captain counts as power 4 for crew and saddle, and its real power
// stays 2.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c86bac69-96d6-4b9f-87e4-2d7faecaa112",
		Name:         "Cloudspire Captain",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:    "Mounts and Vehicles you control get +1/+1.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					// "Mounts" is a word of the source's text (CR 612.2,
					// #3984); Vehicle is an artifact type, which no text
					// change reaches.
					return target.Controller == source.Controller &&
						(target.HasSubtype(source.TextSubtype("Mount")) || target.HasSubtype("Vehicle"))
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
					c.Toughness++
				},
			},
			CrewsAsThoughPowerGreater(2, game.CostPowerCrew|game.CostPowerSaddle),
		},
	})
}
