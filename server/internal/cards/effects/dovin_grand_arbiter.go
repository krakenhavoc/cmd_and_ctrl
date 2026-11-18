package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dovin, Grand Arbiter — Legendary Planeswalker — Dovin {1}{W}{U},
// starting loyalty 3:
//
//	"+1: Until end of turn, whenever a creature you control deals combat
//	 damage to a player, put a loyalty counter on Dovin.
//	 −1: Create a 1/1 colorless Thopter artifact creature token with
//	 flying. You gain 1 life.
//	 −7: Look at the top ten cards of your library. Put three of them
//	 into your hand and the rest on the bottom of your library in a
//	 random order."
//
// THE +1 is a CR 603.7 delayed trigger that repeats until the turn ends
// (DelayedOnEvent.Repeats, Great Train Heist's shape): one trigger for
// each creature you control that deals combat damage to a player, so
// three attackers connecting give three loyalty counters. A Dovin that
// has left the battlefield (or come back as a new object) gets nothing.
//
// THE −1 is the Thopter and the life gain, in that order.
//
// THE −7 looks at ten cards and takes exactly three (all of them if the
// library holds fewer), the rest going to the bottom in a random order
// (Artificer Class's seam).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "596237da-3d51-4052-a8ca-2218efc1afb3",
		Name:            "Dovin, Grand Arbiter",
		Completeness:    CompletenessFull,
		StartingLoyalty: 3,
		Activated: []ActivatedAbility{
			{
				Label: "+1: Until end of turn, whenever a creature you control deals combat damage to a player, put a loyalty counter on Dovin.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DelayedOnEvent{
						Label:     "Dovin, Grand Arbiter — put a loyalty counter on Dovin",
						On:        []game.EventKind{game.EventDealDamage},
						Condition: dovinGrandArbiterConnects,
						Body:      dovinGrandArbiterCounterBody,
						Repeats:   true,
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label:  "−1: Create a 1/1 colorless Thopter artifact creature token with flying. You gain 1 life.",
				Cost:   LoyaltyCost(-1),
				Effect: Do(CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 1}, GainLife{Amount: 1}),
			},
			{
				Label:   "−7: Look at the top ten cards of your library. Put three of them into your hand and the rest on the bottom of your library in a random order.",
				Cost:    LoyaltyCost(-7),
				Purpose: game.Purpose{Draws: 3},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					player := ctx.Controller()
					return TakeFromLibraryToHand{
						Player: player,
						Cards:  g.LookAtTopOfLibraryForEffect(player, 10),
						Max:    3,
						Label:  "Dovin, Grand Arbiter — put three of them into your hand and the rest on the bottom of your library in a random order",
						Key:    takeRestBottomRandom,
					}.Apply(ctx)
				},
			},
		},
	})
}

var (
	// dovinGrandArbiterConnects is "a creature you control deals combat
	// damage to a player", one match per creature dealing it.
	dovinGrandArbiterConnects = game.DelayedCondition("combat-damage/dovin-grand-arbiter-creature-connects",
		func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
			return combatDamageToPlayerBy(ev, dt.Controller, g)
		})

	dovinGrandArbiterCounterBody = game.SimpleDelayedBody("dovin-grand-arbiter/loyalty-counter", dovinGrandArbiterCounter)
)

func dovinGrandArbiterCounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	dovin := item.SourceCardID
	if ctx.isNewSourceObject(dovin) || !onBattlefield(g, dovin) {
		return nil
	}
	return AddCounter{Target: dovin, Kind: game.CounterLoyalty, N: 1}.Apply(ctx)
}
