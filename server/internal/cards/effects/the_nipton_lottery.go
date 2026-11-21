package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Nipton Lottery — Sorcery {2}{B}{R} (Fallout):
//
//	"Choose a creature at random. You gain control of that creature until
//	 end of turn. Untap it. It gains haste until end of turn. Then destroy
//	 all other creatures."
//
// The choice is one draw from the engine's keyed random stream (Call from
// the Grave's shape) over every creature on the battlefield, any
// controller's, mine included: nothing is targeted, so hexproof and
// protection don't matter. The creature is taken until end of turn, untapped
// and given haste; then every other creature is destroyed as one sweep, so
// indestructible ones survive and the chosen one does not die. With no
// creature on the battlefield nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8af798a9-2198-4f39-aa6d-5420a0f1203c",
		Name:         "The Nipton Lottery",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			g := ctx.Game
			var ids []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.IsCreature() {
					ids = append(ids, c.InstanceID)
				}
			}
			chosen := g.ChooseAtRandomForEffect(randomDraw(ctx), ids, 1)
			if len(chosen) == 0 {
				return nil
			}
			pick := chosen[0]
			if err := (GainControl{Target: pick, Label: "The Nipton Lottery — gain control until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			if err := (UntapTarget{Target: pick}).Apply(ctx); err != nil {
				return err
			}
			if err := (GrantKeywordUntilEOT{Target: pick, Keywords: []string{"haste"}, Label: "The Nipton Lottery — haste"}).Apply(ctx); err != nil {
				return err
			}
			return DestroyAllMatching{Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
				return c.IsCreature() && c.InstanceID != pick
			}}.Apply(ctx)
		},
	})
}
