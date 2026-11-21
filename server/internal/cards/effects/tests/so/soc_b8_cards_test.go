package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardskr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/kr"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestSocB8EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		effectstest.SocB8LatticeOracle:           "Lattice Library",
		effectstest.SocB8ComposerOracle:          "Leitmotif Composer",
		cardskr.LoreholdArchivistOracleID:        "Lorehold Archivist",
		cardskr.LoreholdArchivistOracleID + "#1": "Restore Relic",
		effectstest.SocB8CampusOracle:            "Lorehold Campus",
		effectstest.SocB8OpusOracle:              "Magma Opus",
		effectstest.SocB8BloomOracle:             "Mana Bloom",
		effectstest.SocB8HellkiteOracle:          "Manaform Hellkite",
		effectstest.SocB8ImpetusOracle:           "Martial Impetus",
		effectstest.SocB8MerchantOracle:          "Merchant of Venom",
		effectstest.SocB8MirrorwingOracle:        "Mirrorwing Dragon",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// socB8Cast puts a card into the active seat's hand, walks to a main
// phase and casts it.
func socB8Cast(t *testing.T, g *game.Game, c game.Card, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	if c.InstanceID == uuid.Nil {
		c.InstanceID = uuid.New()
	}
	c.Owner, c.Controller = me.ID, me.ID
	me.Hand.PushTop(c)
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, c.InstanceID, params); err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	return c.InstanceID
}

// --- Lattice Library ------------------------------------------------------

func TestSocB8LatticeLibraryEntersWithXAndMakesAFractal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lib := socB8Cast(t, g, game.Card{Name: "Lattice Library", TypeLine: "Enchantment", ManaCost: "{X}{G}{G}", OracleID: effectstest.SocB8LatticeOracle},
		game.CastSpellParams{XValue: 3})
	effectstest.PassPriorityAroundTable(t, g) // the Library resolves
	if got := effectstest.CounterCount(g, lib, "study"); got != 3 {
		t.Fatalf("study counters %d, want 3", got)
	}
	effectstest.PassPriorityAroundTable(t, g) // its enters trigger
	fr := socB4Fractals(g, me.ID)
	if len(fr) != 1 || fr[0].Counters[game.CounterPlusOne] != 3 {
		t.Fatalf("Fractals %v, want one with three +1/+1 counters", fr)
	}
	// The Library was this turn's first spell with {X}: a second one this
	// turn does not trigger it (the ruling).
	effectstest.SocB4Spell(t, g, "X Spell", "Sorcery", "{X}{U}", game.CastSpellParams{XValue: 1})
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(socB4Fractals(g, me.ID)); n != 1 {
		t.Errorf("Fractals %d after the turn's second X spell, want 1", n)
	}
}

func TestSocB8LatticeLibraryTriggersOnTheFirstXSpellEachTurnOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lib := effectstest.PushCatalogPermanent(g, me.ID, "Lattice Library", "Enchantment", effectstest.SocB8LatticeOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(lib, "study", 2) })
	effectstest.SocB4Spell(t, g, "Plain", "Sorcery", "{1}{U}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(socB4Fractals(g, me.ID)); n != 0 {
		t.Fatalf("a spell without {X} made %d Fractals", n)
	}
	effectstest.SocB4Spell(t, g, "First X", "Sorcery", "{X}{U}", game.CastSpellParams{XValue: 0})
	if n := effectstest.TriggersOnStackFrom(g, lib); n != 1 {
		t.Fatalf("Library triggers %d on the first X spell, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	fr := socB4Fractals(g, me.ID)
	if len(fr) != 1 || fr[0].Counters[game.CounterPlusOne] != 2 {
		t.Fatalf("Fractals %v, want one with two +1/+1 counters", fr)
	}
	effectstest.SocB4Spell(t, g, "Second X", "Sorcery", "{X}{U}", game.CastSpellParams{XValue: 2})
	if n := effectstest.TriggersOnStackFrom(g, lib); n != 0 {
		t.Errorf("Library triggers %d on the second X spell, want 0", n)
	}
}

// --- Leitmotif Composer ---------------------------------------------------

func TestSocB8LeitmotifComposerCopiesItselfOnABigSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Leitmotif Composer", "Creature — Human Bard", effectstest.SocB8ComposerOracle, 2, 2)
	effectstest.SocB4Spell(t, g, "Four", "Sorcery", "{3}{U}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.SocB4Spell(t, g, "Big Bear", "Creature — Bear", "{4}{G}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.SocB8Named(g, me.ID, "Leitmotif Composer")); n != 1 {
		t.Fatalf("Composers %d after a four and a creature spell, want 1", n)
	}
	effectstest.SocB4Spell(t, g, "Big X", "Instant", "{X}{U}", game.CastSpellParams{XValue: 4})
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.SocB8Named(g, me.ID, "Leitmotif Composer")
	if len(got) != 2 {
		t.Fatalf("Composers %d after a mana value 5 instant, want 2", len(got))
	}
	// Both copy on the next one: the token has the ability too.
	effectstest.SocB4Spell(t, g, "Five", "Sorcery", "{4}{U}", game.CastSpellParams{})
	// Two Composers trigger at once: their controller orders them.
	if !effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID) {
		t.Fatal("no trigger order prompt for the two Composers")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.SocB8Named(g, me.ID, "Leitmotif Composer")); n != 4 {
		t.Errorf("Composers %d, want 4", n)
	}
}

func TestSocB8LeitmotifComposerCopiesTheComposerThatDiedInResponse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	c := effectstest.B12Push(g, me.ID, "Leitmotif Composer", "Creature — Human Bard", effectstest.SocB8ComposerOracle, 2, 2)
	effectstest.SocB4Spell(t, g, "Five", "Sorcery", "{4}{U}", game.CastSpellParams{})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, c); err != nil {
		t.Fatalf("kill the Composer: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.SocB8Named(g, me.ID, "Leitmotif Composer")
	if len(got) != 1 || !effects.IsToken(got[0]) {
		t.Fatalf("Composers %v, want one token copy of the one that died (CR 608.2h)", got)
	}
}

func TestSocB8LeitmotifComposerDrawsAndMakesComposersUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Push(g, me.ID, "Leitmotif Composer", "Creature — Human Bard", effectstest.SocB8ComposerOracle, 2, 2)
	b := effectstest.B12Push(g, opp.ID, "Leitmotif Composer", "Creature — Human Bard", effectstest.SocB8ComposerOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "UUU")
	if err := g.ActivateCatalogAbility(me.ID, a, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, a, game.CantBeBlocked)
	effectstest.AssertRestrictions(t, g, b, game.CantBeBlocked)
	effectstest.AssertRestrictions(t, g, bear, 0)

	effectstest.SeedLibrary(me, "A", "B")
	hand := me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, a, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d after combat damage to a player", me.Hand.Size(), hand+1)
	}
}

// --- Lorehold Archivist // Restore Relic --------------------------------

func socB8Archivist(g *game.Game, me *game.Player) game.Card {
	c := sos1Prep(me.ID, cardskr.LoreholdArchivistOracleID, []string{"first strike"}, "Lorehold Archivist", "Creature — Dwarf Artificer", "{1}{R}{W}", 3, 2,
		"Restore Relic", "Sorcery", "{2}{R}{W}")
	effectstest.PushBattlefieldCardWithTimestamp(g, c)
	return c
}

func TestSocB8LoreholdArchivistPreparesAndRestoresARelic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	c := socB8Archivist(g, me)
	effectstest.B17GraveyardCard(me, "Old Rock", "Artifact", "{2}")
	effectstest.PushGraveyardCreature(g, me.ID, "Dead Bear", "{1}{G}")
	effectstest.B17GraveyardCard(me, "Old Spell", "Sorcery", "{R}")
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	if n := effectstest.TriggersOnStackFrom(g, c.InstanceID); n != 0 {
		t.Fatalf("two artifact and/or creature cards triggered it %d times", n)
	}
	golem := effectstest.B17GraveyardCard(me, "Old Golem", "Artifact Creature — Golem", "{3}")
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	if n := effectstest.TriggersOnStackFrom(g, c.InstanceID); n != 1 {
		t.Fatalf("three artifact and/or creature cards: triggers %d, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("not prepared")
	}
	effectstest.AdvanceToMain(t, g)
	cp, ok := effectstest.PrepareCopyOf(g, "Restore Relic")
	if !ok {
		t.Fatal("no copy of Restore Relic in exile")
	}
	if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{
		FromZone: string(game.ZoneExile), Targets: effectstest.CardTarget(golem),
	}); err != nil {
		t.Fatalf("cast Restore Relic: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(golem) {
		t.Error("the targeted card was not exiled")
	}
	got := effectstest.SocB8Named(g, me.ID, "Old Golem")
	if len(got) != 1 || !effects.IsToken(got[0]) {
		t.Errorf("Old Golem tokens %v, want one token copy", got)
	}
}

// The "if" is checked again as the ability resolves (CR 603.4).
func TestSocB8LoreholdArchivistRechecksTheGraveyardOnResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	c := socB8Archivist(g, me)
	effectstest.PushGraveyardCreature(g, me.ID, "Dead Bear", "{1}{G}")
	effectstest.PushGraveyardCreature(g, me.ID, "Dead Elk", "{1}{G}")
	last := effectstest.B17GraveyardCard(me, "Old Rock", "Artifact", "{2}")
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	if n := effectstest.TriggersOnStackFrom(g, c.InstanceID); n != 1 {
		t.Fatalf("triggers %d, want 1", n)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneExile}, last); err != nil {
		t.Fatalf("exile from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Error("prepared though only two cards were left as the ability resolved")
	}
}

func TestSocB8RestoreRelicRefusesANoncreatureNonartifactCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	c := socB8Archivist(g, me)
	effectstest.AdvanceToMain(t, g)
	if _, err := g.BecomePreparedForEffect(c.InstanceID); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	spell := effectstest.B17GraveyardCard(me, "Old Spell", "Sorcery", "{R}")
	cp, ok := effectstest.PrepareCopyOf(g, "Restore Relic")
	if !ok {
		t.Fatal("no copy of Restore Relic in exile")
	}
	if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{
		FromZone: string(game.ZoneExile), Targets: effectstest.CardTarget(spell), Strict: true,
	}); err == nil {
		t.Error("Restore Relic targeted a sorcery card")
	}
}

// --- Lorehold Campus ------------------------------------------------------

func TestSocB8LoreholdCampusEntersTappedMakesROrWAndScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Top", "Next")
	id := effectstest.PlayLandFromHand(t, g, "Lorehold Campus", effectstest.SocB8CampusOracle)
	effectstest.PassPriorityAroundTable(t, g)
	card, ok := effectstest.BattlefieldCard(g, id)
	if !ok || !card.Tapped {
		t.Fatalf("on battlefield %v, tapped %v; want it to enter tapped", ok, card.Tapped)
	}
	effectstest.UntapForTest(g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err == nil {
		t.Fatal("the land made {U}")
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"W"}}); err != nil {
		t.Fatalf("{T}: Add {W}: %v", err)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })

	effectstest.UntapForTest(g, id)
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("scry activated with three mana")
	}
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{4},{T}: scry 1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if len(c.ScryCards) != 1 {
		t.Fatalf("scry looked at %d cards, want 1", len(c.ScryCards))
	}
	top := c.ScryCards[0]
	if err := g.ResolveScry(c.ID, me.ID, []uuid.UUID{top}, nil); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	if bottom, err := me.Library.Bottom(); err != nil || bottom.InstanceID != top {
		t.Error("the scried card is not on the bottom")
	}
}

// --- Magma Opus -----------------------------------------------------------

func TestSocB8MagmaOpusDividesTapsMakesAnElementalAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false)
	effectstest.SeedLibrary(me, "A", "B", "C")
	hand, life := me.Hand.Size(), opp.Life
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Magma Opus", TypeLine: "Instant", ManaCost: "{6}{U}{R}", OracleID: effectstest.SocB8OpusOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: ogre, Slot: 0},
			{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
			{Kind: game.TargetCard, ID: ogre, Slot: 1},
			{Kind: game.TargetCard, ID: rock, Slot: 1},
		},
		Distribution: map[uuid.UUID]int{ogre: 3, opp.ID: 1},
	}); err != nil {
		t.Fatalf("cast Magma Opus: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-1)
	}
	if g.Battlefield.Contains(ogre) {
		t.Error("the Ogre survived 3 damage")
	}
	if r, ok := effectstest.BattlefieldCard(g, rock); !ok || !r.Tapped {
		t.Error("the Rock was not tapped")
	}
	if el := effectstest.SocB8Named(g, me.ID, "Elemental"); len(el) != 1 || effectstest.EffectivePower(t, g, el[0].InstanceID) != 4 {
		t.Errorf("Elementals %v, want one 4/4", el)
	}
	// The Opus left the hand as it was cast; two cards came back.
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+2)
	}
}

func TestSocB8MagmaOpusNeedsTwoPermanentsToTap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.B12Creature(g, opp.ID, "Elk", "Creature — Elk", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Magma Opus", TypeLine: "Instant", ManaCost: "{6}{U}{R}", OracleID: effectstest.SocB8OpusOracle, Owner: me.ID, Controller: me.ID})
	err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: ogre, Slot: 0}, {Kind: game.TargetCard, ID: ogre, Slot: 1}},
		Distribution: map[uuid.UUID]int{ogre: 4},
		Strict:       true,
	})
	if err == nil {
		t.Error("Magma Opus was cast with one permanent to tap")
	}
}

func TestSocB8MagmaOpusDiscardsForATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Magma Opus", TypeLine: "Instant", ManaCost: "{6}{U}{R}", OracleID: effectstest.SocB8OpusOracle, Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "UR")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate from hand: %v", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the Opus was not discarded as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("Treasures %d, want 1", n)
	}
}

// --- Mana Bloom -----------------------------------------------------------

func TestSocB8ManaBloomSpendsACounterOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bloom := socB8Cast(t, g, game.Card{Name: "Mana Bloom", TypeLine: "Enchantment", ManaCost: "{X}{G}", OracleID: effectstest.SocB8BloomOracle},
		game.CastSpellParams{XValue: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bloom, game.CounterCharge); got != 2 {
		t.Fatalf("charge counters %d, want 2", got)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := g.ActivateManaAbility(me.ID, bloom, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err != nil {
		t.Fatalf("remove a counter for {B}: %v", err)
	}
	if got := effectstest.CounterCount(g, bloom, game.CounterCharge); got != 1 {
		t.Errorf("charge counters %d, want 1", got)
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("pool %d, want 1", me.ManaPool.Total())
	}
	if err := g.ActivateManaAbility(me.ID, bloom, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err == nil {
		t.Error("a second activation this turn was allowed")
	}
}

func TestSocB8ManaBloomReturnsToHandWithNoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bloom := effectstest.PushCatalogPermanent(g, me.ID, "Mana Bloom", "Enchantment", effectstest.SocB8BloomOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bloom, game.CounterCharge, 1) })
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	if n := effectstest.TriggersOnStackFrom(g, bloom); n != 0 {
		t.Fatalf("a Bloom with a counter triggered %d times", n)
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, bloom, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("remove the last counter: %v", err)
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	if n := effectstest.TriggersOnStackFrom(g, bloom); n != 1 {
		t.Fatalf("an empty Bloom triggered %d times, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bloom) {
		t.Error("Mana Bloom did not return to its owner's hand")
	}
}

// --- Manaform Hellkite ----------------------------------------------------

func TestSocB8ManaformHellkiteMakesADragonOfTheManaSpent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Manaform Hellkite", "Creature — Dragon", effectstest.SocB8HellkiteOracle, 4, 4)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FillPoolColored(me, "R", 3)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Three", TypeLine: "Sorcery", ManaCost: "{2}{R}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	dragons := effectstest.SocB8Named(g, me.ID, "Dragon Illusion")
	if len(dragons) != 1 {
		t.Fatalf("Dragon Illusions %d, want 1", len(dragons))
	}
	d := dragons[0].InstanceID
	if effectstest.EffectivePower(t, g, d) != 3 || effectstest.EffectiveToughness(t, g, d) != 3 {
		t.Errorf("Dragon Illusion is %d/%d, want 3/3", effectstest.EffectivePower(t, g, d), effectstest.EffectiveToughness(t, g, d))
	}
	if !effectstest.HasEffectiveKeyword(t, g, d, "flying") || !effectstest.HasEffectiveKeyword(t, g, d, "haste") {
		t.Error("the Dragon Illusion lacks flying or haste")
	}
	// A creature spell does not trigger it.
	effectstest.FillPoolColored(me, "G", 2)
	effectstest.SocB4Spell(t, g, "Bear", "Creature — Bear", "{1}{G}", game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.SocB8Named(g, me.ID, "Dragon Illusion")); n != 1 {
		t.Errorf("a creature spell made a Dragon Illusion: %d", n)
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(d) {
		t.Error("the Dragon Illusion was not exiled at the end step")
	}
}

// Cast without paying, nothing was spent: a 0/0 that dies (the ruling).
func TestSocB8ManaformHellkiteFreeSpellMakesADyingDragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	h := effectstest.B12Push(g, me.ID, "Manaform Hellkite", "Creature — Dragon", effectstest.SocB8HellkiteOracle, 4, 4)
	effectstest.SocB4Spell(t, g, "Free", "Instant", "", game.CastSpellParams{})
	if n := effectstest.TriggersOnStackFrom(g, h); n != 1 {
		t.Fatalf("triggers %d, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.SocB8Named(g, me.ID, "Dragon Illusion")); n != 0 {
		t.Errorf("a 0/0 Dragon Illusion survived: %d (CR 704.5f)", n)
	}
}

// --- Martial Impetus ------------------------------------------------------

func TestSocB8MartialImpetusGoadsAndPumpsTheOtherAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	elk := effectstest.B12Creature(g, me.ID, "Elk", "Creature — Elk", 2, 2)
	ox := effectstest.B12Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2)
	home := effectstest.B12Creature(g, me.ID, "Home", "Creature — Bear", 2, 2)
	effectstest.AuraCast(t, g, "Martial Impetus", effectstest.SocB8ImpetusOracle, bear)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tg != 3 {
		t.Errorf("enchanted bear is %d/%d, want 3/3", p, tg)
	}
	if c, ok := g.LookupCardForEffect(bear); !ok || !c.Goaded() {
		t.Error("the enchanted creature is not goaded")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, at := range []struct{ id, def uuid.UUID }{{bear, opp.ID}, {elk, opp.ID}, {ox, third.ID}} {
		if err := g.DeclareAttacker(at.id, at.def); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("the enchanted creature is %d power, want 3 (it is not 'other')", p)
	}
	for _, id := range []uuid.UUID{elk, ox} {
		if p, tg := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 3 || tg != 3 {
			t.Errorf("an attacker is %d/%d, want 3/3", p, tg)
		}
	}
	if p := effectstest.EffectivePower(t, g, home); p != 2 {
		t.Errorf("a creature that did not attack is %d power, want 2", p)
	}
}

// --- Merchant of Venom ----------------------------------------------------

func TestSocB8MerchantOfVenomEdictsAndGrowsOnEverySacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.PushVanillaCreature(g, third.ID, "Third Bear", 2, 2)
	merchant := socB8Cast(t, g, game.Card{Name: "Merchant of Venom", TypeLine: "Creature — Cat Warlock", ManaCost: "{3}{B}",
		OracleID: effectstest.SocB8MerchantOracle, Power: 1, Toughness: 1}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g) // the creature spell
	effectstest.PassPriorityAroundTable(t, g) // its enters trigger
	effectstest.AnswerSacrifice(t, g, me.ID, mine)
	effectstest.AnswerSacrifice(t, g, opp.ID, theirs)
	effectstest.AnswerSacrifice(t, g, third.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Fatal("the edict did not take the chosen creatures")
	}
	if !g.Battlefield.Contains(merchant) {
		t.Fatal("the Merchant was sacrificed though another creature was chosen")
	}
	if got := effectstest.CounterCount(g, merchant, game.CounterPlusOne); got != 3 {
		t.Errorf("+1/+1 counters %d, want 3 (one per creature sacrificed)", got)
	}
	// A noncreature permanent sacrificed later counts too.
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(rock) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, merchant, game.CounterPlusOne); got != 4 {
		t.Errorf("+1/+1 counters %d after an artifact was sacrificed, want 4", got)
	}
}

// --- Mirrorwing Dragon ----------------------------------------------------

func TestSocB8MirrorwingDragonCopiesForEachOfTheCastersCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	dragon := effectstest.B12Push(g, opp.ID, "Mirrorwing Dragon", "Creature — Dragon", effectstest.SocB8MirrorwingOracle, 4, 5)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Giant Growth", "Instant", effectstest.SocB8GiantGrowthOracle, effectstest.CardTarget(dragon))
	if n := effectstest.TriggersOnStackFrom(g, dragon); n != 1 {
		t.Fatalf("triggers %d, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no order prompt for the caster")
	}
	if len(pick.ChooseCards) != 2 || !effectstest.HasID(pick.ChooseCards, a) || !effectstest.HasID(pick.ChooseCards, b) {
		t.Fatalf("order prompt over %v, want the caster's two other creatures", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, b, a)
	for i := 0; i < 4 && len(g.Stack.Cards) > 0; i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	for _, id := range []uuid.UUID{a, b} {
		if p := effectstest.EffectivePower(t, g, id); p != 5 {
			t.Errorf("the caster's creature is %d power, want 5", p)
		}
	}
	if p := effectstest.EffectivePower(t, g, dragon); p != 7 {
		t.Errorf("the Dragon is %d power, want 7", p)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
		t.Errorf("the Dragon's controller's creature is %d power, want 2", p)
	}
}

func TestSocB8MirrorwingDragonIgnoresASpellWithAnotherTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	dragon := effectstest.B12Push(g, opp.ID, "Mirrorwing Dragon", "Creature — Dragon", effectstest.SocB8MirrorwingOracle, 4, 5)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Giant Growth", "Instant", effectstest.SocB8GiantGrowthOracle, effectstest.CardTarget(a))
	if n := effectstest.TriggersOnStackFrom(g, dragon); n != 0 {
		t.Errorf("a spell targeting another creature triggered it %d times", n)
	}
}

func TestSocB8MirrorwingDragonSkipsACreatureTheSpellCantTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	dragon := effectstest.B12Push(g, me.ID, "Mirrorwing Dragon", "Creature — Dragon", effectstest.SocB8MirrorwingOracle, 4, 5)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	shy := effectstest.B12Creature(g, me.ID, "Shy Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == shy {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "shroud")
			}
		}
	})
	effectstest.CastCatalogSpell(t, g, "Giant Growth", "Instant", effectstest.SocB8GiantGrowthOracle, effectstest.CardTarget(dragon))
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.ChooseCardsChoiceFor(g, me.ID); pick != nil {
		t.Fatal("an order prompt with one creature the spell could target")
	}
	for i := 0; i < 4 && len(g.Stack.Cards) > 0; i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p := effectstest.EffectivePower(t, g, a); p != 5 {
		t.Errorf("Bear A is %d power, want 5", p)
	}
	if p := effectstest.EffectivePower(t, g, shy); p != 2 {
		t.Errorf("the shroud Bear is %d power, want 2", p)
	}
}
