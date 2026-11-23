package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pillar Tombs of Aku — World Enchantment {2}{B}{B}:
//
//	"At the beginning of each player's upkeep, that player may
//	 sacrifice a creature of their choice. If that player doesn't, they
//	 lose 5 life and you sacrifice this enchantment."
//
// Every player's upkeep, its controller's included. "That player" is
// the active player the upkeep event names, and they are asked through
// UpkeepPayUnless with a sacrifice payment (ADR 0108 §5): it stops the
// table during their own upkeep, offers only creatures they control,
// and a player with no creature cannot pay. Declining loses 5 life and
// the Tombs' controller sacrifices it; a Tombs that has already left
// the battlefield is not sacrificed again (CR 701.21a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b8cdf79f-b247-4346-8122-9a2d7d23b3c3",
		Name:         "Pillar Tombs of Aku",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Pillar Tombs of Aku — that player may sacrifice a creature", pillarTombsOfAkuUpkeep),
		},
	})
}

func pillarTombsOfAkuUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	player := ctx.Trigger().Event.Actor
	if player == uuid.Nil {
		return nil
	}
	cost, action := effects.SacrificePayment(1, "creature", "creatures", effects.QueryType("creature")).Prompt()
	return effects.UpkeepPayUnless{
		Chooser:  player,
		Cost:     cost,
		Action:   action,
		Question: "Pillar Tombs of Aku — sacrifice a creature, or lose 5 life and its controller sacrifices it?",
		Else:     pillarTombsOfAkuDeclined,
	}.Apply(ctx)
}

// pillarTombsOfAkuRecord files the player who declined.
const pillarTombsOfAkuRecord = "pillar-tombs-of-aku/player"

// pillarTombsOfAkuDecline is the life loss and then the sacrifice: the
// sacrifice waits for a prompt the life loss asks (a CR 616 order; CR 608.2c,
// #4155).
var pillarTombsOfAkuDecline = effects.RegisterThenSteps("then-steps/pillar-tombs-of-aku-decline",
	effects.Instruction(func(ctx *effects.Context) error {
		ids := ctx.StepRecord(pillarTombsOfAkuRecord).IDs
		if len(ids) == 0 {
			return nil
		}
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ids[0], -5)
	}),
	effects.Instruction(func(ctx *effects.Context) error {
		return effects.SacrificeThisIfStillOnBattlefield(ctx.Game, ctx.Item)
	}),
)

// pillarTombsOfAkuDeclined is the decline, keyed (#4264): the player who declined is the payer.
var pillarTombsOfAkuDeclined = effects.PayUnlessElse("pay-unless/pillar-tombs-of-aku-decline",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		ctx.RecordStep(pillarTombsOfAkuRecord, []uuid.UUID{r.Chooser})
		return pillarTombsOfAkuDecline.Run(ctx)
	})
