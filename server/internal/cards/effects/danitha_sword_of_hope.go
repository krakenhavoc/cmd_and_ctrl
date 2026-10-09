package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Danitha, Sword of Hope — Legendary Creature — Human Knight {2}{W}, 2/2:
//
//	"First strike
//	 Whenever you cast an Equipment spell or a spell that targets a
//	 creature you control, draw a card. This ability triggers only
//	 once each turn."
//
// The once-per-turn limit reads the per-turn trigger tally, which
// counts a trigger the moment it is queued.
//
// No simplification.
func init() {
	const label = "Danitha, Sword of Hope — draw a card"
	Register(Spec{
		OracleID:        "f7b754d3-9863-420d-895d-00e70b8b078a",
		Name:            "Danitha, Sword of Hope",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller || b11TriggeredThisTurn(g, source.InstanceID, label) {
					return false
				}
				if spell, ok := g.LookupCardForEffect(ev.CardID); ok && spell.HasSubtype("Equipment") {
					return true
				}
				return rfCreatureACastSpellTargetsAny(g, ev.CardID, func(t game.TargetRef) bool {
					c, ok := rfCreatureATargetIsCreatureOnBattlefield(g, t)
					return ok && c.Controller == source.Controller
				})
			}, label, Do(DrawCards{N: 1})),
		},
	})
}
