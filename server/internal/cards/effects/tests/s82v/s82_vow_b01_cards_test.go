package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	vow1MassiveMight   = "4f0ec90b-e5a0-49bd-8fa0-9be6ba259a62"
	vow1WitchsWeb      = "6ec9a14a-0a62-4f8a-9147-3cab01411a65"
	vow1PiercingLight  = "2a6c7826-dcfe-4029-894e-6c82c4dd84b7"
	vow1Sanctify       = "c804ac8e-7427-43c6-b496-3ee6de192f03"
	vow1SporeCrawler   = "fa7c13c0-350b-4ed6-8c16-590d2b988184"
	vow1SporebackWolf  = "6fd5f773-e53e-4a7d-9093-dddd9b2e741f"
	vow1Phalanx        = "94c98a3b-9c3e-497f-b954-8f2248ca01ac"
	vow1PyreSpawn      = "13b0a3f3-04a5-465f-898b-13a28f795a16"
	vow1Daybreak       = "fbb3f959-ded7-4570-bacf-0e1394dbef80"
	vow1ToxicScorpion  = "612ea70d-f4f2-4909-a025-218161ae1059"
	vow1Minister       = "1d789ccd-8192-41b1-a5bb-1e2174657698"
	vow1Officiant      = "41c06726-0360-4488-8dc3-ec257386a372"
	vow1NebelgastBegui = "b1e76ce9-6f39-419c-99e3-ede2dfa4858c"
)

func vow1HasKeyword(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	return effectstest.HasKeywordOf(g, id, kw)
}

func TestVowB01CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		vow1MassiveMight: "Massive Might", vow1WitchsWeb: "Witch's Web", vow1PiercingLight: "Piercing Light",
		vow1Sanctify: "Sanctify", vow1SporeCrawler: "Spore Crawler", vow1SporebackWolf: "Sporeback Wolf",
		vow1Phalanx: "Unhallowed Phalanx", vow1PyreSpawn: "Pyre Spawn", vow1Daybreak: "Daybreak Combatants",
		vow1ToxicScorpion: "Toxic Scorpion", vow1Minister: "Traveling Minister",
		vow1Officiant: "Unholy Officiant", vow1NebelgastBegui: "Nebelgast Beguiler",
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

func TestVowB01MassiveMightPumpsAndGrantsTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.TdCreature(g, me.ID, "Bear", 2, 2)
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Massive Might", TypeLine: "Instant", OracleID: vow1MassiveMight,
		ManaCost: "{G}"}, game.CastSpellParams{Targets: effectstest.TdCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("Bear is %d/%d, want 4/4", p, tough)
	}
	if !vow1HasKeyword(t, g, bear, "trample") {
		t.Error("Bear has no trample")
	}
}

func TestVowB01MassiveMightNeedsACreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.PushTypedCard(g, me.ID, "Rock", "Artifact", "{1}")
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Massive Might", TypeLine: "Instant",
		OracleID: vow1MassiveMight, ManaCost: "{G}", Owner: me.ID, Controller: me.ID})
	id := me.Hand.Cards[me.Hand.Size()-1].InstanceID
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.TdCard(rock)}); err == nil {
		t.Error("Massive Might targeted an artifact")
	}
}

func TestVowB01WitchsWebPumpsGrantsReachAndUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.TdCreature(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Witch's Web", TypeLine: "Instant", OracleID: vow1WitchsWeb,
		ManaCost: "{1}{G}"}, game.CastSpellParams{Targets: effectstest.TdCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("Bear is %d/%d, want 5/5", p, tough)
	}
	if !vow1HasKeyword(t, g, bear, "reach") {
		t.Error("Bear has no reach")
	}
	if effectstest.B16Tapped(t, g, bear) {
		t.Error("Bear is still tapped")
	}
}

func TestVowB01PiercingLightHitsOnlyAttackersAndBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	idle := effectstest.TdCreature(g, opp.ID, "Idle", 2, 2)
	atk := effectstest.TdCreature(g, me.ID, "Attacker", 2, 2)
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Piercing Light", TypeLine: "Instant",
		OracleID: vow1PiercingLight, ManaCost: "{W}", Owner: me.ID, Controller: me.ID})
	id := me.Hand.Cards[me.Hand.Size()-1].InstanceID
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.TdCard(idle)}); err == nil {
		t.Fatal("Piercing Light targeted a creature that is not in combat")
	}
	effectstest.DeclareAttack(t, g, opp.ID, atk)
	effectstest.B06AddMana(me, "W")
	effectstest.SeedLibrary(me, "Top", "Next")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.TdCard(atk)}); err != nil {
		t.Fatalf("cast on attacker: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Fatal("no scry prompt")
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(atk) {
		t.Error("the 2/2 attacker survived 2 damage")
	}
}

func TestVowB01SanctifyDestroysAndGainsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ench := effectstest.B12Push(g, opp.ID, "Their Enchantment", "Enchantment", "", 0, 0)
	life := me.Life
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Sanctify", TypeLine: "Sorcery", OracleID: vow1Sanctify,
		ManaCost: "{1}{W}"}, game.CastSpellParams{Targets: effectstest.TdCard(ench)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment survived")
	}
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3", life, me.Life)
	}
}

func TestVowB01SporeCrawlerDrawsWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	crawler := effectstest.B12Push(g, me.ID, "Spore Crawler", "Creature — Fungus", vow1SporeCrawler, 3, 2)
	effectstest.SeedLibrary(me, "A", "B")
	hand := me.Hand.Size()
	effectstest.B25Destroy(g, crawler)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, got)
	}
}

func TestVowB01SporebackWolfGetsToughnessOnlyOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wolf := effectstest.B12Push(g, me.ID, "Sporeback Wolf", "Creature — Wolf", vow1SporebackWolf, 2, 2)
	if p, tough := effectstest.EoeCDPT(t, g, wolf); p != 2 || tough != 4 {
		t.Errorf("on my turn %d/%d, want 2/4", p, tough)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, tough := effectstest.EoeCDPT(t, g, wolf); p != 2 || tough != 2 {
		t.Errorf("on their turn %d/%d, want 2/2", p, tough)
	}
}

func TestVowB01UnhallowedPhalanxEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.TdCastFromHand(t, g, game.Card{Name: "Unhallowed Phalanx", TypeLine: "Creature — Zombie Soldier",
		OracleID: vow1Phalanx, ManaCost: "{4}{B}", Power: 1, Toughness: 13}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Phalanx did not resolve")
	}
	if !effectstest.B16Tapped(t, g, id) {
		t.Error("Phalanx entered untapped")
	}
}

func TestVowB01PyreSpawnDealsThreeToAnyTargetWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	spawn := effectstest.B12Push(g, me.ID, "Pyre Spawn", "Creature — Elemental", vow1PyreSpawn, 6, 4)
	life := opp.Life
	effectstest.B25Destroy(g, spawn)
	effectstest.Vow1Pick(t, g, me, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	if opp.Life != life-3 {
		t.Errorf("life %d -> %d, want -3", life, opp.Life)
	}
}

func TestVowB01DaybreakCombatantsPumpsATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.TdCreature(g, me.ID, "Bear", 2, 2)
	id := effectstest.TdCastFromHand(t, g, game.Card{Name: "Daybreak Combatants", TypeLine: "Creature — Human Warrior",
		OracleID: vow1Daybreak, ManaCost: "{2}{R}", Power: 2, Toughness: 2}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Vow1Pick(t, g, me, effectstest.TdCard(bear))
	if p, _ := effectstest.PtOf(t, g, bear); p != 4 {
		t.Errorf("Bear power %d, want 4", p)
	}
	if p, _ := effectstest.PtOf(t, g, id); p != 2 {
		t.Errorf("Combatants power %d, want 2", p)
	}
	if !vow1HasKeyword(t, g, id, "haste") {
		t.Error("Combatants has no haste")
	}
}

func TestVowB01ToxicScorpionGivesAnotherCreatureDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.TdCreature(g, me.ID, "Bear", 2, 2)
	id := effectstest.TdCastFromHand(t, g, game.Card{Name: "Toxic Scorpion", TypeLine: "Creature — Scorpion",
		OracleID: vow1ToxicScorpion, ManaCost: "{1}{G}", Power: 1, Toughness: 1}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Vow1Pick(t, g, me, effectstest.TdCard(bear))
	if !vow1HasKeyword(t, g, bear, "deathtouch") {
		t.Error("Bear has no deathtouch")
	}
	if !vow1HasKeyword(t, g, id, "deathtouch") {
		t.Error("Scorpion lost its printed deathtouch")
	}
}

func TestVowB01TravelingMinisterPumpsAndGainsLifeAtSorcerySpeedOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	minister := effectstest.BlbB1Push(g, me.ID, "Traveling Minister", "Creature — Human Cleric", vow1Minister, 1, 1)
	bear := effectstest.TdCreature(g, opp.ID, "Bear", 2, 2)
	life := me.Life
	effectstest.B16Activate(t, g, me.ID, minister, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	if p, _ := effectstest.PtOf(t, g, bear); p != 3 {
		t.Errorf("Bear power %d, want 3", p)
	}
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1", life, me.Life)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g2)
	m2 := effectstest.BlbB1Push(g2, me2.ID, "Traveling Minister", "Creature — Human Cleric", vow1Minister, 1, 1)
	b2 := effectstest.TdCreature(g2, me2.ID, "Bear", 2, 2)
	effectstest.AdvanceTo(t, g2, game.StepDeclareAttackers)
	err := g2.ActivateCatalogAbility(me2.ID, m2, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(b2)})
	if err == nil {
		t.Error("Minister activated outside a main phase")
	}
}

func TestVowB01UnholyOfficiantGrowsForFiveMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	off := effectstest.BlbB1Push(g, me.ID, "Unholy Officiant", "Creature — Vampire Cleric", vow1Officiant, 1, 2)
	if !vow1HasKeyword(t, g, off, "vigilance") {
		t.Error("no vigilance")
	}
	effectstest.B06AddMana(me, "W", "W", "W", "W", "W")
	effectstest.B16Activate(t, g, me.ID, off, 0, game.ActivateAbilityParams{})
	if got := effectstest.CountersOn(g, off, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if p, tough := effectstest.PtOf(t, g, off); p != 2 || tough != 3 {
		t.Errorf("Officiant is %d/%d, want 2/3", p, tough)
	}
}

func TestVowB01NebelgastBeguilerTapsATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bg := effectstest.OdyB3Source(t, g, "Nebelgast Beguiler", "Creature — Spirit", vow1NebelgastBegui, "W")
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B16Activate(t, g, me.ID, bg, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	if !effectstest.B16Tapped(t, g, bear) {
		t.Error("Bear was not tapped")
	}
	if !effectstest.B16Tapped(t, g, bg) {
		t.Error("Beguiler did not tap for its cost")
	}
}
