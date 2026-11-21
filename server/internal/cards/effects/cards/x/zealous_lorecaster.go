package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "f82265fd-e772-4347-a27f-37b00ccd416b",
		Name:         "Zealous Lorecaster",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Zealous Lorecaster — return target instant or sorcery card from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ReturnLegalGraveyardTargetsToHand(effects.NewContext(g, item))
				}),
				effects.TargetCardInGraveyard("target instant or sorcery card from your graveyard",
					effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery())),
			),
		},
	})
}
