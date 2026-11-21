package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mirror Room // Fractured Realm — Enchantment — Room (ADR 0103):
//
//	Mirror Room {2}{U}: "When you unlock this door, create a token
//	  that's a copy of target creature you control, except it's a
//	  Reflection in addition to its other creature types."
//	Fractured Realm {5}{U}{U}: "If a triggered ability of a permanent
//	  you control triggers, that ability triggers an additional time."
//
// The copy is CreateTokenCopy with the Zombie-style "in addition to"
// exception (addedSubtypeTypeLine). Fractured Realm is #752's trigger
// doubler over every permanent you control, itself included, so its own
// "when you unlock this door" ability is doubled too, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Room(effects.RoomSpec{
		OracleID:     "e7a1c21b-5966-4ce0-a925-0b3cb3c3db00",
		Name:         "Mirror Room // Fractured Realm",
		Completeness: effects.CompletenessFull,
		Left: effects.Door{Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenYouUnlockThisDoor(game.DoorLeft, "Mirror Room — copy target creature you control as a Reflection", mirrorRoomCopy),
				effects.TargetCreature("creature you control", effects.YouControl())),
		}},
		Right: effects.Door{TriggerDoublers: []game.TriggerDoubler{fracturedRealmDoubler()}},
	}))
}

func mirrorRoomCopy(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		// "Reflection" is a word of the door's text (CR 612.2, #3984).
		reflection := ctx.TextSubtype("Reflection")
		return effects.CreateTokenCopy{
			Controller: item.Controller,
			Copy:       t.ID,
			N:          1,
			Except:     func(tok *game.Card) { tok.TypeLine = effects.AddedSubtypeTypeLine(tok.TypeLine, reflection) },
		}.Apply(ctx)
	}
	return nil
}

// fracturedRealmDoubler is "a triggered ability of a permanent you
// control".
func fracturedRealmDoubler() game.TriggerDoubler {
	d := effects.DoublesAbilitiesOf(nil, effects.DoublesAbilitiesOfOptions{})
	d.Label = "Fractured Realm"
	return d
}
