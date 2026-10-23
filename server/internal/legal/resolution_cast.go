package legal

import (
	"encoding/json"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_cast.go — the answers to a resolution_cast prompt (CR
// 707.12, CR 608.2g; ADR 0066's 2026-10-22 amendment, #2947).
//
// The prompt's chooser may cast each copy it still offers, during the
// resolution that made the copies. Each such cast is the ordinary cast of
// that copy out of its zone through the copy's own permission
// (castMovesFromZone, the walk the seat's priority moves use), so every
// way of casting it the engine would take is offered, each naming the
// prompt. The other answer, "cast no more of them", is always legal.

// resolutionCastExpansion is the cap on one copy's casts: every target
// and mode of the one card is a choice the caster is owed.
const resolutionCastExpansion = 200

// resolutionCastCasts is every cast of the copy `card` that chooser could
// make now, as the engine would take it. Caller must hold g.mu.
func resolutionCastCasts(g *game.Game, chooser, card uuid.UUID, opts Options) []Move {
	c, zone, ok := g.ResolutionCastCandidateLocked(chooser, card)
	if !ok {
		return nil
	}
	opts.MaxExpansionPerSource = max(opts.MaxExpansionPerSource, resolutionCastExpansion)
	e := enumeratorAs(g, chooser, opts)
	if e == nil {
		return nil
	}
	speed := g.SorcerySpeedOpenLocked(chooser)
	e.castMovesFromZone(c, zone, string(zone), speed, false, true, true)
	return e.out
}

// resolutionCastMoves answers a resolution_cast prompt c owed by e's
// seat: the casts of each copy it offers, each naming the prompt, and
// the decline.
func (e *enumerator) resolutionCastMoves(c *game.PendingChoice) {
	reason := c.Reason
	if reason == "" {
		reason = "Cast the copy"
	}
	for _, id := range c.DiscardOptions {
		for _, m := range resolutionCastCasts(e.g, c.Chooser, id, e.opts) {
			var params map[string]any
			if err := json.Unmarshal(m.Params, &params); err != nil {
				continue
			}
			params["resolution_cast"] = c.ID.String()
			m.Params = mustJSON(params)
			m.Label = reason + ": " + m.Label
			m.ResolutionCast = c.ID.String()
			e.add(m)
		}
	}
	e.add(Move{
		Type:           TypeResolveChoice,
		Player:         e.seat,
		Kind:           KindChoice,
		Label:          reason + ": don't cast",
		Source:         c.Source,
		Params:         mustJSON(choiceParams{ChoiceID: c.ID.String(), CardIDs: []string{}}),
		AlwaysLegal:    true,
		ResolutionCast: c.ID.String(),
	})
}
