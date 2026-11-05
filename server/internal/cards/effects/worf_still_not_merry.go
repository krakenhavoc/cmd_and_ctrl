package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Worf, Still Not Merry — {3}{R}{W} Legendary Creature — Klingon Warrior
// Officer 5/5 (Star Trek Commander, #4974):
//
//	"Whenever Worf attacks, for each attacking non-Faerie creature,
//	 create a tapped and attacking 1/1 blue Faerie creature token with
//	 flying."
//
// The count is read as the trigger resolves (CR 608.2h): every attacking
// creature that is not a Faerie then, Worf included. The tokens are
// CreateTokensAttackingYourChoice (CR 508.4: no player is named, so their
// controller chooses what each attacks, Worf's own defender offered
// first). They were never declared as attackers, so they don't trigger
// "whenever ... attacks" (CR 508.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8153cd58-cb8b-461c-a2c9-aaa9e2ea1b31",
		Name:         "Worf, Still Not Merry",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Worf, Still Not Merry — a tapped and attacking Faerie for each attacking non-Faerie creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					var source *game.Card
					if c, ok := g.LookupCardForEffect(item.SourceCardID); ok {
						source = &c
					}
					return CreateTokensAttackingYourChoice{
						Template: TokenCard(trekB46FaerieToken),
						N:        trekB46AttackingNonFaerieCreatures(g, source),
						Tapped:   true,
						Prefer:   triggerAttackTarget(item),
						Label:    "Worf, Still Not Merry",
					}.Apply(ctx)
				}),
		},
	})
}
