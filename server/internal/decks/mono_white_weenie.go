package decks

// Strength in Numbers — weenie. Cheap white creatures first, tokens
// second, anthems to make the pile lethal. Adeline, Resplendent Cathar
// is a 1WW commander whose power is the number of creatures you
// control, and she makes a 1/1 Human for each opponent every time you
// attack, so the plan is to have more bodies than the table has
// answers. Forge and Fury (izzet-aggro) turns each body it makes into
// damage to each opponent; this deck wins by attacking with them.
//
// Added by #554 as the mono-coloured aggro option, and deliberately
// not a lifegain or enchantress deck: the creature count is the payoff,
// and Glorious Anthem, Intangible Virtue, Spear of Heliod, Cathars'
// Crusade and Elesh Norn are what turn it into damage. Two Swords,
// Bonesplitter, Lightning Greaves and Swiftfoot Boots give the swarm
// something to carry.
//
// There are no sweepers: Wrath of God, Settle the Wreckage and
// Armageddon were cut because this deck is the board. Raise the Alarm
// and Increasing Devotion are the instant- and sorcery-speed token
// makers; #3138 gave both a declared Purpose (ADR 0126 §6), which is
// what let them back in.
//
// Every non-basic card below is in the effect catalog and graded full,
// and none is on full_test.go's minor-caveats list. Brimaz, King of
// Oreskos (its blocking token is not blocking), Land Tax (the land-count
// check happens as the trigger goes on the stack) and Stoneforge Mystic
// (no put-onto-the-battlefield ability) carry caveats, and Thraben
// Inspector has not been graded, so none is here.

var monoWhiteWeenie = Deck{
	ID:        "mono-white-weenie",
	Name:      "Strength in Numbers",
	Archetype: "weenie",
	Identity:  "W",
	Summary:   "Go wide with cheap white creatures and tokens: Adeline makes a Human every time you attack, and anthems turn the swarm into a lethal one.",
	Commander: Card{Name: "Adeline, Resplendent Cathar", OracleID: "38515f89-348b-4cf3-b7bd-1f6fe4ce2fba", Identity: "W"},
	Mainboard: []Card{
		// One-drops.
		{Name: "Mother of Runes", OracleID: "60433b48-d27f-413c-905c-43839b1943f1", Identity: "W"},
		{Name: "Esper Sentinel", OracleID: "5def9f38-0a0b-4e8d-9f9d-29dcb46520b4", Identity: "W"},
		{Name: "Soul Warden", OracleID: "f3fad295-1af2-4ecc-8546-b121ad6be27b", Identity: "W"},
		{Name: "Usher of the Fallen", OracleID: "a6eb06dc-a62d-4fdb-b336-e304d8d68c92", Identity: "W"},
		{Name: "Ocelot Pride", OracleID: "b4ce2c8f-a19c-461e-8a2e-0ec2bd3ebca3", Identity: "W"},
		{Name: "Doomed Traveler", OracleID: "a30907c0-fbde-4fd3-a8c7-f304305fcea7", Identity: "W"},
		{Name: "Mikaeus, the Lunarch", OracleID: "82f3faa8-39fa-450b-843f-d60a4c36d8f7", Identity: "W"},
		{Name: "Skymarcher Aspirant", OracleID: "0b6db929-1b6a-4372-8146-aa047258b552", Identity: "W"},
		{Name: "Kinsbaile Aspirant", OracleID: "845d80ee-1cd5-437a-90a5-40dcc657d3aa", Identity: "W"},

		// Two-drops.
		{Name: "Thalia, Guardian of Thraben", OracleID: "9b7f1d05-707c-4ed3-9f0e-8ced1232c2ee", Identity: "W"},
		{Name: "Selfless Spirit", OracleID: "71d785a9-ddc8-472e-a778-a551b444a4bd", Identity: "W"},
		{Name: "Ajani's Pridemate", OracleID: "95e94dea-5ac0-4d6f-adec-ca147aee861f", Identity: "W"},
		{Name: "Knight of the White Orchid", OracleID: "7d0efc64-151f-45b4-8c0b-c32609f9d862", Identity: "W"},
		{Name: "Charming Prince", OracleID: "c48d844c-3976-4fa5-8e0d-3f0e535e7619", Identity: "W"},
		{Name: "Cathar Commando", OracleID: "774dce79-67e0-4820-8013-c7a7347993ce", Identity: "W"},
		{Name: "Lion Sash", OracleID: "50330e48-db74-4c5f-a0bc-f9a8607e8f31", Identity: "W"},
		{Name: "Sram, Senior Edificer", OracleID: "7e00b0cd-d212-4604-ba07-da21f4fe00b0", Identity: "W"},
		{Name: "Loyal Warhound", OracleID: "cc6a83c7-e645-4a53-9550-be79b42cd851", Identity: "W"},

		// Three-drops.
		{Name: "Thalia, Heretic Cathar", OracleID: "13be47e4-9da2-4f9e-baf0-db96f7777ccb", Identity: "W"},
		{Name: "Mentor of the Meek", OracleID: "b9f4f96b-6e54-4fe6-8df7-623e0fc72409", Identity: "W"},
		{Name: "Monastery Mentor", OracleID: "3979067a-9c68-443d-a85f-d9f07be880b9", Identity: "W"},
		{Name: "Mirror Entity", OracleID: "17e905ca-c0bd-473d-95a7-e180ba5fea43", Identity: "W"},
		{Name: "Ranger-Captain of Eos", OracleID: "cada3481-cc2b-4412-b9b5-0436af53aad2", Identity: "W"},
		{Name: "Heliod, Sun-Crowned", OracleID: "63e596a2-9126-4af6-8782-c38687d664ad", Identity: "W"},
		{Name: "Lyra, Archangel of Dawn", OracleID: "0fca3328-484f-4050-925e-91840710b5d6", Identity: "W"},
		{Name: "Righteous Valkyrie", OracleID: "891d2690-7144-4f87-b6ef-96f2469780a9", Identity: "W"},
		{Name: "Inspiring Overseer", OracleID: "d646e42b-5635-4798-b633-29c093b66a55", Identity: "W"},
		{Name: "Arahbo, the First Fang", OracleID: "dfec65a0-6fb6-463d-aa62-e84a427db217", Identity: "W"},

		// Top end.
		{Name: "Angelic Field Marshal", OracleID: "502575f4-7c56-44bc-a77f-ae28d66c8e1f", Identity: "W"},
		{Name: "Emeria Angel", OracleID: "ea6616e4-db8d-4905-80f8-cb0162906850", Identity: "W"},
		{Name: "Linvala, Keeper of Silence", OracleID: "88dadc31-dfac-41b3-bf2a-65fa89e3c16d", Identity: "W"},
		{Name: "Angel of Invention", OracleID: "28c7f2b6-ba67-4ccf-9e1b-99c89a9d1f72", Identity: "W"},
		{Name: "Archangel of Thune", OracleID: "4f2d4538-dc1d-4c09-964b-b0d7c240fb7d", Identity: "W"},
		{Name: "Baneslayer Angel", OracleID: "0e11792b-7fe5-4208-aa0b-e5d09b2b65fe", Identity: "W"},
		{Name: "Lyra Dawnbringer", OracleID: "592c91fc-6430-4c76-9460-65f047350f67", Identity: "W"},
		{Name: "Reveillark", OracleID: "1be13ede-98f8-497e-800c-03e5802932b3", Identity: "W"},
		{Name: "Sun Titan", OracleID: "b2e950fb-cb7e-40a0-a311-5bbdd0477b29", Identity: "W"},
		{Name: "Elesh Norn, Grand Cenobite", OracleID: "958d71ff-c9f7-46f0-96ca-79e7f4d65a16", Identity: "W"},

		// Tokens and anthems.
		{Name: "Raise the Alarm", OracleID: "5b2364d7-a811-4595-a1b4-224c70555ffa", Identity: "W"},
		{Name: "Call the Coppercoats", OracleID: "73d8c33d-916a-4220-96ae-9622aad36210", Identity: "W"},
		{Name: "Secure the Wastes", OracleID: "b2347910-d6c6-4681-8316-7ef27056485c", Identity: "W"},
		{Name: "Elspeth, Sun's Champion", OracleID: "05e6b243-48a6-4a42-bc5f-413441de9c33", Identity: "W"},
		{Name: "Anointed Procession", OracleID: "7246d45b-2185-4cdd-981b-5419b7d52bce", Identity: "W"},
		{Name: "Intangible Virtue", OracleID: "d21c3c8f-d105-4ba9-bf69-e5f26f0f8ec5", Identity: "W"},
		{Name: "Glorious Anthem", OracleID: "e3886fe8-9b76-4613-8891-4ec74657c087", Identity: "W"},
		{Name: "Spear of Heliod", OracleID: "fd66aa66-75a2-41b6-8d57-dc4b9c221ccb", Identity: "W"},
		{Name: "Cathars' Crusade", OracleID: "cc65ac73-5bef-4ecb-ad8e-39199084c027", Identity: "W"},
		{Name: "Increasing Devotion", OracleID: "7a5ff4d4-27b7-47d4-ba88-970c63c4e3fb", Identity: "W"},
		{Name: "Smothering Tithe", OracleID: "153376c9-dffd-458c-8ce3-a4c8269bc4e9", Identity: "W"},

		// Removal and protection.
		{Name: "Swords to Plowshares", OracleID: "b1544f21-7e98-461b-aed5-e748b0168c52", Identity: "W"},
		{Name: "Path to Exile", OracleID: "d683d985-9888-4d21-8b5f-69e69ce4a03b", Identity: "W"},
		{Name: "Council's Judgment", OracleID: "acf4d685-a718-42f7-9a7a-dee069ae8db2", Identity: "W"},
		{Name: "Oblivion Ring", OracleID: "bd9b9772-f5f9-4c6b-913e-7193bea5d0a7", Identity: "W"},
		{Name: "Fateful Absence", OracleID: "35bba442-1aec-4d33-b502-4c580d61644b", Identity: "W"},
		{Name: "Generous Gift", OracleID: "fae37e28-e137-4177-b973-fa8b4dd8f409", Identity: "W"},
		{Name: "Flawless Maneuver", OracleID: "4e183439-17d2-47ff-9d99-5e22821d91e3", Identity: "W"},
		{Name: "Teferi's Protection", OracleID: "0d4ecdb1-ec90-497f-a7a4-1c68092b8757", Identity: "W"},

		// Equipment and mana.
		{Name: "Bonesplitter", OracleID: "452e3f5f-ce17-4682-966b-5cc100210aee"},
		{Name: "Swiftfoot Boots", OracleID: "c8b143ad-43ec-4e0d-a440-e348daa31391"},
		{Name: "Lightning Greaves", OracleID: "ca204b66-8d0c-431a-8d34-282f7c2d17da"},
		{Name: "Sword of Fire and Ice", OracleID: "2ccdc60a-49a9-44b9-a7af-0ebf18b26785"},
		{Name: "Sword of Feast and Famine", OracleID: "d0901053-6de0-46d0-9ee3-8d40510236c1"},
		{Name: "Sol Ring", OracleID: "6ad8011d-3471-4369-9d68-b264cc027487"},
		{Name: "Arcane Signet", OracleID: "0bc7f093-bef0-4f1a-852c-4b75ebf54838"},
		{Name: "Mind Stone", OracleID: "c97361b5-af16-4a7b-af85-a429dbaf4ad2"},

		// Lands.
		{Name: "Castle Ardenvale", OracleID: "f8f4fc60-725d-46d8-8e8f-e68e00d20589", Identity: "W"},
		{Name: "Myriad Landscape", OracleID: "2549bc57-9ffb-4053-9f10-f2a5f792b845"},
		{Name: "Arch of Orazca", OracleID: "3bb518ff-399b-4ce7-b9ad-a1d563dd7792"},
		{Name: "Plains", Count: 31, Basic: true},
	},
}
