package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Agent of the Iron Throne — Legendary Enchantment — Background {2}{B}
// (EDHREC rank 2270):
//
//	"Commander creatures you own have "Whenever an artifact or creature
//	 you control is put into a graveyard from the battlefield, each
//	 opponent loses 1 life.""
//
// The aristocrats Background. The printed card GRANTS a triggered
// ability to the commander; a static cannot grant one (the layer
// engine rewrites characteristics, and the harvester reads triggers
// off the catalog by oracle ID), so the ability lives on the
// Background itself, gated on the condition the grant implies: its
// controller controls a creature that is a commander they own
// (b21ControlsCommanderCreatureYouOwn). With no such commander on the
// battlefield the ability does not exist, as printed; with one it
// fires once per artifact or creature the controller controlled that
// went to a graveyard from the battlefield — a Treasure cracked, a
// token that died, a creature sacrificed — and each opponent loses
// 1 life (life loss, not damage). The commander's OWN death counts
// too: a leaves-the-battlefield ability looks back (CR 603.10), so
// the gate also passes when the dying card is itself a commander the
// controller owns. "Choose a Background" is a deck-construction rule
// (CR 702.124), the deck importer's business.
//
// Sandbox simplification, declared: because the trigger is on the
// Background and not on the commander, a commander an opponent has
// stolen does not carry it — the printed ability would then work for
// the thief, draining the thief's opponents. Weaker, never stronger.
// A commander whose death is replaced by the command zone (the
// engine's CR 903.9 prompt) never reaches a graveyard, so that death
// does not drain either.
func init() {
	Register(Spec{
		OracleID:     "325032d5-c452-4454-8976-82f86fee5ab8",
		Name:         "Agent of the Iron Throne",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The drain works only while you control your commander — a commander an opponent has stolen doesn't carry it."},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := b21ArtifactOrCreatureYouControlDied(ev, source, g)
				if !ok {
					return false
				}
				if dead.IsCommander && dead.Owner == source.Controller {
					return true
				}
				return b21ControlsCommanderCreatureYouOwn(g, source.Controller)
			}, "Agent of the Iron Throne — each opponent loses 1 life", func(g *game.Game, item *game.StackItem) error {
				return eachOpponentLosesLife(g, item, 1)
			}),
		},
	})
}
