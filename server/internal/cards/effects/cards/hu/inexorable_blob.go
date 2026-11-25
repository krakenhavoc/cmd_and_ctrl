package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Inexorable Blob — Creature — Ooze {2}{G}, 3/3 (Shadows over Innistrad,
// #5543):
//
//	"Delirium — Whenever this creature attacks, if there are four or more
//	 card types among cards in your graveyard, create a 3/3 green Ooze
//	 creature token that's tapped and attacking."
//
// An intervening if (CR 603.4) on the controller's own graveyard, checked as
// the Blob attacks and again as the trigger resolves. The Ooze enters tapped
// and attacking; the effect names nothing for it to attack, so its
// controller chooses (CR 508.4), preferring what the Blob attacks. It was
// never declared, so no "whenever attacks" trigger sees it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bddf0772-503c-42f1-bb86-1dfeeedf4a7e",
		Name:         "Inexorable Blob",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.B24CondDelirium{},
				effects.On(game.EventAttack, effects.AllOf(effects.ThisAttacked, effects.YouHaveDeliriumWhen),
					"Inexorable Blob — delirium: create a 3/3 green Ooze creature token that's tapped and attacking",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						return effects.CreateTokensAttackingYourChoice{
							Template: effects.TokenCard("3/3 green Ooze"),
							N:        1,
							Tapped:   true,
							Prefer:   ctx.Trigger().Event.Target,
							Label:    "Inexorable Blob",
						}.Apply(ctx)
					})),
		},
	})
}
