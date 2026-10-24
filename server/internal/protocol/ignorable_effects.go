package protocol

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// IgnorableEffectView is one CR 116.2d special action a seat may take:
// ignore a static ability's effect until end of turn by sacrificing a
// permanent (Damping Engine, #1986; game/ignore_effect.go). A row on the
// seat's own PlayerView, fired as `special_action {kind: "ignore_effect",
// card_id: source_id, sacrifice_id: <one of sacrifice_ids>}`. The seat
// chooses the permanent, so the client asks (the sacrifice picker); the
// server never picks for a human.
type IgnorableEffectView struct {
	// SourceID is the permanent whose effect it is; the action's card_id.
	SourceID string `json:"source_id"`
	// SourceName is that permanent's name.
	SourceName string `json:"source_name"`
	// Label is the row as the card reads it ("Sacrifice a permanent:
	// ignore Damping Engine's effect until end of turn").
	Label string `json:"label"`
	// SacrificeIDs are the permanents the seat may give up, in the order
	// a "choose for me" would take them. Never empty.
	SacrificeIDs []string `json:"sacrifice_ids"`
}

// viewOfIgnorableEffects projects the seat's CR 116.2d offers from the
// engine's own accessor, so a row the client shows is a row the engine
// would accept and the bot's move names the same permanent. Nil for every
// seat in a game without such a permanent. Runs under the read lock
// ViewOfGame already holds.
func viewOfIgnorableEffects(g *game.Game, seat uuid.UUID) []IgnorableEffectView {
	var out []IgnorableEffectView
	for _, offer := range g.IgnoreEffectOffersLocked(seat) {
		ids := make([]string, len(offer.Candidates))
		for i, id := range offer.Candidates {
			ids[i] = id.String()
		}
		out = append(out, IgnorableEffectView{
			SourceID:     offer.Source.String(),
			SourceName:   offer.SourceName,
			Label:        offer.Label,
			SacrificeIDs: ids,
		})
	}
	return out
}
