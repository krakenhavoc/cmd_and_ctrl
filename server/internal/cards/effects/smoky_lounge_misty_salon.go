package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Smoky Lounge // Misty Salon — Enchantment — Room (ADR 0103):
//
//	Smoky Lounge {2}{R}: "At the beginning of your first main phase,
//	  add {R}{R}. Spend this mana only to cast Room spells and unlock
//	  doors."
//	Misty Salon {3}{U}: "When you unlock this door, create an X/X blue
//	  Spirit creature token with flying, where X is the number of
//	  unlocked doors among Rooms you control."
//
// Smoky Lounge's ability is an ordinary triggered ability (it triggers
// on a phase, not on mana being produced, so it is not a mana ability
// and uses the stack). The mana carries ManaRestrictAnyOf(cast a Room,
// unlock) through AddMana.Restrictions. It empties with the pool at the
// end of the main phase (CR 106.4), as printed.
//
// "First main phase" is the precombat main phase; the engine has no
// extra-combat turn structure that would put another main phase first.
//
// Misty Salon reads X when the ability resolves, so the new token's
// size counts the door just unlocked. No simplification.
func init() {
	Register(Room(RoomSpec{
		OracleID:     "4f70b0ee-c24c-45fa-b878-9ba69266344f",
		Name:         "Smoky Lounge // Misty Salon",
		Completeness: CompletenessFull,
		Left: Door{Triggered: []game.TriggeredAbility{
			AtYourPrecombatMain("Smoky Lounge — add {R}{R} for Room spells and unlocking doors", Do(AddMana{
				Produced: "{R}{R}",
				Restrictions: []string{game.ManaRestrictAnyOf(
					[]string{game.ManaRestrictCast, game.ManaRestrictSubtype("Room")},
					[]string{game.ManaRestrictUnlock},
				)},
			})),
		}},
		Right: Door{Triggered: []game.TriggeredAbility{
			WhenYouUnlockThisDoor(game.DoorRight, "Misty Salon — create an X/X blue Spirit with flying", mistySalonSpirit),
		}},
	}))
}

func mistySalonSpirit(g *game.Game, item *game.StackItem) error {
	x := UnlockedDoorsYouControl(g, item.Controller)
	return CreateToken{
		Controller: item.Controller,
		Template: game.Card{
			Name: "Spirit", TypeLine: "Token Creature — Spirit",
			Power: x, Toughness: x, Colors: []string{"U"}, Keywords: []string{"flying"},
		},
		N: 1,
	}.Apply(NewContext(g, item))
}
