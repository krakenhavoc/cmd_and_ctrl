package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Three Tree Mascot — Artifact Creature — Shapeshifter {2}, 2/1
// (EDHREC rank 4170):
//
//	"Changeling (This card is every creature type.)
//	 {1}: Add one mana of any color. Activate only once each turn."
//
// A two-mana artifact creature that is every creature type at once,
// which is the only reason it is played: it is an Elf for Elvish
// Archdruid, a Sliver for The First Sliver (also in this batch), a
// Goblin for Krenko, and a legal target for any tribal tutor. The
// mana ability is filtering, not ramp — {1} in, one coloured mana out,
// once a turn — so it fixes a splash rather than accelerating.
//
// Changeling is a printed keyword the engine already understands, so
// the type half is one string; the layer engine gives the Mascot every
// creature type from there.
//
// "Activate only once each turn" is a CR 602.1b condition, not a cost
// and not sorcery speed, and the engine keeps no per-ability
// activation tally — so the tally is the event log
// (b40NotActivatedForManaThisTurn). Checked BEFORE any cost is
// validated, so a second activation in a turn spends nothing.
//
// One consequence: the condition drops a used-up Mascot out of the
// auto-tapper's plan for the rest of the turn, which is correct
// rather than a leak.
//
// "Any color" offers all five, with the controller's commander
// identity first — the printed text names no identity, so
// NarrowToCommanderIdentity stays off.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6cb9d153-4eaa-4f32-8930-bd68cd99c128",
		Name:            "Three Tree Mascot",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
		ManaAbilities: []ManaAbility{{
			Cost:      ManaAbilityCost{Mana: "{1}"},
			Produced:  "{W|U|B|R|G}",
			Label:     "{1}: Add one mana of any color. Activate only once each turn",
			Condition: b40NotActivatedForManaThisTurn(),
		}},
	})
}
