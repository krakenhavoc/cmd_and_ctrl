package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nimrodel Watcher — Creature — Elf Scout {1}{U}, 2/1 (The Lord of the
// Rings: Tales of Middle-earth, in The Hobbit's deck):
//
//	"Whenever you scry, this creature gets +1/+0 until end of turn and
//	 can't be blocked this turn. This ability triggers only once each
//	 turn."
//
// Prudent Fateseer's scry trigger with its once-a-turn limit (the tally
// counts an ability the moment it is put on the stack, so a second scry the
// same turn is declined even while the first is unresolved). The pump and
// the evasion are both the Watcher's own, applied only while it is still on
// the battlefield.
//
// No simplification.
func init() {
	const label = "Nimrodel Watcher — +1/+0 and can't be blocked this turn"
	Register(Spec{
		OracleID:     "1b8de1c5-1e23-4c84-bc5a-a14853ad8c7a",
		Name:         "Nimrodel Watcher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventScry,
				func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return ByYou(ev, source, lki, g) && !b11TriggeredThisTurn(g, source.InstanceID, label)
				},
				label,
				func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					ctx := NewContext(g, item)
					if err := (BoostUntilEOT{Target: item.SourceCardID, Power: 1, Label: label}).Apply(ctx); err != nil {
						return err
					}
					return RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked, Label: label}.Apply(ctx)
				}),
		},
	})
}
