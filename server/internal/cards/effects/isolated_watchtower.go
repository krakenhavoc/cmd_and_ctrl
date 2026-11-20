package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Isolated Watchtower — Land (Commander 2018, #5526):
//
//	"{T}: Add {C}.
//	 {2}, {T}: Scry 1, then you may reveal the top card of your library. If
//	 a basic land card is revealed this way, put it onto the battlefield
//	 tapped. Activate only if an opponent controls at least two more lands
//	 than you."
//
// The activation condition (CR 602.5b) is Land Tax's count with a margin of
// two: some opponent still in the game controls at least two more lands
// than you as you activate. In printed order: the scry runs first and the
// rest waits for the card to be filed (Scry.ThenTail), so the card offered
// for the reveal is the top card the scry left. "You may reveal" is a real
// prompt, B'Elanna Torres' shape. On a yes the top card is revealed, and if
// it is a basic land card it enters tapped under your control; any other
// card stays on top.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3893f320-47fd-49ec-a78c-80bfb607a279",
		Name:         "Isolated Watchtower",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Scry 1, then you may reveal the top card of your library. If a basic land card is revealed this way, put it onto the battlefield tapped. Activate only if an opponent controls at least two more lands than you.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				mine := b03LandsControlled(g, controller)
				for _, p := range g.Seats {
					if p == nil || p.Eliminated || p.ID == controller {
						continue
					}
					if b03LandsControlled(g, p.ID) >= mine+2 {
						return true
					}
				}
				return false
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return Scry{N: 1, ThenTail: ctx.ItemTail(isolatedWatchtowerAfterScry)}.Apply(ctx)
			},
		}},
	})
}

// isolatedWatchtowerAfterScry asks the "you may reveal". An on-disk key:
// never renamed, never reused.
var isolatedWatchtowerAfterScry = CardTail("isolated-watchtower/after-scry", func(ctx *Context, _ game.TailOutcome) error {
	return MayChoice{
		Question: "Isolated Watchtower — reveal the top card of your library? (A basic land card is put onto the battlefield tapped.)",
		YesLabel: "Reveal it",
		NoLabel:  "Don't reveal",
		Then:     isolatedWatchtowerRevealAnswered,
	}.Apply(ctx)
})

// isolatedWatchtowerRevealAnswered reveals on a yes and puts a basic land
// card onto the battlefield tapped. An on-disk key: never renamed, never
// reused.
var isolatedWatchtowerRevealAnswered = MayAnswer("confirm/isolated-watchtower-reveal", func(ctx *Context, r game.ConfirmAnswer) error {
	if !r.Accepted {
		return nil
	}
	me := ctx.Controller()
	for _, id := range ctx.Game.RevealTopOfLibraryForEffect(me, ctx.Source(), 1, "Isolated Watchtower — reveals the top card of their library") {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.IsToken() || !IsBasicLand(c) {
			continue
		}
		if _, err := ctx.Game.PutFromLibraryOntoBattlefieldForEffect(id, game.LibraryEntryOptions{
			Controller: me,
			Tapped:     true,
		}); err != nil {
			return err
		}
	}
	return nil
})
