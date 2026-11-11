package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Timeless Witness — Creature — Human Shaman {2}{G}{G}, 2/1:
//
//	"When this creature enters, return target card from your graveyard
//	 to your hand.
//	 Eternalize {5}{G}{G}"
//
// Eternal Witness's body without the "you may": the trigger is
// mandatory, so there is no optional prompt, and with an empty graveyard
// it is removed for want of a target (CR 603.3d). Eternalize is the
// embalm.go body (CR 702.129a); the 4/4 Zombie token copy re-triggers
// as it enters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a52a06aa-56fc-4845-bbcd-a54f16369b34",
		Name:         "Timeless Witness",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetCardInGraveyard("target card in your graveyard", YouOwn()),
			Key:     "Timeless Witness — return target card to hand",
			Effect:  returnTargetCardToHand,
		}},
		Activated: []ActivatedAbility{Eternalize("{5}{G}{G}")},
	})
}
