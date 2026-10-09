package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Saheeli, Consul of Oversight — Legendary Creature — Human Advisor
// {3}{W}{W}, 4/4:
//
//	"Flying
//	 Whenever you scry or surveil, create a 1/1 colorless Thopter
//	 artifact creature token with flying. This ability triggers only
//	 once each turn."
//
// One ability watching EventScry and EventSurveil. "Only once each turn"
// is the per-turn trigger tally (b11TriggeredThisTurn), which counts a
// trigger that was countered, as printed.
//
// No simplification.
func init() {
	const label = "Saheeli, Consul of Oversight — create a 1/1 Thopter with flying"
	Register(Spec{
		OracleID:        "dfa95c0e-393d-4b22-9de3-45a7efc0ba14",
		Name:            "Saheeli, Consul of Oversight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventScry, game.EventSurveil},
				func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return ByYou(ev, source, lki, g) && !b11TriggeredThisTurn(g, source.InstanceID, label)
				}, label,
				Do(CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 1})),
		},
	})
}
