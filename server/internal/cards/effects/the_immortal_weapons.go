package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Immortal Weapons — Legendary Creature — Human Warrior Hero {4}{R},
// 4/4 (Marvel Super Heroes Commander, #4455):
//
//	"When The Immortal Weapons enter, return target instant or sorcery
//	 card from your graveyard to your hand.
//	 Whenever you cast a noncreature spell, target creature gets +2/+0 and
//	 gains menace until end of turn."
//
// Two printed triggers, two registered. The first is Archaeomancer's
// graveyard-to-hand body on an enters trigger (Along the Crooked Way's
// shape). The second is Black Bolt's cast trigger with a target: the
// boost and the keyword are one effect on the one target, both until end
// of turn; a target that is gone when it resolves gets neither (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "760d6cb2-5d8a-495c-9853-f96a1efa5775",
		Name:         "The Immortal Weapons",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("The Immortal Weapons — return target instant or sorcery card from your graveyard to your hand",
					func(g *game.Game, item *game.StackItem) error {
						return returnLegalGraveyardTargetsToHand(NewContext(g, item))
					}),
				TargetCardInGraveyard("target instant or sorcery card in your graveyard", Or(Instant(), Sorcery()), YouOwn()),
			),
			TriggerWithPurpose(Targeting(
				WheneverYouCast(Noncreature(), "The Immortal Weapons — target creature gets +2/+0 and gains menace until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						if err := (BoostUntilEOT{Target: id, Power: 2, Label: "The Immortal Weapons — +2/+0"}).Apply(ctx); err != nil {
							return err
						}
						return GrantKeywordUntilEOT{Target: id, Keywords: []string{"menace"}, Label: "The Immortal Weapons — menace"}.Apply(ctx)
					}),
				TargetCreature("target creature")),
				ForTargets(BenefitUntilEOT(0, 2, 0, "menace"))),
		},
	})
}
