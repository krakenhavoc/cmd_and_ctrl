package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch12_test.go — card-level coverage for the card-coverage
// roadmap's batch 12 (#305, `edhrec_rank` 1316–1420): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play or attack. Helpers from
// the earlier batch test files are reused by name; new ones are
// b12-prefixed.

const (
	B12AwakenTheWoodsOracle       = "3cc79d36-1a24-4395-8ad1-915a65db8a60"
	B12LaeliaOracle               = "a0be9bb2-3234-4c6c-b8ce-0879b1f43003"
	B12ValakutOracle              = "1bc44216-4e06-4f66-89b7-5c327004604e"
	B12KrarkClanIronworksOracle   = "68e1f7e0-a9b3-437f-8086-0c0cb85f2880"
	B12SkirkProspectorOracle      = "c18013e4-0b99-44e3-a2b2-027ace68723a"
	B12ColossalMajestyOracle      = "cac95494-0db0-4bec-8665-998431a6f76b"
	B12PersistOracle              = "367d4cf2-270f-4236-af10-7d5e8ea8c9fe"
	B12PristineTalismanOracle     = "1b3d7fce-e9fe-4176-9d9a-472415826cdd"
	B12SoulGuideLanternOracle     = "1b5e6560-ff2e-4475-96cb-63f64c8a86db"
	B12ThirdPathIconoclastOracle  = "f7156897-2b02-4ecd-868d-d4d59244e9ed"
	B12CrimeNovelistOracle        = "5996b1d0-7fe3-4fec-8730-9db901d887b1"
	B12SigilOfTheEmptyThroneOrcl  = "3b9b5b22-5a7d-4e37-a870-ca0f0efa4f36"
	B12BonehoardDracosaurOracle   = "7ba550a5-81fc-42c2-8df8-ad455d938605"
	B12ArchangelOfThuneOracle     = "4f2d4538-dc1d-4c09-964b-b0d7c240fb7d"
	B12SkullProphetOracle         = "39b51f62-8421-42d1-86d3-74fd6e6f31a2"
	B12MartialCoupOracle          = "2b7c4dab-e432-4b34-b058-3cec5c0d72df"
	B12TerastodonOracle           = "17975a20-2c13-4325-be1e-0bcda5063a2d"
	B12LifecraftersBestiaryOracle = "5f2a3797-28aa-4c7a-ba2b-fd243a1747fd"
	B12AllWillBeOneOracle         = "477374dc-042c-48f7-9ebe-99c15d8ae04f"
	B12ThousandYearStormOracle    = "dd4cf149-2fae-40e5-b50b-639f6bcec65e"
	B12RazorvergeThicketOracle    = "94f6c407-e665-4032-be13-a01e40c1f306"
	B12TimeWipeOracle             = "36c78a5f-0148-4596-a346-f8e35037b694"
	B12CatharCommandoOracle       = "774dce79-67e0-4820-8013-c7a7347993ce"
	B12AltarOfTheBroodOracle      = "c3aafcdd-c890-4971-b8a9-5bfcad794c0b"
	B12DuskshellCrawlerOracle     = "e54ad5a1-e79d-42db-a3e9-5caecae62c9f"
	B12SoddenVerdureOracle        = "0070db93-142b-4d04-afd3-836792dc134b"
	B12GoldMyrOracle              = "bd6af7b3-b30f-4a65-a18f-8655f778e76a"
	B12TangledIsletOracle         = "4b1f68a2-b606-4c64-bd44-a9714808316d"
	B12GraveTitanAlreadyOID       = "f3abd4d1-a975-4e85-8684-aa0fce029670"
	B12ReverberateAlreadyOID      = "a1f55890-31c5-4ed4-a2cd-7a4a9f05f8ca"
)

// b12Push seeds a catalog permanent through the timestamped push so
// its statics, replacements and triggers are live, with no summoning
// sickness.
func B12Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b12Creature seeds a non-catalog creature with a type line, able to
// attack and tap.
func B12Creature(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b12Permanent seeds a non-catalog noncreature permanent.
func B12Permanent(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Owner: owner, Controller: owner,
	})
}

// b12PlayFromHand puts a card with the given type line in the active
// seat's hand and plays it — a land play or a cast, the engine
// decides by type — from a main phase.
func B12PlayFromHand(t *testing.T, g *game.Game, name, typeLine, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	// #500: these fixtures build a board of several lands inside one
	// main phase, so each helper call buys itself a land play. See
	// playLandFromHand (temples_test.go) for the same note.
	active.LandDropsPerTurn++
	if err := g.CastSpell(active.ID, id, params); err != nil {
		t.Fatalf("play %s: %v", name, err)
	}
	return id
}

// b12Card looks a card up wherever it is.
func B12Card(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var out game.Card
	found := false
	g.WithWriteLock(func() { out, found = g.LookupCardForEffect(id) })
	if !found {
		t.Fatalf("card %s is nowhere", id)
	}
	return out
}

// b12ZoneOf is the zone a card sits in, or "" when it is nowhere.
func B12ZoneOf(g *game.Game, id uuid.UUID) game.ZoneKind {
	var kind game.ZoneKind
	g.WithWriteLock(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			kind = z.Kind
		}
	})
	return kind
}

// b12Counter reads one counter kind off a card.
func B12Counter(t *testing.T, g *game.Game, id uuid.UUID, kind string) int {
	t.Helper()
	return B12Card(t, g, id).Counters[kind]
}

// b12ToMyNextUpkeep walks the cursor around the table to seat 0's
// next upkeep, passing priority through every stop.
func B12ToMyNextUpkeep(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 400; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == 0 {
			return
		}
	}
	t.Fatal("never reached seat 0's next upkeep")
}

// b12Bolt casts Lightning Bolt from the active seat at a target and
// resolves it.
func B12Bolt(t *testing.T, g *game.Game, target game.TargetRef) uuid.UUID {
	t.Helper()
	id := CastCatalogSpell(t, g, "Lightning Bolt", "Instant", LightningBoltOracle, []game.TargetRef{target})
	PassPriorityAroundTable(t, g)
	return id
}

// b12ExiledPlayableBy reports whether a card is in exile with a live
// play grant for `player`.
func B12ExiledPlayableBy(t *testing.T, g *game.Game, id, player uuid.UUID) bool {
	t.Helper()
	if B12ZoneOf(g, id) != game.ZoneExile {
		return false
	}
	c := B12Card(t, g, id)
	perm := g.CastPermissionOnCardByIDForEffect(c.InstanceID)
	return PermissionLive(g, perm, player)
}
