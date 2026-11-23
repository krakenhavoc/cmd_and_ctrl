package testsr

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func castVanquisher(t *testing.T, g *game.Game, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	return effectstest.CastWithTapParams(t, g, "Wren's Run Vanquisher", effectstest.VanquisherTypeLine, effectstest.VanquisherManaCost, effectstest.VanquisherOracle, params)
}

func castMentor(t *testing.T, g *game.Game, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	return effectstest.CastWithTapParams(t, g, "Silvergill Mentor", effectstest.MentorTypeLine, effectstest.MentorManaCost, effectstest.MentorOracle, params)
}

// The reveal branch shows the named card to the whole table, moves
// nothing, pays no extra mana, and the creature resolves.
func TestRevealBranchRevealsTheNamedCardAndMovesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elf := effectstest.HandCard(me, effectstest.RevealTestHandName, effectstest.RevealTestHandType)
	id, err := castVanquisher(t, g, game.CastSpellParams{CostBranch: effectstest.Branch(effectstest.RevealBranchIdx), RevealIDs: []uuid.UUID{elf}})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if !me.Hand.Contains(elf) {
		t.Fatal("revealing moved the card out of the hand")
	}
	if got := effectstest.RevealedEvents(g, elf); got != 1 {
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
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, id) {
		t.Fatal("the Vanquisher did not resolve onto the battlefield")
	}
}

// The mana branch charges {3} on top of the printed cost and reveals
// nothing; the reveal branch charges the printed cost alone.
func TestRevealOrPayPricesOnlyTheManaBranch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	card := game.Card{InstanceID: uuid.New(), Name: "Wren's Run Vanquisher", TypeLine: effectstest.VanquisherTypeLine,
		ManaCost: effectstest.VanquisherManaCost, OracleID: effectstest.VanquisherOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	for _, tc := range []struct {
		branch *int
		want   int
	}{{nil, 2}, {effectstest.Branch(effectstest.RevealBranchIdx), 2}, {effectstest.Branch(effectstest.ManaBranchIdx), 5}} {
		price, err := g.PriceCast(me.ID, card, game.CastSpellParams{CostBranch: tc.branch})
		if err != nil {
			t.Fatalf("PriceCast: %v", err)
		}
		if got := price.Total.ManaValue(); got != tc.want {
			t.Errorf("branch %v: mana value %d, want %d", tc.branch, got, tc.want)
		}
	}
	id, err := castVanquisher(t, g, game.CastSpellParams{CostBranch: effectstest.Branch(effectstest.ManaBranchIdx)})
	if err != nil {
		t.Fatalf("mana branch CastSpell: %v", err)
	}
	if got := effectstest.RevealedEvents(g, id); got != 0 {
		t.Fatalf("the mana branch revealed %d cards", got)
	}
}

// Every refusal leaves the table untouched: no card named, a card that
// is not an Elf, a card in someone else's hand, the spell itself, two
// cards, and a reveal sent with the mana branch.
func TestRevealBranchRefusals(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	elf := effectstest.HandCard(me, effectstest.RevealTestHandName, effectstest.RevealTestHandType)
	goblin := effectstest.HandCard(me, "Goblin Friend", effectstest.RevealTestOtherType)
	theirs := effectstest.HandCard(foe, "Their Elf", effectstest.RevealTestHandType)
	for _, tc := range []struct {
		why    string
		params game.CastSpellParams
	}{
		{"no card named", game.CastSpellParams{CostBranch: effectstest.Branch(effectstest.RevealBranchIdx)}},
		{"a Goblin for an Elf", game.CastSpellParams{CostBranch: effectstest.Branch(effectstest.RevealBranchIdx), RevealIDs: []uuid.UUID{goblin}}},
		{"an Elf in another hand", game.CastSpellParams{CostBranch: effectstest.Branch(effectstest.RevealBranchIdx), RevealIDs: []uuid.UUID{theirs}}},
		{"an unknown card", game.CastSpellParams{CostBranch: effectstest.Branch(effectstest.RevealBranchIdx), RevealIDs: []uuid.UUID{uuid.New()}}},
		{"two cards", game.CastSpellParams{CostBranch: effectstest.Branch(effectstest.RevealBranchIdx), RevealIDs: []uuid.UUID{elf, elf}}},
		{"a reveal sent with the mana branch", game.CastSpellParams{CostBranch: effectstest.Branch(effectstest.ManaBranchIdx), RevealIDs: []uuid.UUID{elf}}},
	} {
		if _, err := castVanquisher(t, g, tc.params); err == nil {
			t.Errorf("%s: cast accepted", tc.why)
		}
		if got := effectstest.RevealedEvents(g, elf) + effectstest.RevealedEvents(g, goblin); got != 0 {
			t.Fatalf("%s: a refused cast revealed a card", tc.why)
		}
	}
}

// The spell cannot reveal itself: the Vanquisher is an Elf, and with
// no other Elf in hand the reveal branch is unpayable, though the
// spell is a legal pick on its face (CR 601.2a moved it to the stack).
func TestARevealingSpellCannotRevealItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	card := game.Card{InstanceID: uuid.New(), Name: "Wren's Run Vanquisher", TypeLine: effectstest.VanquisherTypeLine,
		ManaCost: effectstest.VanquisherManaCost, OracleID: effectstest.VanquisherOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	if g.AdditionalCostBranchPayableLocked(me.ID, card, effectstest.RevealBranchIdx) {
		t.Fatal("the reveal branch is payable with no other Elf in hand")
	}
	if !g.AdditionalCostBranchPayableLocked(me.ID, card, effectstest.ManaBranchIdx) {
		t.Fatal("the mana branch must stay payable")
	}
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, card.InstanceID,
		game.CastSpellParams{CostBranch: effectstest.Branch(effectstest.RevealBranchIdx), RevealIDs: []uuid.UUID{card.InstanceID}}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("revealing itself: %v, want ErrInvalidParam", err)
	}
}

// A changeling in hand is every creature type, so it pays a reveal.
func TestAChangelingPaysAReveal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	changeling := game.Card{InstanceID: uuid.New(), Name: "Mirror Entity Kin", TypeLine: "Creature — Shapeshifter",
		Keywords: []string{game.KeywordChangeling}, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(changeling)
	if _, err := castVanquisher(t, g, game.CastSpellParams{CostBranch: effectstest.Branch(effectstest.RevealBranchIdx), RevealIDs: []uuid.UUID{changeling.InstanceID}}); err != nil {
		t.Fatalf("a changeling should be an Elf card: %v", err)
	}
}

// Behold: a matching permanent you control pays, and nothing is
// revealed; a matching card in hand pays and is revealed.
func TestBeholdChoosesAPermanentOrRevealsAHandCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	merfolk := effectstest.PushCatalogPermanent(g, me.ID, "Merfolk Scout", effectstest.RevealTestMerfolkTyp, "", false)
	id, err := castMentor(t, g, game.CastSpellParams{CostBranch: effectstest.Branch(effectstest.RevealBranchIdx), RevealIDs: []uuid.UUID{merfolk}})
	if err != nil {
		t.Fatalf("behold a permanent: %v", err)
	}
	if got := effectstest.RevealedEvents(g, merfolk); got != 0 {
		t.Fatalf("choosing a permanent revealed it (%d events)", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, id) {
		t.Fatal("the Mentor did not resolve")
	}
	// Mentor's ETB token is a white and blue Merfolk.
	if n := effectstest.OnBattlefieldNamed(g, "Merfolk"); n < 1 {
		t.Fatal("the Mentor's Merfolk token was not created")
	}

	hand := effectstest.HandCard(me, "Merfolk Pal", effectstest.RevealTestMerfolkTyp)
	if _, err := castMentor(t, g, game.CastSpellParams{CostBranch: effectstest.Branch(effectstest.RevealBranchIdx), RevealIDs: []uuid.UUID{hand}}); err != nil {
		t.Fatalf("behold a hand card: %v", err)
	}
	if got := effectstest.RevealedEvents(g, hand); got != 1 {
		t.Fatalf("%d reveal events for the beheld hand card, want 1", got)
	}
}

// Behold refuses a permanent of the wrong type and one the caster does
// not control.
func TestBeholdRefusals(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	theirMerfolk := effectstest.PushCatalogPermanent(g, foe.ID, "Their Merfolk", effectstest.RevealTestMerfolkTyp, "", false)
	for _, tc := range []struct {
		why string
		id  uuid.UUID
	}{{"a Bear for a Merfolk", bear}, {"an opponent's Merfolk", theirMerfolk}} {
		if _, err := castMentor(t, g, game.CastSpellParams{CostBranch: effectstest.Branch(effectstest.RevealBranchIdx), RevealIDs: []uuid.UUID{tc.id}}); err == nil {
			t.Errorf("%s: cast accepted", tc.why)
		}
	}
}

// A plain card with no reveal branch refuses reveal_ids.
func TestRevealIDsOnACardWithoutARevealCostAreRefused(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elf := effectstest.HandCard(me, effectstest.RevealTestHandName, effectstest.RevealTestHandType)
	if _, err := effectstest.CastWithTapParams(t, g, "Bone Shards", "Sorcery", "{B}", effectstest.BoneShardsOracle,
		game.CastSpellParams{CostBranch: effectstest.Branch(0), DiscardIDs: []uuid.UUID{effectstest.HandCard(me, "Pitch", "Instant")}, RevealIDs: []uuid.UUID{elf}}); err == nil {
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
		effectstest.SkymarcherOracle:           {"Vampire", false, "{1}"},
		effectstest.AdeptOracle:                {"Merfolk", false, "{3}"},
		effectstest.PieSneakOracle:             {"Goblin", false, "{3}"},
		effectstest.VanquisherOracle:           {"Elf", false, "{3}"},
		effectstest.ThunderherdOracle:          {"Dinosaur", false, "{1}"},
		effectstest.SurtlandOracle:             {"Giant", false, "{2}"},
		effectstest.CausticExhaleOracle:        {"Dragon", true, "{1}"},
		effectstest.AspirantOracle:             {"Kithkin", true, "{2}"},
		effectstest.LysAlanaOracle:             {"Elf", true, "{2}"},
		effectstest.CursetosserOracle:          {"Goblin", true, "{2}"},
		effectstest.MentorOracle:               {"Merfolk", true, "{2}"},
		effectstest.SoulbrightOracle:           {"Elemental", true, "{2}"},
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
	pal := effectstest.HandCard(me, "Pal", "Creature — "+revealSub)
	id, err := effectstest.CastWithTapParams(t, g, name, typeLine, cost, oracle,
		game.CastSpellParams{CostBranch: effectstest.Branch(effectstest.RevealBranchIdx), RevealIDs: []uuid.UUID{pal}})
	if err != nil {
		t.Fatalf("%s: CastSpell: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, id) {
		t.Fatalf("%s did not resolve", name)
	}
	return id
}

func TestSadisticSkymarcherHasFlyingAndLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := castCreatureThroughReveal(t, g, "Sadistic Skymarcher", "Creature — Vampire Soldier", "{2}{B}", effectstest.SkymarcherOracle, "Vampire")
	effectstest.AssertKeywords(t, g, id, "flying", "lifelink")
}

func TestSilvergillAdeptDrawsOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Land"})
	hand := me.Hand.Size()
	castCreatureThroughReveal(t, g, "Silvergill Adept", "Creature — Merfolk Wizard", "{1}{U}", effectstest.AdeptOracle, "Merfolk")
	// The helper adds a pal (+1) and the Adept (+1), the cast removes the
	// Adept (-1), and the enters trigger draws one (+1).
	if got := me.Hand.Size(); got != hand+2 {
		t.Fatalf("hand %d -> %d, want +2 (the pal and the drawn card)", hand, got)
	}
}

func TestSqueakingPieSneakHasFear(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := castCreatureThroughReveal(t, g, "Squeaking Pie Sneak", "Creature — Goblin Rogue", "{1}{B}", effectstest.PieSneakOracle, "Goblin")
	effectstest.AssertKeywords(t, g, id, "fear")
}

func TestWrensRunVanquisherHasDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := castCreatureThroughReveal(t, g, "Wren's Run Vanquisher", effectstest.VanquisherTypeLine, effectstest.VanquisherManaCost, effectstest.VanquisherOracle, "Elf")
	effectstest.AssertKeywords(t, g, id, "deathtouch")
}

func TestVanillaRevealCreaturesResolve(t *testing.T) {
	for _, c := range []struct{ name, typeLine, cost, oracle, sub string }{
		{"Daring Buccaneer", "Creature — Human Pirate", "{R}", effectstest.RevealPlainOracle, "Pirate"},
		{"Flamekin Bladewhirl", "Creature — Elemental Warrior", "{R}", effectstest.BladewhirlOracle, "Elemental"},
		{"Goldmeadow Stalwart", "Creature — Kithkin Soldier", "{W}", effectstest.GoldmeadowOracle, "Kithkin"},
	} {
		g := effectstest.NewCatalogGame(t)
		castCreatureThroughReveal(t, g, c.name, c.typeLine, c.cost, c.oracle, c.sub)
	}
}

func TestThunderherdMigrationFetchesABasicTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	pal := effectstest.HandCard(me, "Rex", "Creature — Dinosaur")
	if _, err := effectstest.CastWithTapParams(t, g, "Thunderherd Migration", "Sorcery", "{1}{G}", effectstest.ThunderherdOracle,
		game.CastSpellParams{CostBranch: effectstest.Branch(effectstest.RevealBranchIdx), RevealIDs: []uuid.UUID{pal}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if c, ok := effectstest.BattlefieldCard(g, forest); !ok || !c.Tapped {
		t.Fatalf("the basic land should be on the battlefield tapped (found %v, %+v)", ok, c)
	}
}

func TestCausticExhaleGivesMinusThreeMinusThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	victim := effectstest.PushVanillaCreature(g, foe.ID, "Ogre", 3, 3)
	dragon := effectstest.HandCard(me, "Whelp", "Creature — Dragon")
	if _, err := effectstest.CastWithTapParams(t, g, "Caustic Exhale", "Instant", "{B}", effectstest.CausticExhaleOracle,
		game.CastSpellParams{CostBranch: effectstest.Branch(effectstest.RevealBranchIdx), RevealIDs: []uuid.UUID{dragon},
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, victim) {
		t.Fatal("the 3/3 survived -3/-3")
	}
}

func TestKinsbaileAspirantGrowsWhenAnotherCreatureEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	aspirant := effectstest.PushCatalogPermanent(g, me.ID, "Kinsbaile Aspirant", "Creature — Kithkin Citizen", effectstest.AspirantOracle, false)
	before := effectstest.EffectivePower(t, g, aspirant)
	effectstest.CastWithTapParams(t, g, "Daring Buccaneer", "Creature — Human Pirate", "{R}", effectstest.RevealPlainOracle,
		game.CastSpellParams{CostBranch: effectstest.Branch(effectstest.ManaBranchIdx)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, aspirant); got != before+1 {
		t.Fatalf("power %d -> %d, want +1 for the creature that entered", before, got)
	}
}

func TestLysAlanaDignitaryMakesManaOnlyWithAnElfInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dignitary := effectstest.PushCatalogPermanent(g, me.ID, "Lys Alana Dignitary", "Creature — Elf Advisor", effectstest.LysAlanaOracle, false)
	if err := g.ActivateManaAbility(me.ID, dignitary, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("the mana ability activated with no Elf card in the graveyard")
	}
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Elf", TypeLine: "Creature — Elf", Owner: me.ID, Controller: me.ID})
	if err := g.ActivateManaAbility(me.ID, dignitary, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate with an Elf in the graveyard: %v", err)
	}
	if n := me.ManaPool.Total(); n != 2 {
		t.Fatalf("%d mana in the pool, want {G}{G}", n)
	}
}

func TestMudbuttonCursetosserCantBlockAndItsDeathDestroysASmallCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	curse := effectstest.PushCatalogPermanent(g, me.ID, "Mudbutton Cursetosser", "Creature — Goblin Warlock", effectstest.CursetosserOracle, false)
	small := effectstest.PushVanillaCreature(g, foe.ID, "Small", 2, 2)
	big := effectstest.PushVanillaCreature(g, foe.ID, "Big", 3, 3)
	g.ReadSnapshot(func() { g.RecomputeLayersIfStaleLocked() })
	c, ok := effectstest.BattlefieldCard(g, curse)
	if !ok || !game.Restricted(&c, game.CantBlock) {
		t.Fatal("the Cursetosser can block")
	}
	if err := g.DestroyPermanentForEffect(curse); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	if latest := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID); latest != nil {
		if err := g.ResolvePickTarget(latest.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: small}); err != nil {
			t.Fatalf("pick target: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, small) {
		t.Fatal("the power-2 creature survived the Cursetosser's death trigger")
	}
	if !effects.OnBattlefield(g, big) {
		t.Fatal("the power-3 creature should not be a legal target")
	}
}

func TestSoulbrightSeekerAddsManaOnTheThirdResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	seeker := effectstest.PushCatalogPermanent(g, me.ID, "Soulbright Seeker", "Creature — Elemental Sorcerer", effectstest.SoulbrightOracle, false)
	for i := 1; i <= 3; i++ {
		if err := g.ActivateCatalogAbility(me.ID, seeker, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: seeker}},
		}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		got := me.ManaPool.Total()
		if i < 3 && got != 0 {
			t.Fatalf("activation %d added mana", i)
		}
		if i == 3 && got != 4 {
			t.Fatalf("third resolution left %d mana in the pool, want {R}{R}{R}{R}", got)
		}
	}
	effectstest.AssertKeywords(t, g, seeker, "trample")
}

func TestSurtlandElementalistOffersAFreeInstantOrSorceryOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	giant := effectstest.PushCatalogPermanent(g, me.ID, "Surtland Elementalist", "Creature — Giant Wizard", effectstest.SurtlandOracle, false)
	instant := effectstest.PushHandCardWithManaCost(me, "Big Instant", "Instant", "{6}")
	creature := effectstest.PushHandCardWithManaCost(me, "Big Creature", "Creature — Bear", "{1}")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(giant, foe.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	// #4250: the free cast is offered inside the attack trigger's resolution.
	offer := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, instant)
	if effectstest.HasID(offer.DiscardOptions, creature) {
		t.Fatalf("offered %v, want only the instant", offer.DiscardOptions)
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, instant, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the free six-mana instant: %v", err)
	}
	effectstest.CastInsideTheResolution(t, g, instant)
}
