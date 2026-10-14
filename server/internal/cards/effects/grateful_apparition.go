package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grateful Apparition — Creature — Spirit {1}{W}, 1/1:
//
//	"Flying
//	 Whenever this creature deals combat damage to a player or
//	 planeswalker, proliferate."
//
// Bloated Contaminator's trigger widened to a planeswalker: the damage
// event's target is either a player or a planeswalker permanent. The
// controller chooses what to proliferate (#2525).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f1b3e0fc-1dc3-4eba-bd54-557f036f5ea9",
		Name:            "Grateful Apparition",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, ThisDealtCombatDamageToAPlayerOrPlaneswalker,
				"Grateful Apparition — proliferate", Do(Proliferate{})),
		},
	})
}

// ThisDealtCombatDamageToAPlayerOrPlaneswalker — the source dealt combat
// damage to a player or to a planeswalker.
func ThisDealtCombatDamageToAPlayerOrPlaneswalker(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || ev.Source != source.InstanceID {
		return false
	}
	if g.PlayerByIDForEffect(ev.Target) != nil {
		return true
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	return ok && c.IsPlaneswalker()
}
