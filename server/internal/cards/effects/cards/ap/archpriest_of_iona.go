package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Archpriest of Iona — Creature — Human Cleric {W}, */2 (Zendikar Rising,
// reprinted in Commander Legends: Battle for Baldur's Gate, #5948):
//
//	"Archpriest of Iona's power is equal to the number of creatures in your
//	 party. (Your party consists of up to one each of Cleric, Rogue, Warrior,
//	 and Wizard.)
//	 At the beginning of combat on your turn, if you have a full party, target
//	 creature gets +1/+1 and gains flying until end of turn."
//
// The power is a characteristic-defining ability in layer 7a (Ixidron's shape)
// over the party count (CR 700.8). The combat trigger has an intervening "if"
// (CR 603.4): a full party is required as the step begins and again as the
// trigger resolves, the target being picked as it goes on the stack.
//
// No simplification.
func init() {
	const label = "Archpriest of Iona — target creature gets +1/+1 and gains flying until end of turn"
	trigger := effects.InterveningIf(effects.FinB9FullParty,
		effects.Targeting(
			effects.On(game.EventStepBegan, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.StepBegan(game.StepBeginCombat, true)(ev, source, lki, g) &&
					effects.FinB9HasFullParty(g, source.Controller)
			}, label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Archpriest of Iona — +1/+1"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Archpriest of Iona — flying"}.Apply(ctx)
			}),
			effects.TargetCreature("target creature")))
	effects.Register(effects.Spec{
		OracleID:     "a8730922-3f47-4c07-bdd5-4a0f557a3d3d",
		Name:         "Archpriest of Iona",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Archpriest of Iona's power is equal to the number of creatures in your party.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(ch *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				ch.Power = effects.FinB9PartySizeOnBoard(g, source.Controller)
			},
		}},
		Triggered: []game.TriggeredAbility{trigger},
	})
}
