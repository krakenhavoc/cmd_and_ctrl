package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_assimilate.go — the Star Trek keyword action assimilate (#4973),
// shared by every card that assimilates from any zone. It was batch 11's
// trekB11Assimilate (a card in a graveyard, Death Is Irrelevant); Borg
// Cube assimilates a card from exile, so the body moved here with the
// zone as a parameter and trekB11Assimilate now calls it unchanged.

// trekAssimilateFrom is assimilate on one card in zone `from`: "Put it
// onto the battlefield under your control with a +1/+1 counter. It's a
// Borg artifact creature and loses all other creature types." CR 205.1b:
// a "[creature type] artifact creature" keeps its other card types and
// its subtypes other than creature types, so the put adds Artifact and
// replaces the creature types with Borg, in one record registered as the
// permanent lands (ADR 0210). The counter is an entry counter, so the
// creature enters with it. `tail` (zero for none) is told the permanent
// that entered as TailOutcome.IDs, empty when nothing did.
func trekAssimilateFrom(ctx *Context, from game.ZoneKind, card uuid.UUID, label string, tail game.TailRef) error {
	return putCardsFromZoneTogetherWith(ctx, from, []uuid.UUID{card},
		game.ZoneEntryOptions{
			Controller:         ctx.Controller(),
			EntersWithCounters: map[string]int{game.CounterPlusOne: 1},
			EntersAs: &game.EntersAs{
				AddTypes:      []string{"Artifact"},
				CreatureTypes: ctx.TextSubtypes("Borg"),
				Label:         label,
			},
		}, tail, 0)
}
