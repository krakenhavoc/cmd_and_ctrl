package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Balduvian Shaman — Creature — Human Cleric Shaman {U}, 1/1:
//
//	"{T}: Change the text of target white enchantment you control that
//	 doesn't have cumulative upkeep by replacing all instances of one
//	 color word with another. That enchantment gains 'Cumulative upkeep
//	 {1}.'"
//
// #3869. The colour-word change is ADR 0160's 2026-10-15 amendment
// (ChangeTextWord, the words chosen as it resolves, per the ruling); what
// it waited on was the two halves of cumulative upkeep another effect
// touches:
//
//   - "that doesn't have cumulative upkeep" is DoesNotHaveCumulativeUpkeep,
//     read through game.HasCumulativeUpkeep, which sees a printed instance
//     and one granted by another effect (a second Shaman's, a Breath of
//     Dreams's) alike. Checked as the ability is activated and again as it
//     resolves (CR 608.2b).
//   - "That enchantment gains 'Cumulative upkeep {1}.'" is an indefinite
//     ADR 0093 grant of the CumulativeUpkeepGrant bundle, pinned to the
//     enchantment (CR 611.2a, 611.2c), so its age counters, its prompt and
//     its sacrifice are the enchantment's own.
//
// "white" is read through the target walk's word lens (OfColor), so a
// Sleight of Mind on the Shaman changes it. The grant follows the text
// change; it is made even when no word was changed (the ruling lets the
// Shaman target an enchantment with no colour words at all).
//
// No simplification.
const balduvianShamanGrant = "balduvian-shaman/cumulative-upkeep-1"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "3354c3d3-5713-4139-beb2-4fa17f6b47d9",
		Name:         "Balduvian Shaman",
		Completeness: effects.CompletenessFull,
		Grants:       []effects.AbilityGrant{effects.CumulativeUpkeepGrant(balduvianShamanGrant, "{1}")},
		Activated: []effects.ActivatedAbility{{
			Label: "{T}: Change a colour word of target white enchantment you control without cumulative upkeep; it gains cumulative upkeep {1}",
			Cost:  effects.TapCost(),
			Targets: effects.TargetPermanent("target white enchantment you control that doesn't have cumulative upkeep",
				effects.Enchantment(), effects.OfColor("W"), effects.YouControl(), effects.DoesNotHaveCumulativeUpkeep()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ChangeTextOfTarget(effects.TextColorWords, false, balduvianShamanGains)(item, effects.NewContext(g, item))
			},
		}},
	})
}

// balduvianShamanGains is "That enchantment gains 'Cumulative upkeep
// {1}.'", run once the words are chosen, against the asking ability. The
// target is read off the item, so a target that became illegal (CR
// 608.2b) gains nothing: the whole ability did nothing. An on-disk key.
var balduvianShamanGains = effects.CardTail("then/balduvian-shaman-gains", func(ctx *effects.Context, _ game.TailOutcome) error {
	g := ctx.Game
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || !effects.OnBattlefield(g, t.ID) {
			continue
		}
		return effects.GrantAbilitiesFor{
			Target:   t.ID,
			Keys:     []string{balduvianShamanGrant},
			Duration: g.PinnedTo(game.IndefiniteDuration(), t.ID),
			Label:    "Balduvian Shaman — gains cumulative upkeep {1}",
		}.Apply(ctx)
	}
	return nil
})
