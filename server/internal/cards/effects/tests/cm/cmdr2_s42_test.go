package testscm

import (
	"slices"

	"github.com/google/uuid"

	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const plasmCaptureOracle = "3673219d-d2b5-43ad-b0c8-58b583491158"

func TestPlasmCaptureCountersAndAddsManaAtNextFirstMain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := effectstest.Batch01OpponentCasts(t, g, opp, "Big Bolt", effectstest.LightningBoltOracle, "{2}{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.CastCatalogSpell(t, g, "Plasm Capture", "Instant", plasmCaptureOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(spell) {
		t.Fatal("the spell was not countered")
	}
	if n := len(g.DelayedTriggers); n != 1 || !g.DelayedTriggers[0].ControllerTurnOnly ||
		g.DelayedTriggers[0].At != game.StepPrecombatMain {
		t.Fatalf("want one delayed trigger on the caster's precombat main, got %+v", g.DelayedTriggers)
	}
	// This turn's postcombat main is not a first main phase.
	effectstest.Batch01AdvanceToStepOf(t, g, 0, game.StepPostcombatMain)
	if len(g.DelayedTriggers) != 1 {
		t.Fatal("the mana fired in the second main phase")
	}
	// The opponent's precombat main is not mine.
	effectstest.Batch01AdvanceToStepOf(t, g, 1, game.StepPrecombatMain)
	if len(g.DelayedTriggers) != 1 {
		t.Fatal("the mana fired on an opponent's first main phase")
	}
	effectstest.Batch01AdvanceToStepOf(t, g, 0, game.StepPrecombatMain)
	if len(g.DelayedTriggers) != 0 {
		t.Fatal("the mana did not fire on my next first main phase")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []string{"W", "U", "U"} {
		pick := effectstest.RiderLatestManaPick(g, me.ID)
		if pick == nil {
			t.Fatalf("missing a colour pick for %s", c)
		}
		if err := g.ResolveManaChoice(pick.ID, me.ID, c); err != nil {
			t.Fatalf("ResolveManaChoice: %v", err)
		}
	}
	if me.ManaPool.Total() != 3 || effectstest.Msc12PoolOf(me, "W") != 1 || effectstest.Msc12PoolOf(me, "U") != 2 {
		t.Errorf("pool %v, want W U U (the countered spell's mana value, 3)", effectstest.Batch01PoolColors(me))
	}
}

const counterpointOracle = "dad4af15-5b26-417d-a41b-29e671e878f4"

func TestCounterpointCastsASpellOfEqualOrLowerManaValueFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := uuid.New()
	big := effectstest.GyCard(me, "Big Sorcery", "Sorcery", "{3}{R}", "")
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	land := effectstest.GyCard(me, "Forest", "Basic Land — Forest", "", "")
	spell := effectstest.Batch01OpponentCasts(t, g, opp, "Big Bolt", effectstest.LightningBoltOracle, "{2}{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.CastCatalogSpell(t, g, "Counterpoint", "Instant", counterpointOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(spell) {
		t.Fatal("the spell was not countered")
	}
	cast := func(id uuid.UUID) error {
		return g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", Strict: true,
			AlternativeCost: game.AltCostKeyGrantedCost})
	}
	if err := cast(big); err == nil {
		t.Error("a mana value 4 spell was cast under a cap of 3")
	}
	if err := cast(land); err == nil {
		t.Error("a land was cast")
	}
	if err := cast(bear); err != nil {
		t.Fatalf("cast the Bear free: %v", err)
	}
	if g.Stack.Size() == 0 {
		t.Fatal("the Bear is not on the stack")
	}
}

func TestCounterpointOffersNothingAboveTheCounteredSpellsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.GyCard(me, "Big Sorcery", "Sorcery", "{3}{R}", "")
	spell := effectstest.Batch01OpponentCasts(t, g, opp, "Big Bolt", effectstest.LightningBoltOracle, "{2}{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.CastCatalogSpell(t, g, "Counterpoint", "Instant", counterpointOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, big, game.CastSpellParams{FromZone: "graveyard", Strict: true,
		AlternativeCost: game.AltCostKeyGrantedCost}); err == nil {
		t.Error("cast a mana value 4 spell under a cap of 3")
	}
}

const diluvianPrimordialOracle = "f21ce158-2925-4658-ab7b-b73718d31965"

func TestDiluvianPrimordialCastsFromEachOpponentsGraveyardAndExilesThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var opps []*game.Player
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opps = append(opps, p)
		}
	}
	a := effectstest.GyCard(opps[0], "A's Sorcery", "Sorcery", "{4}{R}", "")
	creature := effectstest.GyCard(opps[0], "A's Bear", "Creature — Bear", "{1}{G}", "")
	b := effectstest.GyCard(opps[1], "B's Instant", "Instant", "{U}", "")
	mine := effectstest.GyCard(me, "My Instant", "Instant", "{U}", "")
	effectstest.CastHoldCreature(t, g, "Diluvian Primordial", "Creature — Avatar", diluvianPrimordialOracle, 5, 5)

	ch := effectstest.PickTargetPrompt(t, g)
	if slices.Contains(ch.PickTargetCards, creature) || !slices.Contains(ch.PickTargetCards, a) ||
		slices.Contains(ch.PickTargetCards, b) || ch.PickTargetMin != 0 || ch.PickTargetMax != 1 {
		t.Fatalf("first clause: only the first opponent's instants and sorceries, up to one: %v", ch.PickTargetCards)
	}
	effectstest.AnswerPickTarget(t, g, a)
	effectstest.AnswerPickTarget(t, g, b)
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{a, b} {
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", Strict: true,
			AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
			t.Fatalf("cast %s free: %v", id, err)
		}
	}
	if err := g.CastSpell(me.ID, mine, game.CastSpellParams{FromZone: "graveyard", Strict: true,
		AlternativeCost: game.AltCostKeyGrantedCost}); err == nil {
		t.Error("an untargeted card was cast")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			t.Errorf("a spell cast this way was not exiled instead of going to a graveyard")
		}
	}
}

const direFleetDaredevilOracle = "7a3e50a5-c163-4c41-b62d-d52c233c55b7"

func TestDireFleetDaredevilExilesAndLetsYouCastWithAnyManaThenExilesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	mineGY := effectstest.GyCard(me, "My Instant", "Instant", "{U}", "")
	bolt := effectstest.GyCard(opp, "Opp Bolt", "Instant", "{R}", "")
	creature := effectstest.GyCard(opp, "Opp Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.CastHoldCreature(t, g, "Dire Fleet Daredevil", "Creature — Human Pirate", direFleetDaredevilOracle, 2, 1)
	ch := effectstest.PickTargetPrompt(t, g)
	if slices.Contains(ch.PickTargetCards, mineGY) || slices.Contains(ch.PickTargetCards, creature) ||
		!slices.Contains(ch.PickTargetCards, bolt) {
		t.Fatalf("only an opponent's instant or sorcery: %v", ch.PickTargetCards)
	}
	effectstest.AnswerPickTarget(t, g, bolt)
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(bolt); z == nil || z.Kind != game.ZoneExile {
		t.Fatal("the card was not exiled")
	}
	// Not free: a pool that cannot pay refuses. Mana of any type pays {R}.
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{FromZone: "exile", Strict: true, AlternativeCost: game.AltCostKeyGrantedCost,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Fatal("cast it for free with an empty pool")
	}
	effectstest.FloatForTest(g, me, "U")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{FromZone: "exile", Strict: true, AlternativeCost: game.AltCostKeyGrantedCost,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast it with blue mana: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(bolt); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("the spell was not exiled instead of put into a graveyard: %v", z)
	}
}

const missionBriefingOracle = "c3a68018-9eff-47e6-a612-182886d28fe3"

func TestMissionBriefingSurveilsThenLetsYouCastAChosenInstantOrSorceryThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedLibrary(me, "Bin Me", "Keep Me", "Third")
	old := effectstest.GyCard(me, "Old Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
	creature := effectstest.GyCard(me, "Old Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.CastCatalogSpell(t, g, "Mission Briefing", "Instant", missionBriefingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("surveil prompt %+v, want one over two cards", c)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{ids[0]}, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, old)
	effectstest.PassPriorityAroundTable(t, g)

	cast := func(id uuid.UUID) error {
		return g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", Strict: true,
			AlternativeCost: game.AltCostKeyGrantedCost, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}})
	}
	if err := cast(creature); err == nil {
		t.Error("an unchosen card was cast")
	}
	if err := cast(old); err == nil {
		t.Fatal("cast the chosen card without paying its mana cost")
	}
	effectstest.FloatForTest(g, me, "R")
	if err := cast(old); err != nil {
		t.Fatalf("cast the chosen card: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(old); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("the spell was not exiled instead of put into the graveyard")
	}
}

const haloForagerOracle = "210d1077-6e60-4f06-a8a4-10b842979ac5"

func TestHaloForagerPaysXThenCastsAnInstantOrSorceryOfThatManaValueFromAnyGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	two := effectstest.GyCard(opp, "Two Instant", "Instant", "{1}{R}", "")
	three := effectstest.GyCard(me, "Three Sorcery", "Sorcery", "{2}{R}", "")
	bear := effectstest.GyCard(me, "Two Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPool(me, 2)
	effectstest.CastHoldCreature(t, g, "Halo Forager", "Creature — Faerie Rogue", haloForagerOracle, 3, 1)
	effectstest.PassPriorityAroundTable(t, g)
	pa := effectstest.AnswerNumber(t, g, me.ID, 2)
	if pa.Min != 0 {
		t.Fatalf("X prompt = %+v", pa)
	}
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	ch := effectstest.PickTargetPrompt(t, g)
	if !slices.Contains(ch.PickTargetCards, two) || slices.Contains(ch.PickTargetCards, three) || slices.Contains(ch.PickTargetCards, bear) {
		t.Fatalf("only an instant or sorcery with mana value exactly 2: %v", ch.PickTargetCards)
	}
	effectstest.AnswerPickTarget(t, g, two)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, two, game.CastSpellParams{FromZone: "graveyard", Strict: true,
		AlternativeCost: game.AltCostKeyGrantedCost, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast it free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(two); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("the spell was not exiled instead of put into a graveyard")
	}
}

const bloodthirstyAdversaryOracle = "3b943c3e-1f18-4997-b6d0-28ef75aed88e"

func TestBloodthirstyAdversaryPaysThenCountersAndCopiesFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bolt := effectstest.GyCard(me, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
	big := effectstest.GyCard(me, "Big Sorcery", "Sorcery", "{2}{R}{R}", "")
	oppBolt := effectstest.GyCard(opp, "Opp Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "R", 2)
	effectstest.FillPool(me, 4)
	adv := effectstest.CastHoldCreature(t, g, "Bloodthirsty Adversary", "Creature — Vampire", bloodthirstyAdversaryOracle, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	pa := effectstest.AnswerNumber(t, g, me.ID, 2)
	if pa.Max != 2 {
		t.Fatalf("X prompt = %+v, want up to 2 ({2}{R} twice from 6 mana)", pa)
	}
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	ch := effectstest.PickTargetPrompt(t, g)
	if !slices.Contains(ch.PickTargetCards, bolt) || slices.Contains(ch.PickTargetCards, big) ||
		slices.Contains(ch.PickTargetCards, oppBolt) || ch.PickTargetMax != 2 || ch.PickTargetMin != 0 {
		t.Fatalf("up to two of my instants and sorceries with mana value 3 or less: %v (%d..%d)",
			ch.PickTargetCards, ch.PickTargetMin, ch.PickTargetMax)
	}
	effectstest.AnswerPickTarget(t, g, bolt)
	effectstest.PassPriorityAroundTable(t, g)

	c, ok := effectstest.BattlefieldCard(g, adv)
	if !ok || c.Counters[game.CounterPlusOne] != 2 {
		t.Errorf("Adversary counters = %v, want two +1/+1 counters", c.Counters)
	}
	if !effectstest.InExile(g, bolt) {
		t.Fatal("the targeted card was not exiled")
	}
	cp := effectstest.CopyOf(t, g, "Lightning Bolt")
	before := opp.Life
	if err := effectstest.CastCopy(g, me.ID, cp.InstanceID, true, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("casting the copy free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("life = %d, want %d", opp.Life, before-3)
	}
}

const finaleOfPromiseOracle = "d6fafb50-9531-4fbf-bb1e-ebb4dd39281c"

func castFinaleOfPromise(t *testing.T, g *game.Game, me *game.Player, x int, targets ...uuid.UUID) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "R", 2)
	effectstest.FillPool(me, x)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Finale of Promise", TypeLine: "Sorcery", ManaCost: "{X}{R}{R}",
		OracleID: finaleOfPromiseOracle, Owner: me.ID, Controller: me.ID})
	var refs []game.TargetRef
	for _, tg := range targets {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: tg})
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: x, Targets: refs, Strict: true}); err != nil {
		t.Fatalf("cast Finale of Promise: %v", err)
	}
}

func TestFinaleOfPromiseCastsAnInstantAndASorceryFreeAndExilesThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bolt := effectstest.GyCard(me, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
	sorc := effectstest.GyCard(me, "Small Sorcery", "Sorcery", "{1}{R}", "")
	big := effectstest.GyCard(me, "Big Sorcery", "Sorcery", "{3}{R}", "")
	oppBolt := effectstest.GyCard(opp, "Opp Instant", "Instant", "{R}", effectstest.LightningBoltOracle)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Finale of Promise", TypeLine: "Sorcery", ManaCost: "{X}{R}{R}",
		OracleID: finaleOfPromiseOracle, Owner: me.ID, Controller: me.ID})
	effectstest.FillPoolColored(me, "R", 2)
	effectstest.FillPool(me, 3)
	for _, bad := range []uuid.UUID{big, oppBolt} {
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3, Strict: true,
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bad}}}); err == nil {
			t.Errorf("targeted a card above X or in another graveyard: %v", bad)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3, Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: bolt}, {Kind: game.TargetCard, ID: sorc}}}); err != nil {
		t.Fatalf("cast Finale of Promise: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	free := func(card uuid.UUID, targets ...game.TargetRef) error {
		return g.CastSpell(me.ID, card, game.CastSpellParams{FromZone: "graveyard", Strict: true,
			AlternativeCost: game.AltCostKeyGrantedCost, Targets: targets})
	}
	if err := free(bolt, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("cast the instant free: %v", err)
	}
	if err := free(sorc); err != nil {
		t.Fatalf("cast the sorcery free (its own permission): %v", err)
	}
	if err := free(big); err == nil {
		t.Error("cast an untargeted card")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []uuid.UUID{bolt, sorc} {
		if z := g.FindCardZoneForEffect(c); z == nil || z.Kind != game.ZoneExile {
			t.Errorf("a spell cast this way was not exiled instead of put into the graveyard")
		}
	}
}

func TestFinaleOfPromiseCopiesEachSpellTwiceAtXTenOrMore(t *testing.T) {
	for _, tc := range []struct{ x, wantCopies int }{{9, 0}, {10, 2}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
		bolt := effectstest.GyCard(me, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
		castFinaleOfPromise(t, g, me, tc.x, bolt)
		effectstest.PassPriorityAroundTable(t, g)
		if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{FromZone: "graveyard", Strict: true,
			AlternativeCost: game.AltCostKeyGrantedCost, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
			t.Fatalf("X=%d: cast the instant free: %v", tc.x, err)
		}
		// Each copy offers new targets (CR 707.10c): keep the opponent.
		prompts := 0
		for c := effectstest.CopyPromptFor(g, me.ID); c != nil && prompts < 4; c = effectstest.CopyPromptFor(g, me.ID) {
			prompts++
			if err := g.ResolvePickTarget(c.ID, me.ID, effectstest.PlayerTarget(opp.ID)); err != nil {
				t.Fatalf("X=%d: new targets for a copy: %v", tc.x, err)
			}
		}
		if prompts != tc.wantCopies {
			t.Errorf("X=%d: asked for new targets %d times, want %d", tc.x, prompts, tc.wantCopies)
		}
		if got := effectstest.StackCopies(g); got != tc.wantCopies {
			t.Errorf("X=%d: %d copies on the stack, want %d", tc.x, got, tc.wantCopies)
		}
		before := opp.Life
		effectstest.PassPriorityAroundTable(t, g)
		if want := before - 3*(1+tc.wantCopies); opp.Life != want {
			t.Errorf("X=%d: life %d, want %d", tc.x, opp.Life, want)
		}
	}
}
