package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// blocks_by_piles.go — the two "blocks decided by piles" clauses (#6200).
// The engine halves are game/blocks_by_piles.go (Camouflage) and
// game/left_right_piles.go (Raging River).

// BlocksByPilesThisTurn is Camouflage's "This turn, instead of declaring
// blockers, each defending player chooses any number of creatures they
// control and divides them into a number of piles equal to the number of
// attacking creatures for whom that player is the defending player. …
// Assign each pile to a different one of those attacking creatures at
// random. Each creature in a pile that can block the creature that pile
// is assigned to does so." The cast restriction ("Cast this spell only
// during your declare attackers step") is the card's own.
//
//	Effect: BlocksByPilesThisTurn("Camouflage"),
func BlocksByPilesThisTurn(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		g.BlocksByPilesThisTurnForEffect(ctx.Source(), ctx.Controller(),
			name+" — this turn, each defending player's blocks are decided by piles")
		return nil
	}
}

// DivideIntoLeftAndRightPiles is Raging River's "each defending player
// divides all creatures without flying they control into a "left" pile
// and a "right" pile. Then, for each attacking creature you control,
// choose "left" or "right." That creature can't be blocked this combat
// except by creatures with flying and creatures in a pile with the chosen
// label." Put it on the "whenever one or more creatures you control
// attack" trigger.
func DivideIntoLeftAndRightPiles() Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		g.QueueLeftRightPilesForEffect(ctx.Source(), ctx.Controller())
		return nil
	}
}
