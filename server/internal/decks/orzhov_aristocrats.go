package decks

// Grave Dividends — aristocrats. Make small creatures and tokens,
// sacrifice them, and collect on every death. Teysa makes each death
// trigger fire twice and gives the tokens vigilance and lifelink, so
// one Blood Artist drain is two, and a Grave Pact sacrifice is two per
// opponent.
//
// Added by #554 as the picker's second aristocrats deck. Body Count
// (mono-black-aristocrats) wins by reanimating its best creatures;
// this one wins by going wide with tokens, so the two play differently
// for the same reason. It is also the first Orzhov deck.
//
// Every non-basic card below is graded full, and none is on
// full_test.go's minor-caveats list. Left out on purpose:
//
//   - Karmic Guide — with Reveillark and a free sacrifice outlet the two
//     return each other forever. Combo stays out (ADR 0033).
//   - Village Rites, Kaya's Guile, Return to Dust — full, but they
//     declare no Answers (ADR 0142).
//   - Grim Haruspex, Requiem Angel, Spectral Procession, Hallowed
//     Spiritkeeper — not in the catalog.

var orzhovAristocrats = Deck{
	ID:        "orzhov-aristocrats",
	Name:      "Grave Dividends",
	Archetype: "aristocrats",
	Identity:  "WB",
	Summary:   "Make tokens, sacrifice them and drain the table: Teysa doubles every death trigger and gives the tokens lifelink.",
	Commander: Card{Name: "Teysa Karlov", OracleID: "644eeefd-e684-4ca8-8aef-a892ca130c07", Identity: "WB"},
	Mainboard: []Card{
		// Death payoffs. Teysa makes each of them trigger twice.
		{Name: "Blood Artist", OracleID: "310f141c-7f37-4729-aed6-dd9c09db448d", Identity: "B"},
		{Name: "Zulaport Cutthroat", OracleID: "76b003e0-15af-4f22-bdf2-1ade5430964a", Identity: "B"},
		{Name: "Cruel Celebrant", OracleID: "3ee78cfc-0e9e-4737-a7e2-b42f94228040", Identity: "WB"},
		{Name: "Falkenrath Noble", OracleID: "3739b179-bc81-4737-8376-66a57e16b942", Identity: "B"},
		{Name: "Bastion of Remembrance", OracleID: "c7f33cea-2ec8-4081-9208-a5b1d86721b3", Identity: "B"},
		{Name: "Vindictive Vampire", OracleID: "11ad664e-36d3-4d5b-8a87-59ea715877e0", Identity: "B"},
		{Name: "Mirkwood Bats", OracleID: "0636b6c3-0662-420a-b30d-f0a14e7c512d", Identity: "B"},
		{Name: "Pitiless Plunderer", OracleID: "a784481f-eccb-4112-bb38-04a659319660", Identity: "B"},
		{Name: "Midnight Reaper", OracleID: "e8c7566d-7cc0-48af-a986-83223ec7e06c", Identity: "B"},
		{Name: "Morbid Opportunist", OracleID: "322f44f0-e6da-4ee0-b474-e7d5e9a461c5", Identity: "B"},
		{Name: "Nadier's Nightblade", OracleID: "391978f6-0bbc-41e8-9246-f7d0e21c7900", Identity: "B"},
		{Name: "Butcher of Malakir", OracleID: "a85197ab-dc94-4b72-9716-8dbdbbe90ff8", Identity: "B"},
		{Name: "Grave Pact", OracleID: "6f4ac4a4-53ec-4bc9-8f5c-d4b801d867b2", Identity: "B"},
		{Name: "Dictate of Erebos", OracleID: "7c777a41-e40a-4b40-96bf-8ddd5c12924c", Identity: "B"},
		{Name: "Elas il-Kor, Sadistic Pilgrim", OracleID: "a4d33e21-d3fe-4d5b-903c-b75e859d6f7f", Identity: "WB"},
		{Name: "Teysa, Orzhov Scion", OracleID: "8191342b-b25e-4c4d-8f69-aee662148ff4", Identity: "WB"},
		{Name: "Sifter of Skulls", OracleID: "688e5e19-8f28-45c4-ace4-a38144f494f1", Identity: "B"},
		{Name: "Liesa, Forgotten Archangel", OracleID: "efcaadbe-24e3-4dfc-b08c-a910f003d427", Identity: "WB"},
		{Name: "Elenda, the Dusk Rose", OracleID: "ce35f262-b93f-4409-8b33-26a7ed7abac8", Identity: "WB"},
		{Name: "Skullclamp", OracleID: "65986c1b-8e51-4604-b685-d82fa7d1263a"},

		// Sacrifice outlets.
		{Name: "Viscera Seer", OracleID: "f82a4e85-526d-4456-b700-7760043a31be", Identity: "B"},
		{Name: "Carrion Feeder", OracleID: "a1cc5e37-b09a-4b7f-afd5-77c1c35aa425", Identity: "B"},
		{Name: "Yawgmoth, Thran Physician", OracleID: "a1e232c0-dc38-47be-a5a0-f68bc1d86a29", Identity: "B"},
		{Name: "Ashnod's Altar", OracleID: "4d18bcba-a346-445e-a182-6cc30b7e066d"},
		{Name: "Spawning Pit", OracleID: "ea7b0704-7715-4f8a-b32f-d2d2f3b36054"},
		{Name: "Woe Strider", OracleID: "3adbd963-e85d-4569-963a-4472594f06f9", Identity: "B"},
		{Name: "Deadly Dispute", OracleID: "457af74a-02b3-4659-846d-63e482667f34", Identity: "B"},

		// Fodder: tokens and creatures that leave something behind.
		{Name: "Doomed Traveler", OracleID: "a30907c0-fbde-4fd3-a8c7-f304305fcea7", Identity: "W"},
		{Name: "Pawn of Ulamog", OracleID: "9bcaf141-1f1f-491f-aced-13dc093b9e2c", Identity: "B"},
		{Name: "Ophiomancer", OracleID: "55eca80c-dcd8-4c2f-aa0f-fb0aec7b80f7", Identity: "B"},
		{Name: "Bitterblossom", OracleID: "fb868840-09fa-49b1-85cb-b08ad065e972", Identity: "B"},
		{Name: "Lingering Souls", OracleID: "0b8c3337-04dd-4798-8203-6d8b8cfb936b", Identity: "WB"},
		{Name: "Secure the Wastes", OracleID: "b2347910-d6c6-4681-8316-7ef27056485c", Identity: "W"},
		{Name: "Raise the Alarm", OracleID: "5b2364d7-a811-4595-a1b4-224c70555ffa", Identity: "W"},
		{Name: "Bennie Bracks, Zoologist", OracleID: "17bd7ef7-8b4b-4a2d-a667-c751e10e2a47", Identity: "W"},
		{Name: "Mentor of the Meek", OracleID: "b9f4f96b-6e54-4fe6-8df7-623e0fc72409", Identity: "W"},
		{Name: "Selfless Spirit", OracleID: "71d785a9-ddc8-472e-a778-a551b444a4bd", Identity: "W"},
		{Name: "Sorin, Lord of Innistrad", OracleID: "5c780dda-ee76-482a-a610-53c321c4fdec", Identity: "WB"},
		{Name: "Reveillark", OracleID: "1be13ede-98f8-497e-800c-03e5802932b3", Identity: "W"},
		{Name: "Sun Titan", OracleID: "b2e950fb-cb7e-40a0-a311-5bbdd0477b29", Identity: "W"},
		{Name: "Kokusho, the Evening Star", OracleID: "cf631c93-b7fb-4e4f-b405-3778c15b1117", Identity: "B"},
		{Name: "Wurmcoil Engine", OracleID: "d1a60f44-7696-49ee-91fb-cab5b3102962"},
		{Name: "Elspeth, Sun's Champion", OracleID: "05e6b243-48a6-4a42-bc5f-413441de9c33", Identity: "W"},
		{Name: "Smothering Tithe", OracleID: "153376c9-dffd-458c-8ce3-a4c8269bc4e9", Identity: "W"},

		// Mana rocks.
		{Name: "Sol Ring", OracleID: "6ad8011d-3471-4369-9d68-b264cc027487"},
		{Name: "Arcane Signet", OracleID: "0bc7f093-bef0-4f1a-852c-4b75ebf54838"},
		{Name: "Orzhov Signet", OracleID: "de3dcb5d-775a-479f-99f5-d1883ed9b1b5", Identity: "WB"},
		{Name: "Talisman of Hierarchy", OracleID: "b693c3de-2eaf-4850-b405-e79d00adefda", Identity: "WB"},
		{Name: "Mind Stone", OracleID: "c97361b5-af16-4a7b-af85-a429dbaf4ad2"},

		// Removal.
		{Name: "Swords to Plowshares", OracleID: "b1544f21-7e98-461b-aed5-e748b0168c52", Identity: "W"},
		{Name: "Path to Exile", OracleID: "d683d985-9888-4d21-8b5f-69e69ce4a03b", Identity: "W"},
		{Name: "Infernal Grasp", OracleID: "94f0a572-e91c-4b56-a5d1-6cbbeabd210d", Identity: "B"},
		{Name: "Anguished Unmaking", OracleID: "ad09b3c3-c8e7-481c-8c45-e7f234935117", Identity: "WB"},
		{Name: "Mortify", OracleID: "faa01ed1-ccfa-4e58-951f-cd81f9068027", Identity: "WB"},
		{Name: "Vindicate", OracleID: "63c1ac21-e3d8-40c2-8c09-3f31c52992ef", Identity: "WB"},
		{Name: "Despark", OracleID: "bd16434d-55ea-4c5a-a9ef-752971a4af16", Identity: "WB"},
		{Name: "Utter End", OracleID: "cba94732-1f1f-4ffd-9aba-45db990043fa", Identity: "WB"},
		{Name: "Generous Gift", OracleID: "fae37e28-e137-4177-b973-fa8b4dd8f409", Identity: "W"},
		{Name: "Austere Command", OracleID: "09cc8709-fe10-472a-b05c-e89f3523018d", Identity: "W"},
		{Name: "Damn", OracleID: "b01d61cc-9844-4191-86a0-f2db6d42d6e5", Identity: "WB"},

		// Card draw.
		{Name: "Night's Whisper", OracleID: "7ffae8f8-3006-4969-a339-6d30678f87ea", Identity: "B"},
		{Name: "Read the Bones", OracleID: "5bf4d8d9-a2b2-4dba-ac05-9d4470a89db2", Identity: "B"},
		{Name: "Sign in Blood", OracleID: "c6207f6a-a624-4754-88f5-dbe700c841ff", Identity: "B"},
		{Name: "Phyrexian Arena", OracleID: "ee579a32-a048-4335-b966-231ba731cdea", Identity: "B"},

		// Lands.
		{Name: "Command Tower", OracleID: "0895c9b7-ae7d-4bb3-af17-3b75deb50a25"},
		{Name: "Godless Shrine", OracleID: "73864fcc-1bde-4bc0-831e-2b93e546e417", Identity: "WB"},
		{Name: "Scrubland", OracleID: "c8d95ca8-7d12-4072-aeaf-e20f248c7e39", Identity: "WB"},
		{Name: "Marsh Flats", OracleID: "dab520d0-20b4-4273-ba6b-eb07f85ea433"},
		{Name: "Caves of Koilos", OracleID: "33de01e9-ce5a-42d4-afcb-343cd54a6d80", Identity: "WB"},
		{Name: "Isolated Chapel", OracleID: "7e5d9efe-48a9-434b-bb09-056e0e09cc9a", Identity: "WB"},
		{Name: "Shadowy Backstreet", OracleID: "216a2a92-9ca3-4ca3-8af7-686c13b04290", Identity: "WB"},
		{Name: "Scoured Barrens", OracleID: "d37f858e-03c8-4594-9b92-cd03699a1591", Identity: "WB"},
		{Name: "Fetid Heath", OracleID: "42bf259d-4bb9-49c3-b4ec-223dca62f4d6", Identity: "WB"},
		{Name: "Silent Clearing", OracleID: "fd45063f-c83c-431c-9104-f139c497ec0d", Identity: "WB"},
		{Name: "Vault of the Archangel", OracleID: "eeaac65a-3480-475a-bb28-e6375d53f487", Identity: "WB"},
		{Name: "Orzhov Basilica", OracleID: "aa00ae0b-7c0f-427e-8102-ce0e2a6af5df", Identity: "WB"},
		{Name: "Bojuka Bog", OracleID: "04b7362d-0490-4cb0-b5d7-2a7732f659ce", Identity: "B"},
		{Name: "Plains", Count: 11, Basic: true},
		{Name: "Swamp", Count: 11, Basic: true},
	},
}
