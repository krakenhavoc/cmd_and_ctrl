package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chandra, Chill of Compliance — Legendary Planeswalker — Chandra
// {1}{U}{U}, starting loyalty 3:
//
//	"+1: Surveil 1. If you put a noncreature, nonland card into your
//	 graveyard this way, put that card into your hand.
//	 +1: Add {U}. Spend this mana only to cast a noncreature spell.
//	 −X: Tap target artifact or creature. Put X stun counters on it.
//	 −6: You get an emblem with "Whenever you cast a spell, draw a card.""
//
// The first +1 is a surveil whose graveyard is compared before and after
// the answer (surveilKeepNoncreatureNonland): the noncreature, nonland
// card it just binned comes back to hand. The second +1 is a loyalty
// ability, so it uses the stack, unlike a mana ability; the mana carries
// "only to cast a noncreature spell" (it cannot pay for a creature spell
// or an activated ability) and empties with the step. The emblem triggers
// on every spell its owner casts, one draw each.
//
// The −X is a loyalty cost of X (LoyaltyMinusX, #1944): X is announced
// with the activation, no more than her loyalty (CR 606.6), and the
// stun counters read it back. At X = 0 it only taps.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c3dfa1e2-6785-49a0-a194-fb842a8eb63c",
		Name:            "Chandra, Chill of Compliance",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Emblem: &effects.EmblemSpec{
			Label: "Chandra, Chill of Compliance emblem",
			Text:  "Whenever you cast a spell, draw a card.",
			Triggered: []game.TriggeredAbility{
				effects.WheneverYouCast(nil, "Chandra, Chill of Compliance emblem — draw a card", effects.Do(effects.DrawCards{N: 1})),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:  "+1: Surveil 1. If you put a noncreature, nonland card into your graveyard this way, put that card into your hand.",
				Cost:   effects.LoyaltyCost(1),
				Effect: effects.SurveilKeepNoncreatureNonland,
			},
			{
				Label: "+1: Add {U}. Spend this mana only to cast a noncreature spell.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.AddMana{
						Produced:     "{U}",
						Restrictions: []string{effects.ManaRestrictCast, effects.ManaRestrictNotType("Creature")},
					}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "−X: Tap target artifact or creature. Put X stun counters on it.",
				Cost:    effects.LoyaltyMinusX(),
				Targets: effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					ts := ctx.LegalTargets()
					if len(ts) == 0 {
						return nil
					}
					if err := (effects.TapTarget{Target: ts[0].ID}).Apply(ctx); err != nil {
						return err
					}
					if x := ctx.X(); x > 0 {
						return effects.AddCounter{Target: ts[0].ID, Kind: game.CounterStun, N: x}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label: "−6: You get an emblem with \"Whenever you cast a spell, draw a card.\"",
				Cost:  effects.LoyaltyCost(-6),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateEmblem{}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
