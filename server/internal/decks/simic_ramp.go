package decks

// Deep Roots — ramp-stompy. Ramp, draw off the ramp, cast something
// enormous. The deck is deliberately forgiving of poor sequencing:
// almost every ramp spell is fine on any turn, and the top end wins on
// raw stats rather than on a combination the policy would have to
// find.
//
// Every non-basic card below is in the effect catalog; decks_test.go
// fails the build if that stops being true.

var simicRamp = Deck{
	ID:        "simic-ramp",
	Name:      "Deep Roots",
	Archetype: "ramp-stompy",
	Identity:  "UG",
	Summary:   "Mana creatures and land ramp into large green threats, with the commander drawing a card on every land drop.",
	Commander: Card{Name: "Tatyova, Benthic Druid", OracleID: "0715e860-3b3b-4331-9718-207973e94fee", Identity: "UG"},
	Mainboard: []Card{
		// Mana creatures.
		{Name: "Llanowar Elves", OracleID: "68954295-54e3-4303-a6bc-fc4547a4e3a3", Identity: "G"},
		{Name: "Elvish Mystic", OracleID: "3f3b2c10-21f8-4e13-be83-4ef3fa36e123", Identity: "G"},
		{Name: "Fyndhorn Elves", OracleID: "df317532-7d36-40fd-938f-e972749c8792", Identity: "G"},
		{Name: "Birds of Paradise", OracleID: "d3a0b660-358c-41bd-9cd2-41fbf3491b1a", Identity: "G"},
		{Name: "Delighted Halfling", OracleID: "f9d3b046-0b95-4103-a630-4b3fb88bb60b", Identity: "G"},
		{Name: "Lotus Cobra", OracleID: "8ad91f64-ccab-4edc-bd54-b2ee9267d614", Identity: "G"},
		{Name: "Sakura-Tribe Elder", OracleID: "e3afc704-220f-498f-9eaa-0821b17dc24c", Identity: "G"},
		{Name: "Wood Elves", OracleID: "8973bd99-20f8-4867-90ef-50392147ee1b", Identity: "G"},
		{Name: "Ornithopter of Paradise", OracleID: "3a940dfa-a026-4969-8981-ef90cdfbe9ef"},
		{Name: "Palladium Myr", OracleID: "7b0767b8-b504-456e-93bd-218502f73b3d"},
		{Name: "Burnished Hart", OracleID: "893fed41-c144-433f-af88-bc7d419b7fb3"},
		{Name: "Solemn Simulacrum", OracleID: "00c0543c-2a1f-4425-8283-4062d74a1637"},
		{Name: "Filigree Familiar", OracleID: "b544f690-e4bf-4a5b-984d-9256518fd574"},

		// Value creatures.
		{Name: "Tireless Provisioner", OracleID: "ab8d5f5c-1976-4f77-8ed2-8d28ee666741", Identity: "G"},
		{Name: "Beast Whisperer", OracleID: "5da7eea8-bb9e-47ce-a554-8a1ee058bd7a", Identity: "G"},
		{Name: "Acidic Slime", OracleID: "21f45043-5419-4019-8b6c-e5294bd5f549", Identity: "G"},
		{Name: "Reclamation Sage", OracleID: "032ec6e2-6cc3-4a97-9cc7-3233f5e11904", Identity: "G"},
		{Name: "Ambush Viper", OracleID: "8957f7c2-040c-4048-9f21-efa7c97682b7", Identity: "G"},
		{Name: "Giant Spider", OracleID: "e740ce2f-2134-473c-afa1-1b6d2d1e38ef", Identity: "G"},
		{Name: "Colossal Dreadmaw", OracleID: "08c7db90-c0cf-4482-b7ee-bb033e5996d2", Identity: "G"},
		{Name: "Tarmogoyf", OracleID: "45900b2f-f6a9-4c42-9642-008f3c1cf6dd", Identity: "G"},

		// Top end. What the ramp is for.
		{Name: "Rampaging Baloths", OracleID: "2d3e6549-6cc6-434f-a189-ba3b55e64c34", Identity: "G"},
		{Name: "Avenger of Zendikar", OracleID: "4ba5b3f6-503b-43e6-b66e-4f8c55cffed7", Identity: "G"},
		{Name: "Craterhoof Behemoth", OracleID: "8c52bd39-0586-48ca-b263-17210cf9feb6", Identity: "G"},
		{Name: "Mulldrifter", OracleID: "24d0f5e7-0d9e-4b76-900e-a7274e80312d", Identity: "U"},
		{Name: "Peregrine Drake", OracleID: "0bd67481-6bd9-48d6-92bd-8933b5ea1eae", Identity: "U"},
		{Name: "Wurmcoil Engine", OracleID: "d1a60f44-7696-49ee-91fb-cab5b3102962"},

		// Land ramp. Every one of these is also a Tatyova trigger.
		{Name: "Rampant Growth", OracleID: "8539f295-5d58-4436-a73a-b9277c4c7795", Identity: "G"},
		{Name: "Nature's Lore", OracleID: "78826359-fe63-44ad-adc4-a17ffcd710e4", Identity: "G"},
		{Name: "Three Visits", OracleID: "1b882a0e-0ede-4d1a-bd1a-9b7cffbcde8e", Identity: "G"},
		{Name: "Farseek", OracleID: "495e52e6-4c2b-4574-9474-eadbdcc8b4ac", Identity: "G"},
		{Name: "Cultivate", OracleID: "8b755881-a72d-4e21-a369-d2924eb4585a", Identity: "G"},
		{Name: "Kodama's Reach", OracleID: "1593ea18-2f2f-4ab4-83fb-6ccc0bec8a90", Identity: "G"},
		{Name: "Skyshroud Claim", OracleID: "376c9d3f-21d3-4251-bb6a-026fa9e1b0e1", Identity: "G"},
		{Name: "Explosive Vegetation", OracleID: "a0abd957-01a0-4aa9-8fc8-0e6840d21606", Identity: "G"},
		{Name: "Harrow", OracleID: "705509e9-a034-4a5a-9c65-66f58748b8a2", Identity: "G"},
		{Name: "Wayfarer's Bauble", OracleID: "31f15274-301b-47c5-ba19-0ced04520878"},

		// Payoffs and draw engines.
		{Name: "Overrun", OracleID: "204f9afe-c20b-4933-b5cd-aa572784762a", Identity: "G"},
		{Name: "Garruk's Uprising", OracleID: "3127ae9b-a7a7-43ec-89d7-688f8445b33d", Identity: "G"},
		{Name: "Return of the Wildspeaker", OracleID: "2b76f9e9-cd28-4eaf-8674-215c34263f96", Identity: "G"},
		{Name: "Shamanic Revelation", OracleID: "d1d171de-1c6d-4fb9-817a-9c689c709f3d", Identity: "G"},
		{Name: "Beastmaster Ascension", OracleID: "11b5308d-5bc0-4782-875f-a28be36e665d", Identity: "G"},
		{Name: "Elemental Bond", OracleID: "d9a7e5a6-3e41-4fc6-987a-18fe1b9d67dd", Identity: "G"},
		{Name: "Guardian Project", OracleID: "4f9e07ae-6341-4b46-9f77-f17ab659d266", Identity: "G"},

		// Interaction.
		{Name: "Beast Within", OracleID: "7735eeba-693b-47e2-bd51-414379cf1016", Identity: "G"},
		{Name: "Krosan Grip", OracleID: "3e39224c-72ce-4ecc-aa17-12c071ea1f3e", Identity: "G"},
		{Name: "Pongify", OracleID: "05849bd6-8f38-4031-be2b-e2aa03beb8cc", Identity: "U"},
		{Name: "Rapid Hybridization", OracleID: "06692cd9-ac2f-4a32-8fd1-043ba3c0fe71", Identity: "U"},
		{Name: "Counterspell", OracleID: "cc187110-1148-4090-bbb8-e205694a39f5", Identity: "U"},
		{Name: "Swan Song", OracleID: "8ddfc283-c9b4-41a5-af88-cf0068e986cc", Identity: "U"},

		// Blue card flow.
		{Name: "Preordain", OracleID: "ac641490-ca14-48d7-8cc4-b69ce984befa", Identity: "U"},
		{Name: "Divination", OracleID: "273b339c-964b-4a18-8eb5-ceb8abcdfd9e", Identity: "U"},
		{Name: "Frantic Search", OracleID: "16e015b2-f8a3-4b1a-80be-58a8f5fb5e8c", Identity: "U"},
		{Name: "Pull from Tomorrow", OracleID: "b1a23235-3076-475c-a68a-db29cf2a9dba", Identity: "U"},
		{Name: "Stroke of Genius", OracleID: "0cc6d683-366f-4ae4-be60-20ad9621fdaf", Identity: "U"},
		{Name: "Rhystic Study", OracleID: "53236dd7-845a-444c-96d5-f41ed7325d8f", Identity: "U"},

		// Mana rocks.
		{Name: "Sol Ring", OracleID: "6ad8011d-3471-4369-9d68-b264cc027487"},
		{Name: "Arcane Signet", OracleID: "0bc7f093-bef0-4f1a-852c-4b75ebf54838"},
		{Name: "Simic Signet", OracleID: "44503105-3e13-408d-a44f-37d503c61d72", Identity: "UG"},
		{Name: "Talisman of Curiosity", OracleID: "8c34b089-aad1-476e-958a-3077bf1bbb51", Identity: "UG"},
		{Name: "Mind Stone", OracleID: "c97361b5-af16-4a7b-af85-a429dbaf4ad2"},

		// Lands. Heavier than usual because the commander turns a land
		// drop into a cantrip.
		{Name: "Breeding Pool", OracleID: "20283c4a-f1f0-42f0-bc08-6da87474426b", Identity: "UG"},
		{Name: "Tropical Island", OracleID: "74b7fe23-5d3a-4092-8d78-7c0eba8f6f73", Identity: "UG"},
		{Name: "Hinterland Harbor", OracleID: "fb5a3403-7f0b-406c-8c4f-d693be010ca6", Identity: "UG"},
		{Name: "Yavimaya Coast", OracleID: "40b36bc6-c185-4bda-99e7-0118953c2c97", Identity: "UG"},
		{Name: "Dreamroot Cascade", OracleID: "dd8538e6-cd5f-4a88-aff5-eb5e76ce8ddb", Identity: "UG"},
		{Name: "Rejuvenating Springs", OracleID: "d1620449-930a-4895-a143-fd2a0a3c8b17", Identity: "UG"},
		{Name: "Simic Growth Chamber", OracleID: "046f5783-cc7b-416a-8cf6-2bcef9c2cc1a", Identity: "UG"},
		{Name: "Temple of Mystery", OracleID: "7e26f0b7-20e6-46d5-8130-d98c14d6aa29", Identity: "UG"},
		{Name: "Command Tower", OracleID: "0895c9b7-ae7d-4bb3-af17-3b75deb50a25"},
		{Name: "Exotic Orchard", OracleID: "27b047e3-0d41-45e2-98e9-9391d7923a1e"},
		{Name: "Reflecting Pool", OracleID: "67f43ac6-2a58-4b53-b5d7-0330e2a252e2"},
		{Name: "City of Brass", OracleID: "f25351e3-539b-4bbc-b92d-6480acf4d722"},
		{Name: "Mana Confluence", OracleID: "d0ee5bdc-2b69-4b73-9a20-ffcc18783b29"},
		{Name: "Ancient Tomb", OracleID: "23467047-6dba-4498-b783-1ebc4f74b8c2"},
		{Name: "Path of Ancestry", OracleID: "b473e293-59e3-4e04-acf2-622604aeb25f"},
		{Name: "Misty Rainforest", OracleID: "09dd85aa-47bc-4713-a9b9-8b52ff2285ed"},
		{Name: "Verdant Catacombs", OracleID: "67d60b24-d429-4ded-90d9-06e49f28c396"},
		{Name: "Polluted Delta", OracleID: "ef86989d-ce80-4e55-aece-7d11710eeffa"},
		{Name: "Prismatic Vista", OracleID: "032b8a0d-491a-4a12-ab9f-689010054d5b"},
		{Name: "Evolving Wilds", OracleID: "a75445d3-1303-4bb5-89ad-26ea93fecd48"},
		{Name: "Terramorphic Expanse", OracleID: "1bd3e453-aa21-4ee6-95c2-d6d920ee8e7a"},
		{Name: "Fabled Passage", OracleID: "0c85b8f7-0bd0-4680-9ec5-d4b110460a54"},
		{Name: "Myriad Landscape", OracleID: "2549bc57-9ffb-4053-9f10-f2a5f792b845"},
		{Name: "Strip Mine", OracleID: "d21a89eb-7c5b-459a-acc7-12b20b13bf79"},
		{Name: "Gaea's Cradle", OracleID: "7c427c3d-ecd8-45ef-bebd-8f10f4a311db", Identity: "G"},
		{Name: "Field of the Dead", OracleID: "aa959340-c869-4caa-92c7-572bd8d23eef"},
		{Name: "Forest", Count: 7, Basic: true},
		{Name: "Island", Count: 5, Basic: true},
	},
}
