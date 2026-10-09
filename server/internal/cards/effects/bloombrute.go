package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloombrute — Creature — Plant Elemental {2}{G}{W}, 4/4:
//
//	"Whenever you gain life, draw a card. This ability triggers only
//	 once each turn.
//	 {4}{G}{W}: Target creature gains trample and lifelink until end of
//	 turn."
//
// The once-per-turn limit reads the per-turn trigger tally, which
// counts a trigger the moment it is queued, so a second life gain in
// the same turn is declined even while the first draw is still on the
// stack.
//
// No simplification.
func init() {
	const label = "Bloombrute — draw a card"
	Register(Spec{
		OracleID:     "b7040175-7411-43a8-9b46-dba1d7c2135d",
		Name:         "Bloombrute",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventChangeLife, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return YouGainedLife(ev, source, lki, g) && !b11TriggeredThisTurn(g, source.InstanceID, label)
			}, label, Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}{G}{W}: Target creature gains trample and lifelink until end of turn",
			Cost:    ManaCost("{4}{G}{W}"),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				return GrantKeywordUntilEOT{
					Target:   ts[0].ID,
					Keywords: []string{"trample", "lifelink"},
					Label:    "Bloombrute — trample and lifelink until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
