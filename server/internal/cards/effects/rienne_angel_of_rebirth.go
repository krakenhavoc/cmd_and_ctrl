package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rienne, Angel of Rebirth — Legendary Creature — Angel {2}{R}{G}{W}, 5/4:
//
//	"Flying
//	 Other multicolored creatures you control get +1/+0.
//	 Whenever another multicolored creature you control dies, return it to
//	 its owner's hand at the beginning of the next end step."
//
// The anthem is a layer 7c static over the other creatures its controller
// controls that have two or more colours, read after the colour layer.
//
// The trigger reads the dead creature as it last existed on the battlefield
// (CR 603.10a): the controller and the colours it had then, so a creature
// that was multicolored only through an effect counts and one that is
// monocolored in the graveyard but was gold in play counts too. The return is
// a CR 603.7 delayed trigger carrying the card (Liesa's shape): at the next
// end step, any player's, it comes back to hand only if the card is still in a
// graveyard, so one exiled or reanimated meanwhile is a new object and is left
// alone. A token ceases to exist and returns nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1ef2415f-f88f-41fb-b084-ea52631a60c1",
		Name:            "Rienne, Angel of Rebirth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			b16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					target.InstanceID != source.InstanceID && len(target.EffectiveColors()) >= 2
			}, 1, 0),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == source.InstanceID {
					return false
				}
				dead, ok := diedCreature(ev, g)
				return ok && leftUnderControlOf(ev, dead) == source.Controller && leftMulticolored(ev, dead)
			}, "Rienne, Angel of Rebirth — return it to its owner's hand at the beginning of the next end step",
				func(g *game.Game, item *game.StackItem) error {
					return ScheduleDelayedTrigger{
						Label: "Rienne, Angel of Rebirth — return the creature card to its owner's hand",
						Cards: []uuid.UUID{item.Trigger.Event.CardID},
						Body:  returnListedGraveyardToHandBody,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
