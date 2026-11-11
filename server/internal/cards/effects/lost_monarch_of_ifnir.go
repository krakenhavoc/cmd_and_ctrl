package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lost Monarch of Ifnir — Creature — Zombie Noble {3}{B}, 4/4:
//
//	"Afflict 3
//	 Other Zombies you control have afflict 3.
//	 At the beginning of your second main phase, if a player was dealt
//	 combat damage by a Zombie this turn, mill three cards, then you may
//	 return a creature card from your graveyard to your hand."
//
// Its own afflict is the Afflict row; the grant is a bundle holding that same
// row, given to the other Zombies you control (TribalAbilityGrant), so each
// has afflict 3 as an ability it can lose. The intervening "if" (CR 603.4)
// reads this turn's combat damage events for a Zombie source as it dealt
// the damage and is checked again as the trigger resolves. The mill is the
// effect; the return is a "you may" pick over the graveyard after the mill,
// so a creature card the mill just put there can be returned.
//
// No simplification.
const lostMonarchAfflictGrant = "lost-monarch-of-ifnir/afflict-3"

func init() {
	Register(Spec{
		OracleID:     "626a979f-bed7-4cb1-9ab7-c3f2bbe37781",
		Name:         "Lost Monarch of Ifnir",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       lostMonarchAfflictGrant,
			Triggered: []game.TriggeredAbility{Afflict(3, "Lost Monarch of Ifnir")},
			Text:      "Afflict 3",
		}},
		Static: []game.StaticAbility{
			TribalAbilityGrant(TribeFilter{Tribes: []string{"Zombie"}, Others: true, YoursOnly: true}, lostMonarchAfflictGrant),
		},
		Triggered: []game.TriggeredAbility{
			Afflict(3, "Lost Monarch of Ifnir"),
			cmdrB24LostMonarchSecondMain(),
		},
	})
}

// cmdrB24LostMonarchSecondMain is the second-main trigger with its
// intervening "if" (CR 603.4): the condition gates the trigger and is asked
// again as it resolves.
func cmdrB24LostMonarchSecondMain() game.TriggeredAbility {
	t := AtYourPostcombatMain("Lost Monarch of Ifnir — mill three cards, then you may return a creature card from your graveyard to your hand",
		func(g *game.Game, item *game.StackItem) error {
			src, ok := g.LookupCardForEffect(item.SourceCardID)
			if !ok || !cmdrB24ZombieDealtCombatDamageToAPlayerThisTurn(g, &src) {
				return nil
			}
			return cmdrB24MillThenMayReturnCreature(g, item)
		})
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return (when == nil || when(ev, source, lki, g)) && cmdrB24ZombieDealtCombatDamageToAPlayerThisTurn(g, source)
	}
	return t
}
