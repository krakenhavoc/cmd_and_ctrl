package game

import "testing"

// TestManaAbilityInTextReadsCR605 pins the classifier against printed
// cards, Oracle text as Scryfall ships it. Each "no" is a card whose
// mana is not a mana ability by CR 605.1a/b, or not the card's own; most
// of them carry a produced_mana array, which is why that array is not
// the answer (#3943).
func TestManaAbilityInTextReadsCR605(t *testing.T) {
	cases := []struct {
		name, typeLine, text string
		want                 bool
	}{
		// Activated mana abilities, CR 605.1a.
		{"Sol Ring", "Artifact", "{T}: Add {C}{C}.", true},
		{"Mind Stone", "Artifact", "{T}: Add {C}.\n{1}, {T}, Sacrifice this artifact: Draw a card.", true},
		{"Ashnod's Altar", "Artifact", "Sacrifice a creature: Add {C}{C}.", true},
		{"Coldsteel Heart", "Artifact", "This artifact enters tapped.\nAs this artifact enters, choose a color.\n{T}: Add one mana of the chosen color.", true},
		{"Black Mana Battery", "Artifact", "{2}, {T}: Put a charge counter on this artifact.\n{T}, Remove any number of charge counters from this artifact: Add {B}, then add an additional {B} for each charge counter removed this way.", true},
		{"Barbed Sextant (a delayed draw)", "Artifact", "{1}, {T}, Sacrifice this artifact: Add one mana of any color. Draw a card at the beginning of the next turn's upkeep.", true},
		{"Pyromancer's Goggles (a delayed trigger)", "Legendary Artifact", "{T}: Add {R}. When that mana is spent to cast a red instant or sorcery spell, copy that spell and you may choose new targets for the copy.", true},
		{"Codie (a delayed trigger that reaches the library)", "Legendary Artifact Creature — Construct", "You can't cast permanent spells.\n{4}, {T}: Add {W}{U}{B}{R}{G}. When you next cast a spell this turn, exile cards from the top of your library until you exile an instant or sorcery card with lesser mana value.", true},
		{"Jack-in-the-Mox (a die table)", "Artifact", "{T}: Roll a six-sided die. This ability has the indicated effect.\n1 — Sacrifice this artifact and you lose 5 life.\n2 — Add {W}.\n3 — Add {U}.", true},
		{"Cradle Clearcutter (an amount)", "Artifact Creature — Golem", "Prototype {2}{G} — 1/3 (You may cast this spell with different mana cost, color, and size. It keeps its abilities and types.)\n{T}: Add an amount of {G} equal to this creature's power.", true},
		// Triggered mana abilities, CR 605.1b.
		{"Gauntlet of Power", "Artifact", "As this artifact enters, choose a color.\nCreatures of the chosen color get +1/+1.\nWhenever a basic land is tapped for mana of the chosen color, its controller adds an additional one mana of that color.", true},
		{"Forsaken Monument", "Legendary Artifact", "Colorless creatures you control get +2/+2.\nWhenever you tap a permanent for {C}, add an additional {C}.\nWhenever you cast a colorless spell, you gain 2 life.", true},
		{"Caged Sun", "Artifact", "As this artifact enters, choose a color.\nWhenever a land's ability causes you to add one or more mana of the chosen color, add an additional one mana of that color.", true},
		// A basic land type's intrinsic ability, CR 305.6.
		{"a dual land (reminder text only)", "Land — Island Swamp", "({T}: Add {U} or {B}.)", true},

		// Triggers on something other than a mana event, CR 605.5a.
		{"Su-Chi", "Artifact Creature — Construct", "When this creature dies, add {C}{C}{C}{C}.", false},
		{"Tangleroot", "Artifact", "Whenever a player casts a creature spell, that player adds {G}.", false},
		{"Thran Turbine", "Artifact", "At the beginning of your upkeep, you may add {C}{C}. This mana can't be spent to cast spells.", false},
		// Moves a card to or from a library, CR 605.1a.
		{"Millikin", "Artifact Creature — Construct", "{T}, Mill a card: Add {C}. (Activate only as an instant. To mill a card, put the top card of your library into your graveyard.)", false},
		{"Chromatic Sphere", "Artifact", "{1}, {T}, Sacrifice this artifact: Add one mana of any color. Draw a card. (Activate only as an instant.)", false},
		// Requires a target, CR 605.5a.
		{"Radiant Lotus", "Artifact", "{T}, Sacrifice one or more artifacts: Choose a color. Target player adds three mana of the chosen color for each artifact sacrificed this way. (Activate only as an instant.)", false},
		// A loyalty ability, CR 606.
		{"The Aetherspark", "Legendary Artifact — Equipment", "+1: Attach The Aetherspark to up to one target creature you control.\n−10: Add ten mana of any one color.", false},
		// Not the card's own: a grant, a token's ability.
		{"Paradise Mantle", "Artifact — Equipment", "Equipped creature has \"{T}: Add one mana of any color.\"\nEquip {1}", false},
		{"Gold Pan", "Artifact — Equipment", "When this Equipment enters, create a Treasure token. (It's an artifact with \"{T}, Sacrifice this token: Add one mana of any color.\")\nEquipped creature gets +1/+1.\nEquip {1}", false},
		{"Coin of Mastery", "Artifact", "{T}: Create a Treasure token. (It's an artifact with \"{T}, Sacrifice this token: Add one mana of any color.\")", false},
		// No mana at all, and a spell (CR 605.5b).
		{"Moonsilver Key", "Artifact", "{1}, {T}, Sacrifice this artifact: Search your library for an artifact card with a mana ability or a basic land card, reveal it, put it into your hand, then shuffle.", false},
		{"Dark Ritual", "Instant", "Add {B}{B}{B}.", false},
	}
	for _, tc := range cases {
		if got := ManaAbilityInText(tc.typeLine, tc.text); got != tc.want {
			t.Errorf("%s: ManaAbilityInText = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestHasPrintedManaAbilityPrefersTheText: a card whose text was read is
// answered by its text, even when the catalog runs something else; a card
// never read falls back to the catalog; a face-down card has no text.
func TestHasPrintedManaAbilityPrefersTheText(t *testing.T) {
	restore := CatalogManaAbilities
	t.Cleanup(func() { CatalogManaAbilities = restore })
	CatalogManaAbilities = func(key string) []ManaAbilityShape {
		if key == "catalogued" {
			return []ManaAbilityShape{{Produced: "{C}", Label: "Add {C}"}}
		}
		return nil
	}
	catalogued := Card{Name: "Millikin", TypeLine: "Artifact Creature", OracleID: "catalogued"}
	if !HasPrintedManaAbility(catalogued) {
		t.Error("a card never read falls back to the catalog")
	}
	catalogued.ManaAbilityText = PrintedNo
	if HasPrintedManaAbility(catalogued) {
		t.Error("the printed text overrides the catalog")
	}
	rock := Card{Name: "Worn Powerstone", TypeLine: "Artifact", ManaAbilityText: PrintedYes}
	if !HasPrintedManaAbility(rock) {
		t.Error("an uncatalogued card whose text has one")
	}
	rock.SetFaceDown(FaceDownManifested)
	if HasPrintedManaAbility(rock) {
		t.Error("a face-down card has no text (CR 708.2a)")
	}
}
