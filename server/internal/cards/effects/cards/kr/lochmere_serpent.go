package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lochmere Serpent — Creature — Serpent {4}{U}{B}, 7/7 (Throne of Eldraine,
// S82 old sets):
//
//	"Flash
//	 {U}, Sacrifice an Island: This creature can't be blocked this turn.
//	 {B}, Sacrifice a Swamp: You gain 1 life and draw a card.
//	 {U}{B}: Exile five target cards from an opponent's graveyard. Return this
//	 card from your graveyard to your hand. Activate only as a sorcery."
//
// The first two are ordinary sacrifice costs (Rootwater Alligator's, with a
// mana payment), the first going to the Serpent only while it is still the
// permanent that activated. The third works from the graveyard (CR 113.6): it
// needs exactly five target cards, all from one opponent's graveyard (a set
// rule, so a table of smaller graveyards can't pay it), exiles those still
// legal as it resolves and returns this card to its owner's hand. If every
// target is gone the ability does nothing (CR 608.2b), so the Serpent stays.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "99422efb-3b44-4025-ad1b-388fc48a669a",
		Name:            "Lochmere Serpent",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{U}, Sacrifice an Island: This creature can't be blocked this turn.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{U}"), effects.SacrificeN(1, "an Island", effects.Subtype("Island"))),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					return effects.RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked, Label: "Lochmere Serpent — can't be blocked this turn"}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{B}, Sacrifice a Swamp: You gain 1 life and draw a card.",
				Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{B}"), effects.SacrificeN(1, "a Swamp", effects.Subtype("Swamp"))),
				Effect:  effects.Do(effects.GainLife{Amount: 1}, effects.DrawCards{N: 1}),
			},
			{
				Label:        "{U}{B}: Exile five target cards from an opponent's graveyard. Return this card from your graveyard to your hand. Activate only as a sorcery.",
				Cost:         effects.ManaCost("{U}{B}"),
				Zones:        []game.ZoneKind{game.ZoneGraveyard},
				SorcerySpeed: true,
				Targets: effects.TargetCardInGraveyard("five target cards from an opponent's graveyard", effects.Not(effects.YouOwn())).
					WithCount(5, 5).AllShare(effects.FromASingleGraveyard()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if err := effects.ExileTargetCards(g, item); err != nil {
						return err
					}
					return effects.ReturnThisCardFromYourGraveyardToYourHand(g, item)
				},
			},
		},
	})
}
