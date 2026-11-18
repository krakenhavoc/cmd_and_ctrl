package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rotten Reunion — Instant {B} (Midnight Hunt):
//
//	"Exile up to one target card from a graveyard. Create a 2/2 black
//	 Zombie creature token with decayed.
//	 Flashback {1}{B}"
//
// With no target chosen the spell still resolves and makes the token. A
// chosen card that has left the graveyard is skipped and the token is
// still made (CR 608.2b: only a spell with every target illegal is
// countered).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "1bde8eaa-9beb-4529-b540-5e0155c09713",
		Name:             "Rotten Reunion",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Tokens: 1, CreatureTokens: 1},
		Targets:          TargetCardInGraveyard("up to one target card from a graveyard").WithCount(0, 1),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{1}{B}")},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return ExileTargetCards(ctx.Game, ctx.Item) }),
			CreateToken{Template: TokenCard("2/2 black Zombie with decayed"), N: 1},
		),
	})
}
