package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sweet-Gum Recluse — Creature — Spider {4}{G}{G}, 0/3 (Commander Legends:
// Battle for Baldur's Gate):
//
//	"Flash
//	 Cascade
//	 Reach
//	 When this creature enters, put three +1/+1 counters on each of any number
//	 of target creatures that entered this turn."
//
// Flash and reach are printed keywords; cascade is the engine keyword (CR
// 702.85). The entry trigger is Filigree Vector's shape: "any number of" is a
// clause of zero or more picks, each re-checked as the trigger resolves (CR
// 608.2b), restricted to creatures that entered the battlefield this turn —
// the Recluse itself and anything cascade put in included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8ea10a5a-b698-4f7d-a9c8-520f8d835f34",
		Name:            "Sweet-Gum Recluse",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "reach"},
		Triggered: []game.TriggeredAbility{
			effects.Cascade(),
			effects.Targeting(
				effects.WhenThisEnters("Sweet-Gum Recluse — three +1/+1 counters on each of any number of target creatures that entered this turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, t := range ctx.ClauseTargets(0) {
							if !ctx.IsTargetLegal(t) {
								continue
							}
							if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 3}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					}),
				effects.TargetCreature("any number of target creatures that entered this turn",
					func(g *game.Game, _ uuid.UUID, c game.Card) bool { return g.EnteredThisTurn(c.InstanceID) }).WithCount(0, 0),
			),
		},
	})
}
