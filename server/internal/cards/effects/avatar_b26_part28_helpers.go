package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b26_part28_helpers.go — shared pieces for Avatar: The Last
// Airbender Eternal (#5075), batch 26 (part 28).

// avB28CreatureLeftUnderYourControl is "a creature left the battlefield
// under your control this turn" (Tale of Momo, That's Rough Buddy): the
// per-player CreaturesLeft tally, so a creature that left by any route,
// to any zone, earlier in the turn counts, tokens included.
func avB28CreatureLeftUnderYourControl(g *game.Game, you uuid.UUID) bool {
	return g != nil && g.CreatureLeftThisTurn(you)
}

// avB28DrawThenDiscardUnlessThreeLessons is Chakra Meditation's "Then
// discard a card unless there are three or more Lesson cards in your
// graveyard": the graveyard is counted after the draw, as the discard
// would be made. The key is an on-disk identity.
var avB28DrawThenDiscardUnlessThreeLessons = game.RegisterDrawThen("discard-unless-three-lessons", func(g *game.Game, d game.DrawThen) error {
	if avB4LessonsInGraveyard(g, d.Player) >= 3 {
		return nil
	}
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: d.Player, Source: d.Source, N: d.N})
	return nil
})
