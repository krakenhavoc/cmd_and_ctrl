package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// znr_old_b090_test.go — Zendikar Rising kicker cards, S82 OLD batch 90
// (#5508). Every test plays the real catalog entry through the engine.

const (
	oldB090RoostOracle       = "05e73886-8311-4a0d-9c8d-4a03c7f20ea2"
	oldB090InscriptionOracle = "1551fe97-437e-4e89-a0d2-c398cb2155e4"
	oldB090ChargerOracle     = "571f9d80-9c37-474e-9524-400e7a238e2c"
	oldB090ArbormageOracle   = "58254569-0ba6-4ac1-9403-dc21a1b5c2e8"
	oldB090StormcallerOracle = "675de007-a71e-4036-8204-140c67871b7f"
	oldB090ShellShieldOracle = "73535033-2875-4a82-98d5-84548e349f13"
	oldB090ParagonOracle     = "b03a9d03-4853-4856-8cfa-ad4085e0e3e3"
	oldB090BubbleSnareOracle = "be315fa6-afb5-4ab6-a143-7bbf37046d15"
	oldB090ScourgeOracle     = "c6dfc91f-bfac-40c2-99e5-0cec981f7f1f"
	oldB090ReclaimOracle     = "ce6f3c67-8806-416d-9156-6f5bb63d5f4c"
	oldB090SproutlingOracle  = "dfa8c557-f2f5-4763-85c7-e2ec14d9e6b6"
	oldB090MossPitOracle     = "edf0e7c0-d9a1-4d29-9cee-5267d3a3978f"
	oldB090GloomhunterOracle = "70ab65ff-724f-4126-b472-97fa03ca3a34"
	oldB090BearType          = "Creature — Bear"
	oldB090ForestType        = "Basic Land — Forest"
)

func TestOldB090EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB090RoostOracle: "Roost of Drakes", oldB090InscriptionOracle: "Inscription of Insight",
		oldB090ChargerOracle: "Shatterskull Charger", oldB090ArbormageOracle: "Taunting Arbormage",
		oldB090StormcallerOracle: "Sea Gate Stormcaller", oldB090ShellShieldOracle: "Shell Shield",
		oldB090ParagonOracle: "Tajuru Paragon", oldB090BubbleSnareOracle: "Bubble Snare",
		oldB090ScourgeOracle: "Scourge of the Skyclaves", oldB090ReclaimOracle: "Reclaim the Wastes",
		oldB090SproutlingOracle: "Murasa Sproutling", oldB090MossPitOracle: "Moss-Pit Skeleton",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB090Cast puts a card with the given real cost into the active player's
// hand, loads exactly the mana the cast needs and casts it, kicked or not.
func oldB090Cast(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, colors []string, mana string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.HandCardFull(me, name, typeLine, cost, oracle, colors)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
		t.Fatalf("mana: %v", err)
	}
	params.Strict = true
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func oldB090Kicked() game.CastSpellParams { return game.CastSpellParams{OptionalCosts: []int{0}} }

func oldB090Drakes(g *game.Game, owner uuid.UUID) int {
	return effectstest.CountBattlefieldNamed(g, owner, "Drake")
}

func TestOldB090RoostOfDrakesMakesADrakeOnlyWhenKickedAndWheneverYouCastAKickedSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB090Cast(t, g, "Roost of Drakes", "Enchantment", "{U}", oldB090RoostOracle, []string{"U"}, "{U}", game.CastSpellParams{})
	if got := oldB090Drakes(g, me.ID); got != 0 {
		t.Fatalf("%d Drakes from an unkicked Roost, want 0", got)
	}
	// A kicked spell with the Roost in play makes a Drake.
	oldB090Cast(t, g, "Ghastly Gloomhunter", "Creature — Zombie Bat", "{1}{B}", oldB090GloomhunterOracle, []string{"B"},
		"{B}{B}{B}{B}{B}{B}{B}", oldB090Kicked())
	if got := oldB090Drakes(g, me.ID); got != 1 {
		t.Errorf("%d Drakes after a kicked spell, want 1", got)
	}
	// An unkicked one does not.
	oldB090Cast(t, g, "Ghastly Gloomhunter", "Creature — Zombie Bat", "{1}{B}", oldB090GloomhunterOracle, []string{"B"},
		"{B}{B}", game.CastSpellParams{})
	if got := oldB090Drakes(g, me.ID); got != 1 {
		t.Errorf("%d Drakes after an unkicked spell, want still 1", got)
	}
}

func TestOldB090RoostOfDrakesKickedMakesADrakeWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB090Cast(t, g, "Roost of Drakes", "Enchantment", "{U}", oldB090RoostOracle, []string{"U"}, "{U}{U}{U}{U}", oldB090Kicked())
	if got := oldB090Drakes(g, me.ID); got != 1 {
		t.Errorf("%d Drakes from a kicked Roost, want 1", got)
	}
}

func oldB090InscriptionTargets(opp *game.Player, a, b uuid.UUID, modes []int) []game.TargetRef {
	var out []game.TargetRef
	for occ, m := range modes {
		switch m {
		case 0:
			out = append(out,
				game.TargetRef{Kind: game.TargetCard, ID: a, Mode: occ},
				game.TargetRef{Kind: game.TargetCard, ID: b, Mode: occ})
		case 2:
			out = append(out, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID, Mode: occ})
		}
	}
	return out
}

func TestOldB090InscriptionOfInsightUnkickedChoosesOneKickedAnyNumber(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	b := effectstest.B12Creature(g, opp.ID, "Troll", "Creature — Troll", 3, 3)
	for i := 0; i < 3; i++ {
		opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	}
	handSize := opp.Hand.Size()
	cast := func(modes []int, kicked bool) error {
		id := effectstest.HandCardFull(me, "Inscription of Insight", "Sorcery", "{3}{U}", oldB090InscriptionOracle, []string{"U"})
		mana, optional := "{U}{U}{U}{U}", []int(nil)
		if kicked {
			mana, optional = "{U}{U}{U}{U}{U}{U}{U}{U}", []int{0}
		}
		me.ManaPool.EmptyPool()
		if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
			return err
		}
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: modes, OptionalCosts: optional,
			Targets: oldB090InscriptionTargets(opp, a, b, modes)})
		if err != nil {
			_, _ = me.Hand.Remove(id)
			me.ManaPool.EmptyPool()
		}
		return err
	}
	if err := cast([]int{0, 2}, false); err == nil {
		t.Fatal("unkicked: two bullets must be refused")
	}
	if err := cast([]int{2}, false); err != nil {
		t.Fatalf("unkicked, the token bullet: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, opp.ID, "Illusion"); got != 1 {
		t.Fatalf("%d Illusions, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Illusion" && (c.Power != handSize || c.Toughness != handSize) {
			t.Errorf("Illusion is %d/%d, want %d/%d (the cards in their hand)", c.Power, c.Toughness, handSize, handSize)
		}
	}
	handSize += 2 // the bounce below puts two cards into the hand before the Illusion counts them
	if err := cast([]int{0, 2}, true); err != nil {
		t.Fatalf("kicked, bounce and token: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("the creatures were not returned")
	}
	if !opp.Hand.Contains(a) || !opp.Hand.Contains(b) {
		t.Error("the creatures are not in their owner's hand")
	}
	// The Illusion is counted after the bounce put two cards into the hand.
	big := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Illusion" && c.Power == handSize {
			big++
		}
	}
	if big != 1 {
		t.Errorf("%d 5/5 Illusions, want 1 (X counts the returned creatures)", big)
	}
}

func TestOldB090ShatterskullChargerReturnsUnlessKicked(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		params, mana := game.CastSpellParams{}, "{R}{R}{R}"
		if kicked {
			params, mana = oldB090Kicked(), "{R}{R}{R}{R}{R}"
		}
		id := oldB090Cast(t, g, "Shatterskull Charger", "Creature — Giant Warrior", "{1}{R}{R}", oldB090ChargerOracle, []string{"R"}, mana, params)
		if got := effectstest.CountersOn(g, id, game.CounterPlusOne); (got == 1) != kicked {
			t.Fatalf("kicked=%v: %d counters", kicked, got)
		}
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		if on := effectstest.OnBattlefieldByID(g, id); on != kicked {
			t.Errorf("kicked=%v: on the battlefield = %v at the end step", kicked, on)
		}
		if !kicked && !me.Hand.Contains(id) {
			t.Error("the unkicked Charger is not in its owner's hand")
		}
	}
}

func TestOldB090TauntingArbormageLuresOnlyWhenKicked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.ReqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	a := effectstest.ReqCreature(g, opp.ID, "Wall A", "", 0, 4)
	b := effectstest.ReqCreature(g, opp.ID, "Wall B", "", 0, 4)
	oldB090Cast(t, g, "Taunting Arbormage", "Creature — Elf Wizard", "{2}{G}", oldB090ArbormageOracle, []string{"G"},
		"{G}{G}{G}{G}{G}{G}", oldB090Kicked())
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.OnToBlockers(t, g)
	must := effectstest.MustBlockNow(g)
	if must[a] != bear || must[b] != bear {
		t.Errorf("must-block map %v: every able blocker should have to block the Bear", must)
	}
	if err := effectstest.ReqBlock(t, g, bear, a, b); err != nil {
		t.Errorf("both blockers refused: %v", err)
	}
}

func TestOldB090SeaGateStormcallerCopiesTheNextCheapSpell(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		params, mana := game.CastSpellParams{}, "{U}{U}"
		if kicked {
			params, mana = oldB090Kicked(), "{U}{U}{U}{U}{U}{U}{U}"
		}
		oldB090Cast(t, g, "Sea Gate Stormcaller", "Creature — Human Wizard", "{1}{U}", oldB090StormcallerOracle, []string{"U"}, mana, params)
		// An expensive spell is not the next one it wants.
		oldB090Library := func() { effectstest.Msc18Library(me, 10) }
		oldB090Library()
		big := effectstest.HandCardFull(me, "Big Spell", "Sorcery", "{3}", "", nil)
		_ = big
		bear := effectstest.B12Creature(g, me.ID, "Bear", oldB090BearType, 2, 2)
		id := effectstest.HandCardFull(me, "Shell Shield", "Instant", "{U}", oldB090ShellShieldOracle, []string{"U"})
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}"); err != nil {
			t.Fatal(err)
		}
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err != nil {
			t.Fatalf("cast the cheap spell: %v", err)
		}
		g.RunStateChecksForTest()
		effectstest.PassPriorityAroundTable(t, g)
		// Each copy may choose new targets: keep the target.
		for i := 0; i < 16; i++ {
			if effectstest.LatestPickTarget(g, me.ID) != nil {
				effectstest.AnswerPickTarget(t, g, bear)
				continue
			}
			if effectstest.StackFullyEmpty(g) {
				break
			}
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
		}
		want := 6 // the spell and one copy, +0/+3 each
		if kicked {
			want = 9 // the spell and two copies
		}
		if p, tough := effectstest.PtOf(t, g, bear); tough != 2+want {
			t.Errorf("kicked=%v: Bear is %d/%d, want toughness %d", kicked, p, tough, 2+want)
		}
	}
}

func TestOldB090ShellShieldGivesToughnessAndHexproofWhenKicked(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		bear := effectstest.B12Creature(g, me.ID, "Bear", oldB090BearType, 2, 2)
		params, mana := game.CastSpellParams{Targets: effectstest.CardTarget(bear)}, "{U}"
		if kicked {
			params.OptionalCosts, mana = []int{0}, "{U}{U}"
		}
		oldB090Cast(t, g, "Shell Shield", "Instant", "{U}", oldB090ShellShieldOracle, []string{"U"}, mana, params)
		if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 5 {
			t.Errorf("kicked=%v: %d/%d, want 2/5", kicked, p, tough)
		}
		if got := effectstest.HasEffectiveKeyword(t, g, bear, "hexproof"); got != kicked {
			t.Errorf("kicked=%v: hexproof = %v", kicked, got)
		}
	}
}

func TestOldB090TajuruParagonIsAPartyOfFourAndFindsAChosenTypeWhenKicked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	paragon := effectstest.B12Push(g, me.ID, "Tajuru Paragon", "Creature — Elf", oldB090ParagonOracle, 3, 2)
	var party int
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked(); party = effects.FinB9PartySize(g, me.ID) })
	if party != 1 {
		t.Errorf("a lone Paragon is a party of %d, want 1", party)
	}
	_ = paragon
	effectstest.Msc18Library(me, 12)
	wizard, bear := uuid.New(), uuid.New()
	me.Library.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: oldB090BearType, Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: wizard, Name: "Sage", TypeLine: "Creature — Human Wizard", Owner: me.ID, Controller: me.ID})
	oldB090Cast(t, g, "Tajuru Paragon", "Creature — Elf", "{1}{G}", oldB090ParagonOracle, []string{"G"}, "{G}{G}{G}{G}{G}", oldB090Kicked())
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no choice for the revealed cards")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range c.ChooseCards {
		offered[id] = true
	}
	if !offered[wizard] {
		t.Error("a Wizard card sharing a type with the Paragon was not offered")
	}
	if offered[bear] {
		t.Error("a Bear, sharing no type, was offered")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, wizard)
	if !me.Hand.Contains(wizard) {
		t.Error("the Wizard did not reach the hand")
	}
}

func TestOldB090BubbleSnareTapsWhenKickedAndKeepsItTapped(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB090BearType, 2, 2)
		params, mana := game.CastSpellParams{Targets: effectstest.CardTarget(bear)}, "{U}"
		if kicked {
			params.OptionalCosts, mana = []int{0}, "{U}{U}{U}{U}"
		}
		oldB090Cast(t, g, "Bubble Snare", "Enchantment — Aura", "{U}", oldB090BubbleSnareOracle, []string{"U"}, mana, params)
		if got := effectstest.HmlTapped(t, g, bear); got != kicked {
			t.Errorf("kicked=%v: tapped = %v", kicked, got)
		}
	}
}

func TestOldB090ScourgeOfTheSkyclavesIsTwentyMinusTheHighestLifeAndHalvesLifeWhenKicked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, p := range g.Seats {
		p.Life = 10
	}
	g.Seats[1].Life = 12
	scourge := effectstest.B12Push(g, me.ID, "Scourge of the Skyclaves", "Creature — Demon", oldB090ScourgeOracle, 0, 0)
	if p, tough := effectstest.PtOf(t, g, scourge); p != 8 || tough != 8 {
		t.Errorf("%d/%d with a highest life of 12, want 8/8", p, tough)
	}
	oldB090Cast(t, g, "Scourge of the Skyclaves", "Creature — Demon", "{1}{B}", oldB090ScourgeOracle, []string{"B"},
		"{B}{B}{B}{B}{B}{B}{B}", oldB090Kicked())
	for i, p := range g.Seats {
		want := 5
		if i == 1 {
			want = 6
		}
		if p.Life != want {
			t.Errorf("seat %d life %d, want %d after each player lost half rounded up", i, p.Life, want)
		}
	}
}

func TestOldB090ScourgeUnkickedDoesNotHalveLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	for _, p := range g.Seats {
		p.Life = 30
	}
	oldB090Cast(t, g, "Scourge of the Skyclaves", "Creature — Demon", "{1}{B}", oldB090ScourgeOracle, []string{"B"},
		"{B}{B}", game.CastSpellParams{})
	for i, p := range g.Seats {
		if p.Life != 30 {
			t.Errorf("seat %d life %d, want 30", i, p.Life)
		}
	}
}

func TestOldB090ReclaimTheWastesFetchesOneBasicOrTwoWhenKicked(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.SeedSearchLibrary(me,
			effectstest.SearchTestLand("Forest", oldB090ForestType),
			effectstest.SearchTestLand("Island", "Basic Land — Island"),
			game.Card{Name: "Bear", TypeLine: oldB090BearType})
		params, mana := game.CastSpellParams{}, "{G}"
		if kicked {
			params, mana = oldB090Kicked(), "{G}{G}{G}{G}"
		}
		hand := me.Hand.Size()
		oldB090Cast(t, g, "Reclaim the Wastes", "Sorcery", "{G}", oldB090ReclaimOracle, []string{"G"}, mana, params)
		if effectstest.SearchChoiceFor(g, me.ID) == nil {
			t.Fatalf("kicked=%v: no search prompt", kicked)
		}
		if kicked {
			effectstest.AnswerSearchNamed(t, g, me.ID, "Forest", "Island")
		} else {
			effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
		}
		effectstest.PassPriorityAroundTable(t, g)
		want := hand + 1 // the spell was added and left, then the lands
		if kicked {
			want = hand + 2
		}
		if me.Hand.Size() != want {
			t.Errorf("kicked=%v: hand %d, want %d", kicked, me.Hand.Size(), want)
		}
	}
}

func TestOldB090MurasaSproutlingReturnsAKickerCardWhenKicked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kick, plain := uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: plain, Name: "Plain Bear", TypeLine: oldB090BearType, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: kick, Name: "Shell Shield", TypeLine: "Instant", OracleID: oldB090ShellShieldOracle, Owner: me.ID, Controller: me.ID})
	oldB090Cast(t, g, "Murasa Sproutling", "Creature — Plant Elemental", "{2}{G}", oldB090SproutlingOracle, []string{"G"},
		"{G}{G}{G}{G}{G}", oldB090Kicked())
	effectstest.AnswerPickTarget(t, g, kick)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(kick) {
		t.Error("the kicker card did not return")
	}
	if me.Hand.Contains(plain) {
		t.Error("a card with no kicker returned")
	}
}

func TestOldB090MossPitSkeletonEntersWithThreeCountersAndComesBackFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := oldB090Cast(t, g, "Moss-Pit Skeleton", "Creature — Plant Skeleton", "{B}{G}", oldB090MossPitOracle, []string{"B", "G"},
		"{B}{G}{G}{G}{G}", oldB090Kicked())
	if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != 3 {
		t.Errorf("%d counters, want 3", got)
	}
	// Dead in the graveyard, it returns to the top when a creature you control gets a counter.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g2)
	grave := uuid.New()
	me2.Graveyard.PushTop(game.Card{InstanceID: grave, Name: "Moss-Pit Skeleton", TypeLine: "Creature — Plant Skeleton",
		OracleID: oldB090MossPitOracle, Owner: me2.ID, Controller: me2.ID})
	bear := effectstest.B12Creature(g2, me2.ID, "Bear", oldB090BearType, 2, 2)
	effectstest.PutCounters(t, g2, bear, game.CounterPlusOne, 1)
	g2.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AnswerConfirmFor(t, g2, me2, true)
	effectstest.PassPriorityAroundTable(t, g2)
	if me2.Library.Size() == 0 || me2.Library.Cards[len(me2.Library.Cards)-1].InstanceID != grave {
		t.Error("the Skeleton is not on top of its owner's library")
	}
	_ = me
}
