package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 491 (Zendikar kicker, #5583): Oran-Rief Recluse, Aether
// Figment, Sphinx of Lost Truths, Elemental Appeal, Vampire's Bite, Bold
// Defense, Goblin Ruinblaster, Gigantiform, Conqueror's Pledge, Tempest Owl and
// Sadistic Sacrament.

const (
	oldB491Recluse   = "043fa53c-8a46-4fbd-836f-2403c62e0121"
	oldB491Figment   = "15fee0bd-ede2-4f65-a9a1-2da4623679d5"
	oldB491Sphinx    = "25597913-aa17-4e3b-a4e9-371ce1459160"
	oldB491Appeal    = "3bcfc82c-3523-40e3-b236-78ece8a8d69f"
	oldB491Bite      = "9e88796b-f37f-4f9f-9a6f-695db2801f0e"
	oldB491Defense   = "d21994f2-75e1-456c-a25a-729c3269df1a"
	oldB491Ruinblast = "d99fa026-6347-450a-9011-1fd79070a55c"
	oldB491Giganti   = "d9e1bae6-2d56-45d1-bdac-2431a18aab3e"
	oldB491Pledge    = "dcc59bd6-2c5f-48eb-8008-83ce6eb5a442"
	oldB491Owl       = "dfd2bc6e-32c8-4027-9c0d-8a3050b8eeda"
	oldB491Sacrament = "e3d08582-85ba-48c4-a453-4a5a385a0c3e"
)

func TestOldB491EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB491Recluse: "Oran-Rief Recluse", oldB491Figment: "Aether Figment", oldB491Sphinx: "Sphinx of Lost Truths",
		oldB491Appeal: "Elemental Appeal", oldB491Bite: "Vampire's Bite", oldB491Defense: "Bold Defense",
		oldB491Ruinblast: "Goblin Ruinblaster", oldB491Giganti: "Gigantiform", oldB491Pledge: "Conqueror's Pledge",
		oldB491Owl: "Tempest Owl", oldB491Sacrament: "Sadistic Sacrament",
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

// oldB491Cast casts a catalog card from the active seat's hand, kicked or not,
// and leaves it on the stack.
func oldB491Cast(t *testing.T, g *game.Game, c game.Card, kicked bool, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	params := game.CastSpellParams{Targets: targets}
	if kicked {
		params.OptionalCosts = []int{0}
	}
	id, err := effectstest.Msc17Cast(t, g, g.Seats[g.Turn.ActiveSeat], c, params)
	if err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	return id
}

func oldB491Card(name, typeLine, oracle string, power, toughness int) game.Card {
	return game.Card{Name: name, TypeLine: typeLine, OracleID: oracle, Power: power, Toughness: toughness}
}

func oldB491Flyer(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bird", Power: 1, Toughness: 1,
		Keywords: []string{"flying"}, Owner: owner, Controller: owner,
	})
}

func TestOldB491OranRiefRecluseDestroysAFlyerOnlyWhenKicked(t *testing.T) {
	for _, kicked := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		flyer := oldB491Flyer(g, opp.ID, "Bird")
		ground := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
		oldB491Cast(t, g, oldB491Card("Oran-Rief Recluse", "Creature — Spider", oldB491Recluse, 1, 3), kicked, nil)
		effectstest.PassPriorityAroundTable(t, g)
		p := effectstest.LatestPickTarget(g, me.ID)
		if !kicked {
			if p != nil {
				t.Error("an unkicked Recluse asked for a target")
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(flyer) {
				t.Error("an unkicked Recluse destroyed a flyer")
			}
			continue
		}
		if p == nil {
			t.Fatalf("no target prompt: %+v", g.PendingChoices)
		}
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: ground}); err == nil {
			t.Error("a creature without flying was accepted as the target")
		}
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: flyer}); err != nil {
			t.Fatalf("pick the flyer: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(flyer) {
			t.Error("the flyer was not destroyed")
		}
		if !g.Battlefield.Contains(ground) {
			t.Error("the ground creature died")
		}
	}
}

func TestOldB491AetherFigmentIsUnblockableAndGrowsWhenKicked(t *testing.T) {
	for _, kicked := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		id := oldB491Cast(t, g, oldB491Card("Aether Figment", "Creature — Illusion", oldB491Figment, 1, 1), kicked, nil)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if kicked {
			want = 2
		}
		if n := effectstest.B12Counter(t, g, id, game.CounterPlusOne); n != want {
			t.Errorf("kicked=%v: %d +1/+1 counters, want %d", kicked, n, want)
		}
		if p, tt := effectstest.PtOf(t, g, id); p != 1+want || tt != 1+want {
			t.Errorf("kicked=%v: %d/%d, want %d/%d", kicked, p, tt, 1+want, 1+want)
		}
	}
}

func TestOldB491AetherFigmentCantBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	figment := effectstest.HmlCreature(g, me.ID, "Aether Figment", oldB491Figment, 1, 1)
	blocker := effectstest.HmlCreature(g, opp.ID, "Blocker", "", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, figment)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, figment); err == nil {
		t.Error("Aether Figment was blocked")
	}
}

func TestOldB491SphinxOfLostTruthsDrawsThreeAndDiscardsThreeUnlessKicked(t *testing.T) {
	for _, kicked := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AuraSeedLibrary(me, 6)
		hand := len(me.Hand.Cards)
		oldB491Cast(t, g, oldB491Card("Sphinx of Lost Truths", "Creature — Sphinx", oldB491Sphinx, 3, 5), kicked, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if got := len(me.Hand.Cards); got != hand+3 {
			t.Errorf("kicked=%v: hand %d, want %d (three drawn)", kicked, got, hand+3)
		}
		asked := effectstest.LatestChooseCardsFor(g, me.ID) != nil || effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceDiscardFromHand, me.ID) != nil
		if asked == kicked {
			t.Errorf("kicked=%v: asked to discard = %v", kicked, asked)
		}
	}
}

func TestOldB491ElementalAppealMakesASevenOneThatIsExiledAtEndOfTurn(t *testing.T) {
	for _, kicked := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		oldB491Cast(t, g, game.Card{Name: "Elemental Appeal", TypeLine: "Sorcery", OracleID: oldB491Appeal}, kicked, nil)
		effectstest.PassPriorityAroundTable(t, g)
		tok := effectstest.BattlefieldTokenNamed(t, g, "Elemental")
		wantP := 7
		if kicked {
			wantP = 14
		}
		if p, tt := effectstest.PtOf(t, g, tok.InstanceID); p != wantP || tt != 1 {
			t.Errorf("kicked=%v: token %d/%d, want %d/1", kicked, p, tt, wantP)
		}
		if !effectstest.HasEffectiveKeyword(t, g, tok.InstanceID, "trample") || !effectstest.HasEffectiveKeyword(t, g, tok.InstanceID, "haste") {
			t.Error("the token lacks trample or haste")
		}
		if got := effectstest.EffectiveColors(t, g, tok.InstanceID); len(got) != 1 || got[0] != "R" {
			t.Errorf("token colours %v, want red", got)
		}
		effectstest.AdvanceToEndStep(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(tok.InstanceID) {
			t.Errorf("kicked=%v: the token was not exiled at the beginning of the end step", kicked)
		}
		_ = me
	}
}

func TestOldB491VampiresBiteGivesLifelinkOnlyWhenKicked(t *testing.T) {
	for _, kicked := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		oldB491Cast(t, g, game.Card{Name: "Vampire's Bite", TypeLine: "Instant", OracleID: oldB491Bite}, kicked, effectstest.CardTarget(bear))
		effectstest.PassPriorityAroundTable(t, g)
		if p, tt := effectstest.PtOf(t, g, bear); p != 5 || tt != 2 {
			t.Errorf("kicked=%v: %d/%d, want 5/2", kicked, p, tt)
		}
		if got := effectstest.HasEffectiveKeyword(t, g, bear, "lifelink"); got != kicked {
			t.Errorf("kicked=%v: has lifelink = %v", kicked, got)
		}
	}
}

func TestOldB491BoldDefensePumpsYourCreaturesAndKickedGivesFirstStrike(t *testing.T) {
	for _, kicked := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
		oldB491Cast(t, g, game.Card{Name: "Bold Defense", TypeLine: "Instant", OracleID: oldB491Defense}, kicked, nil)
		effectstest.PassPriorityAroundTable(t, g)
		want := 1
		if kicked {
			want = 2
		}
		if p, tt := effectstest.PtOf(t, g, mine); p != 2+want || tt != 2+want {
			t.Errorf("kicked=%v: mine %d/%d, want %d/%d", kicked, p, tt, 2+want, 2+want)
		}
		if p, tt := effectstest.PtOf(t, g, theirs); p != 3 || tt != 3 {
			t.Errorf("kicked=%v: an opponent's creature is %d/%d, want 3/3", kicked, p, tt)
		}
		if got := effectstest.HasEffectiveKeyword(t, g, mine, "first strike"); got != kicked {
			t.Errorf("kicked=%v: has first strike = %v", kicked, got)
		}
	}
}

func TestOldB491GoblinRuinblasterDestroysANonbasicLandOnlyWhenKicked(t *testing.T) {
	for _, kicked := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		nonbasic := effectstest.PushLandFor(g, opp.ID, "Karoo", "Land")
		basic := effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
		oldB491Cast(t, g, oldB491Card("Goblin Ruinblaster", "Creature — Goblin Shaman", oldB491Ruinblast, 2, 1), kicked, nil)
		effectstest.PassPriorityAroundTable(t, g)
		p := effectstest.LatestPickTarget(g, me.ID)
		if !kicked {
			if p != nil {
				t.Error("an unkicked Ruinblaster asked for a target")
			}
			continue
		}
		if p == nil {
			t.Fatalf("no target prompt: %+v", g.PendingChoices)
		}
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: basic}); err == nil {
			t.Error("a basic land was accepted as the target")
		}
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: nonbasic}); err != nil {
			t.Fatalf("pick the nonbasic land: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(nonbasic) || !g.Battlefield.Contains(basic) {
			t.Error("the nonbasic land should be destroyed and the basic land left")
		}
	}
}

func TestOldB491GigantiformMakesAnEightEightAndKickedFetchesAnother(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	one := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	two := effectstest.B12Creature(g, me.ID, "Cub", "Creature — Bear", 1, 1)
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Gigantiform", TypeLine: "Enchantment — Aura", OracleID: oldB491Giganti, Owner: me.ID, Controller: me.ID})
	oldB491Cast(t, g, game.Card{Name: "Gigantiform", TypeLine: "Enchantment — Aura", OracleID: oldB491Giganti}, true, effectstest.CardTarget(one))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, one); p != 8 || tt != 8 {
		t.Fatalf("the first host is %d/%d, want 8/8", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, one, "trample") {
		t.Error("the first host lacks trample")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Gigantiform")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerAuraHost(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: two})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, two); p != 8 || tt != 8 {
		t.Errorf("the second host is %d/%d, want 8/8", p, tt)
	}
	if n := effectstest.CountNamed(g, "Gigantiform"); n != 2 {
		t.Errorf("%d Gigantiforms on the battlefield, want 2", n)
	}
}

func TestOldB491GigantiformUnkickedDoesNotSearch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	one := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oldB491Cast(t, g, game.Card{Name: "Gigantiform", TypeLine: "Enchantment — Aura", OracleID: oldB491Giganti}, false, effectstest.CardTarget(one))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil || effectstest.LatestTriggerPromptFor(g, me.ID) != nil {
		t.Error("an unkicked Gigantiform offered a search")
	}
}

func TestOldB491ConquerorsPledgeMakesSixOrTwelveKorSoldiers(t *testing.T) {
	for _, kicked := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		oldB491Cast(t, g, game.Card{Name: "Conqueror's Pledge", TypeLine: "Sorcery", OracleID: oldB491Pledge}, kicked, nil)
		effectstest.PassPriorityAroundTable(t, g)
		want := 6
		if kicked {
			want = 12
		}
		if n := effectstest.B28TokensNamed(g, me.ID, "Kor Soldier"); n != want {
			t.Errorf("kicked=%v: %d Kor Soldier tokens, want %d", kicked, n, want)
		}
	}
}

func TestOldB491TempestOwlTapsUpToThreePermanentsOnlyWhenKicked(t *testing.T) {
	for _, kicked := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		var ids []uuid.UUID
		for i := 0; i < 4; i++ {
			ids = append(ids, effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2))
		}
		oldB491Cast(t, g, oldB491Card("Tempest Owl", "Creature — Bird", oldB491Owl, 1, 2), kicked, nil)
		effectstest.PassPriorityAroundTable(t, g)
		p := effectstest.LatestPickTarget(g, me.ID)
		if !kicked {
			if p != nil {
				t.Error("an unkicked Owl asked for targets")
			}
			continue
		}
		if p == nil {
			t.Fatalf("no target prompt: %+v", g.PendingChoices)
		}
		var four []game.TargetRef
		for _, id := range ids {
			four = append(four, game.TargetRef{Kind: game.TargetCard, ID: id})
		}
		if err := g.ResolvePickTargets(p.ID, me.ID, four); err == nil {
			t.Error("four targets were accepted for 'up to three'")
		}
		if err := g.ResolvePickTargets(p.ID, me.ID, four[:3]); err != nil {
			t.Fatalf("pick three: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		for i, id := range ids {
			if got := effectstest.TappedForTest(t, g, id); got != (i < 3) {
				t.Errorf("creature %d tapped = %v", i, got)
			}
		}
	}
}

func TestOldB491SadisticSacramentExilesUpToThreeOrFifteen(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		for len(opp.Library.Cards) < 20 {
			opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Test", Owner: opp.ID, Controller: opp.ID})
		}
		before := len(opp.Library.Cards)
		oldB491Cast(t, g, game.Card{Name: "Sadistic Sacrament", TypeLine: "Sorcery", OracleID: oldB491Sacrament}, kicked, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.SearchChoiceFor(g, me.ID)
		if c == nil {
			t.Fatalf("kicked=%v: no search prompt", kicked)
		}
		want := 3
		if kicked {
			want = 15
		}
		if c.SearchMax != want {
			t.Errorf("kicked=%v: the search may take up to %d, want %d", kicked, c.SearchMax, want)
		}
		var picks []uuid.UUID
		for i := 0; i < len(c.SearchCards) && i < want; i++ {
			picks = append(picks, c.SearchCards[i])
		}
		effectstest.AnswerSearchByID(t, g, me.ID, picks...)
		effectstest.PassPriorityAroundTable(t, g)
		if got := before - len(opp.Library.Cards); got != want {
			t.Errorf("kicked=%v: %d cards left the library, want %d", kicked, got, want)
		}
		for _, id := range picks {
			if !effectstest.InExile(g, id) {
				t.Errorf("kicked=%v: a found card is not in exile", kicked)
			}
		}
	}
}
