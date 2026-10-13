package decks

// Forge and Fury — aggro, rebuilt by #3071. Purphoros, God of the
// Forge deals 2 damage to each opponent whenever another creature
// enters under its controller's control, and the deck is built to make
// that happen as often as possible: Goblins that come with Goblins,
// token spells, creatures that make a token every combat, and a second
// rank of the same trigger (Impact Tremors, Witty Roastmaster, Molten
// Gatekeeper, Agate Instigator, Weftstalker Ardent) and of its spell
// and artifact cousins (Guttersnipe, Firebrand Archer, Kessig
// Flamebreather, Reckless Fireweaver). Damage to each opponent is the
// one thing a four-seat table pays three times for, and it is what the
// two strongest curated decks win with (Syr Konrad and Y'shtola).
//
// Why the commander changed. The deck was Raid and Ransack, Mary Read
// and Anne Bonny's Izzet Pirates, and won 2–5% of four-seat heuristic
// games against 25–40% for the other three (#2480). Two to four card
// swaps did not move it, and neither did keeping blue: the same engine
// under Vivi Ornitier or Gandalf the Grey won 8–10%. What moved it is
// the engine in the command zone. Purphoros is indestructible and is
// cast in nearly every game, so the trigger the other cards feed is on
// the table by turn four or five whatever the deck draws. A mono-red
// commander keeps the deck inside the colours the owner allowed (#3071).
//
// The ID was izzet-aggro until #3071 renamed it with the deck.
// Migration 0018 rewrites saved table setups and last-deck records, and
// legacyIDs (decks.go) still resolves the old ID for anything else
// that kept it: a guest's browser, an engine snapshot, a script.
//
// Every non-basic card below is in the effect catalog and graded full,
// apart from Path of Ancestry on full_test.go's minor-caveats list;
// decks_test.go and full_test.go fail the build if that stops being
// true.

var monoRedAggro = Deck{
	ID:        "mono-red-aggro",
	Name:      "Forge and Fury",
	Archetype: "aggro",
	Identity:  "R",
	Summary:   "Goblins and tokens: every creature that enters under your control is 2 damage to each opponent from Purphoros, and more from the pingers that back it up.",
	Commander: Card{Name: "Purphoros, God of the Forge", OracleID: "4fdbbec2-e921-4b63-958d-f9ba1e417197", Identity: "R"},
	Mainboard: []Card{
		// Damage to each opponent when a creature or an artifact enters.
		// Purphoros's own trigger, five more copies of it, and the two
		// that also count Treasures and Thopters.
		{Name: "Impact Tremors", OracleID: "9242cd3e-1a71-4700-8182-9c1005616033", Identity: "R"},
		{Name: "Witty Roastmaster", OracleID: "34edbf1b-0826-49f1-a04d-15cc8c543b22", Identity: "R"},
		{Name: "Molten Gatekeeper", OracleID: "90fdfba8-f29e-44f9-91d2-7bf3c458a9c1", Identity: "R"},
		{Name: "Agate Instigator", OracleID: "e1a2bd31-b800-47cd-8866-1b849a601b09", Identity: "R"},
		{Name: "Weftstalker Ardent", OracleID: "926d52a5-4db1-46ce-9567-17c28bf56ae7", Identity: "R"},
		{Name: "Reckless Fireweaver", OracleID: "180e1a7e-890d-477c-80a5-da8a5f2857b3", Identity: "R"},
		{Name: "Fanatic of Mogis", OracleID: "f19d06af-6caf-41d8-8d0a-d5d50bd67900", Identity: "R"},

		// Damage to each opponent when a spell is cast.
		{Name: "Guttersnipe", OracleID: "c6bdaf76-6a03-4695-9c4b-f040e73435af", Identity: "R"},
		{Name: "Firebrand Archer", OracleID: "2e9289d6-dbc6-456d-88cf-d1f534e731d6", Identity: "R"},
		{Name: "Kessig Flamebreather", OracleID: "eadd9559-6dcb-4b96-8c95-58abddd0e120", Identity: "R"},
		{Name: "Erebor Flamesmith", OracleID: "6eb93545-a1a6-4447-82c5-421d8e9f023d", Identity: "R"},
		{Name: "Coruscation Mage", OracleID: "88bb91b5-2ccd-4ce9-8cd4-e54d63c12abf", Identity: "R"},
		{Name: "Firespitter Whelp", OracleID: "1cad92ce-55c8-4b78-8d8c-56645ef8e6ee", Identity: "R"},
		{Name: "Fiery Inscription", OracleID: "111889d4-bcca-4b1f-ad48-3077e2f5136f", Identity: "R"},
		{Name: "Brimstone Vandal", OracleID: "7af23925-6d32-407d-9a99-a57b2ebb63ac", Identity: "R"},

		// More damage per hit.
		{Name: "Torbran, Thane of Red Fell", OracleID: "8c3495bf-02e7-4ad9-949d-92eb3d2b662a", Identity: "R"},
		{Name: "Hellrider", OracleID: "f02557b0-f422-48cb-875e-2c814c39967d", Identity: "R"},

		// Creatures that bring more creatures: each one is two or more
		// Purphoros triggers.
		{Name: "Goblin Instigator", OracleID: "8b022754-6d16-470e-b754-4df6e4f4709e", Identity: "R"},
		{Name: "Beetleback Chief", OracleID: "f5c5f64c-6911-430c-a825-b32b96d39c7d", Identity: "R"},
		{Name: "Siege-Gang Commander", OracleID: "ddc7f59a-bbb1-4ba1-82c8-6813fd191940", Identity: "R"},
		{Name: "Legion Warboss", OracleID: "0ccb9af3-6902-4130-a59c-4c882dc3a2fd", Identity: "R"},
		{Name: "Goblin Rabblemaster", OracleID: "24661b81-6bee-4ad8-b3ab-53cb65005c51", Identity: "R"},
		{Name: "Krenko, Mob Boss", OracleID: "68418069-f615-40ef-ae0d-764192acae00", Identity: "R"},
		{Name: "Krenko, Tin Street Kingpin", OracleID: "e8065e1d-e937-4b56-8011-78f0d07328a0", Identity: "R"},
		{Name: "Hanweir Garrison", OracleID: "7cb29569-48e1-4782-9906-fad155ebfafe", Identity: "R"},
		{Name: "Loyal Apprentice", OracleID: "7f268b15-ac92-4e98-821b-78d15d9285d9", Identity: "R"},
		{Name: "Young Pyromancer", OracleID: "5fac139a-07d3-4e6c-98e3-d98b199f7a6f", Identity: "R"},
		{Name: "Person of Interest", OracleID: "3d18cd9b-6013-48d7-9269-4787e7585a51", Identity: "R"},
		{Name: "Pashalik Mons", OracleID: "bdb94ccd-1bb5-4bb7-9539-e1b1c97c19c5", Identity: "R"},
		{Name: "Fearless Liberator", OracleID: "b0b93253-6432-401e-9a01-94c41fb72c10", Identity: "R"},
		{Name: "Goblin Matron", OracleID: "145737a7-c597-4dec-b752-207c2d0501e3", Identity: "R"},
		{Name: "Hordeling Outburst", OracleID: "a6450b8e-eb18-431c-9eb7-7daf107978b2", Identity: "R"},
		{Name: "Krenko's Command", OracleID: "9cfe86ae-eebe-44aa-a956-4b3e9e621105", Identity: "R"},
		{Name: "Dragon Fodder", OracleID: "d0d2c45b-b6e3-4999-bdab-976e8f0d6617", Identity: "R"},

		// Treasure makers and card flow. A Treasure is an artifact entering
		// for Reckless Fireweaver and Weftstalker Ardent, and mana for the
		// next creature.
		{Name: "Charming Scoundrel", OracleID: "9eec304d-ffcb-4287-b877-280ac1e4f496", Identity: "R"},
		{Name: "Generous Plunderer", OracleID: "91c835d1-22ca-4c90-9ba6-c8e01bbc0347", Identity: "R"},
		{Name: "Captain Lannery Storm", OracleID: "235bf0ba-658c-463f-b112-7478ba27bd7b", Identity: "R"},
		{Name: "Patron of the Arts", OracleID: "47f8df80-abc9-4d46-8078-8fbc23430259", Identity: "R"},
		{Name: "Professional Face-Breaker", OracleID: "04152e7a-969c-4858-841b-0a569a9fc1bf", Identity: "R"},
		{Name: "Fable of the Mirror-Breaker", OracleID: "c0957e5e-c71b-439c-931c-9f55d2f76ace", Identity: "R"},
		{Name: "Rapacious Dragon", OracleID: "0944ec2e-1dd9-459f-8f1d-667242cf52fe", Identity: "R"},
		{Name: "Goldspan Dragon", OracleID: "716b3ea2-45b7-4a8f-af72-de7f4e510eff", Identity: "R"},
		{Name: "Strike It Rich", OracleID: "c34c17c7-3827-49a2-be25-67f44fdfe150", Identity: "R"},
		{Name: "Big Score", OracleID: "a5cbd257-c836-493e-bb1a-76242619dea2", Identity: "R"},
		{Name: "Unexpected Windfall", OracleID: "498c10c9-253d-4b15-b48c-1509381b17e8", Identity: "R"},
		{Name: "Seize the Spoils", OracleID: "58f83528-9110-4895-b5ea-51b90af30a8d", Identity: "R"},
		{Name: "Faithless Looting", OracleID: "3d6fa57a-aa53-4b5c-b8af-a7612c823117", Identity: "R"},

		// Burn and spot removal.
		{Name: "Lightning Bolt", OracleID: "4457ed35-7c10-48c8-9776-456485fdf070", Identity: "R"},
		{Name: "Shock", OracleID: "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d", Identity: "R"},
		{Name: "Wild Slash", OracleID: "f76a0e5b-74c1-4ee8-b502-e8b988086de4", Identity: "R"},
		{Name: "Arc Trail", OracleID: "f1c26b25-371e-4fbf-a43d-7fd59a364d3a", Identity: "R"},
		{Name: "Abrade", OracleID: "f9db72dc-9a5b-48a4-a86e-7464d9a2166a", Identity: "R"},
		{Name: "Incinerate", OracleID: "d8fd7a34-8418-4e98-b79b-119c4348c667", Identity: "R"},
		{Name: "Lightning Strike", OracleID: "f34b9bc4-7bfe-47fd-ba23-4eeeb46026eb", Identity: "R"},
		{Name: "Chaos Warp", OracleID: "07a0cba9-8768-4fd9-a3d5-b0f83b4bf8e8", Identity: "R"},
		{Name: "Vandalblast", OracleID: "3567c3c8-b3c7-45b7-935b-b1fdbc973720", Identity: "R"},

		// Mana rocks. Each is also an artifact entering.
		{Name: "Sol Ring", OracleID: "6ad8011d-3471-4369-9d68-b264cc027487"},
		{Name: "Arcane Signet", OracleID: "0bc7f093-bef0-4f1a-852c-4b75ebf54838"},
		{Name: "Fire Diamond", OracleID: "97b477d8-2e05-475e-8ed6-7d680cb21cd9", Identity: "R"},
		{Name: "Mind Stone", OracleID: "c97361b5-af16-4a7b-af85-a429dbaf4ad2"},
		{Name: "Thought Vessel", OracleID: "9965d9c5-2ebf-4a6c-930e-55c5890979be"},
		{Name: "Ruby Medallion", OracleID: "0c367958-729e-416d-988b-098e90e7a1fd"},

		// Lands.
		{Name: "Command Tower", OracleID: "0895c9b7-ae7d-4bb3-af17-3b75deb50a25"},
		{Name: "Path of Ancestry", OracleID: "b473e293-59e3-4e04-acf2-622604aeb25f"},
		{Name: "Mana Confluence", OracleID: "d0ee5bdc-2b69-4b73-9a20-ffcc18783b29"},
		{Name: "Evolving Wilds", OracleID: "a75445d3-1303-4bb5-89ad-26ea93fecd48"},
		{Name: "Terramorphic Expanse", OracleID: "1bd3e453-aa21-4ee6-95c2-d6d920ee8e7a"},
		{Name: "Fabled Passage", OracleID: "0c85b8f7-0bd0-4680-9ec5-d4b110460a54"},
		{Name: "Prismatic Vista", OracleID: "032b8a0d-491a-4a12-ab9f-689010054d5b"},
		{Name: "Myriad Landscape", OracleID: "2549bc57-9ffb-4053-9f10-f2a5f792b845"},
		{Name: "Strip Mine", OracleID: "d21a89eb-7c5b-459a-acc7-12b20b13bf79"},
		{Name: "Geier Reach Sanitarium", OracleID: "7b9fafe7-d26a-4ed5-b4c4-ce13763770b5"},
		{Name: "Mountain", Count: 27, Basic: true},
	},
}
