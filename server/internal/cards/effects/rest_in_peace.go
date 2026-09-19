package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rest in Peace — Enchantment {1}{W} (EDHREC rank 2298):
//
//	"When this enchantment enters, exile all graveyards.
//	 If a card or token would be put into a graveyard from anywhere,
//	 exile it instead."
//
// The format's hardest graveyard hate: not a one-shot sweep like
// Bojuka Bog but a standing rewrite, and it is symmetrical — the
// controller's own graveyard is gone too, which is why it is a
// sideboard card in every deck that does not want one.
//
// The ETB is the sweep Bojuka Bog and Farewell already share
// (b02ExileAllGraveyards, every seat including the controller). It
// targets nothing, so it goes on the stack as a plain triggered
// ability.
//
// The static is the CR 614 replacement, and "from anywhere" is the
// whole card: every route into a graveyard has to open the window for
// it, which is what #931 finished. A creature that would DIE is exiled
// instead, so it never died and its dies-triggers never fire
// (CR 700.4); a milled card, a discarded card, a countered spell, an
// Entomb and a surveil's graveyard leg all end in exile as well. It
// exiles ITSELF when it would be put into a graveyard — the
// replacement is gathered while the enchantment is still on the
// battlefield, which is the printed behaviour.
//
// One declared deviation, the one Liesa and Stone of Erech share and
// for the same reason: a COMMANDER whose owner takes CR 903.9's offer
// goes to the command zone rather than to exile. The built-in rewrites
// the destination first, and once it is the command zone this
// replacement no longer applies. Weaker than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:     "087f9ad7-e74f-40e2-8102-1ed2925d0418",
		Name:         "Rest in Peace",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"A dying commander whose owner sends it to the command zone goes there instead of being exiled."},
		Replacements: []game.ReplacementEffect{GraveyardBecomesExile{
			Label: "Rest in Peace: exile instead of a graveyard",
		}.Build()},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Rest in Peace — exile all graveyards",
					b02ExileAllGraveyards)
			},
		}},
	})
}
