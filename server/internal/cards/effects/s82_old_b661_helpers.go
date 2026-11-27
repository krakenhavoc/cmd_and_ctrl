package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b661_helpers.go — shared pieces of Invasion batch 661 (#5616).

// OldB661UpkeepSacrificeUnlessYouPayItsManaCost is "At the beginning of your
// upkeep, sacrifice this permanent unless you pay its mana cost" (Essence
// Leak's granted ability). The cost is read off the permanent as the trigger
// resolves; a permanent with no mana cost has an unpayable one (CR 202.1a), so
// it is sacrificed without a question. {X} in a cost is paid as 0 (CR 107.3i).
func OldB661UpkeepSacrificeUnlessYouPayItsManaCost() game.TriggeredAbility {
	const label = "Essence Leak — sacrifice this permanent unless you pay its mana cost"
	return AtYourUpkeep(label, func(g *game.Game, item *game.StackItem) error {
		cost, hasCost := "", false
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.InstanceID == item.SourceCardID {
				cost = c.ManaCost
				hasCost = cost != ""
				break
			}
		}
		ctx := NewContext(g, item)
		if !hasCost {
			return SacrificeThisIfStillOnBattlefield(g, item)
		}
		cost = oldB661StripX(cost)
		return UpkeepPayUnless{
			Chooser:  item.Controller,
			Cost:     cost,
			Question: "Essence Leak — pay " + cost + " or sacrifice this permanent?",
			Else:     UpkeepSacrificeSource,
		}.Apply(ctx)
	})
}

// oldB661StripX writes each {X} of a mana cost as {0}.
func oldB661StripX(cost string) string {
	out := ""
	for i := 0; i < len(cost); i++ {
		if i+2 < len(cost) && cost[i] == '{' && cost[i+1] == 'X' && cost[i+2] == '}' {
			out += "{0}"
			i += 2
			continue
		}
		out += string(cost[i])
	}
	return out
}

// OldB661ProtectionFromEachColorAmongYourPermanentsSparingThisAura is "Enchanted
// creature has protection from each color among permanents you control. This
// effect doesn't remove this Aura." (Pledge of Loyalty; CR 702.16n). "You" is
// the Aura's controller, and the colors are read as the layers run.
func OldB661ProtectionFromEachColorAmongYourPermanentsSparingThisAura() game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer6Ability,
		Label:     "Enchanted creature has protection from each color among permanents you control. This effect doesn't remove this Aura.",
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			for _, col := range ColorsAmongPermanentsYouControl(g, source.Controller) {
				tok := game.ProtectionSparingAura(game.ProtectionFromColor(col), source.InstanceID)
				if tok != "" {
					c.Abilities = game.AppendKeywordAbility(c.Abilities, tok)
				}
			}
		},
	}
}
