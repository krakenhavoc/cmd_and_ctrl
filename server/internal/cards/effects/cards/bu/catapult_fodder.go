package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Catapult Fodder // Catapult Captain — {2}{B} Creature — Zombie 1/5 //
// Creature — Zombie 2/6 (Innistrad: Crimson Vow):
//
//	Front: "At the beginning of combat on your turn, if you control three or
//	        more creatures that each have toughness greater than their power,
//	        transform this creature."
//	Back:  "{2}{B}, {T}, Sacrifice another creature: Target opponent loses
//	        life equal to the sacrificed creature's toughness."
//
// Power and toughness are read through the layers (the Fodder counts
// itself). The "if" is checked as combat begins and as the trigger
// resolves (CR 603.4). The Captain's sacrifice is a cost, so the toughness
// is the creature's last-known (CR 608.2h).
//
// No simplification.
func init() {
	const oracle = "aa2c0d23-8df8-408b-a8c0-c4cd005f078d"
	three := func(g *game.Game, source *game.Card) bool {
		n := 0
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == source.Controller && c.IsCreature() && c.CurrentToughness() > c.PowerForComparison() {
				n++
			}
		}
		return n >= 3
	}
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Catapult Fodder",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventStepBegan, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.StepBegan(game.StepBeginCombat, true)(ev, source, lki, g) && three(g, source)
			}, "Catapult Fodder — transform this creature", effects.S82VowB14TransformIf(three)),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Catapult Captain",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, {T}, Sacrifice another creature: Target opponent loses life equal to the sacrificed creature's toughness.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.TapCost(), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				fed, ok := effects.B17PermanentSacrificedToPay(g, item)
				if !ok {
					return nil
				}
				n := effects.DepartedCreatureToughness(g, fed)
				if n <= 0 {
					return nil
				}
				for _, t := range effects.NewContext(g, item).LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return g.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -n)
					}
				}
				return nil
			},
		}},
	})
}
