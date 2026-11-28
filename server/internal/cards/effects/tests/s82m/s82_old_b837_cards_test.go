package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b837_cards_test.go — Ice Age's draw-and-library cards, played from the
// real catalog entries (S82 OLD batch 837, #5636).

const (
	b837TouchOracle     = "84f1e596-f56f-48c9-9bff-2c0bb2cea2d3"
	b837ClairOracle     = "851333be-d7d5-472a-95cb-0a28da09ce24"
	b837PykniteOracle   = "857f9a8d-ba77-4afd-922d-994f66e478ef"
	b837RavelOracle     = "8cd0fcac-800a-431c-aca9-804998704606"
	b837SpiritOracle    = "9798d307-a44a-4f44-94c7-03e499b94962"
	b837SextantOracle   = "97ba11a1-16cd-47ee-a187-1e919c468e4e"
	b837FormationOracle = "ad5d27ea-cead-4c95-8561-283855ce952e"
	b837CharmOracle     = "e0d638d2-2c5b-415c-95e7-74bf8d88794f"
	b837PanicOracle     = "e98429a7-a83f-4b69-afad-9eb68b09e38e"
	b837EnervateOracle  = "f763d95e-87af-473f-9b1f-cf16221e5792"
	b837UpdraftOracle   = "ff0d5cbe-99d3-4d3b-9a5b-80223fc18869"
)

func TestS82OldB837CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b837TouchOracle:     "Touch of Vitae",
		b837ClairOracle:     "Clairvoyance",
		b837PykniteOracle:   "Pyknite",
		b837RavelOracle:     "Mind Ravel",
		b837SpiritOracle:    "Sibilant Spirit",
		b837SextantOracle:   "Barbed Sextant",
		b837FormationOracle: "Formation",
		b837CharmOracle:     "Freyalise's Charm",
		b837PanicOracle:     "Panic",
		b837EnervateOracle:  "Enervate",
		b837UpdraftOracle:   "Updraft",
	} {
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

// b837DrawsAtTheNextUpkeep asserts `p` draws one card at the next turn's upkeep
// and not before.
func b837DrawsAtTheNextUpkeep(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	before := p.Hand.Size()
	next := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	effectstest.AdvanceToUpkeepOf(t, g, next)
	if p.Hand.Size() != before {
		t.Fatalf("drew before the next upkeep's trigger resolved: hand %d -> %d", before, p.Hand.Size())
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p.Hand.Size() != before+1 {
		t.Errorf("hand %d -> %d at the next upkeep, want +1", before, p.Hand.Size())
	}
}

func b837CastInstant(t *testing.T, g *game.Game, p *game.Player, name, oracle string, targets []game.TargetRef) error {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", OracleID: oracle, Owner: p.ID, Controller: p.ID})
	return g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true, Targets: targets})
}

func TestClairvoyanceLooksAtTheHandAndDrawsLater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.AdvanceToMain(t, g)
	secret := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: secret, Name: "Secret", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	if err := b837CastInstant(t, g, me, "Clairvoyance", b837ClairOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	known := false
	for _, c := range opp.Hand.Cards {
		if c.InstanceID == secret {
			known = c.IsKnownTo(me.ID)
		}
	}
	if !known {
		t.Error("the caster did not learn the target's hand")
	}
	b837DrawsAtTheNextUpkeep(t, g, me)
}

func TestMindRavelMakesTheTargetDiscardAndDrawsLater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.AdvanceToMain(t, g)
	keep := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: keep, Name: "Only", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Mind Ravel", TypeLine: "Sorcery", OracleID: b837RavelOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, opp.ID, keep)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(keep) {
		t.Error("the target did not discard")
	}
	b837DrawsAtTheNextUpkeep(t, g, me)
}

func TestPyknitesEntersTriggerDrawsAtTheNextUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.CastHoldCreature(t, g, "Pyknite", "Creature — Ouphe", b837PykniteOracle, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	b837DrawsAtTheNextUpkeep(t, g, me)
}

func TestEnervateTapsAnArtifactCreatureOrLandOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.AdvanceToMain(t, g)
	ench := effectstest.B12Push(g, opp.ID, "Aura Thing", "Enchantment", "", 0, 0)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := b837CastInstant(t, g, me, "Enervate", b837EnervateOracle, effectstest.CardTarget(ench)); err == nil {
		t.Fatal("targeted an enchantment: accepted")
	}
	if err := b837CastInstant(t, g, me, "Enervate", b837EnervateOracle, effectstest.CardTarget(bear)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, bear) {
		t.Error("the creature was not tapped")
	}
	b837DrawsAtTheNextUpkeep(t, g, me)
}

func TestUpdraftGrantsFlyingUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := b837CastInstant(t, g, me, "Updraft", b837UpdraftOracle, effectstest.CardTarget(bear)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), "flying") {
		t.Error("the creature did not gain flying")
	}
	b837DrawsAtTheNextUpkeep(t, g, me)
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), "flying") {
		t.Error("flying outlived the turn")
	}
}

func TestFormationGrantsBandingUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := b837CastInstant(t, g, me, "Formation", b837FormationOracle, effectstest.CardTarget(bear)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), game.KeywordBanding) {
		t.Error("the creature did not gain banding")
	}
	b837DrawsAtTheNextUpkeep(t, g, me)
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), game.KeywordBanding) {
		t.Error("banding outlived the turn")
	}
}

func TestPanicIsCastOnlyInCombatBeforeBlockersAndStopsABlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	atk := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	blocker := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4})
	effectstest.AdvanceToMain(t, g)
	if err := b837CastInstant(t, g, me, "Panic", b837PanicOracle, effectstest.CardTarget(blocker)); err == nil {
		t.Fatal("cast in the main phase: accepted")
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := b837CastInstant(t, g, me, "Panic", b837PanicOracle, effectstest.CardTarget(blocker)); err != nil {
		t.Fatalf("cast in combat: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, atk); err == nil {
		t.Error("the creature blocked despite Panic")
	}
	if err := b837CastInstant(t, g, me, "Panic", b837PanicOracle, effectstest.CardTarget(blocker)); err == nil {
		t.Error("Panic was cast in the declare blockers step")
	}
}

func TestTouchOfVitaeGrantsHasteAndOneFreeUntapUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := b837CastInstant(t, g, me, "Touch of Vitae", b837TouchOracle, effectstest.CardTarget(bear)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), "haste") {
		t.Error("the creature did not gain haste")
	}
	idx, ref := effectstest.GrantedActivatedIndex(t, g, bear)
	if idx < 0 {
		t.Fatal("the target has no granted untap")
	}
	effectstest.TapForTest(t, g, bear)
	effectstest.B16Activate(t, g, me.ID, bear, idx, game.ActivateAbilityParams{Ref: ref})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, bear) {
		t.Error("the untap did not untap the creature")
	}
	effectstest.TapForTest(t, g, bear)
	if err := g.ActivateCatalogAbility(me.ID, bear, idx, game.ActivateAbilityParams{Ref: ref}); err == nil {
		t.Error("the untap was activated a second time")
	}
	b837DrawsAtTheNextUpkeep(t, g, me)
}

func TestBarbedSextantMakesAnyColorAndDrawsLater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.AdvanceToMain(t, g)
	sextant := effectstest.B12Push(g, me.ID, "Barbed Sextant", "Artifact", b837SextantOracle, 0, 0)
	if err := g.ActivateManaAbility(me.ID, sextant, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("{1} is a cost: with an empty pool the Sextant stays")
	}
	if !g.Battlefield.Contains(sextant) {
		t.Fatal("a failed activation leaves the Sextant untouched")
	}
	effectstest.B06AddMana(me, "C")
	if err := g.ActivateManaAbility(me.ID, sextant, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if pick := effectstest.RiderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("any colour: %+v", pick)
	}
	if g.Battlefield.Contains(sextant) || !me.Graveyard.Contains(sextant) {
		t.Fatal("the Sextant was not sacrificed as a cost")
	}
	if effectstest.B10ResolveAllManaPicks(t, g, me.ID, "U") != 1 {
		t.Fatal("one colour pick")
	}
	b837DrawsAtTheNextUpkeep(t, g, me)
}

func TestSibilantSpiritLetsTheDefendingPlayerDraw(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.SeedLibrary(opp, "A", "B", "C", "D")
		spirit := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Sibilant Spirit", TypeLine: "Creature — Spirit", OracleID: b837SpiritOracle, Power: 5, Toughness: 6})
		oppHand, myHand := opp.Hand.Size(), me.Hand.Size()
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(spirit, opp.ID); err != nil {
			t.Fatal(err)
		}
		effectstest.LockInAttacks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerMayChoice(t, g, opp.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		want := oppHand
		if yes {
			want++
		}
		if opp.Hand.Size() != want {
			t.Errorf("yes=%v: the defender's hand is %d, want %d", yes, opp.Hand.Size(), want)
		}
		if me.Hand.Size() != myHand {
			t.Errorf("yes=%v: the attacker's hand changed", yes)
		}
	}
}

func TestFreyalisesCharmDrawsOnlyWhenPaidForAnOpponentsBlackSpell(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.SeedLibrary(me, "A", "B", "C", "D")
		effectstest.AdvanceToMainOf(t, g, 1)
		effectstest.B12Push(g, me.ID, "Freyalise's Charm", "Enchantment", b837CharmOracle, 0, 0)
		spell := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: spell, Name: "Black Spell", TypeLine: "Sorcery", ManaCost: "{B}", Colors: []string{"B"}, Owner: opp.ID, Controller: opp.ID})
		effectstest.ApaMana(opp, "B")
		effectstest.ApaMana(me, "G", "G")
		hand := me.Hand.Size()
		if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{}); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		want := hand
		if pay {
			want++
		}
		if me.Hand.Size() != want {
			t.Errorf("pay=%v: hand %d, want %d", pay, me.Hand.Size(), want)
		}
	}
}

func TestFreyalisesCharmIgnoresANonblackSpellAndReturnsToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMainOf(t, g, 1)
	charm := effectstest.B12Push(g, me.ID, "Freyalise's Charm", "Enchantment", b837CharmOracle, 0, 0)
	red := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: red, Name: "Red Spell", TypeLine: "Sorcery", ManaCost: "{R}", Colors: []string{"R"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.ApaMana(opp, "R")
	if err := g.CastSpell(opp.ID, red, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("a red spell offered the payment")
	}
	effectstest.ApaMana(me, "G", "G")
	effectstest.ApaActivate(t, g, me, charm, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(charm) {
		t.Error("the Charm did not return to its owner's hand")
	}
}
