package decks

// The Long Answer — control. Answer everything, then win with whatever
// is left. This is the deck most sensitive to policy quality — a bot
// that fires removal at the first legal target plays it badly — so it
// is also the most useful regression target for the heuristic and
// model tiers.
//
// Every non-basic card below is in the effect catalog; decks_test.go
// fails the build if that stops being true.

var esperControl = Deck{
	ID:        "esper-control",
	Name:      "The Long Answer",
	Archetype: "control",
	Identity:  "WUB",
	Summary:   "Counterspells, one-for-one removal and six board wipes, with just enough creatures to close.",
	Commander: Card{Name: "Hashaton, Scarab's Fist", OracleID: "db266661-f783-4907-9e52-6963eec05431", Identity: "WUB"},
	Mainboard: []Card{
		// Creatures. Few, and every one of them draws or removes.
		{Name: "Esper Sentinel", OracleID: "5def9f38-0a0b-4e8d-9f9d-29dcb46520b4", Identity: "W"},
		{Name: "Thraben Inspector", OracleID: "caa02547-66e3-4e27-a2d3-5e94f3e7a069", Identity: "W"},
		{Name: "Archivist of Oghma", OracleID: "08b13e1f-27ca-40a8-b5ed-88ac933d24bf", Identity: "W"},
		{Name: "Baleful Strix", OracleID: "37688720-03de-4eca-a82d-a0afe8d58adc", Identity: "UB"},
		{Name: "Loran of the Third Path", OracleID: "b3d81980-76f2-44e2-b1c9-01e30c726312", Identity: "W"},
		{Name: "Ravenous Chupacabra", OracleID: "7b459306-149b-4f43-abc1-2dd70c748c0e", Identity: "B"},
		{Name: "Mulldrifter", OracleID: "24d0f5e7-0d9e-4b76-900e-a7274e80312d", Identity: "U"},
		{Name: "Gray Merchant of Asphodel", OracleID: "38f3b157-0df4-409b-89cc-086e1531cd5b", Identity: "B"},
		{Name: "Sheoldred, the Apocalypse", OracleID: "34f34409-326d-4994-a0ea-1a69aa278f03", Identity: "B"},
		{Name: "Consecrated Sphinx", OracleID: "311a449d-dc74-46e6-9a47-6a597931f736", Identity: "U"},
		{Name: "Sun Titan", OracleID: "b2e950fb-cb7e-40a0-a311-5bbdd0477b29", Identity: "W"},

		// Counterspells.
		{Name: "Counterspell", OracleID: "cc187110-1148-4090-bbb8-e205694a39f5", Identity: "U"},
		{Name: "Negate", OracleID: "3407fe41-fdd3-4119-8f70-4bc4590a379f", Identity: "U"},
		{Name: "Swan Song", OracleID: "8ddfc283-c9b4-41a5-af88-cf0068e986cc", Identity: "U"},
		{Name: "An Offer You Can't Refuse", OracleID: "234a734b-ba28-4f1b-9d01-3c3e7d516590", Identity: "U"},
		{Name: "Arcane Denial", OracleID: "ab1cc360-b9de-48d9-9983-4dfe4a7d2a37", Identity: "U"},
		{Name: "Mana Drain", OracleID: "74d3277a-38e5-4732-afed-084a56148f20", Identity: "U"},
		{Name: "Mental Misstep", OracleID: "1a0770e6-b093-4439-baff-6889a50ba12e", Identity: "U"},
		{Name: "Wash Away", OracleID: "a4630da0-fe9b-4ead-9621-eac4b7825c35", Identity: "U"},

		// Spot removal.
		{Name: "Swords to Plowshares", OracleID: "b1544f21-7e98-461b-aed5-e748b0168c52", Identity: "W"},
		{Name: "Path to Exile", OracleID: "d683d985-9888-4d21-8b5f-69e69ce4a03b", Identity: "W"},
		{Name: "Doom Blade", OracleID: "59e7f2ae-4535-4191-98be-3e65b6b2befa", Identity: "B"},
		{Name: "Go for the Throat", OracleID: "2f092562-9e17-43cd-aeb8-d0567f99363e", Identity: "B"},
		{Name: "Infernal Grasp", OracleID: "94f0a572-e91c-4b56-a5d1-6cbbeabd210d", Identity: "B"},
		{Name: "Feed the Swarm", OracleID: "5825997b-10d7-4a36-972c-a80ddd90b8ed", Identity: "B"},
		{Name: "Withering Torment", OracleID: "ffce81c5-1b58-4882-a4e7-6f8d7cb170de", Identity: "B"},
		{Name: "Anguished Unmaking", OracleID: "ad09b3c3-c8e7-481c-8c45-e7f234935117", Identity: "WB"},
		{Name: "Mortify", OracleID: "faa01ed1-ccfa-4e58-951f-cd81f9068027", Identity: "WB"},
		{Name: "Despark", OracleID: "bd16434d-55ea-4c5a-a9ef-752971a4af16", Identity: "WB"},
		{Name: "Generous Gift", OracleID: "fae37e28-e137-4177-b973-fa8b4dd8f409", Identity: "W"},
		{Name: "Stroke of Midnight", OracleID: "9a107e48-3d50-4941-95b1-10f2b29a4245", Identity: "W"},

		// Board wipes.
		{Name: "Wrath of God", OracleID: "34515b16-c9a4-4f98-8c77-416a7a523407", Identity: "W"},
		{Name: "Damnation", OracleID: "d57a8f0b-7989-4db5-8756-6f2690097252", Identity: "B"},
		{Name: "Day of Judgment", OracleID: "d057289d-5e28-43d5-8ff3-4a1bc723477d", Identity: "W"},
		{Name: "Damn", OracleID: "b01d61cc-9844-4191-86a0-f2db6d42d6e5", Identity: "WB"},
		{Name: "Austere Command", OracleID: "09cc8709-fe10-472a-b05c-e89f3523018d", Identity: "W"},
		{Name: "Farewell", OracleID: "4eb813fd-2d5a-4b02-8193-662681ef4e7d", Identity: "W"},
		{Name: "Cyclonic Rift", OracleID: "d75b9c82-1b49-4c3e-a1b5-aeef57d6644b", Identity: "U"},

		// Card advantage.
		{Name: "Rhystic Study", OracleID: "53236dd7-845a-444c-96d5-f41ed7325d8f", Identity: "U"},
		{Name: "Smothering Tithe", OracleID: "153376c9-dffd-458c-8ce3-a4c8269bc4e9", Identity: "W"},
		{Name: "Phyrexian Arena", OracleID: "ee579a32-a048-4335-b966-231ba731cdea", Identity: "B"},
		{Name: "Night's Whisper", OracleID: "7ffae8f8-3006-4969-a339-6d30678f87ea", Identity: "B"},
		{Name: "Sign in Blood", OracleID: "c6207f6a-a624-4754-88f5-dbe700c841ff", Identity: "B"},
		{Name: "Ambition's Cost", OracleID: "84de4fec-2f38-4293-93d3-b3882c5aac14", Identity: "B"},
		{Name: "Divination", OracleID: "273b339c-964b-4a18-8eb5-ceb8abcdfd9e", Identity: "U"},
		{Name: "Preordain", OracleID: "ac641490-ca14-48d7-8cc4-b69ce984befa", Identity: "U"},
		{Name: "Pull from Tomorrow", OracleID: "b1a23235-3076-475c-a68a-db29cf2a9dba", Identity: "U"},
		{Name: "Stroke of Genius", OracleID: "0cc6d683-366f-4ae4-be60-20ad9621fdaf", Identity: "U"},

		// Tutors.
		{Name: "Demonic Tutor", OracleID: "82004860-e589-4e38-8d61-8c0210e4ea39", Identity: "B"},
		{Name: "Vampiric Tutor", OracleID: "ededbdae-d9dc-4206-9335-d7158f2d7700", Identity: "B"},
		{Name: "Diabolic Tutor", OracleID: "14589b6b-1814-46f9-a364-83cc15dacac2", Identity: "B"},

		// Mana rocks.
		{Name: "Sol Ring", OracleID: "6ad8011d-3471-4369-9d68-b264cc027487"},
		{Name: "Arcane Signet", OracleID: "0bc7f093-bef0-4f1a-852c-4b75ebf54838"},
		{Name: "Azorius Signet", OracleID: "e018773f-95b3-49a3-9674-6f04ddef2092", Identity: "WU"},
		{Name: "Dimir Signet", OracleID: "7d881c57-0bd9-4c57-aa4a-b10808b86143", Identity: "UB"},
		{Name: "Orzhov Signet", OracleID: "de3dcb5d-775a-479f-99f5-d1883ed9b1b5", Identity: "WB"},
		{Name: "Talisman of Progress", OracleID: "00e35322-1a9a-41e3-9ce1-359c8eaa3bc7", Identity: "WU"},
		{Name: "Mind Stone", OracleID: "c97361b5-af16-4a7b-af85-a429dbaf4ad2"},
		{Name: "Thought Vessel", OracleID: "9965d9c5-2ebf-4a6c-930e-55c5890979be"},
		{Name: "Commander's Sphere", OracleID: "0b67c4e2-f88b-4e01-85a1-9d5f5b8db13b"},

		// Taxes and tempo.
		{Name: "Land Tax", OracleID: "d2d9ecea-7925-420e-98b9-2f87f41f387c", Identity: "W"},
		{Name: "Kismet", OracleID: "81fdd1c4-d43b-4f8b-8712-7c2bf45a3e0b", Identity: "W"},

		// Lands.
		{Name: "Hallowed Fountain", OracleID: "f1750962-a87c-49f6-b731-02ae971ac6ea", Identity: "WU"},
		{Name: "Watery Grave", OracleID: "fc9ec820-4245-4a96-b009-5308a818ca58", Identity: "UB"},
		{Name: "Godless Shrine", OracleID: "73864fcc-1bde-4bc0-831e-2b93e546e417", Identity: "WB"},
		{Name: "Tundra", OracleID: "02418479-9455-417f-a6a1-004356faff37", Identity: "WU"},
		{Name: "Underground Sea", OracleID: "4b22be3a-8ce1-47d1-b82e-6c3ccfb0548b", Identity: "UB"},
		{Name: "Scrubland", OracleID: "c8d95ca8-7d12-4072-aeaf-e20f248c7e39", Identity: "WB"},
		{Name: "Glacial Fortress", OracleID: "027dd013-baa7-4111-b3c9-f4d1414e9c45", Identity: "WU"},
		{Name: "Drowned Catacomb", OracleID: "819fc966-434e-470f-91e9-a38df974ad17", Identity: "UB"},
		{Name: "Isolated Chapel", OracleID: "7e5d9efe-48a9-434b-bb09-056e0e09cc9a", Identity: "WB"},
		{Name: "Adarkar Wastes", OracleID: "d5ad26cc-2bdb-46b7-b8bf-dd099d5fa09b", Identity: "WU"},
		{Name: "Underground River", OracleID: "857febd9-cdd7-4f8e-a852-d88084b0cfbc", Identity: "UB"},
		{Name: "Caves of Koilos", OracleID: "33de01e9-ce5a-42d4-afcb-343cd54a6d80", Identity: "WB"},
		{Name: "Deserted Beach", OracleID: "f0ec8681-da50-466b-8cdd-1dc710deccd9", Identity: "WU"},
		{Name: "Shipwreck Marsh", OracleID: "5f42b67f-87fd-4f98-a0e8-0c8313f4bbc8", Identity: "UB"},
		{Name: "Shattered Sanctum", OracleID: "c854ecb0-cc60-4c48-a9aa-7f2348a7a8c6", Identity: "WB"},
		{Name: "Arcane Sanctum", OracleID: "7d7cf15c-06b9-4062-a1eb-32614c458a3b", Identity: "WUB"},
		{Name: "Raffine's Tower", OracleID: "6e9ef5ef-6aed-4d3e-a59b-9e3dc8740b1b", Identity: "WUB"},
		{Name: "Command Tower", OracleID: "0895c9b7-ae7d-4bb3-af17-3b75deb50a25"},
		{Name: "Exotic Orchard", OracleID: "27b047e3-0d41-45e2-98e9-9391d7923a1e"},
		{Name: "Reflecting Pool", OracleID: "67f43ac6-2a58-4b53-b5d7-0330e2a252e2"},
		{Name: "City of Brass", OracleID: "f25351e3-539b-4bbc-b92d-6480acf4d722"},
		{Name: "Mana Confluence", OracleID: "d0ee5bdc-2b69-4b73-9a20-ffcc18783b29"},
		{Name: "Ancient Tomb", OracleID: "23467047-6dba-4498-b783-1ebc4f74b8c2"},
		{Name: "Path of Ancestry", OracleID: "b473e293-59e3-4e04-acf2-622604aeb25f"},
		{Name: "Flooded Strand", OracleID: "f3c7af78-a77d-4134-82a2-a5ce84285a84"},
		{Name: "Polluted Delta", OracleID: "ef86989d-ce80-4e55-aece-7d11710eeffa"},
		{Name: "Marsh Flats", OracleID: "dab520d0-20b4-4273-ba6b-eb07f85ea433"},
		{Name: "Prismatic Vista", OracleID: "032b8a0d-491a-4a12-ab9f-689010054d5b"},
		{Name: "Bojuka Bog", OracleID: "04b7362d-0490-4cb0-b5d7-2a7732f659ce", Identity: "B"},
		{Name: "Plains", Count: 3, Basic: true},
		{Name: "Island", Count: 3, Basic: true},
		{Name: "Swamp", Count: 2, Basic: true},
	},
}
