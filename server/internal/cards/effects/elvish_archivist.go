package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elvish Archivist — Creature {1}{G}, 1/2 Elf Artificer (Wilds of
// Eldraine, #4471):
//
//	"Whenever one or more artifacts you control enter, put two +1/+1
//	 counters on this creature. This ability triggers only once each turn.
//	 Whenever one or more enchantments you control enter, draw a card.
//	 This ability triggers only once each turn."
//
// Two separate abilities, each Ingenious Smith's shape: a batch trigger
// (OncePerBatch, so permanents entering together are one) that also
// triggers only once each turn (TriggersOnlyOnceEachTurn, keyed by the
// ability's own label, so the artifact ability and the enchantment
// ability each get their own turn). An artifact enchantment entering
// triggers both. Each ability is checked against the permanent as it
// entered, through the layered view.
//
// No simplification.
func init() {
	const counterLabel = "Elvish Archivist — put two +1/+1 counters on this creature"
	const drawLabel = "Elvish Archivist — draw a card"
	Register(Spec{
		OracleID:     "d7035db0-4bde-4ba3-9028-dd14191c8126",
		Name:         "Elvish Archivist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggersOnlyOnceEachTurn(OncePerBatch(On(game.EventETB,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := enteredUnderYourControl(ev, source, g, false)
					return ok && c.IsArtifact()
				}, counterLabel, Do(CounterOnThis{Kind: game.CounterPlusOne, N: 2})))),
			TriggersOnlyOnceEachTurn(OncePerBatch(On(game.EventETB,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := enteredUnderYourControl(ev, source, g, false)
					return ok && c.IsEnchantment()
				}, drawLabel, Do(DrawCards{N: 1})))),
		},
	})
}
