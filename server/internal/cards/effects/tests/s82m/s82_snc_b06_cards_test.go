package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	snc6BoonOfSafety     = "34f0beae-a8e2-4474-9f7f-4735d3f26863"
	snc6SwoopingProtect  = "094689cc-617a-426e-beb9-3f8de692cd5b"
	snc6BrokersVeteran   = "feee1926-0b63-404f-9b01-df965de26b5f"
	snc6WingshieldAgent  = "513034fc-873a-4877-afb2-39ece1ea6a7c"
	snc6VoiceOfTheVermin = "73829ae3-c8de-4e6d-9607-4b2e980ff144"
	snc6DisciplinedDuel  = "c650f497-92d6-48cf-8b59-5d61eb184dc0"
	snc6RhoxPummeler     = "898c8415-8769-48c4-ad16-a4ba8c3d5d7d"
	snc6DapperShieldmate = "3cdd795b-13b8-4812-89ee-d248368f915b"
	snc6Rigo             = "1f42a809-bbfc-4225-9c31-f7d0fa8fa17a"
	snc6SanctuaryWarden  = "713bcfb6-e5ea-411e-b4b3-52e5eed5cdf1"
	snc6FamilysFavor     = "7f5bebfd-faf3-485f-b906-aa96d621d499"
)

func snc6Shield(t *testing.T, g *game.Game, owner, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddCounterByForEffect(owner, id, game.CounterShield, 1); err != nil {
			t.Fatal(err)
		}
	})
}

// snc6PickTarget answers the open target prompt, passing priority first
// when the trigger has not asked yet.
func snc6PickTarget(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID) {
	t.Helper()
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		effectstest.PassPriorityAroundTable(t, g)
		p = effectstest.LatestPickTarget(g, me.ID)
	}
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.TdCard(id)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestSncB06CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		snc6BoonOfSafety: "Boon of Safety", snc6SwoopingProtect: "Swooping Protector",
		snc6BrokersVeteran: "Brokers Veteran", snc6WingshieldAgent: "Wingshield Agent",
		snc6VoiceOfTheVermin: "Voice of the Vermin", snc6DisciplinedDuel: "Disciplined Duelist",
		snc6RhoxPummeler: "Rhox Pummeler", snc6DapperShieldmate: "Dapper Shieldmate",
		snc6Rigo: "Rigo, Streetwise Mentor", snc6SanctuaryWarden: "Sanctuary Warden",
		snc6FamilysFavor: "Family's Favor",
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

func TestSncB06CreaturesEnterWithTheirShieldCounters(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, cost string
		p, tough, want               int
	}{
		{"Swooping Protector", "Creature — Bird Citizen", snc6SwoopingProtect, "{3}{W}", 2, 1, 1},
		{"Wingshield Agent", "Creature — Human Soldier", snc6WingshieldAgent, "{2}{U}", 2, 3, 1},
		{"Voice of the Vermin", "Creature — Human Citizen", snc6VoiceOfTheVermin, "{3}{G}", 2, 2, 1},
		{"Disciplined Duelist", "Creature — Human Citizen", snc6DisciplinedDuel, "{G}{W}{U}", 2, 1, 1},
		{"Rhox Pummeler", "Creature — Rhino Soldier", snc6RhoxPummeler, "{5}{G}", 6, 3, 1},
		{"Dapper Shieldmate", "Creature — Human Soldier", snc6DapperShieldmate, "{3}{W}", 2, 2, 1},
		{"Rigo, Streetwise Mentor", "Legendary Creature — Cat Citizen", snc6Rigo, "{G/W}{W}{W/U}", 2, 2, 1},
	} {
		g := effectstest.NewCatalogGame(t)
		id := effectstest.TdCastFromHand(t, g, game.Card{Name: tc.name, TypeLine: tc.typeLine, OracleID: tc.oracle,
			ManaCost: tc.cost, Power: tc.p, Toughness: tc.tough}, game.CastSpellParams{})
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(id) {
			t.Fatalf("%s did not resolve", tc.name)
		}
		if got := effectstest.CountersOn(g, id, game.CounterShield); got != tc.want {
			t.Errorf("%s entered with %d shield counters, want %d", tc.name, got, tc.want)
		}
	}
}

func TestSncB06DisciplinedDuelistAndProtectorKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	d := effectstest.B12Push(g, me.ID, "Disciplined Duelist", "Creature — Human Citizen", snc6DisciplinedDuel, 2, 1)
	s := effectstest.B12Push(g, me.ID, "Swooping Protector", "Creature — Bird Citizen", snc6SwoopingProtect, 2, 1)
	if !effectstest.HasKeywordOf(g, d, "double strike") {
		t.Error("Duelist has no double strike")
	}
	for _, kw := range []string{"flash", "flying"} {
		if !effectstest.HasKeywordOf(g, s, kw) {
			t.Errorf("Protector has no %s", kw)
		}
	}
}

func TestSncB06BoonOfSafetyShieldsAndScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	bear := effectstest.TdCreature(g, me.ID, "Bear", 2, 2)
	effectstest.SeedLibrary(me, "Top", "Next")
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Boon of Safety", TypeLine: "Instant", OracleID: snc6BoonOfSafety,
		ManaCost: "{W}"}, game.CastSpellParams{Targets: effectstest.TdCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterShield); got != 1 {
		t.Errorf("Bear has %d shield counters, want 1", got)
	}
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Error("no scry prompt")
	}
	effectstest.B25Destroy(g, bear)
	if !g.Battlefield.Contains(bear) || effectstest.CountersOn(g, bear, game.CounterShield) != 0 {
		t.Error("the counter did not stand in for the destruction")
	}
}

func TestSncB06BoonOfSafetyNeedsACreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.PushTypedCard(g, me.ID, "Rock", "Artifact", "{1}")
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Boon of Safety", TypeLine: "Instant",
		OracleID: snc6BoonOfSafety, ManaCost: "{W}", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	boon := me.Hand.Cards[me.Hand.Size()-1].InstanceID
	if err := g.CastSpell(me.ID, boon, game.CastSpellParams{Strict: true, Targets: effectstest.TdCard(rock)}); err == nil {
		t.Error("cast Boon of Safety on an artifact")
	}
}

func TestSncB06BrokersVeteranShieldsAFriendWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vet := effectstest.B12Push(g, me.ID, "Brokers Veteran", "Creature — Human Soldier", snc6BrokersVeteran, 2, 1)
	bear := effectstest.TdCreature(g, me.ID, "Bear", 2, 2)
	effectstest.B25Destroy(g, vet)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(vet) {
		t.Fatal("the Veteran survived")
	}
	if effectstest.CountersOn(g, bear, game.CounterShield) == 0 {
		snc6PickTarget(t, g, me, bear)
	}
	if got := effectstest.CountersOn(g, bear, game.CounterShield); got != 1 {
		t.Errorf("Bear has %d shield counters, want 1", got)
	}
}

func TestSncB06WingshieldAgentGivesAnotherCreatureFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	agent := effectstest.B12Push(g, me.ID, "Wingshield Agent", "Creature — Human Soldier", snc6WingshieldAgent, 2, 3)
	bear := effectstest.TdCreature(g, me.ID, "Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, agent)
	snc6PickTarget(t, g, me, bear)
	if !effectstest.HasKeywordOf(g, bear, "flying") {
		t.Error("Bear did not gain flying")
	}
	if effectstest.HasKeywordOf(g, agent, "flying") {
		t.Error("the Agent gave itself flying")
	}
}

func TestSncB06VoiceOfTheVerminSetsBase44ButKeepsCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	voice := effectstest.B12Push(g, me.ID, "Voice of the Vermin", "Creature — Human Citizen", snc6VoiceOfTheVermin, 2, 2)
	bear := effectstest.TdCreature(g, me.ID, "Bear", 1, 1)
	if err := g.AddCounterBy(me.ID, bear, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	effectstest.DeclareAttack(t, g, opp.ID, voice)
	snc6PickTarget(t, g, me, bear)
	if p, tough := effectstest.EoeCDPT(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("Bear is %d/%d, want 5/5 (base 4/4 plus the counter)", p, tough)
	}
	if p, tough := effectstest.EoeCDPT(t, g, voice); p != 2 || tough != 2 {
		t.Errorf("Voice is %d/%d, want 2/2", p, tough)
	}
}

func TestSncB06RhoxPummelerTramplesOnlyWithAShieldCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rhox := effectstest.B12Push(g, me.ID, "Rhox Pummeler", "Creature — Rhino Soldier", snc6RhoxPummeler, 6, 3)
	if effectstest.HasKeywordOf(g, rhox, "trample") {
		t.Fatal("trample without a shield counter")
	}
	snc6Shield(t, g, me.ID, rhox)
	if !effectstest.HasKeywordOf(g, rhox, "trample") {
		t.Fatal("no trample with a shield counter")
	}
	effectstest.B25Destroy(g, rhox)
	if !g.Battlefield.Contains(rhox) || effectstest.HasKeywordOf(g, rhox, "trample") {
		t.Error("the spent counter should leave it alive and without trample")
	}
}

func TestSncB06DapperShieldmateGetsPlusTwoOnlyOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	mate := effectstest.B12Push(g, me.ID, "Dapper Shieldmate", "Creature — Human Soldier", snc6DapperShieldmate, 2, 2)
	if p, tough := effectstest.EoeCDPT(t, g, mate); p != 4 || tough != 2 {
		t.Errorf("on my turn %d/%d, want 4/2", p, tough)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, tough := effectstest.EoeCDPT(t, g, mate); p != 2 || tough != 2 {
		t.Errorf("on their turn %d/%d, want 2/2", p, tough)
	}
}

func TestSncB06RigoDrawsOnceForSmallAttackersOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Rigo, Streetwise Mentor", "Legendary Creature — Cat Citizen", snc6Rigo, 2, 2)
	big := effectstest.TdCreature(g, me.ID, "Big", 2, 2)
	effectstest.SeedLibrary(me, "A", "B", "C")
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, big)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Fatalf("drew %d for a power-2 attacker", got-hand)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	effectstest.AdvanceToMain(t, g2)
	effectstest.B12Push(g2, me2.ID, "Rigo, Streetwise Mentor", "Legendary Creature — Cat Citizen", snc6Rigo, 2, 2)
	s1 := effectstest.TdCreature(g2, me2.ID, "Small One", 1, 1)
	s2 := effectstest.TdCreature(g2, me2.ID, "Small Two", 0, 3)
	effectstest.SeedLibrary(me2, "A", "B", "C")
	hand = me2.Hand.Size()
	effectstest.DeclareAttack(t, g2, opp2.ID, s1, s2)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := me2.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d for two small attackers, want exactly 1", got-hand)
	}
}

func snc6WardenOnBattlefield(t *testing.T, g *game.Game) (me *game.Player, warden uuid.UUID) {
	t.Helper()
	me = g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C")
	warden = effectstest.TdCastFromHand(t, g, game.Card{Name: "Sanctuary Warden", TypeLine: "Creature — Angel Soldier",
		OracleID: snc6SanctuaryWarden, ManaCost: "{4}{W}{W}", Power: 5, Toughness: 5}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	return me, warden
}

func TestSncB06SanctuaryWardenRemovesACounterToDrawAndMakeACitizen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, warden := snc6WardenOnBattlefield(t, g)
	if got := effectstest.CountersOn(g, warden, game.CounterShield); got != 2 {
		t.Fatalf("Warden entered with %d shield counters, want 2", got)
	}
	hand := me.Hand.Size()
	if effectstest.LatestOptionPickFor(g, me.ID) == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, warden, game.CounterShield); got != 1 {
		t.Errorf("Warden has %d shield counters, want 1", got)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card", hand, got)
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Citizen"); n != 1 {
		t.Errorf("%d Citizens, want 1", n)
	}
}

func TestSncB06SanctuaryWardenMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, warden := snc6WardenOnBattlefield(t, g)
	hand := me.Hand.Size()
	if effectstest.LatestOptionPickFor(g, me.ID) == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, warden, game.CounterShield); got != 2 {
		t.Errorf("Warden has %d shield counters, want 2", got)
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("drew %d cards after declining", got-hand)
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Citizen"); n != 0 {
		t.Errorf("%d Citizens after declining", n)
	}
}

func TestSncB06SanctuaryWardenAttackTriggersToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	warden := effectstest.B12Push(g, me.ID, "Sanctuary Warden", "Creature — Angel Soldier", snc6SanctuaryWarden, 5, 5)
	snc6Shield(t, g, me.ID, warden)
	effectstest.SeedLibrary(me, "A", "B")
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, warden)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card", hand, got)
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Citizen"); n != 1 {
		t.Errorf("%d Citizens, want 1", n)
	}
}

func TestSncB06SanctuaryWardenWithNoCountersDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	warden := effectstest.B12Push(g, me.ID, "Sanctuary Warden", "Creature — Angel Soldier", snc6SanctuaryWarden, 5, 5)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, warden)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		t.Error("asked to remove a counter with none to remove")
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("drew %d with nothing removed", got-hand)
	}
}

func TestSncB06FamilysFavorShieldsAnAttackerAndItDrawsOnAHit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Family's Favor", "Enchantment", snc6FamilysFavor, false)
	bear := effectstest.TdCreature(g, me.ID, "Bear", 2, 2)
	effectstest.SeedLibrary(me, "A", "B")
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if effectstest.CountersOn(g, bear, game.CounterShield) == 0 {
		snc6PickTarget(t, g, me, bear)
	}
	if got := effectstest.CountersOn(g, bear, game.CounterShield); got != 1 {
		t.Fatalf("Bear has %d shield counters, want 1", got)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterShield); got != 0 {
		t.Errorf("Bear has %d shield counters after connecting, want 0", got)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card", hand, got)
	}
}

func TestSncB06FamilysFavorGrantDrawsNothingWithoutTheCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Family's Favor", "Enchantment", snc6FamilysFavor, false)
	bear := effectstest.TdCreature(g, me.ID, "Bear", 2, 2)
	effectstest.SeedLibrary(me, "A", "B")
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if effectstest.CountersOn(g, bear, game.CounterShield) == 0 {
		snc6PickTarget(t, g, me, bear)
	}
	// The counter is spent before the damage lands (a ping would do it).
	g.WithWriteLock(func() {
		if err := g.AddCounterByForEffect(me.ID, bear, game.CounterShield, -1); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("drew %d with no counter to remove", got-hand)
	}
}
