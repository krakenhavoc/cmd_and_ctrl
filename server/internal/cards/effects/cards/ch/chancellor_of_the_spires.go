package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chancellor of the Spires — Creature — Phyrexian Sphinx {4}{U}{U}{U}, 5/7
// (New Phyrexia, S82 OLD batch 463, #5575):
//
//	"You may reveal this card from your opening hand. If you do, at the
//	 beginning of the first upkeep, each opponent mills seven cards.
//	 Flying
//	 When this creature enters, you may cast target instant or sorcery
//	 card from an opponent's graveyard without paying its mana cost."
//
// The opening-hand reveal is the #6101 seam (Providence's "first upkeep"
// moment, CR 103.6): the delayed trigger belongs to the player who revealed
// it, and mills every opponent still in the game. The enter trigger is
// Memory Plunder's offer (BnsB13FreeCastFromGraveyard) with the
// target chosen as the trigger goes on the stack, as Impulsivity does it.
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Chancellor of the Spires — you may cast target instant or sorcery card from an opponent's graveyard without paying its mana cost",
		func(g *game.Game, item *game.StackItem) error {
			return effects.BnsB13FreeCastFromGraveyard("Chancellor of the Spires")(g, item, game.EffectParams{})
		})
	enters.Targets = effects.TargetCardInGraveyard("target instant or sorcery card from an opponent's graveyard",
		effects.Or(effects.Instant(), effects.Sorcery()), effects.Not(effects.YouOwn()))
	effects.Register(effects.Spec{
		OracleID:        "e766a7ec-2761-453b-bfba-956de7b5e125",
		Name:            "Chancellor of the Spires",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		OpeningHand: effects.RevealFromOpeningHand(game.RevealAtTheFirstUpkeep,
			"Chancellor of the Spires — each opponent mills seven cards", chancellorOfTheSpiresFirstUpkeep),
		Triggered: []game.TriggeredAbility{enters},
	})
}

// chancellorOfTheSpiresFirstUpkeep is the revealed card's delayed trigger:
// each opponent of the player who revealed it mills seven. An on-disk key:
// never renamed, never reused.
var chancellorOfTheSpiresFirstUpkeep = game.SimpleDelayedBody("chancellor-of-the-spires/first-upkeep-mill", func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var err error
	effects.EachOpponent(g, item.Controller, func(opp uuid.UUID) bool {
		err = effects.MillCards{Player: opp, N: 7}.Apply(ctx)
		return err != nil
	})
	return err
})
