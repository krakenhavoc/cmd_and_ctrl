package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mom_b18_test.go — March of the Machine Sieges, batch 18 (#5477): the
// eight Sieges and their back faces, driven through the engine.

// b18Row is a Siege's Scryfall record in the shape deck.ToGameCard
// consumes: a transform layout, the battle face first.
func b18Row(oracle, front, back, cost, defense, backType, power, toughness string, colors ...string) cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     front + " // " + back,
		Layout:   "transform",
		TypeLine: "Battle — Siege // " + backType,
		OracleID: uuid.MustParse(oracle),
		CardFaces: []cards.CardFace{
			{Name: front, TypeLine: "Battle — Siege", ManaCost: cost, Defense: defense, Colors: colors},
			{Name: back, TypeLine: backType, Power: power, Toughness: toughness, Colors: colors},
		},
	}
}

type b18Table struct {
	g        *game.Game
	me, them *game.Player
	siege    uuid.UUID
}

// b18Cast puts the row in the active player's hand, casts it and answers
// the protector prompt, leaving the enters trigger for the caller.
func b18Cast(t *testing.T, row cards.Card, setup func(g *game.Game, me, them *game.Player)) b18Table {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, them := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	if setup != nil {
		setup(g, me, them)
	}
	id := effectstest.ImportToHand(row, me)
	effectstest.ToMainPhaseSameTurn(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the Siege: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerProtectorPrompt(t, g, me.ID, them.ID)
	return b18Table{g: g, me: me, them: them, siege: id}
}

// b18Pick answers the newest pick_target prompt with the given refs.
func b18Pick(t *testing.T, g *game.Game, chooser uuid.UUID, refs ...game.TargetRef) {
	t.Helper()
	p := effectstest.LatestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	if err := g.ResolvePickTargets(p.ID, chooser, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
}

func b18Card(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

// b18DefeatAndCast defeats the Siege and casts its back face from exile.
func b18DefeatAndCast(t *testing.T, tb b18Table, defense int) *game.Card {
	t.Helper()
	effectstest.DefeatBattle(t, tb.g, tb.siege, defense)
	if !effectstest.InExile(tb.g, tb.siege) {
		t.Fatal("the defeated Siege was not exiled")
	}
	if err := effectstest.CastThroughResolutionCast(t, tb.g, tb.me.ID, tb.siege, game.CastSpellParams{}); err != nil {
		t.Fatalf("casting the back face: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, tb.g)
	return effectstest.BattlefieldCardFor(tb.g, tb.siege)
}

func b18Defense(t *testing.T, tb b18Table, want int) {
	t.Helper()
	b := effectstest.BattlefieldCardFor(tb.g, tb.siege)
	if b == nil || b.Counters[game.CounterDefense] != want || !b.IsBattle() {
		t.Fatalf("the Siege: %+v, want a battle with %d defense", b, want)
	}
}

// --- Invasion of Kaladesh // Aetherwing, Golden-Scale Flagship ----------------

const b18KaladeshOracle = "bc6f4f73-e10c-44ae-926d-9e5f6ddc14be"

func b18KaladeshRow() cards.Card {
	return b18Row(b18KaladeshOracle, "Invasion of Kaladesh", "Aetherwing, Golden-Scale Flagship",
		"{U}{R}", "4", "Legendary Artifact — Vehicle", "*", "4", "U", "R")
}

func TestInvasionOfKaladeshMakesAThopterAndBecomesAetherwing(t *testing.T) {
	tb := b18Cast(t, b18KaladeshRow(), nil)
	effectstest.PassPriorityAroundTable(t, tb.g)
	b18Defense(t, tb, 4)
	thopters := effectstest.TokensNamed(tb.g, "Thopter")
	if len(thopters) != 1 {
		t.Fatalf("made %d Thopters, want 1", len(thopters))
	}
	th := thopters[0]
	if th.Controller != tb.me.ID || th.Power != 1 || th.Toughness != 1 || len(th.Colors) != 0 ||
		!th.IsArtifact() || !game.HasKeyword(&th, "flying") {
		t.Errorf("the token is %+v, want a 1/1 colorless artifact Thopter with flying", th)
	}
	wing := b18DefeatAndCast(t, tb, 4)
	if wing == nil || wing.Name != "Aetherwing, Golden-Scale Flagship" || wing.IsBattle() || !wing.IsArtifact() ||
		!wing.HasSubtype("Vehicle") || wing.IsCreature() {
		t.Fatalf("landed %+v, want Aetherwing as an uncrewed Vehicle", wing)
	}
	// Its power counts the artifacts you control: itself and the Thopter.
	if got := wing.CurrentPower(); got != 2 {
		t.Errorf("power = %d, want 2 (Aetherwing and the Thopter)", got)
	}
	if got := wing.CurrentToughness(); got != 4 {
		t.Errorf("toughness = %d, want 4", got)
	}
	if !game.HasKeyword(wing, "flying") {
		t.Error("Aetherwing should have flying")
	}
}

func TestAetherwingPowerGrowsWithArtifactsAndCrewsWithOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	wing := effectstest.B27Push(g, me.ID, "Aetherwing, Golden-Scale Flagship", "Legendary Artifact — Vehicle",
		b18KaladeshOracle+"#1", "", 0, 4, "U", "R")
	effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	if got := effectstest.ApaLive(g, wing).CurrentPower(); got != 3 {
		t.Errorf("power with two other artifacts = %d, want 3", got)
	}
	// Crew 1: a power-0 creature cannot crew it, a power-1 creature can.
	small := effectstest.B12Creature(g, me.ID, "Wisp", "Creature — Spirit", 0, 1)
	crew := effectstest.B12Creature(g, me.ID, "Pilot", "Creature — Human", 1, 1)
	params := func(id uuid.UUID) game.ActivateAbilityParams {
		return game.ActivateAbilityParams{Strict: true, CrewIDs: []uuid.UUID{id}}
	}
	if err := g.ActivateCatalogAbility(me.ID, wing, 0, params(small)); err == nil {
		t.Error("a power-0 creature crewed Aetherwing")
	}
	if err := g.ActivateCatalogAbility(me.ID, wing, 0, params(crew)); err != nil {
		t.Fatalf("crewing with a power-1 creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.ApaLive(g, wing); c == nil || !c.IsCreature() {
		t.Error("crewed Aetherwing should be an artifact creature")
	}
}

// --- Invasion of Theros // Ephara, Ever-Sheltering ----------------------------

const b18TherosOracle = "ba5f7f37-352a-4b9c-b648-0ba5c8d842fc"

func b18TherosRow() cards.Card {
	return b18Row(b18TherosOracle, "Invasion of Theros", "Ephara, Ever-Sheltering",
		"{2}{W}", "4", "Legendary Enchantment Creature — God", "4", "4", "W")
}

func TestInvasionOfTherosFetchesAnAuraGodOrDemigod(t *testing.T) {
	var aura, god, demigod, plain uuid.UUID
	tb := b18Cast(t, b18TherosRow(), func(g *game.Game, me, _ *game.Player) {
		mk := func(name, tl string) uuid.UUID {
			c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: tl, Owner: me.ID, Controller: me.ID}
			me.Library.PushTop(c)
			return c.InstanceID
		}
		plain = mk("Bear", "Creature — Bear")
		aura = mk("Pacifism", "Enchantment — Aura")
		god = mk("Some God", "Legendary Enchantment Creature — God")
		demigod = mk("Some Demigod", "Legendary Enchantment Creature — Demigod")
	})
	effectstest.PassPriorityAroundTable(t, tb.g)
	// The searcher is offered exactly the three qualifying cards.
	found := effectstest.LatestChoiceOfKindFor(tb.g, game.PendingChoiceSearchLibrary, tb.me.ID)
	if found == nil {
		t.Fatal("the enters trigger did not ask what to search for")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range found.SearchCards {
		offered[id] = true
	}
	if !offered[aura] || !offered[god] || !offered[demigod] || offered[plain] {
		t.Fatalf("offered %v, want the Aura, the God and the Demigod but not the Bear", offered)
	}
	if err := tb.g.ResolveSearchLibrary(found.ID, tb.me.ID, []uuid.UUID{god}); err != nil {
		t.Fatalf("choosing the God: %v", err)
	}
	if !effectstest.InHandOf(tb.me, god) || effectstest.InHandOf(tb.me, aura) {
		t.Error("only the chosen God should be in hand")
	}
}

func TestEpharaNeedsThreeOtherEnchantmentsAndDrawsOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	e := effectstest.B27Push(g, me.ID, "Ephara, Ever-Sheltering", "Legendary Enchantment Creature — God",
		b18TherosOracle+"#1", "", 4, 4, "W")
	view := func() *game.Card { return effectstest.ApaLive(g, e) }
	if game.HasKeyword(view(), "lifelink") || game.HasKeyword(view(), "indestructible") {
		t.Fatal("Ephara has keywords with no other enchantments")
	}
	for i := 0; i < 2; i++ {
		effectstest.B12Permanent(g, me.ID, "Ench", "Enchantment")
	}
	// An opponent's enchantment is not "you control".
	effectstest.B12Permanent(g, opp.ID, "Theirs", "Enchantment")
	if game.HasKeyword(view(), "lifelink") {
		t.Fatal("two of mine and one of theirs is not three other enchantments")
	}
	effectstest.B12Permanent(g, me.ID, "Ench", "Enchantment")
	if !game.HasKeyword(view(), "lifelink") || !game.HasKeyword(view(), "indestructible") {
		t.Error("with three other enchantments Ephara should have lifelink and indestructible")
	}

	// "Another enchantment you control enters": a draw, not for a nonenchantment or an opponent's.
	before := me.Hand.Size()
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before {
		t.Error("a nonenchantment entering drew a card")
	}
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Theirs", TypeLine: "Enchantment", Owner: opp.ID, Controller: opp.ID})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before {
		t.Error("an opponent's enchantment entering drew a card")
	}
}

// --- Invasion of Tolvada // The Broken Sky ------------------------------------

const b18TolvadaOracle = "7f8dad7f-a573-469f-9095-4ba3886b4f0b"

func b18TolvadaRow() cards.Card {
	return b18Row(b18TolvadaOracle, "Invasion of Tolvada", "The Broken Sky",
		"{3}{W}{B}", "5", "Enchantment", "", "", "W", "B")
}

func TestInvasionOfTolvadaReturnsANonbattlePermanentCard(t *testing.T) {
	var bear, land, instant uuid.UUID
	tb := b18Cast(t, b18TolvadaRow(), func(g *game.Game, me, _ *game.Player) {
		bear = effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
		land = effectstest.PushGraveyardCardTyped(me, "Dead Land", "Land")
		instant = effectstest.PushGraveyardCardTyped(me, "Spent Spell", "Instant")
	})
	pick := effectstest.LatestPickTarget(tb.g, tb.me.ID)
	if pick == nil {
		effectstest.PassPriorityAroundTable(t, tb.g)
		pick = effectstest.LatestPickTarget(tb.g, tb.me.ID)
	}
	if pick == nil {
		t.Fatal("the enters trigger did not ask for a target")
	}
	if err := tb.g.ResolvePickTargets(pick.ID, tb.me.ID, []game.TargetRef{b18Card(instant)}); err == nil {
		t.Error("an instant card was accepted as the target")
	}
	b18Pick(t, tb.g, tb.me.ID, b18Card(bear))
	effectstest.PassPriorityAroundTable(t, tb.g)
	if !tb.g.Battlefield.Contains(bear) || tb.g.Battlefield.Contains(land) {
		t.Error("only the chosen Bear should return")
	}
	b18Defense(t, tb, 5)
}

func TestTheBrokenSkyPumpsAndLinksTokensAndMakesSpirits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	tokenID := effectstest.B12Creature(g, me.ID, "Goblin", "Token Creature — Goblin", 1, 1)
	nonToken := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Goblin", "Token Creature — Goblin", 1, 1)
	effectstest.B27Push(g, me.ID, "The Broken Sky", "Enchantment", b18TolvadaOracle+"#1", "", 0, 0, "W", "B")
	if c := effectstest.ApaLive(g, tokenID); c.CurrentPower() != 2 || !game.HasKeyword(c, "lifelink") {
		t.Errorf("my token is %d/%d, want 2/1 with lifelink", c.CurrentPower(), c.CurrentToughness())
	}
	if c := effectstest.ApaLive(g, nonToken); c.CurrentPower() != 2 || game.HasKeyword(c, "lifelink") {
		t.Error("a nontoken creature was changed")
	}
	if c := effectstest.ApaLive(g, theirs); c.CurrentPower() != 1 || game.HasKeyword(c, "lifelink") {
		t.Error("an opponent's token was changed")
	}
	// The end step makes a 1/1 white and black Spirit with flying (and it gets the bonus).
	for i := 0; i < 12 && g.Turn.Step != game.StepEnd; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	spirits := effectstest.TokensNamed(g, "Spirit")
	if len(spirits) != 1 {
		t.Fatalf("made %d Spirits, want 1", len(spirits))
	}
	s := spirits[0]
	if s.Controller != me.ID || s.CurrentPower() != 2 || s.CurrentToughness() != 1 || !game.HasKeyword(&s, "flying") ||
		!game.HasKeyword(&s, "lifelink") || !s.HasColorWord("W") || !s.HasColorWord("B") {
		t.Errorf("the Spirit is %+v, want a white and black 1/1 flier that the Sky makes 2/1 with lifelink", s)
	}
}

// --- Invasion of Shandalar // Leyline Surge -----------------------------------

const b18ShandalarOracle = "b5fe62f0-d914-4789-afd3-3831dcdbed7b"

func b18ShandalarRow() cards.Card {
	return b18Row(b18ShandalarOracle, "Invasion of Shandalar", "Leyline Surge",
		"{3}{G}{G}", "4", "Enchantment", "", "", "G")
}

func TestInvasionOfShandalarReturnsUpToThreePermanentCards(t *testing.T) {
	var a, b, c, d, instant uuid.UUID
	tb := b18Cast(t, b18ShandalarRow(), func(g *game.Game, me, _ *game.Player) {
		a = effectstest.PushGraveyardCardTyped(me, "A", "Creature — Bear")
		b = effectstest.PushGraveyardCardTyped(me, "B", "Artifact")
		c = effectstest.PushGraveyardCardTyped(me, "C", "Land")
		d = effectstest.PushGraveyardCardTyped(me, "D", "Enchantment")
		instant = effectstest.PushGraveyardCardTyped(me, "Spell", "Sorcery")
	})
	pick := effectstest.LatestPickTarget(tb.g, tb.me.ID)
	if pick == nil {
		effectstest.PassPriorityAroundTable(t, tb.g)
		pick = effectstest.LatestPickTarget(tb.g, tb.me.ID)
	}
	if pick == nil {
		t.Fatal("the enters trigger did not ask for targets")
	}
	if err := tb.g.ResolvePickTargets(pick.ID, tb.me.ID, []game.TargetRef{b18Card(a), b18Card(b), b18Card(c), b18Card(d)}); err == nil {
		t.Error("four targets were accepted")
	}
	if err := tb.g.ResolvePickTargets(pick.ID, tb.me.ID, []game.TargetRef{b18Card(instant)}); err == nil {
		t.Error("a sorcery card was accepted")
	}
	b18Pick(t, tb.g, tb.me.ID, b18Card(a), b18Card(b), b18Card(c))
	effectstest.PassPriorityAroundTable(t, tb.g)
	for _, id := range []uuid.UUID{a, b, c} {
		if !effectstest.InHandOf(tb.me, id) {
			t.Errorf("%s did not return to hand", id)
		}
	}
	if effectstest.InHandOf(tb.me, d) {
		t.Error("the untargeted card returned")
	}
}

func TestLeylineSurgePutsAPermanentCardFromHandAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B27Push(g, me.ID, "Leyline Surge", "Enchantment", b18ShandalarOracle+"#1", "", 0, 0, "G")
	bear := game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}
	spell := game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() { me.Hand.PushTop(bear); me.Hand.PushTop(spell) })
	// Next own upkeep.
	for i := 0; i < 80 && !(g.Turn.ActiveSeat == seatIdxB18(g, me.ID) && g.Turn.Step == game.StepUpkeep && g.Turn.Seq > 1); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	var ch *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == me.ID && c.Kind == game.PendingChoiceChooseCards {
			ch = c
		}
	}
	if ch == nil {
		t.Fatal("the upkeep trigger did not ask which permanent card to put onto the battlefield")
	}
	for _, id := range ch.ChooseCards {
		if id == spell.InstanceID {
			t.Error("a nonpermanent card was offered")
		}
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{bear.InstanceID}); err != nil {
		t.Fatal(err)
	}
	if !g.Battlefield.Contains(bear.InstanceID) {
		t.Error("the Bear did not enter")
	}
}

func seatIdxB18(g *game.Game, id uuid.UUID) int {
	for i, p := range g.Seats {
		if p.ID == id {
			return i
		}
	}
	return -1
}

// b18PickClauses answers the pick_target prompts one clause at a time, in
// clause order, skipping an empty list (an "up to" clause left unchosen).
func b18PickClauses(t *testing.T, g *game.Game, chooser uuid.UUID, clauses ...[]game.TargetRef) {
	t.Helper()
	for i, refs := range clauses {
		if len(refs) == 0 && effectstest.LatestPickTarget(g, chooser) == nil {
			continue
		}
		p := effectstest.LatestPickTarget(g, chooser)
		if p == nil {
			t.Fatalf("no pick_target prompt for clause %d", i)
		}
		if err := g.ResolvePickTargets(p.ID, chooser, refs); err != nil {
			t.Fatalf("clause %d: ResolvePickTargets: %v", i, err)
		}
	}
}

func b18WaitForPick(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	if effectstest.LatestPickTarget(g, chooser) == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
}

// --- Invasion of Kylem // Valor's Reach Tag Team -----------------------------

const b18KylemOracle = "fee0aa18-5103-4814-94e3-87ad3eaf5a6f"

func b18KylemRow() cards.Card {
	return b18Row(b18KylemOracle, "Invasion of Kylem", "Valor's Reach Tag Team",
		"{2}{R}{W}", "5", "Sorcery", "", "", "R", "W")
}

func TestInvasionOfKylemPumpsUpToTwoCreaturesAndBecomesTwoWarriors(t *testing.T) {
	var mine, theirs, third uuid.UUID
	tb := b18Cast(t, b18KylemRow(), func(g *game.Game, me, them *game.Player) {
		mine = effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		theirs = effectstest.B12Creature(g, them.ID, "Their Bear", "Creature — Bear", 2, 2)
		third = effectstest.B12Creature(g, me.ID, "Third", "Creature — Bear", 2, 2)
	})
	b18WaitForPick(t, tb.g, tb.me.ID)
	if p := effectstest.LatestPickTarget(tb.g, tb.me.ID); p == nil {
		t.Fatal("the enters trigger did not ask for targets")
	} else if err := tb.g.ResolvePickTargets(p.ID, tb.me.ID, []game.TargetRef{b18Card(mine), b18Card(theirs), b18Card(third)}); err == nil {
		t.Error("three targets were accepted")
	}
	b18Pick(t, tb.g, tb.me.ID, b18Card(mine), b18Card(theirs))
	effectstest.PassPriorityAroundTable(t, tb.g)
	for _, id := range []uuid.UUID{mine, theirs} {
		c := effectstest.ApaLive(tb.g, id)
		if c.CurrentPower() != 4 || c.CurrentToughness() != 2 || !game.HasKeyword(c, "vigilance") || !game.HasKeyword(c, "haste") {
			t.Errorf("%s is %d/%d, want 4/2 with vigilance and haste", c.Name, c.CurrentPower(), c.CurrentToughness())
		}
	}
	if c := effectstest.ApaLive(tb.g, third); c.CurrentPower() != 2 || game.HasKeyword(c, "haste") {
		t.Error("the untargeted creature changed")
	}
	b18Defense(t, tb, 5)

	if got := b18DefeatAndCast(t, tb, 5); got != nil {
		t.Errorf("the sorcery stayed on the battlefield: %+v", got)
	}
	warriors := effectstest.TokensNamed(tb.g, "Warrior")
	if len(warriors) != 2 {
		t.Fatalf("made %d Warriors, want 2", len(warriors))
	}
	for _, w := range warriors {
		if w.Controller != tb.me.ID || w.Power != 3 || w.Toughness != 2 || !w.HasColorWord("R") || !w.HasColorWord("W") || len(w.Colors) != 2 {
			t.Errorf("the token is %+v, want a 3/2 red and white Warrior", w)
		}
	}
}

func TestValorsReachWarriorGrowsOnlyWithAnotherCreatureTokenAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mk := func() uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, func() game.Card {
			c := effects.TokenFromCatalog(effects.PrintedValorsReachWarriorToken)
			c.InstanceID = uuid.New()
			c.Owner, c.Controller = me.ID, me.ID
			return c
		}())
	}
	a := mk()
	nonToken := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	counters := func(id uuid.UUID) int { return effectstest.ApaLive(g, id).Counters[game.CounterPlusOne] }

	// A Warrior attacking with only a nontoken creature gets nothing.
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(nonToken, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if counters(a) != 0 {
		t.Fatalf("a Warrior attacking with a nontoken creature got %d counters, want 0", counters(a))
	}
}

func TestValorsReachWarriorsBothGrowWhenTheyAttackTogether(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mk := func() uuid.UUID {
		return effectstest.PushToken(g, me.ID, effects.TokenFromCatalog(effects.PrintedValorsReachWarriorToken))
	}
	a, b := mk(), mk()
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		c := effectstest.ApaLive(g, id)
		if c.Counters[game.CounterPlusOne] != 1 || c.CurrentPower() != 4 {
			t.Errorf("a Warrior attacking with another token has %d counters / power %d, want 1 / 4", c.Counters[game.CounterPlusOne], c.CurrentPower())
		}
	}
}

// --- Invasion of Muraganda // Primordial Plasm -------------------------------

const b18MuragandaOracle = "2e0d12dc-55fa-4363-b28f-612e7cfa12c6"

func b18MuragandaRow() cards.Card {
	return b18Row(b18MuragandaOracle, "Invasion of Muraganda", "Primordial Plasm",
		"{4}{G}", "6", "Creature — Ooze", "4", "4", "G")
}

func TestInvasionOfMuragandaCountersThenFights(t *testing.T) {
	var mine, theirs uuid.UUID
	tb := b18Cast(t, b18MuragandaRow(), func(g *game.Game, me, them *game.Player) {
		mine = effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		theirs = effectstest.B12Creature(g, them.ID, "Their Bear", "Creature — Bear", 3, 3)
	})
	b18WaitForPick(t, tb.g, tb.me.ID)
	b18PickClauses(t, tb.g, tb.me.ID, []game.TargetRef{b18Card(mine)}, []game.TargetRef{b18Card(theirs)})
	effectstest.PassPriorityAroundTable(t, tb.g)
	// The counter first: 3/3 fights a 3/3, so both die.
	if tb.g.Battlefield.Contains(theirs) {
		t.Error("the countered 3/3 should have killed the 3/3 it fought")
	}
	if tb.g.Battlefield.Contains(mine) {
		t.Error("the 3/3 took 3 damage and should have died")
	}
	b18Defense(t, tb, 6)
}

func TestInvasionOfMuragandaFightsNothingWhenTheSecondTargetIsSkipped(t *testing.T) {
	var mine uuid.UUID
	tb := b18Cast(t, b18MuragandaRow(), func(g *game.Game, me, them *game.Player) {
		mine = effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.B12Creature(g, them.ID, "Their Bear", "Creature — Bear", 3, 3)
	})
	b18WaitForPick(t, tb.g, tb.me.ID)
	b18PickClauses(t, tb.g, tb.me.ID, []game.TargetRef{b18Card(mine)}, nil)
	effectstest.PassPriorityAroundTable(t, tb.g)
	c := effectstest.ApaLive(tb.g, mine)
	if c == nil || c.Counters[game.CounterPlusOne] != 1 || c.DamageMarked != 0 {
		t.Errorf("the creature: %+v, want a counter and no damage", c)
	}
}

func TestPrimordialPlasmShrinksAnotherCreatureToItsStatsPlusTwoWithNoAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	plasm := effectstest.B27Push(g, me.ID, "Primordial Plasm", "Creature — Ooze", b18MuragandaOracle+"#1", "", 4, 4, "G")
	target := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Flier", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	b18WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the combat trigger did not ask for a target")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{b18Card(plasm)}); err == nil {
		t.Error("the Plasm targeted itself")
	}
	b18Pick(t, g, me.ID, b18Card(target))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ApaLive(g, target)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 3 || game.HasKeyword(c, "flying") {
		t.Errorf("the target is %d/%d flying=%v, want 3/3 with no abilities", c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(c, "flying"))
	}
}

// --- Invasion of Regatha // Disciples of the Inferno -------------------------

const b18RegathaOracle = "4f0bcfe5-7e52-4249-a26a-482619716f18"

func b18RegathaRow() cards.Card {
	return b18Row(b18RegathaOracle, "Invasion of Regatha", "Disciples of the Inferno",
		"{2}{R}", "5", "Creature — Human Monk", "4", "4", "R")
}

func TestInvasionOfRegathaHitsAnOpponentAndACreature(t *testing.T) {
	var bear uuid.UUID
	tb := b18Cast(t, b18RegathaRow(), func(g *game.Game, me, them *game.Player) {
		bear = effectstest.B12Creature(g, them.ID, "Bear", "Creature — Bear", 1, 1)
	})
	b18WaitForPick(t, tb.g, tb.me.ID)
	before := tb.them.Life
	b18PickClauses(t, tb.g, tb.me.ID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: tb.them.ID}}, []game.TargetRef{b18Card(bear)})
	effectstest.PassPriorityAroundTable(t, tb.g)
	if tb.them.Life != before-4 {
		t.Errorf("the opponent went %d → %d, want 4 damage", before, tb.them.Life)
	}
	if tb.g.Battlefield.Contains(bear) {
		t.Error("1 damage did not kill the 1/1")
	}
	b18Defense(t, tb, 5)
}

func TestInvasionOfRegathaCanHitAnotherBattleButNotItselfOrYou(t *testing.T) {
	var other uuid.UUID
	tb := b18Cast(t, b18RegathaRow(), func(g *game.Game, me, them *game.Player) {
		other = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Other Siege", TypeLine: "Battle — Siege",
			Counters: map[string]int{game.CounterDefense: 6}, Owner: them.ID, Controller: them.ID})
	})
	b18WaitForPick(t, tb.g, tb.me.ID)
	p := effectstest.LatestPickTarget(tb.g, tb.me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	for name, ref := range map[string]game.TargetRef{
		"itself": b18Card(tb.siege), "its caster": {Kind: game.TargetPlayer, ID: tb.me.ID},
	} {
		if err := tb.g.ResolvePickTargets(p.ID, tb.me.ID, []game.TargetRef{ref}); err == nil {
			t.Errorf("the Siege targeted %s", name)
		}
	}
	b18PickClauses(t, tb.g, tb.me.ID, []game.TargetRef{b18Card(other)}, nil)
	effectstest.PassPriorityAroundTable(t, tb.g)
	if c := effectstest.ApaLive(tb.g, other); c == nil || c.Counters[game.CounterDefense] != 2 {
		t.Errorf("the other battle: %+v, want 6-4 = 2 defense", c)
	}
}

func TestDisciplesAddTwoToNoncreatureSourcesDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	effectstest.B27Push(g, me.ID, "Disciples of the Inferno", "Creature — Human Monk", b18RegathaOracle+"#1", "", 4, 4, "R")
	rock := effectstest.B12Permanent(g, me.ID, "Pinger", "Artifact")
	theirRock := effectstest.B12Permanent(g, opp.ID, "Their Pinger", "Artifact")
	biter := effectstest.B12Creature(g, me.ID, "Biter", "Creature — Bear", 2, 2)
	victim := effectstest.B12Creature(g, me.ID, "Own Bear", "Creature — Bear", 9, 9)
	walker := effectstest.PushWalkerForTest(g, opp.ID, "Walker", "", 9)

	ping := func(src, target uuid.UUID, player bool, n int) {
		g.WithWriteLock(func() {
			var err error
			if player {
				err = g.DealDamageToPlayerForEffect(src, target, n)
			} else {
				err = g.DealDamageToCreatureForEffect(src, target, n)
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
	life := opp.Life
	ping(rock, opp.ID, true, 1)
	if opp.Life != life-3 {
		t.Errorf("a noncreature source dealt %d to an opponent, want 3", life-opp.Life)
	}
	life = opp.Life
	ping(biter, opp.ID, true, 1)
	if opp.Life != life-1 {
		t.Errorf("a creature source dealt %d, want 1", life-opp.Life)
	}
	life = opp.Life
	ping(theirRock, opp.ID, true, 1)
	if opp.Life != life-1 {
		t.Errorf("an opponent's noncreature source dealt %d, want 1", life-opp.Life)
	}
	// To a creature of mine: a creature is a creature.
	ping(rock, victim, false, 1)
	if got := effectstest.ApaLive(g, victim).DamageMarked; got != 3 {
		t.Errorf("a noncreature source dealt %d to a creature, want 3", got)
	}
	// Not to yourself, not to a planeswalker.
	life = me.Life
	ping(rock, me.ID, true, 1)
	if me.Life != life-1 {
		t.Errorf("damage to its controller was %d, want 1", life-me.Life)
	}
	ping(rock, walker, false, 1)
	if got := effectstest.ApaLive(g, walker).Counters[game.CounterLoyalty]; got != 8 {
		t.Errorf("a planeswalker has %d loyalty, want 9-1 = 8", got)
	}
}

// --- Invasion of New Capenna // Holy Frazzle-Cannon --------------------------

const b18NewCapennaOracle = "530dc119-e266-4525-abe9-7fab599e9de2"

func b18NewCapennaRow() cards.Card {
	return b18Row(b18NewCapennaOracle, "Invasion of New Capenna", "Holy Frazzle-Cannon",
		"{W}{B}", "4", "Artifact — Equipment", "", "", "W", "B")
}

func TestInvasionOfNewCapennaSacrificesToExileAnOpposingPermanent(t *testing.T) {
	var rock, bear, theirBear, theirRock, theirLand uuid.UUID
	tb := b18Cast(t, b18NewCapennaRow(), func(g *game.Game, me, them *game.Player) {
		rock = effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
		bear = effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		theirBear = effectstest.B12Creature(g, them.ID, "Their Bear", "Creature — Bear", 2, 2)
		theirRock = effectstest.B12Permanent(g, them.ID, "Their Rock", "Artifact")
		theirLand = effectstest.B12Permanent(g, them.ID, "Their Land", "Land")
	})
	effectstest.PassPriorityAroundTable(t, tb.g)
	c := effectstest.LatestChoiceOfKindFor(tb.g, game.PendingChoiceOwnPermanents, tb.me.ID)
	if c == nil {
		t.Fatal("the enters trigger did not offer a sacrifice")
	}
	if !effectstest.HasID(c.ChooseCards, rock) || !effectstest.HasID(c.ChooseCards, bear) || effectstest.HasID(c.ChooseCards, theirBear) {
		t.Errorf("the prompt offers %v, want my artifact and my creature only", c.ChooseCards)
	}
	effectstest.AnswerOwnPermanents(t, tb.g, tb.me.ID, rock)
	b18WaitForPick(t, tb.g, tb.me.ID)
	p := effectstest.LatestPickTarget(tb.g, tb.me.ID)
	if p == nil {
		t.Fatal("the reflexive trigger did not ask for a target")
	}
	for name, id := range map[string]uuid.UUID{"my own creature": bear, "their land": theirLand} {
		if err := tb.g.ResolvePickTargets(p.ID, tb.me.ID, []game.TargetRef{b18Card(id)}); err == nil {
			t.Errorf("targeted %s", name)
		}
	}
	b18Pick(t, tb.g, tb.me.ID, b18Card(theirRock))
	effectstest.PassPriorityAroundTable(t, tb.g)
	if tb.g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if tb.g.Battlefield.Contains(theirRock) || !effectstest.InExile(tb.g, theirRock) {
		t.Error("the opponent's artifact was not exiled")
	}
	if !tb.g.Battlefield.Contains(theirBear) {
		t.Error("the wrong permanent was exiled")
	}
}

func TestInvasionOfNewCapennaDeclinedSacrificeExilesNothing(t *testing.T) {
	var theirBear uuid.UUID
	tb := b18Cast(t, b18NewCapennaRow(), func(g *game.Game, me, them *game.Player) {
		effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
		theirBear = effectstest.B12Creature(g, them.ID, "Their Bear", "Creature — Bear", 2, 2)
	})
	effectstest.PassPriorityAroundTable(t, tb.g)
	c := effectstest.LatestChoiceOfKindFor(tb.g, game.PendingChoiceOwnPermanents, tb.me.ID)
	if c == nil {
		t.Fatal("the enters trigger did not offer a sacrifice")
	}
	if err := tb.g.ResolveOwnPermanents(c.ID, tb.me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, tb.g)
	if !tb.g.Battlefield.Contains(theirBear) || effectstest.LatestPickTarget(tb.g, tb.me.ID) != nil {
		t.Error("declining the sacrifice still exiled or asked for a target")
	}
}

func TestHolyFrazzleCannonPumpsTheAttackerAndItsTypeMates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	cannon := effectstest.B27Push(g, me.ID, "Holy Frazzle-Cannon", "Artifact — Equipment", b18NewCapennaOracle+"#1", "", 0, 0)
	knight := effectstest.B12Creature(g, me.ID, "Knight", "Creature — Human Knight", 2, 2)
	mate := effectstest.B12Creature(g, me.ID, "Mate", "Creature — Elf Knight", 2, 2)
	stranger := effectstest.B12Creature(g, me.ID, "Stranger", "Creature — Goblin", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Knight", "Creature — Human Knight", 2, 2)
	changeling := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Shifter", TypeLine: "Creature — Shapeshifter",
		Keywords: []string{"changeling"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(cannon, game.TargetRef{Kind: game.TargetCard, ID: knight}); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(knight, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	n := func(id uuid.UUID) int { return effectstest.ApaLive(g, id).Counters[game.CounterPlusOne] }
	if n(knight) != 1 || n(mate) != 1 || n(changeling) != 1 {
		t.Errorf("counters: attacker %d, Knight mate %d, changeling %d, want 1 each", n(knight), n(mate), n(changeling))
	}
	if n(stranger) != 0 || n(theirs) != 0 {
		t.Errorf("counters: stranger %d, opponent's Knight %d, want 0", n(stranger), n(theirs))
	}
}
