package effects

// Meloku the Clouded Mirror — Legendary Creature — Moonfolk Wizard
// {4}{U}, 2/4:
//
//	"Flying
//	 {1}, Return a land you control to its owner's hand: Create a
//	 1/1 blue Illusion creature token with flying."
//
// The simplest card that prints #1213's cost component, and the
// reason it is in this PR: the ability is a mana symbol, a return,
// and a token, so a test of it is a test of the COST and of nothing
// else.
//
// The return is a cost, not an effect (CR 601.2h / 602.2b): it is paid
// at announce, in full, or the ability is not activated at all. Two
// consequences a player will notice:
//
//   - a Meloku with no land in play cannot activate at all, however
//     much mana is floating. That is CR 118.3, and the engine reads it
//     off the same candidate walk the client's picker is built from,
//     so the menu row is greyed rather than the click refused.
//   - the land is back in hand BEFORE the token is on the stack, so a
//     landfall payoff or a "whenever a land leaves" watcher sees it
//     and resolves ABOVE the token.
//
// The loop is real and is the card: with enough mana, Meloku turns a
// hand of lands into a board of fliers, one land at a time, and puts
// each land back where it can be replayed next turn. Nothing here
// bounds it — the engine's loop breaker counts activations, and every
// iteration costs {1} and a land drop's worth of tempo.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3b96b8c3-b9c8-4712-a3b1-243a67cc013a",
		Name:            "Meloku the Clouded Mirror",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label: "{1}, Return a land you control to its owner's hand: Create a 1/1 blue Illusion creature token with flying.",
			Cost: Plus(
				ManaCost("{1}"),
				ReturnAPermanentToHand("a land you control", Land()),
			),
			Effect: Do(CreateToken{
				Template: TokenCard("1/1 blue Illusion with flying"),
				N:        1,
			}),
		}},
	})
}
