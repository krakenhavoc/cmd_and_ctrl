package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kaldheim batch 072 (S82 OLD, #5505).

const (
	oldB072Shackles  = "3466e0be-3e59-4f5e-812e-3250c8bf19aa"
	oldB072Absorb    = "4489f362-9b7e-4dc0-a459-a990a487454e"
	oldB072Greatness = "a7b45840-505d-4b7d-a084-549be08b0c20"
	oldB072Crew      = "ced584a8-7548-4b4e-8ccf-04b66d255539"
	oldB072Reaper    = "d72489b5-a213-4bd8-a519-8caeb1acea5f"
)

func TestOldB072CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB072Shackles: "Shackles of Treachery", oldB072Absorb: "Absorb Identity", oldB072Greatness: "In Search of Greatness",
		oldB072Crew: "Reckless Crew", oldB072Reaper: "Renegade Reaper",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB072Hand(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Hand.PushTop(c)
	return c.InstanceID
}

func oldB072Attach(g *game.Game, owner, host uuid.UUID, name, typeLine string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Owner: owner, Controller: owner, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host},
	})
}

// --- Renegade Reaper ---------------------------------------------------------

func TestOldB072RenegadeReaperGainsFourLifeOnlyWhenAnAngelIsMilled(t *testing.T) {
	for _, angel := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		deep := effectstest.SeedLibraryTop(me, "Deep Angel", "Creature — Angel", "{W}")
		tl := "Creature — Bear"
		if angel {
			tl = "Creature — Angel"
		}
		effectstest.SeedLibraryTop(me, "Maybe Angel", tl, "{W}")
		for i := 0; i < 3; i++ {
			effectstest.SeedLibraryTop(me, "Filler", "Creature — Bear", "{G}")
		}
		life := me.Life
		effectstest.CastCatalogSpell(t, g, "Renegade Reaper", "Creature — Angel Berserker", oldB072Reaper, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if n := len(me.Graveyard.Cards); n != 4 {
			t.Fatalf("%d cards milled, want 4", n)
		}
		if !me.Library.Contains(deep) {
			t.Error("a fifth card was milled")
		}
		want := life
		if angel {
			want += 4
		}
		if me.Life != want {
			t.Errorf("angel=%v: life %d, want %d", angel, me.Life, want)
		}
	}
}

// --- Shackles of Treachery ---------------------------------------------------

func TestOldB072ShacklesStealsUntapsHastesAndDestroysAnAttachedEquipmentOnDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(ogre) })
	sword := oldB072Attach(g, opp.ID, ogre, "Their Sword", "Artifact — Equipment")
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
	otherSword := oldB072Attach(g, opp.ID, other, "Other Sword", "Artifact — Equipment")
	effectstest.CastCatalogSpell(t, g, "Shackles of Treachery", "Sorcery", oldB072Shackles, effectstest.CardRefs(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.GaLayered(t, g, ogre)
	if c.Controller != me.ID {
		t.Fatal("the creature was not stolen")
	}
	if c.Tapped {
		t.Error("the creature was not untapped")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, ogre, "haste") {
		t.Error("the creature has no haste")
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(ogre, opp.ID, 3) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if effectstest.HasID(p.PickTargetCards, otherSword) {
			t.Fatal("an Equipment attached to another creature was offered")
		}
		effectstest.PickCard(t, g, me.ID, sword)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(sword) {
		t.Error("the attached Equipment was not destroyed")
	}
	if !g.Battlefield.Contains(otherSword) {
		t.Error("an Equipment attached to another creature was destroyed")
	}
}

func TestOldB072ShacklesDamageWithNoEquipmentAttachedDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Shackles of Treachery", "Sorcery", oldB072Shackles, effectstest.CardRefs(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(ogre, opp.ID, 3) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d prompts for a trigger with no Equipment to target", len(g.PendingChoices))
	}
	_ = me
}

// --- Absorb Identity ---------------------------------------------------------

func TestOldB072AbsorbIdentityBouncesAndTurnsShapeshiftersIntoCopies(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		changeling := effectstest.B12Creature(g, me.ID, "Mask", "Creature — Shapeshifter", 1, 1)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 5, 5)
		effectstest.CastCatalogSpell(t, g, "Absorb Identity", "Instant", oldB072Absorb, effectstest.CardRefs(theirs))
		effectstest.PassPriorityAroundTable(t, g)
		if !opp.Hand.Contains(theirs) {
			t.Fatal("the creature was not returned to its owner's hand")
		}
		c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID)
		if c == nil {
			t.Fatal("no may question")
		}
		if err := g.ResolveConfirm(c.ID, me.ID, yes); err != nil {
			t.Fatalf("ResolveConfirm: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		mask := effectstest.GaLayered(t, g, changeling)
		if yes && mask.Name != "Their Ogre" {
			t.Errorf("the Shapeshifter is %q, want a copy of Their Ogre", mask.Name)
		}
		if !yes && mask.Name != "Mask" {
			t.Errorf("declined, but the Shapeshifter is %q", mask.Name)
		}
		if effectstest.GaLayered(t, g, bear).Name != "Bear" {
			t.Error("a non-Shapeshifter became a copy")
		}
	}
}

// --- In Search of Greatness ----------------------------------------------------

func oldB072Upkeep(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB072InSearchOfGreatnessOffersExactlyOnePlusTheGreatestManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B27Push(g, me.ID, "In Search of Greatness", "Enchantment", oldB072Greatness, "{G}{G}", 0, 0)
	effectstest.B27Push(g, me.ID, "Two", "Creature — Bear", "", "{1}{G}", 2, 2)
	effectstest.B27Push(g, g.Seats[1].ID, "Their Six", "Creature — Ogre", "", "{4}{G}{G}", 6, 6)
	exact := oldB072Hand(me, game.Card{Name: "Three", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 3, Toughness: 3})
	oldB072Hand(me, game.Card{Name: "Two Again", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	oldB072Hand(me, game.Card{Name: "Four", TypeLine: "Creature — Bear", ManaCost: "{3}{G}", Power: 4, Toughness: 4})
	oldB072Hand(me, game.Card{Name: "Sorcery Three", TypeLine: "Sorcery", ManaCost: "{2}{G}"})
	oldB072Upkeep(t, g)
	c := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, exact)
	if len(c.DiscardOptions) != 1 || c.DiscardOptions[0] != exact {
		t.Fatalf("offered %v, want only the permanent card of mana value 3", c.DiscardOptions)
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, exact, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if effectstest.ScryChoiceFor(g, me.ID) != nil {
		t.Error("scried after casting")
	}
}

func TestOldB072InSearchOfGreatnessScriesWhenDeclinedOrNothingFits(t *testing.T) {
	for _, offer := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.B27Push(g, me.ID, "In Search of Greatness", "Enchantment", oldB072Greatness, "{G}{G}", 0, 0)
		effectstest.B27Push(g, me.ID, "Two", "Creature — Bear", "", "{1}{G}", 2, 2)
		if offer {
			oldB072Hand(me, game.Card{Name: "Three", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 3, Toughness: 3})
		} else {
			oldB072Hand(me, game.Card{Name: "Five", TypeLine: "Creature — Bear", ManaCost: "{4}{G}", Power: 5, Toughness: 5})
		}
		effectstest.SeedLibraryTop(me, "Top", "Creature — Bear", "{G}")
		oldB072Upkeep(t, g)
		if offer {
			effectstest.DeclineResolutionCast(t, g, me.ID)
		}
		if effectstest.ScryChoiceFor(g, me.ID) == nil {
			t.Errorf("offer=%v: no scry 1", offer)
		}
	}
}

// --- Reckless Crew -----------------------------------------------------------

func TestOldB072RecklessCrewMakesATokenPerVehicleAndEquipmentAndAttaches(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Cart", "Artifact — Vehicle", "", false)
	sword := effectstest.PushCatalogPermanent(g, me.ID, "Sword", "Artifact — Equipment", "", false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Their Sword", "Artifact — Equipment", "", false)
	effectstest.CastCatalogSpell(t, g, "Reckless Crew", "Sorcery", oldB072Crew, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Dwarf Berserker"); n != 2 {
		t.Fatalf("%d Dwarf Berserker tokens, want 2 (one Vehicle, one Equipment)", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Dwarf Berserker" && (c.Power != 2 || c.Toughness != 1) {
			t.Errorf("token is %d/%d, want 2/1", c.Power, c.Toughness)
		}
	}
	// The first token: attach my Sword (their Equipment is not offered).
	p := effectstest.LatestChooseCardsFor(g, me.ID)
	if p == nil || len(p.ChooseCards) != 1 || p.ChooseCards[0] != sword {
		t.Fatalf("offered %+v, want only my Equipment", p)
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{sword}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	first := effectstest.FindBattlefieldCardByID(g, sword).AttachedTo.ID
	if first == uuid.Nil {
		t.Fatal("the Sword was not attached to the first token")
	}
	// The second token: move it again, or none.
	p = effectstest.LatestChooseCardsFor(g, me.ID)
	if p == nil {
		t.Fatal("no offer for the second token")
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	if got := effectstest.FindBattlefieldCardByID(g, sword).AttachedTo.ID; got != first {
		t.Error("declining the second token moved the Sword")
	}
}

func TestOldB072RecklessCrewWithNothingToCountMakesNoTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Reckless Crew", "Sorcery", oldB072Crew, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Dwarf Berserker"); n != 0 {
		t.Errorf("%d tokens with no Vehicle or Equipment, want 0", n)
	}
}
