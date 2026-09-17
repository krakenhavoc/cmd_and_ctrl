package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hagra Mauling // Hagra Broodpit — modal double-faced card. This file
// is the FRONT face, Instant {2}{B}{B}:
//
//	"This spell costs {1} less to cast if an opponent controls no basic
//	 lands.
//	 Destroy target creature."
//
// The back face, Hagra Broodpit, is registered with the MDFC land
// cycle in mdfc_lands.go under "<oracle>#1".
//
// #746: the reduction is a self cost modifier on face 0 only. The
// engine reads self modifiers under CatalogKey, so playing the land
// back reads the back's spec and gets nothing (ADR 0048 addendum §12).
func init() {
	Register(Spec{
		OracleID:     "37783ce6-af58-4ef6-8ab4-587079970307",
		Name:         "Hagra Mauling",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			CostsLess(1, "This spell costs {1} less to cast if an opponent controls no basic lands.",
				AnOpponentControlsNoBasicLands()),
		},
		OnResolve: destroyTheTargetPermanent,
	})
}
