package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// CreaturesWithoutAnyCantAttack is "Creatures without <keyword> or <keyword>
// can't attack" (Stormtide Leviathan; Moat's shape): a layer-6 static over
// every creature, whoever controls it, that records the keywords one of
// which a creature needs to attack. The engine folds the restriction in once
// every layer has run (game.foldAttackNeedsKeywordLocked), so a creature that
// is granted flying or islandwalk, however late its timestamp, may attack,
// and one that loses it may not. The restriction belongs to the source, so a
// creature that loses all abilities is not freed by it.
func CreaturesWithoutAnyCantAttack(label string, keywords ...string) game.StaticAbility {
	kws := slices.Clone(keywords)
	return game.StaticAbility{
		Label: label,
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
			return target.IsCreature()
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			for _, kw := range kws {
				if !slices.Contains(c.CantAttackWithoutAnyOf, kw) {
					c.CantAttackWithoutAnyOf = append(c.CantAttackWithoutAnyOf, kw)
				}
			}
		},
	}
}

// AllCreaturesCanAttackYourOpponentsAsThoughHaste is "All creatures can
// attack your opponents and planeswalkers your opponents control as though
// those creatures had haste" (Frenzied Saddlebrute): a layer-6 static over
// every creature, whoever controls it, that records the source's controller
// as the player whose opponents may be attacked without regard to summoning
// sickness (game.Characteristic.AttackHasteFor). "Your" is read live, so a
// stolen Saddlebrute answers to its new controller.
func AllCreaturesCanAttackYourOpponentsAsThoughHaste() game.StaticAbility {
	return game.StaticAbility{
		Label: "All creatures can attack your opponents and planeswalkers your opponents control as though those creatures had haste.",
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
			return target.IsCreature()
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			if source != nil && !slices.Contains(c.AttackHasteFor, source.Controller) {
				c.AttackHasteFor = append(c.AttackHasteFor, source.Controller)
			}
		},
	}
}
