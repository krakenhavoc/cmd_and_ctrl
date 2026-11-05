package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kirk, Man of Action — {2}{G}{W}{U} Legendary Creature — Human Officer
// 3/3 (Star Trek Commander, #4974):
//
//	"Other Officers, Artificers, Doctors, Pilots, Scientists, and
//	 Spacecraft creatures you control get +2/+2."
//
// A lord over six types: TribalAnthem with the six named, "other" and
// "you control". A creature that is two of them gets +2/+2 once, and a
// stationed Spacecraft that has become a creature gets it while it is one.
// The types are read through the layered view, so a changeling counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "254177f2-5a03-4a29-8b98-248f6a41a351",
		Name:         "Kirk, Man of Action",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{
				Tribes:    []string{"Officer", "Artificer", "Doctor", "Pilot", "Scientist", "Spacecraft"},
				Others:    true,
				YoursOnly: true,
			}, 2, 2),
		},
	})
}
