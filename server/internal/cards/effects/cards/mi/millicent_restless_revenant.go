package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Millicent, Restless Revenant — Legendary Creature — Spirit Soldier
// {5}{W}{U}, 4/4 (Innistrad: Crimson Vow Commander):
//
//	"Affinity for Spirits (This spell costs {1} less to cast for each
//	 Spirit you control.)
//	 Flying
//	 Whenever Millicent or another nontoken Spirit you control dies or deals
//	 combat damage to a player, create a 1/1 white Spirit creature token
//	 with flying."
//
// Affinity is the shared self cost modifier. The trigger is one registered
// ability watching two events. A death reads the creature as it last
// existed on the battlefield (CR 603.10a: its Spirit type and controller),
// and Millicent counts as a Spirit whether or not she is a token; a
// damage event reads the dealing creature live. Each Spirit that deals
// combat damage, or dies, is its own trigger (CR 603.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "71766a05-62a5-4839-b9bf-f5dcde710251",
		Name:            "Millicent, Restless Revenant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for Spirits", effects.HasSubtype("Spirit")),
		},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventLTB, game.EventDealDamage}, millicentCondition,
				"Millicent — create a 1/1 white Spirit creature token with flying",
				effects.Do(effects.VowB12SpiritToken())),
		},
	})
}

func millicentCondition(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	switch ev.Kind {
	case game.EventLTB:
		dead, ok := effects.DiedCreature(ev, g)
		if !ok || effects.LeftUnderControlOf(ev, dead) != source.Controller || !effects.LeftAsSubtype(ev, dead, source.TextSubtype("Spirit")) {
			return false
		}
		return dead.InstanceID == source.InstanceID || !dead.IsToken()
	case game.EventDealDamage:
		if !effects.CombatDamageToPlayerBy(ev, source.Controller, g) {
			return false
		}
		dealer, ok := g.LookupCardForEffect(ev.Source)
		if !ok || !dealer.HasSubtype(source.TextSubtype("Spirit")) {
			return false
		}
		return dealer.InstanceID == source.InstanceID || !dealer.IsToken()
	}
	return false
}
