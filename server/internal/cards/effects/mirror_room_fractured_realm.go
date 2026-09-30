package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Room(RoomSpec{
		OracleID:     "e7a1c21b-5966-4ce0-a925-0b3cb3c3db00",
		Name:         "Mirror Room // Fractured Realm",
		Completeness: CompletenessFull,
		Left: Door{Triggered: []game.TriggeredAbility{
			Targeting(WhenYouUnlockThisDoor(game.DoorLeft, "Mirror Room — copy target creature you control as a Reflection", mirrorRoomCopy),
				TargetCreature("creature you control", YouControl())),
		}},
		Right: Door{TriggerDoublers: []game.TriggerDoubler{fracturedRealmDoubler()}},
	}))
}

func mirrorRoomCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return CreateTokenCopy{
			Controller: item.Controller,
			Copy:       t.ID,
			N:          1,
			Except:     func(tok *game.Card) { tok.TypeLine = addedSubtypeTypeLine(tok.TypeLine, "Reflection") },
		}.Apply(ctx)
	}
	return nil
}

// fracturedRealmDoubler is "a triggered ability of a permanent you
// control".
func fracturedRealmDoubler() game.TriggerDoubler {
	d := DoublesAbilitiesOf(nil, DoublesAbilitiesOfOptions{})
	d.Label = "Fractured Realm"
	return d
}
