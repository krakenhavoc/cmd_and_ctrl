package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goblin Rabblemaster — Creature — Goblin Warrior, {2}{R}, 2/2:
//
//	"Other Goblin creatures you control attack each combat if able.
//	 At the beginning of combat on your turn, create a 1/1 red
//	 Goblin creature token with haste.
//	 Whenever this creature attacks, it gets +1/+0 until end of turn
//	 for each other attacking Goblin."
//
// #1599, the first card built on #1595's CR 508.1d machinery rather
// than just cleared of a stale caveat. The "other Goblin creatures you
// control" clause is TribeFilter's own shape — Others + YoursOnly, so
// Rabblemaster itself is never required to attack by this line, only
// the horde it makes.
//
// The pump reads the board at RESOLUTION, not continuously: "for each
// other attacking Goblin" counts every attacking Goblin creature on
// the battlefield, any controller's (the printed line has no "you
// control"), excluding Rabblemaster itself, and the +1/+0 lands as a
// one-time until-end-of-turn boost rather than a live-recomputed
// static — exactly like exalted's pump (game/exalted.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "24661b81-6bee-4ad8-b3ab-53cb65005c51",
		Name:         "Goblin Rabblemaster",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			AttacksEachCombatWhere(TribeFilter{Tribes: []string{"Goblin"}, Others: true, YoursOnly: true}.Matches),
		},
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Goblin Rabblemaster — create a 1/1 red Goblin with haste",
				Do(CreateToken{Template: TokenCard("1/1 red Goblin with haste"), N: 1})),
			On(game.EventAttack, ThisAttacked,
				"Goblin Rabblemaster — +1/+0 until end of turn for each other attacking Goblin",
				rabblemasterPumpForOtherAttackingGoblins),
		},
	})
}

// rabblemasterPumpForOtherAttackingGoblins counts every attacking
// Goblin creature on the battlefield other than Rabblemaster itself —
// any controller's, per the printed line — and pumps Rabblemaster
// +1/+0 until end of turn for each one.
func rabblemasterPumpForOtherAttackingGoblins(g *game.Game, item *game.StackItem) error {
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID == item.SourceCardID {
			continue
		}
		if c.IsCreature() && c.HasSubtype("Goblin") && c.AttackingTarget != uuid.Nil {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return BoostUntilEOT{
		Target: item.SourceCardID,
		Power:  n,
		Label:  "Goblin Rabblemaster — +1/+0 for each other attacking Goblin",
	}.Apply(NewContext(g, item))
}
