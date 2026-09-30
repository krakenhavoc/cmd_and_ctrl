package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frostcliff Siege — Enchantment {1}{U}{R}:
//
//	"As this enchantment enters, choose Jeskai or Temur.
//	 • Jeskai — Whenever one or more creatures you control deal combat
//	   damage to a player, draw a card.
//	 • Temur — Creatures you control get +1/+0 and have trample and
//	   haste."
//
// The proof card for #1572, the CR 614.12 anchor-word choice. Each
// bullet is an ordinary ability carrying the ADR 0071 gate for its
// word, so a Jeskai Siege has the draw trigger and nothing else and a
// Temur Siege has the anthem and nothing else — and before the
// controller answers, neither.
//
//   - Jeskai is the shared "one or more creatures … deal combat damage
//     to a player" constructor: one draw per player connected with per
//     damage step (CR 603.2c), never one per creature.
//   - Temur is three layer statics with one gate each: +1/+0 in layer
//     7c, trample and haste in layer 6, all over "creatures you
//     control". The answer landing bumps the layer version
//     (EventOptionChosen), so a creature already on the battlefield
//     gets the bonus the moment Temur is chosen.
//
// No simplification.
func init() {
	yours := TribeFilter{YoursOnly: true}
	Register(Spec{
		OracleID:     "e7af31ff-fa1b-4225-b3bc-fde80ac72d0a",
		Name:         "Frostcliff Siege",
		Completeness: CompletenessFull,
		AsEnters:     ChooseOptionAsEnters("Frostcliff Siege", "Jeskai", "Temur"),
		Triggered: []game.TriggeredAbility{
			WhenChosen("Jeskai", WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil,
				"Frostcliff Siege — draw a card", Do(DrawCards{N: 1}))),
		},
		Static: []game.StaticAbility{
			StaticWhenChosen("Temur", TribalAnthem(yours, 1, 0)),
			StaticWhenChosen("Temur", TribalKeywordGrant(yours, "trample")),
			StaticWhenChosen("Temur", TribalKeywordGrant(yours, "haste")),
		},
	})
}
