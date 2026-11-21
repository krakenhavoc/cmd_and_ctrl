package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. The two
// Turbulent lands and the two cycling lands are near-identical rows,
// so a transposed pair is invisible until someone plays that exact
// land.
func TestBatch40CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B40TurbulentWildernessOracle: "Turbulent Wilderness",
		effectstest.B40TurbulentSteppeOracle:     "Turbulent Steppe",
		effectstest.B40IceTunnelOracle:           "Ice Tunnel",
		effectstest.B40GrandColiseumOracle:       "Grand Coliseum",
		effectstest.B40PollutedMireOracle:        "Polluted Mire",
		effectstest.B40RemoteIsleOracle:          "Remote Isle",
		effectstest.B40VisionSkeinsOracle:        "Vision Skeins",
		effectstest.B40RevitalizeOracle:          "Revitalize",
		effectstest.B40MassCalcifyOracle:         "Mass Calcify",
		effectstest.B40ShootTheSheriffOracle:     "Shoot the Sheriff",
		effectstest.B40LeapOracle:                "Leap",
		effectstest.B40ManakinOracle:             "Manakin",
		effectstest.B40PatronOfTheArtsOracle:     "Patron of the Arts",
		effectstest.B40DrogskolReaverOracle:      "Drogskol Reaver",
		effectstest.B40ZhurTaaDruidOracle:        "Zhur-Taa Druid",
		effectstest.B40ThreeTreeMascotOracle:     "Three Tree Mascot",
		effectstest.B40JacesSanctumOracle:        "Jace's Sanctum",
		effectstest.B40MechanizedWarfareOracle:   "Mechanized Warfare",
		effectstest.B40ShivanDevastatorOracle:    "Shivan Devastator",
		effectstest.B40BygoneColossusOracle:      "Bygone Colossus",
		effectstest.B40VraanOracle:               "Vraan, Executioner Thane",
		effectstest.B40HerdBalothOracle:          "Herd Baloth",
		effectstest.B40GixianPuppeteerOracle:     "Gixian Puppeteer",
		effectstest.B40VeinwitchCovenOracle:      "Veinwitch Coven",
		effectstest.B40GargosOracle:              "Gargos, Vicious Watcher",
		effectstest.B40VincentOracle:             "Vincent, Vengeful Atoner",
		effectstest.B40DazzlingDenialOracle:      "Dazzling Denial",
		effectstest.B40OliphauntOracle:           "Oliphaunt",
		effectstest.B40MurderousRedcapOracle:     "Murderous Redcap",
		effectstest.B40AngelOfInventionOracle:    "Angel of Invention",
		effectstest.B40DragToTheRootsOracle:      "Drag to the Roots",
		effectstest.B40DecoctionModuleOracle:     "Decoction Module",
		effectstest.B40LumberingWorldwagonOracle: "Lumbering Worldwagon",
	}
	if len(want) != 33 {
		t.Fatalf("the batch registers 33 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
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
		{"Turbulent Wilderness", "Land — Forest Island", effectstest.B40TurbulentWildernessOracle, "G"},
		{"Turbulent Steppe", "Land — Mountain Plains", effectstest.B40TurbulentSteppeOracle, "R"},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
		effectstest.B36Lands(g, opp.ID, 4)
		effectstest.B36Lands(g, other.ID, 3)
		// Your own lands never count towards the eight.
		effectstest.B36Lands(g, me.ID, 6)
		tapped := effectstest.B12PlayFromHand(t, g, row.name, row.typeLine, row.oracle, game.CastSpellParams{})
		if !effectstest.B16Tapped(t, g, tapped) {
			t.Fatalf("%s: seven opposing lands is not eight — it enters tapped", row.name)
		}
		effectstest.B36Lands(g, other.ID, 1)
		effectstest.AdvanceToMainOf(t, g, 0)
		untapped := effectstest.B12PlayFromHand(t, g, row.name, row.typeLine, row.oracle, game.CastSpellParams{})
		if effectstest.B16Tapped(t, g, untapped) {
			t.Fatalf("%s: eight opposing lands between two opponents — it enters untapped", row.name)
		}
		effectstest.B28TapForMana(t, g, me.ID, untapped, row.colour)
		if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != row.colour {
			t.Errorf("%s: tapped for {%s}: pool %v", row.name, row.colour, got)
		}
	}
}

func TestB40IceTunnelEntersTappedAndTapsForEitherColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.B12PlayFromHand(t, g, "Ice Tunnel", "Snow Land — Island Swamp", effectstest.B40IceTunnelOracle, game.CastSpellParams{})
	if !effectstest.B16Tapped(t, g, land) {
		t.Fatal("the snow dual enters tapped")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	effectstest.B28TapForMana(t, g, me.ID, land, "B")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("tapped for {B}: pool %v", got)
	}
}

// The colourless half of Grand Coliseum is free and the coloured half
// costs a point — folding the pain into one ability would tax both.
func TestB40GrandColiseumPainsOnlyTheColouredAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.B12PlayFromHand(t, g, "Grand Coliseum", "Land", effectstest.B40GrandColiseumOracle, game.CastSpellParams{})
	if !effectstest.B16Tapped(t, g, land) {
		t.Fatal("the Coliseum enters tapped")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	before := me.Life
	effectstest.B28TapForMana(t, g, me.ID, land, "C")
	if me.Life != before {
		t.Errorf("tapping for {C} costs no life: %d → %d", before, me.Life)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Fatalf("tapped for {C}: pool %v", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	before = me.Life
	effectstest.B40TapForMana(t, g, me.ID, land, 1, "R")
	if me.Life != before-1 {
		t.Errorf("tapping for a colour deals 1 to you: %d → %d", before, me.Life)
	}
}

func TestB40CyclingLandsEnterTappedAndTapForTheirColour(t *testing.T) {
	for _, row := range []struct{ name, oracle, colour string }{
		{"Polluted Mire", effectstest.B40PollutedMireOracle, "B"},
		{"Remote Isle", effectstest.B40RemoteIsleOracle, "U"},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		land := effectstest.B12PlayFromHand(t, g, row.name, "Land", row.oracle, game.CastSpellParams{})
		if !effectstest.B16Tapped(t, g, land) {
			t.Fatalf("%s enters tapped", row.name)
		}
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
		effectstest.B28TapForMana(t, g, me.ID, land, row.colour)
		if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != row.colour {
			t.Errorf("%s: tapped for {%s}: pool %v", row.name, row.colour, got)
		}
		// #1412: cycling is real (#660 shipped it), so the card ships
		// complete rather than caveated.
		spec, _ := effects.Lookup(row.oracle)
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: completeness %v, want CompletenessFull now that cycling is registered", row.name, spec.Completeness)
		}
	}
}

// --- spells --------------------------------------------------------

func TestB40VisionSkeinsDrawsTwoForEveryone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	before := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		before[i] = p.Hand.Size()
	}
	// castCatalogSpell seeds the Skeins into hand and casts it in the
	// same breath, so the caster's hand is net unchanged by the cast
	// and every seat — the caster included — is up exactly two.
	effectstest.CastCatalogSpell(t, g, "Vision Skeins", "Instant", effectstest.B40VisionSkeinsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Hand.Size() != before[i]+2 {
			t.Errorf("seat %d draws two: hand %d → %d", i, before[i], p.Hand.Size())
		}
	}
}

func TestB40RevitalizeGainsThreeThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life, hand := me.Life, me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Revitalize", "Instant", effectstest.B40RevitalizeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	white := effectstest.B16Creature(g, me.ID, "White Knight", "Creature — Human Knight", 2, 2, "W")
	black := effectstest.B16Creature(g, opp.ID, "Black Knight", "Creature — Human Knight", 2, 2, "B")
	colorless := effectstest.B16Creature(g, opp.ID, "Ornithopter", "Artifact Creature — Thopter", 0, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Wastes", "Basic Land")
	effectstest.CastCatalogSpell(t, g, "Mass Calcify", "Sorcery", effectstest.B40MassCalcifyOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rogue := effectstest.B12Creature(g, opp.ID, "Thieving Rogue", "Creature — Human Rogue", 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Shoot the Sheriff", TypeLine: "Instant",
		OracleID: effectstest.B40ShootTheSheriffOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(rogue)}); err == nil {
		t.Fatal("a Rogue is an outlaw and is not a legal target")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("a Bear is fair game: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear is destroyed")
	}
	if !g.Battlefield.Contains(rogue) {
		t.Error("the Rogue is untouched")
	}
}

func TestB40LeapGrantsFlyingAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Leap", "Instant", effectstest.B40LeapOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the Bear has flying until end of turn")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("and draw one: hand %d → %d", hand, me.Hand.Size())
	}
}

func TestB40DragToTheRootsDestroysANonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Sol Ring", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Drag to the Roots", "Instant", effectstest.B40DragToTheRootsOracle, effectstest.CardRefs(rock))
	effectstest.PassPriorityAroundTable(t, g)
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
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		if row.bird {
			effectstest.B12Creature(g, me.ID, "Storm Crow", "Creature — Bird", 1, 2)
		}
		effectstest.AdvanceToMain(t, g)
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
		effectstest.CastCatalogSpell(t, g, "Dazzling Denial", "Instant", effectstest.B40DazzlingDenialOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
		effectstest.PassPriorityAroundTable(t, g)
		ask := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Patron of the Arts", "Creature — Dragon Noble", effectstest.B40PatronOfTheArtsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Treasure"); n != 1 {
		t.Fatalf("one Treasure on entry, got %d", n)
	}
	patron := effectstest.PushCatalogPermanent(g, me.ID, "Patron of the Arts", "Creature — Dragon Noble", effectstest.B40PatronOfTheArtsOracle, false)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(patron) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Treasure"); n != 2 {
		t.Errorf("a second Treasure when it dies, got %d", n)
	}
}

func TestB40DrogskolReaverDrawsOnEachLifegain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Drogskol Reaver", "Creature — Spirit", effectstest.B40DrogskolReaverOracle, false)
	spec, _ := effects.Lookup(effectstest.B40DrogskolReaverOracle)
	for _, kw := range []string{"flying", "double strike", "lifelink"} {
		if !effectstest.HasAbility(spec.PrintedKeywords, kw) {
			t.Fatalf("the Reaver prints %s", kw)
		}
	}
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Revitalize", "Instant", effectstest.B40RevitalizeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// One card from Revitalize's own draw, one from the Reaver's
	// trigger on the three life it gained.
	if me.Hand.Size() != hand+2 {
		t.Errorf("the lifegain draws a card too: hand %d → %d, want +2", hand, me.Hand.Size())
	}
}

func TestB40ZhurTaaDruidPingsEachOpponentWhenTappedForMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	druid := effectstest.PushCatalogPermanent(g, me.ID, "Zhur-Taa Druid", "Creature — Human Druid", effectstest.B40ZhurTaaDruidOracle, false)
	before := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		before[i] = p.Life
	}
	effectstest.B28TapForMana(t, g, me.ID, druid, "G")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Fatalf("tapped for {G}: pool %v", got)
	}
	// #4928: the damage is a trigger on the stack (CR 605.1b).
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mascot := effectstest.PushCatalogPermanent(g, me.ID, "Three Tree Mascot", "Artifact Creature — Shapeshifter", effectstest.B40ThreeTreeMascotOracle, false)
	if !effectstest.HasEffectiveKeyword(t, g, mascot, game.KeywordChangeling) {
		t.Fatal("the Mascot has changeling")
	}
	if c, ok := effectstest.BattlefieldCard(g, mascot); !ok || !c.HasSubtype("Sliver") || !c.HasSubtype("Elf") {
		t.Error("a changeling is every creature type")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B40TapForMana(t, g, me.ID, mascot, 0, "U")
	if err := g.ActivateManaAbility(me.ID, mascot, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("activate only once each turn: the second activation this turn is refused")
	}
}

func TestB40HerdBalothOffersATokenWhenCountersArePut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	baloth := effectstest.PushCatalogPermanent(g, me.ID, "Herd Baloth", "Creature — Beast", effectstest.B40HerdBalothOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(baloth, game.CounterPlusOne, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil {
		t.Fatal("two counters in one placement is ONE trigger, and it asks")
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if n := effectstest.B16CountNamed(g, "Beast"); n != 1 {
		t.Errorf("one 4/4 Beast for one placement event, got %d", n)
	}
	// Removing a counter is not a placement and must not trigger.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(baloth, game.CounterPlusOne, -1) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Beast"); n != 1 {
		t.Errorf("removing a counter makes nothing, got %d Beasts", n)
	}
}

func TestB40VraanDrainsOncePerTurnAndNotForHimself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Vraan, Executioner Thane", "Legendary Creature — Phyrexian Vampire", effectstest.B40VraanOracle, false)
	life, oppLife := me.Life, opp.Life
	first := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(first) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 || opp.Life != oppLife-2 {
		t.Fatalf("the first death drains 2: you %d → %d, opponent %d → %d", life, me.Life, oppLife, opp.Life)
	}
	life, oppLife = me.Life, opp.Life
	second := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(second) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life || opp.Life != oppLife {
		t.Errorf("only once each turn: the second death this turn drains nothing (you %d → %d)", life, me.Life)
	}
}

// The Puppeteer is given to seat 1 and the draws happen on seat 0's
// turn, which pins two things at once: the count starts from zero for
// a player who has not had their draw step, and "each turn" means any
// turn rather than only the controller's.
func TestB40GixianPuppeteerDrainsOnTheSecondDrawOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, owner := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, owner.ID, "Gixian Puppeteer", "Creature — Phyrexian Warlock", effectstest.B40GixianPuppeteerOracle, false)
	effectstest.AdvanceToMain(t, g)
	ownerLife, myLife := owner.Life, me.Life
	g.WithWriteLock(func() { _ = g.DrawNForEffect(owner.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if owner.Life != ownerLife || me.Life != myLife {
		t.Fatalf("the first draw of the turn does nothing: owner %d → %d", ownerLife, owner.Life)
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(owner.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if owner.Life != ownerLife+2 || me.Life != myLife-2 {
		t.Fatalf("the second draw drains 2: owner %d → %d, each opponent %d → %d",
			ownerLife, owner.Life, myLife, me.Life)
	}
	ownerLife, myLife = owner.Life, me.Life
	g.WithWriteLock(func() { _ = g.DrawNForEffect(owner.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if owner.Life != ownerLife || me.Life != myLife {
		t.Errorf("the third draw does nothing: owner %d → %d", ownerLife, owner.Life)
	}
}

func TestB40VeinwitchCovenPaysBlackToReturnACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Veinwitch Coven", "Creature — Vampire Warlock", effectstest.B40VeinwitchCovenOracle, false)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: dead, Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B")
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
	// The target is chosen when the trigger goes on the stack
	// (CR 603.3d), before anyone is asked about the {B}.
	effectstest.B40PickCard(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
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
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	vincent := effectstest.B12Push(g, me.ID, "Vincent, Vengeful Atoner", "Legendary Creature — Assassin", effectstest.B40VincentOracle, 7, 7)
	oppLife, otherLife := opp.Life, other.Life
	effectstest.AttackWith(t, g, opp.ID, vincent)
	// Both of Vincent's abilities trigger off the same damage, so
	// their controller orders them (CR 603.3b) before either resolves.
	effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-7 {
		t.Fatalf("the defender takes 7 in combat: %d → %d", oppLife, opp.Life)
	}
	if other.Life != otherLife-7 {
		t.Errorf("at 7 power the Chaos ability hits each OTHER opponent for 7: %d → %d", otherLife, other.Life)
	}
	if effectstest.CounterOn(g, vincent, game.CounterPlusOne) != 1 {
		t.Errorf("one +1/+1 counter for the combat damage, got %d", effectstest.CounterOn(g, vincent, game.CounterPlusOne))
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
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	vincent := effectstest.B12Push(g, me.ID, "Vincent, Vengeful Atoner", "Legendary Creature — Assassin", effectstest.B40VincentOracle, 6, 6)
	if err := g.AddCounterForEffect(vincent, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounterForEffect: %v", err)
	}
	oppLife, otherLife := opp.Life, other.Life
	effectstest.AttackWith(t, g, opp.ID, vincent)
	effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-7 {
		t.Fatalf("the defender takes 7 in combat (6 printed + the counter): %d → %d", oppLife, opp.Life)
	}
	if other.Life != otherLife-7 {
		t.Errorf("a Vincent at CURRENT power 7 (6 printed + one +1/+1 counter, no anthem) must still "+
			"spill its 7 over: %d → %d", otherLife, other.Life)
	}
}

func TestB40MurderousRedcapDealsDamageEqualToItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	me := g.Seats[0]
	before := opp.Life
	effectstest.B36CastCreature(t, g, "Murderous Redcap", "Creature — Goblin Assassin", effectstest.B40MurderousRedcapOracle, 2, 2)
	// The enters trigger is targeted, so the target is chosen as it
	// goes on the stack.
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("a printed 2/2 pings for 2: %d → %d", before, opp.Life)
	}
	spec, _ := effects.Lookup(effectstest.B40MurderousRedcapOracle)
	if spec.Completeness != effects.CompletenessFull {
		t.Error("persist is implemented (#2075): the Redcap is Full")
	}
}

// TestB40MurderousRedcapCountersItsDamage is #1281: the ETB read
// src.Effective().Power, which excludes +1/+1 / -1/-1 counters, so a
// Redcap pumped by a counter (rather than an anthem) dealt the wrong
// amount. A counter placed while the trigger is still on the stack —
// resolution reads power fresh, per the card's own doc comment — must
// count.
func TestB40MurderousRedcapCountersItsDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	me := g.Seats[0]
	before := opp.Life
	redcap := effectstest.B36CastCreature(t, g, "Murderous Redcap", "Creature — Goblin Assassin", effectstest.B40MurderousRedcapOracle, 2, 2)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	// The trigger is on the stack, un-resolved: a +1/+1 counter lands
	// on the Redcap before it fires.
	if err := g.AddCounterForEffect(redcap, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounterForEffect: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("a 2/2 with a +1/+1 counter pings for 3: %d → %d", before, opp.Life)
	}
}

func TestB40AngelOfInventionFabricatesEitherWay(t *testing.T) {
	for _, counters := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		angel := effectstest.B36CastCreature(t, g, "Angel of Invention", "Creature — Angel", effectstest.B40AngelOfInventionOracle, 2, 1)
		ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
		if ask == nil {
			t.Fatal("fabricate 2 asks which mode")
		}
		if err := g.ResolveConfirm(ask.ID, me.ID, counters); err != nil {
			t.Fatalf("ResolveConfirm: %v", err)
		}
		if counters {
			if effectstest.CounterOn(g, angel, game.CounterPlusOne) != 2 {
				t.Errorf("the counters branch puts two +1/+1 counters on it, got %d", effectstest.CounterOn(g, angel, game.CounterPlusOne))
			}
			if n := effectstest.B16CountNamed(g, "Servo"); n != 0 {
				t.Errorf("and makes no Servos, got %d", n)
			}
			continue
		}
		if n := effectstest.B16CountNamed(g, "Servo"); n != 2 {
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
		if c := effectstest.FindTestCard(g, angel); c != nil {
			if eff := c.Effective(); eff.Power != 2 || eff.Toughness != 1 {
				t.Errorf("the Angel does not pump itself: %d/%d, want 2/1", eff.Power, eff.Toughness)
			}
		}
	}
}

func TestB40DecoctionModuleGivesEnergyAndBouncesYourOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	module := effectstest.PushCatalogPermanent(g, me.ID, "Decoction Module", "Artifact", effectstest.B40DecoctionModuleOracle, false)
	before := me.Energy
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Energy != before+1 {
		t.Fatalf("a creature entering under your control gives you {E}: %d → %d", before, me.Energy)
	}
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, module, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs)}); err == nil {
		t.Fatal("an opponent's creature is not a legal target")
	}
	effectstest.B16Activate(t, g, me.ID, module, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(mine)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || !me.Hand.Contains(mine) {
		t.Error("your own creature goes back to your hand")
	}
}

func TestB40LumberingWorldwagonIsAsBigAsYourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B36Lands(g, me.ID, 3)
	wagon := effectstest.B12Push(g, me.ID, "Lumbering Worldwagon", "Artifact — Vehicle", effectstest.B40LumberingWorldwagonOracle, 0, 4)
	// A Vehicle is not a creature until it is crewed, and a
	// non-creature has no P/T for the layer engine to define — so the
	// characteristic-defining power is only observable once it is.
	crewer := effectstest.PushCrewerForTest(g, me.ID, "Crewer", 4)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, wagon, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{crewer}}); err != nil {
		t.Fatalf("crew 4: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, wagon); got != 3 {
		t.Fatalf("power is the number of lands you control (3), got %d", got)
	}
	// Recomputed every pass, so two more lands is two more power with
	// no invalidation bookkeeping.
	effectstest.B36Lands(g, me.ID, 2)
	if got := effectstest.EffectivePower(t, g, wagon); got != 5 {
		t.Errorf("two more lands, two more power: got %d, want 5", got)
	}
	// The toughness is printed, not derived.
	if got := effectstest.EffectiveToughness(t, g, wagon); got != 4 {
		t.Errorf("toughness stays the printed 4, got %d", got)
	}
}

func TestB40LumberingWorldwagonMayFetchABasicOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Ketria Triome", TypeLine: "Land — Forest Island Mountain"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
	)
	effectstest.CastCatalogSpell(t, g, "Lumbering Worldwagon", "Artifact — Vehicle", effectstest.B40LumberingWorldwagonOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("entering opens the optional search")
	}
	if effectstest.SearchOptionNamed(g, c, "Forest") == uuid.Nil {
		t.Error("a basic land is offered")
	}
	if effectstest.SearchOptionNamed(g, c, "Ketria Triome") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a nonbasic land and a creature are not")
	}
}

// --- statics and cost modifiers ------------------------------------

func TestB40JacesSanctumDiscountsYourInstantsAndScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Jace's Sanctum", "Enchantment", effectstest.B40JacesSanctumOracle, false)
	if got := effectstest.B40PricedFor(t, g, me.ID, "Divination", "Sorcery", "{2}{U}"); got != 2 {
		t.Errorf("your sorcery costs {1} less: mana value %d, want 2", got)
	}
	if got := effectstest.B40PricedFor(t, g, opp.ID, "Divination", "Sorcery", "{2}{U}"); got != 3 {
		t.Errorf("an opponent's sorcery is untouched: mana value %d, want 3", got)
	}
	if got := effectstest.B40PricedFor(t, g, me.ID, "Bear", "Creature — Bear", "{2}{G}"); got != 3 {
		t.Errorf("a creature spell is not an instant or a sorcery: mana value %d, want 3", got)
	}
	effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry) == nil {
		t.Error("casting an instant or sorcery scries 1")
	}
}

func TestB40GargosDiscountsHydraSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Gargos, Vicious Watcher", "Legendary Creature — Hydra", effectstest.B40GargosOracle, false)
	if got := effectstest.B40PricedFor(t, g, me.ID, "Hooded Hydra", "Creature — Snake Hydra", "{4}{G}"); got != 1 {
		t.Errorf("your Hydra costs {4} less, off the generic: mana value %d, want 1", got)
	}
	if got := effectstest.B40PricedFor(t, g, me.ID, "Bear", "Creature — Bear", "{4}{G}"); got != 5 {
		t.Errorf("a Bear is not a Hydra: mana value %d, want 5", got)
	}
	if got := effectstest.B40PricedFor(t, g, opp.ID, "Hooded Hydra", "Creature — Snake Hydra", "{4}{G}"); got != 5 {
		t.Errorf("an opponent's Hydra is untouched: mana value %d, want 5", got)
	}
}

func TestB40MechanizedWarfareBoostsOnlyYourRedOrArtifactDamageAtOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Mechanized Warfare", "Enchantment", effectstest.B40MechanizedWarfareOracle, false)
	red := effectstest.B16Creature(g, me.ID, "Goblin", "Creature — Goblin", 2, 2, "R")
	green := effectstest.B16Creature(g, me.ID, "Elf", "Creature — Elf", 2, 2, "G")
	theirRed := effectstest.B16Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 2, 2, "R")

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
	g := effectstest.NewCatalogGame(t)
	dragon := effectstest.B12PlayFromHand(t, g, "Shivan Devastator", "Creature — Dragon Hydra", effectstest.B40ShivanDevastatorOracle,
		game.CastSpellParams{XValue: 4})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterOn(g, dragon, game.CounterPlusOne) != 4 {
		t.Errorf("X = 4 puts four +1/+1 counters on it, got %d", effectstest.CounterOn(g, dragon, game.CounterPlusOne))
	}
	spec, _ := effects.Lookup(effectstest.B40ShivanDevastatorOracle)
	for _, kw := range []string{"flying", "haste"} {
		if !effectstest.HasAbility(spec.PrintedKeywords, kw) {
			t.Errorf("the Dragon prints %s", kw)
		}
	}
}

func TestB40BygoneColossusOffersItsWarpCost(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.B40BygoneColossusOracle)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Manakin", "Artifact Creature — Construct", effectstest.B40ManakinOracle, false)
	effectstest.B28TapForMana(t, g, me.ID, rock, "C")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("tapped for {C}: pool %v", got)
	}
}

func TestB40OliphauntPumpsAnotherCreatureAndNotItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oliphaunt := effectstest.B12Push(g, me.ID, "Oliphaunt", "Creature — Elephant", effectstest.B40OliphauntOracle, 6, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, oliphaunt)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the attack trigger asks for another creature you control")
	}
	if effectstest.HasID(p.PickTargetCards, oliphaunt) {
		t.Error("\"another\" excludes the Oliphaunt itself")
	}
	effectstest.B40PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, bear)
	if !ok {
		t.Fatal("the Bear is still there")
	}
	if eff := c.Effective(); eff.Power != 4 || eff.Toughness != 2 {
		t.Errorf("the Bear is 4/2 until end of turn: %d/%d", eff.Power, eff.Toughness)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("and it has trample until end of turn")
	}
}

// --- #1412: cycling ---------------------------------------------------

// Polluted Mire and Remote Isle: cycling {2} from hand discards the
// land and draws a card, and is refused without the mana to pay it.
func TestB40CyclingLandsCycleForACard(t *testing.T) {
	for _, row := range []struct{ name, oracle string }{
		{"Polluted Mire", effectstest.B40PollutedMireOracle},
		{"Remote Isle", effectstest.B40RemoteIsleOracle},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]

		refused := effectstest.PushCatalogHandCard(me, row.name, "Land", row.oracle)
		if err := g.ActivateCatalogAbility(me.ID, refused, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
			t.Fatalf("%s: {2} must be paid to cycle", row.name)
		}
		if !me.Hand.Contains(refused) {
			t.Errorf("%s: discarded despite the unpaid cost", row.name)
		}
		g.WithWriteLock(func() { me.Hand.Cards = nil })

		id, _ := effectstest.CycleFromHand(t, g, row.name, "Land", row.oracle, "{C}{C}")
		if !me.Graveyard.Contains(id) {
			t.Fatalf("%s: the cycled land is not in the graveyard", row.name)
		}
		before := len(me.Hand.Cards)
		effectstest.PassPriorityAroundTable(t, g)
		if got := len(me.Hand.Cards); got != before+1 {
			t.Errorf("%s: hand %d -> %d, want the cycling draw", row.name, before, got)
		}
	}
}

// Oliphaunt: mountaincycling {1} discards it and fetches a Mountain
// card (any land with the Mountain type, not only a basic one),
// reveals it and shuffles; refused without the mana.
func TestB40OliphauntMountaincyclesForAMountain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]

	refused := effectstest.PushCatalogHandCard(me, "Oliphaunt", "Creature — Elephant", effectstest.B40OliphauntOracle)
	if err := g.ActivateCatalogAbility(me.ID, refused, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("{1} must be paid to mountaincycle")
	}
	if !me.Hand.Contains(refused) {
		t.Error("discarded despite the unpaid cost")
	}
	g.WithWriteLock(func() { me.Hand.Cards = nil })

	ids := effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Mountain", "Basic Land — Mountain"),
		effectstest.SearchTestLand("Rugged Prairie", "Land — Mountain Plains"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	mountain, dual := ids[0], ids[1]

	id, _ := effectstest.CycleFromHand(t, g, "Oliphaunt", "Creature — Elephant", effectstest.B40OliphauntOracle, "{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the mountaincycled Oliphaunt is not in the graveyard")
	}
	handBefore := len(me.Hand.Cards)
	libraryBefore := len(me.Library.Cards)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.SearchChoiceFor(g, me.ID)
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
