package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dragon's Rage Channeler — Creature — Human Shaman {R}, 1/1 (EDHREC
// rank 1247):
//
//	"Whenever you cast a noncreature spell, surveil 1.
//	 Delirium — As long as there are four or more card types among
//	 cards in your graveyard, this creature gets +2/+2, has flying,
//	 and attacks each combat if able."
//
// The surveil trigger is the shared "noncreature spell cast by you"
// condition (b10NoncreatureSpellCastByYou), one Surveil 1 per cast.
//
// Delirium is three layer-6/7c effects gated on the same live
// predicate, re-evaluated every recompute — a CDA-flavoured static
// rather than a one-shot, so the bonus, the flying, and the "attacks
// each combat if able" all turn on and off together as the graveyard
// changes. "Cards in your graveyard" is scoped to the controller
// (drcDeliriumMet reads the controller's own Graveyard.Cards), unlike
// Tarmogoyf's table-wide count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0c016ccc-a341-4b76-87ba-69c639d2746d",
		Name:         "Dragon's Rage Channeler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, b10NoncreatureSpellCastByYou,
				"Dragon's Rage Channeler — surveil 1", Do(Surveil{N: 1})),
		},
		Static: []game.StaticAbility{
			{
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return target.InstanceID == source.InstanceID && drcDeliriumMet(g, source.Controller)
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power += 2
					c.Toughness += 2
				},
			},
			{
				Layer: game.Layer6Ability,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return target.InstanceID == source.InstanceID && drcDeliriumMet(g, source.Controller)
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					for _, kw := range c.Abilities {
						if kw == "flying" {
							return
						}
					}
					c.Abilities = append(c.Abilities, "flying")
				},
			},
			AttacksEachCombatWhere(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && drcDeliriumMet(g, source.Controller)
			}),
		},
	})
}

// drcDeliriumMet is CR delirium: four or more distinct card types
// among the cards in `controller`'s own graveyard. Types are read
// off the printed type line (IsCreature / IsLand / etc. read the same
// way for a graveyard card, which has no layer-applied
// characteristics of its own).
func drcDeliriumMet(g *game.Game, controller uuid.UUID) bool {
	p := g.PlayerByIDForEffect(controller)
	if p == nil || p.Graveyard == nil {
		return false
	}
	seen := map[string]bool{}
	for _, c := range p.Graveyard.Cards {
		if c.IsArtifact() {
			seen["artifact"] = true
		}
		if c.IsCreature() {
			seen["creature"] = true
		}
		if c.IsEnchantment() {
			seen["enchantment"] = true
		}
		if c.IsInstant() {
			seen["instant"] = true
		}
		if c.IsLand() {
			seen["land"] = true
		}
		if c.IsPlaneswalker() {
			seen["planeswalker"] = true
		}
		if c.IsSorcery() {
			seen["sorcery"] = true
		}
		if c.IsBattle() {
			seen["battle"] = true
		}
		if len(seen) >= 4 {
			return true
		}
	}
	return len(seen) >= 4
}
