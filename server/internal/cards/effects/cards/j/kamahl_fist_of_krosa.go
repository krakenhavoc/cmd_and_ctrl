package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kamahl, Fist of Krosa — Legendary Creature — Human Druid {4}{G}{G}, 4/3
// (Dominaria Remastered):
//
//	"{G}: Target land becomes a 1/1 creature until end of turn. It's still
//	 a land.
//	 {2}{G}{G}{G}: Creatures you control get +3/+3 and gain trample until
//	 end of turn."
//
// The first ability adds the creature type to the land and sets its base
// power and toughness to 1/1 (layers 4 and 7b, BecomeCreatureUntilEOT), for
// the turn and pinned to that object (CR 400.7); it keeps being a land, with
// its mana abilities. The second is Overrun's pump over the creatures its
// controller controls as it resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8b21146d-3e92-46e7-b5d9-8240fba5088e",
		Name:         "Kamahl, Fist of Krosa",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{G}: Target land becomes a 1/1 creature until end of turn. It's still a land.",
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Cost:    effects.ManaCost("{G}"),
				Targets: effects.TargetPermanent("target land", effects.Land()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						return effects.BecomeCreatureUntilEOT{
							Target:       t.ID,
							Types:        []string{"Creature"},
							SetPower:     1,
							SetToughness: 1,
							Label:        "Kamahl, Fist of Krosa — becomes a 1/1 creature",
						}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label:   "{2}{G}{G}{G}: Creatures you control get +3/+3 and gain trample until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump, TeamPump: &game.Pump{Power: 3, Toughness: 3, Keywords: []string{"trample"}}},
				Cost:    effects.ManaCost("{2}{G}{G}{G}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					yours := effects.And(effects.Creature(), effects.YouControl())
					if err := (effects.BoostUntilEOT{Match: yours, Power: 3, Toughness: 3, Label: "Kamahl, Fist of Krosa — +3/+3"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Match: yours, Keywords: []string{"trample"}, Label: "Kamahl, Fist of Krosa — trample"}.Apply(ctx)
				},
			},
		},
	})
}
