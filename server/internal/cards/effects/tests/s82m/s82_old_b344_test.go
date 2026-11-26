package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b344_test.go — Khans of Tarkir cards, S82 OLD batch 344 (#5554),
// played through the real catalog entries.

const (
	oldB344Ascendancy = "06e92ab3-c781-4f32-9ada-09a388047067"
	oldB344Archive    = "260f95c5-4058-4627-981f-98b9eafc2265"
	oldB344Mentor     = "2c47319a-4062-4aec-b68a-e8e8f023cf86"
	oldB344Spoils     = "45fd14b9-d66a-452d-add4-36f512fff9c2"
	oldB344Spray      = "88ce2017-83a6-4b55-afba-371875bf2172"
	oldB344MasterWay  = "909a10f5-a607-474f-8b8b-0176fd840650"
	oldB344Scheming   = "d9c57520-9827-4b8b-8f81-43cca4227c60"
	oldB344Embodiment = "e12bba04-b10f-4bbd-959c-560f31630951"
)

var oldB344Banners = []struct {
	name, oracle, cost string
	colors             [3]string
}{
	{"Abzan Banner", "46535f8e-1bcd-4588-ac6c-a4bc89c379c8", "{W}{B}{G}", [3]string{"W", "B", "G"}},
	{"Jeskai Banner", "fda10ef2-f431-4d9b-96a4-4db4c98d2b0c", "{U}{R}{W}", [3]string{"U", "R", "W"}},
	{"Mardu Banner", "9805b7b3-086a-4c91-810f-52c8f8f95f6a", "{R}{W}{B}", [3]string{"R", "W", "B"}},
	{"Sultai Banner", "93b139a5-2678-46a5-b8eb-45344ee3290a", "{B}{G}{U}", [3]string{"B", "G", "U"}},
	{"Temur Banner", "8a35f23d-f59a-4677-9736-035c60c22a2b", "{G}{U}{R}", [3]string{"G", "U", "R"}},
}

func TestOldB344EveryCardIsRegisteredFull(t *testing.T) {
	ids := map[string]string{
		oldB344Ascendancy: "Sultai Ascendancy", oldB344Archive: "Cranial Archive", oldB344Mentor: "Bloodfire Mentor",
		oldB344Spoils: "Raiders' Spoils", oldB344Spray: "Blinding Spray", oldB344MasterWay: "Master the Way",
		oldB344Scheming: "Taigam's Scheming", oldB344Embodiment: "Embodiment of Spring",
	}
	for _, b := range oldB344Banners {
		ids[b.oracle] = b.name
	}
	for id, name := range ids {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func TestOldB344BannersTapForEachOfTheirColoursAndCashIn(t *testing.T) {
	for _, b := range oldB344Banners {
		for _, color := range b.colors {
			t.Run(b.name+" "+color, func(t *testing.T) {
				g := effectstest.NewCatalogGame(t)
				me := g.Seats[0]
				effectstest.AdvanceToMain(t, g)
				banner := effectstest.PushCatalogPermanent(g, me.ID, b.name, "Artifact", b.oracle, false)
				effectstest.B28TapForMana(t, g, me.ID, banner, color)
				if got := me.ManaPool.Total(); got != 1 {
					t.Errorf("pool = %d, want 1", got)
				}
			})
		}
		t.Run(b.name+" draws", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			banner := effectstest.PushCatalogPermanent(g, me.ID, b.name, "Artifact", b.oracle, false)
			if err := effectstest.MirB1Activate(g, me.ID, banner, 0); err == nil {
				t.Fatal("activated without paying " + b.cost)
			}
			effectstest.DcMana(t, g, me.ID, b.cost)
			hand := me.Hand.Size()
			if err := effectstest.MirB1Activate(g, me.ID, banner, 0); err != nil {
				t.Fatalf("activate: %v", err)
			}
			if g.Battlefield.Contains(banner) {
				t.Error("the Banner was not sacrificed as a cost")
			}
			effectstest.PassPriorityAroundTable(t, g)
			if me.Hand.Size() != hand+1 {
				t.Errorf("hand = %d, want %d", me.Hand.Size(), hand+1)
			}
		})
	}
}

func TestOldB344SultaiAscendancySurveilsTwoEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.MirB1Push(g, me.ID, "Sultai Ascendancy", "Enchantment", oldB344Ascendancy, 0, 0)
	top := effectstest.LibraryTopIDs(me, 2)
	effectstest.BeginUpkeep(t, g, me.ID)
	ch := effectstest.SurveilChoiceFor(g, me.ID)
	if ch == nil || len(ch.ScryCards) != 2 {
		t.Fatalf("surveil prompt = %+v, want two cards", ch)
	}
	effectstest.RfmaAnswerSurveil(t, g, me, top[0])
	if !me.Graveyard.Contains(top[0]) || me.Graveyard.Contains(top[1]) {
		t.Error("only the first card should have gone to the graveyard")
	}
}

func TestOldB344TaigamsSchemingSurveilsFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.P1g2Cast(t, g, "Taigam's Scheming", "Sorcery", oldB344Scheming, "{1}{U}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.SurveilChoiceFor(g, me.ID)
	if ch == nil || len(ch.ScryCards) != 5 {
		t.Fatalf("surveil prompt = %+v, want five cards", ch)
	}
	all := append([]uuid.UUID(nil), ch.ScryCards...)
	effectstest.RfmaAnswerSurveil(t, g, me, all...)
	for _, id := range all {
		if !me.Graveyard.Contains(id) {
			t.Errorf("%s did not reach the graveyard", id)
		}
	}
}

func TestOldB344BloodfireMentorLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	mentor := effectstest.MirB1Push(g, me.ID, "Bloodfire Mentor", "Creature — Efreet Shaman", oldB344Mentor, 0, 5)
	effectstest.HandCardFull(me, "Fodder", "Sorcery", "{1}", "", nil)
	hand := me.Hand.Size()
	effectstest.DcMana(t, g, me.ID, "{C}{C}{U}")
	if err := effectstest.MirB1Activate(g, me.ID, mentor, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCardsOrDiscard(t, g, me)
	if me.Hand.Size() != hand || me.Graveyard.Size() != 1 {
		t.Errorf("hand %d (want %d), graveyard %d (want 1): a card was drawn and a card discarded", me.Hand.Size(), hand, me.Graveyard.Size())
	}
	if !effectstest.TappedOnBattlefield(t, g, mentor) {
		t.Error("the Mentor did not tap")
	}
}

func TestOldB344BlindingSpraySlashesOpposingPowerAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 5, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 6, 6)
	hand := me.Hand.Size()
	effectstest.P1g2Cast(t, g, "Blinding Spray", "Instant", oldB344Spray, "{4}{U}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
		t.Errorf("their power = %d, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, mine); got != 5 {
		t.Errorf("my power = %d, want 5", got)
	}
	if got := effectstest.EffectiveToughness(t, g, theirs); got != 6 {
		t.Errorf("their toughness = %d, want 6", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d", me.Hand.Size(), hand+1)
	}
	late := effectstest.B12Creature(g, opp.ID, "Late", "Creature — Bear", 6, 6)
	if got := effectstest.EffectivePower(t, g, late); got != 6 {
		t.Errorf("a creature that arrived later has power %d, want 6", got)
	}
}

func TestOldB344MasterTheWayDrawsThenHitsForHandSize(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	me.Hand.Cards = nil
	for i := 0; i < 2; i++ {
		effectstest.HandCardFull(me, "Filler", "Sorcery", "{1}", "", nil)
	}
	life := opp.Life
	effectstest.P1g2Cast(t, g, "Master the Way", "Sorcery", oldB344MasterWay, "{3}{U}{R}",
		game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 3 {
		t.Fatalf("hand = %d, want 3 (two fillers and the drawn card)", me.Hand.Size())
	}
	if got := life - opp.Life; got != 3 {
		t.Errorf("damage = %d, want 3 (the card drawn counts)", got)
	}
}

func TestOldB344EmbodimentOfSpringFetchesATappedBasic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	emb := effectstest.MirB1Push(g, me.ID, "Embodiment of Spring", "Creature — Elemental", oldB344Embodiment, 0, 3)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Wasteland", TypeLine: "Land"})
	effectstest.DcMana(t, g, me.ID, "{C}{G}")
	if err := effectstest.MirB1Activate(g, me.ID, emb, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(emb) {
		t.Error("the Embodiment was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.SearchChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no search prompt")
	}
	for _, id := range prompt.SearchCards {
		if zone, c := effectstest.CardWhere(g, me, id); zone != "" && c.Name == "Wasteland" {
			t.Error("a nonbasic land was offered")
		}
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	effectstest.PassPriorityAroundTable(t, g)
	forests := oldB344Named(g, "Forest", me.ID)
	if len(forests) != 1 || !forests[0].Tapped {
		t.Errorf("Forest on the battlefield = %+v, want one tapped", forests)
	}
}

func TestOldB344CranialArchiveShufflesAGraveyardAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	archive := effectstest.MirB1Push(g, me.ID, "Cranial Archive", "Artifact", oldB344Archive, 0, 0)
	effectstest.PushGraveyardCardTyped(opp, "Dead A", "Creature — Bear")
	effectstest.PushGraveyardCardTyped(opp, "Dead B", "Sorcery")
	lib := opp.Library.Size()
	effectstest.DcMana(t, g, me.ID, "{C}{C}")
	hand := me.Hand.Size()
	if err := effectstest.MirB1Activate(g, me.ID, archive, 0, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(archive) {
		t.Error("the Archive was not exiled as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Size() != 0 || opp.Library.Size() != lib+2 {
		t.Errorf("graveyard %d library %d, want 0 and %d", opp.Graveyard.Size(), opp.Library.Size(), lib+2)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestOldB344RaidersSpoilsPumpsCreaturesAndPaysLifeForCardsOnWarriorHits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Raiders' Spoils", "Enchantment", oldB344Spoils, false)
	warrior := effectstest.B12Creature(g, me.ID, "Warrior", "Creature — Human Warrior", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if got := effectstest.EffectivePower(t, g, warrior); got != 3 {
		t.Errorf("my power = %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
		t.Errorf("their power = %d, want 2", got)
	}
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	for _, id := range []uuid.UUID{warrior, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	for i := 0; i < 8 && g.Turn.Step != game.StepCombatDamage; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	life, hand := me.Life, me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.PendingOfKind(g, game.PendingChoicePayUnless); c != nil {
		effectstest.AnswerPayUnless(t, g, me.ID, true)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-1 || me.Hand.Size() != hand+1 {
		t.Errorf("life %d→%d, hand %d→%d: want exactly one 1-life payment and one card (the Bear is not a Warrior)", life, me.Life, hand, me.Hand.Size())
	}
}

func oldB344Named(g *game.Game, name string, controller uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			out = append(out, c)
		}
	}
	return out
}
