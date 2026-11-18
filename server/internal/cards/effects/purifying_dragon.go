package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Purifying Dragon — Creature — Dragon {3}{R}{R}, 4/3 (Midnight Hunt, #5497):
//
//	"Flying
//	 Whenever this creature attacks, it deals 1 damage to target creature
//	 defending player controls. If that creature is a Zombie, this creature
//	 deals 2 damage to it instead."
//
// Coveted Peacock's target (TargetCreatureDefendingPlayerControls reads the
// defending player off this attack). The Zombie check is made as the trigger
// resolves, against the creature as it then is (CR 608.2h), through the
// layered view, so a Changeling or a creature that has been made a Zombie
// takes 2. A target that is no longer legal takes nothing (CR 608.2b). The
// damage is dealt by the Dragon, which a Dragon that has left still does with
// its last-known information.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a45762af-4aac-40b3-a983-ea441b04c5ee",
		Name:            "Purifying Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches:     []game.EventKind{game.EventAttack},
			AppliesTo:   ThisAttacked,
			TargetsFrom: TargetCreatureDefendingPlayerControls,
			Key:         "Purifying Dragon — 1 damage to target creature defending player controls, 2 if it is a Zombie",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				victim, ok := g.LookupCardForEffect(id)
				if !ok {
					return nil
				}
				n := 1
				if victim.HasSubtype(ctx.TextSubtype("Zombie")) {
					n = 2
				}
				return DealDamage{Source: item.SourceCardID, Target: id, Amount: n}.Apply(ctx)
			},
		}},
	})
}
