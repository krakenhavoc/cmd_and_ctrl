package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Savvy Trader — Creature — Human Citizen {3}{G}, 3/3 (Outlaws of Thunder
// Junction Commander):
//
//	"When this creature enters, exile target permanent card from your
//	 graveyard. You may play that card for as long as it remains exiled.
//	 Spells you cast from anywhere other than your hand cost {1} less to
//	 cast."
//
// The enters trigger targets a permanent card of the controller's
// graveyard, exiles it and grants a play permission that lasts while the
// card stays in exile (WhileInZoneDuration): a land may be played, not just
// cast. A target gone by resolution does nothing. The cost clause is the
// shared cmdr2S40CastFromOutsideHandCostsLess, so a spell cast off the
// exiled card is {1} cheaper too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c1aba8c7-d940-4d5d-a176-68cc226906de",
		Name:         "Savvy Trader",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			cmdr2S40CastFromOutsideHandCostsLess(1, "Spells you cast from anywhere other than your hand cost {1} less to cast."),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Savvy Trader — exile target permanent card from your graveyard; you may play it for as long as it remains exiled",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					ids := legalTargetCardIDs(ctx)
					if len(ids) == 0 {
						return nil
					}
					if z := g.FindCardZoneForEffect(ids[0]); z == nil || z.Kind != game.ZoneGraveyard {
						return nil
					}
					return g.ExileCardWithPermissionForEffect(ids[0], game.CastPermission{
						Player:     item.Controller,
						Duration:   game.WhileInZoneDuration(),
						Source:     item.SourceCardID,
						SourceName: "Savvy Trader",
						Label:      "Savvy Trader — you may play this card for as long as it remains exiled",
					})
				}),
				TargetCardInGraveyard("target permanent card from your graveyard", YouOwn(), Permanent())),
		},
	})
}
