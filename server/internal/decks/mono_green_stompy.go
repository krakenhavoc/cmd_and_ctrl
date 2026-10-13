package decks

// Stampede — stompy. Ramp, then land something enormous a turn or two
// before the table expects it. Goreclaw makes every creature spell with
// power 4 or greater cost {2} less, and whenever it attacks every
// creature with power 4 or greater gets +1/+1 and trample, so the
// deck's big creatures are cheaper and meaner than their printed
// numbers. Deep Roots (simic-ramp) wins by what its lands do; this
// deck wins by the size of what it casts.
//
// Added by #554 as the mono-coloured stompy option: mono-green on
// purpose, so there is no colour fixing to read, no dual lands to
// explain, and every mana creature taps for the one colour the whole
// deck uses. Fast mana is Sol Ring and Arcane Signet only, in line with
// the upgraded-precon rule above.
//
// Every non-basic card below is in the effect catalog and graded full,
// and none of them is on full_test.go's minor-caveats list: this deck
// was built to a stricter bar than the first four, and decks_test.go and
// full_test.go fail the build if that stops being true. Cultivate and
// Tireless Provisioner, which Deep Roots plays under a named caveat,
// were left out; Skyshroud Claim and Wood Elves were left out because
// the catalog has not graded them. Carnage Tyrant, Thrun and Primeval
// Titan, the staples a stompy list would want, are not in the catalog
// at all.

var monoGreenStompy = Deck{
	ID:        "mono-green-stompy",
	Name:      "Stampede",
	Archetype: "stompy",
	Identity:  "G",
	Summary:   "Ramp into oversized creatures a turn early: Goreclaw makes every big spell cheaper and every big attacker hit harder, and the table runs out of blockers.",
	Commander: Card{Name: "Goreclaw, Terror of Qal Sisma", OracleID: "befb211f-37ca-4083-98d4-9ff1f28be3f2", Identity: "G"},
	Mainboard: []Card{
		// Mana creatures.
		{Name: "Llanowar Elves", OracleID: "68954295-54e3-4303-a6bc-fc4547a4e3a3", Identity: "G"},
		{Name: "Elvish Mystic", OracleID: "3f3b2c10-21f8-4e13-be83-4ef3fa36e123", Identity: "G"},
		{Name: "Fyndhorn Elves", OracleID: "df317532-7d36-40fd-938f-e972749c8792", Identity: "G"},
		{Name: "Birds of Paradise", OracleID: "d3a0b660-358c-41bd-9cd2-41fbf3491b1a", Identity: "G"},
		{Name: "Arbor Elf", OracleID: "4567a528-75f0-4ea6-b927-3a500caf76ac", Identity: "G"},
		{Name: "Boreal Druid", OracleID: "2fcc69ff-8ab5-4e14-afe3-db892049a872", Identity: "G"},
		{Name: "Sakura-Tribe Elder", OracleID: "e3afc704-220f-498f-9eaa-0821b17dc24c", Identity: "G"},
		{Name: "Burnished Hart", OracleID: "893fed41-c144-433f-af88-bc7d419b7fb3"},

		// Ramp spells and rocks.
		{Name: "Sol Ring", OracleID: "6ad8011d-3471-4369-9d68-b264cc027487"},
		{Name: "Arcane Signet", OracleID: "0bc7f093-bef0-4f1a-852c-4b75ebf54838"},
		{Name: "Mind Stone", OracleID: "c97361b5-af16-4a7b-af85-a429dbaf4ad2"},
		{Name: "Thought Vessel", OracleID: "9965d9c5-2ebf-4a6c-930e-55c5890979be"},
		{Name: "Rampant Growth", OracleID: "8539f295-5d58-4436-a73a-b9277c4c7795", Identity: "G"},
		{Name: "Nature's Lore", OracleID: "78826359-fe63-44ad-adc4-a17ffcd710e4", Identity: "G"},
		{Name: "Three Visits", OracleID: "1b882a0e-0ede-4d1a-bd1a-9b7cffbcde8e", Identity: "G"},
		{Name: "Kodama's Reach", OracleID: "1593ea18-2f2f-4ab4-83fb-6ccc0bec8a90", Identity: "G"},
		{Name: "Explosive Vegetation", OracleID: "a0abd957-01a0-4aa9-8fc8-0e6840d21606", Identity: "G"},
		{Name: "Utopia Sprawl", OracleID: "00d8efa6-a2d9-4249-8da7-b45173675329", Identity: "G"},

		// Early and mid creatures.
		{Name: "Eternal Witness", OracleID: "30b24e8e-3b0e-4d8e-90f3-f66eb7c1858c", Identity: "G"},
		{Name: "Tireless Tracker", OracleID: "e6555cca-e608-4c87-b835-9cd47fd1f543", Identity: "G"},
		{Name: "Scavenging Ooze", OracleID: "1ff25f67-36a7-4cfa-a2b1-2135b5b6fb67", Identity: "G"},
		{Name: "Reclamation Sage", OracleID: "032ec6e2-6cc3-4a97-9cc7-3233f5e11904", Identity: "G"},
		{Name: "Courser of Kruphix", OracleID: "46779609-4fa7-4fd2-b5b4-7d4d749339e6", Identity: "G"},
		{Name: "Beast Whisperer", OracleID: "5da7eea8-bb9e-47ce-a554-8a1ee058bd7a", Identity: "G"},
		{Name: "Questing Beast", OracleID: "b685757b-521e-4353-a233-97052359723d", Identity: "G"},
		{Name: "Polukranos, World Eater", OracleID: "ecbf5560-853a-43a9-bf70-ac8d403b0ce8", Identity: "G"},
		{Name: "Bayou Groff", OracleID: "e83b2790-3cec-429c-a65d-6f4c7d025d37", Identity: "G"},
		{Name: "Thrashing Brontodon", OracleID: "60bc63dc-ac9f-4a2f-aef5-c90d0aa31553", Identity: "G"},
		{Name: "Prowling Serpopard", OracleID: "4bdfd718-f474-4223-88d8-7fa9fb0c86b4", Identity: "G"},
		{Name: "Wolfbriar Elemental", OracleID: "2f8872fe-84dc-4cda-a253-e7503a5c96a3", Identity: "G"},
		{Name: "Imperiosaur", OracleID: "e9ced5d8-8337-403f-86a3-bddb9c77d658", Identity: "G"},
		{Name: "Kalonian Hydra", OracleID: "7bd36106-04fe-481f-b16e-e076dcbb183b", Identity: "G"},
		{Name: "Bristling Hydra", OracleID: "b3b23c58-0b7a-4fe4-a8e8-5320a7605724", Identity: "G"},
		{Name: "Lurking Green Dragon", OracleID: "262e1cf6-61c8-444d-a71b-050eebcaf932", Identity: "G"},

		// Top end. Goreclaw makes every one of these cheaper.
		{Name: "Verdurous Gearhulk", OracleID: "3765e8bb-e70d-4503-b8dc-1e684e434c18", Identity: "G"},
		{Name: "Titania, Protector of Argoth", OracleID: "d0ade00d-a496-441d-9b7e-7dc033d3292c", Identity: "G"},
		{Name: "Ancient Greenwarden", OracleID: "3bcf090c-e890-4a9f-a8aa-6079e4ec9947", Identity: "G"},
		{Name: "Acidic Slime", OracleID: "21f45043-5419-4019-8b6c-e5294bd5f549", Identity: "G"},
		{Name: "Thunderfoot Baloth", OracleID: "f55334d9-1ec2-4667-bf28-5a0384d6f053", Identity: "G"},
		{Name: "Hydra Broodmaster", OracleID: "9e4fb4b1-9b28-4a49-b4ac-d5549c08594d", Identity: "G"},
		{Name: "Rampaging Baloths", OracleID: "2d3e6549-6cc6-434f-a189-ba3b55e64c34", Identity: "G"},
		{Name: "Regal Force", OracleID: "e264ffe3-0252-49d6-b990-dbb3654325a5", Identity: "G"},
		{Name: "Old Gnawbone", OracleID: "dff3f4c7-f792-4ca3-9b0a-0738e70664d9", Identity: "G"},
		{Name: "Avenger of Zendikar", OracleID: "4ba5b3f6-503b-43e6-b66e-4f8c55cffed7", Identity: "G"},
		{Name: "Hornet Queen", OracleID: "3b1f8108-6911-49e9-8f78-f950bb58cb6c", Identity: "G"},
		{Name: "Craterhoof Behemoth", OracleID: "8c52bd39-0586-48ca-b263-17210cf9feb6", Identity: "G"},
		{Name: "Terastodon", OracleID: "17975a20-2c13-4325-be1e-0bcda5063a2d", Identity: "G"},
		{Name: "Ghalta the Unstoppable", OracleID: "8e5ea773-a60b-4502-9db5-ce93fa61cc89", Identity: "G"},
		{Name: "Ghalta, Primal Hunger", OracleID: "b0b6be0c-41cf-4757-9f0e-87227b6ba6b3", Identity: "G"},
		{Name: "Wurmcoil Engine", OracleID: "d1a60f44-7696-49ee-91fb-cab5b3102962"},
		{Name: "Phyrexian Hydra", OracleID: "b16085d5-6d00-4d47-ab8b-d18d55c72141", Identity: "G"},
		{Name: "Annoyed Altisaur", OracleID: "a8135179-51ef-454e-98ff-69137440339f", Identity: "G"},
		{Name: "Autarch Mammoth", OracleID: "eb50b5f2-4840-4f90-9c85-40fc581be9c7", Identity: "G"},

		// Spells.
		{Name: "Rancor", OracleID: "9d2d6479-531c-4ce1-b52b-00e36fa63b64", Identity: "G"},
		{Name: "Overrun", OracleID: "204f9afe-c20b-4933-b5cd-aa572784762a", Identity: "G"},
		{Name: "Beast Within", OracleID: "7735eeba-693b-47e2-bd51-414379cf1016", Identity: "G"},
		{Name: "Naturalize", OracleID: "bdb3ca68-ec1f-4e16-81cc-d23f8f52c728", Identity: "G"},
		{Name: "Bite Down", OracleID: "623903de-3c04-4745-9af3-d7ec9fb2574d", Identity: "G"},
		{Name: "Garruk's Uprising", OracleID: "3127ae9b-a7a7-43ec-89d7-688f8445b33d", Identity: "G"},
		{Name: "Harmonize", OracleID: "7eff84f1-f772-497a-b350-bbc93d0230f7", Identity: "G"},
		{Name: "Shamanic Revelation", OracleID: "d1d171de-1c6d-4fb9-817a-9c689c709f3d", Identity: "G"},
		{Name: "Swiftfoot Boots", OracleID: "c8b143ad-43ec-4e0d-a440-e348daa31391"},
		{Name: "Wild Growth", OracleID: "706ae742-1807-44b7-a4fa-f2e26f61519a", Identity: "G"},
		{Name: "Sylvan Library", OracleID: "92eed395-62ca-4293-882b-8565c40daab5", Identity: "G"},

		// Lands.
		{Name: "Rogue's Passage", OracleID: "f29dc596-2121-4421-8463-15f6c2e8b9b3"},
		{Name: "Forest", Count: 34, Basic: true},
	},
}
