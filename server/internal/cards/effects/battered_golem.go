package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Battered Golem — Artifact Creature — Golem {3}, 3/2 (Fifth Dawn,
// #5604):
//
//	"This creature doesn't untap during your untap step.
//	 Whenever an artifact enters, you may untap this creature."
//
// Brass Man's untap-step restriction and Grinding Station's optional
// untap, which watches every artifact entering under any player's
// control, the Golem's own entry included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "066dfefd-eef5-4c81-8e8a-310bac01674f",
		Name:                  "Battered Golem",
		Completeness:          CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{doesntUntapDuringYourUntapStep()},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventETB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b20ArtifactEntered(ev, g)
			}, "Battered Golem — untap", b31UntapSelf), "Battered Golem — untap it?"),
		},
	})
}
