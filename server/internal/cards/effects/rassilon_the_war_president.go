package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rassilon, the War President — Legendary Creature — Time Lord Noble
// {3}{U}{B}, 3/4 (Doctor Who):
//
//	"At the beginning of your upkeep, you lose 2 life and exile the top card of
//	 your library. You may play that card for as long as it remains exiled.
//	 Each noncreature spell you cast from exile has conspire. (As you cast that
//	 spell, you may tap two untapped creatures you control that share a color
//	 with it. When you do, copy it and you may choose new targets for the copy.)"
//
// The upkeep trigger is Gonti-style impulse with no end (the exiled card stays
// playable for as long as it is exiled). The grant is Raiding Schemes' optional
// conspire cost, offered only to a noncreature spell the controller is casting
// from exile (the card is still in exile as the optional costs are offered).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ebec9957-b777-42ae-a394-85c7ed84d00a",
		Name:         "Rassilon, the War President",
		Completeness: CompletenessFull,
		GrantedOptionalCosts: []game.GrantedOptionalCost{
			ConspireForSpellsYouCast("Each noncreature spell you cast from exile has conspire.", gapB85NoncreatureSpellFromExile),
		},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Rassilon — you lose 2 life and exile the top card of your library; you may play it for as long as it remains exiled", rassilonUpkeep),
		},
	})
}
