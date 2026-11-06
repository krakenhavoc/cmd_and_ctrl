package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Atmospheric Greenhouse — Artifact — Spacecraft {4}{G}, 5/4 (Edge of
// Eternities, #4461):
//
//	"When this Spacecraft enters, put a +1/+1 counter on each creature
//	 you control.
//	 Station (Tap another creature you control: Put charge counters
//	 equal to its power on this Spacecraft. Station only as a sorcery.
//	 It's an artifact creature at 8+.)
//	 8+ | Flying, trample"
//
// The entry trigger is Basri's Solidarity's body (one counter on each
// creature the controller controls as it resolves). The Greenhouse is
// not a creature as it enters, so it gets no counter itself; a
// Greenhouse already stationed to 8 that is flickered still enters as a
// noncreature, because its charge counters do not come with it
// (CR 400.7). Station, the 8+ creature threshold and its flying and
// trample are the shared constructors (Galvanizing Sawship).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "52c6afc6-aff1-46a8-b1ea-b183facec1ae",
		Name:         "Atmospheric Greenhouse",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Atmospheric Greenhouse — put a +1/+1 counter on each creature you control",
				b11PutCounterOnEachCreatureYouControl),
		},
		Activated: []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(8, 5, 4),
			ThresholdKeywords(8, "flying", "trample"),
		),
	})
}
