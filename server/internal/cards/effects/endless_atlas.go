package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Endless Atlas — Artifact {2} (EDHREC rank 2280):
//
//	"{2}, {T}: Draw a card. Activate only if you control three or
//	 more lands with the same name."
//
// A basic-heavy deck's card-draw rock. The gate is the activation
// condition (CR 602.1b, #743): some one name is shared by at least
// three lands you control. Read off the battlefield's names, so three
// Islands open it and an Island, a Snow-Covered Island and a Tropical
// Island do not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cb923c4d-1d2e-4039-90bb-72d1981b9738",
		Name:         "Endless Atlas",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{2}, {T}: Draw a card. Activate only if you control three or more lands with the same name.",
			Cost:      Plus(ManaCost("{2}"), TapCost()),
			Condition: endlessAtlasThreeLandsSharingAName,
			Effect:    b36DrawOne,
		}},
	})
}

// endlessAtlasThreeLandsSharingAName is the Atlas's condition.
func endlessAtlasThreeLandsSharingAName(g *game.Game, controller, _ uuid.UUID) bool {
	byName := map[string]int{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller || !c.IsLand() {
			continue
		}
		byName[c.Name]++
		if byName[c.Name] >= 3 {
			return true
		}
	}
	return false
}
