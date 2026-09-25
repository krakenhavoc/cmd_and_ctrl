package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Greenwarden of Murasa — Creature — Elemental {4}{G}{G}, 5/4
// (EDHREC rank 3743):
//
//	"When this creature enters, you may return target card from your
//	 graveyard to your hand.
//	 When this creature dies, you may exile it. If you do, return
//	 target card from your graveyard to your hand."
//
// Eternal Witness twice, the second time paid for with its own
// corpse. Both triggers are optional and targeted — the controller
// is asked, then picks the card in the zone browser — and both
// return a card of any type from the controller's graveyard (Codex
// Shredder's clause), re-checked at resolution (CR 608.2b). With no
// card in the graveyard the trigger is removed (CR 603.3d) rather
// than prompting.
//
// The dies half reads "you may exile it. If you do" as the
// trigger's own yes/no: a "yes" means the Greenwarden is exiled from
// the graveyard on resolution and the chosen card comes back; a
// Greenwarden that has since left the graveyard — reanimated in
// response, or a commander sent to the command zone — cannot be
// exiled, so "if you do" fails and nothing returns. The Greenwarden
// is itself in the graveyard when the trigger goes on the stack and
// so is a legal pick; choosing it exiles it first, after which it is
// no longer in the graveyard and nothing returns — as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2facb1b3-4522-4610-a0ab-29ac53ca7fcd",
		Name:         "Greenwarden of Murasa",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:        []game.EventKind{game.EventETB},
				AppliesTo:      b06SelfETB,
				OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Greenwarden of Murasa — return a card from your graveyard to your hand?"},
				Targets:        TargetCardInGraveyard("target card from your graveyard", YouOwn()),
				Key:            "Greenwarden of Murasa — return a card from your graveyard to your hand",
				Effect:         b35ReturnChosenGraveyardCardToHand,
			},
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return cardDied(ev, source)
				},
				OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Greenwarden of Murasa — exile it from your graveyard to return a card to your hand?"},
				Targets:        TargetCardInGraveyard("target card from your graveyard", YouOwn()),
				Key:            "Greenwarden of Murasa — exile it and return a card from your graveyard to your hand",
				Effect:         b35ExileSelfFromGraveyardThenReturnChosen,
			},
		},
	})
}
