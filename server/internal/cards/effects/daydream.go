package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Daydream — Sorcery {W} (Secrets of Strixhaven, #4456):
//
//	"Exile target creature you control, then return that card to the
//	 battlefield under its owner's control with a +1/+1 counter on it.
//	 Flashback {2}{W} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// An immediate blink (Spaceshift's Flicker with PlusOneCounters): the card
// comes back as a new object (CR 400.7) under its OWNER's control, so a
// creature you stole goes home, and it re-triggers its enters abilities
// with the counter already on it. A token exiled this way is gone
// (CR 111.8). The flashback is the keyword helper's alternative cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "094c4486-6cae-4268-850e-ff0082e5ad10",
		Name:             "Daydream",
		Completeness:     CompletenessFull,
		Targets:          TargetCreature("target creature you control", YouControl()),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{2}{W}")},
		OnResolve:        socB22FlickerFirstTargetWithACounter,
	})
}
