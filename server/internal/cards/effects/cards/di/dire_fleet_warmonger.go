package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dire Fleet Warmonger — Creature — Orc Pirate {1}{B}{R}, 3/3 (Core Set 2021,
// #5511):
//
//	"At the beginning of combat on your turn, you may sacrifice another
//	 creature. If you do, this creature gets +2/+2 and gains trample until
//	 end of turn."
//
// The sacrifice is asked as the trigger resolves, over the other creatures
// you control (never the Warmonger itself); declining, or having nothing to
// sacrifice, does nothing. "If you do" follows only a creature that really
// left, and the boost goes on the Warmonger only while it is still the
// permanent that triggered (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ee08435b-a2c1-491a-87cb-e60789b09c25",
		Name:         "Dire Fleet Warmonger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventStepBegan, effects.StepBegan(game.StepBeginCombat, true),
				"Dire Fleet Warmonger — you may sacrifice another creature; if you do, +2/+2 and trample until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Dire Fleet Warmonger — sacrifice another creature for +2/+2 and trample until end of turn?",
						effects.Creature(), true, direFleetWarmongerSacrificed)
				}),
		},
	})
}

// direFleetWarmongerSacrificed is "if you do": the +2/+2 and trample. Keyed:
// an on-disk identity, never renamed or reused.
var direFleetWarmongerSacrificed = effects.RegisterSacrificedOne("prompt-run/s82-old-b106-dire-fleet-warmonger-sacrifice", func(ctx *effects.Context, _ uuid.UUID) error {
	if !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	if err := (effects.BoostUntilEOT{Target: ctx.Source(), Power: 2, Toughness: 2, Label: "Dire Fleet Warmonger — +2/+2"}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: []string{"trample"}, Label: "Dire Fleet Warmonger — trample"}.Apply(ctx)
})
