package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// MACH-1, Swooping Scoundrel — Legendary Artifact Creature — Human
// Villain {1}{B}, 1/3 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 When MACH-1 enters and whenever you gain life, surveil 1. This
//	 ability triggers only once each turn."
//
// One ability watching two events, with the once-each-turn limit read
// off the turn's trigger tally for this ability (the same limit Ant-Man
// uses), so an enters trigger and a later life gain in one turn surveil
// once, and a second life gain does nothing. The limit counts the
// trigger, not the resolution.
//
// No simplification.
func init() {
	const label = "MACH-1 — surveil 1 (triggers only once each turn)"
	Register(Spec{
		OracleID:        "75810ce7-328c-4512-8e36-fd3badbd3997",
		Name:            "MACH-1, Swooping Scoundrel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventChangeLife},
				func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
					if b11TriggeredThisTurn(g, source.InstanceID, label) {
						return false
					}
					if ev.Kind == game.EventETB {
						return Self(ev, source, ch, g)
					}
					return YouGainedLife(ev, source, ch, g)
				}, label, Do(Surveil{N: 1})),
		},
	})
}
