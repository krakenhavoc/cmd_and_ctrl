package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alacrian Armory — Artifact {3}{W}:
//
//	"Creatures you control get +0/+1 and have vigilance.
//	 At the beginning of combat on your turn, choose up to one target
//	 Mount or Vehicle you control. Until end of turn, that permanent
//	 becomes saddled if it's a Mount and becomes an artifact creature if
//	 it's a Vehicle."
//
// The target is a real target clause, so the Armory's trigger goes on
// the stack and can be answered. BecomeSaddled leaves a non-Mount
// alone and BecomeCreatureUntilEOT is only applied to a Vehicle, which
// is the printed "if it's a ..." pair.
//
// No simplification.
func init() {
	yourCreatures := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.Controller == source.Controller
	}
	begin := AtBeginningOfYourCombat("Alacrian Armory — a Mount becomes saddled, or a Vehicle becomes an artifact creature", func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			c, ok := g.LookupCardForEffect(t.ID)
			if !ok {
				continue
			}
			if c.HasSubtype("Mount") {
				if err := (BecomeSaddled{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			if c.HasSubtype("Vehicle") {
				if err := (BecomeCreatureUntilEOT{Target: t.ID, Label: "Alacrian Armory — becomes an artifact creature"}).Apply(ctx); err != nil {
					return err
				}
			}
		}
		return nil
	})
	begin.Targets = TargetPermanent("up to one target Mount or Vehicle you control",
		Or(OfSubtype("Mount"), OfSubtype("Vehicle")), YouControl()).WithCount(0, 1)
	Register(Spec{
		OracleID:     "3fa2d51c-d549-4287-ac3b-76bc72bcd8bc",
		Name:         "Alacrian Armory",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: yourCreatures,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Toughness++
				},
			},
			KeywordGrant(yourCreatures, "vigilance"),
		},
		Triggered: []game.TriggeredAbility{begin},
	})
}
