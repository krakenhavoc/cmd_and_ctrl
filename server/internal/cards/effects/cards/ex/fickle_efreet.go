package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fickle Efreet — Creature — Efreet {3}{R}, 5/2 (Prophecy, S82 OLD b671, #5617):
//
//	"Whenever this creature attacks or blocks, flip a coin at end of combat. If
//	 you lose the flip, an opponent gains control of this creature."
//
// Runaway Carriage's attack-or-block trigger (one printed ability, one registered
// trigger) over a delayed trigger at the beginning of this combat's end of combat
// step (CR 511.2, CR 603.7), as its sacrifice is. The flip is the controller's
// called coin flip (Volatile Rig's keyed continuation); on a lost flip the
// controller chooses which opponent gains control (the Wishclaw Talisman ruling),
// with no end (CR 611.2a), and only while the Efreet is still the object that
// attacked or blocked (CR 400.7).
//
// No simplification.
func init() {
	const label = "Fickle Efreet — flip a coin at end of combat; if you lose the flip, an opponent gains control of it"
	effects.Register(effects.Spec{
		OracleID:     "e5cd1024-ec1a-4df9-8bc6-ad2f3f443d06",
		Name:         "Fickle Efreet",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclared(ev, source) || effects.SelfBlocksOnce(ev, source)
			}, label, func(g *game.Game, item *game.StackItem) error {
				return effects.ScheduleDelayedTrigger{
					At:    game.StepEndCombat,
					Label: "Fickle Efreet — flip a coin",
					Body:  fickleEfreetFlipBody,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

// fickleEfreetFlipBody is the delayed trigger: its controller flips a coin if the
// Efreet is still the object that attacked or blocked.
var fickleEfreetFlipBody = game.SimpleDelayedBody("s82-old-b671/fickle-efreet-flip", func(g *game.Game, item *game.StackItem) error {
	if g.AbilitySourceGoneForEffect(item) {
		return nil
	}
	g.FlipCoinForEffect(game.CoinFlipSpec{
		Flipper:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Fickle Efreet — call the coin flip",
		ThenKey:  fickleEfreetFlipped,
		Asked:    game.NewAskedItem(item),
	})
	return nil
})

// fickleEfreetFlipped is "if you lose the flip, an opponent gains control": the
// controller chooses the opponent. An on-disk key: never renamed, never reused.
var fickleEfreetFlipped = effects.CoinFlipThen("coin-flip/s82-old-b671-fickle-efreet", func(ctx *effects.Context, r game.CoinFlipped) error {
	if effects.WonFirstFlip(r) {
		return nil
	}
	return effects.ChoosePlayer{
		Chooser:  ctx.Controller(),
		Only:     ctx.Opponents(),
		Question: "Fickle Efreet — choose an opponent to gain control of it",
		ThenKey:  fickleEfreetHandedOver,
	}.Apply(ctx)
})

// fickleEfreetHandedOver is the opponent chosen. An on-disk key: never renamed,
// never reused.
var fickleEfreetHandedOver = effects.ChoosePlayerThen("option-pick/s82-old-b671-fickle-efreet-hand-over", func(ctx *effects.Context, chosen uuid.UUID, _ []uuid.UUID) error {
	g, item := ctx.Game, ctx.Item
	if chosen == uuid.Nil || effects.SourceIsNewObject(g, item) {
		return nil
	}
	g.GainControlForEffect(item.SourceCardID, item.SourceCardID, chosen, game.IndefiniteDuration(),
		"Fickle Efreet — an opponent gains control of this creature")
	return nil
})
