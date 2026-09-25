package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Caduceus, Staff of Hermes — Legendary Artifact — Equipment {2}{W}:
//
//	"Equipped creature has lifelink.
//	 As long as you have 30 or more life, equipped creature gets
//	 +5/+5 and has indestructible and 'Prevent all damage that would
//	 be dealt to this creature.'
//	 Equip {W}{W}"
//
// The life-total condition is life_gated_statics.go's family
// (Serra Ascendant, Righteous Valkyrie), which supplies
// LifeGatedPump / LifeGatedKeyword and DependsOnLifeTotal so the
// layer cache invalidates the instant the life total crosses 30 —
// but every constructor there answers a fixed shape ("this creature",
// "creatures you control"). "Equipped creature, conditioned on life"
// is neither, so the AppliesTo is built by hand from the same two
// pieces those constructors compose: AttachedToSource and
// controllerLifeAtLeast.
//
// "Prevent all damage that would be dealt to this creature" is the
// THIRD granted ability, and it is not a keyword — no canonical
// keyword token means "immune to all damage". It is a Torbran-shaped
// permanent replacement effect (prevention.go's PreventNextDamage /
// PreventAllCombatDamageThisTurn are one-shot resolution-time
// registrations for a spell; this is a standing Spec.Replacements
// entry instead, gated by the same condition as the two statics), so
// all three grants light up and go dark together as life crosses 30
// in either direction.
//
// No simplification.
func init() {
	caduceusCondition := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return AttachedToSource(target, g, source) && controllerLifeAtLeast(g, source, 30)
	}
	Register(Spec{
		OracleID:     "7cea5b12-9483-4142-9efa-735305581e73",
		Name:         "Caduceus, Staff of Hermes",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			GrantToAttached("lifelink"),
			LifeGatedPump(caduceusCondition, 5, 5),
			LifeGatedKeyword(caduceusCondition, "indestructible"),
		},
		Replacements: []game.ReplacementEffect{
			{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
					if ev.Kind != game.RepEventDamage {
						return false
					}
					if !src.IsAttachedTo(ev.DamageTarget) {
						return false
					}
					return controllerLifeAtLeast(g, src, 30)
				},
				Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
					ev.Cancel()
					return nil
				},
				Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
					return src.Controller
				},
				Label: "Caduceus, Staff of Hermes — prevent all damage to equipped creature",
			},
		},
		Activated: []ActivatedAbility{
			EquipAbility("{W}{W}"),
		},
	})
}
