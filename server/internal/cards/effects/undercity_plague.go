package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Undercity Plague — Sorcery {4}{B}{B}:
//
//	"Target player loses 1 life, discards a card, then sacrifices a
//	 permanent of their choice.
//	 Cipher (Then you may exile this spell card encoded on a creature you
//	 control. Whenever that creature deals combat damage to a player, its
//	 controller may cast a copy of the encoded card without paying its
//	 mana cost.)"
//
// In printed order (CR 608.2c): the life, then the discard the player
// chooses (CR 701.9a), then — once the card is in the graveyard — the
// sacrifice of a permanent they choose. Cipher is the engine's keyword
// (game/cipher.go, CR 702.99, #3867), asked after the sacrifice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "aa997e63-be19-4b1a-8972-eebc8a720ded",
		Name:            "Undercity Plague",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordCipher},
		Targets:         TargetPlayer("target player"),
		OnResolve:       undercityPlagueResolve,
	})
}

func undercityPlagueResolve(_ *game.StackItem, ctx *Context) error {
	var player uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			player = t.ID
			break
		}
	}
	if player == uuid.Nil {
		return nil
	}
	if err := (GainLife{Player: player, Amount: -1}).Apply(ctx); err != nil {
		return err
	}
	source := ctx.Source()
	ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
		Player: player,
		Source: source,
		N:      1,
		Then: func(g *game.Game, seat uuid.UUID, _ []uuid.UUID) error {
			g.PlayerSacrificesNForEffect(source, seat, sacrificeSpec("a permanent"),
				"Undercity Plague — sacrifice a permanent", 1)
			return nil
		},
	})
	return nil
}
