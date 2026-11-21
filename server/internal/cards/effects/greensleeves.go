package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Greensleeves — Legendary Planeswalker — Greensleeves {4}{G}{G}, loyalty 5
// (Mystery Booster Commander, #5948):
//
//	"+2: Create a Giant Badger token. (It's a {1}{G}{G} 2/2 Badger creature with
//	 "Whenever this token blocks, it gets +2/+2 until end of turn.")
//	 −3: Mill three cards. Put all permanent cards from among them into your
//	 hand.
//	 −8: Until end of turn, creatures you control have base power and toughness
//	 8/8 and gain trample.
//	 Greensleeves can be your commander."
//
// The Giant Badger is a catalog token (gapB06GiantBadgerToken) with Giant
// Badger's block trigger. The −3 mills three, then puts into the hand every
// permanent card among the cards that actually reached the graveyard (a mill a
// replacement sent elsewhere is not "among them"), still there when the tail runs.
// The −8 is one layer-7 effect over the creatures its controller controls as it
// resolves (CR 611.2c): base 8/8 and trample until end of turn; a creature that
// arrives later is not in it. "Can be your commander" is deck-building, read from
// the oracle text by the deck validator.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "20b90808-0279-46eb-ae51-6990ac679a83",
		Name:         "Greensleeves",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "+2: Create a Giant Badger token.",
				Cost:    LoyaltyCost(2),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect:  Do(CreateToken{Template: gapB06GiantBadgerToken(), N: 1}),
			},
			{
				Label: "−3: Mill three cards. Put all permanent cards from among them into your hand.",
				Cost:  LoyaltyCost(-3),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MillToZone{Player: item.Controller, N: 3, ThenTail: ctx.ItemTail(greensleevesTakePermanents)}.Apply(ctx)
				},
			},
			{
				Label: "−8: Until end of turn, creatures you control have base power and toughness 8/8 and gain trample.",
				Cost:  LoyaltyCost(-8),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return ScopedEffectFor{
						Match: And(Creature(), YouControl()),
						Mods: []game.Mod{
							game.SetBasePowerMod(8),
							game.SetBaseToughnessMod(8),
							game.AddKeywordsMod("trample"),
						},
						Duration: DurationUntilEndOfTurn(ctx),
						Label:    "Greensleeves — creatures you control have base power and toughness 8/8 and gain trample until end of turn",
					}.Apply(ctx)
				},
			},
		},
	})
}

// greensleevesTakePermanents is "put all permanent cards from among them into
// your hand": each permanent card among the milled cards that is still in the
// graveyard it landed in goes to its owner's hand. An on-disk key.
var greensleevesTakePermanents = CardTail("then/gap-b06-greensleeves-take-permanents", func(ctx *Context, o game.TailOutcome) error {
	g := ctx.Game
	for _, id := range o.IDs {
		c, ok := g.LookupCardForEffect(id)
		if z := g.FindCardZoneForEffect(id); !ok || z == nil || z.Kind != game.ZoneGraveyard || !isPermanentCard(c) {
			continue
		}
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
