package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Consuming Vapors — Sorcery {3}{B}:
//
//	"Target player sacrifices a creature of their choice. You gain life
//	 equal to that creature's toughness.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// The life gain waits for the sacrifice (PlayerSacrificesThenForEffect)
// and reads the creature's toughness as it last existed on the
// battlefield (CR 608.2h), so an anthem or a counter on it counts. A
// player with no creature sacrifices nothing and you gain nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a10ce333-48d7-499b-9355-f381f2395497",
		Name:            "Consuming Vapors",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Targets:         TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			victim, ok := firstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			me := ctx.Controller()
			return ctx.Game.PlayerSacrificesThenForEffect(ctx.Source(), victim,
				sacrificeSpec("a creature", Creature()),
				"Consuming Vapors — sacrifice a creature", 1,
				func(g *game.Game, sacrificed game.PromptedSacrifices) error {
					ids := sacrificed.By(victim)
					if len(ids) == 0 {
						return nil
					}
					toughness := departedCreatureToughness(g, ids[0])
					if toughness <= 0 {
						return nil
					}
					return GainLife{Player: me, Amount: toughness}.Apply(NewContext(g, item))
				})
		},
	})
}

// departedCreatureToughness is departedCreaturePower's twin: the
// toughness of the permanent object `cardID` most recently was, live
// or as it last existed this turn (CR 608.2h), clamped at zero.
//
// Caller holds g.mu in write mode.
func departedCreatureToughness(g *game.Game, cardID uuid.UUID) int {
	ref, ok := g.PermanentRefForEffect(cardID)
	if !ok {
		return 0
	}
	info, ok := g.PermanentForEffect(ref)
	if !ok || info.Toughness < 0 {
		return 0
	}
	return info.Toughness
}
