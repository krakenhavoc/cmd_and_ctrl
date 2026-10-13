package decks

// Armed and Dangerous — voltron. One big flier, dressed. Aurelia, the
// Law Above has flying, vigilance and haste, so whatever is attached to
// her attacks the turn it lands. Twenty-one Equipment and four Auras
// (Swords, Batterskull, Embercleave, Lost Jitte, Shadowspear, Hammer of
// Nazahn, Kaldra Compleat and company) turn that into a clock, and
// Lightning Greaves, Swiftfoot Boots, Mask of Avacyn and Whispersilk
// Cloak, with Mother of Runes, Teferi's Protection and Flawless
// Maneuver, are there to keep her on the table.
//
// This is the Voltron deck the package comment used to say the catalog
// could not fill (#554). The attachment relation shipped in #374 and
// the catalog has since grown well past the fifteen to twenty-five
// attachments such a deck needs, so the count problem is solved.
//
// Not a combo deck, and not one that leans on extra combats: it wins
// by hitting someone with a very large creature. The creature half of
// the list is the Equipment payoffs (Sram, both Danithas, Puresteel
// Paladin, Cloud, Reckless Fireweaver, Lion Sash), cheap bodies to carry
// the gear if Aurelia is answered (Selfless Spirit, Inspiring Overseer,
// Knight of the White Orchid, Thalia), and a few fliers to take over
// (Goldspan Dragon, Thunderbreak Regent, Lyra Dawnbringer, Angel of
// Invention, Archangel of Thune).
//
// Every non-basic card below is in the effect catalog and graded full,
// and none is on full_test.go's minor-caveats list. Stoneforge Mystic
// carries a caveat on the clause this deck would play it for (it cannot
// put an Equipment onto the battlefield), and Wayfarer's Bauble has not
// been graded, so neither is here.

var borosVoltron = Deck{
	ID:        "boros-voltron",
	Name:      "Armed and Dangerous",
	Archetype: "voltron",
	Identity:  "WR",
	Summary:   "Equipment and Auras on one big flier: dress Aurelia in Swords and armour, keep her alive with protection, and win by hitting one player hard.",
	Commander: Card{Name: "Aurelia, the Law Above", OracleID: "2a800427-ff8c-4b3c-baee-85211b70656d", Identity: "WR"},
	Mainboard: []Card{
		// Equipment. Twenty-one of them, so the deck always has something to attach.
		{Name: "Bonesplitter", OracleID: "452e3f5f-ce17-4682-966b-5cc100210aee"},
		{Name: "Lost Jitte", OracleID: "a9d72f78-2ab5-4e2e-ab7b-ef875e0a0609"},
		{Name: "Shadowspear", OracleID: "8b27326f-e7b8-4a4d-b589-df459246d19a"},
		{Name: "Skullclamp", OracleID: "65986c1b-8e51-4604-b685-d82fa7d1263a"},
		{Name: "Sword of Fire and Ice", OracleID: "2ccdc60a-49a9-44b9-a7af-0ebf18b26785"},
		{Name: "Sword of Feast and Famine", OracleID: "d0901053-6de0-46d0-9ee3-8d40510236c1"},
		{Name: "Sword of Hearth and Home", OracleID: "913e6182-706a-4872-8c8a-e146b0ae0738"},
		{Name: "Sword of Body and Mind", OracleID: "fac42229-4f5f-4d04-85dd-5031d4e435aa"},
		{Name: "Sword of Vengeance", OracleID: "e366afb3-c447-4bde-b358-41c8568142d5"},
		{Name: "Batterskull", OracleID: "d12e5ce0-5705-4c35-9a93-b883db52c80c"},
		{Name: "Hammer of Nazahn", OracleID: "e3955573-3db5-490e-903e-65e0172a9202"},
		{Name: "Loxodon Warhammer", OracleID: "dba35ac5-7ad3-488a-a006-6b9a1d54eea5"},
		{Name: "Whispersilk Cloak", OracleID: "9ad4f730-a18e-4a7c-a468-a926c718c741"},
		{Name: "Darksteel Plate", OracleID: "b5b4cf54-ed5e-42d0-9d98-5fec76b0b0b8"},
		{Name: "Mask of Avacyn", OracleID: "ab66f8a8-eb3d-4c2d-95e9-26a53c66b237"},
		{Name: "Hero's Blade", OracleID: "e6bcd25d-39b6-4619-8a27-9f3d87f4a17b"},
		{Name: "Embercleave", OracleID: "4d6120d6-fcce-40bc-9fc6-e1f5beb6c728", Identity: "R"},
		{Name: "Lightning Greaves", OracleID: "ca204b66-8d0c-431a-8d34-282f7c2d17da"},
		{Name: "Swiftfoot Boots", OracleID: "c8b143ad-43ec-4e0d-a440-e348daa31391"},
		{Name: "Kaldra Compleat", OracleID: "7359e82b-db79-488d-a1d4-75a00f12a4cf"},
		{Name: "Cori-Steel Cutter", OracleID: "46b1a169-bc32-47bf-b379-fa0c3d13878f", Identity: "R"},

		// Auras.
		{Name: "Spirit Mantle", OracleID: "81a51328-b995-4f3b-90bc-a20ae21dd254", Identity: "W"},
		{Name: "All That Glitters", OracleID: "a4d751e0-41c1-4e90-853d-512f385acd81", Identity: "W"},
		{Name: "Ethereal Armor", OracleID: "dbba75f5-2404-4bd5-982b-f6c4effa5316", Identity: "W"},
		{Name: "Battle Mastery", OracleID: "a06c7c1a-8534-40fb-8bb8-59f8f2530567", Identity: "W"},

		// Creatures that carry the gear or pay for it.
		{Name: "Sram, Senior Edificer", OracleID: "7e00b0cd-d212-4604-ba07-da21f4fe00b0", Identity: "W"},
		{Name: "Danitha Capashen, Paragon", OracleID: "4b6377da-83e7-4519-9582-16a9c16b8faa", Identity: "W"},
		{Name: "Danitha, Sword of Hope", OracleID: "f7b754d3-9863-420d-895d-00e70b8b078a", Identity: "W"},
		{Name: "Puresteel Paladin", OracleID: "74a62c7b-4753-4af2-b7a1-9a4ae8988801", Identity: "W"},
		{Name: "Cloud, Midgar Mercenary", OracleID: "33d2584b-bf29-4c22-bd45-14ba2fb98c0e", Identity: "W"},
		{Name: "Lion Sash", OracleID: "50330e48-db74-4c5f-a0bc-f9a8607e8f31", Identity: "W"},
		{Name: "Rabbit Battery", OracleID: "c739e180-2f14-41ed-8e7e-50b7df985f35", Identity: "R"},
		{Name: "Reckless Fireweaver", OracleID: "180e1a7e-890d-477c-80a5-da8a5f2857b3", Identity: "R"},
		{Name: "Mother of Runes", OracleID: "60433b48-d27f-413c-905c-43839b1943f1", Identity: "W"},
		{Name: "Selfless Spirit", OracleID: "71d785a9-ddc8-472e-a778-a551b444a4bd", Identity: "W"},
		{Name: "Inspiring Overseer", OracleID: "d646e42b-5635-4798-b633-29c093b66a55", Identity: "W"},
		{Name: "Knight of the White Orchid", OracleID: "7d0efc64-151f-45b4-8c0b-c32609f9d862", Identity: "W"},
		{Name: "Thalia, Guardian of Thraben", OracleID: "9b7f1d05-707c-4ed3-9f0e-8ced1232c2ee", Identity: "W"},
		{Name: "Mirror Entity", OracleID: "17e905ca-c0bd-473d-95a7-e180ba5fea43", Identity: "W"},
		{Name: "Thunderbreak Regent", OracleID: "bb04f927-0348-4a14-9e74-381f18083f6a", Identity: "R"},
		{Name: "Goldspan Dragon", OracleID: "716b3ea2-45b7-4a8f-af72-de7f4e510eff", Identity: "R"},
		{Name: "Lyra Dawnbringer", OracleID: "592c91fc-6430-4c76-9460-65f047350f67", Identity: "W"},
		{Name: "Angel of Invention", OracleID: "28c7f2b6-ba67-4ccf-9e1b-99c89a9d1f72", Identity: "W"},
		{Name: "Archangel of Thune", OracleID: "4f2d4538-dc1d-4c09-964b-b0d7c240fb7d", Identity: "W"},

		// Protection and answers.
		{Name: "Swords to Plowshares", OracleID: "b1544f21-7e98-461b-aed5-e748b0168c52", Identity: "W"},
		{Name: "Path to Exile", OracleID: "d683d985-9888-4d21-8b5f-69e69ce4a03b", Identity: "W"},
		{Name: "Lightning Bolt", OracleID: "4457ed35-7c10-48c8-9776-456485fdf070", Identity: "R"},
		{Name: "Wild Slash", OracleID: "f76a0e5b-74c1-4ee8-b502-e8b988086de4", Identity: "R"},
		{Name: "Chaos Warp", OracleID: "07a0cba9-8768-4fd9-a3d5-b0f83b4bf8e8", Identity: "R"},
		{Name: "Council's Judgment", OracleID: "acf4d685-a718-42f7-9a7a-dee069ae8db2", Identity: "W"},
		{Name: "Oblivion Ring", OracleID: "bd9b9772-f5f9-4c6b-913e-7193bea5d0a7", Identity: "W"},
		{Name: "Fateful Absence", OracleID: "35bba442-1aec-4d33-b502-4c580d61644b", Identity: "W"},
		{Name: "Lightning Helix", OracleID: "800c258a-cfc4-4a54-a667-065ea8dea69e", Identity: "WR"},
		{Name: "Boros Charm", OracleID: "2679d0dd-ba30-4a1c-b6a0-b3ac6c790496", Identity: "WR"},
		{Name: "Akroma's Will", OracleID: "fd949f82-fc10-4e37-8aa9-6c7569fe3c55", Identity: "W"},
		{Name: "Flawless Maneuver", OracleID: "4e183439-17d2-47ff-9d99-5e22821d91e3", Identity: "W"},
		{Name: "Teferi's Protection", OracleID: "0d4ecdb1-ec90-497f-a7a4-1c68092b8757", Identity: "W"},
		{Name: "Generous Gift", OracleID: "fae37e28-e137-4177-b973-fa8b4dd8f409", Identity: "W"},
		{Name: "Smothering Tithe", OracleID: "153376c9-dffd-458c-8ce3-a4c8269bc4e9", Identity: "W"},

		// Mana.
		{Name: "Sol Ring", OracleID: "6ad8011d-3471-4369-9d68-b264cc027487"},
		{Name: "Arcane Signet", OracleID: "0bc7f093-bef0-4f1a-852c-4b75ebf54838"},
		{Name: "Boros Signet", OracleID: "41c84665-1f99-40ab-aaca-1188649eb263", Identity: "WR"},
		{Name: "Mind Stone", OracleID: "c97361b5-af16-4a7b-af85-a429dbaf4ad2"},
		{Name: "Fellwar Stone", OracleID: "95560508-7ac9-4be9-8a3f-3c7d5b52807b"},

		// Lands.
		{Name: "Command Tower", OracleID: "0895c9b7-ae7d-4bb3-af17-3b75deb50a25"},
		{Name: "Clifftop Retreat", OracleID: "d7faa3c8-46cf-46b2-bfa4-89000307cf18", Identity: "WR"},
		{Name: "Sacred Foundry", OracleID: "45181cb8-2090-4471-ba90-e5a8f04d525f", Identity: "WR"},
		{Name: "Battlefield Forge", OracleID: "6b75b94e-83b7-457e-ac41-7ca90b5a59aa", Identity: "WR"},
		{Name: "Inspiring Vantage", OracleID: "3f17c60e-923a-4392-9da8-87d9ded009b7", Identity: "WR"},
		{Name: "Rugged Prairie", OracleID: "8e7641e1-e814-4d5a-9cb3-71ad2f4ceee8", Identity: "WR"},
		{Name: "Boros Garrison", OracleID: "8fa3ac81-3dfe-4565-be99-5554f7597b4b", Identity: "WR"},
		{Name: "Rogue's Passage", OracleID: "f29dc596-2121-4421-8463-15f6c2e8b9b3"},
		{Name: "Plains", Count: 14, Basic: true},
		{Name: "Mountain", Count: 13, Basic: true},
	},
}
