package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Maturation Chamber — Artifact {1}{W} (Star Trek Commander, #4974):
//
//	"Whenever you cast an artifact spell, put a charge counter on this
//	 artifact.
//	 {4}, {T}: Create a 1/1 colorless Borg artifact creature token for
//	 each charge counter on this artifact. Activate only as a sorcery."
//
// The cast trigger is Vedalken Archmage's condition
// (b29ArtifactSpellCastByYou): the spell's types are read off the stack
// as it is cast. The tokens are counted as the
// ability resolves, from the Chamber's last-known information if it has
// left the battlefield (CR 608.2h), and are the set's printed Borg.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "585511e1-4cca-4e70-963d-56f98ae49c29",
		Name:         "Maturation Chamber",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b29ArtifactSpellCastByYou(ev, source, g)
			}, "Maturation Chamber — put a charge counter on this artifact",
				Do(CounterOnThis{Kind: game.CounterCharge, N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:        "{4}, {T}: Create a 1/1 colorless Borg artifact creature token for each charge counter on this artifact. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{4}"), TapCost()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				info, ok := ctx.SourcePermanent()
				if !ok {
					return nil
				}
				n := info.Counters[game.CounterCharge]
				if n <= 0 {
					return nil
				}
				return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 colorless Borg artifact"), N: n}.Apply(ctx)
			},
		}},
	})
}
