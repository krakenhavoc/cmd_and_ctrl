package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shallow Grave — Instant {1}{B}:
//
//	"Return the top creature card of your graveyard to the battlefield.
//	 That creature gains haste until end of turn. Exile it at the
//	 beginning of the next end step."
//
// The graveyard keeps its order (CR 404.2; the top is the last card in
// the zone), and the spell takes the topmost CREATURE card, which need
// not be the top card. No target: the card is found as the spell
// resolves, so a card milled or discarded in response is the one that
// comes back. Whip of Erebos's haste and delayed exile follow, without
// Whip's "if it would leave, exile it instead", which Shallow Grave
// does not print.
//
// No simplification. Mass Casualty deck request (#3245).
func init() {
	Register(Spec{
		OracleID:     "9843d8a6-bcbf-40c2-8c41-316e97db9c62",
		Name:         "Shallow Grave",
		Completeness: CompletenessFull,
		OnResolve:    shallowGraveReturn,
	})
}

func shallowGraveReturn(item *game.StackItem, ctx *Context) error {
	p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
	if p == nil {
		return nil
	}
	var id uuid.UUID
	for i := len(p.Graveyard.Cards) - 1; i >= 0; i-- {
		if p.Graveyard.Cards[i].IsCreature() {
			id = p.Graveyard.Cards[i].InstanceID
			break
		}
	}
	if id == uuid.Nil {
		return nil
	}
	if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
		return err
	}
	if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	if err := (GrantKeywordUntilEOT{
		Target:   id,
		Keywords: []string{"haste"},
		Label:    "Shallow Grave — the returned creature has haste",
	}).Apply(ctx); err != nil {
		return err
	}
	return ScheduleDelayedTrigger{
		Label: "Shallow Grave — exile the returned creature",
		Cards: []uuid.UUID{id},
		Body:  exileListedCardsBody,
	}.Apply(ctx)
}
