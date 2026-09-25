package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eternal Witness — 2/1 Human Shaman for {1}{G}{G}:
//
//	"When Eternal Witness enters the battlefield, you may return
//	target card from your graveyard to your hand."
//
// S20 sub-PR 2: the S14/S19 "top of your graveyard" auto-pick is
// gone — Targets is "target card in your graveyard", the controller
// answers "you may", then clicks the card in their graveyard (the
// zone browser surfaces graveyard cards as targets). With an empty
// graveyard the trigger is removed without a prompt (CR 603.3d).
func init() {
	Register(Spec{
		OracleID:     "30b24e8e-3b0e-4d8e-90f3-f66eb7c1858c",
		Name:         "Eternal Witness",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetCardInGraveyard("target card in your graveyard", YouOwn()),
			Key:     "Eternal Witness — return target card to hand",
			Effect:  returnTargetCardToHand,
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Eternal Witness — return a card from your graveyard?",
			},
		}},
	})
}
