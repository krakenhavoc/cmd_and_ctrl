package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const legolassQuickReflexesGrant = "legolass-quick-reflexes/tap-damage"

// Legolas's Quick Reflexes — Instant {G} (Tales of Middle-earth Commander):
//
//	"Split second (As long as this spell is on the stack, players can't cast
//	 spells or activate abilities that aren't mana abilities.)
//	 Untap target creature. Until end of turn, it gains reach, hexproof, and
//	 "Whenever this creature becomes tapped, it deals damage equal to its
//	 power to up to one target creature.""
//
// Split second is the printed keyword (CR 702.61). The untap, the reach and
// hexproof (Layer 6) and the granted trigger (a bundle declared in
// Spec.Grants, Furnace Reins's shape) all apply to the one target, and each
// ends at cleanup. The granted trigger fires every time the creature becomes
// tapped this turn (an attack that taps it, a tap cost, an effect), takes its
// own "up to one target creature" as it goes on the stack, and deals damage
// equal to the creature's power as the ability resolves, or as it last
// existed if it has left (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "59154c04-6bc5-4b95-af1f-6de5fed05649",
		Name:            "Legolas's Quick Reflexes",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordSplitSecond},
		Grants: []effects.AbilityGrant{{
			Key: legolassQuickReflexesGrant,
			Triggered: []game.TriggeredAbility{
				effects.Targeting(effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
					"Legolas's Quick Reflexes — this creature deals damage equal to its power to up to one target creature",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						ts := ctx.LegalTargets()
						if len(ts) == 0 || ts[0].Kind != game.TargetCard {
							return nil
						}
						power := effects.SourcePowerNow(ctx)
						if power <= 0 {
							return nil
						}
						return effects.DealDamage{Source: item.SourceCardID, Target: ts[0].ID, Amount: power}.Apply(ctx)
					}),
					effects.UpToOneTargetCreature("up to one target creature")),
			},
			Text: "Whenever this creature becomes tapped, it deals damage equal to its power to up to one target creature.",
		}},
		Targets: effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (effects.UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"reach", "hexproof"}, Label: "Legolas's Quick Reflexes — reach and hexproof"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantAbilitiesFor{
				Target: target,
				Keys:   []string{legolassQuickReflexesGrant},
				Label:  "Legolas's Quick Reflexes — deals damage equal to its power whenever it becomes tapped",
			}.Apply(ctx)
		},
	})
}
