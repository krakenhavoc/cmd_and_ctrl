package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Riveteers Overlook — Land (EDHREC rank 714):
//
//	"When this land enters, sacrifice it. When you do, search your
//	 library for a basic Swamp, Mountain, or Forest card, put it onto
//	 the battlefield tapped, then shuffle and you gain 1 life."
//
// The Streets of New Capenna "Overlook": Evolving Wilds that fires on
// entry instead of on activation, fetches only its three colours, and
// pays a life back. It is a land drop that becomes a basic of your
// choice, tapped, plus a life.
//
// Sandbox simplification: the printed card is a trigger ("when this
// enters, sacrifice it") with a REFLEXIVE trigger inside it ("when
// you do, search…"), two stack items with a response window between
// them. Here it is one item: sacrifice, then search, then gain — the
// search cannot be responded to separately from the sacrifice. Weaker
// than printed only for an opponent who wanted that second window.
// The search is the S22 chooser, so the basic is the player's pick.
//
// "When you do" is still honoured: if the Overlook is no longer on
// the battlefield when the trigger resolves (bounced in response),
// it cannot be sacrificed, so the reflexive half never happens and
// nothing is searched — as printed.
func init() {
	Register(Spec{
		OracleID:     "5548ff43-e5f6-4a63-8562-a2b1de06d6f5",
		Name:         "Riveteers Overlook",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The sacrifice and the search happen as one ability, so there is no separate chance to respond between them."},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Riveteers Overlook — sacrifice it, fetch a basic Swamp, Mountain, or Forest tapped, gain 1 life",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneBattlefield {
							return nil
						}
						if err := (SacrificePermanent{Target: item.SourceCardID}).Apply(ctx); err != nil {
							return err
						}
						source, controller := item.SourceCardID, item.Controller
						return SearchLibrary{
							Player: controller,
							Predicate: func(c game.Card) bool {
								return IsBasicLand(c) && (c.HasSubtype("Swamp") || c.HasSubtype("Mountain") || c.HasSubtype("Forest"))
							},
							Dest:          game.ZoneBattlefield,
							Limit:         1,
							Reveal:        true,
							Shuffle:       true,
							TappedOnEntry: true,
							Reason:        "Riveteers Overlook — a basic Swamp, Mountain, or Forest",
							Then: func(g *game.Game, _ []uuid.UUID) error {
								return g.ChangePlayerLifeForEffect(source, controller, 1)
							},
						}.Apply(ctx)
					})
			},
		}},
	})
}
