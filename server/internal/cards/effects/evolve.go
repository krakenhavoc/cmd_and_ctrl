package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// evolve.go — the catalog's vocabulary for evolve (CR 702.100, #1805,
// ADR 0106 §3). Append-only.
//
// Evolve itself needs nothing here: it is a canonical keyword token
// (game.KeywordEvolve) that the engine turns into one trigger per
// instance (game/evolve.go). A card that prints it declares
// PrintedKeywords: []string{game.KeywordEvolve}; a card that grants it
// is an ordinary layer-6 KeywordGrant. Never write an evolve trigger by
// hand.
//
// What does live here is the reader for CR 702.100b's "evolves".

// ThisEvolved is "this creature evolves" (CR 702.100b): one or more
// +1/+1 counters were put on the source as a result of its own evolve
// ability resolving. The engine emits game.EventEvolved only when a
// counter actually landed.
func ThisEvolved(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventEvolved && ev.CardID == source.InstanceID
}

// WhenThisEvolves — "Whenever this creature evolves, …" (Renegade
// Krasis, Watchful Radstag).
func WhenThisEvolves(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventEvolved, ThisEvolved, label, effect)
}
