package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deeproot Wayfinder — Creature — Merfolk Scout {1}{G}, 2/3 (March of the
// Machine, S82 MOM batch 13):
//
//	"Whenever this creature deals combat damage to a player or battle,
//	 surveil 1, then you may return a land card from your graveyard to the
//	 battlefield tapped."
//
// Bilbo's player-or-battle damage condition. The land is picked AFTER the
// surveil, as the tail of the surveil (Teval's shape): a land the surveil
// just binned is a legal pick, and "you may" is a floor of zero. The land
// enters tapped on the entry event, not by a tap after it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ec767f2-40ba-46c9-b7cd-5821986dda0c",
		Name:         "Deeproot Wayfinder",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayerOrBattle,
				"Deeproot Wayfinder — surveil 1, then you may return a land card from your graveyard to the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.Surveil{N: 1, ThenTail: ctx.ItemTail(deeprootWayfinderAfterSurveil)}.Apply(ctx)
				}),
		},
	})
}

// deeprootWayfinderAfterSurveil is "then you may return a land card from
// your graveyard to the battlefield tapped", asked once the surveil is done.
var deeprootWayfinderAfterSurveil = effects.CardTail("s82-mom-b13/deeproot-wayfinder-after-surveil", func(ctx *effects.Context, _ game.TailOutcome) error {
	you := ctx.Controller()
	candidates := effects.GraveyardCardIDs(ctx, you, game.Card.IsLand)
	if len(candidates) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    you,
		FromPlayer: you,
		Source:     ctx.Source(),
		Question:   "Deeproot Wayfinder — you may return a land card from your graveyard to the battlefield tapped",
		Cards:      candidates,
		Min:        0,
		Max:        1,
		ThenKey:    effects.B25ReturnLandTapped,
		LiveZone:   game.ZoneGraveyard,
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
})
