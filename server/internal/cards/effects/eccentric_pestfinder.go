package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eccentric Pestfinder // Turn Stones — a preparation card (CR 722,
// ADR 0090), Secrets of Strixhaven Commander (#4834):
//
//	Eccentric Pestfinder {2}{B}{G}, Creature — Troll Druid 5/5:
//	"Trample
//	 At the beginning of each end step, if you gained life this turn,
//	 this creature becomes prepared."
//	Turn Stones {B}{G}, Sorcery:
//	"For each opponent, you create a 1/1 black and green Pest creature
//	 token with "When this token dies, you gain 1 life.""
//
// "If you gained life this turn" is an intervening if (CR 603.4): asked
// as the end step begins, and again as the trigger resolves. "Each end
// step" is every player's. Turn Stones counts the opponents still in
// the game as it resolves; the Pest is the catalog's Pest template.
//
// No simplification.
const eccentricPestfinderOracleID = "dfaa02d4-4f3c-4e5d-b9e5-283a45395a53"

func init() {
	Register(Spec{
		OracleID:        eccentricPestfinderOracleID,
		Name:            "Eccentric Pestfinder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b32EndStepAndYouGainedLifeThisTurnAtLeast(ev, source, g, 1)
			}, "Eccentric Pestfinder — becomes prepared (you gained life this turn)",
				func(g *game.Game, item *game.StackItem) error {
					if b15LifeGainedThisTurn(g, item.Controller) < 1 {
						return nil
					}
					return socB4ThisBecomesPrepared(g, item)
				}),
		},
	})
	Register(Spec{
		OracleID:     eccentricPestfinderOracleID + "#1",
		Name:         "Turn Stones",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: PestToken(), N: len(ctx.Opponents())}.Apply(ctx)
		},
	})
}
