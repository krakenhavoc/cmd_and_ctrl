package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skyclave Aerialist // Skyclave Invader — {1}{U} Creature — Merfolk Scout
// 2/1 // Creature — Phyrexian Merfolk Scout 2/4 (March of the Machine):
//
//	Front: "Flying
//	        {4}{G/P}: Transform this creature. Activate only as a sorcery."
//	Back:  "Flying
//	        When this creature transforms into Skyclave Invader, look at
//	        the top card of your library. If it's a land card, you may put
//	        it onto the battlefield. If you don't put the card onto the
//	        battlefield, put it into your hand."
//
// The trigger watches the turn-over to the back face only (CR 701.27e).
// The look is private. A land may go onto the battlefield; a land you
// decline, like any nonland card, goes into your hand (a put into hand,
// not a draw), as in Fecund Greenshell.
//
// No simplification.
func init() {
	const oracle = "29e96c47-94ad-4740-ba86-66dc2cf9ae2f"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Skyclave Aerialist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:        "{4}{G/P}: Transform this creature. Activate only as a sorcery.",
			Cost:         ManaCost("{4}{G/P}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Skyclave Invader",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventTransform, ThisTransformedInto(1),
				"Skyclave Invader — look at the top card of your library; a land may go onto the battlefield, otherwise it goes into your hand",
				skyclaveInvaderLook),
		},
	})
}

func skyclaveInvaderLook(g *game.Game, item *game.StackItem) error {
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
		Label:    "Skyclave Invader — you may put the land card onto the battlefield (otherwise it goes into your hand)",
		ThenTail: skyclaveInvaderLandRest,
	}.Apply(NewContext(g, item))
}

// skyclaveInvaderLandRest is the land you didn't put onto the battlefield:
// it goes into your hand. An on-disk key: never renamed, never reused.
var skyclaveInvaderLandRest = PutFromLibraryTail("skyclave-invader/land-declined-into-hand",
	func(ctx *Context, res PutFromLibraryResult) error {
		if len(res.Entered) > 0 || len(res.Rest) == 0 {
			return nil
		}
		ctx.Game.TakeFromLibraryToHandForEffect(res.Player, res.Rest)
		return nil
	})
