package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vexing Devil — Creature — Devil {R}, 4/3 (Innistrad Remastered, S82
// GAP, #5948):
//
//	"When this creature enters, any opponent may have it deal 4 damage to
//	 them. If a player does, sacrifice this creature."
//
// Desecration Demon's any-opponent offer (ADR 0150, #1903): each opponent
// in APNAP order from the active player is asked until one takes the 4
// damage, which the Devil deals through the ordinary damage path (so it can
// be prevented, and taking it still counts). Only if a player does is the
// Devil sacrificed, and only while it is still the permanent the trigger
// came from (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b582e485-c024-47f0-ad6f-b918d32288ba",
		Name:         "Vexing Devil",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Vexing Devil — any opponent may have it deal 4 damage to them; if a player does, sacrifice it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.AnyPlayerPays{
						Pay:          effects.DamagePayment(4),
						Opponents:    true,
						Question:     "Vexing Devil — have it deal 4 damage to you? (if you do, it is sacrificed)",
						AcceptLabel:  "Take 4 damage",
						DeclineLabel: "Don't",
						Then:         vexingDevilTaken,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// vexingDevilTaken is "if a player does, sacrifice this creature", keyed
// (#4264): the asking item carries the Devil's object (CR 400.7).
var vexingDevilTaken = effects.AnyPlayerConsequences("any-player/vexing-devil-sacrifice",
	sacrificeThisIfStillHereAnswer, nil, nil)

func sacrificeThisIfStillHereAnswer(ctx *effects.Context, _ game.AnyPlayerAnswer) error {
	return effects.SacrificeThisIfStillHere(ctx)
}
