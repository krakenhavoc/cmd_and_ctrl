package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Siani, Eye of the Storm — Legendary Creature — Djinn Monk {3}{U}, 3/2
// (Commander Legends, #5507):
//
//	"Flying
//	 Whenever Siani attacks, scry X, where X is the number of attacking
//	 creatures with flying.
//	 Partner (You can have two commanders if both have partner.)"
//
// X is counted as the trigger resolves (the ruling): every creature that
// is attacking and has flying then, Siani included while it does. With
// none, there is nothing to scry. Partner is the deck rule internal/deck
// reads off the oracle text (ADR 0144).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ecf808fd-4828-497e-b84b-1cb018a9924b",
		Name:            "Siani, Eye of the Storm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Siani, Eye of the Storm — scry X, X the number of attacking creatures with flying", sianiScry),
		},
	})
}

// sianiScry counts the attacking creatures with flying and scries that
// many.
func sianiScry(g *game.Game, item *game.StackItem) error {
	x := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.AttackingTarget != uuid.Nil && game.HasKeyword(&c, "flying") {
			x++
		}
	}
	if x == 0 {
		return nil
	}
	return Scry{Player: item.Controller, N: x}.Apply(NewContext(g, item))
}
