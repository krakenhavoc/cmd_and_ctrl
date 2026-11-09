package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fecund Greenshell — Creature — Elemental Turtle {3}{G}{G}, 4/6 (Bloomburrow):
//
//	"Reach
//	 As long as you control ten or more lands, creatures you control get +2/+2.
//	 Whenever this creature or another creature you control with toughness
//	 greater than its power enters, look at the top card of your library. If it's
//	 a land card, you may put it onto the battlefield tapped. Otherwise, put it
//	 into your hand."
//
// The anthem is a layer-7 modification over the creatures you control while the
// land count holds, read live every pass (a static locks no set, CR 611.3a); the
// land count moves with the battlefield, which already invalidates the layers.
// The trigger reads the entering creature's power and toughness through the
// layers as it enters (Greenshell's own 4/6 counts), once per creature. The look is
// private. A land card may go onto the battlefield tapped (the PutFromLibrary
// prompt); a land you decline, like any other card, goes into your hand, as the
// card's ruling says ("you may choose to put it into your hand rather than onto
// the battlefield tapped"). That is a put into hand, not a draw.
//
// No simplification.
func init() {
	anthem := game.StaticAbility{
		Label:    "As long as you control ten or more lands, creatures you control get +2/+2.",
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.Controller == source.Controller &&
				b02CountLandsControlledBy(g, source.Controller) >= 10
		},
		Apply: pumpBy(2, 2),
	}
	Register(Spec{
		OracleID:        "a090be5a-b0f2-4c17-8f7d-e742f9de8f6b",
		Name:            "Fecund Greenshell",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Static:          []game.StaticAbility{anthem},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, fecundCreatureEnters,
				"Fecund Greenshell — look at the top card of your library; a land may go onto the battlefield tapped, otherwise it goes into your hand",
				fecundLook),
		},
	})
}

// fecundCreatureEnters is "this creature or another creature you control with
// toughness greater than its power enters".
func fecundCreatureEnters(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventETB {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Controller == source.Controller && c.IsCreature() && c.CurrentToughness() > c.CurrentPower()
}

func fecundLook(g *game.Game, item *game.StackItem) error {
	player := item.Controller
	top := g.LookAtTopOfLibraryForEffect(player, 1)
	if len(top) == 0 {
		return nil
	}
	c, ok := g.LookupCardForEffect(top[0])
	if !ok {
		return nil
	}
	if !c.IsLand() {
		g.TakeFromLibraryToHandForEffect(player, top)
		return nil
	}
	return PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    top,
		Match:    Land(),
		Max:      1,
		Optional: true,
		Tapped:   true,
		Label:    "Fecund Greenshell — you may put the land card onto the battlefield tapped (otherwise it goes into your hand)",
		ThenTail: fecundLandRest,
	}.Apply(NewContext(g, item))
}

// fecundLandRest is the land you didn't put onto the battlefield: it goes into
// your hand. An on-disk key: never renamed, never reused.
var fecundLandRest = PutFromLibraryTail("fecund-greenshell/land-declined-into-hand",
	func(ctx *Context, res PutFromLibraryResult) error {
		if len(res.Entered) > 0 || len(res.Rest) == 0 {
			return nil
		}
		ctx.Game.TakeFromLibraryToHandForEffect(res.Player, res.Rest)
		return nil
	})
