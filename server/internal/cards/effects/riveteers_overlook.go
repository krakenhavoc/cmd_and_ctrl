package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
// Two stack items, as printed (#636). The entry trigger sacrifices
// the land; "when you do" is a CR 603.12 reflexive trigger that
// carries the search and the life, so the table gets a response
// window between the sacrifice and the fetch. The condition is real
// in both directions: an Overlook bounced in response to the entry
// trigger is not on the battlefield to be sacrificed, so no reflexive
// trigger is created and nothing is searched.
//
// The whole card is b08OverlookSacrifice, the shape four more New
// Capenna lands print; this one keeps its own spec only because it
// names its three basics in its labels.
func init() {
	Register(Spec{
		OracleID:     "5548ff43-e5f6-4a63-8562-a2b1de06d6f5",
		Name:         "Riveteers Overlook",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, b06SelfETB, "Riveteers Overlook — sacrifice it",
				b08OverlookSacrifice(
					"Riveteers Overlook — fetch a basic Swamp, Mountain, or Forest tapped, gain 1 life",
					"Riveteers Overlook — a basic Swamp, Mountain, or Forest",
					func(c game.Card) bool {
						return IsBasicLand(c) && (c.HasSubtype("Swamp") || c.HasSubtype("Mountain") || c.HasSubtype("Forest"))
					},
				)),
		},
	})
}
