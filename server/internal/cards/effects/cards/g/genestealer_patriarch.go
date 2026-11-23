package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Genestealer Patriarch — Creature — Tyranid {4}{U}, 4/4 (Warhammer 40,000
// Commander, S82 GAP batch 195, #5948):
//
//	"Genestealer's Kiss — Whenever this creature attacks, put an infection
//	 counter on target creature defending player controls.
//	 Children of the Cult — Whenever a creature with an infection counter on it
//	 dies, you create a token that's a copy of that creature, except it's a
//	 Tyranid in addition to its other types."
//
// The attack trigger's target is a creature controlled by the player the
// Patriarch is attacking (CR 802.2a), as it goes on the stack. The dies trigger
// is any creature's, anyone's, that last had an infection counter (CR 603.10a);
// the token is a copy of the creature as it last existed (CR 707.2) made for
// the Patriarch's controller, with Tyranid added to its types.
//
// No simplification.
func init() {
	attack := effects.WheneverThisAttacks("Genestealer Patriarch — put an infection counter on target creature defending player controls",
		effects.GapB195InfectTarget)
	attack.TargetsFrom = effects.TargetCreatureDefendingPlayerControls
	effects.Register(effects.Spec{
		OracleID:     "785ef850-4cd7-4e7d-88f2-7b6368c91a9c",
		Name:         "Genestealer Patriarch",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			attack,
			effects.On(game.EventLTB, effects.GapB195AnInfectedCreatureDied,
				"Genestealer Patriarch — create a token that's a copy of that creature, except it's a Tyranid in addition to its other types",
				effects.GapB195CopyTheInfectedCreature),
		},
	})
}
