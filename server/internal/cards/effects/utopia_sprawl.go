package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Utopia Sprawl — Enchantment — Aura {G} (#294):
//
//	"Enchant Forest
//	 As this Aura enters, choose a color.
//	 Whenever enchanted Forest is tapped for mana, its controller adds
//	 an additional one mana of the chosen color."
//
// Two seams meeting: #742's stored colour and #763's triggered mana
// ability. Neither had to learn about the other — the colour lands on
// Card.ChosenColor as the Aura enters and the trigger reads it back
// with AddsOneManaOfTheChosenColor, exactly as a mana ABILITY reads it
// with ProducedChosenColor.
//
// The purpose is ColorForMana (#780): the colour becomes mana, so an
// automated chooser should name a colour it wants to spend.
//
// Before a colour is chosen the trigger adds NOTHING — never "any
// colour". That is the reading every ChosenColor consumer takes, and
// the weaker one.
//
// "Enchant Forest" is a narrower enchant clause than Wild Growth's,
// and it is the same spec the CR 704.5m legality check reruns: a
// Forest that stops being one puts the Sprawl in the graveyard. Arbor
// Elf untaps it, which is why the two cards are always played
// together — and Arbor Elf already works, because its untap is an
// ordinary activated ability (see arbor_elf.go).
//
// Same declared auto-tap simplification as Wild Growth: the planner
// does not count the extra mana (ADR 0074 §7).
//
// No simplification of the card itself.
func init() {
	Register(Spec{
		OracleID:     "00d8efa6-a2d9-4249-8da7-b45173675329",
		Name:         "Utopia Sprawl",
		Completeness: CompletenessFull,
		Targets:      EnchantLand(Subtype("Forest")),
		AsEnters:     ChooseColorAsEnters(game.ColorForMana, "Utopia Sprawl"),
		ManaTriggers: []game.ManaTrigger{
			WheneverAttachedTapsForManaFunc(
				"Utopia Sprawl — add an additional one mana of the chosen color",
				AddsOneManaOfTheChosenColor()),
		},
	})
}
