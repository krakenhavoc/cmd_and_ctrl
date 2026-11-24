package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dakkon, Shadow Slayer — Legendary Planeswalker — Dakkon {W}{U}{B},
// loyalty 0 (Modern Horizons 2, #5501):
//
//	"Dakkon enters with a number of loyalty counters on him equal to the
//	 number of lands you control.
//	 +1: Surveil 2.
//	 −3: Exile target creature.
//	 −6: You may put an artifact card from your hand or graveyard onto the
//	 battlefield."
//
// The entry loyalty is an enters-with-counters replacement (CR 614.1c)
// counting the lands you control as it enters, on top of the printed 0. The
// −6 is the pick over both zones's artifact cards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "335b5302-2272-405a-b2b8-3943b3df335f",
		Name:         "Dakkon, Shadow Slayer",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterLoyalty, s82OldB39LandsYouControl,
				"Dakkon enters with a number of loyalty counters equal to the number of lands you control"),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:  "+1: Surveil 2.",
				Cost:   effects.LoyaltyCost(1),
				Effect: effects.Do(effects.Surveil{N: 2}),
			},
			{
				Label:   "−3: Exile target creature.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.SocB37ExileTheLegalTarget,
			},
			{
				Label:  "−6: You may put an artifact card from your hand or graveyard onto the battlefield.",
				Cost:   effects.LoyaltyCost(-6),
				Effect: effects.S82OldB39PutArtifactFromHandOrGraveyard("Dakkon, Shadow Slayer"),
			},
		},
	})
}

func s82OldB39LandsYouControl(g *game.Game, src *game.Card) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == src.Controller && c.IsLand() {
			n++
		}
	}
	return n
}
