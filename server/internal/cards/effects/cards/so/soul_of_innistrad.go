package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soul of Innistrad — {4}{B}{B} Creature — Avatar 6/6 (Commander 2020):
//
//	"Deathtouch
//	 {3}{B}{B}: Return up to three target creature cards from your graveyard to
//	 your hand.
//	 {3}{B}{B}, Exile this card from your graveyard: Return up to three target
//	 creature cards from your graveyard to your hand."
//
// Two printed abilities, two rows: the battlefield one and Bonebind Orator's
// graveyard one (the card is exiled as a cost, so it is never among its own
// targets). Both return whichever targets are still legal as they resolve
// (CR 608.2b).
func init() {
	targets := func() *game.TargetSpec {
		return effects.TargetCardInGraveyard("up to three target creature cards from your graveyard", effects.Creature(), effects.YouOwn()).WithCount(0, 3)
	}
	effects.Register(effects.Spec{
		OracleID:        "b5ba8155-520b-4053-be31-2744b77fb862",
		Name:            "Soul of Innistrad",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}{B}{B}: Return up to three target creature cards from your graveyard to your hand.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{3}{B}{B}"),
				Targets: targets(),
				Effect:  soulOfInnistradReturn,
			},
			{
				Label:   "{3}{B}{B}, Exile this card from your graveyard: Return up to three target creature cards from your graveyard to your hand.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{3}{B}{B}"), effects.ExileThis()),
				Zones:   []game.ZoneKind{game.ZoneGraveyard},
				Targets: targets(),
				Effect:  soulOfInnistradReturn,
			},
		},
	})
}

func soulOfInnistradReturn(g *game.Game, item *game.StackItem) error {
	return effects.ReturnLegalGraveyardTargetsToHand(effects.NewContext(g, item))
}
