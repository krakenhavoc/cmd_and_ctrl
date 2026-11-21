package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Triskaidekaphobia — Enchantment {3}{B} (Innistrad Remastered):
//
//	"At the beginning of your upkeep, choose one —
//	 • Each player with exactly 13 life loses the game, then each player
//	   gains 1 life.
//	 • Each player with exactly 13 life loses the game, then each player
//	   loses 1 life."
//
// The players at exactly 13 are found as the ability resolves and lose
// together (LoseTheGameTogetherForEffect), so two players left who are both
// at 13 draw rather than one winning. "Each player" then means each player
// still in the game; a player who can't lose (Platinum Angel) still changes
// life.
func init() {
	t := AtYourUpkeep("Triskaidekaphobia — choose one", func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = ChooseOne(
		ModeDoing("Each player with exactly 13 life loses the game, then each player gains 1 life.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return triskaidekaphobiaMode(ctx, item, 1)
			}),
		ModeDoing("Each player with exactly 13 life loses the game, then each player loses 1 life.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return triskaidekaphobiaMode(ctx, item, -1)
			}),
	)
	Register(Spec{
		OracleID:     "d5b131c7-29d4-4745-95e3-70bdf9bf54b3",
		Name:         "Triskaidekaphobia",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}

func triskaidekaphobiaMode(ctx *Context, item *game.StackItem, delta int) error {
	g := ctx.Game
	var thirteen []uuid.UUID
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated && p.Life == 13 {
			thirteen = append(thirteen, p.ID)
		}
	}
	if len(thirteen) > 0 {
		if _, err := g.LoseTheGameTogetherForEffect(thirteen, item.SourceCardID); err != nil {
			return err
		}
		if g.State != game.StateActive {
			return game.ErrStopResolution
		}
	}
	var everyone []uuid.UUID
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated {
			everyone = append(everyone, p.ID)
		}
	}
	for _, id := range everyone {
		if err := g.ChangePlayerLifeForEffect(item.SourceCardID, id, delta); err != nil {
			return err
		}
	}
	return nil
}
