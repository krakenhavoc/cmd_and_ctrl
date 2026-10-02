package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Veiled Serpent — Enchantment {2}{U}:
//
//	"When an opponent casts a spell, if this permanent is an enchantment,
//	 it becomes a 4/4 Serpent creature with "This creature can't attack
//	 unless defending player controls an Island."
//	 Cycling {2}"
//
// The trigger has an intervening "if" (CR 603.4): it triggers only while
// this is an enchantment, and checks again as it resolves. "Becomes a …
// creature" sets the card type (CR 205.1a), so the Serpent is no longer
// an enchantment and never triggers again. The quoted ability is #1879's
// restriction (ADR 0107 §2), gained in the same effect as the type change
// (a cantAttackUnlessDefenderControls mod), so it lasts as long as the
// Serpent is this object and a later "loses all abilities" takes it
// (CR 613.1f). It stays blue.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ff59a95c-28de-44cf-abbe-772417851ffa",
		Name:         "Veiled Serpent",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return source.HasCardType("enchantment") && AnOpponentCast(nil)(ev, source, ch, g)
			}, "Veiled Serpent — it becomes a 4/4 Serpent creature",
				thisEnchantmentBecomesACreature("Serpent", 4, 4, "Veiled Serpent — a 4/4 Serpent creature",
					game.CantAttackUnlessDefenderControlsMod(QuerySubtype("Island")))),
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
