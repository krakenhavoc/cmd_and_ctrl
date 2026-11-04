package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zealous Lorecaster — Creature — Giant Sorcerer {5}{R}, 4/4 (Secrets of
// Strixhaven, #4456):
//
//	"When this creature enters, return target instant or sorcery card from
//	 your graveyard to your hand."
//
// Pinnacle Monk's enters trigger: the target is chosen when the ability goes
// on the stack (CR 603.3d) and the trigger is removed if the graveyard holds
// no instant or sorcery card of yours. A card that left the graveyard in
// response is dropped by the legality re-check (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f82265fd-e772-4347-a27f-37b00ccd416b",
		Name:         "Zealous Lorecaster",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Zealous Lorecaster — return target instant or sorcery card from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return returnLegalGraveyardTargetsToHand(NewContext(g, item))
				}),
				TargetCardInGraveyard("target instant or sorcery card from your graveyard",
					YouOwn(), Or(Instant(), Sorcery())),
			),
		},
	})
}
