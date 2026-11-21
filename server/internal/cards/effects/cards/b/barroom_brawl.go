package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Barroom Brawl — Sorcery {1}{G} (Commander Legends: Battle for Baldur's Gate,
// S82 GAP #5948):
//
//	"Target creature you control fights target creature the opponent to your
//	 left controls. Then that player may copy this spell and may choose new
//	 targets for the copy."
//
// Two clauses, the second bound to the opponent to the caster's left
// (Bowie Base One's reading: the first player still in the game walking left)
// and fought through Ruthless Predation's two-clause fight. "That player" is the
// controller of the second target, read before the fight (CR 608.2h), and is
// asked as the spell finishes whether to copy it. The copy is created by, and
// controlled by, that player (CR 707.10), who may choose new targets for it
// (CR 707.10c); the spell is held for its last-known information as the
// question may outlive its resolution (Wandering Archaic's #4276). Their copy
// reads "left" and "you control" from them.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d1ce7fcf-42f8-4b35-9ba1-55033c65c43e",
		Name:         "Barroom Brawl",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature the opponent to your left controls", effects.ControlledByThePlayerToYourLeft()),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			var that uuid.UUID
			if t, ok := ctx.ClauseTarget(1); ok {
				that, _ = effects.ControllerOfTarget(ctx, t.ID)
			}
			if err := effects.FightTargetClauses(ctx); err != nil {
				return err
			}
			if that == uuid.Nil {
				return nil
			}
			ctx.Game.HoldSpellForEffect(item.SourceCardID)
			return effects.MayChoice{
				Player:   that,
				Question: "Barroom Brawl — copy this spell and choose new targets for the copy?",
				Then:     barroomBrawlCopy,
				Carry:    []uuid.UUID{item.SourceCardID},
			}.Apply(ctx)
		},
	})
}

// barroomBrawlCopy is the answer to "copy this spell?": the copy is made by the
// answering player. An on-disk key.
var barroomBrawlCopy = effects.MayAnswer("confirm/barroom-brawl-copy", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if !r.Accepted || len(r.Carry) == 0 {
		return nil
	}
	return effects.CopySpell{
		StackID:          r.Carry[0],
		Controller:       r.Chooser,
		FromLastKnown:    true,
		ChooseNewTargets: true,
	}.Apply(ctx)
})
