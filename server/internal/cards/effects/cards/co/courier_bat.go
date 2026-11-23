package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Courier Bat — Creature — Bat {2}{B}, 2/2 (Innistrad: Crimson Vow,
// #5495):
//
//	"Flying
//	 When this creature enters, if you gained life this turn, return up
//	 to one target creature card from your graveyard to your hand."
//
// An intervening "if" (CR 603.4) on the life-gained tally
// (CondLifeGained, Crested Sunmare's): it never triggers when you have
// gained no life as the Bat enters, and is checked again as the trigger
// resolves. The target is chosen as the trigger goes on the stack and may
// be left empty; the card returns only if it is still a creature card in
// your graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d8ff239c-cd71-4cd2-9e96-792a116989ca",
		Name:            "Courier Bat",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.InterveningIf(effects.CondLifeGained{},
				effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.CardID == source.InstanceID && effects.B15LifeGainedThisTurn(g, source.Controller) > 0
				}, "Courier Bat — return up to one target creature card from your graveyard to your hand",
					effects.ReturnTargetGraveyardCardToHand)),
				effects.TargetCardInGraveyard("up to one target creature card from your graveyard", effects.Creature(), effects.YouOwn()).WithCount(0, 1)),
		},
	})
}
