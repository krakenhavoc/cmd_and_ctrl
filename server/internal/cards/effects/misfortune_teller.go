package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Misfortune Teller — Creature — Human Warlock {3}{B}, 3/1 (Outlaws of
// Thunder Junction Commander, #5368):
//
//	"Deathtouch
//	 Whenever this creature enters or deals combat damage to a player,
//	 exile target card from a graveyard. If it was a creature card, create
//	 a 2/2 black Rogue creature token. If it was a land card, create a
//	 Treasure token. Otherwise, you gain 3 life."
//
// One ability over two events (Aerial Extortionist's condition). The target
// is any card in any graveyard. What the card was is asked before it moves
// (CR 608.2h), and the answer rides the exile's keyed tail as two bits,
// creature and land, so a creature land makes both tokens and "otherwise"
// is the card being neither. A card that never reached exile (a replacement
// or a cancelled move) gives nothing: "it" has no referent (CR 614.10,
// ExileThenIfItWas's reading).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bd1ccc6f-d281-413e-889e-1e022c3f2a99",
		Name:            "Misfortune Teller",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB, game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b26SelfEnteredOrDealtCombatDamageToPlayer(ev, source, g)
			},
			Targets: TargetCardInGraveyard("target card from a graveyard"),
			Key:     "Misfortune Teller — exile target card from a graveyard; Rogue, Treasure or 3 life",
			Effect:  cmdr2S33MisfortuneTellerExile,
		}},
	})
}
