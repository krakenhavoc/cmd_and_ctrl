package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deep Space 9 — Legendary Artifact — Spacecraft {9} (Star Trek, #4973):
//
//	"When Deep Space 9 enters and at the beginning of your upkeep, exile
//	 the top four cards of your library. Put all permanent cards exiled
//	 this way onto the battlefield.
//	 Station (Tap another creature you control: Put charge counters equal
//	 to its power on this Spacecraft. Station only as a sorcery.)
//	 9+ | Creatures you control get +3/+3."
//
// One printed trigger on two events (Grell Philosopher's
// thisEntersOrYourUpkeep), so one registered trigger. The exile is a
// MillToZone to exile; every permanent card that reached exile goes onto
// the battlefield under the trigger's controller in one simultaneous
// entry, and the instants and sorceries stay exiled
// (trekB27ExileTopFourPutPermanents).
//
// Deep Space 9 has no "It's an artifact creature at N+" clause, so it is
// never a creature. Its 9+ line is a station-gated anthem over the
// creatures its controller controls (trekB27CreaturesYouControlGet).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "14442f5b-0bad-4c58-b268-d775787571d8",
		Name:         "Deep Space 9",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventBeginUpkeep}, thisEntersOrYourUpkeep,
				"Deep Space 9 — exile the top four cards of your library and put the permanent cards onto the battlefield",
				trekB27ExileTopFourPutPermanents),
		},
		Activated: []ActivatedAbility{Station()},
		Static: []game.StaticAbility{
			trekB27CreaturesYouControlGet("9+ | Creatures you control get +3/+3.", 9, 3, 3),
		},
	})
}
