package effects

// conjurable_cards.go — the printed records conjure builds its cards from
// (conjure.go, ADR 0099 amendment 2026-11-24, #5948).
//
// One trimmed Scryfall record per card some catalog card conjures, in
// the dump's own JSON shape, so each becomes a game.Card through the deck
// importer (deck.ToGameCard): a conjured card is byte-for-byte the object
// a decklist would have seated, with the same oracle ID (so the catalog
// runs it) and a real printing ID (so the client shows its art).
// TestConjurableRecordsMatchTheDump (cards/coverage) holds every field
// here to the dump.
//
// Add a record when a new card conjures a name not listed; the name must
// also be a catalog card, or the conjured card would do nothing. The same
// records are the Oracle card reference a copy of a card defined by name
// is built from (CR 707.13, Garth One-Eye, named_card_copy.go).

const conjurableRecordsJSON = `[
	{
		"id": "5d69f9e2-9673-42bc-b2b5-d1c271212cc7",
		"oracle_id": "550c74d4-1fcb-406a-b02a-639a760a4380",
		"name": "Ancestral Recall",
		"layout": "normal",
		"type_line": "Instant",
		"mana_cost": "{U}",
		"cmc": 1.0,
		"colors": [
			"U"
		],
		"color_identity": [
			"U"
		],
		"oracle_text": "Target player draws three cards."
	},
	{
		"id": "f5d24a5b-c950-4fd9-99e6-a4b979d915b3",
		"oracle_id": "5089ec1a-f881-4d55-af14-5d996171203b",
		"name": "Black Lotus",
		"layout": "normal",
		"type_line": "Artifact",
		"mana_cost": "{0}",
		"cmc": 0.0,
		"colors": [],
		"produced_mana": [
			"B",
			"G",
			"R",
			"U",
			"W"
		],
		"oracle_text": "{T}, Sacrifice this artifact: Add three mana of any one color."
	},
	{
		"id": "45b6d1f8-a470-4471-bd4f-bed31cb4d344",
		"oracle_id": "310f141c-7f37-4729-aed6-dd9c09db448d",
		"name": "Blood Artist",
		"layout": "normal",
		"type_line": "Creature — Vampire",
		"mana_cost": "{1}{B}",
		"cmc": 2.0,
		"colors": [
			"B"
		],
		"color_identity": [
			"B"
		],
		"oracle_text": "Whenever this creature or another creature dies, target player loses 1 life and you gain 1 life.",
		"power": "0",
		"toughness": "1"
	},
	{
		"id": "90c99876-b325-432d-ae02-cd34320801fc",
		"oracle_id": "9908e597-9470-4c13-8387-39431b380138",
		"name": "Braingeyser",
		"layout": "normal",
		"type_line": "Sorcery",
		"mana_cost": "{X}{U}{U}",
		"cmc": 2.0,
		"colors": [
			"U"
		],
		"color_identity": [
			"U"
		],
		"oracle_text": "Target player draws X cards."
	},
	{
		"id": "f0b80608-1455-4899-9d19-1b1d2d068744",
		"oracle_id": "4a338863-d599-46e7-9c30-e11b898ae1b0",
		"name": "Chitterspitter",
		"layout": "normal",
		"type_line": "Artifact",
		"mana_cost": "{2}{G}",
		"cmc": 3.0,
		"colors": [
			"G"
		],
		"color_identity": [
			"G"
		],
		"oracle_text": "At the beginning of your upkeep, you may sacrifice a token. If you do, put an acorn counter on this artifact.\nSquirrels you control get +1/+1 for each acorn counter on this artifact.\n{G}, {T}: Create a 1/1 green Squirrel creature token."
	},
	{
		"id": "a31ca907-a274-436e-8433-7f11b0303f75",
		"oracle_id": "25895c28-23c1-439e-858b-5cb31fdc31de",
		"name": "Death of a Thousand Stings",
		"layout": "normal",
		"type_line": "Instant — Arcane",
		"mana_cost": "{4}{B}",
		"cmc": 5.0,
		"colors": [
			"B"
		],
		"color_identity": [
			"B"
		],
		"oracle_text": "Target player loses 1 life and you gain 1 life.\nAt the beginning of your upkeep, if you have more cards in hand than each opponent, you may return this card from your graveyard to your hand."
	},
	{
		"id": "d95ce840-b583-4c48-8a58-b85a4e2b9b2f",
		"oracle_id": "a7e97fa9-4b72-4548-b854-5be5f18a6f1a",
		"name": "Disenchant",
		"layout": "normal",
		"type_line": "Instant",
		"mana_cost": "{1}{W}",
		"cmc": 2.0,
		"colors": [
			"W"
		],
		"color_identity": [
			"W"
		],
		"oracle_text": "Destroy target artifact or enchantment."
	},
	{
		"id": "bd7a7363-03ba-466e-9d61-2dc5ecdcc630",
		"oracle_id": "464749e1-9df2-47d0-8fcd-516f566022b5",
		"name": "Lethal Sting",
		"layout": "normal",
		"type_line": "Sorcery",
		"mana_cost": "{2}{B}",
		"cmc": 3.0,
		"colors": [
			"B"
		],
		"color_identity": [
			"B"
		],
		"oracle_text": "As an additional cost to cast this spell, put a -1/-1 counter on a creature you control.\nDestroy target creature."
	},
	{
		"id": "35187d28-3d8c-4120-8421-f5b1cb7c76b5",
		"oracle_id": "4457ed35-7c10-48c8-9776-456485fdf070",
		"name": "Lightning Bolt",
		"layout": "normal",
		"type_line": "Instant",
		"mana_cost": "{R}",
		"cmc": 1.0,
		"colors": [
			"R"
		],
		"color_identity": [
			"R"
		],
		"oracle_text": "Lightning Bolt deals 3 damage to any target."
	},
	{
		"id": "abb424b6-741e-40ef-937f-e01536cf7479",
		"oracle_id": "c68faebc-b2cd-461b-b93e-e1fcd4816810",
		"name": "Midnight Clock",
		"layout": "normal",
		"type_line": "Artifact",
		"mana_cost": "{2}{U}",
		"cmc": 3.0,
		"colors": [
			"U"
		],
		"color_identity": [
			"U"
		],
		"produced_mana": [
			"U"
		],
		"oracle_text": "{T}: Add {U}.\n{2}{U}: Put an hour counter on this artifact.\nAt the beginning of each upkeep, put an hour counter on this artifact.\nWhen the twelfth hour counter is put on this artifact, shuffle your hand and graveyard into your library, then draw seven cards. Exile this artifact."
	},
	{
		"id": "2e8455b3-a5e8-438f-aacb-79d2719636e3",
		"oracle_id": "376ee366-e082-402f-b4db-6592fcfcacd2",
		"name": "Mox Emerald",
		"layout": "normal",
		"type_line": "Artifact",
		"mana_cost": "{0}",
		"cmc": 0.0,
		"colors": [],
		"color_identity": [
			"G"
		],
		"produced_mana": [
			"G"
		],
		"oracle_text": "{T}: Add {G}."
	},
	{
		"id": "0015d273-5f49-4154-9bd1-d3aef62d1c73",
		"oracle_id": "0677f49e-f8bf-4349-af52-2ccde9287c2e",
		"name": "Mox Jet",
		"layout": "normal",
		"type_line": "Artifact",
		"mana_cost": "{0}",
		"cmc": 0.0,
		"colors": [],
		"color_identity": [
			"B"
		],
		"produced_mana": [
			"B"
		],
		"oracle_text": "{T}: Add {B}."
	},
	{
		"id": "b3020358-9c59-45fb-bf74-9c69bc47580b",
		"oracle_id": "824597b8-c89a-47ec-8526-7efc6e24ef0e",
		"name": "Mox Pearl",
		"layout": "normal",
		"type_line": "Artifact",
		"mana_cost": "{0}",
		"cmc": 0.0,
		"colors": [],
		"color_identity": [
			"W"
		],
		"produced_mana": [
			"W"
		],
		"oracle_text": "{T}: Add {W}."
	},
	{
		"id": "1abdfbaa-5bd7-4732-b62b-ab27581bc6bb",
		"oracle_id": "ed85fa82-e4fa-434b-92a8-36b6075708d1",
		"name": "Mox Ruby",
		"layout": "normal",
		"type_line": "Artifact",
		"mana_cost": "{0}",
		"cmc": 0.0,
		"colors": [],
		"color_identity": [
			"R"
		],
		"produced_mana": [
			"R"
		],
		"oracle_text": "{T}: Add {R}."
	},
	{
		"id": "b1bd8899-c5dc-40a7-a417-0a27e81f4458",
		"oracle_id": "d5ed1233-df87-4b90-8918-13922ec95249",
		"name": "Mox Sapphire",
		"layout": "normal",
		"type_line": "Artifact",
		"mana_cost": "{0}",
		"cmc": 0.0,
		"colors": [],
		"color_identity": [
			"U"
		],
		"produced_mana": [
			"U"
		],
		"oracle_text": "{T}: Add {U}."
	},
	{
		"id": "ace2d2c9-6af0-4728-9265-36730df50e83",
		"oracle_id": "e6e4a8bd-5c40-4654-8de1-0da9afed90fd",
		"name": "Regrowth",
		"layout": "normal",
		"type_line": "Sorcery",
		"mana_cost": "{1}{G}",
		"cmc": 2.0,
		"colors": [
			"G"
		],
		"color_identity": [
			"G"
		],
		"oracle_text": "Return target card from your graveyard to your hand."
	},
	{
		"id": "760d746c-fe1b-4a75-8c71-4fe9966ec9ea",
		"oracle_id": "711eea87-0fa3-46e0-a42b-fa5a86455f04",
		"name": "Shivan Dragon",
		"layout": "normal",
		"type_line": "Creature — Dragon",
		"mana_cost": "{4}{R}{R}",
		"cmc": 6.0,
		"colors": [
			"R"
		],
		"color_identity": [
			"R"
		],
		"keywords": [
			"Flying"
		],
		"oracle_text": "Flying\n{R}: This creature gets +1/+0 until end of turn.",
		"power": "5",
		"toughness": "5"
	},
	{
		"id": "a396551e-882d-48ef-807a-81e5f2dfe517",
		"oracle_id": "a5550795-8cd3-4f29-a9df-02c48b124dbc",
		"name": "Sigardian Evangel",
		"layout": "normal",
		"type_line": "Creature — Human Cleric",
		"mana_cost": "{1}{W}",
		"cmc": 2.0,
		"colors": [
			"W"
		],
		"color_identity": [
			"W"
		],
		"keywords": [
			"Conjure"
		],
		"oracle_text": "When Sigardian Evangel enters the battlefield, conjure a card named Sigardian Evangel into your hand. Discard that card at the beginning of the next end step.\nWhen Sigardian Evangel enters the battlefield, tap target permanent you don't control.",
		"power": "3",
		"toughness": "1"
	},
	{
		"id": "2e5aa787-0d7f-4347-9276-c66be3d8ed3b",
		"oracle_id": "7fd7d20d-d44c-443c-a2cf-320f39a3bec4",
		"name": "Stinging Shot",
		"layout": "normal",
		"type_line": "Instant",
		"mana_cost": "{G}",
		"cmc": 1.0,
		"colors": [
			"G"
		],
		"color_identity": [
			"G"
		],
		"keywords": [
			"Cycling"
		],
		"oracle_text": "Put three -1/-1 counters on target creature with flying.\nCycling {2} ({2}, Discard this card: Draw a card.)"
	},
	{
		"id": "f8c2f285-141b-479b-aca6-ceacf60e3f6a",
		"oracle_id": "b81f041d-98db-4408-9472-c483e4a502bc",
		"name": "Terror",
		"layout": "normal",
		"type_line": "Instant",
		"mana_cost": "{1}{B}",
		"cmc": 2.0,
		"colors": [
			"B"
		],
		"color_identity": [
			"B"
		],
		"oracle_text": "Destroy target nonartifact, nonblack creature. It can't be regenerated."
	},
	{
		"id": "8e9ab027-120f-4bf3-a281-51d7cbb23604",
		"oracle_id": "d0209d3f-3f7e-4fd5-bce5-10bce6f29c86",
		"name": "Time Walk",
		"layout": "normal",
		"type_line": "Sorcery",
		"mana_cost": "{1}{U}",
		"cmc": 2.0,
		"colors": [
			"U"
		],
		"color_identity": [
			"U"
		],
		"oracle_text": "Take an extra turn after this one."
	},
	{
		"id": "d3c0c1da-e14f-4ce5-a639-c9bb8af0baab",
		"oracle_id": "c823e687-6311-4c99-974b-fd77d204141a",
		"name": "Timetwister",
		"layout": "normal",
		"type_line": "Sorcery",
		"mana_cost": "{2}{U}",
		"cmc": 3.0,
		"colors": [
			"U"
		],
		"color_identity": [
			"U"
		],
		"oracle_text": "Each player shuffles their hand and graveyard into their library, then draws seven cards. (Then put Timetwister into its owner's graveyard.)"
	}
]`
