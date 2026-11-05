package decks

// Growing Pains — +1/+1 counters. Grumgully puts an extra +1/+1
// counter on every other non-Human creature that enters, and the
// counter doublers (Hardened Scales, Branching Evolution, Doubling
// Season, Kami of Whispered Hopes) add more on top, so the deck's
// Hydras and X creatures enter well above their printed size and keep
// growing.
//
// Added by #554 for a precon-style theme the picker lacked, and the
// first Gruul deck. Stampede (mono-green-stompy) wins by casting big
// creatures early; this one wins by making small ones big. No creature
// in it is a Human, which Grumgully skips, apart from Taurean Mauler:
// a changeling is every creature type, so it grows on its own instead.
//
// Every non-basic card below is graded full, and none is on
// full_test.go's minor-caveats list. Left out on purpose:
//
//   - Heroic Intervention — full, but it declares no Purpose (ADR 0126
//     §6) and no Answers (ADR 0142). Lightning Greaves and Swiftfoot
//     Boots protect the board instead.
//   - Ram Through, Rabid Bite — not in the catalog.

var gruulCounters = Deck{
	ID:        "gruul-counters",
	Name:      "Growing Pains",
	Archetype: "counters",
	Identity:  "RG",
	Summary:   "Every creature enters with extra +1/+1 counters: Grumgully adds one, the doublers add more, and the Hydras grow out of reach.",
	Commander: Card{Name: "Grumgully, the Generous", OracleID: "fa4afcf7-9cf7-4f67-95b9-1f7fe66bf332", Identity: "RG"},
	Mainboard: []Card{
		// Counter engines. Grumgully adds one to every creature that enters.
		{Name: "Hardened Scales", OracleID: "a1f3da21-af6d-450e-bf0b-985d158418e6", Identity: "G"},
		{Name: "Branching Evolution", OracleID: "28fe909b-06e0-424c-9f75-c824a25f5865", Identity: "G"},
		{Name: "Doubling Season", OracleID: "01546b7d-a233-4176-8843-d732074dc5b6", Identity: "G"},
		{Name: "Ozolith, the Shattered Spire", OracleID: "41fed659-237c-4d8e-ad31-d17fa0d3f764", Identity: "G"},
		{Name: "Kami of Whispered Hopes", OracleID: "f122624f-f30d-444e-a62a-939829241045", Identity: "G"},
		{Name: "Rhythm of the Wild", OracleID: "acfa77fd-3610-4f12-9c3c-bd860ce91700", Identity: "RG"},
		{Name: "Tribute to the World Tree", OracleID: "72deedab-7c17-4505-aeca-4bc8596d80a5", Identity: "G"},
		{Name: "Elemental Bond", OracleID: "d9a7e5a6-3e41-4fc6-987a-18fe1b9d67dd", Identity: "G"},
		{Name: "Armorcraft Judge", OracleID: "d7f49243-a96e-499f-b2d7-8e9842432420", Identity: "G"},
		{Name: "Bitter Work", OracleID: "e9a24ed8-e844-4c50-b386-bc28da9989b1", Identity: "RG"},
		{Name: "Garruk's Uprising", OracleID: "3127ae9b-a7a7-43ec-89d7-688f8445b33d", Identity: "G"},
		{Name: "The Great Henge", OracleID: "78427103-9543-41fb-b6d4-72963fe87275", Identity: "G"},

		// Creatures that grow.
		{Name: "Gyre Sage", OracleID: "e3aa16cf-079b-4737-9ffd-7bfdffef0cb2", Identity: "G"},
		{Name: "Incubation Druid", OracleID: "34428f42-03ac-4795-8286-6cbea796df2b", Identity: "G"},
		{Name: "Duskshell Crawler", OracleID: "e54ad5a1-e79d-42db-a3e9-5caecae62c9f", Identity: "G"},
		{Name: "Scurry Oak", OracleID: "eee6c02b-7d6a-4445-ad27-8b03176147d4", Identity: "G"},
		{Name: "Fertilid", OracleID: "21f1c6d7-8289-44b2-b88f-c09e202be200", Identity: "G"},
		{Name: "Evolution Witness", OracleID: "0e07f1af-5ff6-4da5-8683-aea4cc390975", Identity: "G"},
		{Name: "Mossborn Hydra", OracleID: "84122d81-9634-4d3f-85a5-f6cf4303691a", Identity: "G"},
		{Name: "Managorger Hydra", OracleID: "b3f2265b-dd65-4b74-8b74-35ee0b147617", Identity: "G"},
		{Name: "Pridemalkin", OracleID: "f9672b63-415a-448b-a3da-140df63a0f0c", Identity: "G"},
		{Name: "Taurean Mauler", OracleID: "fd5ee26c-cc54-4dc0-b611-5cda14354a5e", Identity: "R"},
		{Name: "Ivy Lane Denizen", OracleID: "a5da5ad6-4ed2-4041-a983-76a8c87fa109", Identity: "G"},
		{Name: "Forgotten Ancient", OracleID: "2ef75cf3-30dc-4852-a6d1-1d6391fad022", Identity: "G"},
		{Name: "Bristly Bill, Spine Sower", OracleID: "d3b2d8a2-d3bc-448c-9cf6-6bead6010c28", Identity: "G"},
		{Name: "Kalonian Hydra", OracleID: "7bd36106-04fe-481f-b16e-e076dcbb183b", Identity: "G"},
		{Name: "Ouroboroid", OracleID: "50d6fd91-23d3-4d32-804f-6233e4386904", Identity: "G"},
		{Name: "Agonasaur Rex", OracleID: "1ad5766b-9ae3-437b-bd91-c4c988a8b095", Identity: "G"},
		{Name: "Herd Baloth", OracleID: "03873314-6e64-43ab-95c0-3d8692a57a03", Identity: "G"},
		{Name: "Mycoloth", OracleID: "d7fd16ce-282d-49cd-b2b4-0d25935e7a72", Identity: "G"},
		{Name: "Hydra Broodmaster", OracleID: "9e4fb4b1-9b28-4a49-b4ac-d5549c08594d", Identity: "G"},
		{Name: "Avenger of Zendikar", OracleID: "4ba5b3f6-503b-43e6-b66e-4f8c55cffed7", Identity: "G"},
		{Name: "Vigor", OracleID: "70787d8b-0a40-43ea-837a-9617dc13e7e5", Identity: "G"},
		{Name: "Goldvein Hydra", OracleID: "2b62543f-a475-457a-a96b-b5d070383d3c", Identity: "G"},
		{Name: "Primordial Hydra", OracleID: "1c36ed3a-c806-47e5-83f9-e44999c67fe5", Identity: "G"},
		{Name: "Voracious Hydra", OracleID: "ff8f5a4b-112a-425e-b489-7ee26d1d9fb3", Identity: "G"},
		{Name: "Benevolent Hydra", OracleID: "01dbf1bc-ca62-4fb6-959c-ef7c0dc03bb0", Identity: "G"},
		{Name: "Shivan Devastator", OracleID: "b7daa74c-6142-4107-9355-be98af6ccf13", Identity: "R"},
		{Name: "Walking Ballista", OracleID: "4b515bb0-f275-4400-8032-3173b799ab40"},
		{Name: "Hangarback Walker", OracleID: "dde55256-5259-44e7-a267-fca45a7f0d04"},
		{Name: "Stonecoil Serpent", OracleID: "78124cd1-2658-43f3-a339-0755d84154d2"},

		// Ramp.
		{Name: "Llanowar Elves", OracleID: "68954295-54e3-4303-a6bc-fc4547a4e3a3", Identity: "G"},
		{Name: "Elvish Mystic", OracleID: "3f3b2c10-21f8-4e13-be83-4ef3fa36e123", Identity: "G"},
		{Name: "Birds of Paradise", OracleID: "d3a0b660-358c-41bd-9cd2-41fbf3491b1a", Identity: "G"},
		{Name: "Sol Ring", OracleID: "6ad8011d-3471-4369-9d68-b264cc027487"},
		{Name: "Arcane Signet", OracleID: "0bc7f093-bef0-4f1a-852c-4b75ebf54838"},
		{Name: "Gruul Signet", OracleID: "d36e0c9f-c025-4dfe-9644-9cad2461ce38", Identity: "RG"},
		{Name: "Talisman of Impulse", OracleID: "f2ccc9e8-8e92-4f8c-8728-8c748630e0dd", Identity: "RG"},
		{Name: "Rampant Growth", OracleID: "8539f295-5d58-4436-a73a-b9277c4c7795", Identity: "G"},
		{Name: "Nature's Lore", OracleID: "78826359-fe63-44ad-adc4-a17ffcd710e4", Identity: "G"},
		{Name: "Three Visits", OracleID: "1b882a0e-0ede-4d1a-bd1a-9b7cffbcde8e", Identity: "G"},
		{Name: "Farseek", OracleID: "495e52e6-4c2b-4574-9474-eadbdcc8b4ac", Identity: "G"},
		{Name: "Kodama's Reach", OracleID: "1593ea18-2f2f-4ab4-83fb-6ccc0bec8a90", Identity: "G"},

		// Removal and protection.
		{Name: "Beast Within", OracleID: "7735eeba-693b-47e2-bd51-414379cf1016", Identity: "G"},
		{Name: "Chaos Warp", OracleID: "07a0cba9-8768-4fd9-a3d5-b0f83b4bf8e8", Identity: "R"},
		{Name: "Abrade", OracleID: "f9db72dc-9a5b-48a4-a86e-7464d9a2166a", Identity: "R"},
		{Name: "Lightning Bolt", OracleID: "4457ed35-7c10-48c8-9776-456485fdf070", Identity: "R"},
		{Name: "Vandalblast", OracleID: "3567c3c8-b3c7-45b7-935b-b1fdbc973720", Identity: "R"},
		{Name: "Bite Down", OracleID: "623903de-3c04-4745-9af3-d7ec9fb2574d", Identity: "G"},
		{Name: "Primal Might", OracleID: "58d7b8b3-ac47-4805-8259-4ba8e70d7dc3", Identity: "G"},
		{Name: "Lightning Greaves", OracleID: "ca204b66-8d0c-431a-8d34-282f7c2d17da"},
		{Name: "Swiftfoot Boots", OracleID: "c8b143ad-43ec-4e0d-a440-e348daa31391"},

		// Card draw.
		{Name: "Harmonize", OracleID: "7eff84f1-f772-497a-b350-bbc93d0230f7", Identity: "G"},
		{Name: "Shamanic Revelation", OracleID: "d1d171de-1c6d-4fb9-817a-9c689c709f3d", Identity: "G"},
		{Name: "Return of the Wildspeaker", OracleID: "2b76f9e9-cd28-4eaf-8674-215c34263f96", Identity: "G"},

		// Lands.
		{Name: "Command Tower", OracleID: "0895c9b7-ae7d-4bb3-af17-3b75deb50a25"},
		{Name: "Exotic Orchard", OracleID: "27b047e3-0d41-45e2-98e9-9391d7923a1e"},
		{Name: "Stomping Ground", OracleID: "16052b52-ade1-406f-a06b-ce7ea607fb63", Identity: "RG"},
		{Name: "Cinder Glade", OracleID: "dfac0258-e148-4d7d-8ded-fc2466d9caa6", Identity: "RG"},
		{Name: "Rootbound Crag", OracleID: "9516c4c1-d72d-434f-97e1-6a862434a169", Identity: "RG"},
		{Name: "Sheltered Thicket", OracleID: "db8d8643-3d0b-4f20-bf53-f4cd26a0e8df", Identity: "RG"},
		{Name: "Taiga", OracleID: "22e3cf1d-3559-4ce1-954c-8dc815342979", Identity: "RG"},
		{Name: "Karplusan Forest", OracleID: "bd912666-f37f-4767-af6f-9e6d0fcccacf", Identity: "RG"},
		{Name: "Game Trail", OracleID: "00de57d2-7cb6-4337-9bc6-f6711e4dfabf", Identity: "RG"},
		{Name: "Gruul Turf", OracleID: "657243dd-e479-4f4b-99d2-09b55d833a35", Identity: "RG"},
		{Name: "Rugged Highlands", OracleID: "6c922206-6e68-4dcd-9559-88da1074f2c4", Identity: "RG"},
		{Name: "Kessig Wolf Run", OracleID: "c6911265-54ef-4c16-bcf2-1ffb24b7d426", Identity: "RG"},
		{Name: "Oran-Rief, the Vastwood", OracleID: "e88027a6-24cc-4a8b-86db-734f26149ea8", Identity: "G"},
		{Name: "Raging Ravine", OracleID: "8d38194e-b607-4ff4-9c19-0e8636d463bf", Identity: "RG"},
		{Name: "Forest", Count: 13, Basic: true},
		{Name: "Mountain", Count: 7, Basic: true},
	},
}
