package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "63e596a2-9126-4af6-8782-c38687d664ad",
		Name:            "Heliod, Sun-Crowned",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static: []game.StaticAbility{{
			Layer: game.Layer4Type,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID &&
					devotionTo(g, source.Controller, "W") < 5
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				notACreature(c)
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventChangeLife},
			AppliesTo: YouGainedLife,
			Targets:   TargetPermanent("target creature or enchantment you control", Or(Creature(), Enchantment()), YouControl()),
			Key:       "Heliod, Sun-Crowned — a +1/+1 counter on target creature or enchantment",
			Effect:    heliodCounterOnChosenTarget,
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}: Another target creature gains lifelink until end of turn.",
			Cost:    ManaCost("{1}{W}"),
			Targets: Another(TargetCreature("another target creature")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return GrantKeywordUntilEOT{
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
	return AddCounter{Target: item.Targets[0].ID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
}
