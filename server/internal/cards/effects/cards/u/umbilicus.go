package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Umbilicus — Artifact {4} (Dominaria Remastered):
//
//	"At the beginning of each player's upkeep, that player may pay 2 life.
//	 If they don't, they return a permanent they control to its owner's
//	 hand."
//
// Pillar Tombs of Aku's shape: every player's upkeep, its controller's
// included, "that player" being the active player the upkeep event names.
// They are asked through UpkeepPayUnless with a life payment (CR 119.4: a
// player can pay only with at least that much life). Declining, or being
// unable to pay, makes THAT player pick a permanent they control and
// return it to its owner's hand (Cache Raiders' pick, answered by the
// upkeep's player). With no permanent they control nothing is returned.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c9132e72-d016-4cb7-8662-a16b7d932bef",
		Name:         "Umbilicus",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Umbilicus — that player may pay 2 life; if they don't, they return a permanent they control to its owner's hand", umbilicusUpkeep),
		},
	})
}

func umbilicusUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	player := ctx.Trigger().Event.Actor
	if player == uuid.Nil {
		return nil
	}
	cost, action := effects.LifePayment(2).Prompt()
	return effects.UpkeepPayUnless{
		Chooser:  player,
		Cost:     cost,
		Action:   action,
		Question: "Umbilicus — pay 2 life, or return a permanent you control to its owner's hand?",
		Else:     umbilicusDeclined,
	}.Apply(ctx)
}

// umbilicusDeclined is the decline: the player who was asked returns a
// permanent they control. An on-disk key: never renamed, never reused.
var umbilicusDeclined = effects.PayUnlessElse("pay-unless/umbilicus-decline",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		return effects.ChoosePermanents{
			Player:   r.Chooser,
			Question: "Umbilicus — return a permanent you control to its owner's hand",
			Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
				var out []uuid.UUID
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == of {
						out = append(out, c.InstanceID)
					}
				}
				return out, 1, 1
			},
			Keyed: effects.BouncedChosen,
		}.Apply(ctx)
	})
