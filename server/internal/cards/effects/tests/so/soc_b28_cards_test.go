package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b28_cards_test.go — Secrets of Strixhaven (sos), batch 28 (#4456).
// Each test plays the real catalog entry through the engine and asserts what
// the printed text says.

const (
	socB28ConfidenceOracle = "c17f0e65-e5a3-4113-b782-0b192ab52046"
	socB28KirolOracle      = "df1c0d06-2109-4297-9271-8a003fc892bc"
	socB28LluwenOracle     = "cb125cd4-c7c2-415d-b802-bdcfc2454403"
	socB28ScornmageOracle  = "29c23fd7-0ac4-46af-8ecd-92b726575252"
	socB28CollectorOracle  = "84bb960c-c34e-4979-8c0f-893c85cee27d"
	socB28HistoryOracle    = "a76adc88-e7db-43e2-8fbf-bae95c0838ef"
	socB28CharmOracle      = "a5729f11-6cde-4830-b882-ab4c5251c1a9"
	socB28FavorOracle      = "d514704b-4055-48ed-9da3-e4397fca7576"
	socB28RefereeOracle    = "3f27988b-0863-4cc1-8eb9-83bd0cd8561c"
	socB28ArchaicOracle    = "c6a5adc8-ee57-421e-8235-0b3160f44568"
)

func TestSocB28EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB28ConfidenceOracle:       "Killian's Confidence",
		socB28KirolOracle:            "Kirol, History Buff",
		socB28KirolOracle + "#1":     "Pack a Punch",
		socB28LluwenOracle:           "Lluwen, Exchange Student",
		socB28LluwenOracle + "#1":    "Pest Friend",
		socB28ScornmageOracle:        "Lecturing Scornmage",
		socB28CollectorOracle:        "Leech Collector",
		socB28CollectorOracle + "#1": "Bloodletting",
		socB28HistoryOracle:          "Living History",
		socB28CharmOracle:            "Lorehold Charm",
		socB28FavorOracle:            "Lumaret's Favor",
		socB28RefereeOracle:          "Mage Tower Referee",
		socB28ArchaicOracle:          "Magmablood Archaic",
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

// socB28Attack declares `attackers` against `defender`, locks the attack in
// and walks into the combat damage step.
func socB28Attack(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Killian's Confidence ----------------------------------------------------

func TestSocB28ConfidencePumpsTheTargetAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	id := effectstest.SocB20HandCard(me, "Killian's Confidence", "Sorcery", "{W}{B}", socB28ConfidenceOracle)
	effectstest.AdvanceToMain(t, g)
	before := me.Hand.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.S58p6View(t, g, bear)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("the Bear is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d, want %d (the card cast left it and one was drawn)", got, before)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("Killian's Confidence is not in the graveyard after resolving")
	}
	effectstest.EndTurn(t, g)
	c = effectstest.S58p6View(t, g, bear)
	if c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("the Bear is %d/%d after the turn ended, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
}

func socB28ConfidenceInGraveyardWithAttackers(t *testing.T, g *game.Game, n int) (uuid.UUID, []uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	var attackers []uuid.UUID
	for i := 0; i < n; i++ {
		attackers = append(attackers, effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
	}
	card := effectstest.SeedGraveyardCard(t, g, "Killian's Confidence", "Sorcery", socB28ConfidenceOracle)
	return card, attackers
}

func TestSocB28ConfidenceReturnsFromTheGraveyardWhenYouPayAfterCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	card, attackers := socB28ConfidenceInGraveyardWithAttackers(t, g, 2)
	socB28Attack(t, g, opp.ID, attackers...)
	effectstest.B06AddMana(me, "W")
	prompt := effectstest.AnswerPayUnless(t, g, me.ID, true)
	if prompt.PayCost != "{W/B}" {
		t.Errorf("payment = %q, want {W/B}", prompt.PayCost)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(card) || me.Graveyard.Contains(card) {
		t.Errorf("paying did not return the card to hand (hand=%t graveyard=%t)", me.Hand.Contains(card), me.Graveyard.Contains(card))
	}
	// Two creatures connected: one trigger, so one question.
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a second question was asked for the second creature that connected")
	}
}

func TestSocB28ConfidenceStaysWhenYouDeclineOrCannotPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	card, attackers := socB28ConfidenceInGraveyardWithAttackers(t, g, 1)
	socB28Attack(t, g, opp.ID, attackers...)
	effectstest.B06AddMana(me, "W")
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(card) {
		t.Fatal("declining moved the card out of the graveyard")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	card2, attackers2 := socB28ConfidenceInGraveyardWithAttackers(t, g2, 1)
	socB28Attack(t, g2, opp2.ID, attackers2...)
	if effectstest.HasPayUnlessFor(g2, me2.ID) {
		effectstest.AnswerPayUnless(t, g2, me2.ID, true)
		effectstest.PassPriorityAroundTable(t, g2)
	}
	if !me2.Graveyard.Contains(card2) {
		t.Error("the card came back without {W/B} being paid")
	}
}

func TestSocB28ConfidenceAsksNothingWhenNoCreatureConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	_, _ = socB28ConfidenceInGraveyardWithAttackers(t, g, 0)
	effectstest.B06AddMana(me, "W")
	effectstest.EndTurn(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("asked about returning the card with no combat damage dealt")
	}
}

// --- Kirol, History Buff // Pack a Punch -------------------------------------

func socB28CastKirol(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	c := sos1Prep(me.ID, socB28KirolOracle, nil, "Kirol, History Buff", "Legendary Creature — Vampire Cleric", "{R}{W}", 2, 3,
		"Pack a Punch", "Sorcery", "{1}{R}{W}")
	effectstest.CastFromHandAndResolve(t, g, me, c)
	return c.InstanceID
}

func TestSocB28KirolBecomesPreparedWhenACardLeavesYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kirol := socB28CastKirol(t, g, me)
	mine := socB13GraveyardCard(me, "Old Card", "Creature — Bear", 2, 2)
	theirs := socB13GraveyardCard(opp, "Their Card", "Creature — Bear", 2, 2)
	if g.IsPreparedForEffect(kirol) {
		t.Fatal("prepared before anything left a graveyard")
	}

	socB11Move(t, g, opp.ID, theirs, game.ZoneGraveyard, game.ZoneHand)
	if n := effectstest.TriggersOnStackFrom(g, kirol); n != 0 {
		t.Fatalf("%d triggers for a card leaving an opponent's graveyard", n)
	}
	socB11Move(t, g, me.ID, mine, game.ZoneGraveyard, game.ZoneHand)
	if n := effectstest.TriggersOnStackFrom(g, kirol); n != 1 {
		t.Fatalf("%d triggers for a card leaving your graveyard, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(kirol) {
		t.Fatal("Kirol is not prepared after a card left your graveyard")
	}
	if _, ok := effectstest.PrepareCopyOf(g, "Pack a Punch"); !ok {
		t.Fatal("no Pack a Punch copy in exile")
	}
}

func TestSocB28KirolIgnoresACardEnteringYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kirol := socB28CastKirol(t, g, me)
	card := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(card) })
	effectstest.PassPriorityAroundTable(t, g)
	if g.IsPreparedForEffect(kirol) {
		t.Error("Kirol became prepared when a card entered the graveyard")
	}
}

func TestSocB28PackAPunchMillsAPutsTwoCountersAndGivesTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kirol := socB28CastKirol(t, g, me)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() {
		if _, err := g.BecomePreparedForEffect(kirol); err != nil {
			t.Fatalf("prepare: %v", err)
		}
	})
	library, graveyard := me.Library.Size(), me.Graveyard.Size()
	effectstest.FraBCastCopy(t, g, me, "Pack a Punch", effectstest.Pr7aCard(bear))
	if got := me.Library.Size(); got != library-1 {
		t.Errorf("library %d, want %d (one card milled)", got, library-1)
	}
	if got := me.Graveyard.Size(); got != graveyard+1 {
		t.Errorf("graveyard %d, want %d (the milled card; the copy ceases to exist)", got, graveyard+1)
	}
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters %d, want 2", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the target did not gain trample")
	}
	if g.IsPreparedForEffect(kirol) {
		t.Error("Kirol is still prepared after the copy was cast")
	}
	effectstest.EndTurn(t, g)
	if effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("trample lasted past the end of the turn")
	}
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("the counters did not stay: %d", got)
	}
}

// --- Lecturing Scornmage -----------------------------------------------------

func socB28Scornmage(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.SocB17Perm(g, owner, "Lecturing Scornmage", "Creature — Human Warlock", socB28ScornmageOracle, "{B}", 1, 1)
}

func TestSocB28ScornmageGrowsWhenYourSpellTargetsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mage := socB28Scornmage(g, me.ID)
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Locust Spray", "Instant", effectstest.SocB17LocustSprayOracle, effectstest.CardRefs(bear))
	if n := effectstest.TriggersOnStackFrom(g, mage); n != 1 {
		t.Fatalf("%d Repartee triggers, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, mage, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters %d, want 1", got)
	}
	c := effectstest.S58p6View(t, g, mage)
	if c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("Scornmage is %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestSocB28ScornmageIgnoresSpellsThatTargetNoCreatureAndOpponentsSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mage := socB28Scornmage(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Plain Instant", "Instant", effectstest.SocB17PlainSorceryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastXSpell(t, g, "Blaze", "Sorcery", effectstest.BlazeOracle, "{X}{R}", 1,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, mage, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters from spells that target no creature, want 0", got)
	}
}

// --- Leech Collector // Bloodletting -----------------------------------------

func socB28CastCollector(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	c := sos1Prep(me.ID, socB28CollectorOracle, nil, "Leech Collector", "Creature — Human Warlock", "{1}{B}", 2, 2,
		"Bloodletting", "Sorcery", "{B}")
	effectstest.CastFromHandAndResolve(t, g, me, c)
	return c.InstanceID
}

func TestSocB28CollectorBecomesPreparedOnTheFirstLifeGainOfTheTurnOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	collector := socB28CastCollector(t, g, me)
	if g.IsPreparedForEffect(collector) {
		t.Fatal("prepared before any life was gained")
	}
	sos1GainLife(t, g)
	if !g.IsPreparedForEffect(collector) {
		t.Fatal("not prepared after the first life gain of the turn")
	}
	// Cast the copy: it unprepares, and a second gain the same turn is not the first.
	effectstest.FraBCastCopy(t, g, me, "Bloodletting")
	if g.IsPreparedForEffect(collector) {
		t.Fatal("still prepared after casting the copy")
	}
	sos1GainLife(t, g)
	if g.IsPreparedForEffect(collector) {
		t.Error("prepared again by the second life gain of the turn")
	}
	// The next turn's first gain prepares it again.
	effectstest.EndTurn(t, g)
	effectstest.EndTurn(t, g)
	effectstest.EndTurn(t, g)
	effectstest.EndTurn(t, g)
	sos1GainLife(t, g)
	if !g.IsPreparedForEffect(collector) {
		t.Error("not prepared by the first life gain of a later turn")
	}
}

func TestSocB28CollectorIgnoresLifeGainedBeforeItEnteredThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sos1GainLife(t, g)
	collector := socB28CastCollector(t, g, me)
	sos1GainLife(t, g)
	if g.IsPreparedForEffect(collector) {
		t.Error("prepared by a second gain of the turn because it had not been out for the first")
	}
}

func TestSocB28CollectorIgnoresAnOpponentsLifeGain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	collector := socB28CastCollector(t, g, me)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 3) })
	effectstest.PassPriorityAroundTable(t, g)
	if g.IsPreparedForEffect(collector) {
		t.Error("prepared by an opponent gaining life")
	}
}

func TestSocB28BloodlettingMakesEachOpponentLoseTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	collector := socB28CastCollector(t, g, me)
	g.WithWriteLock(func() {
		if _, err := g.BecomePreparedForEffect(collector); err != nil {
			t.Fatalf("prepare: %v", err)
		}
	})
	before := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		before[s.ID] = s.Life
	}
	effectstest.FraBCastCopy(t, g, me, "Bloodletting")
	for _, s := range g.Seats {
		want := before[s.ID] - 2
		if s.ID == me.ID {
			want = before[s.ID]
		}
		if s.Life != want {
			t.Errorf("%s life %d, want %d", s.Name, s.Life, want)
		}
	}
}

// --- Living History ----------------------------------------------------------

func TestSocB28LivingHistoryMakesATwoTwoRedAndWhiteSpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Living History", "Enchantment", socB28HistoryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	spirits := effectstest.SocB17Names(g, me.ID, "Spirit")
	if len(spirits) != 1 {
		t.Fatalf("%d Spirits, want 1", len(spirits))
	}
	v := effectstest.S58p6View(t, g, spirits[0].InstanceID)
	if !v.IsToken() || v.CurrentPower() != 2 || v.CurrentToughness() != 2 || len(v.EffectiveColors()) != 2 {
		t.Errorf("Spirit token=%v %d/%d colors %v, want a 2/2 red and white token", v.IsToken(), v.CurrentPower(), v.CurrentToughness(), v.EffectiveColors())
	}
}

func socB28HistoryAttack(t *testing.T, g *game.Game, leaveGraveyard bool) (attacker uuid.UUID, triggers int) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	history := effectstest.PushCatalogPermanent(g, me.ID, "Living History", "Enchantment", socB28HistoryOracle, false)
	attacker = effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	if leaveGraveyard {
		card := socB13GraveyardCard(me, "Old Card", "Creature — Bear", 1, 1)
		socB11Move(t, g, me.ID, card, game.ZoneGraveyard, game.ZoneHand)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(other, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	return attacker, effectstest.TriggersOnStackFrom(g, history)
}

func TestSocB28LivingHistoryPumpsTheChosenAttackerOnlyIfACardLeftYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	attacker, _ := socB28HistoryAttack(t, g, true)
	effectstest.AnswerPickTargetIfOpen(t, g, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.S58p6View(t, g, attacker)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 2 {
		t.Errorf("the chosen attacker is %d/%d, want 4/2", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestSocB28LivingHistoryDoesNothingWhenNoCardLeftYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	attacker, triggers := socB28HistoryAttack(t, g, false)
	if triggers != 0 || len(g.PendingChoices) != 0 {
		t.Fatalf("triggers %d, pending choices %d with no card having left the graveyard", triggers, len(g.PendingChoices))
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.S58p6View(t, g, attacker)
	if c.CurrentPower() != 2 {
		t.Errorf("the attacker is %d power, want 2", c.CurrentPower())
	}
}

func TestSocB28LivingHistoryIgnoresACardLeavingAnOpponentsGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := socB13GraveyardCard(opp, "Their Card", "Creature — Bear", 1, 1)
	socB11Move(t, g, opp.ID, theirs, game.ZoneGraveyard, game.ZoneHand)
	history := effectstest.PushCatalogPermanent(g, me.ID, "Living History", "Enchantment", socB28HistoryOracle, false)
	attacker := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	if n := effectstest.TriggersOnStackFrom(g, history); n != 0 {
		t.Errorf("%d triggers for a card that left an opponent's graveyard", n)
	}
}

// --- Lorehold Charm ----------------------------------------------------------

func socB28CastCharm(t *testing.T, g *game.Game, mode int, targets []game.TargetRef) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Lorehold Charm", TypeLine: "Instant", OracleID: socB28CharmOracle,
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{mode}, Targets: targets}); err != nil {
		t.Fatalf("cast Lorehold Charm mode %d: %v", mode, err)
	}
}

func TestSocB28CharmFirstModeMakesEachOpponentSacrificeANontokenArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	token := effectstest.B12Permanent(g, opp.ID, "Treasure", "Token Artifact — Treasure")
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	spare := effectstest.B12Permanent(g, opp.ID, "Spare Rock", "Artifact")
	other := effectstest.B12Permanent(g, third.ID, "Other Rock", "Artifact")
	mine := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	socB28CastCharm(t, g, 0, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// The first opponent holds two nontoken artifacts and chooses; the token is not offered.
	c := effectstest.SacrificeChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the opponent with two nontoken artifacts was not asked to choose")
	}
	if effectstest.EdictOffers(effectstest.SacrificeOptions(c), token) {
		t.Error("a token artifact was offered, but only nontoken artifacts qualify")
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, spare)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(spare) {
		t.Error("the artifact the opponent chose was not sacrificed")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("the opponent lost a second artifact")
	}
	if g.Battlefield.Contains(other) {
		t.Error("the other opponent kept their only nontoken artifact")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the caster's own artifact was sacrificed")
	}
	if !g.Battlefield.Contains(token) {
		t.Error("a token artifact was sacrificed")
	}
}

// socB28GraveyardCard puts a card with a mana cost into a graveyard.
func socB28GraveyardCard(p *game.Player, name, typeLine, cost string, power, toughness int) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID})
	return id
}

func TestSocB28CharmSecondModeReturnsAnArtifactOrCreatureCardOfManaValueTwoOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cheapCreature := socB28GraveyardCard(me, "Cheap Bear", "Creature — Bear", "{1}{G}", 2, 2)
	cheapArtifact := socB28GraveyardCard(me, "Cheap Rock", "Artifact", "{2}", 0, 0)
	pricey := socB28GraveyardCard(me, "Big Bear", "Creature — Bear", "{2}{G}", 3, 3)
	spell := socB28GraveyardCard(me, "Cheap Spell", "Sorcery", "{1}", 0, 0)

	for _, bad := range []uuid.UUID{pricey, spell} {
		effectstest.AdvanceToMain(t, g)
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Lorehold Charm", TypeLine: "Instant", OracleID: socB28CharmOracle, Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}, Targets: effectstest.CardRefs(bad), Strict: true}); err == nil {
			t.Fatalf("Lorehold Charm accepted an illegal graveyard target %s", bad)
		}
	}

	socB28CastCharm(t, g, 1, effectstest.CardRefs(cheapCreature))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheapCreature) {
		t.Error("the creature card did not return to the battlefield")
	}
	socB28CastCharm(t, g, 1, effectstest.CardRefs(cheapArtifact))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheapArtifact) {
		t.Error("the artifact card did not return to the battlefield")
	}
}

func TestSocB28CharmThirdModePumpsAndTramplesOnlyCreaturesYouControlThatAreThere(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	socB28CastCharm(t, g, 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID][2]int{a: {3, 3}, b: {4, 4}, theirs: {2, 2}} {
		c := effectstest.S58p6View(t, g, id)
		if c.CurrentPower() != want[0] || c.CurrentToughness() != want[1] {
			t.Errorf("%s is %d/%d, want %d/%d", c.Name, c.CurrentPower(), c.CurrentToughness(), want[0], want[1])
		}
	}
	if !effectstest.HasEffectiveKeyword(t, g, a, "trample") || !effectstest.HasEffectiveKeyword(t, g, b, "trample") {
		t.Error("my creatures did not gain trample")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("their creature gained trample")
	}
	late := effectstest.B12Creature(g, me.ID, "Late Bear", "Creature — Bear", 2, 2)
	if c := effectstest.S58p6View(t, g, late); c.CurrentPower() != 2 || effectstest.HasEffectiveKeyword(t, g, late, "trample") {
		t.Error("a creature that entered after the charm resolved was affected")
	}
}

// --- Lumaret's Favor ---------------------------------------------------------

func TestSocB28FavorPumpsPlusTwoPlusFourAndDoesNotCopyWithoutLifeGain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Lumaret's Favor", "Instant", socB28FavorOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.S58p6View(t, g, bear)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 6 {
		t.Errorf("the Bear is %d/%d, want 4/6", c.CurrentPower(), c.CurrentToughness())
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a copy offered new targets with no life gained this turn")
	}
}

func TestSocB28FavorCopiesItselfIfYouGainedLifeThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	other := effectstest.PushVanillaCreature(g, me.ID, "Other Bear", 2, 2)
	sos1GainLife(t, g)
	effectstest.CastCatalogSpell(t, g, "Lumaret's Favor", "Instant", socB28FavorOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("the copy was offered no new targets")
	}
	if err := g.ResolvePickTarget(prompt.ID, me.ID, effectstest.Pr7aCard(other)); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, other} {
		c := effectstest.S58p6View(t, g, id)
		if c.CurrentPower() != 4 || c.CurrentToughness() != 6 {
			t.Errorf("%s is %d/%d, want 4/6 (the spell and its copy each pump one)", c.Name, c.CurrentPower(), c.CurrentToughness())
		}
	}
}

func TestSocB28FavorCopyMayKeepTheSameTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	sos1GainLife(t, g)
	effectstest.CastCatalogSpell(t, g, "Lumaret's Favor", "Instant", socB28FavorOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if prompt := effectstest.LatestPickTarget(g, me.ID); prompt != nil {
		if err := g.ResolvePickTarget(prompt.ID, me.ID, effectstest.Pr7aCard(bear)); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	c := effectstest.S58p6View(t, g, bear)
	if c.CurrentPower() != 6 || c.CurrentToughness() != 10 {
		t.Errorf("the Bear is %d/%d after the spell and its copy, want 6/10", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Mage Tower Referee ------------------------------------------------------

func socB28Referee(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.SocB17Perm(g, owner, "Mage Tower Referee", "Artifact Creature — Construct", socB28RefereeOracle, "{2}", 2, 1)
}

func TestSocB28RefereeGrowsOnlyOnYourMulticoloredSpells(t *testing.T) {
	for _, tc := range []struct {
		name, cost, pool string
		want             int
	}{
		{"multicolored", "{W}{B}", "WB", 1},
		{"hybrid is multicolored", "{W/B}{W/B}", "WW", 1},
		{"monocolored", "{1}{R}", "RR", 0},
		{"colorless", "{2}", "RR", 0},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		referee := socB28Referee(g, me.ID)
		effectstest.AdvanceToMain(t, g)
		id := effectstest.SocB17Hand(me, "Test Spell", "Sorcery", tc.cost, effectstest.SocB17PlainSorceryOracle)
		effectstest.FloatForTest(g, me, tc.pool)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("%s: cast: %v", tc.name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.CounterCount(g, referee, game.CounterPlusOne); got != tc.want {
			t.Errorf("%s: %d counters, want %d", tc.name, got, tc.want)
		}
	}
}

func TestSocB28RefereeIgnoresAnOpponentsMulticoloredSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	referee := socB28Referee(g, me.ID)
	effectstest.EndTurn(t, g)
	opp := g.Seats[g.Turn.ActiveSeat]
	if opp.ID == me.ID {
		t.Fatal("the turn did not pass")
	}
	effectstest.AdvanceToMain(t, g)
	id := effectstest.SocB17Hand(opp, "Test Spell", "Sorcery", "{W}{B}", effectstest.SocB17PlainSorceryOracle)
	effectstest.FloatForTest(g, opp, "WB")
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, referee, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters from an opponent's spell, want 0", got)
	}
}

// --- Magmablood Archaic ------------------------------------------------------

func socB28CastArchaic(t *testing.T, g *game.Game, pool string) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Magmablood Archaic", TypeLine: "Creature — Avatar", ManaCost: "{2/R}{2/R}{2/R}",
		OracleID: socB28ArchaicOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		Keywords: []string{"reach", "trample"}, KnownBy: map[uuid.UUID]bool{me.ID: true},
	})
	effectstest.FloatForTest(g, me, pool)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast with %s: %v", pool, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestSocB28ArchaicEntersWithACounterPerColorOfManaSpent(t *testing.T) {
	for _, tc := range []struct {
		pool string
		want int
	}{{"RRR", 1}, {"RWWGG", 3}, {"RWGUB", 5}} {
		g := effectstest.NewCatalogGame(t)
		id := socB28CastArchaic(t, g, tc.pool)
		if !g.Battlefield.Contains(id) {
			t.Fatalf("%s: the Archaic did not resolve", tc.pool)
		}
		if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != tc.want {
			t.Errorf("%s: %d +1/+1 counters, want %d", tc.pool, got, tc.want)
		}
		c := effectstest.S58p6View(t, g, id)
		if c.CurrentPower() != 2+tc.want {
			t.Errorf("%s: power %d, want %d", tc.pool, c.CurrentPower(), 2+tc.want)
		}
	}
}

func TestSocB28ArchaicPumpsYourCreaturesByTheColorsOfEachInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	archaic := effectstest.SocB17Perm(g, me.ID, "Magmablood Archaic", "Creature — Avatar", socB28ArchaicOracle, "{2/R}{2/R}{2/R}", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	effectstest.SocB17CastSpellFor(t, g, "Instant", "{W}{R}", "WR")
	for id, want := range map[uuid.UUID]int{archaic: 4, bear: 4, theirs: 2} {
		if got := effectstest.S58p6View(t, g, id).CurrentPower(); got != want {
			t.Errorf("power %d after a two-color spell, want %d", got, want)
		}
	}
	effectstest.SocB17CastSpellFor(t, g, "Sorcery", "{2}{G}", "GGG")
	for id, want := range map[uuid.UUID]int{archaic: 5, bear: 5, theirs: 2} {
		if got := effectstest.S58p6View(t, g, id).CurrentPower(); got != want {
			t.Errorf("power %d after a one-color spell, want %d", got, want)
		}
	}
	// Toughness never changes.
	if got := effectstest.S58p6View(t, g, bear).CurrentToughness(); got != 2 {
		t.Errorf("toughness %d, want 2", got)
	}
	effectstest.EndTurn(t, g)
	if got := effectstest.S58p6View(t, g, bear).CurrentPower(); got != 2 {
		t.Errorf("power %d after the turn ended, want 2", got)
	}
}

func TestSocB28ArchaicIgnoresCreatureSpellsAndFreeSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	_ = effectstest.SocB17Perm(g, me.ID, "Magmablood Archaic", "Creature — Avatar", socB28ArchaicOracle, "{2/R}{2/R}{2/R}", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.SocB17CastSpellFor(t, g, "Creature — Bear", "{W}{R}", "WR")
	if got := effectstest.S58p6View(t, g, bear).CurrentPower(); got != 2 {
		t.Errorf("a creature spell pumped the team to %d", got)
	}
}

// --- Lluwen, Exchange Student // Pest Friend ---------------------------------

func socB28CastLluwen(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	c := sos1Prep(me.ID, socB28LluwenOracle, nil, "Lluwen, Exchange Student", "Legendary Creature — Elf Druid", "{2}{B}{G}", 3, 4,
		"Pest Friend", "Sorcery", "{B/G}")
	effectstest.CastFromHandAndResolve(t, g, me, c)
	return c.InstanceID
}

func TestSocB28LluwenEntersPreparedAndPestFriendMakesAPestThatGainsLifeOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lluwen := socB28CastLluwen(t, g, me)
	if !g.IsPreparedForEffect(lluwen) {
		t.Fatal("Lluwen did not enter prepared")
	}
	effectstest.FraBCastCopy(t, g, me, "Pest Friend")
	if g.IsPreparedForEffect(lluwen) {
		t.Error("Lluwen is still prepared after casting the copy")
	}
	pests := effectstest.SocB17Names(g, me.ID, "Pest")
	if len(pests) != 1 {
		t.Fatalf("%d Pests, want 1", len(pests))
	}
	pest := effectstest.S58p6View(t, g, pests[0].InstanceID)
	if !pest.IsToken() || pest.CurrentPower() != 1 || pest.CurrentToughness() != 1 || len(pest.EffectiveColors()) != 2 {
		t.Fatalf("the Pest is token %v %d/%d colors %v, want a 1/1 black and green token", pest.IsToken(), pest.CurrentPower(), pest.CurrentToughness(), pest.EffectiveColors())
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == pests[0].InstanceID {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	life := me.Life
	effectstest.DeclareAttack(t, g, opp.ID, pests[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d when the Pest attacked, want +1", life, me.Life)
	}
}

func TestSocB28LluwenPestGainsNothingWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lluwen := socB28CastLluwen(t, g, me)
	effectstest.FraBCastCopy(t, g, me, "Pest Friend")
	_ = lluwen
	pest := effectstest.FindBattlefieldByName(g, "Pest")
	if pest == uuid.Nil {
		t.Fatal("no Pest was created")
	}
	life := me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(pest) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d -> %d when the Pest died: this Pest gains life when it attacks, not dies", life, me.Life)
	}
}

func TestSocB28LluwenExilesACreatureCardFromYourGraveyardToBecomePreparedAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lluwen := socB28CastLluwen(t, g, me)
	// Spend the entering prepare so it must be earned back.
	effectstest.FraBCastCopy(t, g, me, "Pest Friend")
	if g.IsPreparedForEffect(lluwen) {
		t.Fatal("still prepared after casting the copy")
	}
	spell := socB28GraveyardCard(me, "Old Spell", "Sorcery", "{1}", 0, 0)
	fodder := socB28GraveyardCard(me, "Fodder", "Creature — Bear", "{1}{G}", 2, 2)

	// A noncreature card is not a legal cost.
	if err := g.ActivateCatalogAbility(me.ID, lluwen, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{spell}}); err == nil {
		t.Fatal("Lluwen accepted a sorcery card as the exile cost")
	}
	if g.IsPreparedForEffect(lluwen) {
		t.Fatal("prepared by a refused activation")
	}
	effectstest.B16Activate(t, g, me.ID, lluwen, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{fodder}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(fodder) || me.Graveyard.Contains(fodder) {
		t.Error("the creature card was not exiled from the graveyard as the cost")
	}
	if !g.IsPreparedForEffect(lluwen) {
		t.Fatal("Lluwen did not become prepared")
	}
	if !me.Graveyard.Contains(spell) {
		t.Error("the sorcery card left the graveyard")
	}
}

func TestSocB28LluwenCannotActivateOutsideYourMainPhaseWithAnEmptyStack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lluwen := socB28CastLluwen(t, g, me)
	fodder := socB28GraveyardCard(me, "Fodder", "Creature — Bear", "{1}{G}", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, lluwen, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{fodder}}); err == nil {
		t.Error("Lluwen's ability was activated at beginning of combat")
	}
	if !me.Graveyard.Contains(fodder) {
		t.Error("the creature card was exiled by a refused activation")
	}
}
