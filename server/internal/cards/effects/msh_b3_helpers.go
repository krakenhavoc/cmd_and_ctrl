package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msh_b3_helpers.go — shared bodies for Marvel Super Heroes half B, batch 3
// (#4455).

// mshB3EnteredOrControlsBasic is the Marvel land cycle's activation clause,
// "Activate only if this land entered this turn or if you control a basic
// land" (Dark Fortress, Gleaming Bastion). The land counts as having entered
// this turn by the engine's entry record, not the summoning-sickness marker
// (see SourceEnteredThisTurn); the basic land is any other basic land its
// controller controls, found the way the Avatar cycle's is.
func mshB3EnteredOrControlsBasic(g *game.Game, _, source uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(source)
	if !ok {
		return false
	}
	return SourceEnteredThisTurn(g, c) || b11ControlsBasicLand(g, &c)
}

// mshB3ConditionalDualLand is the cycle's two mana abilities: "{T}: Add {C}"
// and "{T}: Add <a> or <b>. Activate only if this land entered this turn or
// if you control a basic land." `produced` is the pair, e.g. "{B|R}".
func mshB3ConditionalDualLand(produced, label string) []ManaAbility {
	return []ManaAbility{
		{Cost: ManaAbilityCost{Tap: true}, Produced: "{C}", Label: "Add {C}"},
		{
			Cost:      ManaAbilityCost{Tap: true},
			Produced:  produced,
			Label:     label,
			Condition: mshB3EnteredOrControlsBasic,
		},
	}
}
