package decks

// Body Count — aristocrats. Put a payoff down, put a body down,
// sacrifice the body, then bring it back. The deck teaches death
// triggers: Syr Konrad and the drain payoffs pay out on every creature
// that dies, and the reanimation spells turn the graveyard into a
// second hand. It wins without attacking, which makes it the one
// archetype here that does not depend on the combat heuristics being
// any good.
//
// An upgraded precon (#2436): no free mana (Dark Ritual, Phyrexian
// Altar), no hand disruption and fewer tutors. The cuts went to
// reanimation and death payoffs, not to raw power.
//
// Every non-basic card below is in the effect catalog and graded full,
// apart from the minor caveats full_test.go names; decks_test.go and
// full_test.go fail the build if that stops being true.

var monoBlackAristocrats = Deck{
	ID:        "mono-black-aristocrats",
	Name:      "Body Count",
	Archetype: "aristocrats",
	Identity:  "B",
	Summary:   "Death triggers: sacrifice your creatures for value, drain the table each time one dies, then reanimate the best of them and do it again.",
	Commander: Card{Name: "Syr Konrad, the Grim", OracleID: "14c3ff84-1e82-4606-a433-869fc52cc382", Identity: "B"},
	Mainboard: []Card{
		// Sacrifice outlets.
		{Name: "Carrion Feeder", OracleID: "a1cc5e37-b09a-4b7f-afd5-77c1c35aa425", Identity: "B"},
		{Name: "Viscera Seer", OracleID: "f82a4e85-526d-4456-b700-7760043a31be", Identity: "B"},
		{Name: "Ashnod's Altar", OracleID: "4d18bcba-a346-445e-a182-6cc30b7e066d"},
		{Name: "Warren Soultrader", OracleID: "ace86e56-efde-4eb7-8815-71456a4c3abe", Identity: "B"},
		{Name: "Woe Strider", OracleID: "3adbd963-e85d-4569-963a-4472594f06f9", Identity: "B"},

		// Death payoffs. The reason the outlets above are worth having.
		{Name: "Blood Artist", OracleID: "310f141c-7f37-4729-aed6-dd9c09db448d", Identity: "B"},
		{Name: "Falkenrath Noble", OracleID: "3739b179-bc81-4737-8376-66a57e16b942", Identity: "B"},
		{Name: "Vindictive Vampire", OracleID: "11ad664e-36d3-4d5b-8a87-59ea715877e0", Identity: "B"},
		{Name: "Bastion of Remembrance", OracleID: "c7f33cea-2ec8-4081-9208-a5b1d86721b3", Identity: "B"},
		{Name: "Mirkwood Bats", OracleID: "0636b6c3-0662-420a-b30d-f0a14e7c512d", Identity: "B"},
		{Name: "Midnight Reaper", OracleID: "e8c7566d-7cc0-48af-a986-83223ec7e06c", Identity: "B"},
		{Name: "Pitiless Plunderer", OracleID: "a784481f-eccb-4112-bb38-04a659319660", Identity: "B"},
		{Name: "Morbid Opportunist", OracleID: "322f44f0-e6da-4ee0-b474-e7d5e9a461c5", Identity: "B"},
		{Name: "Plague Belcher", OracleID: "f086a82d-8dbe-4d02-a076-03789704e9d4", Identity: "B"},
		{Name: "Grave Pact", OracleID: "6f4ac4a4-53ec-4bc9-8f5c-d4b801d867b2", Identity: "B"},
		{Name: "Dictate of Erebos", OracleID: "7c777a41-e40a-4b40-96bf-8ddd5c12924c", Identity: "B"},
		{Name: "Butcher of Malakir", OracleID: "a85197ab-dc94-4b72-9716-8dbdbbe90ff8", Identity: "B"},

		// Bodies to feed the outlets, edicts that make the table feed them
		// too, and creatures that come back on their own.
		{Name: "Typhoid Rats", OracleID: "d6ee6cc1-902d-4f56-afa5-6fa4813bfbbc", Identity: "B"},
		{Name: "Fleshbag Marauder", OracleID: "4b1bf05e-753e-4350-a913-894cf3cecc0c", Identity: "B"},
		{Name: "Accursed Marauder", OracleID: "d8ad23a1-0b43-48ea-9fbe-d89b29194509", Identity: "B"},
		{Name: "Plaguecrafter", OracleID: "2e3ee458-fa3f-4452-ab95-5a7fb5a0483b", Identity: "B"},
		{Name: "Merciless Executioner", OracleID: "c3c45d50-9038-41df-bb2f-9bc40071845b", Identity: "B"},
		{Name: "Bloodghast", OracleID: "e97f9c2b-b41e-4f36-9245-77c0ac125647", Identity: "B"},
		{Name: "Undead Butler", OracleID: "426fa8e8-b0f9-4b6c-b22b-04ec2073187a", Identity: "B"},
		{Name: "Phyrexian Delver", OracleID: "a13cbac0-4c76-4970-b61e-5f4e020ee95c", Identity: "B"},
		{Name: "Gray Merchant of Asphodel", OracleID: "38f3b157-0df4-409b-89cc-086e1531cd5b", Identity: "B"},
		{Name: "Sheoldred, the Apocalypse", OracleID: "34f34409-326d-4994-a0ea-1a69aa278f03", Identity: "B"},
		{Name: "Solemn Simulacrum", OracleID: "00c0543c-2a1f-4425-8283-4062d74a1637"},
		{Name: "Burnished Hart", OracleID: "893fed41-c144-433f-af88-bc7d419b7fb3"},
		{Name: "Wurmcoil Engine", OracleID: "d1a60f44-7696-49ee-91fb-cab5b3102962"},

		// Reanimation. The graveyard is a second hand, and the self-mill
		// fills it.
		{Name: "Reanimate", OracleID: "a044474a-cd72-4e9d-bd8d-a08f2de9cdc0", Identity: "B"},
		{Name: "Dread Return", OracleID: "352b64d2-2ae5-44ee-a64f-94932ef545d3", Identity: "B"},
		{Name: "Zombify", OracleID: "bb95db4d-5017-4121-bf79-d68476602d8c", Identity: "B"},
		{Name: "Stitch Together", OracleID: "bc3d5911-3580-4132-9daf-2826495b5739", Identity: "B"},
		{Name: "Living Death", OracleID: "9e6a3df4-67a3-452e-a6ef-f04dbadb21ef", Identity: "B"},
		{Name: "Rise of the Dark Realms", OracleID: "e5223a09-f732-4747-8914-e6546ab0ef4c", Identity: "B"},
		{Name: "Dread Summons", OracleID: "642fa01d-025e-44dd-8360-5325e5a28282", Identity: "B"},
		{Name: "Entomb", OracleID: "299fc083-0834-4064-8344-f895aff68867", Identity: "B"},
		{Name: "Vile Entomber", OracleID: "556dd333-4066-41f7-98f6-41794754de71", Identity: "B"},

		// Draw and drain.
		{Name: "Altar's Reap", OracleID: "6a125750-2b8c-4f9d-8173-ac8d14c91ddb", Identity: "B"},
		{Name: "Deadly Dispute", OracleID: "457af74a-02b3-4659-846d-63e482667f34", Identity: "B"},
		{Name: "Night's Whisper", OracleID: "7ffae8f8-3006-4969-a339-6d30678f87ea", Identity: "B"},
		{Name: "Sign in Blood", OracleID: "c6207f6a-a624-4754-88f5-dbe700c841ff", Identity: "B"},
		{Name: "Ambition's Cost", OracleID: "84de4fec-2f38-4293-93d3-b3882c5aac14", Identity: "B"},
		{Name: "Read the Bones", OracleID: "5bf4d8d9-a2b2-4dba-ac05-9d4470a89db2", Identity: "B"},
		{Name: "Exsanguinate", OracleID: "8164b1e8-3350-465e-8a17-75f57d326344", Identity: "B"},
		{Name: "Sanguine Bond", OracleID: "73089a39-a2f6-4aa2-a058-e6551475153d", Identity: "B"},
		{Name: "Exquisite Blood", OracleID: "8f933fae-6c0c-42d7-a817-14760d8285cd", Identity: "B"},

		// Removal.
		{Name: "Doom Blade", OracleID: "59e7f2ae-4535-4191-98be-3e65b6b2befa", Identity: "B"},
		{Name: "Go for the Throat", OracleID: "2f092562-9e17-43cd-aeb8-d0567f99363e", Identity: "B"},
		{Name: "Infernal Grasp", OracleID: "94f0a572-e91c-4b56-a5d1-6cbbeabd210d", Identity: "B"},
		{Name: "Feed the Swarm", OracleID: "5825997b-10d7-4a36-972c-a80ddd90b8ed", Identity: "B"},
		{Name: "Withering Torment", OracleID: "ffce81c5-1b58-4882-a4e7-6f8d7cb170de", Identity: "B"},
		{Name: "Ashes to Ashes", OracleID: "944a52d9-bb14-43c5-8d05-2afaf023dc9f", Identity: "B"},
		{Name: "Damnation", OracleID: "d57a8f0b-7989-4db5-8756-6f2690097252", Identity: "B"},

		// Tutors.
		{Name: "Demonic Tutor", OracleID: "82004860-e589-4e38-8d61-8c0210e4ea39", Identity: "B"},
		{Name: "Diabolic Tutor", OracleID: "14589b6b-1814-46f9-a364-83cc15dacac2", Identity: "B"},

		// Mana.
		{Name: "Sol Ring", OracleID: "6ad8011d-3471-4369-9d68-b264cc027487"},
		{Name: "Arcane Signet", OracleID: "0bc7f093-bef0-4f1a-852c-4b75ebf54838"},
		{Name: "Mind Stone", OracleID: "c97361b5-af16-4a7b-af85-a429dbaf4ad2"},
		{Name: "Thought Vessel", OracleID: "9965d9c5-2ebf-4a6c-930e-55c5890979be"},
		{Name: "Commander's Sphere", OracleID: "0b67c4e2-f88b-4e01-85a1-9d5f5b8db13b"},
		{Name: "Worn Powerstone", OracleID: "b166b670-febc-4821-855e-f8d465644c03"},
		{Name: "Hedron Archive", OracleID: "32263baa-d3f0-463f-92b3-4e9938476add"},

		// Lands.
		{Name: "Command Tower", OracleID: "0895c9b7-ae7d-4bb3-af17-3b75deb50a25"},
		{Name: "Ancient Tomb", OracleID: "23467047-6dba-4498-b783-1ebc4f74b8c2"},
		{Name: "Path of Ancestry", OracleID: "b473e293-59e3-4e04-acf2-622604aeb25f"},
		{Name: "Cabal Coffers", OracleID: "7358e164-5704-4e78-9b21-6a9bf2a968ce", Identity: "B"},
		{Name: "Urborg, Tomb of Yawgmoth", OracleID: "db6174d7-211d-4817-b8e4-8384594c83f9"},
		{Name: "Phyrexian Tower", OracleID: "1861e642-21d5-4232-89f3-b5557f2946c1", Identity: "B"},
		{Name: "High Market", OracleID: "86fb3749-37d6-48a6-8524-71e996850307"},
		{Name: "Bojuka Bog", OracleID: "04b7362d-0490-4cb0-b5d7-2a7732f659ce", Identity: "B"},
		{Name: "Strip Mine", OracleID: "d21a89eb-7c5b-459a-acc7-12b20b13bf79"},
		{Name: "Myriad Landscape", OracleID: "2549bc57-9ffb-4053-9f10-f2a5f792b845"},
		{Name: "Temple of the False God", OracleID: "cfdd5dc6-593e-495a-8cfe-3a56b3c4c7df"},
		{Name: "Bloodstained Mire", OracleID: "fc0707c7-d504-4ccf-a0d2-3eb6e26e7a57"},
		{Name: "Marsh Flats", OracleID: "dab520d0-20b4-4273-ba6b-eb07f85ea433"},
		{Name: "Prismatic Vista", OracleID: "032b8a0d-491a-4a12-ab9f-689010054d5b"},
		{Name: "Evolving Wilds", OracleID: "a75445d3-1303-4bb5-89ad-26ea93fecd48"},
		{Name: "Terramorphic Expanse", OracleID: "1bd3e453-aa21-4ee6-95c2-d6d920ee8e7a"},
		{Name: "Fabled Passage", OracleID: "0c85b8f7-0bd0-4680-9ec5-d4b110460a54"},
		{Name: "Swamp", Count: 18, Basic: true},
	},
}
