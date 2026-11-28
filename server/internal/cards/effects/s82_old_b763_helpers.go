package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b763_helpers.go — shared pieces of the Tempest batch 763 lands
// (#5628).

// S82OldB763TappedPainland is one Tempest enters-tapped painland: "This land
// enters tapped. {T}: Add {C}. {T}: Add {a} or {b}. This land deals 1 damage to
// you."
func S82OldB763TappedPainland(oracleID, name, a, b string) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{
			PainlessColorless(),
			PainDual(a, b, "This land"),
		},
	}
}

// S82OldB763FreezingLand is one Tempest freezing land: "{T}: Add {C}. {T}: Add
// {a} or {b}. This land doesn't untap during your next untap step."
func S82OldB763FreezingLand(oracleID, name, a, b string) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			PainlessColorless(),
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{" + a + "|" + b + "}",
				Label:    "Add {" + a + "} or {" + b + "}. This land doesn't untap during your next untap step.",
				Rider:    S82OldB763SitOutNextUntap,
			},
		},
	}
}

// S82OldB763SitOutNextUntap is the rider: the land, as the object it is, does
// not untap during its controller's next untap step (CR 502.3).
func S82OldB763SitOutNextUntap(g *game.Game, controller, source uuid.UUID) error {
	return g.SkipNextUntapForEffect(source, controller)
}
