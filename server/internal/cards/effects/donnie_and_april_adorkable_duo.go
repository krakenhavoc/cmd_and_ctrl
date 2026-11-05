package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Donnie & April, Adorkable Duo — Legendary Creature — Mutant Ninja
// Human Turtle {4}{U}, 3/3 (TMNT Commander, #5011):
//
//	"When Donnie & April enter, choose one or both. Each mode must target
//	 a different player.
//	 • Target player draws two cards.
//	 • Target player returns an artifact, instant, or sorcery card from
//	   their graveyard to their hand."
//
// A modal enters trigger whose two "target player" clauses are Distinct
// (tmcB7EachModeADifferentPlayer). The second bullet's card is not a
// target: the target player picks it from their own graveyard (a public
// zone) as the bullet resolves, and a player with no such card returns
// nothing.
//
// No simplification.
func init() {
	enters := WhenThisEnters("Donnie & April — choose one or both; each mode targets a different player",
		func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = tmcB7EachModeADifferentPlayer(
		ModeDoing("Target player draws two cards.", tmcB7TargetPlayer(),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				p, ok := tmcB7ModePlayer(ctx, occ)
				if !ok {
					return nil
				}
				return DrawCards{Player: p, N: 2}.Apply(ctx)
			}),
		ModeDoing("Target player returns an artifact, instant, or sorcery card from their graveyard to their hand.",
			tmcB7TargetPlayer(), donnieAndAprilReturn),
	)
	Register(Spec{
		OracleID:     "7b33368f-0668-4233-8bbf-725c66c771cb",
		Name:         "Donnie & April, Adorkable Duo",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}

// donnieAndAprilReturn is the second bullet: the target player chooses
// one artifact, instant or sorcery card in their graveyard and returns it
// to their hand.
func donnieAndAprilReturn(_ *game.StackItem, ctx *Context, occ int) error {
	p, ok := tmcB7ModePlayer(ctx, occ)
	if !ok {
		return nil
	}
	player := ctx.PlayerByID(p)
	if player == nil || player.Graveyard == nil {
		return nil
	}
	match := tmcB7ArtifactInstantOrSorcery()
	var cards []uuid.UUID
	for _, c := range player.Graveyard.Cards {
		if match(ctx.Game, p, c) {
			cards = append(cards, c.InstanceID)
		}
	}
	if len(cards) == 0 {
		return nil
	}
	queueReturnPickedToHand(ctx, p, "Donnie & April — return an artifact, instant, or sorcery card from your graveyard to your hand", cards)
	return nil
}
