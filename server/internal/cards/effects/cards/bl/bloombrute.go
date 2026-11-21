package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "b7040175-7411-43a8-9b46-dba1d7c2135d",
		Name:         "Bloombrute",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventChangeLife, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.YouGainedLife(ev, source, lki, g) && !effects.B11TriggeredThisTurn(g, source.InstanceID, label)
			}, label, effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{G}{W}: Target creature gains trample and lifelink until end of turn",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "trample", "lifelink")),
			Cost:    effects.ManaCost("{4}{G}{W}"),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				return effects.GrantKeywordUntilEOT{
					Target:   ts[0].ID,
					Keywords: []string{"trample", "lifelink"},
					Label:    "Bloombrute — trample and lifelink until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
