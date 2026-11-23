package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mighty Servant of Leuk-o — Artifact — Vehicle {3}, 6/6 (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"Trample
//	 Ward—Discard a card.
//	 Whenever this Vehicle becomes crewed for the first time each turn, if it
//	 was crewed by exactly two creatures, it gains "Whenever this creature
//	 deals combat damage to a player, draw two cards" until end of turn.
//	 Crew 4"
//
// Ward—Discard a card is Ward(WardDiscard(1)) (CR 702.21a). The crew trigger
// has an intervening "if" (CR 603.4): exactly two distinct creatures paid the
// crew cost (CR 702.122a), checked as it triggers and as it resolves, and
// counting a creature that has since left. The Vehicle gains the draw trigger
// until end of turn (an ADR 0093 duration grant). Crew 4 is the ordinary
// CR 702.122 crew ability.
//
// No simplification.
const mightyServantGrant = "mighty-servant-of-leuk-o/draw-two"

func init() {
	effects.Register(effects.Spec{
		OracleID:        "0b59099f-2172-4eab-9c0d-7048e6771018",
		Name:            "Mighty Servant of Leuk-o",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Grants: []effects.AbilityGrant{{
			Key: mightyServantGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Mighty Servant of Leuk-o — draw two cards",
					effects.Do(effects.DrawCards{N: 2})),
			},
			Text: "Whenever this creature deals combat damage to a player, draw two cards.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardDiscard(1), "Mighty Servant of Leuk-o — ward, discard a card"),
			effects.WhenBecomesCrewedFirstTimeByExactlyTwo("Mighty Servant of Leuk-o — gains \"draw two cards when it deals combat damage to a player\" until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.GrantAbilitiesFor{
						Target: ctx.Source(),
						Keys:   []string{mightyServantGrant},
						Label:  "Mighty Servant of Leuk-o — draws two cards when it deals combat damage to a player",
					}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 4",
			Cost:    effects.CrewCost(4),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Mighty Servant of Leuk-o"),
		}},
	})
}
