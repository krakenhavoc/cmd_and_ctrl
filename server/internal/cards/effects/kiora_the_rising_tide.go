package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kiora, the Rising Tide — Legendary Creature — Merfolk Noble {2}{U},
// 3/2 (EDHREC rank 3597):
//
//	"When Kiora enters, draw two cards, then discard two cards.
//	 Threshold — Whenever Kiora attacks, if there are seven or more
//	 cards in your graveyard, you may create Scion of the Deep, a
//	 legendary 8/8 blue Octopus creature token."
//
// The Foundations self-mill commander. The entry loot is the real
// draw-then-choose-discard, so the drawn cards are legal discards
// as in paper. The attack trigger carries an intervening-if:
// threshold is checked when the attack is declared and again as the
// trigger resolves (CR 603.4), and the "you may" is the ordinary
// trigger prompt. The Scion is a legendary token, so a second one
// meets the legend rule, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5d748bce-dff8-46fa-a3d1-633863b7bbff",
		Name:         "Kiora, the Rising Tide",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: b06SelfETB,
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Kiora, the Rising Tide — draw two, then discard two",
						func(g *game.Game, item *game.StackItem) error {
							return lootOne(g, item, 2)
						})
				},
			},
			{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return attackDeclared(ev, source) && b34GraveyardHasAtLeast(g, source.Controller, 7)
				},
				OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Kiora, the Rising Tide — create Scion of the Deep, a legendary 8/8 Octopus?"},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Kiora, the Rising Tide — create Scion of the Deep", b34CreateScionIfThreshold)
				},
			},
		},
	})
}
