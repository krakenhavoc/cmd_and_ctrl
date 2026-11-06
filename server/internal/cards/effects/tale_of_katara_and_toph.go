package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tale of Katara and Toph — Enchantment {2}{G} (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"Creatures you control have \"Whenever this creature becomes tapped
//	 for the first time during each of your turns, put a +1/+1 counter on
//	 it.\""
//
// A layer 6 grant of a catalog bundle (ADR 0093) to every creature its
// controller controls, so the trigger is each creature's own. "Becomes
// tapped" is any tap, an attack included (ThisBecameTapped); "for the
// first time" reads the engine's became-tapped tally for the object
// (Captain America, Living Legend's), which counts taps from before the
// creature had the ability too, and only during its controller's turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7c376ea9-d50a-4cc3-ae40-7bac3a926f20",
		Name:         "Tale of Katara and Toph",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: taleOfKataraAndTophGrant,
			Triggered: []game.TriggeredAbility{
				OnAny([]game.EventKind{game.EventTapCard, game.EventAttack},
					func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
						return ThisBecameTapped(ev, source, lki, g) && IsYourTurn(g, source.Controller) &&
							g.TimesBecameTappedThisTurn(ev.CardID) == 1
					}, "Tale of Katara and Toph — put a +1/+1 counter on this creature",
					Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			},
			Text: "Whenever this creature becomes tapped for the first time during each of your turns, put a +1/+1 counter on it.",
		}},
		Static: []game.StaticAbility{GrantAbilities(CreaturesYouControl, taleOfKataraAndTophGrant)},
	})
}

// taleOfKataraAndTophGrant is the granted bundle's catalog key. An on-disk
// key: never renamed, never reused.
const taleOfKataraAndTophGrant = "tale-of-katara-and-toph/first-tap-counter"
