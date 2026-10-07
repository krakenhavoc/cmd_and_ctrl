package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reveal_or_pay_test.go — ADR 0100 amendment 2026-10-07: the reveal and
// behold branch components, and the fifteen cards that carry them. The
// component rules are tested on Wren's Run Vanquisher (reveal) and
// Silvergill Mentor (behold); the rest get one test for what is theirs.

const (
	vanquisherOracle     = "fe9cdd15-a390-4a1e-bae9-474ce8d355b8"
	mentorOracle         = "dcf07f38-422f-46c3-aee9-50540b9c9115"
	skymarcherOracle     = "9e1fe5ec-1669-4587-8e42-b189fd6d39e0"
	adeptOracle          = "cb66c5ec-e7a4-4bd2-b825-6c9b846ba40d"
	pieSneakOracle       = "7243b935-73fc-47d2-ac65-ab390ed94f3f"
	thunderherdOracle    = "006bbd8b-2007-419d-8b84-9ff73f2f95b7"
	causticExhaleOracle  = "c4023e1f-26e3-4225-98d0-3d33f6654616"
	aspirantOracle       = "845d80ee-1cd5-437a-90a5-40dcc657d3aa"
	lysAlanaOracle       = "ce7582a2-e014-41fb-8753-00d29cffc81a"
	cursetosserOracle    = "b63e5518-61f3-4cde-a799-5f3ee6aba70b"
	soulbrightOracle     = "762679dd-4345-43c3-929b-15959443aeb0"
	surtlandOracle       = "b1a3b2ab-d9f8-4a0a-b2e2-aaa9989b2e41"
	revealPlainOracle    = "1e6f2997-a84b-4302-9d8b-ca4928961d24" // Daring Buccaneer
	bladewhirlOracle     = "c37770aa-4770-4198-b6a6-f9ff86dbb06b"
	goldmeadowOracle     = "1aff8653-12c5-4549-915f-ac2fd6dfd43b"
	revealBranchIdx      = 0
	manaBranchIdx        = 1
	vanquisherManaCost   = "{1}{G}"
	vanquisherTypeLine   = "Creature — Elf Warrior"
	mentorManaCost       = "{1}{U}"
	mentorTypeLine       = "Creature — Merfolk Wizard"
	revealTestHandName   = "Elf Friend"
	revealTestHandType   = "Creature — Elf Druid"
	revealTestOtherType  = "Creature — Goblin"
	revealTestMerfolkTyp = "Creature — Merfolk"
)

func castVanquisher(t *testing.T, g *game.Game, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	return castWithTapParams(t, g, "Wren's Run Vanquisher", vanquisherTypeLine, vanquisherManaCost, vanquisherOracle, params)
}

func castMentor(t *testing.T, g *game.Game, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	return castWithTapParams(t, g, "Silvergill Mentor", mentorTypeLine, mentorManaCost, mentorOracle, params)
}

// revealedEvents counts the EventRevealCards the table has seen for id.
func revealedEvents(g *game.Game, id uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventRevealCards && ev.CardID == id {
			n++
		}
	}
	return n
}

// The reveal branch shows the named card to the whole table, moves
// nothing, pays no extra mana, and the creature resolves.
func TestRevealBranchRevealsTheNamedCardAndMovesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elf := handCard(me, revealTestHandName, revealTestHandType)
	id, err := castVanquisher(t, g, game.CastSpellParams{CostBranch: branch(revealBranchIdx), RevealIDs: []uuid.UUID{elf}})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if !me.Hand.Contains(elf) {
		t.Fatal("revealing moved the card out of the hand")
	}
	if got := revealedEvents(g, elf); got != 1 {
		t.Fatalf("%d reveal events for the named Elf, want 1", got)
	}
	for _, seat := range g.Seats {
		var known bool
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == elf {
				known = me.Hand.Cards[i].IsKnownTo(seat.ID)
			}
		}
		if !known {
			t.Fatalf("seat %s does not know the revealed card", seat.ID)
		}
	}
	if item := g.StackMeta[id]; item == nil || item.Paid.CostBranch != 1 {
		t.Fatalf("Paid = %+v, want CostBranch 1 (branch 0)", item)
	}
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, id) {
		t.Fatal("the Vanquisher did not resolve onto the battlefield")
	}
}

// The mana branch charges {3} on top of the printed cost and reveals
// nothing; the reveal branch charges the printed cost alone.
func TestRevealOrPayPricesOnlyTheManaBranch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	card := game.Card{InstanceID: uuid.New(), Name: "Wren's Run Vanquisher", TypeLine: vanquisherTypeLine,
		ManaCost: vanquisherManaCost, OracleID: vanquisherOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	for _, tc := range []struct {
		branch *int
		want   int
	}{{nil, 2}, {branch(revealBranchIdx), 2}, {branch(manaBranchIdx), 5}} {
		price, err := g.PriceCast(me.ID, card, game.CastSpellParams{CostBranch: tc.branch})
		if err != nil {
			t.Fatalf("PriceCast: %v", err)
		}
		if got := price.Total.ManaValue(); got != tc.want {
			t.Errorf("branch %v: mana value %d, want %d", tc.branch, got, tc.want)
		}
	}
	id, err := castVanquisher(t, g, game.CastSpellParams{CostBranch: branch(manaBranchIdx)})
	if err != nil {
		t.Fatalf("mana branch CastSpell: %v", err)
	}
	if got := revealedEvents(g, id); got != 0 {
		t.Fatalf("the mana branch revealed %d cards", got)
	}
}

// Every refusal leaves the table untouched: no card named, a card that
// is not an Elf, a card in someone else's hand, the spell itself, two
// cards, and a reveal sent with the mana branch.
func TestRevealBranchRefusals(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	elf := handCard(me, revealTestHandName, revealTestHandType)
	goblin := handCard(me, "Goblin Friend", revealTestOtherType)
	theirs := handCard(foe, "Their Elf", revealTestHandType)
	for _, tc := range []struct {
		why    string
		params game.CastSpellParams
	}{
		{"no card named", game.CastSpellParams{CostBranch: branch(revealBranchIdx)}},
		{"a Goblin for an Elf", game.CastSpellParams{CostBranch: branch(revealBranchIdx), RevealIDs: []uuid.UUID{goblin}}},
		{"an Elf in another hand", game.CastSpellParams{CostBranch: branch(revealBranchIdx), RevealIDs: []uuid.UUID{theirs}}},
		{"an unknown card", game.CastSpellParams{CostBranch: branch(revealBranchIdx), RevealIDs: []uuid.UUID{uuid.New()}}},
		{"two cards", game.CastSpellParams{CostBranch: branch(revealBranchIdx), RevealIDs: []uuid.UUID{elf, elf}}},
		{"a reveal sent with the mana branch", game.CastSpellParams{CostBranch: branch(manaBranchIdx), RevealIDs: []uuid.UUID{elf}}},
	} {
		if _, err := castVanquisher(t, g, tc.params); err == nil {
			t.Errorf("%s: cast accepted", tc.why)
		}
		if got := revealedEvents(g, elf) + revealedEvents(g, goblin); got != 0 {
			t.Fatalf("%s: a refused cast revealed a card", tc.why)
		}
	}
}

// The spell cannot reveal itself: the Vanquisher is an Elf, and with
// no other Elf in hand the reveal branch is unpayable, though the
// spell is a legal pick on its face (CR 601.2a moved it to the stack).
func TestARevealingSpellCannotRevealItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	card := game.Card{InstanceID: uuid.New(), Name: "Wren's Run Vanquisher", TypeLine: vanquisherTypeLine,
		ManaCost: vanquisherManaCost, OracleID: vanquisherOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	if g.AdditionalCostBranchPayableLocked(me.ID, card, revealBranchIdx) {
		t.Fatal("the reveal branch is payable with no other Elf in hand")
	}
	if !g.AdditionalCostBranchPayableLocked(me.ID, card, manaBranchIdx) {
		t.Fatal("the mana branch must stay payable")
	}
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, card.InstanceID,
		game.CastSpellParams{CostBranch: branch(revealBranchIdx), RevealIDs: []uuid.UUID{card.InstanceID}}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("revealing itself: %v, want ErrInvalidParam", err)
	}
}

// A changeling in hand is every creature type, so it pays a reveal.
func TestAChangelingPaysAReveal(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	changeling := game.Card{InstanceID: uuid.New(), Name: "Mirror Entity Kin", TypeLine: "Creature — Shapeshifter",
		Keywords: []string{game.KeywordChangeling}, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(changeling)
	if _, err := castVanquisher(t, g, game.CastSpellParams{CostBranch: branch(revealBranchIdx), RevealIDs: []uuid.UUID{changeling.InstanceID}}); err != nil {
		t.Fatalf("a changeling should be an Elf card: %v", err)
	}
}

// Behold: a matching permanent you control pays, and nothing is
// revealed; a matching card in hand pays and is revealed.
func TestBeholdChoosesAPermanentOrRevealsAHandCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	merfolk := pushCatalogPermanent(g, me.ID, "Merfolk Scout", revealTestMerfolkTyp, "", false)
	id, err := castMentor(t, g, game.CastSpellParams{CostBranch: branch(revealBranchIdx), RevealIDs: []uuid.UUID{merfolk}})
	if err != nil {
		t.Fatalf("behold a permanent: %v", err)
	}
	if got := revealedEvents(g, merfolk); got != 0 {
		t.Fatalf("choosing a permanent revealed it (%d events)", got)
	}
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, id) {
		t.Fatal("the Mentor did not resolve")
	}
	// Mentor's ETB token is a white and blue Merfolk.
	if n := onBattlefieldNamed(g, "Merfolk"); n < 1 {
		t.Fatal("the Mentor's Merfolk token was not created")
	}

	hand := handCard(me, "Merfolk Pal", revealTestMerfolkTyp)
	if _, err := castMentor(t, g, game.CastSpellParams{CostBranch: branch(revealBranchIdx), RevealIDs: []uuid.UUID{hand}}); err != nil {
		t.Fatalf("behold a hand card: %v", err)
	}
	if got := revealedEvents(g, hand); got != 1 {
		t.Fatalf("%d reveal events for the beheld hand card, want 1", got)
	}
}

// Behold refuses a permanent of the wrong type and one the caster does
// not control.
func TestBeholdRefusals(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	bear := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	theirMerfolk := pushCatalogPermanent(g, foe.ID, "Their Merfolk", revealTestMerfolkTyp, "", false)
	for _, tc := range []struct {
		why string
		id  uuid.UUID
	}{{"a Bear for a Merfolk", bear}, {"an opponent's Merfolk", theirMerfolk}} {
		if _, err := castMentor(t, g, game.CastSpellParams{CostBranch: branch(revealBranchIdx), RevealIDs: []uuid.UUID{tc.id}}); err == nil {
			t.Errorf("%s: cast accepted", tc.why)
		}
	}
}

// A plain card with no reveal branch refuses reveal_ids.
func TestRevealIDsOnACardWithoutARevealCostAreRefused(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elf := handCard(me, revealTestHandName, revealTestHandType)
	if _, err := castWithTapParams(t, g, "Bone Shards", "Sorcery", "{B}", boneShardsOracle,
		game.CastSpellParams{CostBranch: branch(0), DiscardIDs: []uuid.UUID{handCard(me, "Pitch", "Instant")}, RevealIDs: []uuid.UUID{elf}}); err == nil {
		t.Fatal("reveal_ids accepted on Bone Shards")
	}
}

// --- the cards --------------------------------------------------------

// All fifteen register an either/or cost whose branch 0 reveals or
// beholds and whose branch 1 is mana, and no card is mislabelled.
func TestEveryRevealOrPayCardIsBranched(t *testing.T) {
	want := map[string]struct {
		sub    string
		behold bool
		mana   string
	}{
		"1e6f2997-a84b-4302-9d8b-ca4928961d24": {"Pirate", false, "{2}"},
		"c37770aa-4770-4198-b6a6-f9ff86dbb06b": {"Elemental", false, "{3}"},
		"1aff8653-12c5-4549-915f-ac2fd6dfd43b": {"Kithkin", false, "{3}"},
		skymarcherOracle:                       {"Vampire", false, "{1}"},
		adeptOracle:                            {"Merfolk", false, "{3}"},
		pieSneakOracle:                         {"Goblin", false, "{3}"},
		vanquisherOracle:                       {"Elf", false, "{3}"},
		thunderherdOracle:                      {"Dinosaur", false, "{1}"},
		surtlandOracle:                         {"Giant", false, "{2}"},
		causticExhaleOracle:                    {"Dragon", true, "{1}"},
		aspirantOracle:                         {"Kithkin", true, "{2}"},
		lysAlanaOracle:                         {"Elf", true, "{2}"},
		cursetosserOracle:                      {"Goblin", true, "{2}"},
		mentorOracle:                           {"Merfolk", true, "{2}"},
		soulbrightOracle:                       {"Elemental", true, "{2}"},
	}
	for oracle, w := range want {
		ac := game.AdditionalCostFor(oracle)
		if !ac.Branched() || len(ac.Either) != 2 {
			t.Errorf("%s: not a two-branch either/or cost", oracle)
			continue
		}
		r := ac.Either[0].Reveal
		if r == nil || r.Subtype != w.sub || r.Behold != w.behold {
			t.Errorf("%s: branch 0 = %+v, want %s behold=%v", oracle, r, w.sub, w.behold)
		}
		if ac.Either[1].ManaCost != w.mana {
			t.Errorf("%s: branch 1 mana %q, want %q", oracle, ac.Either[1].ManaCost, w.mana)
		}
	}
}

func castCreatureThroughReveal(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, revealSub string) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	pal := handCard(me, "Pal", "Creature — "+revealSub)
	id, err := castWithTapParams(t, g, name, typeLine, cost, oracle,
		game.CastSpellParams{CostBranch: branch(revealBranchIdx), RevealIDs: []uuid.UUID{pal}})
	if err != nil {
		t.Fatalf("%s: CastSpell: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, id) {
		t.Fatalf("%s did not resolve", name)
	}
	return id
}

func TestSadisticSkymarcherHasFlyingAndLifelink(t *testing.T) {
	g := newCatalogGame(t)
	id := castCreatureThroughReveal(t, g, "Sadistic Skymarcher", "Creature — Vampire Soldier", "{2}{B}", skymarcherOracle, "Vampire")
	assertKeywords(t, g, id, "flying", "lifelink")
}

func TestSilvergillAdeptDrawsOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Land"})
	hand := me.Hand.Size()
	castCreatureThroughReveal(t, g, "Silvergill Adept", "Creature — Merfolk Wizard", "{1}{U}", adeptOracle, "Merfolk")
	// The helper adds a pal (+1) and the Adept (+1), the cast removes the
	// Adept (-1), and the enters trigger draws one (+1).
	if got := me.Hand.Size(); got != hand+2 {
		t.Fatalf("hand %d -> %d, want +2 (the pal and the drawn card)", hand, got)
	}
}

func TestSqueakingPieSneakHasFear(t *testing.T) {
	g := newCatalogGame(t)
	id := castCreatureThroughReveal(t, g, "Squeaking Pie Sneak", "Creature — Goblin Rogue", "{1}{B}", pieSneakOracle, "Goblin")
	assertKeywords(t, g, id, "fear")
}

func TestWrensRunVanquisherHasDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	id := castCreatureThroughReveal(t, g, "Wren's Run Vanquisher", vanquisherTypeLine, vanquisherManaCost, vanquisherOracle, "Elf")
	assertKeywords(t, g, id, "deathtouch")
}

func TestVanillaRevealCreaturesResolve(t *testing.T) {
	for _, c := range []struct{ name, typeLine, cost, oracle, sub string }{
		{"Daring Buccaneer", "Creature — Human Pirate", "{R}", revealPlainOracle, "Pirate"},
		{"Flamekin Bladewhirl", "Creature — Elemental Warrior", "{R}", bladewhirlOracle, "Elemental"},
		{"Goldmeadow Stalwart", "Creature — Kithkin Soldier", "{W}", goldmeadowOracle, "Kithkin"},
	} {
		g := newCatalogGame(t)
		castCreatureThroughReveal(t, g, c.name, c.typeLine, c.cost, c.oracle, c.sub)
	}
}

func TestThunderherdMigrationFetchesABasicTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	pal := handCard(me, "Rex", "Creature — Dinosaur")
	if _, err := castWithTapParams(t, g, "Thunderherd Migration", "Sorcery", "{1}{G}", thunderherdOracle,
		game.CastSpellParams{CostBranch: branch(revealBranchIdx), RevealIDs: []uuid.UUID{pal}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, forest); !ok || !c.Tapped {
		t.Fatalf("the basic land should be on the battlefield tapped (found %v, %+v)", ok, c)
	}
}

func TestCausticExhaleGivesMinusThreeMinusThree(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	victim := pushVanillaCreature(g, foe.ID, "Ogre", 3, 3)
	dragon := handCard(me, "Whelp", "Creature — Dragon")
	if _, err := castWithTapParams(t, g, "Caustic Exhale", "Instant", "{B}", causticExhaleOracle,
		game.CastSpellParams{CostBranch: branch(revealBranchIdx), RevealIDs: []uuid.UUID{dragon},
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, victim) {
		t.Fatal("the 3/3 survived -3/-3")
	}
}

func TestKinsbaileAspirantGrowsWhenAnotherCreatureEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	aspirant := pushCatalogPermanent(g, me.ID, "Kinsbaile Aspirant", "Creature — Kithkin Citizen", aspirantOracle, false)
	before := effectivePower(t, g, aspirant)
	castWithTapParams(t, g, "Daring Buccaneer", "Creature — Human Pirate", "{R}", revealPlainOracle,
		game.CastSpellParams{CostBranch: branch(manaBranchIdx)})
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, aspirant); got != before+1 {
		t.Fatalf("power %d -> %d, want +1 for the creature that entered", before, got)
	}
}

func TestLysAlanaDignitaryMakesManaOnlyWithAnElfInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dignitary := pushCatalogPermanent(g, me.ID, "Lys Alana Dignitary", "Creature — Elf Advisor", lysAlanaOracle, false)
	if err := g.ActivateManaAbility(me.ID, dignitary, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("the mana ability activated with no Elf card in the graveyard")
	}
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Elf", TypeLine: "Creature — Elf", Owner: me.ID, Controller: me.ID})
	if err := g.ActivateManaAbility(me.ID, dignitary, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate with an Elf in the graveyard: %v", err)
	}
	if n := len(me.ManaPool); n != 2 {
		t.Fatalf("%d mana in the pool, want {G}{G}", n)
	}
}

func TestMudbuttonCursetosserCantBlockAndItsDeathDestroysASmallCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	curse := pushCatalogPermanent(g, me.ID, "Mudbutton Cursetosser", "Creature — Goblin Warlock", cursetosserOracle, false)
	small := pushVanillaCreature(g, foe.ID, "Small", 2, 2)
	big := pushVanillaCreature(g, foe.ID, "Big", 3, 3)
	g.ReadSnapshot(func() { g.RecomputeLayersIfStaleLocked() })
	c, ok := battlefieldCard(g, curse)
	if !ok || !game.Restricted(&c, game.CantBlock) {
		t.Fatal("the Cursetosser can block")
	}
	if err := g.DestroyPermanentForEffect(curse); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	if latest := latestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID); latest != nil {
		if err := g.ResolvePickTarget(latest.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: small}); err != nil {
			t.Fatalf("pick target: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, small) {
		t.Fatal("the power-2 creature survived the Cursetosser's death trigger")
	}
	if !onBattlefield(g, big) {
		t.Fatal("the power-3 creature should not be a legal target")
	}
}

func TestSoulbrightSeekerAddsManaOnTheThirdResolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seeker := pushCatalogPermanent(g, me.ID, "Soulbright Seeker", "Creature — Elemental Sorcerer", soulbrightOracle, false)
	for i := 1; i <= 3; i++ {
		if err := g.ActivateCatalogAbility(me.ID, seeker, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: seeker}},
		}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		passPriorityAroundTable(t, g)
		got := len(me.ManaPool)
		if i < 3 && got != 0 {
			t.Fatalf("activation %d added mana", i)
		}
		if i == 3 && got != 4 {
			t.Fatalf("third resolution left %d mana in the pool, want {R}{R}{R}{R}", got)
		}
	}
	assertKeywords(t, g, seeker, "trample")
}

func TestSurtlandElementalistOffersAFreeInstantOrSorceryOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	giant := pushCatalogPermanent(g, me.ID, "Surtland Elementalist", "Creature — Giant Wizard", surtlandOracle, false)
	instant := pushHandCardWithManaCost(me, "Big Instant", "Instant", "{6}")
	creature := pushHandCardWithManaCost(me, "Big Creature", "Creature — Bear", "{1}")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(giant, foe.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no free-cast pick on attacking")
	}
	if !hasID(pick.ChooseCards, instant) || hasID(pick.ChooseCards, creature) {
		t.Fatalf("offered %v, want only the instant", pick.ChooseCards)
	}
}
