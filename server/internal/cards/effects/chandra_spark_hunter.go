package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chandra, Spark Hunter — Legendary Planeswalker — Chandra {3}{R},
// starting loyalty 4 (Aetherdrift, #4464):
//
//	"At the beginning of combat on your turn, choose up to one target
//	 Vehicle you control. Until end of turn, it becomes an artifact
//	 creature and gains haste.
//	 +2: You may sacrifice an artifact or discard a card. If you do, draw
//	 a card.
//	 0: Create a 3/2 colorless Vehicle artifact token with crew 1.
//	 −7: You get an emblem with "Whenever an artifact you control enters,
//	 this emblem deals 3 damage to any target.""
//
// The combat trigger is Alacrian Armory's shape (CR 301.7b: a Vehicle
// that becomes a creature keeps its type, and the haste is a keyword
// until end of turn). The +2 is #4566's either/or may-pay, and the 0 is
// Mu Yanling's 3/2 Vehicle token. The emblem is colourless, so a
// permanent with protection from red can be the target of its damage, and
// the emblem is the source (CR 114.1); it sees every artifact that enters
// under its controller's control, a token and the emblem's own first
// artifact included.
//
// No simplification.
func init() {
	begin := AtBeginningOfYourCombat("Chandra, Spark Hunter — up to one target Vehicle becomes an artifact creature with haste",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			t, ok := firstLegalTarget(ctx)
			if !ok {
				return nil
			}
			if err := (BecomeCreatureUntilEOT{Target: t.ID, Label: "Chandra, Spark Hunter — becomes an artifact creature"}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Target:   t.ID,
				Keywords: []string{"haste"},
				Label:    "Chandra, Spark Hunter — haste until end of turn",
			}.Apply(ctx)
		})
	begin.Targets = TargetPermanent("up to one target Vehicle you control",
		OfSubtype("Vehicle"), YouControl()).WithCount(0, 1)
	Register(Spec{
		OracleID:        "30bb01c1-9253-4251-80e7-fb8236b3d858",
		Name:            "Chandra, Spark Hunter",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Triggered:       []game.TriggeredAbility{begin},
		Emblem: &EmblemSpec{
			Label: "Chandra, Spark Hunter emblem",
			Text:  "Whenever an artifact you control enters, this emblem deals 3 damage to any target.",
			Triggered: []game.TriggeredAbility{
				Targeting(
					On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						c, ok := enteredUnderYourControl(ev, source, g, false)
						return ok && c.IsArtifact()
					}, "Chandra, Spark Hunter emblem — 3 damage to any target", chandraSparkHunterEmblemDamage),
					TargetAny()),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label: "+2: You may sacrifice an artifact or discard a card. If you do, draw a card.",
				Cost:  LoyaltyCost(2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return mayPayArtifactOrCard(NewContext(g, item),
						"Chandra, Spark Hunter — sacrifice an artifact or discard a card to draw a card?", mayPayDrawACard)
				},
			},
			{
				Label:  "0: Create a 3/2 colorless Vehicle artifact token with crew 1.",
				Cost:   LoyaltyCost(0),
				Effect: Do(CreateToken{Template: VehicleCrew1Token(), N: 1}),
			},
			{
				Label:  "−7: You get an emblem with \"Whenever an artifact you control enters, this emblem deals 3 damage to any target.\"",
				Cost:   LoyaltyCost(-7),
				Effect: Do(CreateEmblem{}),
			},
		},
	})
}

func chandraSparkHunterEmblemDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := firstLegalTarget(ctx)
	if !ok {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 3}.Apply(ctx)
}
