package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alela, Cunning Conqueror — Legendary Creature — Faerie Warlock
// {2}{U}{B}, 2/4 (EDHREC rank 3466):
//
//	"Flying
//	 Whenever you cast your first spell during each opponent's turn,
//	 create a 1/1 black Faerie Rogue creature token with flying.
//	 Whenever one or more Faeries you control deal combat damage to a
//	 player, goad target creature that player controls."
//
// The flash deck's commander: a Faerie for every opponent's turn you
// act on, and the flock turns the table against whoever it hits.
// The cast trigger is Wavebreak Hippocamp's condition
// (b22FirstSpellOnAnOpponentsTurn) making the Faerie Rogue. The
// damage trigger is "one or more … to a player": the engine emits
// one EventDealDamage per creature, so the first Faerie to connect
// with a given player fires it and the rest of that damage step's
// Faeries aimed at the SAME player are declined as later events of
// the same batch, while a Faerie connecting with a second player is
// its own occurrence — OncePerBatchPerPlayer, the CR 603.2c key with
// its player dimension (#784; see AGENTS.md §7). That is what makes
// the printed per-player target clause work: each hit player's
// trigger goads a creature THAT player controls. Alela herself is a
// Faerie and counts.
//
// The goad is the engine's goad. The sandbox has carried a goad
// marker since S10 — the client badges the creature and the context
// menu clears it — and enforces the must-attack constraint for
// nobody, manual goads included; this trigger stamps that marker,
// and a delayed trigger clears it at the beginning of the
// controller's next turn, which is "until your next turn". The
// target clause is Trygon Predator's shape: a target predicate is
// not handed the trigger's event, so it admits any creature of a
// player one of the controller's Faeries dealt combat damage to this
// turn, and the body re-checks the specific player the firing Faerie
// hit — a pick under some other hit player's control does nothing.
// That set is the per-turn tally's, recorded at the damage (#1009):
// the Faerie Rogue tokens Alela makes are the commonest thing to
// connect and then trade, and CR 704.5d takes a dead token out of the
// graveyard, so a set computed later by looking the dealer up would
// lose exactly the player the goad is for.
//
// Sandbox simplification, declared: the goad is a marker the engine
// shows and does not enforce — the goaded creature is not made to
// attack, and is not stopped from attacking Alela's controller.
// Weaker than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:        "1cae5752-b4af-4a8f-8c8c-2493e163083b",
		Name:            "Alela, Cunning Conqueror",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"Goad only marks the creature — the game doesn't force it to attack, or stop it attacking you."},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b22FirstSpellOnAnOpponentsTurn(ev, source, g)
			}, "Alela, Cunning Conqueror — create a 1/1 black Faerie Rogue with flying", b33CreateTokenBody(FaerieRogueToken)),
			{
				OncePerBatch: true,
				BatchKey:     PerPlayer,
				Watches:      []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b33FaerieYouControlDealtCombatDamageToPlayer(ev, source, g)
				},
				Targets: TargetCreature("target creature that player controls", b33CreatureOfPlayerHitByYourFaeries),
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, b33AlelaGoadLabel, b33GoadChosenIfControlledBy(ev.Target))
				},
			},
		},
	})
}
