package decks

// Dragon Hoard — tribal. Five colours, one tribe. The Ur-Dragon makes
// every other Dragon spell cost {1} less from the command zone or the
// battlefield, and whenever Dragons attack it draws that many cards and
// puts a permanent from the hand onto the battlefield, so the deck is
// ramp into a table of fliers that pay for themselves. Tiamat fetches
// five Dragons with different names the turn it lands.
//
// Added for ADR 0172 decision 8 (#3401): a built-in deck whose
// commander is five-colour, so it can carry Jegantha, the Wellspring
// as its companion and the companion path runs in ordinary games.
// Jegantha asks that no card in the starting deck has more than one of
// the same mana symbol in its mana cost, which rules out every card
// with {W}{W}, {2}{U}{U} and the like: no Wrath of God, no Sun Titan.
// The list is built around that, and TestCompanionDeckFitsItsCompanion
// checks it against the same registry the server validates with.
//
// Every non-basic card below is graded full in the catalog (no entry
// on full_test.go's minor-caveats list); decks_test.go and full_test.go
// fail the build if that stops being true.

var dragonHoard = Deck{
	ID:        "wubrg-dragons",
	Name:      "Dragon Hoard",
	Archetype: "tribal",
	Identity:  "WUBRG",
	Summary:   "Five-colour Dragons: ramp into big fliers that The Ur-Dragon makes cheaper and turns into cards and free permanents, with Jegantha, the Wellspring waiting outside the game as its companion.",
	Commander: Card{Name: "The Ur-Dragon", OracleID: "87b22b09-4f6d-4bc5-9cfc-663e4c7c6981", Identity: "WUBRG"},
	Companion: &Card{Name: "Jegantha, the Wellspring", OracleID: "f0587c55-06c1-4930-a62e-668f37464ce9", Identity: "WUBRG"},
	Mainboard: []Card{
		// Dragons. Tiamat fetches five of them; the rest are the big fliers and ETB bodies that pay for the mana.
		{Name: "Tiamat", OracleID: "f00e4df1-13fb-4514-8abb-92954068689a", Identity: "WUBRG"},
		{Name: "Miirym, Sentinel Wyrm", OracleID: "880fcf86-e6a9-482c-b91d-750f293127b2", Identity: "URG"},
		{Name: "Korvold, Fae-Cursed King", OracleID: "9ae669dd-7e60-4649-b96e-35da28be641a", Identity: "BRG"},
		{Name: "Dragonlord Dromoka", OracleID: "82cdd612-3322-4372-810f-1ff106ea8e6a", Identity: "WG"},
		{Name: "Dragonlord Atarka", OracleID: "daff70a4-a990-47c9-b6ef-2c379272c8a3", Identity: "RG"},
		{Name: "Silverquill, the Disputant", OracleID: "66deb3f3-0314-427e-9cb8-6befa365ae94", Identity: "WB"},
		{Name: "Neriv, Crackling Vanguard", OracleID: "c8f827d7-4740-43c8-b494-c9e0479e0bf7", Identity: "WBR"},
		{Name: "Ziatora, the Incinerator", OracleID: "d46c3fb6-f1c7-4a96-ae42-5bce17fc7c1d", Identity: "BRG"},
		{Name: "Savage Ventmaw", OracleID: "f743d85f-754e-4e70-9907-af54acb9e8d4", Identity: "RG"},
		{Name: "Armament Dragon", OracleID: "23c75386-8011-4ee4-97e2-eaafeb20b788", Identity: "WBG"},
		{Name: "Rapacious Dragon", OracleID: "0944ec2e-1dd9-459f-8f1d-667242cf52fe", Identity: "R"},
		{Name: "Lurking Green Dragon", OracleID: "262e1cf6-61c8-444d-a71b-050eebcaf932", Identity: "G"},
		{Name: "Covetous Dragon", OracleID: "bf7710cb-70cb-42fb-b840-cc4f385daa7a", Identity: "R"},
		{Name: "Ganax, Astral Hunter", OracleID: "a3112971-3af4-46e8-b2d6-a8759b39d0d1", Identity: "R"},
		{Name: "Dragonborn Champion", OracleID: "9e51ccf9-d55d-4231-8ff5-ef3ca1485c8d", Identity: "RG"},
		{Name: "Betor, Ancestor's Voice", OracleID: "990b5e12-6e04-4832-9d64-87278f12cbda", Identity: "WBG"},
		{Name: "Teval, Arbiter of Virtue", OracleID: "f6cf359a-91cb-4b3c-8837-be53e4ed9c93", Identity: "UBG"},
		{Name: "Ureni, the Song Unending", OracleID: "1c995c80-3301-409b-822b-9297aa260823", Identity: "URG"},
		{Name: "Beledros Witherbloom", OracleID: "90194ff1-db61-463f-b5a3-15cd85311d0e", Identity: "BG"},
		// Other creatures.
		{Name: "Birds of Paradise", OracleID: "d3a0b660-358c-41bd-9cd2-41fbf3491b1a", Identity: "G"},
		{Name: "Llanowar Elves", OracleID: "68954295-54e3-4303-a6bc-fc4547a4e3a3", Identity: "G"},
		{Name: "Solemn Simulacrum", OracleID: "00c0543c-2a1f-4425-8283-4062d74a1637"},
		{Name: "Burnished Hart", OracleID: "893fed41-c144-433f-af88-bc7d419b7fb3"},
		{Name: "Wurmcoil Engine", OracleID: "d1a60f44-7696-49ee-91fb-cab5b3102962"},
		{Name: "Reveillark", OracleID: "1be13ede-98f8-497e-800c-03e5802932b3", Identity: "W"},
		{Name: "Mulldrifter", OracleID: "24d0f5e7-0d9e-4b76-900e-a7274e80312d", Identity: "U"},
		{Name: "Hydroid Krasis", OracleID: "6bd872b2-5c40-4e11-9a7f-0136a51b0642", Identity: "UG"},
		// Mana rocks.
		{Name: "Sol Ring", OracleID: "6ad8011d-3471-4369-9d68-b264cc027487"},
		{Name: "Arcane Signet", OracleID: "0bc7f093-bef0-4f1a-852c-4b75ebf54838"},
		{Name: "Fellwar Stone", OracleID: "95560508-7ac9-4be9-8a3f-3c7d5b52807b"},
		{Name: "Mind Stone", OracleID: "c97361b5-af16-4a7b-af85-a429dbaf4ad2"},
		{Name: "Thought Vessel", OracleID: "9965d9c5-2ebf-4a6c-930e-55c5890979be"},
		{Name: "Azorius Signet", OracleID: "e018773f-95b3-49a3-9674-6f04ddef2092", Identity: "WU"},
		{Name: "Dimir Signet", OracleID: "7d881c57-0bd9-4c57-aa4a-b10808b86143", Identity: "UB"},
		{Name: "Orzhov Signet", OracleID: "de3dcb5d-775a-479f-99f5-d1883ed9b1b5", Identity: "WB"},
		{Name: "Boros Signet", OracleID: "41c84665-1f99-40ab-aaca-1188649eb263", Identity: "WR"},
		{Name: "Simic Signet", OracleID: "44503105-3e13-408d-a44f-37d503c61d72", Identity: "UG"},
		{Name: "Talisman of Curiosity", OracleID: "8c34b089-aad1-476e-958a-3077bf1bbb51", Identity: "UG"},
		// Ramp spells.
		{Name: "Cultivate", OracleID: "8b755881-a72d-4e21-a369-d2924eb4585a", Identity: "G"},
		{Name: "Kodama's Reach", OracleID: "1593ea18-2f2f-4ab4-83fb-6ccc0bec8a90", Identity: "G"},
		{Name: "Farseek", OracleID: "495e52e6-4c2b-4574-9474-eadbdcc8b4ac", Identity: "G"},
		{Name: "Nature's Lore", OracleID: "78826359-fe63-44ad-adc4-a17ffcd710e4", Identity: "G"},
		{Name: "Three Visits", OracleID: "1b882a0e-0ede-4d1a-bd1a-9b7cffbcde8e", Identity: "G"},
		// Removal.
		{Name: "Swords to Plowshares", OracleID: "b1544f21-7e98-461b-aed5-e748b0168c52", Identity: "W"},
		{Name: "Path to Exile", OracleID: "d683d985-9888-4d21-8b5f-69e69ce4a03b", Identity: "W"},
		{Name: "Doom Blade", OracleID: "59e7f2ae-4535-4191-98be-3e65b6b2befa", Identity: "B"},
		{Name: "Go for the Throat", OracleID: "2f092562-9e17-43cd-aeb8-d0567f99363e", Identity: "B"},
		{Name: "Infernal Grasp", OracleID: "94f0a572-e91c-4b56-a5d1-6cbbeabd210d", Identity: "B"},
		{Name: "Despark", OracleID: "bd16434d-55ea-4c5a-a9ef-752971a4af16", Identity: "WB"},
		{Name: "Mortify", OracleID: "faa01ed1-ccfa-4e58-951f-cd81f9068027", Identity: "WB"},
		{Name: "Utter End", OracleID: "cba94732-1f1f-4ffd-9aba-45db990043fa", Identity: "WB"},
		{Name: "Anguished Unmaking", OracleID: "ad09b3c3-c8e7-481c-8c45-e7f234935117", Identity: "WB"},
		{Name: "Chaos Warp", OracleID: "07a0cba9-8768-4fd9-a3d5-b0f83b4bf8e8", Identity: "R"},
		{Name: "Beast Within", OracleID: "7735eeba-693b-47e2-bd51-414379cf1016", Identity: "G"},
		// Card advantage and protection.
		{Name: "Rhystic Study", OracleID: "53236dd7-845a-444c-96d5-f41ed7325d8f", Identity: "U"},
		{Name: "Fact or Fiction", OracleID: "437b2dab-15e0-4b9a-a204-58622d37a3b3", Identity: "U"},
		{Name: "Night's Whisper", OracleID: "7ffae8f8-3006-4969-a339-6d30678f87ea", Identity: "B"},
		{Name: "Read the Bones", OracleID: "5bf4d8d9-a2b2-4dba-ac05-9d4470a89db2", Identity: "B"},
		{Name: "Cyclonic Rift", OracleID: "d75b9c82-1b49-4c3e-a1b5-aeef57d6644b", Identity: "U"},
		{Name: "Sylvan Library", OracleID: "92eed395-62ca-4293-882b-8565c40daab5", Identity: "G"},
		{Name: "Lightning Greaves", OracleID: "ca204b66-8d0c-431a-8d34-282f7c2d17da"},
		{Name: "Swiftfoot Boots", OracleID: "c8b143ad-43ec-4e0d-a440-e348daa31391"},
		{Name: "Teferi's Protection", OracleID: "0d4ecdb1-ec90-497f-a7a4-1c68092b8757", Identity: "W"},
		{Name: "Fierce Guardianship", OracleID: "d09c9cba-fdd2-479b-ad5d-d05181c3e3f9", Identity: "U"},
		{Name: "Swan Song", OracleID: "8ddfc283-c9b4-41a5-af88-cf0068e986cc", Identity: "U"},
		// Lands.
		{Name: "Command Tower", OracleID: "0895c9b7-ae7d-4bb3-af17-3b75deb50a25"},
		{Name: "Exotic Orchard", OracleID: "27b047e3-0d41-45e2-98e9-9391d7923a1e"},
		{Name: "Reflecting Pool", OracleID: "67f43ac6-2a58-4b53-b5d7-0330e2a252e2"},
		{Name: "City of Brass", OracleID: "f25351e3-539b-4bbc-b92d-6480acf4d722"},
		{Name: "Mana Confluence", OracleID: "d0ee5bdc-2b69-4b73-9a20-ffcc18783b29"},
		{Name: "Prismatic Vista", OracleID: "032b8a0d-491a-4a12-ab9f-689010054d5b"},
		{Name: "Fabled Passage", OracleID: "0c85b8f7-0bd0-4680-9ec5-d4b110460a54"},
		{Name: "Hallowed Fountain", OracleID: "f1750962-a87c-49f6-b731-02ae971ac6ea", Identity: "WU"},
		{Name: "Watery Grave", OracleID: "fc9ec820-4245-4a96-b009-5308a818ca58", Identity: "UB"},
		{Name: "Godless Shrine", OracleID: "73864fcc-1bde-4bc0-831e-2b93e546e417", Identity: "WB"},
		{Name: "Breeding Pool", OracleID: "20283c4a-f1f0-42f0-bc08-6da87474426b", Identity: "UG"},
		{Name: "Sacred Foundry", OracleID: "45181cb8-2090-4471-ba90-e5a8f04d525f", Identity: "WR"},
		{Name: "Tundra", OracleID: "02418479-9455-417f-a6a1-004356faff37", Identity: "WU"},
		{Name: "Underground Sea", OracleID: "4b22be3a-8ce1-47d1-b82e-6c3ccfb0548b", Identity: "UB"},
		{Name: "Scrubland", OracleID: "c8d95ca8-7d12-4072-aeaf-e20f248c7e39", Identity: "WB"},
		{Name: "Tropical Island", OracleID: "74b7fe23-5d3a-4092-8d78-7c0eba8f6f73", Identity: "UG"},
		{Name: "Flooded Strand", OracleID: "f3c7af78-a77d-4134-82a2-a5ce84285a84"},
		{Name: "Polluted Delta", OracleID: "ef86989d-ce80-4e55-aece-7d11710eeffa"},
		{Name: "Marsh Flats", OracleID: "dab520d0-20b4-4273-ba6b-eb07f85ea433"},
		{Name: "Misty Rainforest", OracleID: "09dd85aa-47bc-4713-a9b9-8b52ff2285ed"},
		{Name: "Verdant Catacombs", OracleID: "67d60b24-d429-4ded-90d9-06e49f28c396"},
		{Name: "Bloodstained Mire", OracleID: "fc0707c7-d504-4ccf-a0d2-3eb6e26e7a57"},
		{Name: "Arcane Sanctum", OracleID: "7d7cf15c-06b9-4062-a1eb-32614c458a3b", Identity: "WUB"},
		{Name: "Raffine's Tower", OracleID: "6e9ef5ef-6aed-4d3e-a59b-9e3dc8740b1b", Identity: "WUB"},
		{Name: "Plains", Count: 2, Basic: true},
		{Name: "Island", Count: 2, Basic: true},
		{Name: "Swamp", Count: 2, Basic: true},
		{Name: "Mountain", Count: 2, Basic: true},
		{Name: "Forest", Count: 2, Basic: true},
	},
}
