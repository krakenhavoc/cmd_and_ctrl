package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grist, the Hunger Tide — Legendary Planeswalker — Grist {1}{B}{G},
// loyalty 3 (Duskmourn Commander, #5427):
//
//	"As long as Grist isn't on the battlefield, it's a 1/1 Insect creature
//	 in addition to its other types.
//	 +1: Create a 1/1 black and green Insect creature token, then mill a
//	     card. If an Insect card was milled this way, put a loyalty counter
//	     on Grist and repeat this process.
//	 −2: You may sacrifice a creature. When you do, destroy target creature
//	     or planeswalker.
//	 −5: Each opponent loses life equal to the number of creature cards in
//	     your graveyard."
//
// The first sentence is Spec.OffBattlefield (game/off_battlefield_self.go):
// CR 113.6c, "An ability that states which zones it doesn't function in
// functions everywhere except for the specified zones, even outside the game
// and before the game begins." In a hand, a library, a graveyard, exile and
// the command zone Grist is a 1/1 Insect creature card as well as a
// planeswalker card, and on the stack it is a creature spell, so a creature
// tutor finds it, a reanimation spell returns it, a "creature spell" counter
// stops it, and it is a legal commander (CR 903.3: "(a) a creature card").
// On the battlefield it is a planeswalker and nothing else.
//
// +1 is one process, each step waiting for the one before: the token (through
// the CR 701.7b window), then the mill, whose tail is told every card the
// mill took out of the library, wherever a replacement sent it (CR 701.17c:
// "An effect that refers to a milled card can find that card in the zone it
// moved to from the library, as long as that zone is a public zone"). If one
// of them is an Insect card (Grist itself is one, in a graveyard or exile),
// Grist gets a loyalty counter and the process starts again. The repeat does
// not need Grist: a Grist that has left the battlefield gets no counter, and
// the tokens and mills go on. An empty library mills nothing and ends it.
//
// −2 asks "you may sacrifice a creature" as it resolves (maySacrificeOne,
// CR 608.2c); a creature that really went raises the printed "when you do" as
// a CR 603.12 reflexive trigger, which chooses its own target as it goes on
// the stack, after the sacrifice.
//
// −5 counts the creature cards in the graveyard as it resolves; a Grist that
// is there by then is one of them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0efb0d7e-dea0-4817-a243-15066e9ef333",
		Name:            "Grist, the Hunger Tide",
		Completeness:    CompletenessFull,
		StartingLoyalty: 3,
		OffBattlefield: &game.OffBattlefieldSelf{
			Types:     []string{"Creature"},
			Subtypes:  []string{"Insect"},
			Power:     1,
			Toughness: 1,
		},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Create a 1/1 black and green Insect creature token, then mill a card. If an Insect card was milled this way, put a loyalty counter on Grist and repeat this process.",
				Cost:    LoyaltyCost(1),
				Purpose: game.Purpose{Tokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return gristProcess(NewContext(g, item))
				},
			},
			{
				Label: "−2: You may sacrifice a creature. When you do, destroy target creature or planeswalker.",
				Cost:  LoyaltyCost(-2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Grist, the Hunger Tide — sacrifice a creature to destroy target creature or planeswalker?", Creature(), false, gristSacrificed)
				},
			},
			{
				Label: "−5: Each opponent loses life equal to the number of creature cards in your graveyard.",
				Cost:  LoyaltyCost(-5),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return eachOpponentLosesLife(g, item, b11CreatureCardsInGraveyard(g, item.Controller))
				},
			},
		},
	})
}

// gristProcess is one turn of the +1's process: the token, then (once it
// is made) the mill.
func gristProcess(ctx *Context) error {
	return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: ctx.Controller(),
		Source:     ctx.Item.SourceCardID,
		Groups:     []game.TokenGroup{{Template: ctx.TextToken(TokenCard("1/1 black and green Insect")), Count: 1}},
	}, ctx.ItemTail(gristMill))
}

// gristMill is "then mill a card", after the token. An on-disk key: never
// renamed, never reused.
var gristMill = CardTail("then/grist-the-hunger-tide-mill", func(ctx *Context, _ game.TailOutcome) error {
	// The "moved" form: a card Rest in Peace exiled instead is still the
	// milled card (CR 701.17c).
	return ctx.Game.MillToZoneKeyedMovedForEffect(ctx.Controller(), 1, game.ZoneGraveyard, ctx.ItemTail(gristMilled))
})

// gristMilled is "If an Insect card was milled this way, put a loyalty
// counter on Grist and repeat this process." An on-disk key.
//
// Assigned in init: the tail starts the process again, and the process
// names the tail, which a package-level initializer cannot say.
var gristMilled game.TailRef

func init() {
	gristMilled = CardTail("then/grist-the-hunger-tide-milled", func(ctx *Context, o game.TailOutcome) error {
		if !gristMilledAnInsect(ctx, o.IDs) {
			return nil
		}
		// AddCounter does nothing for a Grist that left the battlefield or
		// came back as a new object; the process repeats all the same.
		if err := (AddCounter{Target: ctx.Item.SourceCardID, Kind: game.CounterLoyalty, N: 1}).Apply(ctx); err != nil {
			return err
		}
		return gristProcess(ctx)
	})
}

// gristMilledAnInsect reports whether one of the milled cards is an Insect
// card, read where it went, as long as that is a public zone (CR 701.17c).
func gristMilledAnInsect(ctx *Context, milled []uuid.UUID) bool {
	g := ctx.Game
	// The word as Grist's text now says it (CR 612.2).
	insect := ctx.TextSubtype("Insect")
	for _, id := range milled {
		z := g.FindCardZoneForEffect(id)
		if z == nil {
			continue
		}
		switch z.Kind {
		case game.ZoneGraveyard, game.ZoneExile, game.ZoneCommand:
		default:
			continue
		}
		if c, ok := g.LookupCardForEffect(id); ok && !c.FaceDown && c.HasSubtype(insect) {
			return true
		}
	}
	return false
}

// gristSacrificed raises the −2's reflexive trigger once a creature has
// really gone, keyed (#4290).
var gristSacrificed = RegisterSacrificedOne("prompt-run/grist-the-hunger-tide-sacrifice", func(ctx *Context, _ uuid.UUID) error {
	return ReflexiveTrigger{
		Label: "Grist, the Hunger Tide — destroy target creature or planeswalker",
		Body:  gristDestroyBody,
	}.Apply(ctx)
})

// gristDestroyBody is the reflexive trigger's registration, with its own
// target clause.
var gristDestroyBody = game.ReflexiveBody("grist-the-hunger-tide/destroy",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		return destroyTheTargetPermanent(item, NewContext(g, item))
	}),
	constTargets(func() *game.TargetSpec {
		return TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker()))
	}))
