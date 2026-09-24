package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch40_test.go — card-level coverage for the card-coverage roadmap's
// batch 40 (#403, `edhrec_rank` 4154–4253). One test per observable
// behaviour, driven through a real cast, activation or attack rather
// than by calling primitives.

const (
	b40TurbulentWildernessOracle = "bd8adca6-4f16-45f8-994a-fe55bd573bd0"
	b40TurbulentSteppeOracle     = "db444f9d-4dde-4308-b0f2-7acfe6de871a"
	b40IceTunnelOracle           = "40c5d6fe-854a-436d-9f80-13eb5f1f8f68"
	b40GrandColiseumOracle       = "1cea9b82-d2e9-4758-8ec8-729fcf4bb7d7"
	b40PollutedMireOracle        = "9809d975-7ef8-4946-9041-607c4e954b13"
	b40RemoteIsleOracle          = "24aebda0-315f-4d2f-8bd9-00bbaf5bd76a"
	b40VisionSkeinsOracle        = "6d42dff5-97ec-4768-b112-84a3584d0b87"
	b40RevitalizeOracle          = "b1385b03-cb4b-4812-857f-7421f1df39af"
	b40MassCalcifyOracle         = "3ab3996f-aa3f-4041-8634-8e197d51f108"
	b40ShootTheSheriffOracle     = "2038138e-2e52-41e4-90fc-443fa054c5c0"
	b40LeapOracle                = "e89a3ce0-6b38-4326-a7af-8575af371baa"
	b40ManakinOracle             = "d2343af0-468b-42bc-8a0c-347c10f7e2f3"
	b40PatronOfTheArtsOracle     = "47f8df80-abc9-4d46-8078-8fbc23430259"
	b40DrogskolReaverOracle      = "68175b49-f1e6-4b34-aad9-20d61a43d427"
	b40ZhurTaaDruidOracle        = "5979310e-38c8-489f-ab9e-2723af98a3a5"
	b40ThreeTreeMascotOracle     = "6cb9d153-4eaa-4f32-8930-bd68cd99c128"
	b40JacesSanctumOracle        = "22d7048e-d538-4fd3-9fb3-2e99e74875fe"
	b40MechanizedWarfareOracle   = "c912a43b-8994-434e-84c0-f4cf58abbd42"
	b40ShivanDevastatorOracle    = "b7daa74c-6142-4107-9355-be98af6ccf13"
	b40BygoneColossusOracle      = "1bd584d5-4e11-428c-b51e-462e4292b07f"
	b40VraanOracle               = "b2f2645f-5f74-456a-bd02-83169d8b8a7e"
	b40HerdBalothOracle          = "03873314-6e64-43ab-95c0-3d8692a57a03"
	b40GixianPuppeteerOracle     = "9d6a9a37-a245-4503-baeb-9488553798ab"
	b40VeinwitchCovenOracle      = "ab478ac9-af59-4df1-afef-5e9806c06643"
	b40GargosOracle              = "378fcd0f-1096-4769-9149-55b4a889ff56"
	b40VincentOracle             = "b8961205-87fa-4cce-aba9-84fa2f91a67f"
	b40DazzlingDenialOracle      = "6d56bd32-f47a-4e54-a88e-b69d16283ea4"
	b40OliphauntOracle           = "186b2256-4af3-48cb-96b0-b0e80a7ee6dc"
	b40MurderousRedcapOracle     = "a498bc70-36e7-4454-bc44-906893df38b8"
	b40AngelOfInventionOracle    = "28c7f2b6-ba67-4ccf-9e1b-99c89a9d1f72"
	b40DragToTheRootsOracle      = "ef4c478c-7019-4ec0-8edf-2a3078a8e97a"
	b40DecoctionModuleOracle     = "1daba4f6-2a7d-426a-97d1-0298a7100c45"
	b40LumberingWorldwagonOracle = "7e60a641-1f2d-45ff-b3f2-d389be938b22"
)

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. The two
// Turbulent lands and the two cycling lands are near-identical rows,
// so a transposed pair is invisible until someone plays that exact
// land.
func TestBatch40CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b40TurbulentWildernessOracle: "Turbulent Wilderness",
		b40TurbulentSteppeOracle:     "Turbulent Steppe",
		b40IceTunnelOracle:           "Ice Tunnel",
		b40GrandColiseumOracle:       "Grand Coliseum",
		b40PollutedMireOracle:        "Polluted Mire",
		b40RemoteIsleOracle:          "Remote Isle",
		b40VisionSkeinsOracle:        "Vision Skeins",
		b40RevitalizeOracle:          "Revitalize",
		b40MassCalcifyOracle:         "Mass Calcify",
		b40ShootTheSheriffOracle:     "Shoot the Sheriff",
		b40LeapOracle:                "Leap",
		b40ManakinOracle:             "Manakin",
		b40PatronOfTheArtsOracle:     "Patron of the Arts",
		b40DrogskolReaverOracle:      "Drogskol Reaver",
		b40ZhurTaaDruidOracle:        "Zhur-Taa Druid",
		b40ThreeTreeMascotOracle:     "Three Tree Mascot",
		b40JacesSanctumOracle:        "Jace's Sanctum",
		b40MechanizedWarfareOracle:   "Mechanized Warfare",
		b40ShivanDevastatorOracle:    "Shivan Devastator",
		b40BygoneColossusOracle:      "Bygone Colossus",
		b40VraanOracle:               "Vraan, Executioner Thane",
		b40HerdBalothOracle:          "Herd Baloth",
		b40GixianPuppeteerOracle:     "Gixian Puppeteer",
		b40VeinwitchCovenOracle:      "Veinwitch Coven",
		b40GargosOracle:              "Gargos, Vicious Watcher",
		b40VincentOracle:             "Vincent, Vengeful Atoner",
		b40DazzlingDenialOracle:      "Dazzling Denial",
		b40OliphauntOracle:           "Oliphaunt",
		b40MurderousRedcapOracle:     "Murderous Redcap",
		b40AngelOfInventionOracle:    "Angel of Invention",
		b40DragToTheRootsOracle:      "Drag to the Roots",
		b40DecoctionModuleOracle:     "Decoction Module",
		b40LumberingWorldwagonOracle: "Lumbering Worldwagon",
	}
	if len(want) != 33 {
		t.Fatalf("the batch registers 33 cards, the table lists %d", len(want))
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

// --- lands ---------------------------------------------------------

func TestB40TurbulentDualsEnterUntappedOnlyAgainstEightOpposingLands(t *testing.T) {
	for _, row := range []struct {
		name, typeLine, oracle, colour string
	}{
		{"Turbulent Wilderness", "Land — Forest Island", b40TurbulentWildernessOracle, "G"},
		{"Turbulent Steppe", "Land — Mountain Plains", b40TurbulentSteppeOracle, "R"},
	} {
		g := newCatalogGame(t)
		me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
		b36Lands(g, opp.ID, 4)
		b36Lands(g, other.ID, 3)
		// Your own lands never count towards the eight.
		b36Lands(g, me.ID, 6)
		tapped := b12PlayFromHand(t, g, row.name, row.typeLine, row.oracle, game.CastSpellParams{})
		if !b16Tapped(t, g, tapped) {
			t.Fatalf("%s: seven opposing lands is not eight — it enters tapped", row.name)
		}
		b36Lands(g, other.ID, 1)
		advanceToMainOf(t, g, 0)
		untapped := b12PlayFromHand(t, g, row.name, row.typeLine, row.oracle, game.CastSpellParams{})
		if b16Tapped(t, g, untapped) {
			t.Fatalf("%s: eight opposing lands between two opponents — it enters untapped", row.name)
		}
		b28TapForMana(t, g, me.ID, untapped, row.colour)
		if got := poolColors(me); len(got) != 1 || got[0] != row.colour {
			t.Errorf("%s: tapped for {%s}: pool %v", row.name, row.colour, got)
		}
	}
}

func TestB40IceTunnelEntersTappedAndTapsForEitherColour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := b12PlayFromHand(t, g, "Ice Tunnel", "Snow Land — Island Swamp", b40IceTunnelOracle, game.CastSpellParams{})
	if !b16Tapped(t, g, land) {
		t.Fatal("the snow dual enters tapped")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	b28TapForMana(t, g, me.ID, land, "B")
	if got := poolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("tapped for {B}: pool %v", got)
	}
}

// The colourless half of Grand Coliseum is free and the coloured half
// costs a point — folding the pain into one ability would tax both.
func TestB40GrandColiseumPainsOnlyTheColouredAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := b12PlayFromHand(t, g, "Grand Coliseum", "Land", b40GrandColiseumOracle, game.CastSpellParams{})
	if !b16Tapped(t, g, land) {
		t.Fatal("the Coliseum enters tapped")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	before := me.Life
	b28TapForMana(t, g, me.ID, land, "C")
	if me.Life != before {
		t.Errorf("tapping for {C} costs no life: %d → %d", before, me.Life)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "C" {
		t.Fatalf("tapped for {C}: pool %v", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	before = me.Life
	b40TapForMana(t, g, me.ID, land, 1, "R")
	if me.Life != before-1 {
		t.Errorf("tapping for a colour deals 1 to you: %d → %d", before, me.Life)
	}
}

func TestB40CyclingLandsEnterTappedAndTapForTheirColour(t *testing.T) {
	for _, row := range []struct{ name, oracle, colour string }{
		{"Polluted Mire", b40PollutedMireOracle, "B"},
		{"Remote Isle", b40RemoteIsleOracle, "U"},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		land := b12PlayFromHand(t, g, row.name, "Land", row.oracle, game.CastSpellParams{})
		if !b16Tapped(t, g, land) {
			t.Fatalf("%s enters tapped", row.name)
		}
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
		b28TapForMana(t, g, me.ID, land, row.colour)
		if got := poolColors(me); len(got) != 1 || got[0] != row.colour {
			t.Errorf("%s: tapped for {%s}: pool %v", row.name, row.colour, got)
		}
		// #1412: cycling is real (#660 shipped it), so the card ships
		// complete rather than caveated.
		spec, _ := Lookup(row.oracle)
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s: completeness %v, want CompletenessFull now that cycling is registered", row.name, spec.Completeness)
		}
	}
}

// --- spells --------------------------------------------------------

func TestB40VisionSkeinsDrawsTwoForEveryone(t *testing.T) {
	g := newCatalogGame(t)
	before := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		before[i] = p.Hand.Size()
	}
	// castCatalogSpell seeds the Skeins into hand and casts it in the
	// same breath, so the caster's hand is net unchanged by the cast
	// and every seat — the caster included — is up exactly two.
	castCatalogSpell(t, g, "Vision Skeins", "Instant", b40VisionSkeinsOracle, nil)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Hand.Size() != before[i]+2 {
			t.Errorf("seat %d draws two: hand %d → %d", i, before[i], p.Hand.Size())
		}
	}
}

func TestB40RevitalizeGainsThreeThenDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life, hand := me.Life, me.Hand.Size()
	castCatalogSpell(t, g, "Revitalize", "Instant", b40RevitalizeOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("you gain 3: %d → %d", life, me.Life)
	}
	// The Revitalize was seeded and cast in one breath, so the hand is
	// up exactly the one card it drew.
	if me.Hand.Size() != hand+1 {
		t.Errorf("and draw one: hand %d → %d", hand, me.Hand.Size())
	}
}

func TestB40MassCalcifyDestroysOnlyNonwhiteCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	white := b16Creature(g, me.ID, "White Knight", "Creature — Human Knight", 2, 2, "W")
	black := b16Creature(g, opp.ID, "Black Knight", "Creature — Human Knight", 2, 2, "B")
	colorless := b16Creature(g, opp.ID, "Ornithopter", "Artifact Creature — Thopter", 0, 2)
	land := b12Permanent(g, opp.ID, "Wastes", "Basic Land")
	castCatalogSpell(t, g, "Mass Calcify", "Sorcery", b40MassCalcifyOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(white) {
		t.Error("a white creature survives")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("a land is not a creature and survives")
	}
	for _, dead := range []uuid.UUID{black, colorless} {
		if g.Battlefield.Contains(dead) {
			t.Error("every nonwhite creature — colourless included — is destroyed")
		}
	}
}

func TestB40ShootTheSheriffSparesOutlaws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rogue := b12Creature(g, opp.ID, "Thieving Rogue", "Creature — Human Rogue", 2, 2)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Shoot the Sheriff", TypeLine: "Instant",
		OracleID: b40ShootTheSheriffOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(rogue)}); err == nil {
		t.Fatal("a Rogue is an outlaw and is not a legal target")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("a Bear is fair game: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear is destroyed")
	}
	if !g.Battlefield.Contains(rogue) {
		t.Error("the Rogue is untouched")
	}
}

func TestB40LeapGrantsFlyingAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Leap", "Instant", b40LeapOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the Bear has flying until end of turn")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("and draw one: hand %d → %d", hand, me.Hand.Size())
	}
}

func TestB40DragToTheRootsDestroysANonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Sol Ring", "Artifact")
	castCatalogSpell(t, g, "Drag to the Roots", "Instant", b40DragToTheRootsOracle, cardRefs(rock))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact is destroyed")
	}
}

func TestB40DazzlingDenialTaxesFourOnlyWithABird(t *testing.T) {
	for _, row := range []struct {
		name string
		bird bool
		cost string
	}{
		{"no Bird", false, "{2}"},
		{"with a Bird", true, "{4}"},
	} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		if row.bird {
			b12Creature(g, me.ID, "Storm Crow", "Creature — Bird", 1, 2)
		}
		advanceToMain(t, g)
		// An instant, because it is seat 0's turn and an opponent
		// cannot cast a sorcery-speed spell during it.
		victim := uuid.New()
		opp.Hand.PushTop(game.Card{
			InstanceID: victim, Name: "Shock", TypeLine: "Instant",
			Owner: opp.ID, Controller: opp.ID,
		})
		if err := g.CastSpell(opp.ID, victim, game.CastSpellParams{}); err != nil {
			t.Fatalf("%s: opponent casts an instant: %v", row.name, err)
		}
		castCatalogSpell(t, g, "Dazzling Denial", "Instant", b40DazzlingDenialOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
		passPriorityAroundTable(t, g)
		ask := latestChoiceOfKind(g, game.PendingChoicePayUnless)
		if ask == nil {
			t.Fatalf("%s: the spell's controller is asked to pay", row.name)
		}
		if ask.Chooser != opp.ID {
			t.Errorf("%s: the prompt goes to the spell's controller, not the counterspell's", row.name)
		}
		if ask.PayCost != row.cost {
			t.Errorf("%s: tax is %s, want %s", row.name, ask.PayCost, row.cost)
		}
	}
}

// --- permanents with triggers --------------------------------------

func TestB40PatronOfTheArtsMakesATreasureOnEntryAndOnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Patron of the Arts", "Creature — Dragon Noble", b40PatronOfTheArtsOracle, nil)
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Treasure"); n != 1 {
		t.Fatalf("one Treasure on entry, got %d", n)
	}
	patron := pushCatalogPermanent(g, me.ID, "Patron of the Arts", "Creature — Dragon Noble", b40PatronOfTheArtsOracle, false)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(patron) })
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Treasure"); n != 2 {
		t.Errorf("a second Treasure when it dies, got %d", n)
	}
}

func TestB40DrogskolReaverDrawsOnEachLifegain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Drogskol Reaver", "Creature — Spirit", b40DrogskolReaverOracle, false)
	spec, _ := Lookup(b40DrogskolReaverOracle)
	for _, kw := range []string{"flying", "double strike", "lifelink"} {
		if !hasAbility(spec.PrintedKeywords, kw) {
			t.Fatalf("the Reaver prints %s", kw)
		}
	}
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Revitalize", "Instant", b40RevitalizeOracle, nil)
	passPriorityAroundTable(t, g)
	// One card from Revitalize's own draw, one from the Reaver's
	// trigger on the three life it gained.
	if me.Hand.Size() != hand+2 {
		t.Errorf("the lifegain draws a card too: hand %d → %d, want +2", hand, me.Hand.Size())
	}
}

func TestB40ZhurTaaDruidPingsEachOpponentWhenTappedForMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	druid := pushCatalogPermanent(g, me.ID, "Zhur-Taa Druid", "Creature — Human Druid", b40ZhurTaaDruidOracle, false)
	before := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		before[i] = p.Life
	}
	b28TapForMana(t, g, me.ID, druid, "G")
	if got := poolColors(me); len(got) != 1 || got[0] != "G" {
		t.Fatalf("tapped for {G}: pool %v", got)
	}
	for i, p := range g.Seats {
		want := before[i]
		if i != 0 {
			want--
		}
		if p.Life != want {
			t.Errorf("seat %d: %d → %d, want %d (you take nothing, each opponent takes 1)", i, before[i], p.Life, want)
		}
	}
}

func TestB40ThreeTreeMascotIsEveryTypeAndTapsOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mascot := pushCatalogPermanent(g, me.ID, "Three Tree Mascot", "Artifact Creature — Shapeshifter", b40ThreeTreeMascotOracle, false)
	if !hasEffectiveKeyword(t, g, mascot, game.KeywordChangeling) {
		t.Fatal("the Mascot has changeling")
	}
	if c, ok := battlefieldCard(g, mascot); !ok || !c.HasSubtype("Sliver") || !c.HasSubtype("Elf") {
		t.Error("a changeling is every creature type")
	}
	advanceToMain(t, g)
	b06AddMana(me, "C", "C")
	b40TapForMana(t, g, me.ID, mascot, 0, "U")
	if err := g.ActivateManaAbility(me.ID, mascot, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("activate only once each turn: the second activation this turn is refused")
	}
}

func TestB40HerdBalothOffersATokenWhenCountersArePut(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	baloth := pushCatalogPermanent(g, me.ID, "Herd Baloth", "Creature — Beast", b40HerdBalothOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(baloth, game.CounterPlusOne, 2) })
	passPriorityAroundTable(t, g)
	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil {
		t.Fatal("two counters in one placement is ONE trigger, and it asks")
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if n := b16CountNamed(g, "Beast"); n != 1 {
		t.Errorf("one 4/4 Beast for one placement event, got %d", n)
	}
	// Removing a counter is not a placement and must not trigger.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(baloth, game.CounterPlusOne, -1) })
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Beast"); n != 1 {
		t.Errorf("removing a counter makes nothing, got %d Beasts", n)
	}
}

func TestB40VraanDrainsOncePerTurnAndNotForHimself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Vraan, Executioner Thane", "Legendary Creature — Phyrexian Vampire", b40VraanOracle, false)
	life, oppLife := me.Life, opp.Life
	first := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(first) })
	passPriorityAroundTable(t, g)
	if me.Life != life+2 || opp.Life != oppLife-2 {
		t.Fatalf("the first death drains 2: you %d → %d, opponent %d → %d", life, me.Life, oppLife, opp.Life)
	}
	life, oppLife = me.Life, opp.Life
	second := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(second) })
	passPriorityAroundTable(t, g)
	if me.Life != life || opp.Life != oppLife {
		t.Errorf("only once each turn: the second death this turn drains nothing (you %d → %d)", life, me.Life)
	}
}

// The Puppeteer is given to seat 1 and the draws happen on seat 0's
// turn, which pins two things at once: the count starts from zero for
// a player who has not had their draw step, and "each turn" means any
// turn rather than only the controller's.
func TestB40GixianPuppeteerDrainsOnTheSecondDrawOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, owner := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, owner.ID, "Gixian Puppeteer", "Creature — Phyrexian Warlock", b40GixianPuppeteerOracle, false)
	advanceToMain(t, g)
	ownerLife, myLife := owner.Life, me.Life
	g.WithWriteLock(func() { _ = g.DrawNForEffect(owner.ID, 1) })
	passPriorityAroundTable(t, g)
	if owner.Life != ownerLife || me.Life != myLife {
		t.Fatalf("the first draw of the turn does nothing: owner %d → %d", ownerLife, owner.Life)
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(owner.ID, 1) })
	passPriorityAroundTable(t, g)
	if owner.Life != ownerLife+2 || me.Life != myLife-2 {
		t.Fatalf("the second draw drains 2: owner %d → %d, each opponent %d → %d",
			ownerLife, owner.Life, myLife, me.Life)
	}
	ownerLife, myLife = owner.Life, me.Life
	g.WithWriteLock(func() { _ = g.DrawNForEffect(owner.ID, 1) })
	passPriorityAroundTable(t, g)
	if owner.Life != ownerLife || me.Life != myLife {
		t.Errorf("the third draw does nothing: owner %d → %d", ownerLife, owner.Life)
	}
}

func TestB40VeinwitchCovenPaysBlackToReturnACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Veinwitch Coven", "Creature — Vampire Warlock", b40VeinwitchCovenOracle, false)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: dead, Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	b06AddMana(me, "B")
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
	// The target is chosen when the trigger goes on the stack
	// (CR 603.3d), before anyone is asked about the {B}.
	b40PickCard(t, g, me.ID, dead)
	passPriorityAroundTable(t, g)
	ask := latestChoiceOfKind(g, game.PendingChoicePayUnless)
	if ask == nil {
		t.Fatal("the lifegain asks whether to pay {B}")
	}
	if ask.PayCost != "{B}" {
		t.Errorf("the optional cost is {B}: %s", ask.PayCost)
	}
	if err := g.ResolvePayUnless(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if !me.Hand.Contains(dead) {
		t.Error("paying {B} returns the creature card to your hand")
	}
}

func TestB40VincentGrowsOnCombatDamageAndSpillsOverAtSeven(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	vincent := b12Push(g, me.ID, "Vincent, Vengeful Atoner", "Legendary Creature — Assassin", b40VincentOracle, 7, 7)
	oppLife, otherLife := opp.Life, other.Life
	attackWith(t, g, opp.ID, vincent)
	// Both of Vincent's abilities trigger off the same damage, so
	// their controller orders them (CR 603.3b) before either resolves.
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-7 {
		t.Fatalf("the defender takes 7 in combat: %d → %d", oppLife, opp.Life)
	}
	if other.Life != otherLife-7 {
		t.Errorf("at 7 power the Chaos ability hits each OTHER opponent for 7: %d → %d", otherLife, other.Life)
	}
	if counterOn(g, vincent, game.CounterPlusOne) != 1 {
		t.Errorf("one +1/+1 counter for the combat damage, got %d", counterOn(g, vincent, game.CounterPlusOne))
	}
}

// TestB40VincentChaosCountsCounterPumpedPower is #1281: the Chaos
// gate read src.Effective().Power, which excludes +1/+1 / -1/-1
// counters, so a Vincent pumped to 7 by counters rather than by an
// anthem never reached the threshold. A printed 6/6 with one +1/+1
// counter already on it (CurrentPower 7, Effective().Power still 6 —
// no anthem in play) deals combat damage AS a 7/7 (the counter is
// real, CR 613.4 layer 7d applies before combat) and the Chaos
// ability must still spill that 7 over.
func TestB40VincentChaosCountsCounterPumpedPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	vincent := b12Push(g, me.ID, "Vincent, Vengeful Atoner", "Legendary Creature — Assassin", b40VincentOracle, 6, 6)
	if err := g.AddCounterForEffect(vincent, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounterForEffect: %v", err)
	}
	oppLife, otherLife := opp.Life, other.Life
	attackWith(t, g, opp.ID, vincent)
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-7 {
		t.Fatalf("the defender takes 7 in combat (6 printed + the counter): %d → %d", oppLife, opp.Life)
	}
	if other.Life != otherLife-7 {
		t.Errorf("a Vincent at CURRENT power 7 (6 printed + one +1/+1 counter, no anthem) must still "+
			"spill its 7 over: %d → %d", otherLife, other.Life)
	}
}

func TestB40MurderousRedcapDealsDamageEqualToItsPower(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	me := g.Seats[0]
	before := opp.Life
	b36CastCreature(t, g, "Murderous Redcap", "Creature — Goblin Assassin", b40MurderousRedcapOracle, 2, 2)
	// The enters trigger is targeted, so the target is chosen as it
	// goes on the stack.
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("a printed 2/2 pings for 2: %d → %d", before, opp.Life)
	}
	spec, _ := Lookup(b40MurderousRedcapOracle)
	if spec.Completeness != CompletenessCaveats {
		t.Error("persist is not implemented and the card says so")
	}
}

// TestB40MurderousRedcapCountersItsDamage is #1281: the ETB read
// src.Effective().Power, which excludes +1/+1 / -1/-1 counters, so a
// Redcap pumped by a counter (rather than an anthem) dealt the wrong
// amount. A counter placed while the trigger is still on the stack —
// resolution reads power fresh, per the card's own doc comment — must
// count.
func TestB40MurderousRedcapCountersItsDamage(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	me := g.Seats[0]
	before := opp.Life
	redcap := b36CastCreature(t, g, "Murderous Redcap", "Creature — Goblin Assassin", b40MurderousRedcapOracle, 2, 2)
	b16PickPlayer(t, g, me.ID, opp.ID)
	// The trigger is on the stack, un-resolved: a +1/+1 counter lands
	// on the Redcap before it fires.
	if err := g.AddCounterForEffect(redcap, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounterForEffect: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("a 2/2 with a +1/+1 counter pings for 3: %d → %d", before, opp.Life)
	}
}

func TestB40AngelOfInventionFabricatesEitherWay(t *testing.T) {
	for _, counters := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		angel := b36CastCreature(t, g, "Angel of Invention", "Creature — Angel", b40AngelOfInventionOracle, 2, 1)
		ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
		if ask == nil {
			t.Fatal("fabricate 2 asks which mode")
		}
		if err := g.ResolveConfirm(ask.ID, me.ID, counters); err != nil {
			t.Fatalf("ResolveConfirm: %v", err)
		}
		if counters {
			if counterOn(g, angel, game.CounterPlusOne) != 2 {
				t.Errorf("the counters branch puts two +1/+1 counters on it, got %d", counterOn(g, angel, game.CounterPlusOne))
			}
			if n := b16CountNamed(g, "Servo"); n != 0 {
				t.Errorf("and makes no Servos, got %d", n)
			}
			continue
		}
		if n := b16CountNamed(g, "Servo"); n != 2 {
			t.Fatalf("the token branch makes two Servos, got %d", n)
		}
		// The anthem is "OTHER creatures you control", so a Servo is a
		// 2/2 and the Angel stays a printed 2/1.
		for _, c := range g.Battlefield.Cards {
			if c.Name != "Servo" {
				continue
			}
			if eff := c.Effective(); eff.Power != 2 || eff.Toughness != 2 {
				t.Errorf("a Servo under the Angel's anthem is 2/2, got %d/%d", eff.Power, eff.Toughness)
			}
		}
		if c := findTestCard(g, angel); c != nil {
			if eff := c.Effective(); eff.Power != 2 || eff.Toughness != 1 {
				t.Errorf("the Angel does not pump itself: %d/%d, want 2/1", eff.Power, eff.Toughness)
			}
		}
	}
}

func TestB40DecoctionModuleGivesEnergyAndBouncesYourOwn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	module := pushCatalogPermanent(g, me.ID, "Decoction Module", "Artifact", b40DecoctionModuleOracle, false)
	before := me.Energy
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if me.Energy != before+1 {
		t.Fatalf("a creature entering under your control gives you {E}: %d → %d", before, me.Energy)
	}
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	b06AddMana(me, "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, module, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs)}); err == nil {
		t.Fatal("an opponent's creature is not a legal target")
	}
	b16Activate(t, g, me.ID, module, 0, game.ActivateAbilityParams{Targets: cardRefs(mine)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || !me.Hand.Contains(mine) {
		t.Error("your own creature goes back to your hand")
	}
}

func TestB40LumberingWorldwagonIsAsBigAsYourLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b36Lands(g, me.ID, 3)
	wagon := b12Push(g, me.ID, "Lumbering Worldwagon", "Artifact — Vehicle", b40LumberingWorldwagonOracle, 0, 4)
	// A Vehicle is not a creature until it is crewed, and a
	// non-creature has no P/T for the layer engine to define — so the
	// characteristic-defining power is only observable once it is.
	crewer := pushCrewerForTest(g, me.ID, "Crewer", 4)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, wagon, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{crewer}}); err != nil {
		t.Fatalf("crew 4: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, wagon); got != 3 {
		t.Fatalf("power is the number of lands you control (3), got %d", got)
	}
	// Recomputed every pass, so two more lands is two more power with
	// no invalidation bookkeeping.
	b36Lands(g, me.ID, 2)
	if got := effectivePower(t, g, wagon); got != 5 {
		t.Errorf("two more lands, two more power: got %d, want 5", got)
	}
	// The toughness is printed, not derived.
	if got := effectiveToughness(t, g, wagon); got != 4 {
		t.Errorf("toughness stays the printed 4, got %d", got)
	}
}

func TestB40LumberingWorldwagonMayFetchABasicOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Ketria Triome", TypeLine: "Land — Forest Island Mountain"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
	)
	castCatalogSpell(t, g, "Lumbering Worldwagon", "Artifact — Vehicle", b40LumberingWorldwagonOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("entering opens the optional search")
	}
	if searchOptionNamed(g, c, "Forest") == uuid.Nil {
		t.Error("a basic land is offered")
	}
	if searchOptionNamed(g, c, "Ketria Triome") != uuid.Nil || searchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a nonbasic land and a creature are not")
	}
}

// --- statics and cost modifiers ------------------------------------

func TestB40JacesSanctumDiscountsYourInstantsAndScries(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Jace's Sanctum", "Enchantment", b40JacesSanctumOracle, false)
	if got := b40PricedFor(t, g, me.ID, "Divination", "Sorcery", "{2}{U}"); got != 2 {
		t.Errorf("your sorcery costs {1} less: mana value %d, want 2", got)
	}
	if got := b40PricedFor(t, g, opp.ID, "Divination", "Sorcery", "{2}{U}"); got != 3 {
		t.Errorf("an opponent's sorcery is untouched: mana value %d, want 3", got)
	}
	if got := b40PricedFor(t, g, me.ID, "Bear", "Creature — Bear", "{2}{G}"); got != 3 {
		t.Errorf("a creature spell is not an instant or a sorcery: mana value %d, want 3", got)
	}
	castCatalogSpell(t, g, "Divination", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if latestChoiceOfKind(g, game.PendingChoiceScry) == nil {
		t.Error("casting an instant or sorcery scries 1")
	}
}

func TestB40GargosDiscountsHydraSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Gargos, Vicious Watcher", "Legendary Creature — Hydra", b40GargosOracle, false)
	if got := b40PricedFor(t, g, me.ID, "Hooded Hydra", "Creature — Snake Hydra", "{4}{G}"); got != 1 {
		t.Errorf("your Hydra costs {4} less, off the generic: mana value %d, want 1", got)
	}
	if got := b40PricedFor(t, g, me.ID, "Bear", "Creature — Bear", "{4}{G}"); got != 5 {
		t.Errorf("a Bear is not a Hydra: mana value %d, want 5", got)
	}
	if got := b40PricedFor(t, g, opp.ID, "Hooded Hydra", "Creature — Snake Hydra", "{4}{G}"); got != 5 {
		t.Errorf("an opponent's Hydra is untouched: mana value %d, want 5", got)
	}
}

func TestB40MechanizedWarfareBoostsOnlyYourRedOrArtifactDamageAtOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Mechanized Warfare", "Enchantment", b40MechanizedWarfareOracle, false)
	red := b16Creature(g, me.ID, "Goblin", "Creature — Goblin", 2, 2, "R")
	green := b16Creature(g, me.ID, "Elf", "Creature — Elf", 2, 2, "G")
	theirRed := b16Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 2, 2, "R")

	before := opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(red, opp.ID, 2) })
	if opp.Life != before-3 {
		t.Errorf("your red source deals 2+1 to an opponent: %d → %d", before, opp.Life)
	}
	before = opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(green, opp.ID, 2) })
	if opp.Life != before-2 {
		t.Errorf("a green non-artifact source of yours is untouched: %d → %d", before, opp.Life)
	}
	before = me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(red, me.ID, 2) })
	if me.Life != before-2 {
		t.Errorf("damage to YOU is untouched: %d → %d", before, me.Life)
	}
	before = me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(theirRed, me.ID, 2) })
	if me.Life != before-2 {
		t.Errorf("an opponent's red source is untouched: %d → %d", before, me.Life)
	}
}

func TestB40ShivanDevastatorEntersWithXCounters(t *testing.T) {
	g := newCatalogGame(t)
	dragon := b12PlayFromHand(t, g, "Shivan Devastator", "Creature — Dragon Hydra", b40ShivanDevastatorOracle,
		game.CastSpellParams{XValue: 4})
	passPriorityAroundTable(t, g)
	if counterOn(g, dragon, game.CounterPlusOne) != 4 {
		t.Errorf("X = 4 puts four +1/+1 counters on it, got %d", counterOn(g, dragon, game.CounterPlusOne))
	}
	spec, _ := Lookup(b40ShivanDevastatorOracle)
	for _, kw := range []string{"flying", "haste"} {
		if !hasAbility(spec.PrintedKeywords, kw) {
			t.Errorf("the Dragon prints %s", kw)
		}
	}
}

func TestB40BygoneColossusOffersItsWarpCost(t *testing.T) {
	spec, ok := Lookup(b40BygoneColossusOracle)
	if !ok {
		t.Fatal("Bygone Colossus is registered")
	}
	if len(spec.AlternativeCosts) != 1 {
		t.Fatalf("one alternative cost, got %d", len(spec.AlternativeCosts))
	}
	warp := spec.AlternativeCosts[0]
	if warp.Key != "warp" || warp.ManaCost != "{3}" {
		t.Errorf("warp {3}: key %q cost %q", warp.Key, warp.ManaCost)
	}
	// The exile leg is what stops warp being "a 9/9 for {3}, forever".
	if !warp.WarpExile {
		t.Error("the warp cost carries its end-step exile")
	}
	// Warp is paid from hand, so no extra castable zone is declared.
	if len(spec.CastableZones) != 0 {
		t.Errorf("warp needs no CastableZones, got %v", spec.CastableZones)
	}
}

func TestB40ManakinTapsForColorless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := pushCatalogPermanent(g, me.ID, "Manakin", "Artifact Creature — Construct", b40ManakinOracle, false)
	b28TapForMana(t, g, me.ID, rock, "C")
	if got := poolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("tapped for {C}: pool %v", got)
	}
}

func TestB40OliphauntPumpsAnotherCreatureAndNotItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oliphaunt := b12Push(g, me.ID, "Oliphaunt", "Creature — Elephant", b40OliphauntOracle, 6, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, oliphaunt)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the attack trigger asks for another creature you control")
	}
	if hasID(p.PickTargetCards, oliphaunt) {
		t.Error("\"another\" excludes the Oliphaunt itself")
	}
	b40PickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, bear)
	if !ok {
		t.Fatal("the Bear is still there")
	}
	if eff := c.Effective(); eff.Power != 4 || eff.Toughness != 2 {
		t.Errorf("the Bear is 4/2 until end of turn: %d/%d", eff.Power, eff.Toughness)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("and it has trample until end of turn")
	}
}

// b40PricedFor runs one card through the cost pipeline as `caster`
// would pay it, and returns the total mana value they owe.
//
// Routed through ApplyCostModifiers rather than through a real cast,
// for the reason priceInHand is: the question is what the board
// charges, not whether the caster could satisfy timing and a pool.
func b40PricedFor(t *testing.T, g *game.Game, caster uuid.UUID, name, typeLine, manaCost string) int {
	t.Helper()
	base, err := game.ParseCost(manaCost)
	if err != nil {
		t.Fatalf("ParseCost(%q): %v", manaCost, err)
	}
	out, err := g.ApplyCostModifiers(base, game.CostQuery{
		Card: game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: manaCost,
			Owner: caster, Controller: caster,
		},
		Controller: caster,
		FromZone:   game.ZoneHand,
	})
	if err != nil {
		t.Fatalf("ApplyCostModifiers(%s): %v", name, err)
	}
	return out.ManaValue()
}

// b40PickCard answers a targeted trigger's pick_target prompt with a
// card — b16PickPlayer's sibling for the card slot.
func b40PickCard(t *testing.T, g *game.Game, chooser, card uuid.UUID) {
	t.Helper()
	p := latestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	if err := g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: card}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

// b40TapForMana activates one named mana ability by index and answers
// the colour pick it opens — b28TapForMana with the index exposed,
// which Grand Coliseum's second ability and Three Tree Mascot both
// need.
func b40TapForMana(t *testing.T, g *game.Game, controller, card uuid.UUID, idx int, color string) {
	t.Helper()
	if err := g.ActivateManaAbility(controller, card, idx, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility(%d): %v", idx, err)
	}
	if color != "" {
		b10ResolveAllManaPicks(t, g, controller, color)
	}
}

// --- #1412: cycling ---------------------------------------------------

// Polluted Mire and Remote Isle: cycling {2} from hand discards the
// land and draws a card, and is refused without the mana to pay it.
func TestB40CyclingLandsCycleForACard(t *testing.T) {
	for _, row := range []struct{ name, oracle string }{
		{"Polluted Mire", b40PollutedMireOracle},
		{"Remote Isle", b40RemoteIsleOracle},
	} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]

		refused := pushCatalogHandCard(me, row.name, "Land", row.oracle)
		if err := g.ActivateCatalogAbility(me.ID, refused, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
			t.Fatalf("%s: {2} must be paid to cycle", row.name)
		}
		if !me.Hand.Contains(refused) {
			t.Errorf("%s: discarded despite the unpaid cost", row.name)
		}
		g.WithWriteLock(func() { me.Hand.Cards = nil })

		id, _ := cycleFromHand(t, g, row.name, "Land", row.oracle, "{C}{C}")
		if !me.Graveyard.Contains(id) {
			t.Fatalf("%s: the cycled land is not in the graveyard", row.name)
		}
		before := len(me.Hand.Cards)
		passPriorityAroundTable(t, g)
		if got := len(me.Hand.Cards); got != before+1 {
			t.Errorf("%s: hand %d -> %d, want the cycling draw", row.name, before, got)
		}
	}
}

// Oliphaunt: mountaincycling {1} discards it and fetches a Mountain
// card (any land with the Mountain type, not only a basic one),
// reveals it and shuffles; refused without the mana.
func TestB40OliphauntMountaincyclesForAMountain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]

	refused := pushCatalogHandCard(me, "Oliphaunt", "Creature — Elephant", b40OliphauntOracle)
	if err := g.ActivateCatalogAbility(me.ID, refused, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("{1} must be paid to mountaincycle")
	}
	if !me.Hand.Contains(refused) {
		t.Error("discarded despite the unpaid cost")
	}
	g.WithWriteLock(func() { me.Hand.Cards = nil })

	ids := seedSearchLibrary(me,
		searchTestLand("Mountain", "Basic Land — Mountain"),
		searchTestLand("Rugged Prairie", "Land — Mountain Plains"),
		searchTestLand("Island", "Basic Land — Island"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	mountain, dual := ids[0], ids[1]

	id, _ := cycleFromHand(t, g, "Oliphaunt", "Creature — Elephant", b40OliphauntOracle, "{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the mountaincycled Oliphaunt is not in the graveyard")
	}
	handBefore := len(me.Hand.Cards)
	libraryBefore := len(me.Library.Cards)
	passPriorityAroundTable(t, g)

	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt after mountaincycling")
	}
	if len(c.SearchCards) != 2 {
		t.Fatalf("search offers %d cards, want the two Mountain-typed lands %v %v", len(c.SearchCards), mountain, dual)
	}
	for _, got := range c.SearchCards {
		if got != mountain && got != dual {
			t.Errorf("search offers %v, which is not a Mountain card", got)
		}
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{dual}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if got := len(me.Hand.Cards); got != handBefore+1 {
		t.Errorf("hand %d -> %d, want the fetched Mountain", handBefore, got)
	}
	if !me.Hand.Contains(dual) {
		t.Error("the fetched land did not reach the hand")
	}
	// "then shuffle" — CR 702.29e. One card left the library.
	if got := len(me.Library.Cards); got != libraryBefore-1 {
		t.Errorf("library %d -> %d, want one card taken", libraryBefore, got)
	}
}
