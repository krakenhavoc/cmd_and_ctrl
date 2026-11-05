package decks

// Spell Echo — spellslinger. Cheap instants and sorceries behind a
// table of creatures that reward each one: a token, a ping to each
// opponent, a card, a Treasure. Veyran makes every one of those
// rewards trigger twice, so a two-mana burn spell is also two Drakes
// from Talrand and two damage to everyone from Guttersnipe.
//
// Added by #554 for the spellslinger theme the picker lacked, and the
// first Izzet deck since izzet-aggro became mono-red (#3071). The
// spells are burn, card draw, counterspells and bounce, so the deck is
// also the picker's second deck with real interaction after The Long
// Answer.
//
// Every non-basic card below is graded full, and none is on
// full_test.go's minor-caveats list. Every instant declares what it
// answers (ADR 0142), so none of them stops smart autopass for
// nothing. Left out on purpose:
//
//   - Brainstorm, Ponder, Opt, Expressive Iteration — full, but they
//     declare no Purpose (ADR 0126 §6).
//   - Izzet Charm, Prismari Command, Thrill of Possibility, Dualcaster
//     Mage — full, but they declare no Answers.
//   - Sprite Dragon, Baral, Crackling Drake, Fire // Ice, Expansion //
//     Explosion — not in the catalog.

var izzetSpellslinger = Deck{
	ID:        "izzet-spellslinger",
	Name:      "Spell Echo",
	Archetype: "spellslinger",
	Identity:  "UR",
	Summary:   "Cast cheap instants and sorceries behind creatures that reward every spell: Veyran makes each of those rewards happen twice.",
	Commander: Card{Name: "Veyran, Voice of Duality", OracleID: "d57a6b19-bd06-4b07-8ee1-0b45612ba9e4", Identity: "UR"},
	Mainboard: []Card{
		// Spell payoffs. Veyran makes each of them trigger twice.
		{Name: "Young Pyromancer", OracleID: "5fac139a-07d3-4e6c-98e3-d98b199f7a6f", Identity: "R"},
		{Name: "Third Path Iconoclast", OracleID: "f7156897-2b02-4ecd-868d-d4d59244e9ed", Identity: "UR"},
		{Name: "Talrand, Sky Summoner", OracleID: "ea1eb902-a23c-44ff-9169-19baf71de238", Identity: "U"},
		{Name: "Murmuring Mystic", OracleID: "dcd4da46-5438-4454-8b1b-43ca51bda1f9", Identity: "U"},
		{Name: "Archmage Emeritus", OracleID: "8305d576-21d8-4ce7-8eda-a7cd9793aca5", Identity: "U"},
		{Name: "Storm-Kiln Artist", OracleID: "a145ff8c-5812-4bcb-bd16-9839dc25121d", Identity: "R"},
		{Name: "Guttersnipe", OracleID: "c6bdaf76-6a03-4695-9c4b-f040e73435af", Identity: "R"},
		{Name: "Harmonic Prodigy", OracleID: "2e2ace5b-4018-43af-8e72-ebafec1a7739", Identity: "R"},
		{Name: "Kessig Flamebreather", OracleID: "eadd9559-6dcb-4b96-8c95-58abddd0e120", Identity: "R"},
		{Name: "Firebrand Archer", OracleID: "2e9289d6-dbc6-456d-88cf-d1f534e731d6", Identity: "R"},
		{Name: "Thermo-Alchemist", OracleID: "228fbae1-423e-461d-b8c3-55786938a3cb", Identity: "R"},
		{Name: "Stormcatch Mentor", OracleID: "b2252471-01ca-4d58-aa99-4ab0aa5eae12", Identity: "UR"},
		{Name: "Archmage of Runes", OracleID: "28045b32-4c1a-40e5-a15d-524d0f8fe6ec", Identity: "U"},
		{Name: "Goblin Electromancer", OracleID: "81f06f84-1580-43c0-89d5-08d34541a519", Identity: "UR"},
		{Name: "Niv-Mizzet, Parun", OracleID: "33666a98-812f-4892-9f8d-33e0cbecc340", Identity: "UR"},
		{Name: "Ledger Shredder", OracleID: "e9117015-1050-44dd-a46b-e7ffe2085fae", Identity: "U"},
		{Name: "Deekah, Fractal Theorist", OracleID: "6b071146-2f83-4fda-a55a-46fcb74e6f28", Identity: "U"},
		{Name: "Electrostatic Field", OracleID: "2fa94a07-c932-4f85-b6e0-a97d2b29eb52", Identity: "R"},
		{Name: "Thousand-Year Storm", OracleID: "dd4cf149-2fae-40e5-b50b-639f6bcec65e", Identity: "UR"},
		{Name: "Arcane Bombardment", OracleID: "1e36a42f-58e2-49bc-8c14-7f75e1d8ef98", Identity: "R"},
		{Name: "Fiery Inscription", OracleID: "111889d4-bcca-4b1f-ad48-3077e2f5136f", Identity: "R"},
		{Name: "Primal Amulet", OracleID: "8e4d0da0-c7d8-4a20-9bfd-02c1331a7a49"},
		{Name: "Rhystic Study", OracleID: "53236dd7-845a-444c-96d5-f41ed7325d8f", Identity: "U"},
		{Name: "Mystic Remora", OracleID: "8a52f3c0-2552-4425-b2e3-5496eb2232a7", Identity: "U"},

		// Mana rocks.
		{Name: "Sol Ring", OracleID: "6ad8011d-3471-4369-9d68-b264cc027487"},
		{Name: "Arcane Signet", OracleID: "0bc7f093-bef0-4f1a-852c-4b75ebf54838"},
		{Name: "Izzet Signet", OracleID: "2fda4fe7-8b0c-489c-a000-6d358e614e34", Identity: "UR"},
		{Name: "Talisman of Creativity", OracleID: "14d2979d-5728-42d7-a027-0eb1f754655d", Identity: "UR"},
		{Name: "Mind Stone", OracleID: "c97361b5-af16-4a7b-af85-a429dbaf4ad2"},

		// Burn.
		{Name: "Lightning Bolt", OracleID: "4457ed35-7c10-48c8-9776-456485fdf070", Identity: "R"},
		{Name: "Chain Lightning", OracleID: "8785f42f-87b9-4a0a-89ce-ea423649ba9c", Identity: "R"},
		{Name: "Shock", OracleID: "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d", Identity: "R"},
		{Name: "Lightning Strike", OracleID: "f34b9bc4-7bfe-47fd-ba23-4eeeb46026eb", Identity: "R"},
		{Name: "Incinerate", OracleID: "d8fd7a34-8418-4e98-b79b-119c4348c667", Identity: "R"},
		{Name: "Arc Trail", OracleID: "f1c26b25-371e-4fbf-a43d-7fd59a364d3a", Identity: "R"},
		{Name: "Abrade", OracleID: "f9db72dc-9a5b-48a4-a86e-7464d9a2166a", Identity: "R"},

		// Card draw and selection.
		{Name: "Preordain", OracleID: "ac641490-ca14-48d7-8cc4-b69ce984befa", Identity: "U"},
		{Name: "Faithless Looting", OracleID: "3d6fa57a-aa53-4b5c-b8af-a7612c823117", Identity: "R"},
		{Name: "Frantic Search", OracleID: "16e015b2-f8a3-4b1a-80be-58a8f5fb5e8c", Identity: "U"},
		{Name: "Big Score", OracleID: "a5cbd257-c836-493e-bb1a-76242619dea2", Identity: "R"},
		{Name: "Unexpected Windfall", OracleID: "498c10c9-253d-4b15-b48c-1509381b17e8", Identity: "R"},
		{Name: "Strike It Rich", OracleID: "c34c17c7-3827-49a2-be25-67f44fdfe150", Identity: "R"},
		{Name: "Divination", OracleID: "273b339c-964b-4a18-8eb5-ceb8abcdfd9e", Identity: "U"},
		{Name: "Compulsive Research", OracleID: "f947808c-a1cc-41ea-9d06-e1279a0da527", Identity: "U"},
		{Name: "Deep Analysis", OracleID: "579cbd92-797f-4cdf-91ed-fca7a523eae5", Identity: "U"},
		{Name: "Lórien Revealed", OracleID: "66f28905-c7fc-4ada-8fa0-199626d9bedb", Identity: "U"},
		{Name: "Treasure Cruise", OracleID: "5b6bdf5a-2742-4851-92cd-a857a3852836", Identity: "U"},
		{Name: "Dig Through Time", OracleID: "f8b17b89-26ce-4208-874a-9e1d66514640", Identity: "U"},
		{Name: "Enter the Enigma", OracleID: "67554654-e751-4679-9e92-f3588525ae4f", Identity: "U"},

		// Counterspells.
		{Name: "Counterspell", OracleID: "cc187110-1148-4090-bbb8-e205694a39f5", Identity: "U"},
		{Name: "Negate", OracleID: "3407fe41-fdd3-4119-8f70-4bc4590a379f", Identity: "U"},
		{Name: "Swan Song", OracleID: "8ddfc283-c9b4-41a5-af88-cf0068e986cc", Identity: "U"},
		{Name: "Mana Drain", OracleID: "74d3277a-38e5-4732-afed-084a56148f20", Identity: "U"},
		{Name: "An Offer You Can't Refuse", OracleID: "234a734b-ba28-4f1b-9d01-3c3e7d516590", Identity: "U"},
		{Name: "Fierce Guardianship", OracleID: "d09c9cba-fdd2-479b-ad5d-d05181c3e3f9", Identity: "U"},
		{Name: "Force of Negation", OracleID: "ac2173f9-f223-440a-9231-fd98762bdc6f", Identity: "U"},

		// Removal.
		{Name: "Pongify", OracleID: "05849bd6-8f38-4031-be2b-e2aa03beb8cc", Identity: "U"},
		{Name: "Rapid Hybridization", OracleID: "06692cd9-ac2f-4a32-8fd1-043ba3c0fe71", Identity: "U"},
		{Name: "Chaos Warp", OracleID: "07a0cba9-8768-4fd9-a3d5-b0f83b4bf8e8", Identity: "R"},
		{Name: "Cyclonic Rift", OracleID: "d75b9c82-1b49-4c3e-a1b5-aeef57d6644b", Identity: "U"},
		{Name: "Aetherize", OracleID: "7c779721-cd1b-4696-9ae9-68ccc284ed2a", Identity: "U"},
		{Name: "Mystic Confluence", OracleID: "cd11c27b-9368-4622-bc5a-2e0b993cc42b", Identity: "U"},
		{Name: "Mizzium Mortars", OracleID: "48ddba1e-2ad7-463f-9307-d2379a800e51", Identity: "R"},
		{Name: "Vandalblast", OracleID: "3567c3c8-b3c7-45b7-935b-b1fdbc973720", Identity: "R"},
		{Name: "Blasphemous Act", OracleID: "7a2484a9-04fd-41a0-8224-610c1c07ed10", Identity: "R"},

		// Lands.
		{Name: "Command Tower", OracleID: "0895c9b7-ae7d-4bb3-af17-3b75deb50a25"},
		{Name: "Exotic Orchard", OracleID: "27b047e3-0d41-45e2-98e9-9391d7923a1e"},
		{Name: "Steam Vents", OracleID: "17039058-822d-409f-938c-b727a366ba63", Identity: "UR"},
		{Name: "Spirebluff Canal", OracleID: "eb0d8093-5f93-4b25-9384-08f9731bfb28", Identity: "UR"},
		{Name: "Shivan Reef", OracleID: "0fe16212-66c3-4e45-a641-7391e9b2e304", Identity: "UR"},
		{Name: "Sulfur Falls", OracleID: "6a6c5e17-6465-4a1f-9d63-8a3ce2edc522", Identity: "UR"},
		{Name: "Volcanic Island", OracleID: "c718911c-c955-4eb9-9e16-be4bd49a4e4e", Identity: "UR"},
		{Name: "Fiery Islet", OracleID: "026f4a4b-eedd-44e1-9d37-ca4fb8d6db98", Identity: "UR"},
		{Name: "Frostboil Snarl", OracleID: "7137aae6-260d-41de-8b4e-42a8cf752697", Identity: "UR"},
		{Name: "Izzet Boilerworks", OracleID: "1cb9d94a-3039-4f2e-8fcc-6996f9a45f74", Identity: "UR"},
		{Name: "Swiftwater Cliffs", OracleID: "2f4ad084-2062-44c0-9975-15f100204531", Identity: "UR"},
		{Name: "Training Center", OracleID: "e3570ac7-c593-40e3-bbd6-ec3da6d8158d", Identity: "UR"},
		{Name: "Reliquary Tower", OracleID: "c23e5b80-08d2-4e24-9908-fe2aa4f30f6f"},
		{Name: "Island", Count: 11, Basic: true},
		{Name: "Mountain", Count: 10, Basic: true},
	},
}
