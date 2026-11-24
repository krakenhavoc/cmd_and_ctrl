package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thijarian Witness — Creature — Alien Cleric {1}{G}, 0/4 (Doctor Who):
//
//	"Flash
//	 Bear Witness — Whenever another creature dies, if it was attacking or
//	 blocking alone, exile it and investigate."
//
// "Attacking or blocking alone" is CR 506.5 as the creature last existed
// (CR 603.10a): EventLTB's AttackingAlone / BlockingAlone (#6098), judged
// against the whole simultaneous event, so two attackers that die together
// in one combat damage step were not attacking alone. It is a fact about the
// past, so the intervening "if" (CR 603.4) cannot change between the trigger
// and its resolution. "It" is the card that died, exiled only while it is
// still that object in the graveyard; the investigate happens either way,
// after the exile (thijarianWitnessInvestigate is the exile's tail).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d273033b-3b8b-4ce9-afe8-208d82dd20f1",
		Name:            "Thijarian Witness",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.AnotherCreatureDied(ev, source, lki, g) && (ev.AttackingAlone || ev.BlockingAlone)
			}, "Bear Witness — exile it and investigate", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				dead := item.Trigger.Event.CardID
				if !effects.DiedCardStillInGraveyard(ctx, dead) {
					return effects.Investigate{N: 1}.Apply(ctx)
				}
				return effects.ExileTarget{Target: dead, ThenTail: ctx.ItemTail(thijarianWitnessInvestigate)}.Apply(ctx)
			}),
		},
	})
}

// thijarianWitnessInvestigate is "and investigate", after the exile has
// landed (or been replaced). An on-disk identity: never renamed.
var thijarianWitnessInvestigate = effects.CardTail("thijarian-witness/investigate", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.Investigate{N: 1}.Apply(ctx)
})
