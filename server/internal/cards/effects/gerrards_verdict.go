package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gerrard's Verdict — Sorcery {W}{B} (Apocalypse, #5614):
//
//	"Target player discards two cards. You gain 3 life for each land card
//	 discarded this way."
//
// The target chooses the two cards from their own hand (CR 701.9b; a
// hand of one or none discards what it has), Mind Roots' keyed discard.
// The life counts the land cards that really left the hand that way,
// read after the discard, so the caster gains 0, 3 or 6. The caster may
// target themselves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "670b5dfb-ffdf-43ed-acc3-855006e352fc",
		Name:         "Gerrard's Verdict",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			victim := TargetedPlayer(ctx)
			if victim == uuid.Nil {
				return nil
			}
			return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
				Player:   victim,
				Source:   item.SourceCardID,
				N:        2,
				Question: "Gerrard's Verdict — discard two cards",
				Asked:    game.NewAskedItem(item),
			}, gerrardsVerdictLife, nil)
		},
	})
}

// gerrardsVerdictLife is the clause after the discard: the caster gains
// 3 life for each land card discarded this way. An on-disk key.
var gerrardsVerdictLife = game.RegisterDiscardThen("choose-cards/gerrards-verdict-life", func(g *game.Game, r game.DiscardsDone) error {
	if r.Item == nil {
		return nil
	}
	lands := 0
	for _, id := range r.Discarded.Cards() {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsLand() {
			lands++
		}
	}
	if lands == 0 {
		return nil
	}
	return GainLife{Player: r.Item.Controller, Amount: 3 * lands}.Apply(NewContext(g, r.Item))
})
