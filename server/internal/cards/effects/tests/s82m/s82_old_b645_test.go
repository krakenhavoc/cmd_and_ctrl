package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b645_test.go — Invasion, "Activated abilities" (S82 OLD batch 645, #5616): each
// card driven through a real activation, cast or trigger.

func oldB645AllFull(t *testing.T, ids ...string) {
	t.Helper()
	for _, id := range ids {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB645Ready(g *game.Game, id uuid.UUID) {
	effectstest.UntapForTest(g, id)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
}

// oldB645Perm seeds a ready catalog permanent for the active seat.
func oldB645Perm(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := effectstest.SeedPermanentWithOracle(g, g.Seats[0].ID, name, typeLine, oracle)
	oldB645Ready(g, id)
	return id
}

func oldB645Activate(t *testing.T, g *game.Game, src uuid.UUID, ability int, mana string, params game.ActivateAbilityParams) error {
	t.Helper()
	me := g.Seats[0]
	if mana != "" {
		effectstest.FloatMana(t, g, me, mana)
	}
	params.Strict = true
	return g.ActivateCatalogAbility(me.ID, src, ability, params)
}

func oldB645Effective(g *game.Game, id uuid.UUID) game.Card {
	c, _ := g.LookupCardForEffect(id)
	return c
}

func oldB645Colored(g *game.Game, owner uuid.UUID, name, color string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", Colors: []string{color},
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

func oldB645Target(id uuid.UUID) game.ActivateAbilityParams {
	return game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}}}
}

const (
	oldB645NomadicElf    = "025bbfc8-bccd-4062-b758-7f9e1f8ad420"
	oldB645Aerostat      = "093daf01-d83b-4d49-a6b3-ec9a799b9582"
	oldB645LotusGuardian = "143d7a32-c15d-4a1f-ad81-cf1e6aa92bf3"
	oldB645SkyWeaver     = "25933e8b-fa48-4152-9eae-388aef10ad37"
	oldB645Cavalry       = "3e6e70eb-6638-4aa6-997b-90f5810e12a0"
	oldB645Apprentice    = "429de3c3-3f2e-4d55-a04a-f414912aff3a"
	oldB645Armored       = "46752289-01e8-4839-8889-49db1efd2e1a"
	oldB645Ranger        = "4ceb49f3-a997-468d-9653-8ca5ceb2bcfb"
	oldB645HoodedKavu    = "52934f16-eb9b-4d0a-b372-155461660f6e"
	oldB645Hypnotist     = "586d64c7-6aa9-42a0-9d8a-9dece52bc5d6"
	oldB645AncientKavu   = "598a7347-b772-4839-a02f-b930076c1354"
)

func TestS82OldB645AllAreFullWithNoCaveats(t *testing.T) {
	oldB645AllFull(t, oldB645NomadicElf, oldB645Aerostat, oldB645LotusGuardian, oldB645SkyWeaver, oldB645Cavalry,
		oldB645Apprentice, oldB645Armored, oldB645Ranger, oldB645HoodedKavu, oldB645Hypnotist, oldB645AncientKavu)
}

func TestS82OldB645NomadicElfPaysTwoForOneManaOfAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elf := oldB645Perm(t, g, "Nomadic Elf", "Creature — Elf Nomad", oldB645NomadicElf)
	effectstest.FloatMana(t, g, me, "{C}{G}")
	effectstest.ActivateManaFor(t, g, me.ID, elf, 0, game.ManaAbilityParams{})
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil {
		t.Fatal("no colour pick was asked for")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("the pool holds %d mana, want exactly the one produced", me.ManaPool.Total())
	}
	if effectstest.TapCostTapped(g, elf) {
		t.Error("the ability has no {T} in its cost")
	}
}

func TestS82OldB645LotusGuardianTapsForAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lotus := oldB645Perm(t, g, "Lotus Guardian", "Artifact Creature — Dragon", oldB645LotusGuardian)
	effectstest.TapForMana(t, g, me.ID, lotus)
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil {
		t.Fatal("no colour pick was asked for")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("the pool holds %d mana, want 1", me.ManaPool.Total())
	}
	if !effectstest.P1g2HasKeyword(g, lotus, "flying") {
		t.Error("the Guardian has no flying")
	}
}

func TestS82OldB645SelfGrantRows(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle, mana, keyword string }{
		{"Llanowar Cavalry", "Creature — Human Soldier", oldB645Cavalry, "{W}", "vigilance"},
		{"Hooded Kavu", "Creature — Kavu", oldB645HoodedKavu, "{B}", "fear"},
	} {
		g := effectstest.NewCatalogGame(t)
		id := oldB645Perm(t, g, tc.name, tc.typeLine, tc.oracle)
		if effectstest.P1g2HasKeyword(g, id, tc.keyword) {
			t.Fatalf("%s starts with %s", tc.name, tc.keyword)
		}
		if err := oldB645Activate(t, g, id, 0, "", game.ActivateAbilityParams{}); err == nil {
			t.Errorf("%s: activated without mana", tc.name)
		}
		if err := oldB645Activate(t, g, id, 0, tc.mana, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("%s: activate: %v", tc.name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.P1g2HasKeyword(g, id, tc.keyword) {
			t.Errorf("%s did not gain %s", tc.name, tc.keyword)
		}
	}
}

func TestS82OldB645SkyWeaverGivesFlyingToWhiteOrBlackCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	weaver := oldB645Perm(t, g, "Sky Weaver", "Creature — Metathran Wizard", oldB645SkyWeaver)
	white := oldB645Colored(g, me.ID, "Knight", "W")
	green := oldB645Colored(g, me.ID, "Bear", "G")
	if err := oldB645Activate(t, g, weaver, 0, "{C}{C}", oldB645Target(green)); err == nil {
		t.Fatal("a green creature was a legal target")
	}
	if err := oldB645Activate(t, g, weaver, 0, "", oldB645Target(white)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P1g2HasKeyword(g, white, "flying") {
		t.Error("the white creature did not gain flying")
	}
}

func TestS82OldB645NightscapeApprenticeHasTwoTapAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	app := oldB645Perm(t, g, "Nightscape Apprentice", "Creature — Zombie Wizard", oldB645Apprentice)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	libBefore := me.Library.Size()
	if err := oldB645Activate(t, g, app, 0, "{U}", oldB645Target(theirs)); err == nil {
		t.Fatal("an opponent's creature was a legal target for the library tuck")
	}
	if err := oldB645Activate(t, g, app, 0, "", oldB645Target(mine)); err != nil {
		t.Fatalf("tuck: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("the creature stayed on the battlefield")
	}
	if me.Library.Size() != libBefore+1 || me.Library.Cards[0].InstanceID != mine && me.Library.Cards[len(me.Library.Cards)-1].InstanceID != mine {
		t.Errorf("the creature is not on the library: size %d->%d", libBefore, me.Library.Size())
	}
	oldB645Ready(g, app)
	if err := oldB645Activate(t, g, app, 1, "{R}", oldB645Target(theirs)); err != nil {
		t.Fatalf("first strike: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P1g2HasKeyword(g, theirs, "first strike") {
		t.Error("no first strike")
	}
}

func TestS82OldB645ArmoredGuardianProtectsAndGainsShroud(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	guard := oldB645Perm(t, g, "Armored Guardian", "Creature — Cat Soldier", oldB645Armored)
	theirs := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	if err := oldB645Activate(t, g, guard, 0, "{C}{W}{W}", oldB645Target(theirs)); err == nil {
		t.Fatal("an opponent's creature was a legal target")
	}
	if err := oldB645Activate(t, g, guard, 0, "", oldB645Target(guard)); err != nil {
		t.Fatalf("protection: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "R")
	if !effectstest.P1g2HasKeyword(g, guard, game.ProtectionFromColor("R")) {
		t.Error("no protection from red")
	}
	if err := oldB645Activate(t, g, guard, 1, "{C}{U}{U}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("shroud: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P1g2HasKeyword(g, guard, "shroud") {
		t.Error("no shroud")
	}
}

func TestS82OldB645AncientKavuBecomesColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	kavu := oldB645Perm(t, g, "Ancient Kavu", "Creature — Kavu", oldB645AncientKavu)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == kavu {
				g.Battlefield.Cards[i].Colors = []string{"R"}
			}
		}
	})
	if !oldB645Effective(g, kavu).HasColor("R") {
		t.Fatal("fixture: the Kavu is not red")
	}
	if err := oldB645Activate(t, g, kavu, 0, "{C}{C}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB645Effective(g, kavu).HasColor("R") {
		t.Error("the Kavu is still red")
	}
}

func TestS82OldB645VodalianHypnotistDiscardsAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	hyp := oldB645Perm(t, g, "Vodalian Hypnotist", "Creature — Merfolk Wizard", oldB645Hypnotist)
	discardMe := effectstest.HandSpell(opp, "Spare", "Sorcery", "")
	hand := opp.Hand.Size()
	if err := oldB645Activate(t, g, hyp, 0, "", game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Fatal("activated without paying")
	}
	if err := oldB645Activate(t, g, hyp, 0, "{C}{C}{B}", game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, opp.ID, discardMe)
	if opp.Hand.Size() != hand-1 {
		t.Errorf("hand %d, want %d", opp.Hand.Size(), hand-1)
	}
}

func TestS82OldB645FirebrandRangerPutsABasicLandFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ranger := oldB645Perm(t, g, "Firebrand Ranger", "Creature — Human Soldier Ranger", oldB645Ranger)
	basic := effectstest.HandCardForTest(me, "Forest", "Basic Land — Forest", "")
	nonbasic := effectstest.HandCardForTest(me, "Hideout", "Land", "")
	if err := oldB645Activate(t, g, ranger, 0, "{G}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, basic)
	if !g.Battlefield.Contains(basic) {
		t.Error("the basic land did not enter")
	}
	if g.Battlefield.Contains(nonbasic) {
		t.Error("the nonbasic land entered")
	}
}

func TestS82OldB645MetathranAerostatPutsAnXCreatureAndReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	aero := oldB645Perm(t, g, "Metathran Aerostat", "Creature — Metathran", oldB645Aerostat)
	two := effectstest.HandCardFull(me, "Two Drop", "Creature — Bear", "{1}{G}", "", []string{"G"})
	three := effectstest.HandCardFull(me, "Three Drop", "Creature — Bear", "{2}{G}", "", []string{"G"})
	effectstest.FloatMana(t, g, me, "{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, aero, 0, game.ActivateAbilityParams{Strict: true, XValue: 2}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick was offered")
	}
	for _, id := range pick.ChooseCards {
		if id == three {
			t.Error("the mana value 3 creature was offered for X = 2")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, two)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(two) {
		t.Error("the creature card did not enter")
	}
	if g.Battlefield.Contains(aero) {
		t.Error("the Aerostat stayed on the battlefield")
	}
	if _, ok := effectstest.HandCardByID(g, me, aero); !ok {
		t.Error("the Aerostat is not in its owner's hand")
	}
}

func TestS82OldB645MetathranAerostatStaysWhenNothingEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	aero := oldB645Perm(t, g, "Metathran Aerostat", "Creature — Metathran", oldB645Aerostat)
	effectstest.HandCardFull(me, "Three Drop", "Creature — Bear", "{2}{G}", "", []string{"G"})
	effectstest.FloatMana(t, g, me, "{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, aero, 0, game.ActivateAbilityParams{Strict: true, XValue: 1}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aero) {
		t.Error("the Aerostat left although no creature entered")
	}
}
