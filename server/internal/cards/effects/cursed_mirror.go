package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cursed Mirror — Artifact {2}{R}:
//
//	"{T}: Add {R}.
//	 As this artifact enters, you may have it become a copy of any
//	 creature on the battlefield until end of turn, except it has
//	 haste."
//
// An entry copy with a DURATION (#1593): EntersAsCopyOfUntilEndOfTurn
// lands the copy as the artifact enters — before any event, so the
// copied creature's own ETB triggers fire for it — and the engine then
// files it as a duration copy, so at the cleanup step it is Cursed
// Mirror again, an artifact that taps for {R}. While it is the creature
// it has none of Cursed Mirror's own text, the mana ability included,
// which is what "become a copy" means.
//
// "Except it has haste" is part of the copiable values (CR 707.9a), so
// a Clone copying the mirror-creature gets haste too.
//
// Declining is always legal; the artifact then enters as itself.
func init() {
	Register(Spec{
		OracleID:     "4d67e2a7-4aa7-44cc-853b-500d7aac046d",
		Name:         "Cursed Mirror",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOfUntilEndOfTurn(
				"Cursed Mirror",
				anyCreatureOnBattlefield,
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.AddKeyword("haste")
				},
			),
		},
	})
}
