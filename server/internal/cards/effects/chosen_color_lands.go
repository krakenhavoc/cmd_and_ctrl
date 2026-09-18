package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// chosen_color_lands.go — the lands that ask for a colour as they
// enter and tap for it afterwards (#742). Two printed shapes:
//
// The Thriving lands and the Gates:
//
//	"This land enters tapped. As it enters, choose a color other than
//	 blue.
//	 {T}: Add {U} or one mana of the chosen color."
//
// Thriving Isle / Heath / Grove / Bluff / Moor (Jumpstart), and Sea
// Gate / Citadel Gate / Manor Gate / Cliffgate / Black Dragon Gate
// (Baldur's Gate). The Gates are the same card with a Gate subtype,
// which the deck importer stamps from the type line.
//
// The one-colour lands:
//
//	"This land enters tapped. As it enters, choose a color.
//	 {T}: Add one mana of the chosen color."
//
// Uncharted Haven, Crossroads Village, Mirage Mesa, and Valgavoth's
// Lair (which also has hexproof, a printed keyword the importer
// stamps).
//
// Every one is a tapped entry (SelfEntersTapped, the CR 614
// replacement), an AsEnters colour prompt (ChooseColorAsEnters, the
// S26 creature-type pattern — see game/color_choice.go), and a mana
// ability whose output is read off the stored colour at activation.
// The "or" in "{U} or one mana of the chosen color" is one ability
// with a two-colour pick, not two abilities; it does not say "in your
// commander's color identity", so the pick is never narrowed.
//
// Before the controller answers, the chosen colour is empty: a
// Thriving land taps for its printed colour only and a one-colour land
// taps for nothing. Nobody can act in that window (an open prompt holds
// priority), so the difference is unobservable, and it is the weaker
// direction if it ever were.
//
// No simplification.
func init() {
	for _, l := range []struct{ oracleID, name, color string }{
		{"69fc70b8-b143-4662-ac95-e2743037239d", "Thriving Isle", "U"},
		{"d1946630-e224-40db-8f0d-388b09622288", "Thriving Heath", "W"},
		{"a8052556-8962-4130-86a8-6fb7b6a324f7", "Thriving Grove", "G"},
		{"91fceb34-0f2d-4392-be27-00dcd765637f", "Thriving Bluff", "R"},
		{"bff416bb-d193-4c45-b2c1-7c297dbfad08", "Thriving Moor", "B"},
		{"b574c540-9f8a-4fd4-8809-d02c9b099ddc", "Sea Gate", "U"},
		{"15f1fe23-5af4-4fc4-8cde-2e0bf9f9be0c", "Citadel Gate", "W"},
		{"dd6e67c0-66a1-49b7-8a86-3cf4b209fd07", "Manor Gate", "G"},
		{"1999b5ac-21fb-4d99-ad72-58bf507f9a59", "Cliffgate", "R"},
		{"dde6bce5-8bbe-4866-b5aa-2c05c7d37241", "Black Dragon Gate", "B"},
	} {
		Register(Spec{
			OracleID:     l.oracleID,
			Name:         l.name,
			Completeness: CompletenessFull,
			Replacements: []game.ReplacementEffect{SelfEntersTapped()},
			AsEnters:     ChooseColorOtherThanAsEnters(l.name, l.color),
			ManaAbilities: []ManaAbility{{
				Cost:         ManaAbilityCost{Tap: true},
				ProducedFunc: ProducedColorOrChosen(l.color),
				Label:        "Add {" + l.color + "} or one mana of the chosen color",
			}},
		})
	}

	for _, l := range []struct{ oracleID, name string }{
		{"d23c3613-bc5e-4fc5-939c-62a090c53a79", "Uncharted Haven"},
		{"b26cfeb0-7bbe-4d93-8eed-e832f175a80c", "Crossroads Village"},
		{"e103f422-85c0-43f8-8a2f-8b7863e503fa", "Mirage Mesa"},
		{"660d44a2-391a-416c-b46c-ddcc3739f527", "Valgavoth's Lair"},
	} {
		Register(Spec{
			OracleID:     l.oracleID,
			Name:         l.name,
			Completeness: CompletenessFull,
			Replacements: []game.ReplacementEffect{SelfEntersTapped()},
			AsEnters:     ChooseColorAsEnters(l.name),
			ManaAbilities: []ManaAbility{{
				Cost:         ManaAbilityCost{Tap: true},
				ProducedFunc: ProducedChosenColor(),
				Label:        "Add one mana of the chosen color",
			}},
		})
	}
}
