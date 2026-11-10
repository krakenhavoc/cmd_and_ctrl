package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Bean — Plane — Chicago (Unknown Event, #4870):
//
//	"When you planeswalk here, each player chooses a creature they control
//	 and creates a token that's a copy of it, except it's a Reflection in
//	 addition to its other types.
//	 All Reflections have shadow. (They can block or be blocked only by
//	 creatures with shadow.)
//	 Whenever chaos ensues, create a token that's a copy of a creature you
//	 control, except it's a Reflection in addition to its other types."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6). The
// planeswalk-to ability asks each seat with a creature, in turn order from
// the planar controller, to choose one of their own (the ordered collector)
// and each of them creates the copy, so a token copy is made under its
// chooser's control (CR 707.9b adds the type; the copy keeps its other
// types). "All Reflections have shadow" is a layer-6 grant to every
// Reflection on the battlefield, whoever controls it, so it covers the
// copies and any other Reflection. The chaos ability is not targeted: the
// planar controller chooses as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b977ff66-6c52-4f33-8f25-e310d835be8b",
		Name:         "The Bean",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "All Reflections have shadow.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.HasSubtype(source.TextSubtype("Reflection"))
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = append(c.Abilities, "shadow")
			},
		}},
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThis("The Bean — each player chooses a creature they control and creates a token copy of it, a Reflection in addition",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return b13EachPlayerChoosesACreature(ctx, item.Controller,
						"The Bean — choose a creature you control; create a token copy of it that's also a Reflection",
						theBeanEveryoneCopies, nil)
				}),
			WheneverChaosEnsues("The Bean — create a token copy of a creature you control, a Reflection in addition",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return ChoosePermanents{
						Question: "The Bean — choose a creature you control to copy",
						Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
							return b13CreaturesOf(g, of), 1, 1
						},
						Keyed: theBeanChaosCopies,
					}.Apply(ctx)
				}),
		},
	})
}

// b13ReflectionCopy is the exception clause: "it's a Reflection in addition
// to its other types" (CR 707.9b).
func b13ReflectionCopy(t *game.Card) { t.TypeLine = addedSubtypeTypeLine(t.TypeLine, "Reflection") }

// theBeanEveryoneCopies is each chooser creating their token copy. An on-disk
// key: never renamed, never reused.
var theBeanEveryoneCopies = game.RegisterPromptRunThen("prompt-run/the-bean-planeswalk-copies", func(g *game.Game, r game.RunDone) error {
	ctx := NewContext(g, r.Item())
	for _, leg := range r.Picks() {
		for _, id := range leg.Cards {
			if err := (CreateTokenCopy{Controller: leg.Seat, Copy: id, N: 1, Except: b13ReflectionCopy}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return nil
})

// theBeanChaosCopies is the chaos ability's copy of the chosen creature. An
// on-disk key: never renamed, never reused.
var theBeanChaosCopies = RegisterPermanentsPick("prompt-run/the-bean-chaos-copy", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if err := (CreateTokenCopy{Controller: ctx.Controller(), Copy: id, N: 1, Except: b13ReflectionCopy}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	},
})
