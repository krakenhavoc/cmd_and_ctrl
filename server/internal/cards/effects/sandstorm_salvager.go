package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sandstorm Salvager — Creature — Human Artificer {2}{G}, 3/3 (The Big
// Score, #4468):
//
//	"When this creature enters, create a 3/3 colorless Golem artifact
//	 creature token.
//	 {2}, {T}: Put a +1/+1 counter on each creature token you control.
//	 They gain trample until end of turn."
//
// The creature tokens are read once at resolution (CR 611.2c); "they" is
// that same set, so a token that arrives later gains neither.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ec19a9bc-6fc1-4150-976b-37a47b165016",
		Name:         "Sandstorm Salvager",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sandstorm Salvager — create a 3/3 Golem", Do(CreateToken{Template: b35GolemToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}: Put a +1/+1 counter on each creature token you control. They gain trample until end of turn.",
			Cost:  Plus(ManaCost("{2}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ids := b35CreatureTokensYouControl(g, item.Controller)
				if err := (GrantKeywordUntilEOT{Match: b35IDSet(ids), Keywords: []string{"trample"}, Label: "Sandstorm Salvager — trample"}).Apply(ctx); err != nil {
					return err
				}
				return b13PutCounterOnEach(ctx, ids)
			},
		}},
	})
}
