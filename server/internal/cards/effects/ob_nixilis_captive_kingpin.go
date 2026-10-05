package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ob Nixilis, Captive Kingpin — Legendary Creature — Demon {2}{B}{R},
// 4/3:
//
//	"Flying, trample
//	 Whenever one or more opponents each lose exactly 1 life, put a
//	 +1/+1 counter on Ob Nixilis. Exile the top card of your library.
//	 Until your next end step, you may play that card."
//
// The trigger reads the TOTAL each opponent lost in one simultaneous
// event batch (#2183, game/life_batch.go), so two 1-damage sources
// into one opponent in the same combat damage step are a loss of 2
// and do nothing, one drain of 1 to each of two opponents is one
// trigger, and two separate resolutions (two Blood Artist triggers)
// are two. A damage doubler changes the total; life paid counts.
//
// The play window is game.UntilYourNextEndStep (#2373): from your own
// turn before its end step it closes as this turn's end step begins,
// otherwise as your next turn's does.
func init() {
	Register(Spec{
		OracleID:        "55b6434b-1542-40fd-b12a-697d40976582",
		Name:            "Ob Nixilis, Captive Kingpin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreOpponentsEachLoseExactly(1,
				"Ob Nixilis, Captive Kingpin — +1/+1 counter, exile the top card, you may play it",
				obNixilisKingpinEffect),
		},
	})
}

// obNixilisKingpinEffect is the resolution: the counter (only while
// Ob Nixilis is still the object that triggered), then the impulse.
func obNixilisKingpinEffect(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (AddCounter{Target: ctx.Source(), Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
		return err
	}
	return ExileTopNUntilYourNextEndStep(ctx, 1)
}
