package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Merieke Ri Berit — Legendary Creature — Human {W}{U}{B}, 1/1 (Ice Age, S82 OLD #5636):
//
//	"Merieke Ri Berit doesn't untap during your untap step.
//	 {T}: Gain control of target creature for as long as you control Merieke Ri
//	 Berit. When Merieke Ri Berit leaves the battlefield or becomes untapped,
//	 destroy that creature. It can't be regenerated."
//
// The control change lasts while Merieke is under your control; the creature is
// remembered on her as her linked permanent (CR 607.2e), so the one trigger,
// fired by either event, destroys it while it is still that object (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "288fdfbb-1236-4732-afd7-e7eaba860007",
		Name:                  "Merieke Ri Berit",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringYourUntapStep()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Gain control of target creature for as long as you control Merieke Ri Berit.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.MeriekeStealB831,
		}},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventLTB, game.EventUntapCard}, effects.MeriekeLeftOrUntappedB831,
				"Merieke Ri Berit — destroy the creature she took; it can't be regenerated", effects.MeriekeDestroyB831),
		},
	})
}
