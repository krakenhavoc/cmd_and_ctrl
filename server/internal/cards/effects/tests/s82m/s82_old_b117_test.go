package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ikoria: Lair of Behemoths, batch 117 (#5512).

const (
	oldB117TitansNest     = "04505544-8341-4bd9-afa7-b31d6b415c47"
	oldB117MythosBrokkos  = "14596108-a8ed-4e68-b787-d655b28bdab4"
	oldB117DeathsOasis    = "b10f6915-96b6-4e87-8171-4d3e96bfd48c"
	oldB117Offsprings     = "e9fa3e3c-4a7e-4a18-b2ef-15d53ff55321"
	oldB117SurvivorsBond  = "ff53c3a0-988a-40a4-b5f9-e5a4b634f5eb"
	oldB117DireTactics    = "0c8cb2ec-5b58-497c-a6e7-2d25512d5a44"
	oldB117MythosNethroi  = "16da72a3-d980-4dd8-99f2-8191cce00978"
	oldB117GeneralKudro   = "a5a024a5-16a8-4083-9157-5905c5e96d76"
	oldB117MutualDestruct = "b01cd710-3906-486b-8bb5-8c4efa46e4d9"
	oldB117EscapeProtocol = "e875b468-b475-4573-adf6-312ca4ab6994"
)

func TestOldB117CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		oldB117TitansNest: "Titans' Nest", oldB117MythosBrokkos: "Mythos of Brokkos", oldB117DeathsOasis: "Death's Oasis",
		oldB117Offsprings: "Offspring's Revenge", oldB117SurvivorsBond: "Survivors' Bond", oldB117DireTactics: "Dire Tactics",
		oldB117MythosNethroi: "Mythos of Nethroi", oldB117GeneralKudro: "General Kudro of Drannith",
		oldB117MutualDestruct: "Mutual Destruction", oldB117EscapeProtocol: "Escape Protocol",
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

// oldB117Cast puts a spell in hand, floats the pool and casts it.
func oldB117Cast(t *testing.T, g *game.Game, me *game.Player, c game.Card, pool string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = me.ID, me.ID
	me.Hand.PushTop(c)
	effectstest.AddPool(t, g, me.ID, pool)
	return c.InstanceID, g.CastSpell(me.ID, c.InstanceID, params)
}

// oldB117Graveyard seeds a card with a mana cost and type line into a graveyard.
func oldB117Graveyard(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID})
	return id
}

// --- Survivors' Bond ---------------------------------------------------------

func oldB117Bond() game.Card {
	return game.Card{Name: "Survivors' Bond", TypeLine: "Sorcery", OracleID: oldB117SurvivorsBond, ManaCost: "{1}{G}", Colors: []string{"G"}}
}

func TestOldB117SurvivorsBondBothModesReturnTheHumanAndTheNonHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	human := oldB117Graveyard(me, "Dead Human", "Creature — Human Soldier", "{1}{W}")
	beast := oldB117Graveyard(me, "Dead Beast", "Creature — Beast", "{2}{G}")
	if _, err := oldB117Cast(t, g, me, oldB117Bond(), "{G}{C}", game.CastSpellParams{
		Strict: true, TargetsStamped: true, Modes: []int{0, 1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: human, Mode: 0}, {Kind: game.TargetCard, ID: beast, Mode: 1}},
	}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(human) || !me.Hand.Contains(beast) {
		t.Errorf("hand has human=%v beast=%v, want both returned", me.Hand.Contains(human), me.Hand.Contains(beast))
	}
}

func TestOldB117SurvivorsBondEachModeReadsItsOwnTargetClause(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	human := oldB117Graveyard(me, "Dead Human", "Creature — Human Soldier", "{1}{W}")
	beast := oldB117Graveyard(me, "Dead Beast", "Creature — Beast", "{2}{G}")
	// A non-Human can't be the first bullet's target, nor a Human the second's.
	if _, err := oldB117Cast(t, g, me, oldB117Bond(), "{G}{C}", game.CastSpellParams{
		Strict: true, TargetsStamped: true, Modes: []int{0},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: beast, Mode: 0}},
	}); err == nil {
		t.Fatal("a non-Human was accepted for the Human bullet")
	}
	if _, err := oldB117Cast(t, g, me, oldB117Bond(), "{G}{C}", game.CastSpellParams{
		Strict: true, TargetsStamped: true, Modes: []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: human, Mode: 0}},
	}); err == nil {
		t.Fatal("a Human was accepted for the non-Human bullet")
	}
	// The second bullet alone returns only its card.
	if _, err := oldB117Cast(t, g, me, oldB117Bond(), "{G}{C}", game.CastSpellParams{
		Strict: true, TargetsStamped: true, Modes: []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: beast, Mode: 0}},
	}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(beast) || me.Hand.Contains(human) {
		t.Errorf("beast in hand=%v, human in hand=%v", me.Hand.Contains(beast), me.Hand.Contains(human))
	}
}

// --- Dire Tactics ------------------------------------------------------------

func oldB117Tactics() game.Card {
	return game.Card{Name: "Dire Tactics", TypeLine: "Instant", OracleID: oldB117DireTactics, ManaCost: "{W}{B}", Colors: []string{"W", "B"}}
}

func TestOldB117DireTacticsCostsLifeEqualToToughnessWithoutAHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foe := effectstest.B12Creature(g, opp.ID, "Big Foe", "Creature — Beast", 3, 5)
	life := me.Life
	if _, err := oldB117Cast(t, g, me, oldB117Tactics(), "{W}{B}", game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(foe)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(foe) || !effectstest.InExile(g, foe) {
		t.Error("the creature was not exiled")
	}
	if me.Life != life-5 {
		t.Errorf("life %d, want %d (lost the creature's toughness)", me.Life, life-5)
	}
}

func TestOldB117DireTacticsIsFreeWhileYouControlAHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "My Human", "Creature — Human Soldier", 1, 1)
	foe := effectstest.B12Creature(g, opp.ID, "Big Foe", "Creature — Beast", 3, 5)
	life := me.Life
	if _, err := oldB117Cast(t, g, me, oldB117Tactics(), "{W}{B}", game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(foe)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, foe) {
		t.Error("the creature was not exiled")
	}
	if me.Life != life {
		t.Errorf("life %d, want %d", me.Life, life)
	}
}

func TestOldB117DireTacticsExilingYourOwnOnlyHumanCostsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "My Human", "Creature — Human Soldier", 1, 3)
	life := me.Life
	if _, err := oldB117Cast(t, g, me, oldB117Tactics(), "{W}{B}", game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(mine)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Errorf("life %d, want %d: the Human is gone when the clause is read", me.Life, life-3)
	}
}

// --- Mythos of Nethroi -------------------------------------------------------

func oldB117Nethroi() game.Card {
	return game.Card{Name: "Mythos of Nethroi", TypeLine: "Instant", OracleID: oldB117MythosNethroi, ManaCost: "{2}{B}", Colors: []string{"B"}}
}

func TestOldB117MythosOfNethroiDestroysACreatureForBlackAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Beast", 2, 2)
	if _, err := oldB117Cast(t, g, me, oldB117Nethroi(), "{B}{C}{C}", game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(foe)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(foe) {
		t.Error("the creature survived")
	}
}

func TestOldB117MythosOfNethroiNeedsGreenAndWhiteForANoncreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.SeedPermanentWithOracle(g, opp.ID, "Rock", "Artifact", uuid.NewString())
	if _, err := oldB117Cast(t, g, me, oldB117Nethroi(), "{B}{C}{C}", game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(rock)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Fatal("a noncreature permanent was destroyed without {G}{W}")
	}
	// Green alone is not enough either.
	if _, err := oldB117Cast(t, g, me, oldB117Nethroi(), "{B}{G}{C}", game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(rock)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Fatal("a noncreature permanent was destroyed with {G} but no {W}")
	}
	if _, err := oldB117Cast(t, g, me, oldB117Nethroi(), "{B}{G}{W}", game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(rock)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("{G}{W} was spent and the permanent survived")
	}
}

func TestOldB117MythosOfNethroiCannotTargetALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.SeedPermanentWithOracle(g, opp.ID, "Plains", "Basic Land — Plains", uuid.NewString())
	if _, err := oldB117Cast(t, g, me, oldB117Nethroi(), "{B}{G}{W}", game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(land)}); err == nil {
		t.Error("a land was a legal target")
	}
}

// --- Mutual Destruction ------------------------------------------------------

func oldB117Mutual() game.Card {
	return game.Card{Name: "Mutual Destruction", TypeLine: "Sorcery", OracleID: oldB117MutualDestruct, ManaCost: "{B}", Colors: []string{"B"}}
}

func TestOldB117MutualDestructionSacrificesACreatureToDestroyAnother(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Beast", 4, 4)
	if _, err := oldB117Cast(t, g, me, oldB117Mutual(), "{B}", game.CastSpellParams{
		Strict: true, Targets: effectstest.CardTarget(foe), SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) || g.Battlefield.Contains(foe) {
		t.Error("the sacrificed creature or the target is still on the battlefield")
	}
}

func TestOldB117MutualDestructionIsRefusedWithNothingToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Beast", 4, 4)
	if _, err := oldB117Cast(t, g, me, oldB117Mutual(), "{B}", game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(foe)}); err == nil {
		t.Error("cast without a creature to sacrifice")
	}
}

func TestOldB117MutualDestructionHasFlashOnlyWhileYouControlAPermanentWithFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Beast", 4, 4)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	c := oldB117Mutual()
	c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
	me.Hand.PushTop(c)
	effectstest.AddPool(t, g, me.ID, "{B}")
	params := game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(foe), SacrificeIDs: []uuid.UUID{fodder}}
	if err := g.CastSpell(me.ID, c.InstanceID, params); err == nil {
		t.Fatal("cast at instant speed with no permanent with flash")
	}
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flash Bear", TypeLine: "Creature — Bear", Keywords: []string{"flash"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, c.InstanceID, params); err != nil {
		t.Fatalf("with a permanent with flash: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(foe) {
		t.Error("the target survived")
	}
}

// --- Mythos of Brokkos -------------------------------------------------------

func oldB117Brokkos() game.Card {
	return game.Card{Name: "Mythos of Brokkos", TypeLine: "Sorcery", OracleID: oldB117MythosBrokkos, ManaCost: "{2}{G}{G}", Colors: []string{"G"}}
}

func oldB117PickFromGraveyard(t *testing.T, g *game.Game, who uuid.UUID, want ...uuid.UUID) *game.PendingChoice {
	t.Helper()
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, who)
	if p == nil {
		t.Fatal("no card pick is open")
	}
	shown := append([]uuid.UUID(nil), p.ChooseCards...)
	if err := g.ResolveChooseCards(p.ID, who, want); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	return &game.PendingChoice{ChooseCards: shown}
}

func TestOldB117MythosOfBrokkosReturnsUpToTwoPermanentsWithoutBlueBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := oldB117Graveyard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	b := oldB117Graveyard(me, "Dead Relic", "Artifact", "{2}")
	c := oldB117Graveyard(me, "Dead Land", "Land", "")
	spell := oldB117Graveyard(me, "Dead Spell", "Sorcery", "{R}")
	library := me.Library.Size()
	if _, err := oldB117Cast(t, g, me, oldB117Brokkos(), "{G}{G}{C}{C}", game.CastSpellParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	shown := oldB117PickFromGraveyard(t, g, me.ID, a, b)
	for _, id := range []uuid.UUID{a, b, c} {
		if !contains(shown.ChooseCards, id) {
			t.Errorf("permanent card %s was not offered", id)
		}
	}
	if contains(shown.ChooseCards, spell) {
		t.Error("a sorcery card was offered")
	}
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) || me.Hand.Contains(c) {
		t.Errorf("hand has a=%v b=%v c=%v, want the two picked", me.Hand.Contains(a), me.Hand.Contains(b), me.Hand.Contains(c))
	}
	if me.Library.Size() != library {
		t.Errorf("the library changed from %d to %d without {U}{B}", library, me.Library.Size())
	}
}

func contains(ids []uuid.UUID, id uuid.UUID) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

func TestOldB117MythosOfBrokkosWithBlueBlackSearchesACardIntoTheGraveyardFirst(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	target := game.Card{InstanceID: uuid.New(), Name: "Buried Titan", TypeLine: "Creature — Giant", ManaCost: "{5}{G}", Owner: me.ID, Controller: me.ID}
	me.Library.PushTop(target)
	if _, err := oldB117Cast(t, g, me, oldB117Brokkos(), "{U}{B}{G}{G}", game.CastSpellParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, target.InstanceID)
	shown := oldB117PickFromGraveyard(t, g, me.ID, target.InstanceID)
	if !contains(shown.ChooseCards, target.InstanceID) {
		t.Error("the card just searched into the graveyard was not offered")
	}
	if !me.Hand.Contains(target.InstanceID) {
		t.Error("the searched permanent card did not reach the hand")
	}
}

// --- Death's Oasis -----------------------------------------------------------

func oldB117Sac(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(id) })
}

func TestOldB117DeathsOasisMillsTwoThenReturnsALesserCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedPermanentWithOracle(g, me.ID, "Death's Oasis", "Enchantment", oldB117DeathsOasis)
	dying := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dying Giant", TypeLine: "Creature — Giant", ManaCost: "{2}{G}",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	cheap := oldB117Graveyard(me, "Cheap Bear", "Creature — Bear", "{1}")
	pricey := oldB117Graveyard(me, "Pricey Bear", "Creature — Bear", "{3}{G}")
	equal := oldB117Graveyard(me, "Equal Bear", "Creature — Bear", "{1}{G}{G}")
	spell := oldB117Graveyard(me, "Old Spell", "Sorcery", "{1}")
	for i := 0; i < 2; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Milled Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	library := me.Library.Size()
	oldB117Sac(t, g, dying)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != library-2 {
		t.Fatalf("library %d, want %d (two milled)", got, library-2)
	}
	shown := oldB117PickFromGraveyard(t, g, me.ID, cheap)
	if len(shown.ChooseCards) != 1 || shown.ChooseCards[0] != cheap {
		t.Errorf("offered %v, want only the lesser creature card %s", shown.ChooseCards, cheap)
	}
	if !me.Hand.Contains(cheap) {
		t.Error("the lesser creature card did not reach the hand")
	}
	for _, id := range []uuid.UUID{pricey, equal, spell, dying} {
		if me.Hand.Contains(id) {
			t.Errorf("card %s returned although it is not a lesser creature card", id)
		}
	}
}

func TestOldB117DeathsOasisIgnoresTokensAndOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedPermanentWithOracle(g, me.ID, "Death's Oasis", "Enchantment", oldB117DeathsOasis)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Spirit", TypeLine: "Token Creature — Spirit",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	oldB117Graveyard(me, "Cheap Bear", "Creature — Bear", "{0}")
	library := me.Library.Size()
	oldB117Sac(t, g, token)
	oldB117Sac(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != library {
		t.Errorf("library %d, want %d: neither a token nor an opponent's creature triggers it", me.Library.Size(), library)
	}
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) != nil {
		t.Error("a pick is open")
	}
}

func TestOldB117DeathsOasisSacrificeGainsLifeEqualToTheGreatestManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	oasis := effectstest.SeedPermanentWithOracle(g, me.ID, "Death's Oasis", "Enchantment", oldB117DeathsOasis)
	for _, mv := range []string{"{2}", "{4}{G}", "{1}"} {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Body " + mv, TypeLine: "Creature — Bear", ManaCost: mv,
			Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		})
	}
	life := me.Life
	effectstest.AddPool(t, g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, oasis, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(oasis) {
		t.Error("Death's Oasis was not sacrificed")
	}
	if me.Life != life+5 {
		t.Errorf("life %d, want %d (the greatest mana value among my creatures)", me.Life, life+5)
	}
}

// --- Titans' Nest ------------------------------------------------------------

func TestOldB117TitansNestManaPaysOnlyForColoredSpellsWithoutX(t *testing.T) {
	setup := func() (*game.Game, *game.Player) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		nest := effectstest.SeedPermanentWithOracle(g, me.ID, "Titans' Nest", "Enchantment", oldB117TitansNest)
		fuel := oldB117Graveyard(me, "Fuel", "Sorcery", "{1}")
		if err := g.ActivateManaAbility(me.ID, nest, 0, game.ManaAbilityParams{ExileIDs: []uuid.UUID{fuel}}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		if !effectstest.InExile(g, fuel) || me.Graveyard.Contains(fuel) {
			t.Fatal("the graveyard card was not exiled as the cost")
		}
		if me.ManaPool.Total() != 1 {
			t.Fatalf("pool has %d mana, want {C}", me.ManaPool.Total())
		}
		return g, me
	}
	t.Run("a colored spell is payable", func(t *testing.T) {
		g, me := setup()
		id := effectstest.HandCardFull(me, "Green One", "Sorcery", "{1}", "", []string{"G"})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("a colored spell paid with the restricted mana: %v", err)
		}
	})
	t.Run("a multicolored spell is payable", func(t *testing.T) {
		g, me := setup()
		id := effectstest.HandCardFull(me, "Gold One", "Sorcery", "{1}", "", []string{"G", "U"})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("a multicolored spell paid with the restricted mana: %v", err)
		}
	})
	t.Run("a colorless spell is not", func(t *testing.T) {
		g, me := setup()
		id := effectstest.HandCardFull(me, "Grey One", "Artifact", "{1}", "", nil)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
			t.Error("a colorless spell was paid with the restricted mana")
		}
	})
	t.Run("a colored spell with X in its mana cost is not", func(t *testing.T) {
		g, me := setup()
		id := effectstest.HandCardFull(me, "X Spell", "Sorcery", "{X}{G}", "", []string{"G"})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 0}); err == nil {
			t.Error("a spell with {X} was paid with the restricted mana")
		}
	})
}

func TestOldB117TitansNestSurveilsAtYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Top", "Next")
	effectstest.SeedPermanentWithOracle(g, me.ID, "Titans' Nest", "Enchantment", oldB117TitansNest)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToUpkeepOfSeat(t, g, effects.SeatIndexOf(g, me.ID))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil at the upkeep")
	}
	if len(c.ScryCards) != 1 {
		t.Errorf("surveil looked at %d cards, want 1", len(c.ScryCards))
	}
}

// --- Offspring's Revenge -----------------------------------------------------

func TestOldB117OffspringsRevengeCopiesAColoredCreatureCardAs1x1WithHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedPermanentWithOracle(g, me.ID, "Offspring's Revenge", "Enchantment", oldB117Offsprings)
	red := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: red, Name: "Fallen Ogre", TypeLine: "Creature — Ogre", ManaCost: "{3}{R}", Colors: []string{"R"}, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	green := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: green, Name: "Fallen Elf", TypeLine: "Creature — Elf", ManaCost: "{G}", Colors: []string{"G"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardTarget(red)); err != nil {
			t.Fatalf("pick target: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(red) || !effectstest.InExile(g, red) {
		t.Fatal("the red creature card was not exiled")
	}
	if !me.Graveyard.Contains(green) {
		t.Error("the green card, an illegal target, was exiled")
	}
	var token *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Fallen Ogre" && g.Battlefield.Cards[i].IsToken() {
			token = &g.Battlefield.Cards[i]
		}
	}
	if token == nil {
		t.Fatal("no token copy of the Ogre")
	}
	eff := token.Effective()
	if eff.Power != 1 || eff.Toughness != 1 {
		t.Errorf("token is %d/%d, want 1/1", eff.Power, eff.Toughness)
	}
	if !effectstest.HasKeywordOf(g, token.InstanceID, "haste") {
		t.Error("the token lacks haste")
	}
}

// --- General Kudro of Drannith -----------------------------------------------

func TestOldB117GeneralKudroPumpsOtherHumansAndExilesOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	card := oldB117Graveyard(opp, "Their Dead", "Creature — Bear", "{1}")
	mine := oldB117Graveyard(me, "My Dead", "Creature — Bear", "{1}")
	human := effectstest.B12Creature(g, me.ID, "Plain Human", "Creature — Human", 1, 1)
	kudro := effectstest.B12Push(g, me.ID, "General Kudro of Drannith", "Legendary Creature — Human Soldier", oldB117GeneralKudro, 3, 3)
	if p, tough := effectstest.PtOf(t, g, human); p != 2 || tough != 2 {
		t.Errorf("another Human is %d/%d, want 2/2", p, tough)
	}
	if p, _ := effectstest.PtOf(t, g, kudro); p != 3 {
		t.Errorf("the General pumped himself: power %d", p)
	}
	// A Human entering under my control triggers; the pick is an opponent's graveyard card.
	_, err := effectstest.Msc17Cast(t, g, me, game.Card{Name: "Fresh Human", TypeLine: "Creature — Human", Power: 1, Toughness: 1}, game.CastSpellParams{})
	if err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickTriggerTarget(t, g, me.ID, card)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(card) || !effectstest.InExile(g, card) {
		t.Error("the opponent's graveyard card was not exiled")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("my own graveyard card was exiled")
	}
}

func TestOldB117GeneralKudroSacrificesTwoHumansToDestroyAPowerFourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	kudro := effectstest.B12Push(g, me.ID, "General Kudro of Drannith", "Legendary Creature — Human Soldier", oldB117GeneralKudro, 3, 3)
	a := effectstest.B12Creature(g, me.ID, "Human A", "Creature — Human", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "Human B", "Creature — Human", 1, 1)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Beast", 4, 4)
	effectstest.AddPool(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, kudro, 0, game.ActivateAbilityParams{
		Strict: true, Targets: effectstest.CardTarget(small), SacrificeIDs: []uuid.UUID{a, b},
	}); err == nil {
		t.Fatal("a power 3 creature was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, kudro, 0, game.ActivateAbilityParams{
		Strict: true, Targets: effectstest.CardTarget(big), SacrificeIDs: []uuid.UUID{a},
	}); err == nil {
		t.Fatal("activated with only one Human sacrificed")
	}
	if err := g.ActivateCatalogAbility(me.ID, kudro, 0, game.ActivateAbilityParams{
		Strict: true, Targets: effectstest.CardTarget(big), SacrificeIDs: []uuid.UUID{a, b},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) || g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("the target or a sacrificed Human is still on the battlefield")
	}
	if !g.Battlefield.Contains(small) {
		t.Error("the small creature was destroyed")
	}
}

// --- Escape Protocol ---------------------------------------------------------

func oldB117EscapeCycle(t *testing.T, g *game.Game, pool string) *game.Player {
	t.Helper()
	_, me := effectstest.CycleFromHand(t, g, "Scrounging Skyray", "Creature — Bird", effectstest.ScroungingSkyrayOracle, pool)
	effectstest.PassPriorityAroundTable(t, g)
	return me
}

func oldB117BlinkedBear(g *game.Game, me *game.Player) (uuid.UUID, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Marked Bear" && c.Controller == me.ID {
			return c.InstanceID, true
		}
	}
	return uuid.Nil, false
}

func TestOldB117EscapeProtocolPaysOneToBlinkAPermanentWhenYouCycle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedPermanentWithOracle(g, me.ID, "Escape Protocol", "Enchantment", oldB117EscapeProtocol)
	bear := effectstest.B12Creature(g, me.ID, "Marked Bear", "Creature — Bear", 2, 2)
	me = oldB117EscapeCycle(t, g, "{C}{C}{C}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the old object is still on the battlefield")
	}
	if _, ok := oldB117BlinkedBear(g, me); !ok {
		t.Error("the creature did not come back")
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("%d mana left, want the {1} paid", got)
	}
}

func TestOldB117EscapeProtocolDecliningOrCyclingWithNothingToPayBlinksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedPermanentWithOracle(g, me.ID, "Escape Protocol", "Enchantment", oldB117EscapeProtocol)
	bear := effectstest.B12Creature(g, me.ID, "Marked Bear", "Creature — Bear", 2, 2)
	me = oldB117EscapeCycle(t, g, "{C}{C}{C}")
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature was blinked without paying")
	}
}

func TestOldB117EscapeProtocolOnlyWatchesYourOwnCycling(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedPermanentWithOracle(g, opp.ID, "Escape Protocol", "Enchantment", oldB117EscapeProtocol)
	effectstest.B12Creature(g, opp.ID, "Marked Bear", "Creature — Bear", 2, 2)
	oldB117EscapeCycle(t, g, "{C}{C}{C}")
	if effectstest.HasPayUnlessFor(g, opp.ID) || effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("another player's cycling offered a payment")
	}
}
