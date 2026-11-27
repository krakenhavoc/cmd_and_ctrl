package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tek — Artifact Creature — Dragon {5}, 2/2 (Invasion, S82 OLD #5616):
//
//	"This creature gets +0/+2 as long as you control a Plains, has flying as
//	 long as you control an Island, gets +2/+0 as long as you control a Swamp,
//	 has first strike as long as you control a Mountain, and has trample as long
//	 as you control a Forest."
//
// Five independent clauses of one sentence, so five static rows, each read
// live off the controller's lands (Wild Nacatl's SelfWhileYouControlA): the
// two pumps are layer 7c (CR 613.4c) and the three keywords layer 6 grants
// (CR 613.1f). A land that has a basic land type counts, whatever it is
// called; each row comes and goes on its own as lands arrive and leave. The
// land types are words of Tek's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f85aef9b-b6ad-4716-b48b-5e26b474a59d",
		Name:         "Tek",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "This creature gets +0/+2 as long as you control a Plains.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: effects.SelfWhileYouControlA(effects.Subtype("Plains")),
				Apply:     effects.PumpBy(0, 2),
			},
			effects.OldB652LabeledKeywordGrant("This creature has flying as long as you control an Island.", "flying",
				effects.SelfWhileYouControlA(effects.Subtype("Island"))),
			{
				Label:     "This creature gets +2/+0 as long as you control a Swamp.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: effects.SelfWhileYouControlA(effects.Subtype("Swamp")),
				Apply:     effects.PumpBy(2, 0),
			},
			effects.OldB652LabeledKeywordGrant("This creature has first strike as long as you control a Mountain.", "first strike",
				effects.SelfWhileYouControlA(effects.Subtype("Mountain"))),
			effects.OldB652LabeledKeywordGrant("This creature has trample as long as you control a Forest.", "trample",
				effects.SelfWhileYouControlA(effects.Subtype("Forest"))),
		},
	})
}
