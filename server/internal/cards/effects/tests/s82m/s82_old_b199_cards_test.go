package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b199_cards_test.go — Core Set 2019, batch 199 (S82 older sets,
// #5527). Every test plays the real catalog entry through the engine.

const (
	oldB199ChromiumOracle         = "58d8ff36-4a4d-40e0-b812-3d23b2c7606a"
	oldB199DryadGreenseekerOracle = "5c43c1d7-1899-45c1-992b-443873890d78"
	oldB199MistcallerOracle       = "808d5a67-8c38-42f5-9413-8771d8b4ae38"
	oldB199SkyriderPatrolOracle   = "104386d9-93b4-4a18-86d5-68718b474f4a"
	oldB199AjanisInfluenceOracle  = "2a3e3f3b-1f74-4ac7-b5ac-7c3cb9cc38a6"
	oldB199InfernalReckoningOracl = "32aa920c-0d91-4bc9-803c-d13857688d10"
	oldB199IsolateOracle          = "f42d4556-5be2-4861-9f88-1b630b43eff2"
	oldB199PalladiaMorsOracle     = "4c2b10ba-1930-4da8-9570-c89b2e239e3e"
	oldB199NoviceKnightOracle     = "916d2a0d-1c3e-4102-958d-524b6b4df509"
	oldB199ArisenGorgonOracle     = "69858691-1190-4454-af0a-87c2ac4ab6ec"
	oldB199StarCrownedStagOracle  = "a375b739-6a6c-4e07-9b1c-a304f84ea2cc"

	oldB199BearType = "Creature — Bear"
)

func TestOldB199EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		oldB199ChromiumOracle, oldB199DryadGreenseekerOracle, oldB199MistcallerOracle, oldB199SkyriderPatrolOracle,
		oldB199AjanisInfluenceOracle, oldB199InfernalReckoningOracl, oldB199IsolateOracle, oldB199PalladiaMorsOracle,
		oldB199NoviceKnightOracle, oldB199ArisenGorgonOracle, oldB199StarCrownedStagOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func oldB199LibraryTop(p *game.Player, cards ...game.Card) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		p.Library.PushTop(c)
		ids[i] = c.InstanceID
	}
	return ids
}

func oldB199Planeswalker(g *game.Game, owner uuid.UUID, subtype string) uuid.UUID {
	id := effectstest.B12Permanent(g, owner, subtype+" Test", "Legendary Planeswalker — "+subtype)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterLoyalty, 4) })
	return id
}

func TestOldB199ChromiumHasItsKeywordsAndCantBeCountered(t *testing.T) {
	spec, _ := effects.Lookup(oldB199ChromiumOracle)
	if !spec.CantBeCountered {
		t.Error("the spell can be countered")
	}
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Chromium, the Mutable", "Legendary Creature — Elder Dragon", oldB199ChromiumOracle, 7, 7)
	for _, kw := range []string{"flash", "flying"} {
		if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("no %s", kw)
		}
	}
}

func TestOldB199ChromiumDiscardsToBecomeAHexproofHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.SeatsForCombat(g)
	dragon := effectstest.B12Push(g, me.ID, "Chromium, the Mutable", "Legendary Creature — Elder Dragon", oldB199ChromiumOracle, 7, 7)
	blocker := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB199BearType, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, dragon, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no card discarded")
	}
	card := effectstest.PushCatalogHandCard(me, "Spare", "Sorcery", "")
	if err := g.ActivateCatalogAbility(me.ID, dragon, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{card}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(card) {
		t.Error("the card was not discarded as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, dragon); p != 1 || tough != 1 {
		t.Errorf("Chromium is %d/%d, want 1/1", p, tough)
	}
	if subs := effectstest.EffectiveSubtypes(t, g, dragon); len(subs) != 1 || subs[0] != "Human" {
		t.Errorf("Chromium's subtypes are %v, want only Human", subs)
	}
	if effectstest.HasEffectiveKeyword(t, g, dragon, "flying") {
		t.Error("Chromium kept flying after losing all abilities")
	}
	if !effectstest.HasEffectiveKeyword(t, g, dragon, "hexproof") {
		t.Error("Chromium has no hexproof")
	}
	if err := effectstest.PorBlock(t, g, dragon, opp.ID, blocker); err == nil {
		t.Error("Chromium was blocked")
	}
}

func TestOldB199DryadGreenseekerTakesATopLand(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		confirm        bool
		wantInHand     bool
	}{
		{"Forest", "Basic Land — Forest", true, true},
		{"Forest", "Basic Land — Forest", false, false},
		{"Bear", oldB199BearType, false, false},
	} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, _ := effectstest.SeatsForCombat(g)
		dryad := effectstest.B12Push(g, me.ID, "Dryad Greenseeker", "Creature — Dryad", oldB199DryadGreenseekerOracle, 1, 3)
		ids := oldB199LibraryTop(me, game.Card{Name: tc.name, TypeLine: tc.typeLine})
		if err := g.ActivateCatalogAbility(me.ID, dryad, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if tc.typeLine != oldB199BearType {
			effectstest.AnswerConfirmFor(t, g, me, tc.confirm)
		}
		if got := me.Hand.Contains(ids[0]); got != tc.wantInHand {
			t.Errorf("%s (confirm %v): in hand %v, want %v", tc.name, tc.confirm, got, tc.wantInHand)
		}
		if !tc.wantInHand && !me.Library.Contains(ids[0]) {
			t.Errorf("%s (confirm %v): the card did not stay in the library", tc.name, tc.confirm)
		}
	}
}

func TestOldB199MistcallerExilesCreaturesThatWereNotCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.SeatsForCombat(g)
	mist := effectstest.B12Push(g, me.ID, "Mistcaller", "Creature — Merfolk Wizard", oldB199MistcallerOracle, 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, mist, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(mist) {
		t.Error("Mistcaller was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	// A creature put onto the battlefield is exiled instead.
	put := effectstest.PushCatalogHandCard(opp, "Put Bear", oldB199BearType, "")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, put); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if g.Battlefield.Contains(put) || !g.Exile.Contains(put) {
		t.Error("a creature that wasn't cast entered instead of being exiled")
	}
	// A creature that was cast still enters.
	cast := effectstest.CastAndResolveCreature(t, g, "Cast Bear", oldB199BearType, "")
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cast) {
		t.Error("a cast creature was exiled")
	}
	// The effect ends with the turn.
	effectstest.AdvanceToMainOf(t, g, 1)
	later := effectstest.PushCatalogHandCard(opp, "Later Bear", oldB199BearType, "")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, later); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if !g.Battlefield.Contains(later) {
		t.Error("the replacement outlived the turn")
	}
}

func TestOldB199SkyriderPatrolPaysForACounterAndFlying(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.B12Push(g, me.ID, "Skyrider Patrol", "Creature — Elf Scout", oldB199SkyriderPatrolOracle, 2, 3)
		bear := effectstest.B12Creature(g, me.ID, "Bear", oldB199BearType, 2, 2)
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		effectstest.B06AddMana(me, "G", "U")
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		if pay {
			effectstest.AnswerPickTarget(t, g, bear)
		}
		effectstest.PassPriorityAroundTable(t, g)
		wantCounters := 0
		if pay {
			wantCounters = 1
		}
		if n := effectstest.CounterCount(g, bear, game.CounterPlusOne); n != wantCounters {
			t.Errorf("pay %v: Bear has %d counters, want %d", pay, n, wantCounters)
		}
		if got := effectstest.HasEffectiveKeyword(t, g, bear, "flying"); got != pay {
			t.Errorf("pay %v: Bear flying = %v", pay, got)
		}
	}
}

func TestOldB199AjanisInfluenceCountersAndTakesAWhiteCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", oldB199BearType, 2, 2)
	ids := oldB199LibraryTop(me,
		game.Card{Name: "Red Card", TypeLine: "Sorcery", ManaCost: "{R}", Colors: []string{"R"}},
		game.Card{Name: "White Card", TypeLine: "Sorcery", ManaCost: "{W}", Colors: []string{"W"}},
		game.Card{Name: "Land", TypeLine: "Land"}, game.Card{Name: "Land", TypeLine: "Land"}, game.Card{Name: "Land", TypeLine: "Land"},
		game.Card{Name: "Deep White", TypeLine: "Sorcery", ManaCost: "{W}", Colors: []string{"W"}})
	red, white, deep := ids[0], ids[1], ids[5]
	effectstest.CastCatalogSpell(t, g, "Ajani's Influence", "Sorcery", oldB199AjanisInfluenceOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, bear, game.CounterPlusOne); n != 2 {
		t.Errorf("Bear has %d +1/+1 counters, want 2", n)
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to take a white card")
	}
	for _, id := range pick.ChooseCards {
		if id == red || id == deep {
			t.Errorf("offered a card it may not take: %s", id)
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, white)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(white) {
		t.Error("the white card is not in hand")
	}
}

func TestOldB199InfernalReckoningExilesAColorlessCreatureForItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	golem := effectstest.B12Creature(g, opp.ID, "Golem", "Artifact Creature — Golem", 4, 4)
	green := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Green Bear", TypeLine: oldB199BearType, Power: 2, Toughness: 2, Colors: []string{"G"}})
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Infernal Reckoning", "Instant", oldB199InfernalReckoningOracl, effectstest.CardTarget(green)); err == nil {
		t.Fatal("a green creature was a legal target")
	}
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Infernal Reckoning", "Instant", oldB199InfernalReckoningOracl, effectstest.CardTarget(golem))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(golem) || !g.Exile.Contains(golem) {
		t.Error("the Golem was not exiled")
	}
	if me.Life != life+4 {
		t.Errorf("life %d -> %d, want +4", life, me.Life)
	}
}

func TestOldB199IsolateExilesOnlyAManaValueOnePermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	one := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "One Drop", TypeLine: oldB199BearType, ManaCost: "{G}", Power: 1, Toughness: 1})
	two := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Two Drop", TypeLine: oldB199BearType, ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	land := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	for _, bad := range []uuid.UUID{two, land} {
		if err := effectstest.CastCatalogSpellErrStrict(t, g, "Isolate", "Instant", oldB199IsolateOracle, effectstest.CardTarget(bad)); err == nil {
			t.Fatal("a permanent without mana value 1 was a legal target")
		}
	}
	effectstest.CastCatalogSpell(t, g, "Isolate", "Instant", oldB199IsolateOracle, effectstest.CardTarget(one))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(one) || !g.Exile.Contains(one) {
		t.Error("the mana value 1 permanent was not exiled")
	}
}

func TestOldB199PalladiaMorsIsHexproofUntilItDealsDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	dragon := effectstest.B12Push(g, me.ID, "Palladia-Mors, the Ruiner", "Legendary Creature — Elder Dragon", oldB199PalladiaMorsOracle, 6, 6)
	for _, kw := range []string{"flying", "vigilance", "trample", "hexproof"} {
		if !effectstest.HasEffectiveKeyword(t, g, dragon, kw) {
			t.Errorf("no %s before dealing damage", kw)
		}
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(dragon, opp.ID, 1); err != nil {
			t.Fatalf("deal damage: %v", err)
		}
	})
	if effectstest.HasEffectiveKeyword(t, g, dragon, "hexproof") {
		t.Error("still hexproof after dealing damage")
	}
	if !effectstest.HasEffectiveKeyword(t, g, dragon, "flying") {
		t.Error("lost flying")
	}
}

func TestOldB199NoviceKnightAttacksOnlyWhenEnchantedOrEquipped(t *testing.T) {
	for _, attach := range []string{"", "Enchantment — Aura", "Artifact — Equipment"} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		knight := effectstest.B12Push(g, me.ID, "Novice Knight", "Creature — Human Knight", oldB199NoviceKnightOracle, 2, 3)
		if attach != "" {
			att := effectstest.B12Permanent(g, me.ID, "Attachment", attach)
			effectstest.BrAttach(t, g, att, knight)
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		err := g.DeclareAttacker(knight, opp.ID)
		if (err == nil) != (attach != "") {
			t.Errorf("attachment %q: DeclareAttacker err = %v", attach, err)
		}
	}
}

func TestOldB199ArisenGorgonHasDeathtouchWithALiliana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	gorgon := effectstest.B12Push(g, me.ID, "Arisen Gorgon", "Creature — Zombie Gorgon", oldB199ArisenGorgonOracle, 3, 3)
	if effectstest.HasEffectiveKeyword(t, g, gorgon, "deathtouch") {
		t.Error("deathtouch with no planeswalker")
	}
	oldB199Planeswalker(g, me.ID, "Ajani")
	if effectstest.HasEffectiveKeyword(t, g, gorgon, "deathtouch") {
		t.Error("deathtouch with an Ajani")
	}
	oldB199Planeswalker(g, g.Seats[1].ID, "Liliana")
	if effectstest.HasEffectiveKeyword(t, g, gorgon, "deathtouch") {
		t.Error("deathtouch with an opponent's Liliana")
	}
	oldB199Planeswalker(g, me.ID, "Liliana")
	if !effectstest.HasEffectiveKeyword(t, g, gorgon, "deathtouch") {
		t.Error("no deathtouch with my Liliana")
	}
}

func TestOldB199StarCrownedStagTapsADefendersCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	stag := effectstest.B12Push(g, me.ID, "Star-Crowned Stag", "Creature — Elk", oldB199StarCrownedStagOracle, 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB199BearType, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", oldB199BearType, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(stag, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldTapped(g, theirs) {
		t.Error("the defending player's creature was not tapped")
	}
	if effectstest.BattlefieldTapped(g, mine) {
		t.Error("my own creature was tapped")
	}
}
