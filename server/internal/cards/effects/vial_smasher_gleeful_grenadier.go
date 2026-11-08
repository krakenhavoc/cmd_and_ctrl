package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vial Smasher, Gleeful Grenadier — Legendary Creature — Goblin
// Mercenary {B}{R}, 3/2 (Outlaws of Thunder Junction, #4468):
//
//	"Whenever another outlaw you control enters, Vial Smasher deals 1
//	 damage to target opponent. (Assassins, Mercenaries, Pirates,
//	 Rogues, and Warlocks are outlaws.)"
//
// An outlaw is a permanent with one of the five types (CR 700.12), read
// off the permanent as it enters, so a changeling counts and a Mercenary
// token counts. The five types are words of Vial Smasher's text (CR
// 612.2). "Another" is object identity: Vial Smasher, a Mercenary
// itself, does not trigger on its own entry. The opponent is chosen as
// the trigger goes on the stack (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a66d8dc6-ced4-4b24-b0c2-3557b2a851be",
		Name:         "Vial Smasher, Gleeful Grenadier",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventETB, vialSmasherAnotherOutlawEntered,
				"Vial Smasher, Gleeful Grenadier — 1 damage to target opponent", DealDamageToTheTarget(1)),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

// vialSmasherAnotherOutlawEntered is "another outlaw you control enters".
func vialSmasherAnotherOutlawEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	if !ok {
		return false
	}
	for _, t := range []string{"Assassin", "Mercenary", "Pirate", "Rogue", "Warlock"} {
		if c.HasSubtype(source.TextSubtype(t)) {
			return true
		}
	}
	return false
}
