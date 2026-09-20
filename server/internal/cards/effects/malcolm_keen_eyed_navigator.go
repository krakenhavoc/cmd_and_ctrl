package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Malcolm, Keen-Eyed Navigator — 2/2 Legendary Creature — Siren
// Pirate for {2}{U}:
//
//	"Flying
//	 Whenever one or more Pirates you control deal damage to your
//	 opponents, you create a Treasure token for each opponent dealt
//	 damage.
//	 Partner"
//
// Breeches, Brazen Plunderer's trigger paying out in mana instead of
// cards — the ramp half of the same attack. Both partners in the
// same deck is the intended line.
//
// "One or more … deal damage to your opponents" — no "combat" in
// the printed text, so a Pirate's non-combat damage counts too:
// damagedOpponentByAnyDamage, not damagedOpponent.
//
// It is also CR 603.2c's per-player collapse (#784):
// OncePerBatchPerPlayer fires this trigger at most once per opponent
// per damage batch, matching the printed "for each opponent dealt
// damage" — the count is opponents damaged, not Pirates that dealt
// the damage. Three Pirates into three DIFFERENT opponents still
// makes three Treasures (one trigger per opponent, one Treasure
// each), and three Pirates into the SAME opponent now makes one, not
// three.
//
// Partner is a deck-construction rule, not a game action, and is
// not modelled.
func init() {
	Register(Spec{
		OracleID:        "a66f8b44-0163-4456-b152-4acefab896a4",
		Name:            "Malcolm, Keen-Eyed Navigator",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"Partner isn't supported, so Malcolm can't be your commander."},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			OncePerBatchPerPlayer(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if damagedOpponentByAnyDamage(ev, source.Controller, g) == uuid.Nil {
					return false
				}
				dealer, ok := g.LookupCardForEffect(ev.Source)
				return ok && isPirate(dealer)
			}, "Malcolm — create a Treasure", Do(CreateToken{
				Template: TreasureToken(),
				N:        1,
			}))),
		},
	})
}
