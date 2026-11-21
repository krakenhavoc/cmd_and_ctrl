package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b127_helpers.go — S82 set GAP, batch 127 (#5948): the Lord of the
// Rings white, blue and gold rows. Append-only.

// gapB127ScryTargetsFromLookedAt is the target clause of "whenever you scry,
// <do something> to up to X target <things>, where X is the number of cards
// looked at while scrying this way": built from the scry event the trigger
// fired on, so a scry 2 offers up to two targets.
func GapB127ScryTargetsFromLookedAt(build func(x int) *game.TargetSpec) func(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
	return func(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
		if tc.Event.LookedAt <= 0 {
			return nil
		}
		return build(tc.Event.LookedAt)
	}
}

// gapB127ColorProtections is "protection from each of <that creature's>
// colors": the protection tokens for every color the card has, as it last
// stood. A colorless creature has none.
func GapB127ColorProtections(c game.Card) []string {
	var out []string
	for _, col := range c.EffectiveColors() {
		if tok := game.ProtectionFromColor(col); tok != "" {
			out = append(out, tok)
		}
	}
	return out
}

// gapB127DrewTwoOrMoreCards is "as long as you've drawn two or more cards
// this turn" as a cost condition: the caster's draws this turn.
func GapB127DrewTwoOrMoreCards(q game.CostQuery) bool {
	return q.Game != nil && q.Game.TurnTallyFor(q.Controller).CardsDrawn >= 2
}

var _ = uuid.Nil

// gapB127IsOpponentsTurn is "during an opponent's turn": the active player
// is someone other than `you`.
func GapB127IsOpponentsTurn(g *game.Game, you uuid.UUID) bool {
	s := g.Turn.ActiveSeat
	if s < 0 || s >= len(g.Seats) || g.Seats[s] == nil {
		return false
	}
	return g.Seats[s].ID != you
}

// gapB127PlayerAttackedYouThisTurn is "they attacked you this turn": some
// creature of `player`'s was declared attacking `you` (the player, not one
// of their planeswalkers or battles) this turn.
func GapB127PlayerAttackedYouThisTurn(g *game.Game, player, you uuid.UUID) bool {
	if player == uuid.Nil || you == uuid.Nil || player == you {
		return false
	}
	for _, a := range g.TurnTally.Attacks {
		if a.Defender == you && a.PlaneswalkerController == uuid.Nil && !a.Battle && a.Controller == player {
			return true
		}
	}
	return false
}
