package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Defiant Thundermaw — Creature — Dragon 4/4, the BACK FACE of
// Invasion of Tarkir (oracle 5c7f02ad…, face 1):
//
//	"Flying, trample
//	 Whenever a Dragon you control attacks, it deals 2 damage to any
//	 target."
//
// Registered under "<oracle_id>#1" — see deluge_of_the_dead.go for
// why the key is composite.
//
// "A Dragon you control", not "another": the Thundermaw triggers off
// its own attack, which is Utvara Hellkite's reading and the same
// helper (b14DragonYouControlAttacked). EventAttack is emitted once
// per attacking creature, so a three-Dragon swing is three triggers,
// each with its own target.
//
// "IT deals 2 damage" — the ATTACKING Dragon is the source, not the
// Thundermaw. That matters for lifelink and deathtouch on the
// attacker and for who a damage-reflection effect points back at, so
// the attacker's id is read off item.Trigger.Event.CardID at
// resolution rather than defaulting to item.SourceCardID the way most
// damage effects do.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        invasionOfTarkirOracleID + "#1",
		Name:            "Defiant Thundermaw",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b14DragonYouControlAttacked(ev, source, g)
			},
			Targets: TargetAny(),
			Key:     "Defiant Thundermaw — the attacking Dragon deals 2 damage to any target",
			Effect:  thundermawDamage,
		}},
	})
}

// thundermawDamage deals the attacking Dragon's 2 damage to the
// chosen target, skipping a target that has stopped being legal
// (CR 608.2b) and an attacker that has left the battlefield — a
// Dragon killed in response deals nothing, because "it deals" names
// a source that is no longer there. The attacker is read off
// item.Trigger.Event.CardID rather than a closure (ADR 0041 P9).
//
// Caller holds g.mu.
func thundermawDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
		return nil
	}
	attacker := item.Trigger.Event.CardID
	if z := g.FindCardZoneForEffect(attacker); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return DealDamage{
		Source: attacker,
		Target: item.Targets[0].ID,
		Amount: 2,
	}.Apply(ctx)
}
