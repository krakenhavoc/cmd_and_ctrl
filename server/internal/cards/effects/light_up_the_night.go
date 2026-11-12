package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Light Up the Night — Sorcery {X}{R} (Innistrad: Midnight Hunt, #5497,
// seam #5679):
//
//	"Light Up the Night deals X damage to any target. It deals X plus 1
//	 damage instead if that target is a creature or planeswalker.
//	 Flashback—{3}{R}, Remove X loyalty counters from among planeswalkers
//	 you control. If you cast this spell this way, X can't be 0."
//
// Blaze's X burn, with one more damage when the target is a creature or
// planeswalker as the spell resolves (read through the layers, so an
// animated land counts and a battle or a player does not).
//
// THE FLASHBACK is FlashbackRemoveXCounters: {3}{R} and X loyalty counters
// from among planeswalkers the caster controls, split as they choose. It
// is an alternative cost that includes X (CR 107.3a), so CR 107.3b does
// not lock X at 0: the number removed is the spell's X, at least 1, and
// the printed {X}{R} is not paid. A planeswalker left with no loyalty is
// put into its owner's graveyard the next time state-based actions are
// checked, with the spell already on the stack. The card is exiled as it
// leaves the stack (CR 702.34a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "4669926b-93f9-4d80-aee3-492fbf8e0a0a",
		Name:          "Light Up the Night",
		XMatters:      true,
		Completeness:  CompletenessFull,
		Targets:       TargetAny(),
		CastableZones: []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{
			FlashbackRemoveXCounters("{3}{R}", game.CounterLoyalty, "planeswalkers you control", Planeswalker()),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			t := item.Targets[0]
			amount := ctx.X()
			if t.Kind == game.TargetCard {
				if c, ok := ctx.Game.LookupCardForEffect(t.ID); ok && (c.IsCreature() || c.IsPlaneswalker()) {
					amount++
				}
			}
			return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: amount}.Apply(ctx)
		},
	})
}
