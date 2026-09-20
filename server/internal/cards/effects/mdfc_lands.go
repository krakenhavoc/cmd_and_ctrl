package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mdfc_lands.go — the BACK faces of every modal double-faced card
// whose second half is a land. Sixty of the hundred modal_dfc oracle
// IDs in Magic, generated from the Scryfall dump and grouped by the
// one clause that distinguishes them.
//
// # Why these are a cycle and not sixty cards
//
// An MDFC land back prints, at most, two things: a mana ability and
// an entry clause. The entry clauses come in exactly three shapes:
//
//	"As this land enters, you may pay 3 life. If you don't,     (15)
//	 it enters tapped."
//	"This land enters tapped."                                  (35)
//	(nothing — it just enters)                                  (10)
//
// The first is BYTE-IDENTICAL to the Ravnica shockland clause with a
// different number, so it is EntersTappedUnlessYouPayLife(name, 3) —
// the machinery PR #268 shipped, reused whole. The second is
// SelfEntersTapped(). The third needs no replacement at all. None of
// the sixty needed a line of new rules code; what they needed was a
// face model to hang off, which is ADR 0034.
//
// # Why they register under "<oracle_id>#1"
//
// Scryfall issues ONE oracle_id per card, so Sea Gate Restoration and
// Sea Gate, Reborn share one — and Register panics on a duplicate,
// deliberately. game.CatalogKey makes the key composite: face 0 keeps
// the bare oracle ID (so every existing spec and every single-faced
// card is untouched) and face N takes "<oracle_id>#N". These specs
// are the first users of the back-face half of that keyspace, and the
// panic stays exactly as strict as it was — it now simply has two
// distinct keys to police instead of one.
//
// # The entry prompt actually works here
//
// The pay-life prompt only fires where ReplacementEvent.entryResumable
// is set. That is CastSpell's land branch — the path an MDFC back
// takes when a player plays it as a land — and, since #478, stack
// resolution, the library search, the exile return and the
// reanimation, so a fetched or blinked MDFC land back is asked too.
// The one entry that still takes the un-paid branch is
// putOntoBattlefieldFromZoneLocked's batch (see shocklands.go, and
// server/internal/game/battlefield_put.go:313 for the event built
// without the flag): weaker than printed, never stronger.
//
// The caveat below says that narrower thing since #1051. It used to
// say "by another spell", which #478 made an overstatement — a search
// is a spell, and a fetched back face is prompted. The shocklands'
// caveat is the same sentence with 2 life for 3, and
// TestWhichShocklandEntrySitesOfferThePayment holds both halves of it
// against the engine for the whole family.

// mdfcLandLifeCost is what every pay-life MDFC land back charges.
// All fifteen ask for the same number; naming it keeps the prompt
// copy and the cost from drifting apart, as shocklandLifeCost does
// for the shocklands' 2.
const mdfcLandLifeCost = 3

// mdfcLandBack is one row of the cycle: the card's oracle ID, the
// FRONT face's name (for readability when scanning the table against
// a decklist), the BACK face's name (which is what the prompt and
// the mana-ability label say), and the colours that back face taps
// for.
type mdfcLandBack struct {
	oracleID string
	front    string
	back     string
	colors   []string
}

// backManaAbility builds the tap-for-mana ability for a land back.
// One colour is the common case; a handful of the Duskmourn and
// Bloomburrow backs tap for either of two, which is the same
// "{A|B}" pipe the dual lands use so the client offers one picker
// rather than two menu entries.
func backManaAbility(colors []string) ManaAbility {
	switch len(colors) {
	case 0:
		return ManaAbility{}
	case 1:
		return ManaAbility{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{" + colors[0] + "}",
			Label:    "Add {" + colors[0] + "}",
		}
	default:
		return dualManaAbility(colors[0], colors[1])
	}
}

// mdfcLandBackKeys is every catalog key this file registers, in
// registration order.
//
// It exists because "#1" stopped meaning "an MDFC land back" in S32,
// when the Siege back faces became castable and took keys in the same
// keyspace. TestMDFCLandBackCycleIsComplete used to count suffixes,
// which quietly turned into "count every back face in the catalog"
// the moment a second kind existed — a guard that grows a false
// failure every time an unrelated card ships is worse than no guard.
// The list is the cycle's own record of itself.
var mdfcLandBackKeys []string

func registerMDFCLandBacks(rows []mdfcLandBack, reps func(back string) []game.ReplacementEffect) {
	for _, r := range rows {
		mdfcLandBackKeys = append(mdfcLandBackKeys, game.CatalogKeyForFace(r.oracleID, 1))
		// Capture per iteration — the closures inside the
		// replacement outlive the loop body.
		row := r
		Register(Spec{
			// The back face's catalog key. game.CatalogKey(card)
			// produces exactly this string once CastSpell has called
			// SetFace(1), which is what routes a player's "play it as
			// a land" choice to this spec rather than the front
			// face's.
			OracleID:      game.CatalogKeyForFace(row.oracleID, 1),
			Name:          row.back,
			Completeness:  CompletenessCaveats,
			Caveats:       []string{"A back face a spell PUTS onto the battlefield out of a hand or library — Genesis Wave, Coiling Oracle, Arboreal Grazer — always enters tapped. Playing it as a land, or fetching it with a search (a fetchland, Farseek), does offer the 3 life."},
			Replacements:  reps(row.back),
			ManaAbilities: []ManaAbility{backManaAbility(row.colors)},
		})
	}
}

func init() {
	// --- "you may pay 3 life. If you don't, it enters tapped" ----
	//
	// The Zendikar Rising mythic cycle (Sea Gate Restoration,
	// Agadeem's Awakening, Emeria's Call, Shatterskull Smashing,
	// Turntimber Symbiosis) plus the later designs that copied the
	// template. Same shape as a shockland, three life instead of two.
	registerMDFCLandBacks([]mdfcLandBack{
		{"562d71b9-1646-474e-9293-55da6947a758", "Agadeem's Awakening", "Agadeem, the Undercrypt", []string{"B"}},
		{"727f3201-1cfc-4ab2-9dfe-be4f7251f42f", "Boggart Trawler", "Boggart Bog", []string{"B"}},
		{"9d581188-ce80-494e-bd38-f411e1f4efb5", "Bridgeworks Battle", "Tanglespan Bridgeworks", []string{"G"}},
		{"2699005b-a471-429f-a9d8-fbf2077ee2fd", "Disciple of Freyalise", "Garden of Freyalise", []string{"G"}},
		{"6ec2a242-9068-4ee2-8ac8-8341cc570f56", "Emeria's Call", "Emeria, Shattered Skyclave", []string{"W"}},
		{"053a69d8-2b5e-4f14-8b02-ca405891dc4a", "Fell the Profane", "Fell Mire", []string{"B"}},
		{"573151f0-00d4-4a8a-8a09-745c5f376532", "Hydroelectric Specimen", "Hydroelectric Laboratory", []string{"U"}},
		{"f3d48efa-910a-4872-a5b1-a353c5dbce99", "Pinnacle Monk", "Mystic Peak", []string{"R"}},
		{"5da954fa-9001-4557-825c-1462035d21ed", "Razorgrass Ambush", "Razorgrass Field", []string{"W"}},
		{"4a8d41fe-e04d-484b-a7d1-19be311e6ca7", "Sea Gate Restoration", "Sea Gate, Reborn", []string{"U"}},
		{"78301998-fd9b-4cd5-afad-dbcb43cac2a7", "Shatterskull Smashing", "Shatterskull, the Hammer Pass", []string{"R"}},
		{"bcc6eece-75ea-494c-b33a-d4477d504e0b", "Sink into Stupor", "Soporific Springs", []string{"U"}},
		{"c95309e9-5c2f-4518-b2fd-825d3d0a4ae0", "Sundering Eruption", "Volcanic Fissure", []string{"R"}},
		{"403b59f3-7ade-4bc2-a3e6-de0c3c700f18", "Turntimber Symbiosis", "Turntimber, Serpentine Wood", []string{"G"}},
		{"0355249a-8e4e-41db-9cea-1b901faffbe6", "Witch Enchanter", "Witch-Blessed Meadow", []string{"W"}},
	}, func(back string) []game.ReplacementEffect {
		return []game.ReplacementEffect{
			EntersTappedUnlessYouPayLife(back, mdfcLandLifeCost),
		}
	})

	// --- "This land enters tapped." -----------------------------
	//
	// The Zendikar Rising commons and uncommons, and every MDFC land
	// back printed since. No choice, no prompt.
	registerMDFCLandBacks([]mdfcLandBack{
		{"afedce7b-0e18-40ad-a26a-1933fddb560d", "Akoum Warrior", "Akoum Teeth", []string{"R"}},
		{"d2075f58-b0e9-4e85-b7e6-0523a27a1d5b", "Bala Ged Recovery", "Bala Ged Sanctuary", []string{"G"}},
		{"b03de49d-246f-44e2-9487-9e4e43ec7be4", "Beyeen Veil", "Beyeen Coast", []string{"U"}},
		{"34320ebf-da97-44a4-bbeb-a9da06548289", "Blackbloom Rogue", "Blackbloom Bog", []string{"B"}},
		{"c52fc8a1-43c6-41f8-b010-03be7c89ef1d", "Bloodsoaked Insight", "Sanguine Morass", []string{"B", "R"}},
		{"db19a27a-ee22-4931-ae3c-0ce21f456ea6", "Drowner of Truth", "Drowned Jungle", []string{"G", "U"}},
		{"c178953c-3888-4edd-9d0c-265bd82b1d24", "Glasspool Mimic", "Glasspool Shore", []string{"U"}},
		{"3a3e8c9b-e458-4661-980d-0a84a4c2452b", "Glasswing Grace", "Age-Graced Chapel", []string{"W", "B"}},
		{"37783ce6-af58-4ef6-8ab4-587079970307", "Hagra Mauling", "Hagra Broodpit", []string{"B"}},
		{"941a4b14-ea2a-4bd0-8cc2-d609f80df32c", "Jwari Disruption", "Jwari Ruins", []string{"U"}},
		{"0bb73c07-0220-4ba9-8d85-3c357c223833", "Kabira Takedown", "Kabira Plateau", []string{"W"}},
		{"2ac1c95c-2a9d-40bc-9cad-9cadfa3f19f7", "Kazandu Mammoth", "Kazandu Valley", []string{"G"}},
		{"f8410804-632b-4f18-9a73-6dccc7e4582d", "Kazuul's Fury", "Kazuul's Cliffs", []string{"R"}},
		{"37a55560-6e32-4f54-b9a8-fd157aea6eb5", "Khalni Ambush", "Khalni Territory", []string{"G"}},
		{"ad225ec2-ff3a-48f6-81a7-dfdd1b75e1f7", "Legion Leadership", "Legion Stronghold", []string{"R", "W"}},
		{"342e08f9-d4d0-4408-8621-66e087058616", "Makindi Stampede", "Makindi Mesas", []string{"W"}},
		{"a731e87b-8d99-4b64-8ee3-8e540d652366", "Malakir Rebirth", "Malakir Mire", []string{"B"}},
		{"15fc4e74-300e-4c2d-8ed7-004553b2f7c2", "Ondu Inversion", "Ondu Skyruins", []string{"W"}},
		{"b0fd6889-20b4-439b-aa97-2e90aca1675a", "Pelakka Predation", "Pelakka Caverns", []string{"B"}},
		{"8dd6d060-d023-48a6-85cb-7a5521b6257b", "Revitalizing Repast", "Old-Growth Grove", []string{"B", "G"}},
		{"bbd569cc-bc21-46df-b8eb-5b5bcd8fe762", "Rush of Inspiration", "Crackling Falls", []string{"U", "R"}},
		{"d54e4e37-042b-44a5-918d-757308545d4d", "Sejiri Shelter", "Sejiri Glacier", []string{"W"}},
		{"b0182ca0-f353-4012-9121-6f4ac9f7a046", "Silundi Vision", "Silundi Isle", []string{"U"}},
		{"da9e3910-9a1c-43a9-9138-ca971b2bccae", "Skyclave Cleric", "Skyclave Basilica", []string{"W"}},
		{"81b61770-2ed5-4a50-84d0-97790002fc5a", "Song-Mad Treachery", "Song-Mad Ruins", []string{"R"}},
		{"81036c9f-fe0a-45a7-bcd5-0d344f31055a", "Spikefield Hazard", "Spikefield Cave", []string{"R"}},
		{"1a8c996d-ca93-4c17-ace5-66ecd6b99317", "Strength of the Harvest", "Haven of the Harvest", []string{"G", "W"}},
		{"eb7b1284-0b2c-4b6a-a389-b2b932838083", "Stump Stomp", "Burnwillow Clearing", []string{"R", "G"}},
		{"b592568b-11b0-4081-90a7-30cfb9c1ba80", "Suppression Ray", "Orderly Plaza", []string{"W", "U"}},
		{"53542c79-a62a-4d6a-97db-5296e9c68302", "Tangled Florahedron", "Tangled Vale", []string{"G"}},
		{"6bc668f4-8fc7-4aaf-891b-277d8328b376", "Umara Wizard", "Umara Skyfalls", []string{"U"}},
		{"ff0ab867-b710-4b1a-baed-95fc3cf68f79", "Valakut Awakening", "Valakut Stoneforge", []string{"R"}},
		{"ce148a0c-6c63-49d5-a156-99efae4e367a", "Vastwood Fortification", "Vastwood Thicket", []string{"G"}},
		{"e6ad1be9-f13d-4590-b3db-e2d0fff46f03", "Waterlogged Teachings", "Inundated Archive", []string{"U", "B"}},
		{"d9f11985-e460-425d-b083-9cb0edf1983a", "Zof Consumption", "Zof Bloodbog", []string{"B"}},
	}, func(string) []game.ReplacementEffect {
		return []game.ReplacementEffect{SelfEntersTapped()}
	})

	// --- no entry clause ----------------------------------------
	//
	// The Kaldheim Pathway cycle. The back is an untapped
	// single-colour land, so the spec exists purely to declare the
	// mana ability: the synthetic basic-land ability is derived from
	// the type line, and "Land" with no basic land type derives
	// nothing.
	registerMDFCLandBacks([]mdfcLandBack{
		{"59d22de5-e310-44d7-89cf-ef3529e40cef", "Barkchannel Pathway", "Tidechannel Pathway", []string{"U"}},
		{"e580a229-e800-4746-9d37-c32fcef8de28", "Blightstep Pathway", "Searstep Pathway", []string{"R"}},
		{"7c304547-a4b1-46c9-baed-16d2bfbe16eb", "Branchloft Pathway", "Boulderloft Pathway", []string{"W"}},
		{"1c633e02-95ef-445e-b4e0-fbfbc5ed9cc9", "Brightclimb Pathway", "Grimclimb Pathway", []string{"B"}},
		{"144119bc-7fd1-45c5-9e29-f742e7c255ac", "Clearwater Pathway", "Murkwater Pathway", []string{"B"}},
		{"727ca426-f4cc-4218-8ae5-8c427af2e816", "Cragcrown Pathway", "Timbercrown Pathway", []string{"G"}},
		{"868e6e68-4367-4073-a864-235d5961ae56", "Darkbore Pathway", "Slitherbore Pathway", []string{"G"}},
		{"461b3f2f-fcee-4160-abfa-061f8b6a784f", "Hengegate Pathway", "Mistgate Pathway", []string{"U"}},
		{"a9b8d020-4d72-4934-8942-df29ef19fc1d", "Needleverge Pathway", "Pillarverge Pathway", []string{"W"}},
		{"4924b3a4-a218-4783-8a4d-82361fdecc78", "Riverglide Pathway", "Lavaglide Pathway", []string{"R"}},
	}, func(string) []game.ReplacementEffect { return nil })
}
