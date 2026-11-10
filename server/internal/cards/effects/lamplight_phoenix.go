package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lamplight Phoenix — Creature — Phoenix {1}{R}{R}, 3/3 (Murders at
// Karlov Manor, #4469):
//
//	"Flying
//	 When this creature dies, you may exile it and collect evidence 4. If
//	 you do, return this card to the battlefield tapped."
//
// One decision: "exile it and collect evidence 4" is done together or not
// at all. The offer is Izoni's collect evidence prompt over the graveyard,
// with the Phoenix itself left out of it (the printed ruling: it can't pay
// for its own collect evidence), so a graveyard whose OTHER cards total
// under 4 asks nothing (CR 701.59b); choosing none declines and the
// Phoenix stays put. On a pick the chosen cards and the Phoenix are
// exiled, "whenever you collect evidence" sees the collection, and the
// card returns tapped as a new object. A Phoenix that has left the
// graveyard by resolution (exiled in response) can't be exiled, so nothing
// is asked or done.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c6f7efad-381c-45ba-9976-159b7e78f7ed",
		Name:            "Lamplight Phoenix",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Lamplight Phoenix — you may exile it and collect evidence 4; if you do, return it to the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					return stdB41PhoenixAsk(NewContext(g, item))
				}),
		},
	})
}

// stdB41PhoenixCollect is "exile it and collect evidence 4. If you do,
// return this card to the battlefield tapped". An on-disk key: never
// renamed, never reused.
var stdB41PhoenixCollect = RegisterCollectEvidence("choose-cards/lamplight-phoenix-collect-evidence", 4,
	func(ctx *Context, _ []uuid.UUID) error {
		id := ctx.Item.SourceCardID
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			return nil
		}
		if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
		return ReturnFromExile{Target: id, Tapped: true}.Apply(ctx)
	})

// stdB41PhoenixAsk offers the collection over the graveyard without the
// Phoenix, when the Phoenix is still there to exile.
func stdB41PhoenixAsk(ctx *Context) error {
	id := ctx.Item.SourceCardID
	if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return stdB41PhoenixCollect.AskExcluding(ctx, id,
		"Lamplight Phoenix — exile it and collect evidence 4 (exile cards with total mana value 4 or more from your graveyard) to return it tapped? Choose none to decline.")
}
