package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Crowd Goes Wild — Sorcery {X}{G}:
//
//	"Assist (Another player can pay up to {X} of this spell's cost. You
//	 choose the value of X.)
//	 Support X. (Put a +1/+1 counter on each of up to X target creatures.)
//	 Each creature with a +1/+1 counter on it gains trample until end of
//	 turn."
//
// Support X is its reminder text: a +1/+1 counter on each of up to X
// target creatures. The pick is "up to X" (CountFromX and UpToX, as
// March of Swirling Mist's pick is), so X = 0 picks nothing. The
// counters land first, then the trample grant reads every creature
// with a +1/+1 counter as it resolves (CR 611.2c), so the freshly
// countered creatures are in the set.
func init() {
	Register(Spec{
		OracleID:        "21302f37-0632-4202-9823-f08c5aec016e",
		Name:            "The Crowd Goes Wild",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAssist},
		Targets:         marchOfSwirlingMistTargets(),
		// The trample waits for a prompt the counters ask (a CR 616 order; CR 608.2c, #4155).
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
				}
				return nil
			}),
			GrantKeywordUntilEOT{
				Match:    crowdGoesWildCountered,
				Keywords: []string{"trample"},
				Label:    "The Crowd Goes Wild — trample until end of turn",
			},
		),
	})
}

// crowdGoesWildCountered is "a creature with a +1/+1 counter on it".
func crowdGoesWildCountered(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return c.IsCreature() && c.Counters[game.CounterPlusOne] > 0
}
