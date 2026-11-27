package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flamespeaker Adept — Creature — Human Shaman {2}{R}, 2/3 (Theros, S82 OLD b389, #5561):
//
//	"Whenever you scry, this creature gets +2/+0 and gains first strike
//	 until end of turn."
//
// One trigger on EventScry (emitted once a scry is complete, and not for a
// scry of an empty library), pumping the Adept itself and granting first
// strike, as Rot Farm Mortipede does for its own pump and keywords.
//
// No simplification.
func init() {
	const label = "Flamespeaker Adept — gets +2/+0 and gains first strike until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "ab0707fb-b404-433d-b845-0d50d01cd72e",
		Name:         "Flamespeaker Adept",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventScry, effects.ByYou, label,
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					if err := (effects.BoostUntilEOT{Target: ctx.Source(), Power: 2, Label: "Flamespeaker Adept — +2/+0"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: []string{"first strike"},
						Label: "Flamespeaker Adept — first strike"}.Apply(ctx)
				}))),
		},
	})
}
