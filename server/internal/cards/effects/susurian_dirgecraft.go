package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Susurian Dirgecraft — Artifact — Spacecraft {4}{B}, 4/3 (Edge of
// Eternities, #4461):
//
//	"When this Spacecraft enters, each opponent sacrifices a nontoken
//	 creature of their choice.
//	 Station
//	 7+ | Flying"
//
// The enters trigger is Sheoldred's edict narrowed to nontoken creatures:
// each opponent chooses their own as it resolves, and an opponent with
// none is not asked. Station and the 7+ threshold are Wurmwall Sweeper's
// helpers: at seven or more charge counters it is a 4/3 artifact creature
// with flying (CR 702.184).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f1d826d1-d2da-46ce-b1db-81393c88d865",
		Name:         "Susurian Dirgecraft",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerEdict(WhenThisEnters("Susurian Dirgecraft — each opponent sacrifices a nontoken creature", Do(EachPlayerSacrifices{
				ExceptController: true,
				Match:            And(Creature(), Not(IsTokenPredicate())),
				Label:            "a nontoken creature",
			}))),
		},
		Activated: []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(7, 4, 3),
			ThresholdKeywords(7, "flying"),
		),
	})
}
