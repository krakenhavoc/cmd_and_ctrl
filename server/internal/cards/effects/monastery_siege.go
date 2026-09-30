package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Monastery Siege — Enchantment {2}{U}:
//
//	"As this enchantment enters, choose Khans or Dragons.
//	 • Khans — At the beginning of your draw step, draw an additional
//	   card, then discard a card.
//	 • Dragons — Spells your opponents cast that target you or a
//	   permanent you control cost {2} more to cast."
//
// #1647: the last of the three remaining Sieges.
//
//   - Khans reuses lootOne, the shared "draw, then discard" body —
//     the printed order matters (the drawn card is a legal discard),
//     and lootOne already gets it right.
//   - Dragons is a permanent-declared CostModifier that reads the
//     announced targets (TargetsYouOrYourPermanent), gated to the
//     Dragons word. It is asymmetric like Aura of Silence
//     (OpponentsSpell()): a Khans Monastery Siege taxes nobody, and a
//     Dragons one never taxes its own controller's spells.
//
// No simplification.
func init() {
	dragons := CostsMore(2,
		"Spells your opponents cast that target you or a permanent you control cost {2} more to cast.",
		OpponentsSpell(), TargetsYouOrYourPermanent())
	dragons.ReadsTargets = true
	dragons.ActiveWhen = ChosenIs("Dragons")
	Register(Spec{
		OracleID:     "240a33a8-23a0-4cb2-a9fd-1a2a21966bb8",
		Name:         "Monastery Siege",
		Completeness: CompletenessFull,
		AsEnters:     ChooseOptionAsEnters("Monastery Siege", "Khans", "Dragons"),
		Triggered: []game.TriggeredAbility{
			WhenChosen("Khans", On(game.EventBeginDrawStep,
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Actor == source.Controller
				}, "Monastery Siege — draw an additional card, then discard a card", monasterySiegeKhans)),
		},
		CostModifiers: []game.CostModifier{dragons},
	})
}

// monasterySiegeKhans is the Khans body.
func monasterySiegeKhans(g *game.Game, item *game.StackItem) error {
	return lootOne(g, item, 1)
}
