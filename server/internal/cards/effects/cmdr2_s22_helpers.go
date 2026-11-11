package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s22_helpers.go — shared bodies of the "draw-matters, hand size
// and flash grants" Commander batch (#5368).

// cmdr2S22HandHas reports whether `player` is still in the game with
// exactly n cards in hand (Triskaidekaphile's intervening "if").
func cmdr2S22HandHas(g *game.Game, player uuid.UUID, n int) bool {
	p := g.PlayerByIDForEffect(player)
	return p != nil && !p.Eliminated && cardsInHand(g, player) == n
}

// cmdr2S22WinIfHandHas is "you win the game" behind that intervening
// "if", checked again as the trigger resolves (CR 603.4).
func cmdr2S22WinIfHandHas(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !cmdr2S22HandHas(g, item.Controller, n) {
			return nil
		}
		return WinTheGame{Player: item.Controller}.Apply(NewContext(g, item))
	}
}

// cmdr2S22TargetPlayerDrawsThenYouDraw is "target player draws N cards
// and you draw M cards", the target first as printed. A target that
// became illegal draws nothing; the controller still draws.
func cmdr2S22TargetPlayerDrawsThenYouDraw(theirs, yours int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetPlayer {
				continue
			}
			if err := (DrawCards{Player: t.ID, N: theirs}).Apply(ctx); err != nil {
				return err
			}
			break
		}
		return DrawCards{N: yours}.Apply(ctx)
	}
}

// cmdr2S22HexproofAndTapThis is "This creature gains hexproof until end
// of turn. Tap it." as an activated ability's effect, both halves the
// resolving ability's own.
func cmdr2S22HexproofAndTapThis(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if err := (GrantKeywordUntilEOT{
			Target:   item.SourceCardID,
			Keywords: []string{"hexproof"},
			Label:    label,
		}).Apply(ctx); err != nil {
			return err
		}
		return TapTarget{Target: item.SourceCardID}.Apply(ctx)
	}
}
