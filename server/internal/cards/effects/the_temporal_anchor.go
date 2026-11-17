package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Temporal Anchor — Legendary Artifact {3}{U}{U}{U} (The Brothers' War,
// S82):
//
//	"At the beginning of your upkeep, scry 2.
//	 Whenever you choose to put one or more cards on the bottom of your
//	 library while scrying, exile that many cards from the bottom of your
//	 library.
//	 During your turn, you may play cards exiled with The Temporal Anchor."
//
// The scry is a plain upkeep trigger. The second ability watches the end of
// any scry of yours (EventScry, whose Amount is how many cards the player put
// on the bottom, and every card a scry puts on the bottom is one it chose to
// put there); it exiles that many cards from the bottom of your library, a
// linked exile each (CR 607.2a), after the scry has put its own cards there.
// A library with fewer cards exiles what it has. The permission is Valgavoth's
// shape: a standing play permission over the cards exiled with this object, in
// your turn only, the cards' own timing and costs in force (lands included, so
// a land spends the land drop). It is not an alternative cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "168f2fbb-8cf7-43e0-8bb6-5e4cc3c1e363",
		Name:         "The Temporal Anchor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("The Temporal Anchor — scry 2", Do(Scry{N: 2})),
			On(game.EventScry, AllOf(ByYou, temporalAnchorBottomedCards),
				"The Temporal Anchor — exile that many cards from the bottom of your library",
				temporalAnchorExileBottom),
		},
		CastPermissions: []game.CastPermission{{
			Zone:   game.ZoneExile,
			Filter: game.PermissionFilter{ExiledWithSource: true},
			Timing: game.TimingYourTurnOnly,
			Label:  "During your turn, you may play cards exiled with The Temporal Anchor",
		}},
	})
}

// temporalAnchorBottomedCards is "you choose to put one or more cards on the
// bottom of your library while scrying": the scry finished with some cards put
// on the bottom.
func temporalAnchorBottomedCards(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Amount > 0
}

// temporalAnchorExileBottom exiles that many cards from the bottom of the
// controller's library, each linked to this Anchor.
func temporalAnchorExileBottom(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := item.Trigger.Event.Amount
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Library == nil {
		return nil
	}
	// Index 0 is the bottom (Zone.Bottom): take the ids first, so a card
	// that is exiled does not shift the ones still to be.
	var ids []uuid.UUID
	for i := 0; i < n && i < len(p.Library.Cards); i++ {
		ids = append(ids, p.Library.Cards[i].InstanceID)
	}
	for _, id := range ids {
		if err := (ExileTarget{Target: id, Linked: true}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
