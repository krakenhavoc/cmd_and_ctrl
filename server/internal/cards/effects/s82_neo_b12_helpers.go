package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_neo_b12_helpers.go — Neon Dynasty batch 12 (#5493): the Auras.
//
// "Enchant artifact or creature" plus "As long as enchanted permanent is a
// creature, it gets ..." is one conditional static per clause, all on the
// condition Careful Cultivation reads (cmdr2B19AttachedIsCreature): the host
// as the earlier layers leave it this pass. The pump and the keyword grant
// come from b43PumpAttachedWhen and b43GrantToAttachedWhen; the three
// clauses below are the ones no earlier card needed.

// neoB12CantBlockWhenCreature is "As long as enchanted permanent is a
// creature, it can't block" (Clawing Torment): a restriction bit (CR 509.1b)
// that holds only while the host is a creature.
func neoB12CantBlockWhenCreature() game.StaticAbility {
	return game.StaticAbility{
		Label:     "As long as enchanted permanent is a creature, it can't block.",
		Layer:     game.Layer6Ability,
		AppliesTo: cmdr2B19AttachedIsCreature,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Restrictions |= game.CantBlock
		},
	}
}

// neoB12LosesKeywordWhenCreature is "As long as enchanted permanent is a
// creature, it ... loses <keyword>" (Short Circuit), layer 6. Like every
// plain "loses X" it strips what the layers have granted so far (CR 613.7).
func neoB12LosesKeywordWhenCreature(keyword string) game.StaticAbility {
	return game.StaticAbility{
		Label:     "As long as enchanted permanent is a creature, it loses " + keyword + ".",
		Layer:     game.Layer6Ability,
		AppliesTo: cmdr2B19AttachedIsCreature,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			kept := c.Abilities[:0]
			for _, a := range c.Abilities {
				if !equalFoldASCIIEffects(a, keyword) {
					kept = append(kept, a)
				}
			}
			c.Abilities = kept
		},
	}
}

// neoB12BasePTWhenCreature is "As long as enchanted permanent is a creature,
// it has base power and toughness P/T" (Awakened Awareness): a layer 7b set,
// so the +1/+1 counters the Aura puts on the host still count (CR 613.4b).
func neoB12BasePTWhenCreature(power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Label:     "As long as enchanted permanent is a creature, it has base power and toughness 1/1.",
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7B_Set,
		AppliesTo: cmdr2B19AttachedIsCreature,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power = power
			c.Toughness = toughness
		},
	}
}

// neoB12HostHasGreatestPower is Historian's Wisdom's condition: the
// permanent `host` is a creature with the greatest power among creatures on
// the battlefield (a tie for greatest counts), read through the layers.
func neoB12HostHasGreatestPower(g *game.Game, host uuid.UUID) bool {
	var hostPower int
	found := false
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == host && c.IsCreature() {
			hostPower, found = c.PowerForComparison(), true
		}
	}
	if !found {
		return false
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.PowerForComparison() > hostPower {
			return false
		}
	}
	return true
}

// neoB12EnchantedHasGreatestPower is Historian's Wisdom's intervening "if"
// (CR 603.4) as a CondBoard: it reads the permanent the Aura enchants as the
// ability resolves, or the one it last enchanted (CR 608.2h), and the
// creatures on the battlefield through the layers.
var neoB12EnchantedHasGreatestPower = CondBoard{
	Name: "enchanted permanent is a creature with the greatest power among creatures on the battlefield",
	Test: func(ctx *Context) bool {
		host, ok := enchantedCreatureAsItResolves(ctx)
		return ok && neoB12HostHasGreatestPower(ctx.Game, host)
	},
}

// neoB12AnimateAttachedLand is "Enchanted land is a P/T <color> Spirit
// creature with <keywords>. It's still a land." (Crackling and Harmonious
// Emergence): one continuous effect that starts in layer 4 (CR 613.1d), so
// every half keeps applying should the Aura lose its abilities part-way
// through the pass (CR 613.6). Layer 4 ADDS the creature type and Spirit
// (the land keeps being a land), layer 5 sets the color, layer 6 grants the
// keywords, and layer 7b sets the base power and toughness.
func neoB12AnimateAttachedLand(color string, power, toughness int, keywords ...string) []game.StaticAbility {
	color4 := game.StaticAbility{
		Label:                 "Enchanted land is a Spirit creature. It's still a land.",
		Layer:                 game.Layer4Type,
		ContinuesAfterRemoval: true,
		AppliesTo:             AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			if !typeListHasEffects(c.Types, "Creature") {
				c.Types = append(c.Types, "Creature")
			}
			for _, s := range source.TextSubtypes("Spirit") {
				if !typeListHasEffects(c.Subtypes, s) {
					c.Subtypes = append(c.Subtypes, s)
				}
			}
		},
	}
	grant := GrantToAttached(keywords...)
	grant.ContinuesAfterRemoval = true
	return []game.StaticAbility{
		color4,
		SetAttachedColors(color),
		grant,
		SetAttachedBasePT(power, toughness),
	}
}

func typeListHasEffects(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

// neoB12EmergenceReplacement is "If enchanted land would be destroyed,
// instead sacrifice this Aura and that land gains indestructible until end
// of turn" (CR 614.1a). It replaces a destruction of the enchanted land only
// (the gate umbra armor reads: sacrifice, exile and bounce pass through): the
// destruction is cancelled, the Aura is sacrificed, and the land has
// indestructible until end of turn, so the damage or deathtouch that would
// have destroyed it does not. The land's controller orders it against another
// replacement of the same destruction (CR 616.1), and two Auras are two
// replacements, so it acts on its own source.
func neoB12EmergenceReplacement() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			return src != nil && g.UmbraArmorAppliesForEffect(ev, src.InstanceID)
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
			if src == nil {
				return nil
			}
			land, ok := g.LookupCardForEffect(ev.CardID)
			if !ok {
				return nil
			}
			ev.Cancel()
			if err := g.SacrificePermanentForEffect(src.InstanceID); err != nil {
				return err
			}
			g.RegisterScopedEffectForEffect(src.InstanceID,
				[]game.AffectedObject{game.PinObject(land.InstanceID, land.EnteredBattlefieldAt)},
				[]game.Mod{game.AddKeywordsMod("indestructible")},
				g.UntilEndOfTurnDuration(), "enchanted land gains indestructible until end of turn")
			return nil
		},
		Controller: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) uuid.UUID {
			return g.AffectedPermanentControllerForEffect(ev)
		},
		ActsOnItsSource: true,
		Label:           "If enchanted land would be destroyed, instead sacrifice this Aura and that land gains indestructible until end of turn.",
	}
}
