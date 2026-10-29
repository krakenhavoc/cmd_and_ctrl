package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bilbo, Unexpected Adventurer — Legendary Creature — Halfling Rogue
// {3}{W}, 2/2:
//
//	"Bilbo can't be blocked by creatures with power 3 or greater.
//	 Whenever Bilbo deals combat damage to a player or battle, put up to
//	 one target nonland permanent card with mana value 3 or less from a
//	 graveyard onto the battlefield under its owner's control."
//
// The block rule is Old Fat Spider's (CantBeBlockedBy) with the other
// bound. The trigger is Sun Titan's reanimation from ANY graveyard: the
// target is chosen when the trigger goes on the stack, "up to one" so it
// may be declined, and the card comes back under its owner's control (the
// primitive's default), not Bilbo's. Damage to a battle is the same
// EventDealDamage as damage to a player (Beamtown Beatstick, #3172).
//
// No simplification. The Hobbit (#4454).
func init() {
	Register(Spec{
		OracleID:     "1f4e4750-93e9-4aa2-b105-9482f7b287ec",
		Name:         "Bilbo, Unexpected Adventurer",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), PowerGE(3), "creatures with power 3 or greater"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventDealDamage, thisDealtCombatDamageToAPlayerOrBattle,
					"Bilbo, Unexpected Adventurer — put a nonland permanent card with mana value 3 or less from a graveyard onto the battlefield",
					hobB8ReturnTargetGraveyardCardToBattlefield),
				TargetCardInGraveyard("up to one target nonland permanent card with mana value 3 or less in a graveyard",
					Nonland(), Permanent(), ManaValueLE(3)).WithCount(0, 1)),
		},
	})
}

// thisDealtCombatDamageToAPlayerOrBattle is "this creature deals combat
// damage to a player or battle".
func thisDealtCombatDamageToAPlayerOrBattle(ev game.Event, source *game.Card, c game.Characteristic, g *game.Game) bool {
	if ThisDealtCombatDamageToAPlayer(ev, source, c, g) {
		return true
	}
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || ev.Source != source.InstanceID {
		return false
	}
	target, ok := g.LookupCardForEffect(ev.Target)
	return ok && target.IsBattle()
}
