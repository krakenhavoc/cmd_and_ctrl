package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Charmbreaker Devils — #4380, S76. Uses the existing catalog seams.
func init() {
	Register(Spec{
		OracleID: "1b9df437-6988-4ddc-80c4-893e11076067", Name: "Charmbreaker Devils", Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Charmbreaker Devils — return a random instant or sorcery", returnRandomInstantOrSorceryFromYourGraveyard),
			On(game.EventCast, YouCast(Or(Instant(), Sorcery())), "Charmbreaker Devils — +4/+0", func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{Target: item.SourceCardID, Power: 4, Label: "Charmbreaker Devils — +4/+0"}.Apply(NewContext(g, item))
			}),
		},
	})
}

// returnRandomInstantOrSorceryFromYourGraveyard is shared with Exalted
// Flamer of Tzeentch: no target or choice, one uniformly random eligible card.
func returnRandomInstantOrSorceryFromYourGraveyard(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := graveyardIDs(g, item.Controller, func(c game.Card) bool { return c.IsInstant() || c.IsSorcery() })
	picked := randomPick(ctx, ids, 1)
	if len(picked) == 0 {
		return nil
	}
	return g.ReturnFromGraveyardForEffect(picked[0], game.ZoneHand)
}
