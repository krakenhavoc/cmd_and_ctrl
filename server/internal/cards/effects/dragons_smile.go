package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragon's Smile — Legendary Artifact — Vehicle {5}{R}{R}, 4/5 (Mystery
// Booster Commander, #5948):
//
//	"Flying
//	 Once during each of your turns, you may pay {0} rather than pay the mana
//	 cost for a spell you cast from exile.
//	 Whenever Dragon's Smile deals combat damage to a player, that player
//	 exiles that many cards from the top of their library. You may play those
//	 cards this turn.
//	 Crew 3"
//
// The second line is Warped Space's granted alternative cost (ADR 0118) limited
// to your own turn as Access Maze's is, once per Dragon's Smile per turn, so it
// covers any spell cast from exile, the cards its own trigger exiles included.
// The trigger exiles as many cards as damage was dealt, off the damaged
// player's library, with the standard impulse permission for its controller
// (CR 406.3 and the "this turn" window of ExileTopWithPermission). The cards
// are exiled by that player's library, so a short library gives fewer.
// Crew 3 is the ordinary CR 702.122 ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                "e7611ef9-e5fc-45eb-bc3f-4358a27107fd",
		Name:                    "Dragon's Smile",
		Completeness:            CompletenessFull,
		PrintedKeywords:         []string{"flying"},
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{PayZeroForASpellFromExileOnceEachOfYourTurns()},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Dragon's Smile — that player exiles that many cards from the top of their library; you may play those cards this turn",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
						return nil
					}
					return ExileTopWithPermission{
						From:    item.Trigger.Event.Target,
						GrantTo: item.Controller,
						N:       item.Trigger.Event.Amount,
					}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Dragon's Smile"),
		}},
	})
}
