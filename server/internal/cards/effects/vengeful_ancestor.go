package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vengeful Ancestor — Creature — Spirit Dragon, {2}{R}{R}, 3/4:
//
//	"Flying
//	 Whenever this creature enters or attacks, goad target creature.
//	 (Until your next turn, that creature attacks each combat if
//	 able and attacks a player other than you if able.)
//	 Whenever a goaded creature attacks, it deals 1 damage to its
//	 controller."
//
// #1599: the first line is WhenThisEntersOrAttacks's shape — one
// ability, two trigger conditions (Sun Titan's pattern) — over
// GoadTarget (goad.go). The second line watches EventAttack for ANY
// creature — not only ones this card goaded — that is currently
// goaded (Card.GoadedBy set by anybody), and has THAT creature deal 1
// damage to its own controller: "it deals 1 damage to its controller"
// names the attacker as the source, not Vengeful Ancestor, so the
// attacker's own deathtouch or lifelink (if any) applies to the ping.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f3b3173b-f7ae-420e-84ec-ea61414674a0",
		Name:            "Vengeful Ancestor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB, game.EventAttack},
				AppliesTo: Self,
				Targets:   TargetCreature("target creature"),
				Key:       "Vengeful Ancestor — goad target creature",
				Effect:    GoadTarget,
			},
			{
				Watches:   []game.EventKind{game.EventAttack},
				AppliesTo: vengefulAncestorGoadedAttacker,
				Key:       "Vengeful Ancestor — the goaded attacker deals 1 damage to its controller",
				Effect:    vengefulAncestorDamageGoadedAttacker,
			},
		},
	})
}

// vengefulAncestorGoadedAttacker reports whether the attacking
// creature this EventAttack is about carries a goad marker right now
// — any goader's, not only this card's.
func vengefulAncestorGoadedAttacker(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.GoadedBy != uuid.Nil
}

// vengefulAncestorDamageGoadedAttacker reads the attacking creature
// back off the item's carried trigger event (ADR 0041 P9) and has it
// deal 1 damage to its own controller.
func vengefulAncestorDamageGoadedAttacker(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	return DealDamage{Source: id, Target: c.Controller, Amount: 1}.Apply(NewContext(g, item))
}
