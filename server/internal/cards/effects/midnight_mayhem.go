package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Midnight Mayhem — Sorcery {2}{R}{W} (Duskmourn: House of Horror,
// #4466):
//
//	"Create three 1/1 red Gremlin creature tokens. Gremlins you control
//	 gain menace, lifelink, and haste until end of turn. (A creature with
//	 menace can't be blocked except by two or more creatures.)"
//
// Circle of Power's shape: the tokens are created first, then the grant
// picks out every Gremlin creature you control at that moment, the new
// tokens and any other Gremlin alike (CR 611.2c: the set is fixed as
// the effect begins, so a Gremlin that arrives later this turn does not
// gain them).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5236a6b5-0a96-4470-ba03-37b6d7a889bb",
		Name:         "Midnight Mayhem",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3},
		// The grant waits for a prompt the token creation asks (CR 608.2c, #4155).
		OnResolve: ResolveSteps(
			CreateToken{Template: TokenCard("1/1 red Gremlin"), N: 3},
			GrantKeywordUntilEOT{
				Match:    And(OfCreatureType("Gremlin"), YouControl()),
				Keywords: []string{"menace", "lifelink", "haste"},
				Label:    "Midnight Mayhem — Gremlins gain menace, lifelink, and haste until end of turn",
			},
		),
	})
}
