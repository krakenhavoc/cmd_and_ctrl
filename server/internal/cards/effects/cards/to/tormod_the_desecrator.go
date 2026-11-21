package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tormod, the Desecrator — Legendary Creature — Zombie Wizard {3}{B},
// 4/2 (EDHREC rank 3623):
//
//	"Whenever one or more cards leave your graveyard, create a tapped
//	 2/2 black Zombie creature token.
//	 Partner (You can have two commanders if both have partner.)"
//
// The graveyard-recursion partner. The trigger is Teval's Judgment's
// condition (cardLeftYourGraveyard — a move out of the
// controller's own graveyard, or a flashback cast from it) with the
// per-label "one or more" dedup, so an escape that exiles five cards
// makes one Zombie and a single regrowth makes one; the token is
// Overseer of the Damned's tapped 2/2 black Zombie. Partner is a
// deck-construction rule (CR 702.124), the deck importer's business.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6dc0150d-7145-41e2-bd3d-2564d9d32301",
		Name:         "Tormod, the Desecrator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.OnAny([]game.EventKind{game.EventZoneMove, game.EventCast}, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.CardLeftYourGraveyard(ev, source, g)
			}, b34TormodLabel, effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.BlackZombieToken()).EntersTapped(), N: 1}))),
		},
	})
}

// b34TormodLabel is the stack label of Tormod's trigger — the "one or
// more" dedup keys on it.
const b34TormodLabel = "Tormod, the Desecrator — create a tapped 2/2 black Zombie"
