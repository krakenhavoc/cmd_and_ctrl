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
// A COMMANDER is exiled like any other card (CR 903.9a, ADR 0115):
// the exile lands, its triggers fire, and only then is its owner
// offered the command zone, so the replacement never competes with the
// commander rule and no ordering prompt is asked.
func init() {
	Register(Spec{
		OracleID:     "087f9ad7-e74f-40e2-8102-1ed2925d0418",
		Name:         "Rest in Peace",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{GraveyardBecomesExile{
			Label: "Rest in Peace: exile instead of a graveyard",
		}.Build()},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Key:    "Rest in Peace — exile all graveyards",
			Effect: b02ExileAllGraveyards,
		}},
	})
}
