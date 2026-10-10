package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Escape Tunnel — Land:
//
//	"{T}, Sacrifice this land: Search your library for a basic land
//	 card, put it onto the battlefield tapped, then shuffle.
//	 {T}, Sacrifice this land: Target creature with power 2 or less
//	 can't be blocked this turn."
//
// Two separate activated abilities, each with its own sacrifice cost:
// Terramorphic Expanse's fetch (the shared fetchBasicTapped body) and
// Access Tunnel's evasion grant with a smaller power cap. The land is
// sacrificed as a cost, so the second ability still resolves if the
// Tunnel is gone; a target that left in response is skipped
// (CR 608.2b). The Tunnel has no mana ability of its own.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0056fc91-4398-471c-b561-7ff99750ac8a",
		Name:         "Escape Tunnel",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(TapCost(), SacrificeThis()),
				Effect:  fetchBasicTapped,
			},
			{
				Label:   "{T}, Sacrifice this land: Target creature with power 2 or less can't be blocked this turn.",
				Cost:    Plus(TapCost(), SacrificeThis()),
				Targets: TargetCreature("target creature with power 2 or less", PowerLE(2)),
				Effect:  escapeTunnelUnblockable,
			},
		},
	})
}

func escapeTunnelUnblockable(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return RestrictUntilEOT{
			Target:       t.ID,
			Restrictions: game.CantBeBlocked,
			Label:        "Escape Tunnel — can't be blocked",
		}.Apply(ctx)
	}
	return nil
}
