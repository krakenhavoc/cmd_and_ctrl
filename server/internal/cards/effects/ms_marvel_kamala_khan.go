package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ms. Marvel, Kamala Khan — {2}{U} Legendary Creature — Mutant Inhuman Hero
// 1/4 (Marvel Super Heroes, #4455):
//
//	"Reach, vigilance
//	 You have no maximum hand size.
//	 Embiggen Fist — Whenever you cast a spell that targets a creature you
//	 control, draw a card. Until end of turn, Ms. Marvel gains "Ms. Marvel's
//	 base power is equal to the number of cards in your hand.""
//
// The gained ability is an ADR 0093 bundle given by a duration grant, so it
// is a real layer-6 ability of Ms. Marvel and goes if she leaves. It carries
// one layer 7b static (a granted static may be 7b, not a CDA), which sets her
// base power to your hand size and is re-read whenever a hand changes
// (DependsOnHandSize); her toughness, counters and pumps are untouched. The
// draw comes first, so the hand already holds the card when the power is
// read. Casting a second spell grants it again, which changes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ebedff30-ae43-4861-aab2-afe1fc88337e",
		Name:            "Ms. Marvel, Kamala Khan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance"},
		NoMaxHandSize:   true,
		Grants: []AbilityGrant{{
			Key: msMarvelEmbiggenGrant,
			Static: []game.StaticAbility{{
				Label:             "Ms. Marvel's base power is equal to the number of cards in your hand.",
				Layer:             game.Layer7PT,
				SubLayer:          game.SubLayer7B_Set,
				DependsOnHandSize: true,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.InstanceID == source.InstanceID
				},
				Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
					c.Power = cardsInHand(g, source.Controller)
				},
			}},
			Text: "Ms. Marvel's base power is equal to the number of cards in your hand.",
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, mshB4CastSpellTargetingYourCreature,
				"Ms. Marvel, Kamala Khan — Embiggen Fist: draw a card, and her base power is your hand size until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
						return err
					}
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					return GrantAbilitiesFor{
						Target: item.SourceCardID,
						Keys:   []string{msMarvelEmbiggenGrant},
						Label:  "Ms. Marvel — base power is the number of cards in your hand",
					}.Apply(ctx)
				}),
		},
	})
}

const msMarvelEmbiggenGrant = "ms-marvel-kamala-khan/embiggen-fist"
