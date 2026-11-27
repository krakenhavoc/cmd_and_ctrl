package effects

import (
	"fmt"
	"slices"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b651_helpers.go — Invasion, S82 older sets batch 651 (#5616).

// OldB651CostsMoreMana is "<matching> spells you cast cost <mana> more to
// cast" for a coloured unit (Andradite Leech's {B}): CostsMore with one point
// of a fixed mana Unit, which the increase pass adds as ordinary coloured
// requirements (CR 601.2f). Panics at init on a unit the cost parser rejects.
func OldB651CostsMoreMana(unit, label string, when ...CostPredicate) game.CostModifier {
	parsed, err := game.ParseCost(unit)
	if err != nil || parsed.XSlots > 0 || parsed.ManaValue() == 0 {
		panic(fmt.Sprintf("effects.OldB651CostsMoreMana: %q is not a fixed mana amount (%v)", unit, err))
	}
	m := CostsMore(1, label, when...)
	m.Unit = &parsed
	return m
}

// OldB651SelfPumpWhileNoOpponentControlsCreatureOf is "This creature gets
// +p/+t as long as no opponent controls a <colour> or <colour> creature": a
// layer 7c modification (CR 613.4c) whose condition reads every opponent's
// creatures through the layers, the colours settled by layer 5 already.
func OldB651SelfPumpWhileNoOpponentControlsCreatureOf(power, toughness int, colors ...string) game.StaticAbility {
	words := make([]string, 0, len(colors))
	for _, col := range colors {
		words = append(words, colorWords([]string{col}))
	}
	return game.StaticAbility{
		Label: "This creature gets " + SignedPT(power, toughness) + " as long as no opponent controls a " +
			strings.Join(words, " or ") + " creature.",
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			if !SelfCreature(target, g, source) {
				return false
			}
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == source.Controller || !c.IsCreature() {
					continue
				}
				for _, col := range c.EffectiveColors() {
					if slices.Contains(colors, col) {
						return false
					}
				}
			}
			return true
		},
		Apply: PumpBy(power, toughness),
	}
}
