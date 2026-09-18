package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aether Tunnel — Enchantment — Aura {1}{U} (EDHREC rank 4327):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+0 and can't be blocked."
//
// Two mana that makes a commander connect, which in a format where
// twenty-one damage is a clock is the only job the card has. The
// +1/+0 is rent; the unblockable is the card.
//
// "Can't be blocked" is a RESTRICTION rather than a keyword, and the
// distinction is the reason it took an extra sprint to exist: it is a
// restriction on the DEFENDER's options (CR 509.1b) carried on the
// attacker, read inside the block-pair check rather than anywhere on
// the attacking side. Menace, flying and the landwalks are keywords;
// this is not one, and RestrictAttached is how an attachment says so.
//
// Both halves ride AttachedToSource, which re-reads the attachment on
// every recompute — destroy the Tunnel and the creature is blockable
// again in the same pass, with no bookkeeping.
//
// The Aura's enchant clause is an ordinary target spec, and the same
// spec is what the CR 704.5m legality check re-runs every turn: a host
// that stops being a creature takes the Tunnel to the graveyard with
// it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0a90d1fb-7a76-4b62-b001-4a539a8e5df1",
		Name:         "Aether Tunnel",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(1, 0),
			RestrictAttached(game.CantBeBlocked),
		},
	})
}
