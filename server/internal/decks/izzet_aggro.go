package decks

// Raid and Ransack — aggro. Pirates and Islands: the deck teaches
// discarding for value. Mary Read and Anne Bonny makes a Treasure for
// every Island, Pirate or Vehicle card discarded, and the loots,
// rummages and wheels below are how the deck discards them. The other
// discard payoffs (Magmakin Artillerist, Marauding Mako, Scrounging
// Skyray, Glint-Horn Buccaneer) mean a loot is never a dead draw,
// which matters for a policy that will otherwise treat rummaging as
// card disadvantage.
//
// An upgraded precon (#2436): no fast mana beyond Sol Ring, and the
// Krenko goblins that did not play to the commander are gone.
//
// Every non-basic card below is in the effect catalog and graded full,
// apart from the minor caveats full_test.go names; decks_test.go and
// full_test.go fail the build if that stops being true.

var izzetAggro = Deck{
	ID:        "izzet-aggro",
	Name:      "Raid and Ransack",
	Archetype: "aggro",
	Identity:  "UR",
	Summary:   "Pirates and Islands: loot and rummage away the cards you don't need, and every Island, Pirate or Vehicle you discard becomes a Treasure for Mary Read and Anne Bonny.",
	Commander: Card{Name: "Mary Read and Anne Bonny", OracleID: "5182de2d-aceb-450e-bd20-8bc7db124334", Identity: "UR"},
	Mainboard: []Card{
		// Pirates. Every one is a card Mary Read and Anne Bonny turns into a
		// Treasure when it is discarded, and most of them make Treasure or
		// loot on their own.
		{Name: "Ragavan, Nimble Pilferer", OracleID: "37108cd4-bbab-4ce3-9ed6-f60e8422e703", Identity: "R"},
		{Name: "Breeches, Brazen Plunderer", OracleID: "eb77f7dc-e9e4-44ef-8616-9f4e737e8ca5", Identity: "R"},
		{Name: "Breeches, Eager Pillager", OracleID: "15361770-c6dc-4db7-b745-aa03e4b866ea", Identity: "R"},
		{Name: "Captain Lannery Storm", OracleID: "235bf0ba-658c-463f-b112-7478ba27bd7b", Identity: "R"},
		{Name: "Corsair Captain", OracleID: "a7ec13c6-7ade-433a-b5a2-047854eef486", Identity: "U"},
		{Name: "Brineborn Cutthroat", OracleID: "916cb70f-3b06-48ed-972d-75f805aa0892", Identity: "U"},
		{Name: "Captain Storm, Cosmium Raider", OracleID: "431e85e3-15e6-471b-ba71-6058394c9a96", Identity: "UR"},
		{Name: "Malcolm, Keen-Eyed Navigator", OracleID: "a66f8b44-0163-4456-b152-4acefab896a4", Identity: "U"},
		{Name: "Malcolm, Alluring Scoundrel", OracleID: "3bba3b36-f8d7-4fd6-892a-797226210b6e", Identity: "U"},
		{Name: "Spectral Sailor", OracleID: "a8fdbcdf-479d-4582-9ad5-9fbd4c740c29", Identity: "U"},
		{Name: "Zara, Renegade Recruiter", OracleID: "cd2720c2-522c-4fdc-9cef-9c5ce250fe7b", Identity: "UR"},
		{Name: "Corsairs of Umbar", OracleID: "edc74fb9-a368-4f73-8147-f318a32a3d06", Identity: "U"},
		{Name: "Coercive Recruiter", OracleID: "ad33530c-a8df-4c1c-a863-501e583290b6", Identity: "R"},
		{Name: "Angrath's Marauders", OracleID: "2d4976d4-649c-4d42-ac5a-ada4b46a480c", Identity: "R"},
		{Name: "Glint-Horn Buccaneer", OracleID: "64ad5657-78e9-4f34-8877-18c4f51fff9a", Identity: "R"},
		{Name: "Marauding Mako", OracleID: "e349be42-5f14-44a9-9608-281985c10e2d", Identity: "R"},
		{Name: "Scrounging Skyray", OracleID: "3a46d85b-ce1a-4842-a342-92a5bddb1053", Identity: "U"},
		{Name: "Magmakin Artillerist", OracleID: "900b9409-9c16-414d-8674-2ea42c2415a1", Identity: "R"},

		// Other discard and artifact payoffs, Treasure makers and threats.
		// Smuggler's Copter is a Vehicle, which Mary Read pays on too.
		{Name: "Ingenious Artillerist", OracleID: "752c7723-90f8-4e3a-8266-f251ee0dadd8", Identity: "R"},
		{Name: "Weftstalker Ardent", OracleID: "926d52a5-4db1-46ce-9567-17c28bf56ae7", Identity: "R"},
		{Name: "Generous Plunderer", OracleID: "91c835d1-22ca-4c90-9ba6-c8e01bbc0347", Identity: "R"},
		{Name: "Professional Face-Breaker", OracleID: "04152e7a-969c-4858-841b-0a569a9fc1bf", Identity: "R"},
		{Name: "Reckless Fireweaver", OracleID: "180e1a7e-890d-477c-80a5-da8a5f2857b3", Identity: "R"},
		{Name: "Guttersnipe", OracleID: "c6bdaf76-6a03-4695-9c4b-f040e73435af", Identity: "R"},
		{Name: "Ledger Shredder", OracleID: "e9117015-1050-44dd-a46b-e7ffe2085fae", Identity: "U"},
		{Name: "Mulldrifter", OracleID: "24d0f5e7-0d9e-4b76-900e-a7274e80312d", Identity: "U"},
		{Name: "Solphim, Mayhem Dominus", OracleID: "895f23a2-55b7-4cc0-8939-2efaaf097e6f", Identity: "R"},
		{Name: "Goldspan Dragon", OracleID: "716b3ea2-45b7-4a8f-af72-de7f4e510eff", Identity: "R"},
		{Name: "Ancient Copper Dragon", OracleID: "48daee9d-ddaf-410f-8c3a-12fa1064ab56", Identity: "R"},
		{Name: "Smuggler's Copter", OracleID: "49136bdc-bc50-49a2-999a-1ef9c16ea130"},

		// Burn and spot removal.
		{Name: "Lightning Bolt", OracleID: "4457ed35-7c10-48c8-9776-456485fdf070", Identity: "R"},
		{Name: "Shock", OracleID: "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d", Identity: "R"},
		{Name: "Abrade", OracleID: "f9db72dc-9a5b-48a4-a86e-7464d9a2166a", Identity: "R"},
		{Name: "Arc Trail", OracleID: "f1c26b25-371e-4fbf-a43d-7fd59a364d3a", Identity: "R"},
		{Name: "Blaze", OracleID: "0596920f-9946-42f4-a03b-24aab67f9f1b", Identity: "R"},
		{Name: "Fiery Temper", OracleID: "f07bd49d-8e71-4d56-be2a-638514011318", Identity: "R"},
		{Name: "Izzet Charm", OracleID: "a07698f6-5ad5-49a3-9da2-f82d407f5cd7", Identity: "UR"},
		{Name: "Prismari Command", OracleID: "fa3e28b1-131c-4223-81e0-18dfbab22c26", Identity: "UR"},
		{Name: "Chaos Warp", OracleID: "07a0cba9-8768-4fd9-a3d5-b0f83b4bf8e8", Identity: "R"},
		{Name: "Vandalblast", OracleID: "3567c3c8-b3c7-45b7-935b-b1fdbc973720", Identity: "R"},
		{Name: "Cyclonic Rift", OracleID: "d75b9c82-1b49-4c3e-a1b5-aeef57d6644b", Identity: "U"},
		{Name: "Rapid Hybridization", OracleID: "06692cd9-ac2f-4a32-8fd1-043ba3c0fe71", Identity: "U"},
		{Name: "Pongify", OracleID: "05849bd6-8f38-4031-be2b-e2aa03beb8cc", Identity: "U"},

		// Counterspells. Cheap enough to hold up behind a board.
		{Name: "Counterspell", OracleID: "cc187110-1148-4090-bbb8-e205694a39f5", Identity: "U"},
		{Name: "Negate", OracleID: "3407fe41-fdd3-4119-8f70-4bc4590a379f", Identity: "U"},
		{Name: "Swan Song", OracleID: "8ddfc283-c9b4-41a5-af88-cf0068e986cc", Identity: "U"},
		{Name: "An Offer You Can't Refuse", OracleID: "234a734b-ba28-4f1b-9d01-3c3e7d516590", Identity: "U"},
		{Name: "Pyroblast", OracleID: "ecc435e2-deb1-420a-a79f-01dd08747314", Identity: "R"},
		{Name: "Wash Away", OracleID: "a4630da0-fe9b-4ead-9621-eac4b7825c35", Identity: "U"},

		// Card flow. Every loot and rummage here is also a Magmakin
		// Artillerist / Marauding Mako / Scrounging Skyray trigger, and an
		// Island or a Pirate discarded to it is a Mary Read Treasure.
		{Name: "Preordain", OracleID: "ac641490-ca14-48d7-8cc4-b69ce984befa", Identity: "U"},
		{Name: "Frantic Search", OracleID: "16e015b2-f8a3-4b1a-80be-58a8f5fb5e8c", Identity: "U"},
		{Name: "Faithless Looting", OracleID: "3d6fa57a-aa53-4b5c-b8af-a7612c823117", Identity: "R"},
		{Name: "Big Score", OracleID: "a5cbd257-c836-493e-bb1a-76242619dea2", Identity: "R"},
		{Name: "Unexpected Windfall", OracleID: "498c10c9-253d-4b15-b48c-1509381b17e8", Identity: "R"},
		{Name: "Windfall", OracleID: "08becc07-28bc-4a2f-a6b0-28a2998d2f50", Identity: "U"},
		{Name: "Wheel of Fortune", OracleID: "a8abd966-de7b-46a3-8ac7-8747ab35653a", Identity: "R"},

		// Mana rocks.
		{Name: "Sol Ring", OracleID: "6ad8011d-3471-4369-9d68-b264cc027487"},
		{Name: "Arcane Signet", OracleID: "0bc7f093-bef0-4f1a-852c-4b75ebf54838"},
		{Name: "Izzet Signet", OracleID: "2fda4fe7-8b0c-489c-a000-6d358e614e34", Identity: "UR"},
		{Name: "Talisman of Creativity", OracleID: "14d2979d-5728-42d7-a027-0eb1f754655d", Identity: "UR"},
		{Name: "Mind Stone", OracleID: "c97361b5-af16-4a7b-af85-a429dbaf4ad2"},
		{Name: "Thought Vessel", OracleID: "9965d9c5-2ebf-4a6c-930e-55c5890979be"},
		{Name: "Commander's Sphere", OracleID: "0b67c4e2-f88b-4e01-85a1-9d5f5b8db13b"},

		// Lands.
		{Name: "Steam Vents", OracleID: "17039058-822d-409f-938c-b727a366ba63", Identity: "UR"},
		{Name: "Volcanic Island", OracleID: "c718911c-c955-4eb9-9e16-be4bd49a4e4e", Identity: "UR"},
		{Name: "Sulfur Falls", OracleID: "6a6c5e17-6465-4a1f-9d63-8a3ce2edc522", Identity: "UR"},
		{Name: "Shivan Reef", OracleID: "0fe16212-66c3-4e45-a641-7391e9b2e304", Identity: "UR"},
		{Name: "Stormcarved Coast", OracleID: "4722105b-0085-4bb8-bca1-9de0d3eb5600", Identity: "UR"},
		{Name: "Training Center", OracleID: "e3570ac7-c593-40e3-bbd6-ec3da6d8158d", Identity: "UR"},
		{Name: "Izzet Boilerworks", OracleID: "1cb9d94a-3039-4f2e-8fcc-6996f9a45f74", Identity: "UR"},
		{Name: "Temple of Epiphany", OracleID: "79f94050-d850-41ca-b1db-5ae0cf743f0a", Identity: "UR"},
		{Name: "Command Tower", OracleID: "0895c9b7-ae7d-4bb3-af17-3b75deb50a25"},
		{Name: "Exotic Orchard", OracleID: "27b047e3-0d41-45e2-98e9-9391d7923a1e"},
		{Name: "Reflecting Pool", OracleID: "67f43ac6-2a58-4b53-b5d7-0330e2a252e2"},
		{Name: "Mana Confluence", OracleID: "d0ee5bdc-2b69-4b73-9a20-ffcc18783b29"},
		{Name: "Path of Ancestry", OracleID: "b473e293-59e3-4e04-acf2-622604aeb25f"},
		{Name: "Scalding Tarn", OracleID: "cb027150-848c-4a66-88ad-e20222304dd8"},
		{Name: "Polluted Delta", OracleID: "ef86989d-ce80-4e55-aece-7d11710eeffa"},
		{Name: "Prismatic Vista", OracleID: "032b8a0d-491a-4a12-ab9f-689010054d5b"},
		{Name: "Evolving Wilds", OracleID: "a75445d3-1303-4bb5-89ad-26ea93fecd48"},
		{Name: "Terramorphic Expanse", OracleID: "1bd3e453-aa21-4ee6-95c2-d6d920ee8e7a"},
		{Name: "Fabled Passage", OracleID: "0c85b8f7-0bd0-4680-9ec5-d4b110460a54"},
		{Name: "Strip Mine", OracleID: "d21a89eb-7c5b-459a-acc7-12b20b13bf79"},
		{Name: "Myriad Landscape", OracleID: "2549bc57-9ffb-4053-9f10-f2a5f792b845"},
		{Name: "Geier Reach Sanitarium", OracleID: "7b9fafe7-d26a-4ed5-b4c4-ce13763770b5"},
		{Name: "Mountain", Count: 7, Basic: true},
		{Name: "Island", Count: 7, Basic: true},
	},
}
