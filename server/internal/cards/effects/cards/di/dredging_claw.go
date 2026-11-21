package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dredging Claw — Artifact — Equipment {2} (The Brothers' War, #5482):
//
//	"Equipped creature gets +1/+0 and has menace.
//	 Whenever a creature enters from your graveyard, you may attach this
//	 Equipment to it.
//	 Equip {1}{B}"
//
// The trigger reads where the creature entered from (EventETB.EnteredFrom
// and EnteredFromOwner, as Treacherous Pit-Dweller does): any creature, whoever
// controls it, that came out of THE CONTROLLER'S graveyard, so a unearthed
// creature qualifies and a cast one does not. The attach is Doc Ock's
// Tentacles' optional trigger, which does nothing if the creature or the
// Claw left before it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3808afb9-962f-48ba-94c5-6b3eae7053d0",
		Name:         "Dredging Claw",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 0),
			effects.GrantToAttached("menace"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.S82Bro13CreatureEnteredFromYourGraveyard,
			Key:       "Dredging Claw — attach to the creature that entered from your graveyard",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return g.AttachSourceForEffect(item, game.TargetRef{Kind: game.TargetCard, ID: item.Trigger.Event.CardID})
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Dredging Claw — attach to it?"},
		}},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}{B}")},
	})
}
