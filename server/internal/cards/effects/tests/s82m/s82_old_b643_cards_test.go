package testss82m

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b643_cards_test.go — Planeshift's enters triggers, familiars and land-type
// cards, batch 643 (#5615).

const (
	b643Crocodile = "1d6f3934-d8f1-4758-b87a-8df8d7dc492c"
	b643Snidd     = "bbcd2045-cf51-4346-a561-ebae8ffdc181"
	b643Emergence = "ce19d2dd-4b2c-493e-8963-ad97649248e2"
	b643Drake     = "d5978943-52fc-4b57-9c1c-d802ade1b014"
	b643Spider    = "28537019-3061-4efc-8e87-9bb805bf520f"
	b643Thunder   = "40f6ea6f-71e1-46bd-8d9a-ccb5f437892c"
	b643Sunscape  = "615603bb-3899-4419-9444-e4b64c94a438"
	b643Storm     = "a8f5a11f-3291-4ee9-a96d-4fa38af14932"
	b643Judgment  = "41998d4d-ca01-4a9c-8bd6-399e4c3e25b1"
	b643Overlay   = "a7be1624-04cc-4ed5-8b05-63b633893dee"
	b643Shriek    = "ee5a4902-f993-4fe6-92da-38636962c558"
)

func TestOldB643EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b643Crocodile: "Marsh Crocodile", b643Snidd: "Razing Snidd", b643Emergence: "Natural Emergence",
		b643Drake: "Hunting Drake", b643Spider: "Ancient Spider", b643Thunder: "Thunderscape Familiar",
		b643Sunscape: "Sunscape Familiar", b643Storm: "Stormscape Familiar", b643Judgment: "Dominaria's Judgment",
		b643Overlay: "Planar Overlay", b643Shriek: "Shriek of Dread",
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

func b643Table(t *testing.T) (g *game.Game, me, opp *game.Player) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	return g, me, opp
}

func b643Perm(g *game.Game, owner uuid.UUID, name, typeLine string, p, tough int, colors ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Colors: colors,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

func b643Land(g *game.Game, owner uuid.UUID, typeLine string) uuid.UUID {
	return b643Perm(g, owner, typeLine, "Land — "+typeLine, 0, 0)
}

func b643Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, oracle string, colors []string, targets []game.TargetRef, mana ...string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Colors: colors,
			OracleID: oracle, Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
	})
	effectstest.B06AddMana(me, mana...)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets, Strict: true}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// b643Until passes priority, ordering any triggers, until `who` has an open prompt of `kind`.
func b643Until(t *testing.T, g *game.Game, who uuid.UUID, kind game.PendingChoiceKind) {
	t.Helper()
	for i := 0; i < 24; i++ {
		if effectstest.LatestChoiceOfKindFor(g, kind, who) != nil {
			return
		}
		if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceTriggerOrder, who); c != nil {
			if err := g.ResolveTriggerOrder(c.ID, who, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) && !strings.Contains(err.Error(), "unanswered prompt") {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatalf("no %v prompt for %s", kind, who)
}

func b643AnswerOwn(t *testing.T, g *game.Game, who uuid.UUID, pick uuid.UUID) {
	t.Helper()
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, who)
	if p == nil {
		t.Fatalf("no own_permanents prompt for %s", who)
	}
	if !effectstest.HasID(p.ChooseCards, pick) {
		t.Fatalf("%s is not offered (%v)", pick, p.ChooseCards)
	}
	if err := g.ResolveOwnPermanents(p.ID, who, []uuid.UUID{pick}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
}

func TestOldB643MarshCrocodileReturnsABlueOrBlackCreatureAndEveryoneDiscards(t *testing.T) {
	g, me, opp := b643Table(t)
	blue := b643Perm(g, me.ID, "Merfolk", "Creature — Merfolk", 1, 1, "U")
	green := b643Perm(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	me.Hand.Cards, opp.Hand.Cards = nil, nil
	effectstest.FillHand(me, "M1")
	effectstest.FillHand(opp, "O1")
	croc := b643Cast(t, g, me, "Marsh Crocodile", "Creature — Crocodile", "{2}{U}{B}", b643Crocodile, []string{"U", "B"}, nil, "U", "B", "C", "C")
	b643Until(t, g, me.ID, game.PendingChoiceOwnPermanents)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("no return pick")
	}
	if effectstest.HasID(p.ChooseCards, green) || !effectstest.HasID(p.ChooseCards, blue) || !effectstest.HasID(p.ChooseCards, croc) {
		t.Fatalf("the pick offers %v: want the Merfolk and the Crocodile, not the Bear", p.ChooseCards)
	}
	b643AnswerOwn(t, g, me.ID, blue)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(blue) || g.Battlefield.Contains(blue) || !g.Battlefield.Contains(green) {
		t.Error("the Merfolk should be back in hand and the Bear still on the battlefield")
	}
	for _, p := range []*game.Player{me, opp} {
		if effectstest.DiscardChoiceFor(g, p.ID) == nil {
			t.Fatalf("%s was not asked to discard", p.Name)
		}
	}
}

func TestOldB643RazingSniddReturnsAndEveryPlayerSacrificesALand(t *testing.T) {
	g, me, opp := b643Table(t)
	black := b643Perm(g, me.ID, "Zombie", "Creature — Zombie", 1, 1, "B")
	green := b643Perm(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	myLand, oppLand := b643Land(g, me.ID, "Swamp"), b643Land(g, opp.ID, "Mountain")
	snidd := b643Cast(t, g, me, "Razing Snidd", "Creature — Beast", "{4}{B}{R}", b643Snidd, []string{"B", "R"}, nil, "B", "R", "C", "C", "C", "C")
	b643Until(t, g, me.ID, game.PendingChoiceOwnPermanents)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil || effectstest.HasID(p.ChooseCards, green) || !effectstest.HasID(p.ChooseCards, black) || !effectstest.HasID(p.ChooseCards, snidd) {
		t.Fatalf("return pick %+v: want the Zombie and the Snidd, not the Bear", p)
	}
	b643AnswerOwn(t, g, me.ID, black)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(black) {
		t.Error("the Zombie did not return")
	}
	for i := 0; i < 4 && (g.Battlefield.Contains(myLand) || g.Battlefield.Contains(oppLand)); i++ {
		for _, pl := range []*game.Player{me, opp} {
			for _, k := range []game.PendingChoiceKind{game.PendingChoiceSacrifice, game.PendingChoiceOwnPermanents} {
				if c := effectstest.LatestChoiceOfKindFor(g, k, pl.ID); c != nil && len(c.ChooseCards) > 0 {
					if k == game.PendingChoiceSacrifice {
						_ = g.ResolveSacrificeChoice(c.ID, pl.ID, c.ChooseCards[0])
					} else {
						_ = g.ResolveOwnPermanents(c.ID, pl.ID, c.ChooseCards[:1])
					}
				}
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(myLand) || g.Battlefield.Contains(oppLand) {
		t.Error("each player's only land should have been sacrificed")
	}
}

func TestOldB643NaturalEmergenceAnimatesYourLandsAndReturnsARedOrGreenEnchantment(t *testing.T) {
	g, me, opp := b643Table(t)
	mine, theirs := b643Land(g, me.ID, "Forest"), b643Land(g, opp.ID, "Mountain")
	other := b643Perm(g, me.ID, "Aura", "Enchantment", 0, 0, "G")
	b643Perm(g, me.ID, "WhiteEnch", "Enchantment", 0, 0, "W")
	b643Cast(t, g, me, "Natural Emergence", "Enchantment", "{2}{R}{G}", b643Emergence, []string{"R", "G"}, nil, "R", "G", "C", "C")
	b643Until(t, g, me.ID, game.PendingChoiceOwnPermanents)
	b643AnswerOwn(t, g, me.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(other) {
		t.Error("the green enchantment did not return")
	}
	c := effectstest.E2Card(t, g, mine)
	if !c.IsCreature() || !c.IsLand() || c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("my Forest: creature=%v land=%v %d/%d, want a 2/2 creature land", c.IsCreature(), c.IsLand(), c.CurrentPower(), c.CurrentToughness())
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, mine, "first strike") {
		t.Error("my Forest has no first strike")
	}
	if effectstest.E2Card(t, g, theirs).IsCreature() {
		t.Error("an opponent's land became a creature")
	}
}

func TestOldB643HuntingDrakePutsARedOrGreenCreatureOnTopOfItsLibrary(t *testing.T) {
	g, me, opp := b643Table(t)
	red := b643Perm(g, opp.ID, "Goblin", "Creature — Goblin", 1, 1, "R")
	white := b643Perm(g, opp.ID, "Soldier", "Creature — Soldier", 1, 1, "W")
	drake := b643Cast(t, g, me, "Hunting Drake", "Creature — Drake", "{4}{U}", b643Drake, []string{"U"}, nil, "U", "C", "C", "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asked for no target")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(white)); err == nil {
		t.Fatal("a white creature was accepted")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(red)); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(red) || opp.Library.Size() == 0 || !b643OnTop(opp, red) {
		t.Error("the Goblin is not on top of its owner's library")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, drake, "flying") {
		t.Error("the Drake has no flying")
	}
}

func TestOldB643AncientSpiderHasReachAndFirstStrike(t *testing.T) {
	g, me, _ := b643Table(t)
	id := effectstest.PushCatalogPermanent(g, me.ID, "Ancient Spider", "Creature — Spider", b643Spider, false)
	for _, kw := range []string{"reach", "first strike"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
			t.Errorf("no %s", kw)
		}
	}
}

func TestOldB643FamiliarsDiscountTheirTwoColours(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		discounted, other      string
		kw                     string
	}{
		{"Thunderscape Familiar", "Creature — Kavu", b643Thunder, "B", "W", "first strike"},
		{"Sunscape Familiar", "Creature — Wall", b643Sunscape, "G", "R", "defender"},
		{"Stormscape Familiar", "Creature — Bird", b643Storm, "W", "G", "flying"},
	} {
		g, me, _ := b643Table(t)
		fam := effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
		if !effectstest.HasKeywordOnBattlefield(t, g, fam, tc.kw) {
			t.Errorf("%s: no %s", tc.name, tc.kw)
		}
		// A {1}{X} spell of a discounted colour is castable for the one coloured mana alone.
		other := uuid.New()
		g.WithWriteLock(func() {
			me.Hand.PushTop(game.Card{InstanceID: other, Name: "Other", TypeLine: "Instant", ManaCost: "{1}{" + tc.other + "}",
				Colors: []string{tc.other}, Owner: me.ID, Controller: me.ID})
		})
		effectstest.B06AddMana(me, tc.other)
		if err := g.CastSpell(me.ID, other, game.CastSpellParams{Strict: true}); err == nil {
			t.Errorf("%s: a %s spell got a discount", tc.name, tc.other)
		}
		me.ManaPool = game.ManaPool{}
		good := uuid.New()
		g.WithWriteLock(func() {
			me.Hand.PushTop(game.Card{InstanceID: good, Name: "Good", TypeLine: "Instant", ManaCost: "{1}{" + tc.discounted + "}",
				Colors: []string{tc.discounted}, Owner: me.ID, Controller: me.ID})
		})
		effectstest.B06AddMana(me, tc.discounted)
		if err := g.CastSpell(me.ID, good, game.CastSpellParams{Strict: true}); err != nil {
			t.Errorf("%s: a %s spell for one mana: %v", tc.name, tc.discounted, err)
		}
	}
}

func TestOldB643DominariasJudgmentGrantsProtectionPerLandType(t *testing.T) {
	g, me, opp := b643Table(t)
	b643Land(g, me.ID, "Plains")
	b643Land(g, me.ID, "Island")
	b643Land(g, opp.ID, "Swamp")
	mine := b643Perm(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := b643Perm(g, opp.ID, "Ogre", "Creature — Ogre", 2, 2, "R")
	b643Cast(t, g, me, "Dominaria's Judgment", "Instant", "{2}{W}", b643Judgment, []string{"W"}, nil, "W", "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	for kw, want := range map[string]bool{"protection from white": true, "protection from blue": true,
		"protection from black": false, "protection from red": false, "protection from green": false} {
		if got := effectstest.HasKeywordOnBattlefield(t, g, mine, kw); got != want {
			t.Errorf("my Bear %s = %v, want %v", kw, got, want)
		}
	}
	if effectstest.HasKeywordOnBattlefield(t, g, theirs, "protection from white") {
		t.Error("an opponent's creature gained protection")
	}
}

func TestOldB643ShriekOfDreadGrantsFear(t *testing.T) {
	g, me, opp := b643Table(t)
	bear := b643Perm(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	b643Cast(t, g, me, "Shriek of Dread", "Instant", "{1}{B}", b643Shriek, []string{"B"}, effectstest.CardRefs(bear), "B", "C")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "fear") {
		t.Error("the target did not gain fear")
	}
}

func TestOldB643PlanarOverlayReturnsALandOfEachTypeForEachPlayer(t *testing.T) {
	g, me, opp := b643Table(t)
	f1, f2 := b643Land(g, me.ID, "Forest"), b643Land(g, me.ID, "Forest")
	dual := b643Perm(g, me.ID, "Dual", "Land — Island Swamp", 0, 0)
	// The opponent has one Mountain: it is chosen without a say, and the lone Plains too.
	m, pl := b643Land(g, opp.ID, "Mountain"), b643Land(g, opp.ID, "Plains")
	b643Cast(t, g, me, "Planar Overlay", "Sorcery", "{2}{U}", b643Overlay, []string{"U"}, nil, "U", "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	answered := map[uuid.UUID]int{}
	for i := 0; i < 20; i++ {
		var did bool
		for _, pl := range []*game.Player{me, opp} {
			p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, pl.ID)
			if p == nil || len(p.ChooseCards) == 0 {
				continue
			}
			pick := p.ChooseCards[0]
			if pl == me && effectstest.HasID(p.ChooseCards, f2) {
				pick = f2
			}
			if err := g.ResolveOwnPermanents(p.ID, pl.ID, []uuid.UUID{pick}); err != nil {
				t.Fatalf("answer: %v", err)
			}
			answered[pl.ID]++
			did = true
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !did {
			break
		}
	}
	if g.Battlefield.Contains(f2) || g.Battlefield.Contains(dual) || g.Battlefield.Contains(m) || g.Battlefield.Contains(pl) {
		t.Errorf("a chosen land is still on the battlefield")
	}
	if !g.Battlefield.Contains(f1) {
		t.Error("the Forest that was not chosen was returned")
	}
	if !me.Hand.Contains(f2) || !me.Hand.Contains(dual) || !opp.Hand.Contains(m) || !opp.Hand.Contains(pl) {
		t.Error("the chosen lands are not in their owners' hands")
	}
}

func b643OnTop(p *game.Player, id uuid.UUID) bool {
	top, err := p.Library.Top()
	return err == nil && top.InstanceID == id
}
