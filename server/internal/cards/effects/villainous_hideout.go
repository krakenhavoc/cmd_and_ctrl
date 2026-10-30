package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Villainous Hideout — Land (Marvel Super Heroes, #4455):
//
//	"{T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast a Villain
//	 spell or to activate an ability of a Villain source.
//	 {3}, {T}: Target Villain you control connives. Activate only as a
//	 sorcery. (Draw a card, then discard a card. If you discarded a nonland
//	 card, put a +1/+1 counter on that creature.)"
//
// Avengers Tower's two mana abilities with the word Villain: one
// ManaRestrictAnyOf tag, "cast a Villain spell" or "activate an ability of a
// Villain source" (CR 106.6). The connive is Baron Strucker's ConniveEach on
// the one target Villain you control, activated at sorcery speed (CR 701.50).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cd2888aa-71f3-47ee-ba33-7bb95d5bc836",
		Name:         "Villainous Hideout",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			colorlessTapRow(),
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "{T}: Add one mana of any color (Villain spells and Villain abilities only).",
				Restrictions: []string{game.ManaRestrictAnyOf(
					[]string{game.ManaRestrictCast, game.ManaRestrictSubtype("Villain")},
					[]string{game.ManaRestrictActivate, game.ManaRestrictSubtype("Villain")},
				)},
			},
		},
		Activated: []ActivatedAbility{{
			Label:        "{3}, {T}: Target Villain you control connives. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{3}"), TapCost()),
			SorcerySpeed: true,
			Targets:      TargetCreature("target Villain you control", YouControl(), HasSubtype("Villain")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				target, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return ConniveEach{Controller: item.Controller, Creatures: []uuid.UUID{target}}.Apply(ctx)
			},
		}},
	})
}
