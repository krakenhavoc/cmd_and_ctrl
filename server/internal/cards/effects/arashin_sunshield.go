package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arashin Sunshield — Creature — Human Warrior {3}{W}, 3/4:
//
//	"When this creature enters, exile up to two target cards from a
//	 single graveyard.
//	 {W}, {T}: Tap target creature."
//
// #1807, ADR 0106 §5: Griffnaut Tracker's entry trigger, plus a tapper
// that shares Staff of Domination's body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1175d482-d8a2-467f-bc50-2f4b241966bb",
		Name:         "Arashin Sunshield",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersExileFromASingleGraveyard("Arashin Sunshield", 2, ExileTargetCards),
		},
		Activated: []ActivatedAbility{{
			Label:   "{W}, {T}: Tap target creature.",
			Cost:    Plus(ManaCost("{W}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
