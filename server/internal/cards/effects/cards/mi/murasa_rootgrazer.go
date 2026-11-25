package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Murasa Rootgrazer — Creature — Beast {G}{W}, 2/3 (Zendikar Rising, #5508):
//
//	"Vigilance
//	 {T}: You may put a basic land card from your hand onto the battlefield.
//	 {T}: Return target basic land you control to its owner's hand."
//
// Two tap abilities, so the Rootgrazer taps for either. The put is the family
// of Growth Spiral's "you may": a pick over the basic land cards in hand with
// zero a legal answer. The bounce targets a basic land, which can be one the
// first ability just put (the first ability resolves before the Rootgrazer
// can untap, so they are never both used in one activation).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ca7964a1-b4ce-412b-bcae-e0dfbb95389a",
		Name:            "Murasa Rootgrazer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: You may put a basic land card from your hand onto the battlefield.",
			Purpose: game.Purpose{Lands: 1, Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect: effects.Do(effects.PutFromHandOntoBattlefield{
				Match:    effects.OldB096BasicLandMatch,
				Optional: true,
				Label:    "Murasa Rootgrazer — you may put a basic land card from your hand onto the battlefield",
			}),
		}, {
			Label:   "{T}: Return target basic land you control to its owner's hand.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target basic land you control", effects.OldB096BasicLandMatch, effects.YouControl()),
			Effect:  effects.BounceTheTargetToHand,
		}},
	})
}
