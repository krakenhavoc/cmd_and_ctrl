package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Plaza of Heroes — Land:
//
//	"{T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast a
//	 legendary spell.
//	 {T}: Add one mana of any color among legendary permanents you
//	 control.
//	 {3}, {T}, Exile this land: Target legendary creature gains hexproof
//	 and indestructible until end of turn."
//
// Four lines, three mana abilities and one stack ability, each a
// shape the catalog already has:
//
//   - The colored mana for legendary spells is Delighted Halfling's
//     restricted slot: cast-only and Legendary-supertype, enforced at
//     SPEND time (ManaRestrictCast + ManaRestrictSupertype), so the mana
//     cannot pay for a non-legendary spell or an ability.
//   - "Any color among legendary permanents you control" is Mox Amber's
//     board-derived pick, widened from "creatures and planeswalkers" to
//     every legendary permanent (a legendary artifact or land offers its
//     colors too). It reads colors, not what the permanents could
//     produce; a colorless legendary offers nothing, and the ability
//     adds no mana when no legendary permanent has a color.
//   - The last ability exiles the land as a COST (ExileThis), so the
//     land is gone before the target is protected and nothing can
//     respond by sacrificing it. The target is any legendary creature,
//     not only yours; the grant is evaluated at resolution and lasts
//     until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9c58d241-4d9f-4b46-b8ee-f4587f9acfd6",
		Name:         "Plaza of Heroes",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color (legendary spells only)",
				// "one mana of any color" names no commander
				// identity, so all five colors are offered and the
				// spend restriction is the only limit.
				Restrictions: []string{
					ManaRestrictCast,
					ManaRestrictSupertype("Legendary"),
				},
			},
			{
				Cost:         ManaAbilityCost{Tap: true},
				ProducedFunc: producedFromAnyLegendaryPermanent(),
				Label:        "Add one mana of any color among legendary permanents you control",
			},
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}, Exile this land: Target legendary creature gains hexproof and indestructible until end of turn.",
			Cost:    Plus(ManaCost("{3}"), TapCost(), ExileThis()),
			Targets: TargetCreature("target legendary creature", Legendary()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if err := (GrantKeywordUntilEOT{
						Target:   t.ID,
						Keywords: []string{"hexproof", "indestructible"},
						Label:    "Plaza of Heroes — hexproof and indestructible until end of turn",
					}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}

// producedFromAnyLegendaryPermanent is "one mana of any color among
// legendary permanents you control": ProducedFromLegendaryPermanents
// (Mox Amber) without its creature-or-planeswalker filter.
func producedFromAnyLegendaryPermanent() func(*game.Game, uuid.UUID, uuid.UUID) string {
	return func(g *game.Game, controller, _ uuid.UUID) string {
		seen := map[string]bool{}
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller != controller {
				continue
			}
			if !hasFold(c.Effective().Supertypes, "Legendary") {
				continue
			}
			for _, col := range c.EffectiveColors() {
				seen[strings.ToUpper(col)] = true
			}
		}
		return pipeString(colorsOnly(orderedManaSymbols(seen)))
	}
}
