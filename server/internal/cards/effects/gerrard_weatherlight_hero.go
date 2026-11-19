package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gerrard, Weatherlight Hero — Legendary Creature — Human Soldier
// {2}{R}{W}, 3/3 (Mystery Booster 2):
//
//	"First strike
//	 When Gerrard dies, exile it and return to the battlefield all artifact
//	 and creature cards in your graveyard that were put there from the
//	 battlefield this turn."
//
// Gerrard is exiled from the graveyard first, so it is never one of the cards
// returned; if it has already left the graveyard it is not exiled and the
// return still happens. The returned set is read as the ability resolves: the
// artifact and creature cards in the controller's graveyard whose latest
// arrival was from the battlefield this turn (Samwise the Stouthearted's
// reading of "put there from the battlefield this turn"), so a card that was
// milled, discarded or died on an earlier turn stays. They all return at once
// (CR 614.12), under their owner's control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8a1e8435-16f2-475b-a8a9-7b9f0cf4594a",
		Name:            "Gerrard, Weatherlight Hero",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Gerrard, Weatherlight Hero — exile it and return the artifact and creature cards put into your graveyard from the battlefield this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if zone := g.FindCardZoneForEffect(item.SourceCardID); zone != nil && zone.Kind == game.ZoneGraveyard {
						if err := (ExileTarget{Target: item.SourceCardID}).Apply(ctx); err != nil {
							return err
						}
					}
					return ReturnFromGraveyardTogether{
						Targets: graveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
							return (c.IsCreature() || c.IsArtifact()) && putIntoGraveyardFromBattlefieldThisTurn(g, c.InstanceID)
						}),
						Controller: ctx.Controller(),
					}.Apply(ctx)
				}),
		},
	})
}
