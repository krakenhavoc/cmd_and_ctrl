package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gideon Blackblade — Legendary Planeswalker — Gideon {1}{W}{W},
// starting loyalty 4:
//
//	"During your turn, Gideon Blackblade is a 4/4 Human Soldier creature
//	 with indestructible that's still a planeswalker.
//	 Prevent all damage that would be dealt to Gideon Blackblade during
//	 your turn.
//	 +1: Up to one other target creature you control gains your choice
//	 of vigilance, lifelink, or indestructible until end of turn.
//	 −6: Exile target nonland permanent."
//
// THE FIRST STATIC is three layered effects on the same condition, "it
// is its controller's turn", read at every layer pass: layer 4 adds the
// Creature type and the Human and Soldier subtypes (he keeps Planeswalker
// and Gideon), layer 6 adds indestructible, layer 7b sets the base power
// and toughness. They disappear the moment the turn passes, and a
// permanent that is a creature only on its controller's turn can attack
// on it once it has been under their control since the turn began
// (CR 302.6). THE SECOND is a standing CR 615 prevention that stops
// every damage event to him while it is that turn, the Marble Priest
// shape; damage that can't be prevented (CR 615.12) still removes loyalty
// and is marked on him, and both state-based actions apply (#2046, ADR
// 0032 amendment of 2026-10-07). THE +1 asks for the keyword as it
// resolves, on "up to one other target creature you control"; with no
// target it does nothing. THE −6 is a plain targeted exile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "813c19f5-3580-488d-9eee-c7a563def532",
		Name:            "Gideon Blackblade",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Static: []game.StaticAbility{
			{
				Layer: game.Layer4Type,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return gideonBlackbladeIsACreature(target, g, source)
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !slices.Contains(c.Types, "Creature") {
						c.Types = append(c.Types, "Creature")
					}
					for _, st := range []string{"Human", "Soldier"} {
						if !slices.Contains(c.Subtypes, st) {
							c.Subtypes = append(c.Subtypes, st)
						}
					}
				},
			},
			KeywordGrant(gideonBlackbladeIsACreature, "indestructible"),
			{
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7B_Set,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return gideonBlackbladeIsACreature(target, g, source)
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power, c.Toughness = 4, 4
				},
			},
		},
		Replacements: []game.ReplacementEffect{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Prevention: true, // CR 615.1a — "prevent"; CR 615.12 reads it
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventDamage && ev.DamageTarget == src.InstanceID &&
					isActivePlayer(g, src.Controller)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Gideon Blackblade — prevent all damage that would be dealt to him during your turn",
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Up to one other target creature you control gains your choice of vigilance, lifelink, or indestructible until end of turn.",
				Cost:    LoyaltyCost(1),
				Targets: Another(TargetCreature("up to one other target creature you control", YouControl())).WithCount(0, 1),
				Effect:  gideonBlackbladePlusOne,
			},
			{
				Label:   "−6: Exile target nonland permanent.",
				Cost:    LoyaltyCost(-6),
				Targets: TargetPermanent("target nonland permanent", Nonland()),
				Effect:  ExileFirstTarget,
			},
		},
	})
}

// gideonBlackbladeIsACreature is the condition of his first sentence:
// it is his controller's turn, read at every layer pass.
func gideonBlackbladeIsACreature(target *game.Card, g *game.Game, source *game.Card) bool {
	return target.InstanceID == source.InstanceID && isActivePlayer(g, source.Controller)
}

// gideonBlackbladeKeywords are the three keywords the +1 offers, in
// printed order.
var gideonBlackbladeKeywords = []string{"vigilance", "lifelink", "indestructible"}

// gideonBlackbladePlusOne asks the controller for a keyword as the
// ability resolves and grants it to the target until end of turn. No
// target, or a target that left, does nothing (CR 608.2b).
func gideonBlackbladePlusOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 || targets[0].Kind != game.TargetCard {
		return nil
	}
	target := targets[0].ID
	options := make([]game.ChoiceOption, len(gideonBlackbladeKeywords))
	for i, kw := range gideonBlackbladeKeywords {
		options[i] = game.ChoiceOption{Label: "Gains " + kw + " until end of turn"}
	}
	return PickOption{
		Question: "Gideon Blackblade — the creature gains your choice of vigilance, lifelink, or indestructible until end of turn",
		Options:  options,
		Then: func(ctx *Context, index int) error {
			if index < 0 || index >= len(gideonBlackbladeKeywords) {
				return nil
			}
			return GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{gideonBlackbladeKeywords[index]},
				Label:    "Gideon Blackblade — gains " + gideonBlackbladeKeywords[index] + " until end of turn",
			}.Apply(ctx)
		},
	}.Apply(ctx)
}
