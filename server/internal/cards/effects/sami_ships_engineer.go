package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sami, Ship's Engineer — Legendary Creature — Human Artificer
// {2}{R}{W}, 2/4 (Edge of Eternities, #4461):
//
//	"At the beginning of your end step, if you control two or more
//	 tapped creatures, create a tapped 2/2 colorless Robot artifact
//	 creature token."
//
// Dawnstrike Vanguard's end-step intervening "if" (CR 603.4): checked as
// the end step begins and again as the ability resolves, so untapping a
// creature in response makes nothing. Sami counts toward the two when
// tapped. The Robot is Secluded Starforge's 2/2 row, made with a
// TokenSpec, so it enters tapped rather than being tapped afterwards.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "470c16c5-b078-47f5-89b0-42991e1f7157",
		Name:         "Sami, Ship's Engineer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you control two or more tapped creatures",
				Test: func(ctx *Context) bool { return b24TappedCreaturesControlled(ctx.Game, ctx.Controller()) >= 2 },
			}, On(game.EventBeginEndStep, AllOf(ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b24TappedCreaturesControlled(g, source.Controller) >= 2
			}), "Sami, Ship's Engineer — create a tapped 2/2 colorless Robot artifact creature token",
				Do(CreateTokenAdvanced{Spec: Token(TokenCard("2/2 colorless Robot artifact")).EntersTapped(), N: 1}))),
		},
	})
}
