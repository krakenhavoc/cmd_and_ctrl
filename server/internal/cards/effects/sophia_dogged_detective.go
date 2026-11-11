package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sophia, Dogged Detective — Legendary Creature — Human Detective
// {1}{G}{W}{U}, 3/4 (Murders at Karlov Manor Commander):
//
//	"When Sophia enters, create Tiny, a legendary 2/2 green Dog Detective
//	 creature token with trample.
//	 {1}, Sacrifice an artifact token: Put a +1/+1 counter on each Dog you
//	 control.
//	 Whenever a Dog you control deals combat damage to a player, create a
//	 Food token, then investigate."
//
// Three pieces. The token is the printed Tiny. The sacrifice cost takes any
// artifact token (a Clue, a Food, a Treasure), paid up front, and the
// counters go on every Dog read as the ability resolves
// (avB30aPutCounterOnEachOfSubtypeYouControl, Lo and Li's shape). The
// third is per Dog, not per batch: each Dog that connects triggers once,
// so two Dogs hitting the same player make two Foods and two Clues.
// The Food comes first, then the Clue (CR 701.16a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eaf8027a-017a-4471-ae7b-a372fd8e0ff3",
		Name:         "Sophia, Dogged Detective",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sophia, Dogged Detective — create Tiny, a legendary 2/2 Dog Detective with trample",
				Do(CreateToken{Template: TokenCard("2/2 green Legendary Dog Detective named Tiny with trample"), N: 1})),
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !combatDamageToPlayerBy(ev, source.Controller, g) {
					return false
				}
				dealer, ok := g.LookupCardForEffect(ev.Source)
				return ok && game.ThroughTextOf(source, dealer).HasSubtype("Dog")
			}, "Sophia, Dogged Detective — create a Food token, then investigate",
				Do(CreateToken{Template: FoodToken(), N: 1}, Investigate{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice an artifact token: Put a +1/+1 counter on each Dog you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(ManaCost("{1}"), SacrificeN(1, "an artifact token", Artifact(), IsTokenPredicate())),
			Effect:  avB30aPutCounterOnEachOfSubtypeYouControl("Dog"),
		}},
	})
}
