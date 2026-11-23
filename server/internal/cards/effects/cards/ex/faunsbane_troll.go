package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "b707c131-de13-4d4d-839d-b9f47d62f090",
		Name:         "Faunsbane Troll",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Faunsbane Troll — create a Monster Role token attached to it",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.B09SourceStillOnBattlefield(g, item) {
						return nil
					}
					return effects.CreateRoleToken{Role: effects.RoleMonster, Host: item.SourceCardID}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{1}, Sacrifice an Aura attached to this creature: This creature fights target creature you don't control. If that creature would die this turn, exile it instead. Activate only as a sorcery.",
			Cost: effects.Plus(effects.ManaCost("{1}"), game.AbilityCost{
				SacrificeOther: effects.AttachedToThis("an Aura attached to this creature", effects.HasSubtype("Aura")),
			}),
			Targets:      effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (effects.ExileIfItWouldDieThisTurn{Target: id}).Apply(ctx); err != nil {
					return err
				}
				if !effects.B09SourceStillOnBattlefield(g, item) {
					return nil
				}
				return effects.B10Fight(ctx, item.SourceCardID, id)
			},
		}},
	})
}
