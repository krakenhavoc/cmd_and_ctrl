package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gaea's Liege — Creature — Avatar {3}{G}{G}{G}, */*:
//
//	"As long as Gaea's Liege isn't attacking, its power and toughness are
//	 each equal to the number of Forests you control. As long as Gaea's
//	 Liege is attacking, its power and toughness are each equal to the
//	 number of Forests defending player controls.
//	 {T}: Target land becomes a Forest until this creature leaves the
//	 battlefield."
//
// The size is one characteristic-defining ability (CR 604.3) in layer 7a:
// it counts Forests by their effective subtype, the controller's while it
// is not attacking and the defending player's while it is (CR 508.5: the
// player it is attacking, or the controller of the planeswalker or the
// protector of the battle it attacks). It declares DependsOnAttackingStatus
// so the layer cache follows it into and out of combat.
//
// The ability is ADR 0109 §1's (#1881) CR 305.7 type set for as long as
// the Liege stays on the battlefield (CR 611.2b): the land is a Forest
// (other subtypes stay, CR 205.1a), loses its rules-text abilities and
// taps for {G} (CR 305.6). Each one makes the Liege bigger, and turning
// an opponent's lands into Forests makes it bigger when it attacks them.
// A Liege gone before the ability resolves changes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8d134a60-e1e5-4163-8bdc-36af91567185",
		Name:         "Gaea's Liege",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:                    game.Layer7PT,
			SubLayer:                 game.SubLayer7A_CDA,
			DependsOnAttackingStatus: true,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				whose := source.Controller
				if source.AttackingTarget != uuid.Nil {
					whose = g.DefendingPlayerForAttackerForEffect(source.InstanceID)
				}
				n := b08LandsWithSubtypeControlled(g, whose, "Forest")
				c.Power = n
				c.Toughness = n
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target land becomes a Forest until this creature leaves the battlefield.",
			Cost:    TapCost(),
			Targets: TargetPermanent("target land", Land()),
			Effect:  TargetLandBecomesWhileSourceRemains("Gaea's Liege", "Forest"),
		}},
	})
}
