package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blb_mp2_cards_test.go — Bloomburrow cards M through P, second batch
// (#4467). Each test plays the real catalog entry and asserts what its
// text says.

const (
	blbMP2MabelOracle      = "4b11e435-5888-4e3b-a255-db817392163f"
	blbMP2ManifoldOracle   = "b254ca74-2a96-4590-91bc-650f407daf5d"
	blbMP2SpiralOracle     = "03e356ae-f1e8-4730-8d2b-b9e2f32db915"
	blbMP2WhiskerOracle    = "f823b122-94ac-42ba-b511-9e54d64db112"
	blbMP2HarbingerOracle  = "fcae4f9d-8f81-42f9-93a3-82a482851bdd"
	blbMP2MudflatOracle    = "aeeab1df-0b8b-4bc4-a5f9-aac413449bec"
	blbMP2MuerraOracle     = "c115159c-6991-4812-b3b6-6ed80221cab4"
	blbMP2HermitOracle     = "99007c40-1eb5-4c3a-9889-6d12417d8191"
	blbMP2OakhollowOracle  = "177b7fe4-8565-4631-b4d9-8b2b4282f3ac"
	blbMP2PawpatchOracle   = "50ab0194-c7dc-4f11-abc0-1a82393cc42e"
	blbMP2PlumecreedOracle = "0eb8983e-4e44-4bfe-81dd-ddb41007a166"
)

func TestBlbMP2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		blbMP2MabelOracle:      "Mabel, Heir to Cragflame",
		blbMP2ManifoldOracle:   "Manifold Mouse",
		blbMP2SpiralOracle:     "Mind Spiral",
		blbMP2WhiskerOracle:    "Mindwhisker",
		blbMP2HarbingerOracle:  "Moonstone Harbinger",
		blbMP2MudflatOracle:    "Mudflat Village",
		blbMP2MuerraOracle:     "Muerra, Trash Tactician",
		blbMP2HermitOracle:     "Nightwhorl Hermit",
		blbMP2OakhollowOracle:  "Oakhollow Village",
		blbMP2PawpatchOracle:   "Pawpatch Recruit",
		blbMP2PlumecreedOracle: "Plumecreed Mentor",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// blbMP2Named returns the battlefield permanents called name.
func blbMP2Named(g *game.Game, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

// blbMP2CastHold casts a plain creature card from the active seat's hand
// and lets it resolve. It enters this turn.
func blbMP2CastHold(t *testing.T, g *game.Game, name, typeLine string, keywords ...string) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, Power: 1, Toughness: 1,
		Keywords: keywords, Owner: active.ID, Controller: active.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

// --- Mabel, Heir to Cragflame --------------------------------------------------

func TestBlbMP2MabelPumpsOtherMiceAndMakesCragflame(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	mouse := blbB1Push(g, me.ID, "Mouse", "Creature — Mouse Soldier", "", 1, 1)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirMouse := blbB1Push(g, opp.ID, "Their Mouse", "Creature — Mouse", "", 1, 1)
	mabel := castHoldCreature(t, g, "Mabel, Heir to Cragflame", "Legendary Creature — Mouse Soldier", blbMP2MabelOracle, 3, 3)
	passPriorityAroundTable(t, g)

	if p, tough := ptOf(t, g, mouse); p != 2 || tough != 2 {
		t.Errorf("my other Mouse is %d/%d, want 2/2", p, tough)
	}
	if p, _ := ptOf(t, g, mabel); p != 3 {
		t.Errorf("Mabel's power is %d, want 3 (other Mice only)", p)
	}
	if p, _ := ptOf(t, g, bear); p != 2 {
		t.Errorf("the Bear's power is %d, want 2 (not a Mouse)", p)
	}
	if p, _ := ptOf(t, g, theirMouse); p != 1 {
		t.Errorf("the opponent's Mouse has power %d, want 1 (you control)", p)
	}

	tokens := blbMP2Named(g, "Cragflame")
	if len(tokens) != 1 {
		t.Fatalf("%d Cragflame tokens, want 1", len(tokens))
	}
	crag := tokens[0]
	if !crag.IsToken() || !crag.IsLegendary() || crag.Controller != me.ID {
		t.Errorf("Cragflame: token %v, legendary %v, controller ok %v", crag.IsToken(), crag.IsLegendary(), crag.Controller == me.ID)
	}
	b06AddMana(me, "C", "C")
	b16Activate(t, g, me.ID, crag.InstanceID, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the equipped Bear is %d/%d, want 3/3", p, tough)
	}
	for _, kw := range []string{"vigilance", "trample", "haste"} {
		if !hasKeywordOnBattlefield(t, g, bear, kw) {
			t.Errorf("the equipped Bear should have %s", kw)
		}
	}
}

// --- Manifold Mouse ------------------------------------------------------------

func TestBlbMP2ManifoldMouseGivesTheChosenKeywordAtCombat(t *testing.T) {
	for _, tc := range []struct {
		pick int
		want string
		not  string
	}{
		{0, "double strike", "trample"},
		{1, "trample", "double strike"},
	} {
		t.Run(tc.want, func(t *testing.T) {
			g := newCatalogGame(t)
			me, _, _ := blbB1Seats(g)
			blbB1Push(g, me.ID, "Manifold Mouse", "Creature — Mouse Soldier", blbMP2ManifoldOracle, 1, 2)
			other := blbB1Push(g, me.ID, "Other Mouse", "Creature — Mouse", "", 1, 1)
			bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
			advanceTo(t, g, game.StepBeginCombat)
			ch := pickTargetPrompt(t, g)
			if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
				t.Fatal("a Bear was accepted; the target is a Mouse you control")
			}
			answerPickTarget(t, g, other)
			passPriorityAroundTable(t, g)
			answerOptionPick(t, g, me.ID, tc.pick)
			passPriorityAroundTable(t, g)
			if !hasKeywordOnBattlefield(t, g, other, tc.want) {
				t.Errorf("the Mouse should have %s", tc.want)
			}
			if hasKeywordOnBattlefield(t, g, other, tc.not) {
				t.Errorf("the Mouse should not have %s", tc.not)
			}
		})
	}
}

func TestBlbMP2ManifoldMouseOffspringMakesAOneOneCopy(t *testing.T) {
	g := newCatalogGame(t)
	if _, err := castWithOptionalCosts(t, g, "Manifold Mouse", "Creature — Mouse Soldier", blbMP2ManifoldOracle, nil, []int{0}, nil); err != nil {
		t.Fatalf("cast with offspring: %v", err)
	}
	passPriorityAroundTable(t, g)
	mice := blbMP2Named(g, "Manifold Mouse")
	if len(mice) != 2 {
		t.Fatalf("%d Manifold Mice, want 2 (offspring paid)", len(mice))
	}
	tokens := 0
	for _, c := range mice {
		if c.IsToken() {
			tokens++
			if c.Power != 1 || c.Toughness != 1 {
				t.Errorf("the offspring token is %d/%d, want 1/1", c.Power, c.Toughness)
			}
		}
	}
	if tokens != 1 {
		t.Errorf("%d tokens, want 1", tokens)
	}
}

// --- Mind Spiral ---------------------------------------------------------------

func TestBlbMP2MindSpiralPromisedDrawsAndStuns(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	victim := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	before := me.Hand.Size()
	mustCastWithGift(t, g, "Mind Spiral", "Sorcery", blbMP2SpiralOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: me.ID, Slot: 0}, {Kind: game.TargetCard, ID: victim, Slot: 1},
	}, opp.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+3 {
		t.Errorf("hand %d, want %d", got, before+3)
	}
	if !tappedOnBattlefield(t, g, victim) {
		t.Error("the opponent's creature should be tapped")
	}
	if got := countersOn(g, victim, game.CounterStun); got != 1 {
		t.Errorf("the creature has %d stun counters, want 1", got)
	}
	if n, tapped := countTokensNamed(g, opp.ID, "Fish"); n != 1 || tapped != 1 {
		t.Errorf("the opponent has %d Fish (%d tapped), want one tapped Fish", n, tapped)
	}
}

func TestBlbMP2MindSpiralUnpromisedOnlyDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	victim := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	before := me.Hand.Size()
	mustCastWithGift(t, g, "Mind Spiral", "Sorcery", blbMP2SpiralOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: me.ID, Slot: 0},
	}, uuid.Nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+3 {
		t.Errorf("hand %d, want %d", got, before+3)
	}
	if tappedOnBattlefield(t, g, victim) || countersOn(g, victim, game.CounterStun) != 0 {
		t.Error("no gift: the opponent's creature should be untouched")
	}
	if n, _ := countTokensNamed(g, opp.ID, "Fish"); n != 0 {
		t.Errorf("%d Fish without the gift, want 0", n)
	}
}

// --- Mindwhisker ---------------------------------------------------------------

func TestBlbMP2MindwhiskerThresholdShrinksOpponentsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	blbB1Push(g, me.ID, "Mindwhisker", "Creature — Rat Wizard", blbMP2WhiskerOracle, 3, 2)
	mine := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	blbMPFillGraveyard(g, me, 6)
	if p, _ := ptOf(t, g, theirs); p != 2 {
		t.Fatalf("six cards in the graveyard: their power is %d, want 2", p)
	}
	blbMPFillGraveyard(g, me, 1)
	if p, tough := ptOf(t, g, theirs); p != 1 || tough != 2 {
		t.Errorf("threshold: their Bear is %d/%d, want 1/2", p, tough)
	}
	if p, _ := ptOf(t, g, mine); p != 2 {
		t.Errorf("threshold: my Bear's power is %d, want 2", p)
	}
}

func TestBlbMP2MindwhiskerSurveilsAtYourUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	seedLibrary(me, "Top", "Next")
	blbB1Push(g, me.ID, "Mindwhisker", "Creature — Rat Wizard", blbMP2WhiskerOracle, 3, 2)
	advanceToNextSeatsTurn(t, g)
	advanceToUpkeepOfSeat(t, g, seatIndexOf(g, me.ID))
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil at the upkeep")
	}
	if len(c.ScryCards) != 1 {
		t.Errorf("surveil looked at %d cards, want 1", len(c.ScryCards))
	}
}

// --- Moonstone Harbinger -------------------------------------------------------

func TestBlbMP2MoonstoneHarbingerPumpsBatsOnceOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	harb := blbB1Push(g, me.ID, "Moonstone Harbinger", "Creature — Bat Warrior", blbMP2HarbingerOracle, 1, 3)
	bat := blbB1Push(g, me.ID, "Bat", "Creature — Bat", "", 1, 1)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	advanceToMain(t, g)
	gainLifeNow(t, g, harb, me, 1)
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, bat); p != 2 {
		t.Errorf("the Bat's power is %d, want 2", p)
	}
	if p, _ := ptOf(t, g, harb); p != 2 {
		t.Errorf("the Harbinger's power is %d, want 2 (it is a Bat)", p)
	}
	if !hasKeywordOnBattlefield(t, g, bat, "deathtouch") {
		t.Error("the Bat should have deathtouch")
	}
	if p, _ := ptOf(t, g, bear); p != 2 || hasKeywordOnBattlefield(t, g, bear, "deathtouch") {
		t.Error("the Bear is not a Bat")
	}
	gainLifeNow(t, g, harb, me, -1)
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, bat); p != 2 {
		t.Errorf("only once each turn: the Bat's power is %d, want 2", p)
	}
}

func TestBlbMP2MoonstoneHarbingerIgnoresAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	harb := blbB1Push(g, opp.ID, "Moonstone Harbinger", "Creature — Bat Warrior", blbMP2HarbingerOracle, 1, 3)
	advanceToMain(t, g)
	gainLifeNow(t, g, harb, opp, 2)
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, harb); p != 1 {
		t.Errorf("not its controller's turn: power %d, want 1", p)
	}
}

// --- Mudflat Village and Oakhollow Village ------------------------------------

func TestBlbMP2MudflatVillageReturnsARat(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	village := blbB1Push(g, me.ID, "Mudflat Village", "Land", blbMP2MudflatOracle, 0, 0)
	rat := pushGraveyardCardTyped(me, "Dead Rat", "Creature — Rat Rogue")
	bear := pushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	advanceToMain(t, g)
	b06AddMana(me, "C", "B")
	if err := g.ActivateCatalogAbility(me.ID, village, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err == nil {
		t.Fatal("a Bear card was accepted; the target is a Bat, Lizard, Rat, or Squirrel card")
	}
	b16Activate(t, g, me.ID, village, 0, game.ActivateAbilityParams{Targets: cardRefs(rat)})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(rat) {
		t.Error("the Rat card should be in hand")
	}
	if g.Battlefield.Contains(village) {
		t.Error("the Village should have been sacrificed")
	}
}

func TestBlbMP2OakhollowVillageCountersAnimalsThatEnteredThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	village := blbB1Push(g, me.ID, "Oakhollow Village", "Land", blbMP2OakhollowOracle, 0, 0)
	oldFrog := blbB1Push(g, me.ID, "Old Frog", "Creature — Frog", "", 1, 1)
	newFrog := blbMP2CastHold(t, g, "New Frog", "Creature — Frog")
	newBear := blbMP2CastHold(t, g, "New Bear", "Creature — Bear")
	b06AddMana(me, "G")
	b16Activate(t, g, me.ID, village, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, newFrog); got != 1 {
		t.Errorf("the Frog that entered this turn has %d counters, want 1", got)
	}
	if got := plusOnes(g, oldFrog); got != 0 {
		t.Errorf("the Frog that did not enter this turn has %d counters, want 0", got)
	}
	if got := plusOnes(g, newBear); got != 0 {
		t.Errorf("the Bear has %d counters, want 0", got)
	}
}

// --- Muerra, Trash Tactician ---------------------------------------------------

func TestBlbMP2MuerraAddsManaPerRaccoonAtYourFirstMain(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	blbB1Push(g, me.ID, "Muerra, Trash Tactician", "Legendary Creature — Raccoon Warrior", blbMP2MuerraOracle, 2, 4)
	blbB1Push(g, me.ID, "Raccoon", "Creature — Raccoon", "", 1, 1)
	blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	for _, c := range []string{"R", "G"} {
		pick := riderLatestManaPick(g, me.ID)
		if pick == nil {
			t.Fatalf("missing a colour pick for %s", c)
		}
		if err := g.ResolveManaChoice(pick.ID, me.ID, c); err != nil {
			t.Fatalf("ResolveManaChoice: %v", err)
		}
	}
	if riderLatestManaPick(g, me.ID) != nil {
		t.Error("a third pick: there are two Raccoons")
	}
	if msc12PoolOf(me, "R") != 1 || msc12PoolOf(me, "G") != 1 {
		t.Errorf("pool %v, want one R and one G", poolColors(me))
	}
}

func TestBlbMP2MuerraExpendsFourAndEight(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	ids := seedLibrary(me, "A", "B", "C")
	blbB1Push(g, me.ID, "Muerra, Trash Tactician", "Legendary Creature — Raccoon Warrior", blbMP2MuerraOracle, 2, 4)
	toMainForCost(t, g)
	passPriorityAroundTable(t, g)
	for riderLatestManaPick(g, me.ID) != nil {
		pick := riderLatestManaPick(g, me.ID)
		if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
			t.Fatalf("ResolveManaChoice: %v", err)
		}
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	life := lifeOf(g, me.ID)
	castPaid(t, g, me, "Four", "{3}{R}", "CCCR")
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, me.ID); got != life+3 {
		t.Errorf("expend 4: I am at %d, want %d", got, life+3)
	}
	castPaid(t, g, me, "Four More", "{3}{R}", "CCCR")
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(ids[0]) || !g.Exile.Contains(ids[1]) || g.Exile.Contains(ids[2]) {
		t.Error("expend 8: the top two cards should be exiled, and only those")
	}
	if got := lifeOf(g, me.ID); got != life+3 {
		t.Errorf("expend 4 triggers once: I am at %d, want %d", got, life+3)
	}
}

// --- Nightwhorl Hermit ---------------------------------------------------------

func TestBlbMP2NightwhorlHermitThresholdPumpsAndEvades(t *testing.T) {
	for _, threshold := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp, _ := blbB1Seats(g)
		hermit := blbB1Push(g, me.ID, "Nightwhorl Hermit", "Creature — Rat Rogue", blbMP2HermitOracle, 1, 4)
		wall := blbB1Push(g, opp.ID, "Their Wall", "Creature — Wall", "", 0, 4)
		n := 6
		if threshold {
			n = 7
		}
		blbMPFillGraveyard(g, me, n)
		if !hasKeywordOnBattlefield(t, g, hermit, "vigilance") {
			t.Error("the Hermit should have vigilance")
		}
		want := 1
		if threshold {
			want = 2
		}
		if p, _ := ptOf(t, g, hermit); p != want {
			t.Errorf("threshold %v: power %d, want %d", threshold, p, want)
		}
		advanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(hermit, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		advanceTo(t, g, game.StepDeclareBlockers)
		err := g.DeclareBlocker(wall, hermit)
		if threshold && err == nil {
			t.Error("blocked with threshold")
		}
		if !threshold && err != nil {
			t.Errorf("could not be blocked without threshold: %v", err)
		}
	}
}

// --- Pawpatch Recruit ----------------------------------------------------------

func TestBlbMP2PawpatchRecruitCountersAnotherCreatureWhenAnOpponentTargets(t *testing.T) {
	g := newCatalogGame(t)
	// The opponent is the active seat, so castCatalogSpell casts for them.
	me := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	recruit := blbB1Push(g, me.ID, "Pawpatch Recruit", "Creature — Rabbit Warrior", blbMP2PawpatchOracle, 2, 1)
	if !hasKeywordOnBattlefield(t, g, recruit, "trample") {
		t.Error("the Recruit should have trample")
	}
	big := blbB1Push(g, me.ID, "My Giant", "Creature — Giant", "", 5, 5)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, cardRefs(big))
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: big}); err == nil {
		t.Fatal("the targeted creature was accepted; the counter goes to another creature")
	}
	answerPickTarget(t, g, recruit)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, recruit); got != 1 {
		t.Errorf("the Recruit has %d counters, want 1", got)
	}
}

func TestBlbMP2PawpatchRecruitIgnoresYourOwnSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	recruit := blbB1Push(g, me.ID, "Pawpatch Recruit", "Creature — Rabbit Warrior", blbMP2PawpatchOracle, 2, 1)
	big := blbB1Push(g, me.ID, "My Giant", "Creature — Giant", "", 5, 5)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, cardRefs(big))
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			t.Fatal("my own spell triggered the Recruit")
		}
	}
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, recruit); got != 0 {
		t.Errorf("the Recruit has %d counters, want 0", got)
	}
}

// --- Plumecreed Mentor ---------------------------------------------------------

func TestBlbMP2PlumecreedMentorCountersOnItsAndAnotherFliersEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	mentor := castHoldCreature(t, g, "Plumecreed Mentor", "Creature — Bird Scout", blbMP2PlumecreedOracle, 2, 3)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: mentor}); err == nil {
		t.Fatal("the Mentor (flying) was accepted; the target is a creature without flying")
	}
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, bear); got != 1 {
		t.Fatalf("its own entry: the Bear has %d counters, want 1", got)
	}

	blbMP2CastHold(t, g, "Walker", "Creature — Bear")
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			t.Fatal("a creature without flying entering triggered the Mentor")
		}
	}
	blbMP2CastHold(t, g, "Flier", "Creature — Bird", "flying")
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, bear); got != 2 {
		t.Errorf("another flier entered: the Bear has %d counters, want 2", got)
	}
}
