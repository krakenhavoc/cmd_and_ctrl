package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b414_test.go — Gatecrash, batch 414 (#5565): every card played
// through the engine.

const (
	oldB414Blockade   = "5cb2e859-7737-4b25-a6be-b72e7b5bfada"
	oldB414OozeFlux   = "631d0de7-4810-4bc3-b86b-8e2505ed91f4"
	oldB414Guildmage  = "7a49f3a1-abdb-4bf3-b5ed-f54349ab40fd"
	oldB414Watcher    = "a6297680-5d32-41e6-aaaf-b194e6be6b96"
	oldB414Beckon     = "10762505-ebaa-47de-9f3e-08d48f2ac240"
	oldB414Charm      = "26f8626a-ce64-450c-bda5-833df2fbb4fb"
	oldB414Servitude  = "c946be2c-1e8f-46bc-ab75-861926d4aa63"
	oldB414Strike     = "17694111-5623-48b3-8c54-c383418e8d2b"
	oldB414Genesis    = "58940a7e-3827-43fa-8bf5-cc88dc295df0"
	oldB414Rupture    = "934c50aa-7d2e-4c9e-b91b-1e8c4055e1c6"
	oldB414CreatureTL = "Creature — Bear"
)

func TestOldB414AllRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB414Blockade:  "Corpse Blockade",
		oldB414OozeFlux:  "Ooze Flux",
		oldB414Guildmage: "Skarrg Guildmage",
		oldB414Watcher:   "Greenside Watcher",
		oldB414Beckon:    "Beckon Apparition",
		oldB414Charm:     "Orzhov Charm",
		oldB414Servitude: "Immortal Servitude",
		oldB414Strike:    "Psychic Strike",
		oldB414Genesis:   "Mystic Genesis",
		oldB414Rupture:   "Spell Rupture",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func oldB414OnBattlefield(g *game.Game, id uuid.UUID) bool { return g.Battlefield.Contains(id) }

// --- Corpse Blockade ---------------------------------------------------------

func TestOldB414CorpseBlockadeSacrificesAnotherCreatureForDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	blockade := effectstest.PushCatalogPermanent(g, me.ID, "Corpse Blockade", "Creature — Zombie", oldB414Blockade, false)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", oldB414CreatureTL, 1, 1)
	if effectstest.HasEffectiveKeyword(t, g, blockade, "deathtouch") {
		t.Fatal("the Blockade starts with deathtouch")
	}
	if err := g.ActivateCatalogAbility(me.ID, blockade, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{blockade}}); err == nil {
		t.Fatal("the Blockade sacrificed itself to its own ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, blockade, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB414OnBattlefield(g, fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if !effectstest.HasEffectiveKeyword(t, g, blockade, "deathtouch") {
		t.Error("the Blockade did not gain deathtouch")
	}
	if !effectstest.HasEffectiveKeyword(t, g, blockade, "defender") {
		t.Error("the Blockade lost defender")
	}
}

// --- Ooze Flux -----------------------------------------------------------------

func TestOldB414OozeFluxMakesAnOozeAsBigAsTheCountersRemovedAcrossCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	flux := effectstest.PushCatalogPermanent(g, me.ID, "Ooze Flux", "Enchantment", oldB414OozeFlux, false)
	a := effectstest.B12Creature(g, me.ID, "Bear A", oldB414CreatureTL, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", oldB414CreatureTL, 2, 2)
	effectstest.AddCountersForTest(g, a, game.CounterPlusOne, 2)
	effectstest.AddCountersForTest(g, b, game.CounterPlusOne, 1)
	effectstest.B06AddMana(me, "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, flux, 0, game.ActivateAbilityParams{Strict: true,
		CounterSourceIDs: []uuid.UUID{a, b}, CounterCounts: []int{2, 1}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	oozes := effectstest.SocB17Names(g, me.ID, "Ooze")
	if len(oozes) != 1 {
		t.Fatalf("%d Oozes, want 1", len(oozes))
	}
	if p, tough := effectstest.PtOf(t, g, oozes[0].InstanceID); p != 3 || tough != 3 {
		t.Errorf("the Ooze is %d/%d, want 3/3", p, tough)
	}
	if n := effectstest.CounterOn(g, a, game.CounterPlusOne) + effectstest.CounterOn(g, b, game.CounterPlusOne); n != 0 {
		t.Errorf("%d +1/+1 counters left on the creatures, want 0", n)
	}
}

func TestOldB414OozeFluxNeedsAtLeastOneCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	flux := effectstest.PushCatalogPermanent(g, me.ID, "Ooze Flux", "Enchantment", oldB414OozeFlux, false)
	a := effectstest.B12Creature(g, me.ID, "Bear A", oldB414CreatureTL, 2, 2)
	effectstest.B06AddMana(me, "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, flux, 0, game.ActivateAbilityParams{Strict: true,
		CounterSourceIDs: []uuid.UUID{a}, CounterCounts: []int{1}}); err == nil {
		t.Error("the cost was paid with a counter the creature does not have")
	}
}

// --- Skarrg Guildmage ------------------------------------------------------------

func TestOldB414SkarrgGuildmageGivesTrampleToCreaturesYouControlNow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	if opp.ID == me.ID {
		opp = g.Seats[0]
	}
	effectstest.AdvanceToMain(t, g)
	mage := effectstest.PushCatalogPermanent(g, me.ID, "Skarrg Guildmage", "Creature — Human Shaman", oldB414Guildmage, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB414CreatureTL, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB414CreatureTL, 2, 2)
	effectstest.B06AddMana(me, "R", "G")
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	late := effectstest.B12Creature(g, me.ID, "Late Bear", oldB414CreatureTL, 2, 2)
	for _, id := range []uuid.UUID{mage, bear} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "trample") {
			t.Error("one of my creatures did not gain trample")
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature gained trample")
	}
	if effectstest.HasEffectiveKeyword(t, g, late, "trample") {
		t.Error("a creature that arrived afterwards gained trample")
	}
}

func TestOldB414SkarrgGuildmageAnimatesATargetLandIntoA4x4Elemental(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	if opp.ID == me.ID {
		opp = g.Seats[0]
	}
	effectstest.AdvanceToMain(t, g)
	mage := effectstest.PushCatalogPermanent(g, me.ID, "Skarrg Guildmage", "Creature — Human Shaman", oldB414Guildmage, false)
	land := effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	theirs := effectstest.SeedLand(g, opp.ID, "Island", "Basic Land — Island", "")
	effectstest.B06AddMana(me, "R", "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, mage, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(theirs)}); err == nil {
		t.Fatal("targeted a land I do not control")
	}
	if err := g.ActivateCatalogAbility(me.ID, mage, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(land)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	view := effectstest.GaLayered(t, g, land)
	if !view.IsCreature() || !view.IsLand() {
		t.Errorf("the land is creature=%v land=%v, want both", view.IsCreature(), view.IsLand())
	}
	if !view.HasSubtype("Elemental") {
		t.Error("the land is not an Elemental")
	}
	if p, tough := effectstest.PtOf(t, g, land); p != 4 || tough != 4 {
		t.Errorf("the land is %d/%d, want 4/4", p, tough)
	}
	if effectstest.GaLayered(t, g, theirs).IsCreature() {
		t.Error("their land became a creature")
	}
}

// --- Greenside Watcher -----------------------------------------------------------------

func TestOldB414GreensideWatcherUntapsOnlyAGate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	watcher := effectstest.PushCatalogPermanent(g, me.ID, "Greenside Watcher", "Creature — Elf Druid", oldB414Watcher, false)
	gate := effectstest.SeedLand(g, me.ID, "Test Gate", "Land — Gate", "")
	forest := effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, gate).Tapped = true
		effectstest.FindBattlefieldCardForTest(g, forest).Tapped = true
	})
	if err := g.ActivateCatalogAbility(me.ID, watcher, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(forest)}); err == nil {
		t.Fatal("untapped a non-Gate land")
	}
	if err := g.ActivateCatalogAbility(me.ID, watcher, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(gate)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.GaLayered(t, g, gate).Tapped {
		t.Error("the Gate is still tapped")
	}
	if !effectstest.GaLayered(t, g, watcher).Tapped {
		t.Error("the Watcher did not tap as a cost")
	}
}

// --- Beckon Apparition ----------------------------------------------------------------------

func TestOldB414BeckonApparitionExilesACardAndMakesAFlyingSpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	if opp.ID == me.ID {
		opp = g.Seats[0]
	}
	card := effectstest.PushGraveyardCardForTest(opp, "Their Dead Card")
	effectstest.CastCatalogSpell(t, g, "Beckon Apparition", "Instant", oldB414Beckon, effectstest.CardTargets(card))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(card) {
		t.Error("the targeted card is still in the graveyard")
	}
	spirits := effectstest.SocB17Names(g, me.ID, "Spirit")
	if len(spirits) != 1 {
		t.Fatalf("%d Spirits, want 1", len(spirits))
	}
	view := effectstest.GaLayered(t, g, spirits[0].InstanceID)
	if p, tough := effectstest.PtOf(t, g, spirits[0].InstanceID); p != 1 || tough != 1 {
		t.Errorf("the Spirit is %d/%d, want 1/1", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, spirits[0].InstanceID, "flying") {
		t.Error("the Spirit has no flying")
	}
	if len(view.Colors) != 2 {
		t.Errorf("the Spirit's colours are %v, want white and black", view.Colors)
	}
}

func TestOldB414BeckonApparitionNeedsACardInAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.SeedBear(g, g.Seats[1].ID)
	if err := effectstest.CastCatalogSpellErr(t, g, "Beckon Apparition", "Instant", oldB414Beckon, effectstest.CardTargets(bear)); err == nil {
		t.Error("Beckon Apparition targeted a card on the battlefield")
	}
}

// --- Orzhov Charm ---------------------------------------------------------------------------------

func oldB414CastCharm(t *testing.T, g *game.Game, mode int, target uuid.UUID) error {
	t.Helper()
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Orzhov Charm", TypeLine: "Instant", OracleID: oldB414Charm, Owner: me.ID, Controller: me.ID})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{mode}, Targets: effectstest.CardTargets(target)})
}

func TestOldB414OrzhovCharmBouncesMyCreatureAndMyAurasButNotAnOpponentsAura(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[1]
	if opp.ID == me.ID {
		opp = g.Seats[0]
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB414CreatureTL, 2, 2)
	mine := effectstest.B15Aura(g, me.ID, bear)
	theirs := effectstest.B15Aura(g, opp.ID, bear)
	theirBear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB414CreatureTL, 2, 2)
	if err := oldB414CastCharm(t, g, 0, theirBear); err == nil {
		t.Fatal("mode 1 targeted a creature I do not control")
	}
	if err := oldB414CastCharm(t, g, 0, bear); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) || !me.Hand.Contains(mine) {
		t.Error("the creature and my Aura did not both return to my hand")
	}
	// The opponent's Aura was never returned to a hand; it fell off the host (CR 704.5m).
	if opp.Hand.Contains(theirs) || !opp.Graveyard.Contains(theirs) {
		t.Error("an Aura an opponent controls was not left to fall into its owner's graveyard")
	}
}

func TestOldB414OrzhovCharmDestroysACreatureAndCostsLifeEqualToItsToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[1]
	if opp.ID == me.ID {
		opp = g.Seats[0]
	}
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
	before := effectstest.LifeOf(g, me.ID)
	if err := oldB414CastCharm(t, g, 1, wall); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wall) {
		t.Error("the creature survived")
	}
	if got := effectstest.LifeOf(g, me.ID); got != before-5 {
		t.Errorf("life %d -> %d, want -5", before, got)
	}
}

func TestOldB414OrzhovCharmReturnsAnInexpensiveCreatureFromMyGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cheap, dear := uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: cheap, Name: "Cheap Bear", TypeLine: oldB414CreatureTL, ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: dear, Name: "Dear Bear", TypeLine: oldB414CreatureTL, ManaCost: "{2}", Owner: me.ID, Controller: me.ID})
	if err := oldB414CastCharm(t, g, 2, dear); err == nil {
		t.Fatal("returned a creature card with mana value 2")
	}
	if err := oldB414CastCharm(t, g, 2, cheap); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Error("the creature card did not return to the battlefield")
	}
}

// --- Immortal Servitude ----------------------------------------------------------------------------

func TestOldB414ImmortalServitudeReturnsEveryCreatureCardWithManaValueX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	add := func(name, typ, cost string) uuid.UUID {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typ, ManaCost: cost, Owner: me.ID, Controller: me.ID})
		return id
	}
	two1, two2 := add("Two A", oldB414CreatureTL, "{1}{W}"), add("Two B", oldB414CreatureTL, "{B}{B}")
	three, spell := add("Three", oldB414CreatureTL, "{2}{B}"), add("Two Spell", "Sorcery", "{1}{B}")
	effectstest.ToMainForCost(t, g)
	effectstest.B06AddMana(me, "W", "W", "W", "W", "W")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Immortal Servitude", TypeLine: "Sorcery", OracleID: oldB414Servitude, ManaCost: "{X}{W/B}{W/B}{W/B}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 2}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []uuid.UUID{two1, two2} {
		if !g.Battlefield.Contains(c) {
			t.Error("a creature card with mana value 2 stayed in the graveyard")
		}
	}
	if g.Battlefield.Contains(three) || !me.Graveyard.Contains(three) {
		t.Error("the mana value 3 creature card returned")
	}
	if g.Battlefield.Contains(spell) {
		t.Error("a sorcery card returned")
	}
}

// --- Psychic Strike / Mystic Genesis / Spell Rupture ---------------------------------------------------

// oldB414TheirSpell puts a sorcery the active seat cast on the stack and passes
// priority to the other seat, which will answer it.
func oldB414TheirSpell(t *testing.T, g *game.Game, typeLine, cost string) uuid.UUID {
	t.Helper()
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Their Spell", TypeLine: typeLine, ManaCost: cost, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	return id
}

func oldB414Responder(g *game.Game) *game.Player {
	for _, p := range g.Seats {
		if p.ID != g.Seats[g.Turn.ActiveSeat].ID {
			return p
		}
	}
	return nil
}

func TestOldB414PsychicStrikeCountersTheSpellAndMillsItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	effectstest.Msc18Library(caster, 6)
	victim := oldB414TheirSpell(t, g, "Sorcery", "")
	libBefore := caster.Library.Size()
	effectstest.CastFromSeat(t, g, oldB414Responder(g), "Psychic Strike", "Instant", oldB414Strike, effectstest.CardTargets(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(victim) {
		t.Error("the spell is still on the stack")
	}
	if got := caster.Library.Size(); got != libBefore-2 {
		t.Errorf("the spell's controller's library %d -> %d, want -2", libBefore, got)
	}
}

func TestOldB414MysticGenesisMakesAnOozeEqualToTheSpellsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := oldB414TheirSpell(t, g, "Sorcery", "{2}{R}")
	responder := oldB414Responder(g)
	effectstest.CastFromSeat(t, g, responder, "Mystic Genesis", "Instant", oldB414Genesis, effectstest.CardTargets(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(victim) {
		t.Error("the spell is still on the stack")
	}
	oozes := effectstest.SocB17Names(g, responder.ID, "Ooze")
	if len(oozes) != 1 {
		t.Fatalf("%d Oozes, want 1", len(oozes))
	}
	if p, tough := effectstest.PtOf(t, g, oozes[0].InstanceID); p != 3 || tough != 3 {
		t.Errorf("the Ooze is %d/%d, want 3/3", p, tough)
	}
}

func oldB414Rupture4(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	victim := oldB414TheirSpell(t, g, oldB414CreatureTL, "")
	responder := oldB414Responder(g)
	effectstest.B12Creature(g, responder.ID, "Big", oldB414CreatureTL, 4, 4)
	effectstest.B12Creature(g, responder.ID, "Small", oldB414CreatureTL, 1, 1)
	effectstest.B12Creature(g, caster.ID, "Their Giant", oldB414CreatureTL, 9, 9)
	effectstest.CastFromSeat(t, g, responder, "Spell Rupture", "Instant", oldB414Rupture, effectstest.CardTargets(victim))
	effectstest.PassUntilTaxed(t, g, caster.ID)
	if !g.Stack.Contains(victim) {
		t.Fatal("the spell resolved while the tax was unanswered")
	}
	return g, caster, victim
}

func TestOldB414SpellRuptureTaxesTheGreatestPowerOfItsControllersCreatures(t *testing.T) {
	g, caster, victim := oldB414Rupture4(t)
	// X is the greatest power among creatures the Rupture's controller controls: 4, not 9.
	effectstest.FloatMana(t, g, caster, "{C}{C}{C}{C}")
	effectstest.AnswerPayUnless(t, g, caster.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := caster.ManaPool.Total(); got != 0 {
		t.Errorf("%d mana left in the pool, want 0 (X = 4 was paid)", got)
	}
	if caster.Graveyard.Contains(victim) || !g.Battlefield.Contains(victim) {
		t.Error("the creature spell was countered though its controller paid")
	}
}

func TestOldB414SpellRuptureCountersTheSpellWhenTheTaxIsDeclined(t *testing.T) {
	g, caster, victim := oldB414Rupture4(t)
	effectstest.AnswerPayUnless(t, g, caster.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || !caster.Graveyard.Contains(victim) {
		t.Error("the creature spell was not countered")
	}
}

func TestOldB414SpellRuptureDoesNothingWithoutCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := oldB414TheirSpell(t, g, oldB414CreatureTL, "")
	responder := oldB414Responder(g)
	effectstest.CastFromSeat(t, g, responder, "Spell Rupture", "Instant", oldB414Rupture, effectstest.CardTargets(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Error("the creature spell should have resolved untouched (X = 0)")
	}
}
