package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Faunsbane Troll — Creature — Troll {3}{G}{G}:
//
//	"When this creature enters, create a Monster Role token attached to
//	 it. (Enchanted creature gets +1/+1 and has trample.)
//	 {1}, Sacrifice an Aura attached to this creature: This creature
//	 fights target creature you don't control. If that creature would
//	 die this turn, exile it instead. Activate only as a sorcery."
//
// The cost is the #1945 seam: a sacrifice clause restricted to
// permanents attached to the paying creature (AttachedToThis), judged
// by one answer for the payment, the picker and the bot's enumerator.
// The Aura is any Aura attached to it, not only the Monster Role.
//
// The fight and the exile replacement are ADR 0108's. The replacement is
// the ability's, so it is registered whether or not the fight happens
// (the target left, or the Troll did).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "b707c131-de13-4d4d-839d-b9f47d62f090",
		Name:         "Faunsbane Troll",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Faunsbane Troll — create a Monster Role token attached to it",
				func(g *game.Game, item *game.StackItem) error {
					if !b09SourceStillOnBattlefield(g, item) {
						return nil
					}
					return CreateRoleToken{Role: RoleMonster, Host: item.SourceCardID}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label: "{1}, Sacrifice an Aura attached to this creature: This creature fights target creature you don't control. If that creature would die this turn, exile it instead. Activate only as a sorcery.",
			Cost: Plus(ManaCost("{1}"), game.AbilityCost{
				SacrificeOther: AttachedToThis("an Aura attached to this creature", HasSubtype("Aura")),
			}),
			Targets:      TargetCreature("target creature you don't control", OpponentControls()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (ExileIfItWouldDieThisTurn{Target: id}).Apply(ctx); err != nil {
					return err
				}
				if !b09SourceStillOnBattlefield(g, item) {
					return nil
				}
				return b10Fight(ctx, item.SourceCardID, id)
			},
		}},
	})
}
