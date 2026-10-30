package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// White Tiger, Ava Ayala — {1}{G} Legendary Creature — Human Hero 2/2
// (Marvel Super Heroes, #4455):
//
//	"Power-up — {5}{G}: Put a +1/+1 counter on White Tiger and create The
//	 Tiger God, a legendary 4/4 green Cat God creature token with "The
//	 Tiger God can't be blocked by more than one creature." (Activate each
//	 power-up ability only once. Reduce the cost by her mana cost if she
//	 entered this turn.)"
//
// Power-up is the engine's (CR 702.193). The Tiger God is a catalog token
// (ADR 0083) with Outland Colossus's block limit on itself, MaxBlockers
// (OnSelf(), 1); legendary, so a second one costs its controller one
// (CR 704.5j). The counter goes on White Tiger only if she is still the same
// object; the token is made either way.
//
// No simplification.
func init() {
	ab := PowerUp("{5}{G}: Put a +1/+1 counter on White Tiger and create The Tiger God, a legendary 4/4 green Cat God creature token with \"The Tiger God can't be blocked by more than one creature.\"",
		ManaCost("{5}{G}"), func(g *game.Game, item *game.StackItem) error {
			if err := plusOneCountersOnThis(1)(g, item); err != nil {
				return err
			}
			return CreateToken{Template: tigerGodToken(), N: 1}.Apply(NewContext(g, item))
		})
	ab.Purpose.Answers = game.AnswerValue
	Register(Spec{
		OracleID:     "e2084d73-5a4c-4b17-942a-93c1a1f7ac2a",
		Name:         "White Tiger, Ava Ayala",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{ab},
	})
}

// tigerGodToken is White Tiger's The Tiger God.
func tigerGodToken() game.Card { return tokenFromCatalog(printedTigerGodToken) }

// printedTigerGodToken is The Tiger God as PRINTED. Its slug is an on-disk
// identity and is never renamed.
func printedTigerGodToken() tokenTemplate {
	return tokenTemplate{
		Slug: "the-tiger-god",
		Card: game.Card{
			Name:      "The Tiger God",
			TypeLine:  "Token Legendary Creature — Cat God",
			Power:     4,
			Toughness: 4,
			Colors:    []string{"G"},
		},
		BlockRules: []game.BlockRule{MaxBlockers(OnSelf(), 1)},
		Text:       "The Tiger God can't be blocked by more than one creature.",
	}
}
