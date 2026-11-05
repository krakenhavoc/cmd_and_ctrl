package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// B'Elanna Torres, Scrapper — Legendary Creature — Klingon Human
// Artificer {1}{U}{R}, 2/2 (Star Trek Commander, #4974):
//
//	"Vigilance, trample
//	 {2}, {T}: Surveil 2, then you may reveal the top card of your
//	 library. Draw a card if it's an artifact card. If you don't draw a
//	 card this way, put two +1/+1 counters on B'Elanna Torres."
//
// In printed order: the surveil runs first, and the rest waits for the
// player to file the two cards (Surveil.ThenTail), so the card offered
// for the reveal is the top card the surveil left. "You may reveal" is a
// real prompt. On a yes the top card is revealed, and if it is an
// artifact card its controller draws a card. Whether a card was drawn
// is read off the turn's draw tally, so a declined reveal, a non-artifact
// card, an empty library or a draw a rule stopped (CR 121.2b) all put
// the two counters on B'Elanna, and only on her as the object that
// activated the ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c905227d-b5fe-42e6-a11d-5008b7a7922f",
		Name:            "B'Elanna Torres, Scrapper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Surveil 2, then you may reveal the top card of your library. Draw a card if it's an artifact card. If you don't draw a card this way, put two +1/+1 counters on B'Elanna Torres.",
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return Surveil{N: 2, ThenTail: ctx.ItemTail(belannaTorresAfterSurveil)}.Apply(ctx)
			},
		}},
	})
}

// belannaTorresAfterSurveil asks the "you may reveal". An on-disk key:
// never renamed, never reused.
var belannaTorresAfterSurveil = CardTail("belanna-torres-scrapper/after-surveil", func(ctx *Context, _ game.TailOutcome) error {
	return MayChoice{
		Question: "B'Elanna Torres, Scrapper — reveal the top card of your library? (Draw a card if it's an artifact card; otherwise put two +1/+1 counters on B'Elanna Torres.)",
		YesLabel: "Reveal it",
		NoLabel:  "Don't reveal",
		Then:     belannaTorresRevealAnswered,
	}.Apply(ctx)
})

// belannaTorresRevealAnswered reveals on a yes, draws for an artifact
// card, and puts the counters on her unless a card was drawn. An on-disk
// key: never renamed, never reused.
var belannaTorresRevealAnswered = MayAnswer("confirm/belanna-torres-scrapper-reveal", func(ctx *Context, r game.ConfirmAnswer) error {
	me := ctx.Controller()
	drew := false
	if r.Accepted {
		before := ctx.Game.TurnTallyFor(me).CardsDrawn
		for _, id := range ctx.Game.RevealTopOfLibraryForEffect(me, ctx.Source(), 1, "B'Elanna Torres, Scrapper — reveals the top card of their library") {
			if c, ok := ctx.Game.LookupCardForEffect(id); ok && !c.IsToken() && c.IsArtifact() {
				if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
		}
		drew = ctx.Game.TurnTallyFor(me).CardsDrawn > before
	}
	if drew {
		return nil
	}
	return CounterOnThis{Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
})
