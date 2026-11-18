package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     oracle,
		Name:         "Catapult Fodder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventStepBegan, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return StepBegan(game.StepBeginCombat, true)(ev, source, lki, g) && three(g, source)
			}, "Catapult Fodder — transform this creature", s82VowB14TransformIf(three)),
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Catapult Captain",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}, {T}, Sacrifice another creature: Target opponent loses life equal to the sacrificed creature's toughness.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{2}{B}"), TapCost(), SacrificeAnotherN(1, "another creature", Creature())),
			Targets: TargetPlayer("target opponent", Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				fed, ok := b17PermanentSacrificedToPay(g, item)
				if !ok {
					return nil
				}
				n := departedCreatureToughness(g, fed)
				if n <= 0 {
					return nil
				}
				for _, t := range NewContext(g, item).LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return g.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -n)
					}
				}
				return nil
			},
		}},
	})
}
