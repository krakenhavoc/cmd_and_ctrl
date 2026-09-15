package effects

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
	b12AwakenTheWoodsOracle       = "3cc79d36-1a24-4395-8ad1-915a65db8a60"
	b12LaeliaOracle               = "a0be9bb2-3234-4c6c-b8ce-0879b1f43003"
	b12ValakutOracle              = "1bc44216-4e06-4f66-89b7-5c327004604e"
	b12KrarkClanIronworksOracle   = "68e1f7e0-a9b3-437f-8086-0c0cb85f2880"
	b12SkirkProspectorOracle      = "c18013e4-0b99-44e3-a2b2-027ace68723a"
	b12ColossalMajestyOracle      = "cac95494-0db0-4bec-8665-998431a6f76b"
	b12PersistOracle              = "367d4cf2-270f-4236-af10-7d5e8ea8c9fe"
	b12PristineTalismanOracle     = "1b3d7fce-e9fe-4176-9d9a-472415826cdd"
	b12SoulGuideLanternOracle     = "1b5e6560-ff2e-4475-96cb-63f64c8a86db"
	b12ThirdPathIconoclastOracle  = "f7156897-2b02-4ecd-868d-d4d59244e9ed"
	b12CrimeNovelistOracle        = "5996b1d0-7fe3-4fec-8730-9db901d887b1"
	b12SigilOfTheEmptyThroneOrcl  = "3b9b5b22-5a7d-4e37-a870-ca0f0efa4f36"
	b12BonehoardDracosaurOracle   = "7ba550a5-81fc-42c2-8df8-ad455d938605"
	b12ArchangelOfThuneOracle     = "4f2d4538-dc1d-4c09-964b-b0d7c240fb7d"
	b12SkullProphetOracle         = "39b51f62-8421-42d1-86d3-74fd6e6f31a2"
	b12MartialCoupOracle          = "2b7c4dab-e432-4b34-b058-3cec5c0d72df"
	b12TerastodonOracle           = "17975a20-2c13-4325-be1e-0bcda5063a2d"
	b12LifecraftersBestiaryOracle = "5f2a3797-28aa-4c7a-ba2b-fd243a1747fd"
	b12AllWillBeOneOracle         = "477374dc-042c-48f7-9ebe-99c15d8ae04f"
	b12ThousandYearStormOracle    = "dd4cf149-2fae-40e5-b50b-639f6bcec65e"
	b12RazorvergeThicketOracle    = "94f6c407-e665-4032-be13-a01e40c1f306"
	b12TimeWipeOracle             = "36c78a5f-0148-4596-a346-f8e35037b694"
	b12CatharCommandoOracle       = "774dce79-67e0-4820-8013-c7a7347993ce"
	b12AltarOfTheBroodOracle      = "c3aafcdd-c890-4971-b8a9-5bfcad794c0b"
	b12DuskshellCrawlerOracle     = "e54ad5a1-e79d-42db-a3e9-5caecae62c9f"
	b12SoddenVerdureOracle        = "0070db93-142b-4d04-afd3-836792dc134b"
	b12GoldMyrOracle              = "bd6af7b3-b30f-4a65-a18f-8655f778e76a"
	b12TangledIsletOracle         = "4b1f68a2-b606-4c64-bd44-a9714808316d"
	b12GraveTitanAlreadyOID       = "f3abd4d1-a975-4e85-8684-aa0fce029670"
	b12ReverberateAlreadyOID      = "a1f55890-31c5-4ed4-a2cd-7a4a9f05f8ca"
)

// b12Push seeds a catalog permanent through the timestamped push so
// its statics, replacements and triggers are live, with no summoning
// sickness.
func b12Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b12Creature seeds a non-catalog creature with a type line, able to
// attack and tap.
func b12Creature(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b12Permanent seeds a non-catalog noncreature permanent.
func b12Permanent(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Owner: owner, Controller: owner,
	})
}

// b12PlayFromHand puts a card with the given type line in the active
// seat's hand and plays it — a land play or a cast, the engine
// decides by type — from a main phase.
func b12PlayFromHand(t *testing.T, g *game.Game, name, typeLine, oracle string, params game.CastSpellParams) uuid.UUID {
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
	if err := g.CastSpell(active.ID, id, params); err != nil {
		t.Fatalf("play %s: %v", name, err)
	}
	return id
}

// b12Card looks a card up wherever it is.
func b12Card(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
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
func b12ZoneOf(g *game.Game, id uuid.UUID) game.ZoneKind {
	var kind game.ZoneKind
	g.WithWriteLock(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			kind = z.Kind
		}
	})
	return kind
}

// b12Counter reads one counter kind off a card.
func b12Counter(t *testing.T, g *game.Game, id uuid.UUID, kind string) int {
	t.Helper()
	return b12Card(t, g, id).Counters[kind]
}

// b12ToMyNextUpkeep walks the cursor around the table to seat 0's
// next upkeep, passing priority through every stop.
func b12ToMyNextUpkeep(t *testing.T, g *game.Game) {
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
func b12Bolt(t *testing.T, g *game.Game, target game.TargetRef) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{target})
	passPriorityAroundTable(t, g)
	return id
}

// b12ExiledPlayableBy reports whether a card is in exile with a live
// play grant for `player`.
func b12ExiledPlayableBy(t *testing.T, g *game.Game, id, player uuid.UUID) bool {
	t.Helper()
	if b12ZoneOf(g, id) != game.ZoneExile {
		return false
	}
	c := b12Card(t, g, id)
	return c.ExilePlay.Active(player, g.Turn.Number)
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Four are rows
// in cycle tables (Razorverge Thicket in fastlands.go, Sodden Verdure
// in battle_lands.go, Gold Myr in mana_myr.go, Tangled Islet in
// dominaria_united_duals.go), so a transposed row is invisible until
// someone plays that exact card. Two of the issue's 36 were already
// on main — Grave Titan (#73) and Reverberate (#419) — pinned here so
// the table matches the issue.
func TestBatch12CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b12AwakenTheWoodsOracle:       "Awaken the Woods",
		b12LaeliaOracle:               "Laelia, the Blade Reforged",
		b12ValakutOracle:              "Valakut, the Molten Pinnacle",
		b12KrarkClanIronworksOracle:   "Krark-Clan Ironworks",
		b12SkirkProspectorOracle:      "Skirk Prospector",
		b12ColossalMajestyOracle:      "Colossal Majesty",
		b12PersistOracle:              "Persist",
		b12PristineTalismanOracle:     "Pristine Talisman",
		b12SoulGuideLanternOracle:     "Soul-Guide Lantern",
		b12ThirdPathIconoclastOracle:  "Third Path Iconoclast",
		b12CrimeNovelistOracle:        "Crime Novelist",
		b12SigilOfTheEmptyThroneOrcl:  "Sigil of the Empty Throne",
		b12BonehoardDracosaurOracle:   "Bonehoard Dracosaur",
		b12ArchangelOfThuneOracle:     "Archangel of Thune",
		b12SkullProphetOracle:         "Skull Prophet",
		b12MartialCoupOracle:          "Martial Coup",
		b12TerastodonOracle:           "Terastodon",
		b12LifecraftersBestiaryOracle: "Lifecrafter's Bestiary",
		b12AllWillBeOneOracle:         "All Will Be One",
		b12ThousandYearStormOracle:    "Thousand-Year Storm",
		b12RazorvergeThicketOracle:    "Razorverge Thicket",
		b12TimeWipeOracle:             "Time Wipe",
		b12CatharCommandoOracle:       "Cathar Commando",
		b12AltarOfTheBroodOracle:      "Altar of the Brood",
		b12DuskshellCrawlerOracle:     "Duskshell Crawler",
		b12SoddenVerdureOracle:        "Sodden Verdure",
		b12GoldMyrOracle:              "Gold Myr",
		b12TangledIsletOracle:         "Tangled Islet",
		b12GraveTitanAlreadyOID:       "Grave Titan",
		b12ReverberateAlreadyOID:      "Reverberate",
	}
	if len(want) != 30 {
		t.Fatalf("the batch registers 30 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

// The two cycle rows produce their printed pairs behind their printed
// conditions — the loop-leak canary every land table carries.
func TestB12LandRowsProduceTheirPrintedColours(t *testing.T) {
	for _, tc := range []struct{ oracle, name, produced string }{
		{b12RazorvergeThicketOracle, "Razorverge Thicket", "{G|W}"},
		{b12SoddenVerdureOracle, "Sodden Verdure", "{G|U}"},
		{b12TangledIsletOracle, "Tangled Islet", "{G|U}"},
	} {
		spec, ok := Lookup(tc.oracle)
		if !ok {
			t.Fatalf("%s not registered", tc.name)
		}
		if len(spec.Replacements) != 1 {
			t.Errorf("%s: %d replacements, want 1 (enters tapped unless …)", tc.name, len(spec.Replacements))
		}
		if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != tc.produced {
			t.Errorf("%s: mana abilities %+v, want one producing %s", tc.name, spec.ManaAbilities, tc.produced)
		}
	}
}

func TestB12GoldMyrTapsForWhiteOnceAwake(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	myr := pushCatalogPermanent(g, me.ID, "Gold Myr", "Artifact Creature — Myr", b12GoldMyrOracle, true)
	if err := g.ActivateManaAbility(me.ID, myr, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("a Myr that entered this turn is summoning sick")
	}
	b10Awake(g, myr)
	if err := g.ActivateManaAbility(me.ID, myr, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "W" {
		t.Errorf("pool %v, want [W]", got)
	}
}

func TestB12TangledIsletEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	id := b12PlayFromHand(t, g, "Tangled Islet", "Land — Forest Island", b12TangledIsletOracle, game.CastSpellParams{})
	if !b12Card(t, g, id).Tapped {
		t.Fatal("Tangled Islet enters tapped")
	}
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("a tapped entry is a replacement, not a tap: %d tap events", n)
	}
}

func TestB12SoddenVerdureEntersUntappedWithTwoBasics(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	seedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	id := b12PlayFromHand(t, g, "Sodden Verdure", "Land — Forest Island", b12SoddenVerdureOracle, game.CastSpellParams{})
	if b12Card(t, g, id).Tapped {
		t.Error("with two basics Sodden Verdure enters untapped")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	seedLandOnBattlefield(g2, me2.ID, "Forest", "Basic Land — Forest")
	id2 := b12PlayFromHand(t, g2, "Sodden Verdure", "Land — Forest Island", b12SoddenVerdureOracle, game.CastSpellParams{})
	if !b12Card(t, g2, id2).Tapped {
		t.Error("with one basic Sodden Verdure enters tapped")
	}
}

func TestB12RazorvergeThicketEntersUntappedEarly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	seedLandOnBattlefield(g, me.ID, "Plains", "Basic Land — Plains")
	id := b12PlayFromHand(t, g, "Razorverge Thicket", "Land", b12RazorvergeThicketOracle, game.CastSpellParams{})
	if b12Card(t, g, id).Tapped {
		t.Error("with two other lands the fastland enters untapped")
	}
	seedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	id2 := b12PlayFromHand(t, g, "Razorverge Thicket", "Land", b12RazorvergeThicketOracle, game.CastSpellParams{})
	if !b12Card(t, g, id2).Tapped {
		t.Error("with four other lands the fastland enters tapped")
	}
}

// --- Awaken the Woods ----------------------------------------------

func TestB12AwakenTheWoodsMakesXForestDryadLandCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12PlayFromHand(t, g, "Awaken the Woods", "Sorcery", b12AwakenTheWoodsOracle, game.CastSpellParams{XValue: 3})
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, id) != game.ZoneGraveyard {
		t.Fatal("the sorcery should be in the graveyard")
	}
	var dryads []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Dryad" && c.Controller == me.ID {
			dryads = append(dryads, c)
		}
	}
	if len(dryads) != 3 {
		t.Fatalf("X=3 made %d Dryads, want 3", len(dryads))
	}
	d := dryads[0]
	if !d.IsLand() || !d.IsCreature() || !d.HasSubtype("Forest") || !d.HasSubtype("Dryad") {
		t.Errorf("the token must be a Forest Dryad land creature, got %q", d.TypeLine)
	}
	if !IsToken(d) {
		t.Error("the Dryad must be a token")
	}
	abilities := game.ManaAbilitiesForCard(d)
	if len(abilities) != 1 || abilities[0].Produced != "{G}" {
		t.Errorf("a Forest token taps for {G} from its type line, got %+v", abilities)
	}
	// It is a creature that just entered, so it is summoning sick.
	if err := g.ActivateManaAbility(me.ID, d.InstanceID, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a Dryad that entered this turn must be summoning sick")
	}
}

// --- Laelia, the Blade Reforged ------------------------------------

func TestB12LaeliaAttacksForACardAndGrows(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	laelia := b12Push(g, me.ID, "Laelia, the Blade Reforged", "Legendary Creature — Spirit Warrior", b12LaeliaOracle, 2, 2)
	top := b10LibraryTop(me, "Top Card", "Sorcery", "{R}", 0, 0)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(laelia, victim.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !b12ExiledPlayableBy(t, g, top, me.ID) {
		t.Fatal("the top card should be in exile with a play grant for Laelia's controller")
	}
	if got := b12Counter(t, g, laelia, "+1/+1"); got != 1 {
		t.Errorf("Laelia's own exile grows her: %d counters, want 1", got)
	}
	if p := b12Card(t, g, laelia).CurrentPower(); p != 3 {
		t.Errorf("Laelia power %d, want 3", p)
	}
}

func TestB12LaeliaGrowsOncePerBatchOfExiledCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	laelia := b12Push(g, me.ID, "Laelia, the Blade Reforged", "Legendary Creature — Spirit Warrior", b12LaeliaOracle, 2, 2)
	a := pushGraveyardCardForTest(me, "Dead A")
	b := pushGraveyardCardForTest(me, "Dead B")
	theirs := pushGraveyardCardForTest(opp, "Their Dead")

	// One effect exiling two of your graveyard cards is one batch.
	g.WithWriteLock(func() {
		_ = g.ExileCardForEffect(a)
		_ = g.ExileCardForEffect(b)
	})
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, laelia, "+1/+1"); got != 1 {
		t.Errorf("two cards in one batch: %d counters, want 1", got)
	}

	// An opponent's graveyard is not yours.
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(theirs) })
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, laelia, "+1/+1"); got != 1 {
		t.Errorf("an opponent's card leaving their graveyard grew Laelia: %d counters", got)
	}
}

// --- Valakut, the Molten Pinnacle ----------------------------------

func TestB12ValakutEntersTappedAndTapsForRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12PlayFromHand(t, g, "Valakut, the Molten Pinnacle", "Land", b12ValakutOracle, game.CastSpellParams{})
	if !b12Card(t, g, id).Tapped {
		t.Fatal("Valakut enters tapped")
	}
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("a tapped entry is a replacement, not a tap: %d tap events", n)
	}
	b08Untap(g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want [R]", got)
	}
}

func TestB12ValakutBoltsOnTheSixthMountain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Valakut, the Molten Pinnacle", "Land", b12ValakutOracle, 0, 0)
	for i := 0; i < 4; i++ {
		seedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	}
	before := opp.Life

	// Four others: no trigger at all.
	b12PlayFromHand(t, g, "Mountain", "Basic Land — Mountain", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("four other Mountains must not trigger Valakut")
	}

	// Five others: the prompt, the target, the damage.
	b12PlayFromHand(t, g, "Mountain", "Basic Land — Mountain", "", game.CastSpellParams{})
	answerLatestTriggerPrompt(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("opponent %d → %d, want -3", before, opp.Life)
	}
	// The subtype is what matters: a Forest is not a Mountain.
	prompts := b06TriggerPromptsFor(g, me.ID)
	b12PlayFromHand(t, g, "Dryad Arbor", "Land Creature — Forest Dryad", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if b06TriggerPromptsFor(g, me.ID) != prompts {
		t.Error("a Forest is not a Mountain")
	}
}

func TestB12ValakutRechecksTheFiveMountainsOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Valakut, the Molten Pinnacle", "Land", b12ValakutOracle, 0, 0)
	var mountains []uuid.UUID
	for i := 0; i < 5; i++ {
		mountains = append(mountains, seedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain"))
	}
	before := opp.Life
	b12PlayFromHand(t, g, "Mountain", "Basic Land — Mountain", "", game.CastSpellParams{})
	answerLatestTriggerPrompt(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	// In response, one of the other Mountains leaves.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mountains[0]) })
	passPriorityAroundTable(t, g)
	if opp.Life != before {
		t.Errorf("with only four other Mountains at resolution nothing happens; opponent %d → %d", before, opp.Life)
	}
}

// --- Krark-Clan Ironworks / Skirk Prospector -----------------------

func TestB12KrarkClanIronworksEatsArtifactsForTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kci := b12Push(g, me.ID, "Krark-Clan Ironworks", "Artifact", b12KrarkClanIronworksOracle, 0, 0)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	if err := g.ActivateManaAbility(me.ID, kci, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a creature that is not an artifact is not a legal sacrifice")
	}
	if err := g.ActivateManaAbility(me.ID, kci, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 2 || got[0] != "C" || got[1] != "C" {
		t.Errorf("pool %v, want [C C]", got)
	}
	if b12ZoneOf(g, rock) != game.ZoneGraveyard {
		t.Error("the sacrificed artifact should be in the graveyard")
	}
	// The Ironworks is an artifact, so it can eat itself.
	if err := g.ActivateManaAbility(me.ID, kci, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{kci}}); err != nil {
		t.Fatalf("self-sacrifice: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 4 {
		t.Errorf("pool %v, want four colourless", got)
	}
	if g.Battlefield.Contains(kci) {
		t.Error("the Ironworks should have sacrificed itself")
	}
}

func TestB12SkirkProspectorEatsGoblinsForRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	prospector := b12Push(g, me.ID, "Skirk Prospector", "Creature — Goblin", b12SkirkProspectorOracle, 1, 1)
	goblin := b12Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	if err := g.ActivateManaAbility(me.ID, prospector, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a Bear is not a Goblin")
	}
	if err := g.ActivateManaAbility(me.ID, prospector, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{goblin}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want [R]", got)
	}
	if err := g.ActivateManaAbility(me.ID, prospector, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{prospector}}); err != nil {
		t.Fatalf("self-sacrifice: %v", err)
	}
	if g.Battlefield.Contains(prospector) {
		t.Error("the Prospector is a Goblin and can eat itself")
	}
}

// --- Colossal Majesty ----------------------------------------------

func TestB12ColossalMajestyDrawsWithAFourPowerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Colossal Majesty", "Enchantment", b12ColossalMajestyOracle, 0, 0)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	before := me.Hand.Size()
	b12ToMyNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	// The draw step adds one; the Majesty must add nothing.
	if me.Hand.Size() != before {
		t.Fatalf("no power-4 creature: hand %d → %d during the upkeep", before, me.Hand.Size())
	}

	big := b12Creature(g, me.ID, "Big", "Creature — Beast", 4, 4)
	b12ToMyNextUpkeep(t, g)
	before = me.Hand.Size()
	if len(g.PendingTriggers) == 0 && triggerOnStack(g, findBattlefieldByName(g, "Colossal Majesty")) == nil {
		t.Fatal("the upkeep trigger should be waiting")
	}
	// Removed in response: the intervening-if fails on resolution.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(big) })
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before {
		t.Errorf("the power-4 creature left in response; hand %d → %d, want unchanged", before, me.Hand.Size())
	}

	b12Creature(g, me.ID, "Bigger", "Creature — Beast", 5, 5)
	b12ToMyNextUpkeep(t, g)
	before = me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("with a power-5 creature: hand %d → %d, want +1", before, me.Hand.Size())
	}
}

// --- Persist -------------------------------------------------------

func TestB12PersistReturnsANonlegendaryCreatureWithAMinusCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	legend := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: legend, Name: "Dead Legend", TypeLine: "Legendary Creature — Human",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})

	if err := b09TryCast(t, g, "Persist", "Sorcery", b12PersistOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: legend}}); err == nil {
		t.Fatal("a legendary creature card is not a legal target")
	}
	castCatalogSpell(t, g, "Persist", "Sorcery", b12PersistOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: dead}})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the Bear should be back on the battlefield")
	}
	if got := b12Counter(t, g, dead, "-1/-1"); got != 1 {
		t.Errorf("-1/-1 counters %d, want 1", got)
	}
	if p := b12Card(t, g, dead).CurrentPower(); p != 1 {
		t.Errorf("a 2/2 with a -1/-1 counter has power %d, want 1", p)
	}
	if c := b12Card(t, g, dead); c.Controller != me.ID {
		t.Error("the creature returns under its owner's control")
	}
}

// --- Pristine Talisman ---------------------------------------------

func TestB12PristineTalismanAddsColorlessAndALife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := b12Push(g, me.ID, "Pristine Talisman", "Artifact", b12PristineTalismanOracle, 0, 0)
	before := me.Life
	if err := g.ActivateManaAbility(me.ID, rock, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want +1", before, me.Life)
	}
}

// --- Soul-Guide Lantern --------------------------------------------

func TestB12SoulGuideLanternExilesATargetCardOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushGraveyardCardForTest(opp, "Their Dead")
	mine := pushGraveyardCardForTest(me, "My Dead")

	castCatalogSpell(t, g, "Soul-Guide Lantern", "Artifact", b12SoulGuideLanternOracle, nil)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, theirs) != game.ZoneExile {
		t.Error("the targeted card should be exiled")
	}
	if b12ZoneOf(g, mine) != game.ZoneGraveyard {
		t.Error("the other card stays")
	}
}

func TestB12SoulGuideLanternSweepsOpponentsGraveyardsOrDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	lantern := b12Push(g, me.ID, "Soul-Guide Lantern", "Artifact", b12SoulGuideLanternOracle, 0, 0)
	theirs := pushGraveyardCardForTest(opp, "Their Dead")
	others := pushGraveyardCardForTest(other, "Other Dead")
	mine := pushGraveyardCardForTest(me, "My Dead")

	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, lantern, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, theirs) != game.ZoneExile || b12ZoneOf(g, others) != game.ZoneExile {
		t.Error("every opponent's graveyard should be exiled")
	}
	if b12ZoneOf(g, mine) != game.ZoneGraveyard {
		t.Error("your own graveyard is untouched")
	}
	if b12ZoneOf(g, lantern) != game.ZoneGraveyard {
		t.Error("the Lantern was sacrificed")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	lantern2 := b12Push(g2, me2.ID, "Soul-Guide Lantern", "Artifact", b12SoulGuideLanternOracle, 0, 0)
	advanceToMain(t, g2)
	me2.ManaPool.AddMana(game.ManaToken{Color: "C"})
	before := me2.Hand.Size()
	if err := g2.ActivateCatalogAbility(me2.ID, lantern2, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if me2.Hand.Size() != before+1 {
		t.Errorf("hand %d → %d, want +1", before, me2.Hand.Size())
	}
	if g2.Battlefield.Contains(lantern2) {
		t.Error("the Lantern was sacrificed")
	}
}

// --- Third Path Iconoclast / Sigil of the Empty Throne --------------

func TestB12ThirdPathIconoclastMakesArtifactSoldiersForNoncreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Third Path Iconoclast", "Creature — Human Monk", b12ThirdPathIconoclastOracle, 2, 1)
	b12Bolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	soldiers := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Soldier" && c.Controller == me.ID {
			soldiers++
			if !c.IsArtifact() || !c.IsCreature() || !c.IsColorless() {
				t.Errorf("the Soldier is a colorless artifact creature, got %q colours %v", c.TypeLine, c.Colors)
			}
		}
	}
	if soldiers != 1 {
		t.Fatalf("a Bolt should make one Soldier, made %d", soldiers)
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Soldier"); n != 1 {
		t.Errorf("a creature spell must not trigger it: %d Soldiers", n)
	}
}

func TestB12SigilOfTheEmptyThroneMakesAnAngelPerEnchantmentCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Sigil of the Empty Throne", "Enchantment", b12SigilOfTheEmptyThroneOrcl, 0, 0)
	castCatalogSpell(t, g, "Some Enchantment", "Enchantment", "", nil)
	passPriorityAroundTable(t, g)
	angel := findBattlefieldByName(g, "Angel")
	if angel == uuid.Nil {
		t.Fatal("an enchantment cast should make an Angel")
	}
	if p := effectivePower(t, g, angel); p != 4 {
		t.Errorf("Angel power %d, want 4", p)
	}
	if !containsString(effectiveAbilities(t, g, angel), "flying") {
		t.Error("the Angel has flying")
	}
	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Angel"); n != 1 {
		t.Errorf("an artifact must not trigger it: %d Angels", n)
	}
}

// --- Crime Novelist ------------------------------------------------

func TestB12CrimeNovelistGrowsAndAddsRedWhenAnArtifactIsSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	novelist := b12Push(g, me.ID, "Crime Novelist", "Creature — Goblin Bard", b12CrimeNovelistOracle, 1, 3)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1) })
	treasure := findBattlefieldByName(g, "Treasure")
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, treasure, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("crack the Treasure: %v", err)
	}
	b10ResolveAllManaPicks(t, g, me.ID, "G")
	if triggerOnStack(g, novelist) == nil && len(g.PendingTriggers) == 0 {
		t.Fatal("sacrificing the Treasure should trigger the Novelist")
	}
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, novelist, "+1/+1"); got != 1 {
		t.Errorf("+1/+1 counters %d, want 1", got)
	}
	red := 0
	for _, c := range batch01PoolColors(me) {
		if c == "R" {
			red++
		}
	}
	if red != 1 {
		t.Errorf("pool %v, want one {R} from the Novelist", batch01PoolColors(me))
	}
}

// --- Bonehoard Dracosaur -------------------------------------------

func TestB12BonehoardDracosaurExilesTwoAndPaysForEachKind(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Bonehoard Dracosaur", "Creature — Dinosaur Dragon", b12BonehoardDracosaurOracle, 5, 5)
	spell := b10LibraryTop(me, "A Spell", "Sorcery", "{R}", 0, 0)
	land := b10LibraryTop(me, "A Mountain", "Basic Land — Mountain", "", 0, 0)

	b12ToMyNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{spell, land} {
		if !b12ExiledPlayableBy(t, g, id, me.ID) {
			t.Errorf("%s should be exiled with a play grant", id)
		}
	}
	if n := countBattlefieldNamed(g, me.ID, "Dinosaur"); n != 1 {
		t.Errorf("a land exiled: %d Dinosaurs, want 1", n)
	}
	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("a nonland exiled: %d Treasures, want 1", n)
	}

	// Two lands: one Dinosaur, no Treasure. Past this turn's draw step
	// first, so the draw takes a filler rather than a seeded land.
	advanceTo(t, g, game.StepDraw)
	b10LibraryTop(me, "Mountain A", "Basic Land — Mountain", "", 0, 0)
	b10LibraryTop(me, "Mountain B", "Basic Land — Mountain", "", 0, 0)
	b12ToMyNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Dinosaur"); n != 2 {
		t.Errorf("two lands make one more Dinosaur: %d, want 2", n)
	}
	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("two lands make no Treasure: %d, want 1", n)
	}
}

// --- Archangel of Thune --------------------------------------------

func TestB12ArchangelOfThuneGrowsTheTeamOnLifegain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := b12Push(g, me.ID, "Archangel of Thune", "Creature — Angel", b12ArchangelOfThuneOracle, 3, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{angel, bear} {
		if got := b12Counter(t, g, id, "+1/+1"); got != 1 {
			t.Errorf("one lifegain event: %d counters, want 1", got)
		}
	}
	if got := b12Counter(t, g, theirs, "+1/+1"); got != 0 {
		t.Errorf("an opponent's creature grew: %d", got)
	}
	// Losing life is not gaining it.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -2) })
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, bear, "+1/+1"); got != 1 {
		t.Errorf("life loss triggered the Archangel: %d counters", got)
	}
}

// --- Skull Prophet -------------------------------------------------

func TestB12SkullProphetTapsForManaOrMillsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	prophet := b12Push(g, me.ID, "Skull Prophet", "Creature — Human Druid", b12SkullProphetOracle, 3, 1)
	advanceToMain(t, g)
	before := me.Graveyard.Size()
	lib := me.Library.Size()
	if err := g.ActivateCatalogAbility(me.ID, prophet, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("mill: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Graveyard.Size() != before+2 || me.Library.Size() != lib-2 {
		t.Errorf("graveyard %d → %d, library %d → %d; want two milled", before, me.Graveyard.Size(), lib, me.Library.Size())
	}
	if err := g.ActivateManaAbility(me.ID, prophet, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("a tapped Prophet cannot tap for mana")
	}
	b08Untap(g, prophet)
	if err := g.ActivateManaAbility(me.ID, prophet, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if n := b10ResolveAllManaPicks(t, g, me.ID, "B"); n != 1 {
		t.Fatalf("expected one colour pick, answered %d", n)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("pool %v, want [B]", got)
	}
}

// --- Martial Coup --------------------------------------------------

func TestB12MartialCoupWipesOnlyAtFive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	b12PlayFromHand(t, g, "Martial Coup", "Sorcery", b12MartialCoupOracle, game.CastSpellParams{XValue: 2})
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Soldier"); n != 2 {
		t.Fatalf("X=2: %d Soldiers, want 2", n)
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(theirs) {
		t.Fatal("X=2 destroys nothing")
	}

	b12PlayFromHand(t, g, "Martial Coup", "Sorcery", b12MartialCoupOracle, game.CastSpellParams{XValue: 5})
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Soldier"); n != 5 {
		t.Errorf("X=5: %d Soldiers survive, want the five new ones", n)
	}
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("X=5 destroys every other creature, yours included")
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Soldier" && (len(c.Colors) != 1 || c.Colors[0] != "W") {
			t.Errorf("the Soldiers are white, got %v", c.Colors)
		}
	}
}

// --- Terastodon ----------------------------------------------------

func TestB12TerastodonTradesNoncreaturePermanentsForElephants(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myLand := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	theirRock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	theirBear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Terastodon", "Creature — Elephant", b12TerastodonOracle, nil)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if p.PickTargetMin != 0 || p.PickTargetMax != 3 {
		t.Fatalf("up to three: min %d max %d", p.PickTargetMin, p.PickTargetMax)
	}
	if hasID(p.PickTargetCards, theirBear) {
		t.Error("a creature is not a legal target")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: myLand},
		{Kind: game.TargetCard, ID: theirRock},
	}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(myLand) || g.Battlefield.Contains(theirRock) {
		t.Error("both targets should be destroyed")
	}
	if n := countBattlefieldNamed(g, me.ID, "Elephant"); n != 1 {
		t.Errorf("my land's controller (me) gets one Elephant, got %d", n)
	}
	if n := countBattlefieldNamed(g, opp.ID, "Elephant"); n != 1 {
		t.Errorf("their Rock's controller gets one Elephant, got %d", n)
	}
	if g.Battlefield.Contains(theirBear) == false {
		t.Error("the creature was never a target")
	}
}

func TestB12TerastodonPaysNothingForAnIndestructiblePermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forge := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Darksteel Rock", TypeLine: "Artifact",
		Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID,
	})
	castCatalogSpell(t, g, "Terastodon", "Creature — Elephant", b12TerastodonOracle, nil)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, forge)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(forge) {
		t.Fatal("an indestructible artifact survives")
	}
	if n := countBattlefieldNamed(g, opp.ID, "Elephant"); n != 0 {
		t.Errorf("nothing reached a graveyard, so no Elephant; got %d", n)
	}
}

func TestB12TerastodonMayChooseNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	myLand := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	castCatalogSpell(t, g, "Terastodon", "Creature — Elephant", b12TerastodonOracle, nil)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
		t.Fatalf("declining with zero targets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(myLand) {
		t.Error("nothing chosen, nothing destroyed")
	}
	if n := countBattlefieldNamed(g, me.ID, "Elephant"); n != 0 {
		t.Errorf("no Elephants for nothing, got %d", n)
	}
}

// --- Lifecrafter's Bestiary ----------------------------------------

func TestB12LifecraftersBestiaryScriesOnUpkeepAndSellsADrawPerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Lifecrafter's Bestiary", "Artifact", b12LifecraftersBestiaryOracle, 0, 0)

	b12ToMyNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	var scry *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceScry && c.Chooser == me.ID {
			scry = c
		}
	}
	if scry == nil || len(scry.ScryCards) != 1 {
		t.Fatalf("the upkeep should queue a scry 1, got %+v", scry)
	}
	if err := g.ResolveScry(scry.ID, me.ID, nil, scry.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}

	advanceToMain(t, g)
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("a creature cast should ask the controller to pay {G}")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	answerPayUnless(t, g, me.ID, true)
	if me.Hand.Size() != before+1 {
		t.Errorf("paying draws one: hand %d → %d", before, me.Hand.Size())
	}

	before = me.Hand.Size()
	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("an artifact is not a creature spell")
	}
	if me.Hand.Size() != before {
		t.Errorf("hand %d → %d, want unchanged", before, me.Hand.Size())
	}
}

// --- All Will Be One -----------------------------------------------

func TestB12AllWillBeOnePingsForCountersYouPlace(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "All Will Be One", "Enchantment", b12AllWillBeOneOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirBear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 5, 5)
	before := opp.Life

	// A Duskshell Crawler's entry puts one counter on the Bear: one
	// placement, one trigger, one damage.
	castCatalogSpell(t, g, "Duskshell Crawler", "Creature — Insect", b12DuskshellCrawlerOracle, nil)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if hasID(p.PickTargetCards, bear) || !hasID(p.PickTargetCards, theirBear) {
		t.Errorf("the target is an opponent or something they control: %v", p.PickTargetCards)
	}
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("one counter: opponent %d → %d, want -1", before, opp.Life)
	}

	// Three counters at once is one trigger for three.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventResolve, Actor: me.ID})
		_ = g.AddCounterForEffect(bear, "+1/+1", 3)
	})
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, theirBear)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirBear) {
		t.Fatal("three damage does not kill a 5/5")
	}
	if got := b12Card(t, g, theirBear).DamageMarked; got != 3 {
		t.Errorf("their Bear took %d, want 3", got)
	}
}

func TestB12AllWillBeOneIgnoresOtherPlayersCountersAndRemovals(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "All Will Be One", "Enchantment", b12AllWillBeOneOracle, 0, 0)
	theirBear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	// An opponent's resolution putting counters on their own creature.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventResolve, Actor: opp.ID})
		_ = g.AddCounterForEffect(theirBear, "+1/+1", 2)
	})
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("an opponent putting counters must not trigger it")
	}
	// Removing counters is not placing them.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventResolve, Actor: me.ID})
		_ = g.AddCounterForEffect(theirBear, "+1/+1", -1)
	})
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("a removal must not trigger it")
	}
}

// --- Thousand-Year Storm -------------------------------------------

func TestB12ThousandYearStormCopiesForEachEarlierSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Thousand-Year Storm", "Enchantment", b12ThousandYearStormOracle, 0, 0)
	before := opp.Life

	// First spell of the turn: no copies.
	b12Bolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if opp.Life != before-3 {
		t.Fatalf("first Bolt: %d → %d, want -3", before, opp.Life)
	}
	// Second: one copy.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	// The trigger resolves first and offers new targets for the copy.
	for i := 0; i < 8 && latestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != before-9 {
		t.Errorf("second Bolt plus one copy: %d → %d, want -9", before, opp.Life)
	}
	// A creature spell is not counted and does not trigger.
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	// Third instant: two copies.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	for copies := 0; copies < 2; copies++ {
		for i := 0; i < 8 && latestPickTarget(g, me.ID) == nil; i++ {
			if err := g.PassPriority(); err != nil {
				t.Fatal(err)
			}
		}
		pickPlayer(t, g, me.ID, opp.ID)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-18 {
		t.Errorf("third Bolt plus two copies: %d → %d, want -18", before, opp.Life)
	}
}

// --- Time Wipe -----------------------------------------------------

func TestB12TimeWipeSavesOneCreatureAndWipesTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	saved := b12Creature(g, me.ID, "Saved", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)

	if err := b09TryCast(t, g, "Time Wipe", "Sorcery", b12TimeWipeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err == nil {
		t.Fatal("only a creature you control can be returned")
	}
	castCatalogSpell(t, g, "Time Wipe", "Sorcery", b12TimeWipeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: saved}})
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, saved) != game.ZoneHand {
		t.Error("the chosen creature returns to hand")
	}
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("every other creature is destroyed")
	}
}

// --- Cathar Commando -----------------------------------------------

func TestB12CatharCommandoSacrificesToDestroyAnArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	commando := b12Push(g, me.ID, "Cathar Commando", "Creature — Human Soldier", b12CatharCommandoOracle, 3, 1)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if !containsString(effectiveAbilities(t, g, commando), "flash") {
		t.Error("printed flash did not reach the effective abilities")
	}
	advanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, commando, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("a creature is not an artifact or enchantment")
	}
	if err := g.ActivateCatalogAbility(me.ID, commando, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if g.Battlefield.Contains(commando) {
		t.Fatal("the Commando is sacrificed at announce")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact should be destroyed")
	}
}

// --- Altar of the Brood --------------------------------------------

func TestB12AltarOfTheBroodMillsEachOpponentPerPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Altar of the Brood", "Artifact", b12AltarOfTheBroodOracle, 0, 0)
	before := make([]int, 0, 3)
	for _, p := range g.Seats[1:] {
		before = append(before, p.Graveyard.Size())
	}
	mine := me.Graveyard.Size()
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats[1:] {
		if p.Graveyard.Size() != before[i]+1 {
			t.Errorf("opponent %d: graveyard %d → %d, want +1", i+1, before[i], p.Graveyard.Size())
		}
	}
	if me.Graveyard.Size() != mine {
		t.Error("the controller does not mill")
	}
	// Two tokens at once: two triggers, two cards each.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 2) })
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats[1:] {
		if p.Graveyard.Size() != before[i]+3 {
			t.Errorf("opponent %d after two tokens: graveyard %d, want %d", i+1, p.Graveyard.Size(), before[i]+3)
		}
	}
}

// --- Duskshell Crawler ---------------------------------------------

func TestB12DuskshellCrawlerCountersATargetAndGrantsTrampleToCounteredCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirBear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	crawler := castCatalogSpell(t, g, "Duskshell Crawler", "Creature — Insect", b12DuskshellCrawlerOracle, nil)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, bear, "+1/+1"); got != 1 {
		t.Fatalf("the Bear has %d counters, want 1", got)
	}
	if !containsString(effectiveAbilities(t, g, bear), "trample") {
		t.Error("a creature you control with a +1/+1 counter has trample")
	}
	if containsString(effectiveAbilities(t, g, crawler), "trample") {
		t.Error("the Crawler itself has no counter and no trample")
	}
	// An opponent's countered creature gets nothing from your Crawler.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(theirBear, "+1/+1", 1) })
	if containsString(effectiveAbilities(t, g, theirBear), "trample") {
		t.Error("the grant is for creatures you control")
	}
	// The counter comes off, the trample goes with it.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, "+1/+1", -1) })
	if containsString(effectiveAbilities(t, g, bear), "trample") {
		t.Error("no counter, no trample")
	}
}
