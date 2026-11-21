package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "0056fc91-4398-471c-b561-7ff99750ac8a",
		Name:         "Escape Tunnel",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 1},
				Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
				Effect:  effects.FetchBasicTapped,
			},
			{
				Label:   "{T}, Sacrifice this land: Target creature with power 2 or less can't be blocked this turn.",
				Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
				Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
				Targets: effects.TargetCreature("target creature with power 2 or less", effects.PowerLE(2)),
				Effect:  escapeTunnelUnblockable,
			},
		},
	})
}

func escapeTunnelUnblockable(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return effects.RestrictUntilEOT{
			Target:       t.ID,
			Restrictions: game.CantBeBlocked,
			Label:        "Escape Tunnel — can't be blocked",
		}.Apply(ctx)
	}
	return nil
}
