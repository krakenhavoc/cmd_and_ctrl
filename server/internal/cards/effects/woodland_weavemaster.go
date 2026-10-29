package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Woodland Weavemaster — Creature — Elf Druid {1}{G}, 1/2 (The Hobbit, HOB):
//
//	"Vigilance
//	 Whenever another Elf you control enters, this creature gets +1/+1
//	 until end of turn.
//	 {T}: Add X mana of any one color, where X is this creature's power.
//	 Spend this mana only to cast Elf spells and activate abilities of Elf
//	 sources."
//
// The trigger is Marwyn's reading of "another Elf you control enters" and
// Kinsbaile Aspirant's boost, so a Weavemaster that has left by resolution
// gives nothing to nobody. The mana ability is Kami of Whispered Hopes':
// X is the Weavemaster's power as the ability is activated, counted after
// the layers, so this turn's +1/+1 counts, and the mana is all of one colour
// the player picks. Its one restriction tag is Gwenna's two-alternative
// reading (ManaRestrictAnyOf): casting a spell that is an Elf, or paying for
// an activated ability whose SOURCE is an Elf (CR 106.6). Mana abilities with
// a cost are activations too. Anything else is refused the mana, and the
// Elf is read off the object's types as they stand (a changeling counts).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "28ffddca-a3c4-467d-a2dc-2c5e62393e11",
		Name:            "Woodland Weavemaster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Elf"))
			}, "Woodland Weavemaster — +1/+1 until end of turn", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return BoostUntilEOT{
					Target:    ctx.Source(),
					Power:     1,
					Toughness: 1,
					Label:     "Woodland Weavemaster — +1/+1 until end of turn",
				}.Apply(ctx)
			}),
		},
		ManaAbilities: []ManaAbility{{
			Cost:           ManaAbilityCost{Tap: true},
			ProducedAmount: ProducedOneColor(SourcePower),
			Label:          "Add X mana of any one color, where X is this creature's power. Spend this mana only to cast Elf spells and activate abilities of Elf sources",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictSubtype("Elf")},
				[]string{game.ManaRestrictActivate, game.ManaRestrictSubtype("Elf")},
			)},
		}},
	})
}
