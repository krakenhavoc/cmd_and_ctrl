package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Horn of Gondor — Legendary Artifact {3}:
//
//	"When Horn of Gondor enters, create a 1/1 white Human Soldier
//	 creature token.
//	 {3}, {T}: Create X 1/1 white Human Soldier creature tokens, where
//	 X is the number of Humans you control."
//
// X is counted as the ability resolves, not when it is activated, so
// a Human that dies in response is not counted and one that arrives
// is. The Soldier tokens are Humans themselves, so the Horn feeds
// on its own output.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "88777fc0-286d-47fc-8e94-1e289b4964c4",
		Name:         "Horn of Gondor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Horn of Gondor — create a 1/1 white Human Soldier creature token",
				Do(CreateToken{Template: TokenCard("1/1 white Human Soldier"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label: "{3}, {T}: Create X 1/1 white Human Soldier creature tokens, where X is the number of Humans you control",
			Cost:  Plus(ManaCost("{3}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				n := b43CreaturesOfSubtypeControlled(g, item.Controller, "Human")
				if n <= 0 {
					return nil
				}
				return CreateToken{Template: TokenCard("1/1 white Human Soldier"), N: n}.Apply(NewContext(g, item))
			},
		}},
	})
}
