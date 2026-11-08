package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Into the Fae Court — Sorcery {3}{U}{U} (Wilds of Eldraine, #4471):
//
//	"Draw three cards. Create a 1/1 blue Faerie creature token with flying
//	 and "This token can block only creatures with flying.""
//
// The draw comes first and the token waits for any question the draws
// ask (a draw replacement's prompt). The Faerie is a catalog token
// template because its block rule is printed on the token: Stormbound
// Geist's "can block only creatures with flying", a restriction checked
// as blockers are declared (CR 509.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8db0a379-4cf2-43a7-932b-c5646a604578",
		Name:         "Into the Fae Court",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 3, Tokens: 1},
		OnResolve: ResolveSteps(
			DrawCards{N: 3},
			CreateToken{Template: FlyingBlockerFaerieToken(), N: 1},
		),
	})
}

// printedFlyingBlockerFaerieToken is Into the Fae Court's 1/1 blue Faerie
// with flying and "This token can block only creatures with flying." Its
// slug is an on-disk identity: never renamed, never reused.
func printedFlyingBlockerFaerieToken() tokenTemplate {
	return tokenTemplate{
		Slug: "faerie-blocks-only-fliers",
		Card: game.Card{
			Name:      "Faerie",
			TypeLine:  "Token Creature — Faerie",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"U"},
			Keywords:  []string{"flying"},
		},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnSelf(), OnMatching(WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
		Text: "Flying\nThis token can block only creatures with flying.",
	}
}

// FlyingBlockerFaerieToken is the Faerie's card value for CreateToken.
func FlyingBlockerFaerieToken() game.Card { return tokenFromCatalog(printedFlyingBlockerFaerieToken) }
