package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	rprB28DrogskolCaptainOracle  = "c33604ac-636a-43ad-9efe-8d7b3a28bf81"
	rprB28FalkenrathGorgerOracle = "26aba6fa-830f-4a7f-8618-488d8e00215a"
	rprB28FieldOfSoulsOracle     = "4d7a5b14-8fce-41f2-a0d5-fff3d15f41f6"
	rprB28GeistOracle            = "4304035c-c3e4-4a19-aa1f-92a83d8aed1f"
	rprB28SpiritkeeperOracle     = "5852ed2f-4332-4c30-b145-373deea471d9"
	rprB28KamiOracle             = "38cfe4a0-6ead-435b-9f86-928205c0a52c"
	rprB28KirtarsWrathOracle     = "4f66d82a-492f-4638-9f77-190d4a33ad7f"
	rprB28BloodwitchOracle       = "870e199e-1d54-44ec-8e24-f24af93ae9c5"
	rprB28MobRuleOracle          = "acca2128-6ebb-48a5-a6a2-1bc5e4aa86ce"
	rprB28MoltenEchoesOracle     = "26486a42-4dc6-472b-9a9b-3c3ca78cc3b5"
	rprB28HeraldOracle           = "cf336e4c-a0d3-43aa-ad23-b98effb2b751"
)

func rprB28Tokens(g *game.Game, controller uuid.UUID, name string) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == controller && IsToken(c) {
				out = append(out, c)
			}
		}
	})
	return out
}

func TestRprB28CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		rprB28DrogskolCaptainOracle:  "Drogskol Captain",
		rprB28FalkenrathGorgerOracle: "Falkenrath Gorger",
		rprB28FieldOfSoulsOracle:     "Field of Souls",
		rprB28GeistOracle:            "Geist of Saint Traft",
		rprB28SpiritkeeperOracle:     "Hallowed Spiritkeeper",
		rprB28KamiOracle:             "Kami of the Crescent Moon",
		rprB28KirtarsWrathOracle:     "Kirtar's Wrath",
		rprB28BloodwitchOracle:       "Malakir Bloodwitch",
		rprB28MobRuleOracle:          "Mob Rule",
		rprB28MoltenEchoesOracle:     "Molten Echoes",
		rprB28HeraldOracle:           "Nebelgast Herald",
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

func TestRprB28DrogskolCaptainLordsOtherSpiritsYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	captain := b12Push(g, me.ID, "Drogskol Captain", "Creature — Spirit Soldier", rprB28DrogskolCaptainOracle, 2, 2)
	mine := b12Creature(g, me.ID, "My Spirit", "Creature — Spirit", 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Spirit", "Creature — Spirit", 1, 1)
	if p, tough := ptOf(t, g, mine); p != 2 || tough != 2 {
		t.Errorf("my other Spirit is %d/%d, want 2/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, mine, "hexproof") {
		t.Error("my other Spirit lacks hexproof")
	}
	if p, tough := ptOf(t, g, captain); p != 2 || tough != 2 || hasEffectiveKeyword(t, g, captain, "hexproof") {
		t.Errorf("the Captain itself is %d/%d, hexproof %v; want a plain 2/2", p, tough, hasEffectiveKeyword(t, g, captain, "hexproof"))
	}
	if !hasEffectiveKeyword(t, g, captain, "flying") {
		t.Error("the Captain lost flying")
	}
	if p, _ := ptOf(t, g, bear); p != 1 || hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("a non-Spirit was lorded")
	}
	if p, _ := ptOf(t, g, theirs); p != 1 || hasEffectiveKeyword(t, g, theirs, "hexproof") {
		t.Error("an opponent's Spirit was lorded (it says you control)")
	}
}

func TestRprB28KamiGivesEachPlayerAnExtraCardInTheirDrawStep(t *testing.T) {
	g := newCatalogGame(t)
	b12Push(g, g.Seats[0].ID, "Kami of the Crescent Moon", "Legendary Creature — Spirit", rprB28KamiOracle, 1, 3)
	for seat := 1; seat < 4; seat++ {
		p := g.Seats[seat]
		hand := p.Hand.Size()
		batch01AdvanceToStepOf(t, g, seat, game.StepDraw)
		passPriorityAroundTable(t, g)
		if p.Hand.Size() != hand+2 {
			t.Errorf("seat %d drew %d, want 2", seat, p.Hand.Size()-hand)
		}
	}
	me := g.Seats[0]
	hand := me.Hand.Size()
	batch01AdvanceToStepOf(t, g, 0, game.StepDraw)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("the controller drew %d, want 2", me.Hand.Size()-hand)
	}
}

func TestRprB28FieldOfSoulsMakesASpiritForYourNontokenCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Field of Souls", "Enchantment", rprB28FieldOfSoulsOracle, 0, 0)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	killCreature(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	if n := len(rprB28Tokens(g, me.ID, "Spirit")); n != 1 {
		t.Fatalf("%d Spirits after my creature died, want 1", n)
	}
	sp := rprB28Tokens(g, me.ID, "Spirit")[0]
	if sp.Power != 1 || sp.Toughness != 1 || !hasEffectiveKeyword(t, g, sp.InstanceID, "flying") {
		t.Errorf("Spirit = %d/%d, want a 1/1 flier", sp.Power, sp.Toughness)
	}
	// A token dying is not a nontoken creature.
	killCreature(t, g, me.ID, sp.InstanceID)
	passPriorityAroundTable(t, g)
	if n := len(rprB28Tokens(g, me.ID, "Spirit")); n != 0 {
		t.Errorf("a token's death made %d Spirits", n)
	}
	// An opponent's creature goes to their graveyard, not mine.
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	killCreature(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if n := len(rprB28Tokens(g, me.ID, "Spirit")); n != 0 {
		t.Errorf("an opponent's creature dying made %d Spirits", n)
	}
	// My creature an opponent controls goes to MY graveyard.
	stolen := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stolen Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: opp.ID,
	})
	killCreature(t, g, me.ID, stolen)
	passPriorityAroundTable(t, g)
	if n := len(rprB28Tokens(g, me.ID, "Spirit")); n != 1 {
		t.Errorf("my stolen creature dying made %d Spirits, want 1", n)
	}
}

func TestRprB28NebelgastHeraldTapsOnItsOwnEntryAndOnOtherSpirits(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	herald := castCatalogSpell(t, g, "Nebelgast Herald", "Creature — Spirit", rprB28HeraldOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(herald) {
		t.Fatal("the Herald did not resolve")
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the Herald's own entry")
	}
	if hasID(p.PickTargetCards, mine) || !hasID(p.PickTargetCards, theirs) {
		t.Errorf("legal targets %v: only a creature an opponent controls", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if c := findBattlefieldCardByID(g, theirs); c == nil || !c.Tapped {
		t.Fatal("their Bear should be tapped")
	}
	// Another Spirit entering under my control triggers it again.
	other := b12Creature(g, opp.ID, "Other Bear", "Creature — Bear", 2, 2)
	b12PlayFromHand(t, g, "Wisp", "Creature — Spirit", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) == nil {
		t.Fatal("another Spirit entering did not trigger the Herald")
	}
	pickCard(t, g, me.ID, other)
	passPriorityAroundTable(t, g)
	if c := findBattlefieldCardByID(g, other); c == nil || !c.Tapped {
		t.Error("the second target should be tapped")
	}
	// A non-Spirit does not.
	b12PlayFromHand(t, g, "Plain Bear", "Creature — Bear", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a non-Spirit entering triggered the Herald")
	}
}

func TestRprB28MalakirBloodwitchDrainsByVampireCount(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "Vamp A", "Creature — Vampire", 1, 1)
	b12Creature(g, me.ID, "Vamp B", "Creature — Vampire", 1, 1)
	b12Creature(g, g.Seats[1].ID, "Their Vamp", "Creature — Vampire", 1, 1)
	before := b17Life(g)
	witch := castCatalogSpell(t, g, "Malakir Bloodwitch", "Creature — Vampire Shaman", rprB28BloodwitchOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(witch) {
		t.Fatal("no Bloodwitch")
	}
	after := b17Life(g)
	for i := 1; i < 4; i++ {
		if before[i]-after[i] != 3 {
			t.Errorf("seat %d lost %d, want 3 (the Bloodwitch and two other Vampires of mine)", i, before[i]-after[i])
		}
	}
	if after[0]-before[0] != 9 {
		t.Errorf("I gained %d, want 9", after[0]-before[0])
	}
	if !hasEffectiveKeyword(t, g, witch, "flying") || !hasEffectiveKeyword(t, g, witch, game.ProtectionFromColor("W")) {
		t.Error("the Bloodwitch lacks flying or protection from white")
	}
}

func TestRprB28HallowedSpiritkeeperCountsItselfAmongTheCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	keeper := b12Push(g, me.ID, "Hallowed Spiritkeeper", "Creature — Avatar", rprB28SpiritkeeperOracle, 3, 2)
	for i := 0; i < 2; i++ {
		b17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	}
	b17GraveyardCard(me, "Dead Spell", "Sorcery", "{1}{R}")
	killCreature(t, g, me.ID, keeper)
	passPriorityAroundTable(t, g)
	if n := len(rprB28Tokens(g, me.ID, "Spirit")); n != 3 {
		t.Errorf("%d Spirits, want 3 (two Bears and the Spiritkeeper itself)", n)
	}
}

func TestRprB28GeistMakesAnAttackingAngelAndExilesItAtEndOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	geist := b12Push(g, me.ID, "Geist of Saint Traft", "Legendary Creature — Spirit Cleric", rprB28GeistOracle, 2, 2)
	if !hasEffectiveKeyword(t, g, geist, "hexproof") {
		t.Error("the Geist lacks hexproof")
	}
	declareAttack(t, g, opp.ID, geist)
	passPriorityAroundTable(t, g)
	passAndSettleAttackSplit(t, g, me.ID)
	toks := rprB28Tokens(g, me.ID, "Angel")
	if len(toks) != 1 {
		t.Fatalf("%d Angels, want 1", len(toks))
	}
	c := toks[0]
	if !c.Tapped || c.AttackingTarget == uuid.Nil || c.Power != 4 || c.Toughness != 4 || !hasEffectiveKeyword(t, g, c.InstanceID, "flying") {
		t.Errorf("Angel = %+v, want a tapped attacking 4/4 flier", c)
	}
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(c.InstanceID) {
		t.Error("the Angel survived the end of combat")
	}
	if !g.Battlefield.Contains(geist) {
		t.Error("the Geist was exiled")
	}
}

func rprB28FillGraveyard(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		b17GraveyardCard(p, "Filler", "Sorcery", "{R}")
	}
}

func TestRprB28KirtarsWrathDestroysAllAndOnlyMakesSpiritsAtThreshold(t *testing.T) {
	for _, tc := range []struct {
		name   string
		filler int
		spirit int
	}{{"six cards", 5, 0}, {"seven cards", 7, 2}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			a := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
			b := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
			rprB28FillGraveyard(me, tc.filler)
			castCatalogSpell(t, g, "Kirtar's Wrath", "Sorcery", rprB28KirtarsWrathOracle, nil)
			passPriorityAroundTable(t, g)
			if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
				t.Error("a creature survived the wrath")
			}
			toks := rprB28Tokens(g, me.ID, "Spirit")
			if len(toks) != tc.spirit {
				t.Fatalf("%d Spirits, want %d", len(toks), tc.spirit)
			}
			for _, c := range toks {
				if c.Power != 1 || c.Toughness != 1 || !hasEffectiveKeyword(t, g, c.InstanceID, "flying") {
					t.Errorf("Spirit = %d/%d, want a 1/1 flier", c.Power, c.Toughness)
				}
			}
		})
	}
}

func TestRprB28MobRuleTakesTheBigOrTheSmallCreatures(t *testing.T) {
	for _, tc := range []struct {
		mode        int
		bigStolen   bool
		smallStolen bool
	}{{0, true, false}, {1, false, true}} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		big := b12Creature(g, opp.ID, "Big", "Creature — Giant", 4, 4)
		small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
		assistCardTapped(t, g, big)
		assistCardTapped(t, g, small)
		castCatalogSpellWithModes(t, g, "Mob Rule", "Sorcery", rprB28MobRuleOracle, []int{tc.mode})
		passPriorityAroundTable(t, g)
		for _, e := range []struct {
			id     uuid.UUID
			stolen bool
			name   string
		}{{big, tc.bigStolen, "power 4"}, {small, tc.smallStolen, "power 3"}} {
			c := findBattlefieldCardByID(g, e.id)
			if c == nil {
				t.Fatalf("mode %d: %s creature left", tc.mode, e.name)
			}
			if got := c.Controller == me.ID; got != e.stolen {
				t.Errorf("mode %d: %s creature taken = %v, want %v", tc.mode, e.name, got, e.stolen)
			}
			if e.stolen && (c.Tapped || !hasEffectiveKeyword(t, g, e.id, "haste")) {
				t.Errorf("mode %d: %s creature should be untapped with haste", tc.mode, e.name)
			}
			if !e.stolen && !c.Tapped {
				t.Errorf("mode %d: %s creature should stay tapped", tc.mode, e.name)
			}
		}
	}
}

func TestRprB28MoltenEchoesCopiesNontokenCreaturesOfTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castAndPause(t, g, "Molten Echoes", "Enchantment", rprB28MoltenEchoesOracle)
	answerCreatureType(t, g, me.ID, "Goblin")
	passPriorityAroundTable(t, g)
	b12PlayFromHand(t, g, "Goblin Guide", "Creature — Goblin Scout", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	toks := rprB28Tokens(g, me.ID, "Goblin Guide")
	if len(toks) != 1 {
		t.Fatalf("%d Goblin Guide tokens, want 1", len(toks))
	}
	if !hasEffectiveKeyword(t, g, toks[0].InstanceID, "haste") {
		t.Error("the copy lacks haste")
	}
	// The token entering is not a nontoken creature: no loop.
	if n := len(rprB28Tokens(g, me.ID, "Goblin Guide")); n != 1 {
		t.Errorf("%d tokens after settling, want 1", n)
	}
	b12PlayFromHand(t, g, "Plain Bear", "Creature — Bear", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := len(rprB28Tokens(g, me.ID, "Plain Bear")); n != 0 {
		t.Errorf("a creature of another type was copied (%d)", n)
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if n := len(rprB28Tokens(g, me.ID, "Goblin Guide")); n != 0 {
		t.Errorf("%d copies survive the end step, want 0", n)
	}
}

func TestRprB28FalkenrathGorgerGrantsMadnessToDiscardedVampires(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Falkenrath Gorger", "Creature — Vampire Berserker", rprB28FalkenrathGorgerOracle, 2, 1)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	vamp := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: vamp, Name: "Test Vampire", TypeLine: "Creature — Vampire", ManaCost: "{1}{B}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		KnownBy: map[uuid.UUID]bool{me.ID: true},
	})
	bear := pushCatalogHandCard(me, "Test Bear", "Creature — Bear", "")
	castCatalogSpell(t, g, "Faithless Looting", "Sorcery", faithlessLootingOracl, nil)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, vamp, bear)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(vamp) || me.Graveyard.Contains(vamp) {
		t.Fatal("the discarded Vampire is not in exile (CR 702.35a)")
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the discarded Bear should go to the graveyard: the Gorger grants madness to Vampires only")
	}
	if latestChoiceOfKind(g, game.PendingChoiceResolutionCast) == nil {
		t.Fatal("no madness cast was offered")
	}
	dcMana(t, g, me.ID, "{C}{B}")
	if err := g.CastSpell(me.ID, vamp, game.CastSpellParams{
		Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyMadness,
	}); err != nil {
		t.Fatalf("the madness cast for its mana cost: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(vamp) {
		t.Error("the madness-cast Vampire is not on the battlefield")
	}
}

func TestRprB28FalkenrathGorgerDeclinedMadnessGoesToTheGraveyardAndNothingWithoutIt(t *testing.T) {
	// Without the Gorger a discarded Vampire goes to the graveyard.
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	vamp := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: vamp, Name: "Test Vampire", TypeLine: "Creature — Vampire", ManaCost: "{1}{B}",
		Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
	})
	spare := pushCatalogHandCard(me, "Spare", "Sorcery", "")
	castCatalogSpell(t, g, "Faithless Looting", "Sorcery", faithlessLootingOracl, nil)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, vamp, spare)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(vamp) || g.Exile.Contains(vamp) {
		t.Fatal("with no Gorger the Vampire should simply be in the graveyard")
	}
}
