package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Li'l Giri Saves the Day — Plane — Secret Lair (pssc, #4870):
//
//	"Legendary Food tokens you control have hexproof and indestructible.
//	 When you planeswalk to this plane, create Li'l Giri, a legendary 1/1
//	 colorless Food Warrior artifact creature token.
//	 Whenever chaos ensues, put X +1/+1 counters on target legendary
//	 artifact creature you control, where X is the number of times you've
//	 rolled the planar die this turn."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6). The
// static grants the two keywords to legendary Food tokens the planar
// controller controls, Li'l Giri among them. The chaos ability (CR 311.7)
// counts the planar die rolls of the turn as it resolves, the roll that
// caused it included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5b35fbd0-65a0-4973-b2c9-841b11c0d426",
		Name:         "Li'l Giri Saves the Day",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{b16GrantKeywords(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.Controller == source.Controller && target.IsToken() && target.IsLegendary() && target.HasSubtype("Food")
		}, "hexproof", "indestructible")},
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThis("Li'l Giri Saves the Day — create Li'l Giri, a legendary 1/1 colorless Food Warrior artifact creature token",
				Do(CreateToken{Template: TokenCard("1/1 colorless Legendary Food Warrior artifact named Li'l Giri"), N: 1})),
			Targeting(WheneverChaosEnsues("Li'l Giri Saves the Day — X +1/+1 counters on target legendary artifact creature you control (X = planar die rolls this turn)",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return AddCounter{Target: t, Kind: game.CounterPlusOne, N: g.PlanarDieRollsThisTurnForEffect()}.Apply(ctx)
				}),
				TargetCreature("target legendary artifact creature you control", And(Legendary(), Artifact(), YouControl()))),
		},
	})
}
