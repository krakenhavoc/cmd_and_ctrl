package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "b8cdf79f-b247-4346-8122-9a2d7d23b3c3",
		Name:         "Pillar Tombs of Aku",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtEachUpkeep("Pillar Tombs of Aku — that player may sacrifice a creature", pillarTombsOfAkuUpkeep),
		},
	})
}

func pillarTombsOfAkuUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Trigger().Event.Actor
	if player == uuid.Nil {
		return nil
	}
	cost, action := SacrificePayment(1, "creature", "creatures", QueryType("creature")).prompt()
	return UpkeepPayUnless{
		Chooser:  player,
		Cost:     cost,
		Action:   action,
		Question: "Pillar Tombs of Aku — sacrifice a creature, or lose 5 life and its controller sacrifices it?",
		OnDecline: func(ctx *Context) error {
			ctx.RecordStep(pillarTombsOfAkuRecord, []uuid.UUID{player})
			return pillarTombsOfAkuDecline.Run(ctx)
		},
	}.Apply(ctx)
}

// pillarTombsOfAkuRecord files the player who declined.
const pillarTombsOfAkuRecord = "pillar-tombs-of-aku/player"

// pillarTombsOfAkuDecline is the life loss and then the sacrifice: the
// sacrifice waits for a prompt the life loss asks (a CR 616 order; CR 608.2c,
// #4155).
var pillarTombsOfAkuDecline = RegisterThenSteps("then-steps/pillar-tombs-of-aku-decline",
	Instruction(func(ctx *Context) error {
		ids := ctx.StepRecord(pillarTombsOfAkuRecord).IDs
		if len(ids) == 0 {
			return nil
		}
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ids[0], -5)
	}),
	Instruction(func(ctx *Context) error {
		return SacrificeThisIfStillOnBattlefield(ctx.Game, ctx.Item)
	}),
)
