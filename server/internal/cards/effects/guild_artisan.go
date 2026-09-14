package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Guild Artisan — Legendary Enchantment — Background {1}{R} (EDHREC
// rank 3588):
//
//	"Commander creatures you own have "Whenever this creature attacks
//	 a player, if no opponent has more life than that player, you
//	 create two Treasure tokens." (They're artifacts with "{T},
//	 Sacrifice this token: Add one mana of any color.")"
//
// The Treasure Background. The printed card GRANTS a triggered
// ability to the commander; a static cannot grant one (the layer
// engine rewrites characteristics, and the harvester reads triggers
// off the catalog by oracle ID), so the ability lives on the
// Background itself — Agent of the Iron Throne's posture — gated on
// the attacker being a creature that is a commander the Background's
// controller owns and controls. The intervening-if — the attacked
// player's life is not exceeded by any opponent of the Background's
// controller — is checked when the attack is declared and again as
// the trigger resolves (CR 603.4), reading live life totals each
// time. "Attacks a player": an attack on a planeswalker or a battle
// does not fire it. Two commanders with partner each fire it. The
// Treasures are the real token, with its own mana ability. "Choose
// a Background" is a deck-construction rule (CR 702.124), the deck
// importer's business.
//
// Sandbox simplification, declared: because the trigger is on the
// Background and not on the commander, a commander an opponent has
// stolen does not carry it — the printed ability would then work for
// the thief. Weaker, never stronger.
func init() {
	Register(Spec{
		OracleID:     "aceac269-5100-428a-a4a3-bac57031e30b",
		Name:         "Guild Artisan",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The Treasures come only while you control your commander — a commander an opponent has stolen doesn't carry the ability."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b34CommanderCreatureYouOwnAttackedAPlayer(ev, source, g) &&
					b34NoOpponentHasMoreLifeThan(g, source.Controller, ev.Target)
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				attacked := ev.Target
				return game.NewTriggeredItem(source, "Guild Artisan — create two Treasures",
					func(g *game.Game, item *game.StackItem) error {
						if !b34NoOpponentHasMoreLifeThan(g, item.Controller, attacked) {
							return nil
						}
						return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 2}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
