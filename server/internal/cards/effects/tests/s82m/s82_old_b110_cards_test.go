package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b110_cards_test.go — Core Set 2021, batch 110 (#5511). Every test
// plays the real catalog entry through the engine.

const (
	o110Idol       = "6b946317-8a66-41b7-b902-308784d69b90"
	o110Overload   = "7851e4c8-020c-4a3d-88ed-0c97609cdfed"
	o110Arcanist   = "1d07856b-d9ef-418c-bd06-3888dbcaddf4"
	o110Sanctum    = "a1d8fcb9-26b2-4d09-9db7-e644294749f8"
	o110SeeTheTrue = "b23c1f40-291d-43b3-971d-97ded8b367cb"
	o110Bear       = "Creature — Bear"
)

func TestS82OldB110EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{o110Idol, o110Overload, o110Arcanist, o110Sanctum, o110SeeTheTrue} {
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

func o110Cast(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, p game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: me.ID, Controller: me.ID})
	})
	p.Strict = true
	return id, g.CastSpell(me.ID, id, p)
}

func o110MustCast(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, p game.CastSpellParams) uuid.UUID {
	t.Helper()
	id, err := o110Cast(t, g, name, typeLine, cost, oracle, p)
	if err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

func o110Pay(t *testing.T, g *game.Game, spec string) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, g.Seats[g.Turn.ActiveSeat], spec)
}

func o110Grave(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return id
}

func o110Zone(g *game.Game, id uuid.UUID) game.ZoneKind {
	if z := g.FindCardZoneForEffect(id); z != nil {
		return z.Kind
	}
	return ""
}

// --- Idol of Endurance -------------------------------------------------------------------

func TestS82OldB110IdolExilesSmallCreaturesUntilItLeavesThenTheyReturnToTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
	small := o110Grave(me, "Small", o110Bear, "{2}")
	three := o110Grave(me, "Three", o110Bear, "{1}{G}{G}")
	big := o110Grave(me, "Big", o110Bear, "{2}{G}{G}")
	spell := o110Grave(me, "Spell", "Instant", "{1}")
	theirs := o110Grave(opp, "Theirs", o110Bear, "{1}")
	idol := effectstest.EnterFromHand(t, g, me.ID, "Idol of Endurance", "Artifact", o110Idol)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{small, three} {
		if o110Zone(g, id) != game.ZoneExile {
			t.Errorf("a creature card with mana value 3 or less is in %v, want exile", o110Zone(g, id))
		}
	}
	for name, id := range map[string]uuid.UUID{"big": big, "spell": spell, "theirs": theirs} {
		if o110Zone(g, id) == game.ZoneExile {
			t.Errorf("the %s card was exiled", name)
		}
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(idol, game.DestroyOptions{}); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	g.RunStateChecksForTest()
	for _, id := range []uuid.UUID{small, three} {
		if o110Zone(g, id) != game.ZoneGraveyard {
			t.Errorf("an exiled card is in %v after the Idol left, want graveyard", o110Zone(g, id))
		}
	}
	if !me.Graveyard.Contains(small) {
		t.Error("the card did not return to its owner's graveyard")
	}
}

func TestS82OldB110IdolOpensOneFreeCreatureCastFromItsExiledCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := o110Grave(me, "Alpha", o110Bear, "{2}{G}")
	b := o110Grave(me, "Beta", o110Bear, "{G}")
	idol := effectstest.EnterFromHand(t, g, me.ID, "Idol of Endurance", "Artifact", o110Idol)
	effectstest.PassPriorityAroundTable(t, g)
	// Not before the ability: with no mana the card cannot be cast.
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, a, game.CastSpellParams{FromZone: string(game.ZoneExile), Strict: true}); err == nil {
		t.Fatal("an exiled card was cast before the Idol's ability")
	}
	effectstest.FloatForTest(g, me, "W")
	if err := g.ActivateCatalogAbility(me.ID, idol, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated for {W}")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "WW")
	if err := g.ActivateCatalogAbility(me.ID, idol, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, a, game.CastSpellParams{FromZone: string(game.ZoneExile), Strict: true}); err != nil {
		t.Fatalf("free cast from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if o110Zone(g, a) != game.ZoneBattlefield {
		t.Errorf("the cast creature is in %v, want the battlefield", o110Zone(g, a))
	}
	if err := g.CastSpell(me.ID, b, game.CastSpellParams{FromZone: string(game.ZoneExile), Strict: true}); err == nil {
		t.Error("a second creature spell was cast: the permission is for one")
	}
}

// --- Vodalian Arcanist -------------------------------------------------------------------

func TestS82OldB110ArcanistManaPaysOnlyForInstantsAndSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	arc := effectstest.PushCatalogPermanent(g, me.ID, "Vodalian Arcanist", "Creature — Merfolk Wizard", o110Arcanist, false)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.ActivateManaFor(t, g, me.ID, arc, 0, game.ManaAbilityParams{})
	if _, err := o110Cast(t, g, "Bear Cub", o110Bear, "{1}", "", game.CastSpellParams{}); err == nil {
		t.Fatal("the Arcanist's mana paid for a creature spell")
	}
	if _, err := o110Cast(t, g, "Cheap Sorcery", "Sorcery", "{1}", "", game.CastSpellParams{}); err != nil {
		t.Fatalf("the Arcanist's mana could not pay for a sorcery: %v", err)
	}
}

// --- Experimental Overload ---------------------------------------------------------------

func TestS82OldB110ExperimentalOverloadMakesAWeirdReturnsACardAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	o110Grave(me, "Bolt", "Instant", "{R}")
	back := o110Grave(me, "Divination", "Sorcery", "{2}{U}")
	o110Grave(me, "Bear", o110Bear, "{1}")
	o110Pay(t, g, "UURR")
	spell := o110MustCast(t, g, "Experimental Overload", "Sorcery", "{2}{U}{R}", o110Overload, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	weirds := effectstest.TokensNamed(g, "Weird")
	if len(weirds) != 1 {
		t.Fatalf("%d Weird tokens, want 1", len(weirds))
	}
	if p, tough := effectstest.PtOf(t, g, weirds[0].InstanceID); p != 2 || tough != 2 {
		t.Errorf("the Weird is %d/%d, want 2/2 (two instants and sorceries)", p, tough)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, back)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(back) {
		t.Error("the chosen sorcery did not return to hand")
	}
	if o110Zone(g, spell) != game.ZoneExile {
		t.Errorf("Experimental Overload is in %v, want exile", o110Zone(g, spell))
	}
}

func TestS82OldB110ExperimentalOverloadAllowsDecliningTheReturnAndMakesAZeroZeroWithNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_ = g.Seats[g.Turn.ActiveSeat]
	o110Pay(t, g, "UURR")
	spell := o110MustCast(t, g, "Experimental Overload", "Sorcery", "{2}{U}{R}", o110Overload, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if o110Zone(g, spell) != game.ZoneExile {
		t.Errorf("Experimental Overload is in %v, want exile", o110Zone(g, spell))
	}
	// A 0/0 Weird dies as a state-based action; nothing was returned.
	if n := len(effectstest.TokensNamed(g, "Weird")); n != 0 {
		t.Errorf("%d Weird tokens survived with X = 0", n)
	}
}

func TestS82OldB110ExperimentalOverloadRefusesAPoolThatCannotPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	o110Pay(t, g, "UU")
	if _, err := o110Cast(t, g, "Experimental Overload", "Sorcery", "{2}{U}{R}", o110Overload, game.CastSpellParams{}); err == nil {
		t.Error("Experimental Overload was cast without {R}")
	}
}

// --- Sanctum of Stone Fangs --------------------------------------------------------------

func TestS82OldB110SanctumDrainsForEachShrineInTheFirstMainPhase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Sanctum of Stone Fangs", "Legendary Enchantment — Shrine", o110Sanctum, false)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Other Shrine", TypeLine: "Enchantment — Shrine",
		Owner: me.ID, Controller: me.ID})
	opp := effectstest.Top100Opponent(g)
	// The game starts in my own turn; move to the next one and count from there.
	effectstest.AdvanceToNextSeatsTurn(t, g)
	myLife, theirLife := me.Life, opp.Life
	effectstest.AdvanceToMainOf(t, g, me.Seat)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife+2 {
		t.Errorf("life %d, want %d (two Shrines)", me.Life, myLife+2)
	}
	if opp.Life != theirLife-2 {
		t.Errorf("opponent life %d, want %d", opp.Life, theirLife-2)
	}
	for _, p := range g.Seats {
		if p != nil && p.ID != me.ID && p.Life != theirLife-2 {
			t.Errorf("an opponent has %d life, want %d", p.Life, theirLife-2)
		}
	}
}

// --- See the Truth -----------------------------------------------------------------------

func TestS82OldB110SeeTheTruthTakesOneAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedLibrary(me, "A", "B", "C", "D")
	o110Pay(t, g, "UU")
	o110MustCast(t, g, "See the Truth", "Sorcery", "{1}{U}", o110SeeTheTrue, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no pick opened")
	}
	if pick.ChooseMin != 1 || pick.ChooseMax != 1 || len(pick.ChooseCards) != 3 {
		t.Errorf("the pick offers %d cards (min %d max %d), want 3 of which exactly one is taken", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("no order prompt for the rest on the bottom")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, order.ScryCards, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if !me.Hand.Contains(ids[1]) {
		t.Error("the chosen card is not in hand")
	}
	if me.Hand.Contains(ids[0]) || me.Hand.Contains(ids[2]) {
		t.Error("an unchosen card went to hand")
	}
	if top, err := me.Library.Top(); err != nil || top.InstanceID != ids[3] {
		t.Errorf("the card below the look is not on top now")
	}
}

func TestS82OldB110SeeTheTruthPutsAllThreeInHandWhenCastFromAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.AdvanceToMain(t, g)
	spell := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "See the Truth", TypeLine: "Sorcery", ManaCost: "{1}{U}",
			OracleID: o110SeeTheTrue, Owner: me.ID, Controller: me.ID})
	})
	g.GrantCastPermissionOverCardForEffect(spell, game.CastPermission{Player: me.ID, Zone: game.ZoneGraveyard,
		Duration: game.WhileInZoneDuration(), Source: uuid.New()})
	effectstest.FloatForTest(g, me, "UU")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: string(game.ZoneGraveyard), Strict: true}); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 3; i++ {
		if !me.Hand.Contains(ids[i]) {
			t.Errorf("card %d of the look is not in hand", i)
		}
	}
	if me.Hand.Contains(ids[3]) {
		t.Error("the fourth card went to hand")
	}
}

func TestS82OldB110SeeTheTruthRefusesAPoolThatCannotPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	o110Pay(t, g, "U")
	if _, err := o110Cast(t, g, "See the Truth", "Sorcery", "{1}{U}", o110SeeTheTrue, game.CastSpellParams{}); err == nil {
		t.Error("See the Truth was cast for {U}")
	}
}
