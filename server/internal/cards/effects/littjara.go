package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Littjara — Plane — Kaldheim (Commander Masters, #4870):
//
//	"When you planeswalk to Littjara and at the beginning of your upkeep,
//	 create a 2/2 blue Shapeshifter creature token with changeling. (It is
//	 every creature type.)
//	 Whenever chaos ensues, choose a creature type. Put a +1/+1 counter on
//	 each creature you control of that type."
//
// A plane (ADR 0216 §6). The first ability is one ability with two trigger
// conditions making a changeling Shapeshifter for the planar controller. The
// chaos ability (CR 311.7) asks the planar controller to choose a creature
// type as it resolves, then puts a +1/+1 counter on each creature they
// control of that type, read through its effective subtypes so a changeling
// counts for every type (CR 702.73a) and the ruling about a creature that
// lost changeling holds.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0ff2bbed-e435-4108-b02a-564a75abba40",
		Name:         "Littjara",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThisAndAtYourUpkeep("Littjara — create a 2/2 blue Shapeshifter creature token with changeling",
				Do(CreateToken{Template: BlueShapeshifterToken(), N: 1})),
			WheneverChaosEnsues("Littjara — choose a creature type; put a +1/+1 counter on each creature you control of that type",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					ChooseCreatureTypeTail(g, item.Controller, item.SourceCardID, "Littjara — choose a creature type",
						ctx.ItemTail(littjaraTypeChosen))
					return nil
				}),
		},
	})
}

// littjaraTypeChosen puts the counters once the type is chosen. An on-disk
// key: never renamed, never reused.
var littjaraTypeChosen = creatureTypeChosen("littjara/counters-of-the-chosen-type", func(ctx *Context, creatureType string, _ game.TailOutcome) error {
	for _, id := range creaturesOfTypeYouControl(ctx.Game, ctx.Controller(), creatureType) {
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
