package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_replacements.go — replacements of the act of COUNTERING a spell
// (CR 701.6a, CR 614.1a; RepEventCounterSpell, #4056, ADR 0113's
// 2026-10-19 amendment).
//
// The event is opened by the effect counter verbs (the hard counter, a
// declined "counter unless pays", a ward, and the destination counters
// Remand, Memory Lapse and Hinder) after the "can't be countered" gate
// and before the spell leaves the stack. It names who is countering:
// ev.Actor / ev.CounterBy is the controller of the countering spell or
// ability, so "you control" is judged against it and not against the
// permanent's own controller's spells in general.

// ExileSpellsYouWouldCounterInstead is "If a spell or ability you
// control would counter a spell, instead exile that spell and you may
// play that card without paying its mana cost" (Guile). Declare it on
// Spec.Replacements.
//
//   - The spell is not countered (CR 614.1a: the replaced event does not
//     happen). A spell that can't be countered never reaches the window,
//     so it is not exiled either; nothing keyed to "countered" fires.
//   - "You control" is the countering spell's or ability's controller,
//     which is the resolving item's, so a stolen permanent's ability is
//     its thief's. An opponent's counterspell is untouched, and so is a
//     counter you did not cause (the sandbox's counter action).
//   - Only a SPELL is replaced. A countered ability is not covered.
//   - The play is the countering player's, not the spell owner's, and is
//     made during the resolution of the countering spell or ability
//     (CR 608.2g, #4286): the exile is followed by the resolution_cast
//     prompt over the exiled card (game.OfferResolutionCastForEffect), so
//     the cast happens before the counterspell finishes resolving and
//     nobody receives priority in between. Declining leaves the card in
//     exile. The permission behind it is game.FreeCastFromExilePermission:
//     free, cast-only (a spell is never a land) and X is 0. It is offered
//     only if the card really landed in exile: a commander spell whose
//     owner took the command zone (CR 903.9) is not. Only a counter the
//     engine cannot ask the prompt for (a cause carried past the end of its
//     resolution, #4279) falls back to the same permission granted alone,
//     any timing and closing on that player's next priority pass.
//   - Two of these on one table (two Guiles) are one replacement: the
//     first turns the event into an exile, and the second no longer has
//     a counter to replace (CR 616.1f).
func ExileSpellsYouWouldCounterInstead() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventCounterSpell},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return ev.CounterSpellStillCounters() && src != nil &&
				ev.CounterBy != uuid.Nil && ev.CounterBy == src.Controller
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
			return g.ExileCounteredSpellInsteadForEffect(ev, ev.CounterBy, src.InstanceID, src.Name)
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: "Exile a spell you would counter instead, and you may play it for free",
	}
}
