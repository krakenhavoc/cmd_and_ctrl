package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Medusa, Inhuman Queen — Legendary Creature — Inhuman Noble Hero {2}{G},
// 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"Reach, vigilance
//	 Whenever a player casts a noncreature spell, put a +1/+1 counter on
//	 Medusa."
//
// Reach and vigilance ride PrintedKeywords. The trigger watches every
// player's casts, Medusa's controller's included ("a player"), and the
// spell is read as it is on the stack: an artifact creature spell is a
// creature spell and does not count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "91accef6-378f-4777-ad58-fa1af7893d8a",
		Name:            "Medusa, Inhuman Queen",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, aPlayerCastsANoncreatureSpell,
				"Medusa, Inhuman Queen — a +1/+1 counter on Medusa", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}

// aPlayerCastsANoncreatureSpell is "a player casts a noncreature spell":
// any caster, the spell not a creature.
func aPlayerCastsANoncreatureSpell(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !spell.IsCreature()
}
