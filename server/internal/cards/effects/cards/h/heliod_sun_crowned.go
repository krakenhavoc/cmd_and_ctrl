package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heliod, Sun-Crowned — Legendary Enchantment Creature — God {2}{W},
// 5/5:
//
//	"Indestructible
//	 As long as your devotion to white is less than five, Heliod
//	 isn't a creature.
//	 Whenever you gain life, put a +1/+1 counter on target creature or
//	 enchantment you control.
//	 {1}{W}: Another target creature gains lifelink until end of
//	 turn."
//
// Thassa, Deep-Dwelling's exact shape, one god over: Indestructible is
// PrintedKeywords, the devotion-gated "isn't a creature" clause is the
// same Layer 4 self-static reading `devotionTo` and calling
// `notACreature` (swap the colour), and the activated ability's
// "another target creature" is object identity (`Another`, CR 109.1).
//
// The lifegain trigger is new to this card: `YouGainedLife` (the
// predicate Archangel of Thune's own "whenever you gain life" uses)
// with a target clause riding the ability, exactly like Acidic
// Slime's ETB — the engine computes the legal set when it fires,
// removes it if nothing qualifies (CR 603.3d), and re-checks the pick
// at resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "63e596a2-9126-4af6-8782-c38687d664ad",
		Name:            "Heliod, Sun-Crowned",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static: []game.StaticAbility{{
			Label: "Isn't a creature unless your devotion to white is five or more.",
			Layer: game.Layer4Type,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID &&
					effects.DevotionTo(g, source.Controller, source.TextColor("W")) < 5
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				effects.NotACreature(c)
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventChangeLife},
			AppliesTo: effects.YouGainedLife,
			Targets:   effects.TargetPermanent("target creature or enchantment you control", effects.Or(effects.Creature(), effects.Enchantment()), effects.YouControl()),
			Key:       "Heliod, Sun-Crowned — a +1/+1 counter on target creature or enchantment",
			Effect:    heliodCounterOnChosenTarget,
			// #4979: one counter per life-gain event (CR 119.9), which the
			// bot reads when it prices a lifelink attack.
			Purpose: game.Purpose{LifeGainPayoff: &game.LifeGainPayoff{Counters: 1}},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}: Another target creature gains lifelink until end of turn.",
			Cost:    effects.ManaCost("{1}{W}"),
			Targets: effects.Another(effects.TargetCreature("another target creature")),
			// #4302: lifelink helps whoever controls the creature, so the
			// bot gives it to its own and only when it matters this turn.
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "lifelink")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{
					Target:   id,
					Keywords: []string{"lifelink"},
					Label:    "Heliod, Sun-Crowned — lifelink",
				}.Apply(ctx)
			},
		}},
	})
}

func heliodCounterOnChosenTarget(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	return effects.AddCounter{Target: item.Targets[0].ID, Kind: game.CounterPlusOne, N: 1}.Apply(effects.NewContext(g, item))
}
