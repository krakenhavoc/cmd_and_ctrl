package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ophidian Eye — Enchantment — Aura for {2}{U} (EDHREC rank 1629):
//
//	"Flash
//	 Enchant creature
//	 Whenever enchanted creature deals damage to an opponent, you may
//	 draw a card."
//
// Curiosity with flash, two mana more, and the flash is the whole
// reason to run both: the Eye goes on a creature AFTER blockers are
// declared, or on an opponent's unblocked attacker, or in response to
// the removal spell that was going to kill the Curiosity'd creature.
//
// FLASH IS A PRINTED KEYWORD, NOT A GRANT, so it rides
// Spec.PrintedKeywords and the engine's cast-timing gate reads it
// through the off-battlefield HasKeyword fallback — which is the
// branch that exists precisely so a card still in hand can be cast at
// instant speed. Nothing else about the card differs from Curiosity,
// and the shared trigger condition is the same function.
//
// One combination worth knowing, because the card is played for it:
// an Ophidian Eye on a creature that already has Curiosity, with a
// Niv-Mizzet on the board, is the classic infinite. Neither piece is
// in this catalog yet; the Eye is not the half that would make it
// work incorrectly.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e83b8830-fa97-4e60-b042-005e552a3ceb",
		Name:            "Ophidian Eye",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtDamageToOpponent(ev, source, g)
			}, "Ophidian Eye — draw a card", Do(DrawCards{N: 1})), "Ophidian Eye — draw a card?"),
		},
	})
}
