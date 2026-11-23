package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grismold, the Dreadsower — Legendary Creature — Troll Shaman {1}{B}{G},
// 3/3 (Commander 2019, #5517):
//
//	"Trample
//	 At the beginning of your end step, each player creates a 1/1 green
//	 Plant creature token.
//	 Whenever a creature token dies, put a +1/+1 counter on Grismold."
//
// The end-step trigger is Elephant-Mandrill's "each player creates" shape:
// every player still in the game, in APNAP order, creates the token under
// their own control. The counter trigger watches every creature token that
// dies, anyone's, read off the event's last-known types (CR 603.10a); a
// token that is put into a graveyard really dies before it ceases to exist
// (CR 704.5d, 111.7), so it triggers Grismold.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d0d28e2c-416e-432c-b4bf-07f6302c52ea",
		Name:            "Grismold, the Dreadsower",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Grismold, the Dreadsower — each player creates a 1/1 green Plant", c19EachPlayerCreatesAPlant),
			effects.On(game.EventLTB, c19ACreatureTokenDied,
				"Grismold, the Dreadsower — put a +1/+1 counter on Grismold",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}

// c19EachPlayerCreatesAPlant is "each player creates a 1/1 green Plant
// creature token".
func c19EachPlayerCreatesAPlant(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, p := range effects.SeatsFromActive(g) {
		if err := (effects.CreateToken{Controller: p, Template: effects.TokenCard("1/1 green Plant"), N: 1}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// c19ACreatureTokenDied is "whenever a creature token dies": any player's.
func c19ACreatureTokenDied(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := effects.DiedCreature(ev, g)
	return ok && effects.IsToken(dead)
}
