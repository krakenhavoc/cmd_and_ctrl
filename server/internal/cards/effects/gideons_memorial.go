package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gideon's Memorial — Legendary Artifact {1}{W} (Reality Fracture):
//
//	"Creature tokens you control get +1/+0 and have vigilance.
//	 {T}: Add one mana of any color. Spend this mana only to cast a
//	 planeswalker spell.
//	 {1}{W}, Discard this card: It deals 4 damage to target attacking or
//	 blocking creature."
//
// The anthem is Intangible Virtue's token filter in layers 7c and 6. The
// mana is restricted to casting (ManaRestrictCast) a planeswalker
// (ManaRestrictType), so it cannot pay for an ability. The last line is
// the channel shape: an activated ability that works only from the hand
// (Eiganjo, Seat of the Empire), with no discount.
func init() {
	Register(Spec{
		OracleID:     "c8f08704-7330-4fa9-b456-2e96aa252a2c",
		Name:         "Gideon's Memorial",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: b05CreatureTokenYouControl,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
				},
			},
			{
				Layer:     game.Layer6Ability,
				AppliesTo: b05CreatureTokenYouControl,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !eotHasAbility(c.Abilities, "vigilance") {
						c.Abilities = append(c.Abilities, "vigilance")
					}
				},
			},
		},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			Produced:     "{W|U|B|R|G}",
			Restrictions: []string{ManaRestrictCast, ManaRestrictType("Planeswalker")},
			Label:        "Add one mana of any color. Spend this mana only to cast a planeswalker spell",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}, Discard this card: It deals 4 damage to target attacking or blocking creature.",
			Cost:    game.AbilityCost{Mana: "{1}{W}", DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
			Purpose: ForTargets(DamageToTarget(0, 4)),
			Effect:  rfDamageFirstLegalTarget(4),
		}},
	})
}
