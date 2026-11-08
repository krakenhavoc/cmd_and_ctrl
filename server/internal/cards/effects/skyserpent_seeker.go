package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skyserpent Seeker — Creature — Snake {G}{U}, 1/1 (Aetherdrift, #4464):
//
//	"Flying, deathtouch
//	 Exhaust — {4}: Reveal cards from the top of your library until you
//	 reveal two land cards. Put those land cards onto the battlefield
//	 tapped and the rest on the bottom of your library in a random
//	 order. Put a +1/+1 counter on this creature. (Activate each exhaust
//	 ability only once.)"
//
// Will Riker, First in the Field's reveal (the run stops on the second
// land card; a library that runs out first reveals itself whole and puts
// whatever lands it found) with Open the Way's tapped entry. The put is
// keyed, and its tail puts the rest on the bottom in a random order and
// then the counter on the Seeker, in the printed order. A Seeker that
// left or came back as a new object in response gets no counter
// (CR 400.7, 608.2b); the lands still come.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5da8b86a-a7be-4fb5-9a45-a47d5cb4b372",
		Name:            "Skyserpent Seeker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Activated: []ActivatedAbility{{
			Label:   "Exhaust — {4}: Reveal cards from the top of your library until you reveal two land cards. Put those land cards onto the battlefield tapped and the rest on the bottom of your library in a random order. Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Exhaust: true,
			Cost:    ManaCost("{4}"),
			Effect:  skyserpentSeekerReveal,
		}},
	})
}

// skyserpentSeekerReveal reveals until two land cards and puts them onto
// the battlefield tapped; skyserpentSeekerTail does the rest.
func skyserpentSeekerReveal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	run, _ := revealUntilCount(ctx, ctx.Controller(), func(c game.Card) bool { return c.IsLand() }, 2,
		"Skyserpent Seeker — revealed until two land cards")
	return PutFromLibraryOntoBattlefield{
		Player:   ctx.Controller(),
		Cards:    run,
		Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsLand() },
		All:      true,
		Tapped:   true,
		ThenTail: skyserpentSeekerTail,
	}.Apply(ctx)
}

// skyserpentSeekerTail puts the rest on the bottom in a random order,
// then a +1/+1 counter on the Seeker. An on-disk key.
var skyserpentSeekerTail = PutFromLibraryTail("skyserpent-seeker/rest-on-bottom-then-counter",
	func(ctx *Context, res PutFromLibraryResult) error {
		if err := PutRestOnBottomInRandomOrder(ctx.Game, res); err != nil {
			return err
		}
		return AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	})
