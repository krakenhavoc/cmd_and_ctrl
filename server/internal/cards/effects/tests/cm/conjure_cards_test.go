package testscm

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// conjure_cards_test.go — #5948 class (c) Q1: conjure, perpetual and
// boons (ADR 0099 amendment 2026-11-24), and the cards a conjure names.

const (
	cjEuruOracle     = "b01657ad-120f-4d00-a367-fd112d403700"
	cjJetOracle      = "191f5f95-bfb8-4d3c-a223-c7654cc507d8"
	cjPearlOracle    = "a5ce937f-425f-4c24-a9d3-1bf8a244632b"
	cjEmeraldOracle  = "d12534c8-6be7-41b0-852b-4461e7f962a1"
	cjSapphireOracle = "7de34763-6d81-4bec-92cf-61a0d15ad45f"
	cjRubyOracle     = "ca927d6b-77f7-49fa-a686-12fa981a8bd4"
	cjPinchyOracle   = "4ad0664d-fede-4b75-bb3f-53d5060ca01d"
	cjRuskoOracle    = "ccb456a0-3abd-4bde-a453-20d2da210e05"
	cjPupOracle      = "70c9cfcf-6f23-41de-af63-dbe212018035"
	cjToralfOracle   = "5bce8636-aa2f-4f0f-b374-60c1eb1d360a"
	cjOracleOracle   = "b77dcee1-f42e-4c9d-b6b7-6d5885826060"
	cjEvangelOracle  = "a5550795-8cd3-4f29-a9df-02c48b124dbc"
	cjSanguineOracle = "d2c6502d-fefd-4c9f-9d7b-6dfcc43316e1"

	cjMoxJetOracle      = "0677f49e-f8bf-4349-af52-2ccde9287c2e"
	cjMoxPearlOracle    = "824597b8-c89a-47ec-8526-7efc6e24ef0e"
	cjMoxEmeraldOracle  = "376ee366-e082-402f-b4db-6592fcfcacd2"
	cjMoxSapphireOracle = "d5ed1233-df87-4b90-8918-13922ec95249"
	cjMoxRubyOracle     = "ed85fa82-e4fa-434b-92a8-36b6075708d1"
	cjLotusOracle       = "5089ec1a-f881-4d55-af14-5d996171203b"
	cjRecallOracle      = "550c74d4-1fcb-406a-b02a-639a760a4380"
	cjTimeWalkOracle    = "d0209d3f-3f7e-4fd5-bce5-10bce6f29c86"
	cjTwisterOracle     = "c823e687-6311-4c99-974b-fd77d204141a"
	cjLethalOracle      = "464749e1-9df2-47d0-8fcd-516f566022b5"
	cjDeathOracle       = "25895c28-23c1-439e-858b-5cb31fdc31de"
)

func TestConjureCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		cjEuruOracle: "Euru, Acorn Scrounger", cjJetOracle: "Jet Collector", cjPearlOracle: "Pearl Collector",
		cjEmeraldOracle: "Emerald Collector", cjSapphireOracle: "Sapphire Collector", cjRubyOracle: "Ruby Collector",
		cjPinchyOracle: "Pinchy McStingbutt", cjRuskoOracle: "Rusko, Clockmaker", cjPupOracle: "Tenacious Pup",
		cjToralfOracle: "Toralf's Disciple", cjOracleOracle: "Oracle of the Alpha", cjEvangelOracle: "Sigardian Evangel",
		cjSanguineOracle: "Sanguine Brushstroke",
		cjMoxJetOracle:   "Mox Jet", cjMoxPearlOracle: "Mox Pearl", cjMoxEmeraldOracle: "Mox Emerald",
		cjMoxSapphireOracle: "Mox Sapphire", cjMoxRubyOracle: "Mox Ruby", cjLotusOracle: "Black Lotus",
		cjRecallOracle: "Ancestral Recall", cjTimeWalkOracle: "Time Walk", cjTwisterOracle: "Timetwister",
		cjLethalOracle: "Lethal Sting", cjDeathOracle: "Death of a Thousand Stings",
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

// Every conjurable record is a catalog card, so a conjured card does what
// it prints.
func TestEveryConjurableIsACatalogCard(t *testing.T) {
	names := game.ConjurableNames()
	if len(names) < 17 {
		t.Fatalf("%d conjurable records, want at least 17", len(names))
	}
	for _, n := range names {
		tmpl, _ := game.ConjurableTemplate(n)
		spec, ok := effects.Lookup(tmpl.OracleID)
		if !ok || spec.Name != n {
			t.Errorf("conjurable %q (oracle %s) is not the catalog card of that name", n, tmpl.OracleID)
		}
		if tmpl.ScryfallID == "" || tmpl.TypeLine == "" {
			t.Errorf("conjurable %q has no printing id or type line", n)
		}
	}
}

// --- helpers -------------------------------------------------------------

// cjConjured lists the conjured cards named `name` in a zone.
func cjConjured(z *game.Zone, name string) []game.Card {
	var out []game.Card
	if z == nil {
		return nil
	}
	for _, c := range z.Cards {
		if c.Conjured && c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

// cjEnter puts a catalog card onto the battlefield from its owner's hand
// and settles what it set off.
func cjEnter(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	id := effectstest.EnterFromHand(t, g, owner, name, typeLine, oracle)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

// cjCard is a catalog card for a hand.
func cjCard(owner uuid.UUID, name, typeLine, cost, oracle string) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle, Owner: owner, Controller: owner}
}

// cjSettle answers any trigger-ordering prompt in the order offered and
// passes priority until the stack is empty.
func cjSettle(t *testing.T, g *game.Game, me uuid.UUID) {
	t.Helper()
	for i := 0; i < 8; i++ {
		effectstest.AnswerOpenOrderingIfAny(t, g, me)
		effectstest.PassPriorityAroundTable(t, g)
		if len(g.PendingChoices) == 0 {
			return
		}
	}
}

func cjConjuredEvents(g *game.Game, name string) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventConjured && ev.Label == name {
			n++
		}
	}
	return n
}

// --- conjure into a library ----------------------------------------------

// Toralf's Disciple attacks: four real Lightning Bolts join the library,
// then it is shuffled. A conjured Bolt drawn and cast deals 3.
func TestToralfsDiscipleConjuresFourBoltsAndTheyAreRealCards(t *testing.T) {
	g, me, opp := effectstest.EoTable(t)
	disciple := effectstest.PushCatalogPermanent(g, me.ID, "Toralf's Disciple", "Creature — Human Warrior", cjToralfOracle, false)
	before := me.Library.Size()
	effectstest.DeclareAttack(t, g, opp.ID, disciple)
	effectstest.PassPriorityAroundTable(t, g)
	bolts := cjConjured(me.Library, "Lightning Bolt")
	if len(bolts) != 4 || me.Library.Size() != before+4 {
		t.Fatalf("library has %d conjured Bolts and %d cards, want 4 and %d", len(bolts), me.Library.Size(), before+4)
	}
	for _, b := range bolts {
		if b.Owner != me.ID || b.OracleID != "4457ed35-7c10-48c8-9776-456485fdf070" || b.ManaCost != "{R}" || b.ScryfallID == "" {
			t.Errorf("conjured Bolt %+v is not the printed card owned by me", b)
		}
		if len(b.KnownBy) != 0 {
			t.Errorf("a card conjured into a library is known to %v; nobody should know it", b.KnownBy)
		}
	}
	if n := cjConjuredEvents(g, "Lightning Bolt"); n != 4 {
		t.Errorf("%d conjure events, want 4", n)
	}

	// "Treat conjured cards like cards": move one to hand and cast it.
	id := bolts[0].InstanceID
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneLibrary, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, id); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceToStepOf(t, g, g.Turn.ActiveSeat, game.StepPostcombatMain)
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	life := opp.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast the conjured Bolt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-3)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the conjured Bolt did not go to its owner's graveyard")
	}
}

// Oracle of the Alpha: the Power Nine, nine cards, into the library.
func TestOracleOfTheAlphaConjuresThePowerNine(t *testing.T) {
	g, me, _ := effectstest.EoTable(t)
	cjEnter(t, g, me.ID, "Oracle of the Alpha", "Creature — Bird Wizard", cjOracleOracle)
	for _, n := range effects.PowerNine {
		if got := len(cjConjured(me.Library, n)); got != 1 {
			t.Errorf("%d conjured %s in the library, want 1", got, n)
		}
	}
}

// --- conjure into a hand ---------------------------------------------------

// Jet Collector: the intervening if (four cards in your graveyard) and
// "triggers only once".
func TestJetCollectorConjuresMoxJetOnceWithFourCardsInGraveyard(t *testing.T) {
	g, me, _ := effectstest.EoTable(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Jet Collector", "Creature — Merfolk Scout", cjJetOracle, false)
	effectstest.FgGraveyard(me, 3)
	seat := g.Turn.ActiveSeat
	effectstest.AdvanceToStepOf(t, g, seat, game.StepPostcombatMain)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(cjConjured(me.Hand, "Mox Jet")); n != 0 {
		t.Fatalf("three cards in the graveyard conjured %d Mox Jet", n)
	}
	effectstest.FgGraveyard(me, 1)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToStepOf(t, g, seat, game.StepPostcombatMain)
	effectstest.PassPriorityAroundTable(t, g)
	moxen := cjConjured(me.Hand, "Mox Jet")
	if len(moxen) != 1 {
		t.Fatalf("%d Mox Jet in hand, want 1", len(moxen))
	}
	if !moxen[0].KnownBy[me.ID] {
		t.Error("the conjured card in hand is not known to its owner")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToStepOf(t, g, seat, game.StepPostcombatMain)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(cjConjured(me.Hand, "Mox Jet")); n != 1 {
		t.Errorf("%d Mox Jet after a second turn, want 1: it triggers only once", n)
	}

	// The conjured Mox is a real Mox: cast it and tap it for {B}.
	id := moxen[0].InstanceID
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Mox Jet: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap Mox Jet: %v", err)
	}
	if me.ManaPool.CountColor("B") != 1 {
		t.Errorf("Mox Jet made %d {B}, want 1", me.ManaPool.CountColor("B"))
	}
}

// Pearl Collector: four life gained conjures Mox Pearl; "perpetually
// gains lifelink" stays with the card through zone changes.
func TestPearlCollectorPerpetualLifelinkFollowsTheCard(t *testing.T) {
	g, me, _ := effectstest.EoTable(t)
	pearl := effectstest.PushCatalogPermanent(g, me.ID, "Pearl Collector", "Creature — Human Warlock", cjPearlOracle, false)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	for _, c := range []string{"W", "C", "C"} {
		me.ManaPool.AddMana(game.ManaToken{Color: c})
	}
	if err := g.ActivateCatalogAbility(me.ID, pearl, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: pearl}}}); err == nil {
		t.Fatal("Pearl Collector targeted itself; the ability says another creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, pearl, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, bear); !game.HasKeyword(&c, "lifelink") {
		t.Fatal("the Bear did not gain lifelink")
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, bear); err != nil {
		t.Fatal(err)
	}
	var inHand game.Card
	for _, c := range me.Hand.Cards {
		if c.InstanceID == bear {
			inHand = c
		}
	}
	if !slices.Contains(inHand.Effective().Abilities, "lifelink") {
		t.Error("the perpetual lifelink did not follow the card to its owner's hand")
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, bear); err != nil {
		t.Fatal(err)
	}
	if c, _ := effectstest.BattlefieldCard(g, bear); !game.HasKeyword(&c, "lifelink") {
		t.Error("the Bear lost its perpetual lifelink on returning to the battlefield")
	}

	// The second-main conjure needs four life gained this turn.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(pearl, me.ID, 4) })
	effectstest.AdvanceToStepOf(t, g, g.Turn.ActiveSeat, game.StepPostcombatMain)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(cjConjured(me.Hand, "Mox Pearl")); n != 1 {
		t.Errorf("%d Mox Pearl in hand after gaining 4 life, want 1", n)
	}
}

// Emerald Collector: the third card drawn in a turn conjures Mox Emerald.
func TestEmeraldCollectorConjuresOnTheThirdDraw(t *testing.T) {
	g, me, _ := effectstest.EoTable(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Emerald Collector", "Creature — Human Pirate", cjEmeraldOracle, false)
	drawn := g.TurnTallyFor(me.ID).CardsDrawn
	for i := drawn; i < 2; i++ {
		if err := g.DrawCard(me.ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(cjConjured(me.Hand, "Mox Emerald")); n != 0 {
		t.Fatalf("%d Mox Emerald before the third draw", n)
	}
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(cjConjured(me.Hand, "Mox Emerald")); n != 1 {
		t.Errorf("%d Mox Emerald after the third draw, want 1", n)
	}
}

// Sapphire Collector: the second noncreature spell in a turn.
func TestSapphireCollectorConjuresOnTheSecondNoncreatureSpell(t *testing.T) {
	g, me, _ := effectstest.EoTable(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Sapphire Collector", "Creature — Human Mercenary", cjSapphireOracle, false)
	for i, c := range []game.Card{
		cjCard(me.ID, "Mox Jet", "Artifact", "{0}", cjMoxJetOracle),
		cjCard(me.ID, "Mox Ruby", "Artifact", "{0}", cjMoxRubyOracle),
	} {
		me.Hand.PushTop(c)
		if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", c.Name, err)
		}
		cjSettle(t, g, me.ID)
		want := 0
		if i == 1 {
			want = 1
		}
		if n := len(cjConjured(me.Hand, "Mox Sapphire")); n != want {
			t.Errorf("after %d noncreature spells: %d Mox Sapphire, want %d", i+1, n, want)
		}
	}
}

// Ruby Collector: three attackers conjure Mox Ruby; two do not.
func TestRubyCollectorConjuresWhenYouAttackWithThree(t *testing.T) {
	g, me, opp := effectstest.EoTable(t)
	ruby := effectstest.PushCatalogPermanent(g, me.ID, "Ruby Collector", "Creature — Human Detective", cjRubyOracle, false)
	a := effectstest.PushCatalogPermanent(g, me.ID, "Bear A", "Creature — Bear", "", false)
	b := effectstest.PushCatalogPermanent(g, me.ID, "Bear B", "Creature — Bear", "", false)
	effectstest.DeclareAttack(t, g, opp.ID, ruby, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(cjConjured(me.Hand, "Mox Ruby")); n != 1 {
		t.Errorf("%d Mox Ruby after attacking with three, want 1", n)
	}
}

func TestRubyCollectorTwoAttackersConjureNothing(t *testing.T) {
	g, me, opp := effectstest.EoTable(t)
	ruby := effectstest.PushCatalogPermanent(g, me.ID, "Ruby Collector", "Creature — Human Detective", cjRubyOracle, false)
	a := effectstest.PushCatalogPermanent(g, me.ID, "Bear A", "Creature — Bear", "", false)
	effectstest.DeclareAttack(t, g, opp.ID, ruby, a)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(cjConjured(me.Hand, "Mox Ruby")); n != 0 {
		t.Errorf("%d Mox Ruby after attacking with two, want 0", n)
	}
}

// Pinchy McStingbutt: combat damage conjures one of three at random, and
// reveals it.
func TestPinchyConjuresOneStingAtRandomAndRevealsIt(t *testing.T) {
	g, me, opp := effectstest.EoTable(t)
	pinchy := effectstest.PushCatalogPermanent(g, me.ID, "Pinchy McStingbutt", "Creature — Scorpion", cjPinchyOracle, false)
	effectstest.TrekB25ToCombatDamage(t, g, opp.ID, pinchy)
	effectstest.PassPriorityAroundTable(t, g)
	var got []game.Card
	for _, n := range []string{"Death of a Thousand Stings", "Lethal Sting", "Stinging Shot"} {
		got = append(got, cjConjured(me.Hand, n)...)
	}
	if len(got) != 1 {
		t.Fatalf("%d stings conjured, want 1", len(got))
	}
	if !got[0].KnownBy[opp.ID] {
		t.Error("the conjured sting was not revealed to the opponent")
	}
}

// Sigardian Evangel: a conjured Evangel in hand, a tapped opposing
// permanent, and the conjured card discarded at the next end step.
func TestSigardianEvangelConjuresItselfThenDiscardsIt(t *testing.T) {
	g, me, opp := effectstest.EoTable(t)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)
	effectstest.EnterFromHand(t, g, me.ID, "Sigardian Evangel", "Creature — Human Cleric", cjEvangelOracle)
	for i := 0; i < 8; i++ {
		effectstest.AnswerOpenOrderingIfAny(t, g, me.ID)
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoicePickTarget {
				effectstest.AnswerPickTarget(t, g, theirs)
				break
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		if len(g.PendingChoices) == 0 && effectstest.StackFullyEmpty(g) {
			break
		}
	}
	if c, _ := effectstest.BattlefieldCard(g, theirs); !c.Tapped {
		t.Error("the opponent's permanent was not tapped")
	}
	conj := cjConjured(me.Hand, "Sigardian Evangel")
	if len(conj) != 1 {
		t.Fatalf("%d conjured Evangels in hand, want 1", len(conj))
	}
	effectstest.AdvanceToStepOf(t, g, g.Turn.ActiveSeat, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(conj[0].InstanceID) || !me.Graveyard.Contains(conj[0].InstanceID) {
		t.Error("the conjured Evangel was not discarded at the end step")
	}
}

// --- conjure onto the battlefield ------------------------------------------

// Rusko, Clockmaker: a real Midnight Clock enters; a noncreature spell
// puts an hour counter on it and drains.
func TestRuskoConjuresMidnightClockAndWindsIt(t *testing.T) {
	g, me, opp := effectstest.EoTable(t)
	cjEnter(t, g, me.ID, "Rusko, Clockmaker", "Legendary Creature — Human Artificer", cjRuskoOracle)
	clocks := cjConjured(g.Battlefield, "Midnight Clock")
	if len(clocks) != 1 || clocks[0].Controller != me.ID || clocks[0].Owner != me.ID {
		t.Fatalf("battlefield has %d conjured Midnight Clocks under me", len(clocks))
	}
	clock := clocks[0].InstanceID
	mine, theirs := me.Life, opp.Life
	mox := cjCard(me.ID, "Mox Pearl", "Artifact", "{0}", cjMoxPearlOracle)
	me.Hand.PushTop(mox)
	if err := g.CastSpell(me.ID, mox.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effects.CounterCountOn(g, clock, "hour"); n != 1 {
		t.Errorf("Midnight Clock has %d hour counters, want 1", n)
	}
	if me.Life != mine+1 || opp.Life != theirs-1 {
		t.Errorf("life %d/%d, want %d/%d", me.Life, opp.Life, mine+1, theirs-1)
	}
	// The clock is the real card: it taps for {U}.
	if err := g.ActivateManaAbility(me.ID, clock, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the conjured Midnight Clock: %v", err)
	}
	if me.ManaPool.CountColor("U") != 1 {
		t.Error("the conjured Midnight Clock did not make {U}")
	}
}

// Sanguine Brushstroke: a Blood token and a conjured Blood Artist that
// works; sacrificing the Blood token drains.
func TestSanguineBrushstrokeConjuresAWorkingBloodArtist(t *testing.T) {
	g, me, opp := effectstest.EoTable(t)
	cjEnter(t, g, me.ID, "Sanguine Brushstroke", "Enchantment", cjSanguineOracle)
	if n := len(cjConjured(g.Battlefield, "Blood Artist")); n != 1 {
		t.Fatalf("%d conjured Blood Artists, want 1", n)
	}
	var blood uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Blood" {
			blood = c.InstanceID
		}
	}
	if blood == uuid.Nil {
		t.Fatal("no Blood token")
	}
	effectstest.PushHandCardOfType(me, "Spare", "Sorcery")
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	mine, theirs := me.Life, opp.Life
	if err := g.ActivateCatalogAbility(me.ID, blood, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("crack the Blood token: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine+1 || opp.Life != theirs-1 {
		t.Errorf("life %d/%d after sacrificing Blood, want %d/%d", me.Life, opp.Life, mine+1, theirs-1)
	}

	// The conjured Blood Artist watches deaths.
	bear := effectstest.PushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	mine, theirs = me.Life, opp.Life
	var derr error
	g.WithWriteLock(func() { derr = g.DestroyPermanentForEffect(bear) })
	if derr != nil {
		t.Fatal(derr)
	}
	for i := 0; i < 4 && len(g.PendingChoices) > 0; i++ {
		effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine+1 || opp.Life != theirs-1 {
		t.Errorf("life %d/%d after a creature died, want %d/%d (Blood Artist)", me.Life, opp.Life, mine+1, theirs-1)
	}
}

// Euru, Acorn Scrounger: forage, then the conjured Chitterspitter.
func TestEuruForagesAndConjuresChitterspitter(t *testing.T) {
	g, me, _ := effectstest.EoTable(t)
	gy := effectstest.FgGraveyard(me, 3)
	effectstest.EnterFromHand(t, g, me.ID, "Euru, Acorn Scrounger", "Legendary Creature — Squirrel Soldier", cjEuruOracle)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.EoPay(t, g, me.ID, 0, gy...)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(cjConjured(g.Battlefield, "Chitterspitter")); n != 1 {
		t.Errorf("%d conjured Chitterspitters after foraging, want 1", n)
	}
}

func TestEuruDeclinedForageConjuresNothing(t *testing.T) {
	g, me, _ := effectstest.EoTable(t)
	effectstest.FgGraveyard(me, 3)
	effectstest.EnterFromHand(t, g, me.ID, "Euru, Acorn Scrounger", "Legendary Creature — Squirrel Soldier", cjEuruOracle)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(cjConjured(g.Battlefield, "Chitterspitter")); n != 0 {
		t.Errorf("%d Chitterspitters after declining, want 0", n)
	}
}

// --- the boon ---------------------------------------------------------------

// Tenacious Pup: 1 life and a one-time boon. The first creature spell
// gets the three counters and spends the boon; the second gets nothing.
func TestTenaciousPupBoonIsOneTime(t *testing.T) {
	g, me, _ := effectstest.EoTable(t)
	life := me.Life
	cjEnter(t, g, me.ID, "Tenacious Pup", "Creature — Wolf", cjPupOracle)
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
	if em := g.EmblemsForPlayer(me.ID); len(em) != 1 || em[0].Label != "Tenacious Pup boon" {
		t.Fatalf("emblems %+v, want the Tenacious Pup boon", em)
	}
	cast := func(name string) uuid.UUID {
		c := cjCard(me.ID, name, "Creature — Bear", "{0}", "")
		c.Power, c.Toughness = 2, 2
		me.Hand.PushTop(c)
		if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		return c.InstanceID
	}
	first := cast("First Bear")
	c, ok := effectstest.BattlefieldCard(g, first)
	if !ok {
		t.Fatal("the first creature did not enter")
	}
	for _, kind := range []string{game.CounterPlusOne, game.CounterTrample, game.CounterVigilance} {
		if c.Counters[kind] != 1 {
			t.Errorf("first creature has %d %s counters, want 1", c.Counters[kind], kind)
		}
	}
	if em := g.EmblemsForPlayer(me.ID); len(em) != 0 {
		t.Errorf("the boon is still there after triggering: %+v", em)
	}
	second := cast("Second Bear")
	if c, _ := effectstest.BattlefieldCard(g, second); len(c.Counters) != 0 {
		t.Errorf("second creature has counters %v; the boon is one-time", c.Counters)
	}
}

// --- the conjured cards themselves -----------------------------------------

func TestAncestralRecallTargetPlayerDrawsThree(t *testing.T) {
	g, me, opp := effectstest.EoTable(t)
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	c := cjCard(me.ID, "Ancestral Recall", "Instant", "{U}", cjRecallOracle)
	me.Hand.PushTop(c)
	before := opp.Hand.Size()
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != before+3 {
		t.Errorf("target drew %d, want 3", opp.Hand.Size()-before)
	}
}

func TestTimeWalkTakesAnExtraTurn(t *testing.T) {
	g, me, _ := effectstest.EoTable(t)
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	c := cjCard(me.ID, "Time Walk", "Sorcery", "{1}{U}", cjTimeWalkOracle)
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.ExtraTurns) != 1 || g.Seats[g.ExtraTurns[0].Seat].ID != me.ID {
		t.Errorf("extra turns %+v, want one for me", g.ExtraTurns)
	}
}

func TestTimetwisterEachPlayerDrawsSeven(t *testing.T) {
	g, me, opp := effectstest.EoTable(t)
	effectstest.FgGraveyard(opp, 2)
	for _, col := range []string{"U", "C", "C"} {
		me.ManaPool.AddMana(game.ManaToken{Color: col})
	}
	c := cjCard(me.ID, "Timetwister", "Sorcery", "{2}{U}", cjTwisterOracle)
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 7 || opp.Hand.Size() != 7 || opp.Graveyard.Size() != 0 {
		t.Errorf("hands %d/%d, opponent graveyard %d; want 7/7/0", me.Hand.Size(), opp.Hand.Size(), opp.Graveyard.Size())
	}
	if !me.Graveyard.Contains(c.InstanceID) {
		t.Error("Timetwister did not go to the graveyard")
	}
}

func TestBlackLotusMakesThreeOfOneColor(t *testing.T) {
	g, me, _ := effectstest.EoTable(t)
	lotus := effectstest.PushCatalogPermanent(g, me.ID, "Black Lotus", "Artifact", cjLotusOracle, false)
	if err := g.ActivateManaAbility(me.ID, lotus, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatal(err)
	}
	if me.ManaPool.CountColor("G") != 3 {
		t.Errorf("%d {G}, want 3", me.ManaPool.CountColor("G"))
	}
	if _, ok := effectstest.BattlefieldCard(g, lotus); ok {
		t.Error("Black Lotus was not sacrificed")
	}
}

func TestLethalStingPutsTheCounterThenDestroys(t *testing.T) {
	g, me, opp := effectstest.EoTable(t)
	mine := effectstest.PushTribalCreature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.PushTribalCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	c := cjCard(me.ID, "Lethal Sting", "Sorcery", "{2}{B}", cjLethalOracle)
	me.Hand.PushTop(c)
	for _, col := range []string{"B", "C", "C"} {
		me.ManaPool.AddMana(game.ManaToken{Color: col})
	}
	target := []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Targets: target, Strict: true}); err == nil {
		t.Fatal("Lethal Sting was cast without its -1/-1 counter")
	}
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Targets: target, BlightIDs: []uuid.UUID{mine}}); err != nil {
		t.Fatal(err)
	}
	if got, _ := effectstest.BattlefieldCard(g, mine); got.Counters[game.CounterMinusOne] != 1 {
		t.Errorf("my creature has %d -1/-1 counters, want 1", got.Counters[game.CounterMinusOne])
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, theirs); ok {
		t.Error("the target was not destroyed")
	}
}

func TestDeathOfAThousandStingsDrainsAndReturns(t *testing.T) {
	g, me, opp := effectstest.EoTable(t)
	for _, col := range []string{"B", "C", "C", "C", "C"} {
		me.ManaPool.AddMana(game.ManaToken{Color: col})
	}
	c := cjCard(me.ID, "Death of a Thousand Stings", "Instant — Arcane", "{4}{B}", cjDeathOracle)
	me.Hand.PushTop(c)
	mine, theirs := me.Life, opp.Life
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine+1 || opp.Life != theirs-1 {
		t.Errorf("life %d/%d, want %d/%d", me.Life, opp.Life, mine+1, theirs-1)
	}
	if !me.Graveyard.Contains(c.InstanceID) {
		t.Fatal("not in the graveyard")
	}
	// More cards in hand than each opponent at my next upkeep: it may
	// return. Seven cards in my hand (no discard at cleanup) and none in
	// any other; each opponent draws at most one on the way.
	for _, p := range g.Seats {
		for p.Hand.Size() > 0 {
			_, _ = p.Hand.PopTop()
		}
	}
	for i := 0; i < 7; i++ {
		effectstest.PushHandCardOfType(me, "Spare", "Sorcery")
	}
	seat := g.Turn.ActiveSeat
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToStepOf(t, g, seat, game.StepUpkeep)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestTriggerPromptFor(g, me.ID)
	if p == nil {
		t.Fatal("the upkeep trigger did not ask its \"you may\"")
	}
	if err := g.ResolveTriggerPrompt(p.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(c.InstanceID) {
		t.Error("Death of a Thousand Stings did not return to hand")
	}
}

// --- engine: snapshot and redirect -----------------------------------------

// A conjured card and a perpetual keyword survive a snapshot round trip.
func TestConjuredCardAndPerpetualKeywordSurviveASnapshot(t *testing.T) {
	g, me, _ := effectstest.EoTable(t)
	cjEnter(t, g, me.ID, "Rusko, Clockmaker", "Legendary Creature — Human Artificer", cjRuskoOracle)
	clock := cjConjured(g.Battlefield, "Midnight Clock")[0].InstanceID
	g.WithWriteLock(func() {
		if err := g.GainKeywordPerpetuallyForEffect(clock, "lifelink"); err != nil {
			t.Fatal(err)
		}
	})
	data, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap game.GameSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		t.Fatal(err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatal(err)
	}
	c, ok := effectstest.BattlefieldCard(restored, clock)
	if !ok || !c.Conjured || !slices.Contains(c.PerpetualKeywords, "lifelink") {
		t.Errorf("restored clock %+v lost Conjured or its perpetual lifelink", c)
	}
}
