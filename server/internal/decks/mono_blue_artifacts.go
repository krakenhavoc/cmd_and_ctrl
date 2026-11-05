package decks

// Thopter Swarm — artifacts. Sai makes a 1/1 flying Thopter whenever
// you cast an artifact spell, and trades two artifacts for a card, so
// the deck is mostly cheap artifacts: zero- and one-mana bodies, mana
// rocks, and the payoffs that care how many artifacts are out. Each
// rock is a Thopter as well as a mana, and the Thopters are both the
// army and the card draw.
//
// Added by #554 as the mono-blue deck the picker lacked: one colour,
// nothing to fix, and a theme (cast artifacts, get fliers) a new player
// reads off the commander. It is also the picker's first artifacts
// deck. Fast mana is Sol Ring and Arcane Signet only, as in the other
// mono-coloured lists.
//
// Every non-basic card below is graded full, and none is on
// full_test.go's minor-caveats list. Left out on purpose:
//
//   - Metalwork Colossus — with Sai and Stridehangar Automaton out it
//     casts for free, makes two Thopters, and returns itself by
//     sacrificing them, forever. Combo stays out (ADR 0033).
//   - Thirst for Knowledge and Tezzeret's Gambit — graded full, but they
//     declare no Purpose (ADR 0126 §6), so the bot would price them as
//     blanks.
//   - Thopter Assembly, Retrofitter Foundry, Whirler Rogue — not in the
//     catalog.

var monoBlueArtifacts = Deck{
	ID:        "mono-blue-artifacts",
	Name:      "Thopter Swarm",
	Archetype: "artifacts",
	Identity:  "U",
	Summary:   "Cast cheap artifacts and Sai turns each one into a flying Thopter, then cashes spare Thopters in for cards until the skies are full.",
	Commander: Card{Name: "Sai, Master Thopterist", OracleID: "52241b9c-7a69-4176-9234-8bdab09d8e64", Identity: "U"},
	Mainboard: []Card{
		// Cheap artifacts: every one is a Thopter with Sai out.
		{Name: "Ornithopter", OracleID: "a3a98bc9-caa0-49b7-951c-fe4e4f54e4ba"},
		{Name: "Memnite", OracleID: "7663ac7c-1de3-4250-b96a-fae9dbd66a27"},
		{Name: "Ornithopter of Paradise", OracleID: "3a940dfa-a026-4969-8981-ef90cdfbe9ef"},
		{Name: "Hangarback Walker", OracleID: "dde55256-5259-44e7-a267-fca45a7f0d04"},
		{Name: "Myr Retriever", OracleID: "d07d3be3-f69d-4484-8467-cffd43871788"},
		{Name: "Scrap Trawler", OracleID: "164f3f85-21fc-40b7-9871-4f303ba98428"},
		{Name: "Etherium Sculptor", OracleID: "96b87445-6362-4a17-91b2-cbf203fd03fd", Identity: "U"},
		{Name: "Steel Overseer", OracleID: "986ae327-f433-4c58-93dc-afc544b9bfcb"},
		{Name: "Foundry Inspector", OracleID: "18f22960-87ec-43cd-82ea-ec5cabf49ad3"},
		{Name: "Master of Etherium", OracleID: "181e6b3e-c33e-49e7-acb8-67473289a856", Identity: "U"},
		{Name: "Jhoira's Familiar", OracleID: "df4b5f7b-9d83-49b1-bd5e-77d0652eb34c"},
		{Name: "Stridehangar Automaton", OracleID: "9070c98b-fd01-4eeb-a4ec-fc464946c7c0"},
		{Name: "Phyrexian Metamorph", OracleID: "340bbe8b-e987-4c3e-ab4e-9dee63e57d4f", Identity: "U"},
		{Name: "Solemn Simulacrum", OracleID: "00c0543c-2a1f-4425-8283-4062d74a1637"},
		{Name: "Burnished Hart", OracleID: "893fed41-c144-433f-af88-bc7d419b7fb3"},
		{Name: "Palladium Myr", OracleID: "7b0767b8-b504-456e-93bd-218502f73b3d"},
		{Name: "Walking Ballista", OracleID: "4b515bb0-f275-4400-8032-3173b799ab40"},

		// Mana rocks.
		{Name: "Sol Ring", OracleID: "6ad8011d-3471-4369-9d68-b264cc027487"},
		{Name: "Arcane Signet", OracleID: "0bc7f093-bef0-4f1a-852c-4b75ebf54838"},
		{Name: "Mind Stone", OracleID: "c97361b5-af16-4a7b-af85-a429dbaf4ad2"},
		{Name: "Thought Vessel", OracleID: "9965d9c5-2ebf-4a6c-930e-55c5890979be"},
		{Name: "Fellwar Stone", OracleID: "95560508-7ac9-4be9-8a3f-3c7d5b52807b"},
		{Name: "Everflowing Chalice", OracleID: "0a79237e-0811-4a8a-bd4d-db3ca91bff22"},
		{Name: "Commander's Sphere", OracleID: "0b67c4e2-f88b-4e01-85a1-9d5f5b8db13b"},
		{Name: "Worn Powerstone", OracleID: "b166b670-febc-4821-855e-f8d465644c03"},
		{Name: "Hedron Archive", OracleID: "32263baa-d3f0-463f-92b3-4e9938476add"},
		{Name: "Thran Dynamo", OracleID: "a699c663-8131-4045-9265-a83e86609374"},
		{Name: "Wayfarer's Bauble", OracleID: "31f15274-301b-47c5-ba19-0ced04520878"},

		// Thopter and token payoffs.
		{Name: "Thopter Spy Network", OracleID: "49be65fd-3755-410d-b0dc-2e5861ea2552", Identity: "U"},
		{Name: "Efficient Construction", OracleID: "5af48f87-7b94-44de-90e3-91f10ced00d3", Identity: "U"},
		{Name: "Sharding Sphinx", OracleID: "9ebc0144-fc45-4de7-b3d8-ef8cf2e211ae", Identity: "U"},
		{Name: "Thopter Fabricator", OracleID: "8aab9181-c306-471a-8061-9a7a7654dab5", Identity: "U"},
		{Name: "Simulacrum Synthesizer", OracleID: "eb7a1f21-a66d-415b-8520-710b44890bb6", Identity: "U"},
		{Name: "Inspiring Statuary", OracleID: "78b2b68d-6e27-49f3-8373-6e19faf6f679"},
		{Name: "Myr Battlesphere", OracleID: "c53ba31a-ba27-4e17-9a92-311acb1cab29"},
		{Name: "Wurmcoil Engine", OracleID: "d1a60f44-7696-49ee-91fb-cab5b3102962"},
		{Name: "Chief of the Foundry", OracleID: "4fdfa41a-75a5-4b36-8c9e-083e26d137e2"},
		{Name: "Karn, Legacy Reforged", OracleID: "bc8802db-8c94-4eb4-817f-5f4f5408b7a4"},
		{Name: "Tezzeret the Seeker", OracleID: "4aeafdf3-7828-4114-8c74-3bd46357caec", Identity: "U"},
		{Name: "Padeem, Consul of Innovation", OracleID: "0c7ba712-6a99-4d2f-9242-a2163a11f69c", Identity: "U"},
		{Name: "Skullclamp", OracleID: "65986c1b-8e51-4604-b685-d82fa7d1263a"},
		{Name: "Swiftfoot Boots", OracleID: "c8b143ad-43ec-4e0d-a440-e348daa31391"},
		{Name: "Unwinding Clock", OracleID: "153fac93-5d2b-4348-a468-a5eef6a12da3"},
		{Name: "Vedalken Archmage", OracleID: "568cf486-0261-4634-ac36-a6507101b2d0", Identity: "U"},

		// Card draw.
		{Name: "Rhystic Study", OracleID: "53236dd7-845a-444c-96d5-f41ed7325d8f", Identity: "U"},
		{Name: "Mystic Remora", OracleID: "8a52f3c0-2552-4425-b2e3-5496eb2232a7", Identity: "U"},
		{Name: "Thought Monitor", OracleID: "9deded8b-cec4-4ede-a50b-131404d456d4", Identity: "U"},
		{Name: "Reverse Engineer", OracleID: "f56596be-e12f-4ab6-af61-3ef8dc87134c", Identity: "U"},
		{Name: "Ichor Wellspring", OracleID: "5b5ef43b-13fd-4461-8d2d-18be65e9a790"},
		{Name: "Fact or Fiction", OracleID: "437b2dab-15e0-4b9a-a204-58622d37a3b3", Identity: "U"},
		{Name: "Mulldrifter", OracleID: "24d0f5e7-0d9e-4b76-900e-a7274e80312d", Identity: "U"},
		{Name: "Preordain", OracleID: "ac641490-ca14-48d7-8cc4-b69ce984befa", Identity: "U"},

		// Interaction.
		{Name: "Counterspell", OracleID: "cc187110-1148-4090-bbb8-e205694a39f5", Identity: "U"},
		{Name: "Negate", OracleID: "3407fe41-fdd3-4119-8f70-4bc4590a379f", Identity: "U"},
		{Name: "Swan Song", OracleID: "8ddfc283-c9b4-41a5-af88-cf0068e986cc", Identity: "U"},
		{Name: "An Offer You Can't Refuse", OracleID: "234a734b-ba28-4f1b-9d01-3c3e7d516590", Identity: "U"},
		{Name: "Mana Drain", OracleID: "74d3277a-38e5-4732-afed-084a56148f20", Identity: "U"},
		{Name: "Pongify", OracleID: "05849bd6-8f38-4031-be2b-e2aa03beb8cc", Identity: "U"},
		{Name: "Rapid Hybridization", OracleID: "06692cd9-ac2f-4a32-8fd1-043ba3c0fe71", Identity: "U"},
		{Name: "Cyclonic Rift", OracleID: "d75b9c82-1b49-4c3e-a1b5-aeef57d6644b", Identity: "U"},
		{Name: "Aetherize", OracleID: "7c779721-cd1b-4696-9ae9-68ccc284ed2a", Identity: "U"},
		{Name: "Imprisoned in the Moon", OracleID: "339a7418-9a29-4eae-a3c2-ea590d175936", Identity: "U"},

		// Lands.
		{Name: "Seat of the Synod", OracleID: "39451b4d-cd7a-40da-b457-cb51b609173f", Identity: "U"},
		{Name: "Darksteel Citadel", OracleID: "8dc067bf-f78f-4ac4-b6e7-b305c42cf0bc"},
		{Name: "Treasure Vault", OracleID: "3c43efd6-b1a8-452c-ae20-9a936c3340ab"},
		{Name: "Inventors' Fair", OracleID: "91d4a5fe-fd6d-4b14-a63f-61b4d0ecd9c4"},
		{Name: "Buried Ruin", OracleID: "3644f316-f9a3-46c9-9b1e-747f86cf4ead"},
		{Name: "Academy Ruins", OracleID: "a3da7d5b-2c2b-45fe-b9c5-413b8c8fc0a2", Identity: "U"},
		{Name: "Otawara, Soaring City", OracleID: "e9b6a394-691c-425a-9307-76d8edc7375e", Identity: "U"},
		{Name: "Rogue's Passage", OracleID: "f29dc596-2121-4421-8463-15f6c2e8b9b3"},
		{Name: "Reliquary Tower", OracleID: "c23e5b80-08d2-4e24-9908-fe2aa4f30f6f"},
		{Name: "Island", Count: 28, Basic: true},
	},
}
