package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dire Flail // Dire Blunderbuss — a transforming Equipment (#2709,
// ADR 0137):
//
//	Dire Flail — Artifact — Equipment {R}
//	  "Equipped creature gets +2/+0.
//	   Equip {1}
//	   Craft with artifact {3}{R}{R}"
//	Dire Blunderbuss — Artifact — Equipment
//	  "Equipped creature gets +3/+0 and has "Whenever this creature
//	   attacks, you may sacrifice an artifact other than Dire
//	   Blunderbuss. When you do, this creature deals damage equal to its
//	   power to target creature."
//	   Equip {1}"
//
// The Blunderbuss grants its creature an ADR 0093 bundle, so the attack
// trigger is the creature's own: its controller controls it, and an
// effect that takes the creature's abilities takes it. "Dire
// Blunderbuss" in the granted text is the Equipment that granted it,
// which the trigger's item names (StackItem.GrantedBy, stamped for a
// granted triggered row since #2709): two Blunderbusses give two
// triggers, and each may sacrifice the other.
//
// The sacrifice is the parent's "you may" (MayChoice, then the
// sacrifice prompt), and the damage is a CR 603.12 reflexive trigger
// created only once an artifact has really gone, targeting as it goes
// on the stack (the ruling). The creature deals the damage, at its
// power as the reflexive trigger resolves, or as it last existed if it
// has left (the rulings, CR 608.2h), so its lifelink and deathtouch
// apply.
//
// No simplification.
const direFlailOracleID = "5a043512-fa56-4361-bb92-0a0f229f34d3"

// direBlunderbussGrant is the bundle the Blunderbuss gives its creature.
// An on-disk identity: never renamed.
const direBlunderbussGrant = "dire-blunderbuss/sacrifice-artifact"

func init() {
	Register(Spec{
		OracleID:     direFlailOracleID,
		Name:         "Dire Flail",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 0)},
		Activated: []ActivatedAbility{
			EquipAbility("{1}"),
			Craft("Craft with artifact {3}{R}{R}", "{3}{R}{R}", CraftWith("artifact")),
		},
	})

	Register(Spec{
		OracleID:     direFlailOracleID + "#1",
		Name:         "Dire Blunderbuss",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(3, 0),
			GrantAbilitiesToAttached(direBlunderbussGrant),
		},
		Grants: []AbilityGrant{{
			Key: direBlunderbussGrant,
			Triggered: []game.TriggeredAbility{
				WheneverThisAttacks("Dire Blunderbuss — you may sacrifice another artifact to deal damage", direBlunderbussAttacks),
			},
			Text: "Whenever this creature attacks, you may sacrifice an artifact other than Dire Blunderbuss. When you do, this creature deals damage equal to its power to target creature.",
		}},
		Activated: []ActivatedAbility{
			EquipAbility("{1}"),
		},
	})
}

// direBlunderbussAttacks is the granted trigger's resolution: the
// optional sacrifice, and the reflexive damage once it happened.
func direBlunderbussAttacks(g *game.Game, item *game.StackItem) error {
	exclude := direBlunderbussGrantors(g, item)
	eligible := func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsArtifact() && !exclude[c.InstanceID]
	}
	if countControlled(g, item.Controller, func(c game.Card) bool { return eligible(g, item.Controller, c) }) == 0 {
		return nil
	}
	return MayChoice{
		Question: "Dire Blunderbuss — sacrifice an artifact other than Dire Blunderbuss to deal damage to target creature?",
		YesLabel: "Sacrifice an artifact",
		NoLabel:  "Don't",
		OnYes: func(ctx *Context) error {
			return ctx.Game.PlayerSacrificesThenForEffect(
				item.SourceCardID, item.Controller,
				sacrificeSpec("an artifact other than Dire Blunderbuss", eligible),
				"Dire Blunderbuss — sacrifice an artifact",
				1,
				func(g *game.Game, sacrificed game.PromptedSacrifices) error {
					if sacrificed.Count() == 0 {
						return nil
					}
					return WhenYouDo("Dire Blunderbuss — this creature deals damage equal to its power to target creature",
						direBlunderbussDamageBody).Apply(NewContext(g, item))
				})
		},
	}.Apply(NewContext(g, item))
}

// direBlunderbussGrantors is the "Dire Blunderbuss" the granted text
// names: the Equipment that granted this trigger. An item restored from
// a file written before triggers carried their grantor names none, and
// then every Blunderbuss granting the bundle to the creature is read as
// it, which is the one reading that never lets a trigger sacrifice its
// own grantor.
func direBlunderbussGrantors(g *game.Game, item *game.StackItem) map[uuid.UUID]bool {
	if item.GrantedBy != uuid.Nil {
		return map[uuid.UUID]bool{item.GrantedBy: true}
	}
	out := map[uuid.UUID]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsAttachedTo(item.SourceCardID) && c.Name == "Dire Blunderbuss" {
			out[c.InstanceID] = true
		}
	}
	return out
}

// direBlunderbussDamage is the reflexive trigger: the creature (the
// trigger's source) deals damage equal to its power, live or as it last
// existed, to the target creature.
func direBlunderbussDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: info.Power}.Apply(ctx)
	}
	return nil
}
