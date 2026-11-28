package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b826_cards_test.go — Homelands: card draw and library, other
// creatures, creatures with keywords, batch 826 (S82 older sets, #5635).
// Every test plays the real catalog entry through the engine.

const (
	oldB826Forget    = "619ef7e1-33cd-4470-a1d4-83c5f1f5c31e"
	oldB826Prophecy  = "8dcf15fd-f000-4424-b74c-22fc2642b8c3"
	oldB826Renewal   = "8df46910-ee13-416e-8282-dfe85a90e6d5"
	oldB826WinterSky = "fb150d6a-5f81-4fd8-92b2-22a9615bfd4f"
	oldB826Truce     = "fe4aa72b-f176-461a-aca0-1519721609b2"
	oldB826Irini     = "2727a7ec-1cf7-4cd7-b5d4-955db940ddd9"
	oldB826Aysen     = "3ebbe05c-7cc4-4845-87e5-ab93d7ac9623"
	oldB826AnHavva   = "7131cdb6-15d0-407e-a7e2-b2a0deef9363"
	oldB826Mist      = "dedfbd92-4f12-423c-9d56-167753f03bff"
	oldB826Hounds    = "26957a44-6155-4688-a850-a3b49b011a9f"
	oldB826Faerie    = "5390bb44-985e-473b-883e-c9562c06f210"
	oldB826Shade     = "5eb29367-4ec3-4bd2-887e-13eeeecae8fa"
)

func TestOldB826EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB826Forget: "Forget", oldB826Prophecy: "Prophecy", oldB826Renewal: "Renewal",
		oldB826WinterSky: "Winter Sky", oldB826Truce: "Truce", oldB826Irini: "Irini Sengir",
		oldB826Aysen: "Aysen Crusader", oldB826AnHavva: "An-Havva Constable", oldB826Mist: "Hungry Mist",
		oldB826Hounds: "Ghost Hounds", oldB826Faerie: "Willow Faerie", oldB826Shade: "Ihsan's Shade",
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

// oldB826Cast puts a card in the active seat's hand with a pool and casts it
// strictly.
func oldB826Cast(t *testing.T, g *game.Game, c game.Card, pool string, p game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = me.ID, me.ID
	me.Hand.PushTop(c)
	effectstest.FloatForTest(g, me, pool)
	p.Strict = true
	return c.InstanceID, g.CastSpell(me.ID, c.InstanceID, p)
}

func oldB826HandIDs(p *game.Player) []uuid.UUID {
	var ids []uuid.UUID
	for _, c := range p.Hand.Cards {
		ids = append(ids, c.InstanceID)
	}
	return ids
}

func oldB826Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB826DrawsAtTheNextUpkeep(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	before := p.Hand.Size()
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if p.Hand.Size() != before {
		t.Fatalf("drew before the upkeep trigger resolved: hand %d -> %d", before, p.Hand.Size())
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p.Hand.Size() != before+1 {
		t.Errorf("hand %d -> %d at the next upkeep, want +1", before, p.Hand.Size())
	}
}

// --- Forget ------------------------------------------------------------------

func TestOldB826ForgetDiscardsTwoThenDrawsAsManyAsWereDiscarded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := oldB826Seats(g)
	effectstest.SeedLibrary(opp, "A", "B", "C", "D")
	hand, yard, lib := opp.Hand.Size(), opp.Graveyard.Size(), opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Forget", "Sorcery", oldB826Forget, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	pick := append([]uuid.UUID(nil), oldB826HandIDs(opp)[:2]...)
	effectstest.AnswerDiscard(t, g, opp.ID, pick...)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Graveyard.Size(); got != yard+2 {
		t.Errorf("graveyard %d -> %d, want +2", yard, got)
	}
	if got := opp.Library.Size(); got != lib-2 {
		t.Errorf("library %d -> %d, want -2 (two drawn)", lib, got)
	}
	if got := opp.Hand.Size(); got != hand {
		t.Errorf("hand %d -> %d, want unchanged (two out, two in)", hand, got)
	}
}

func TestOldB826ForgetWithOneCardInHandDrawsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := oldB826Seats(g)
	effectstest.SeedLibrary(opp, "A", "B", "C", "D")
	for _, id := range oldB826HandIDs(opp)[1:] {
		_, _ = opp.Hand.Remove(id)
	}
	lib := opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Forget", "Sorcery", oldB826Forget, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, opp.ID) != nil {
		effectstest.DiscardFromHand(t, g, opp.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != lib-1 {
		t.Errorf("library %d -> %d, want -1 (one discarded, one drawn)", lib, got)
	}
	if got := opp.Hand.Size(); got != 1 {
		t.Errorf("hand %d, want 1", got)
	}
}

func TestOldB826ForgetWithAnEmptyHandDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := oldB826Seats(g)
	effectstest.SeedLibrary(opp, "A", "B", "C")
	for _, id := range oldB826HandIDs(opp) {
		_, _ = opp.Hand.Remove(id)
	}
	lib := opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Forget", "Sorcery", oldB826Forget, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != lib || opp.Hand.Size() != 0 {
		t.Errorf("library %d -> %d, hand %d: a player with no cards discarded none and drew none", lib, got, opp.Hand.Size())
	}
}

func TestOldB826ForgetNeedsATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	if _, err := oldB826Cast(t, g, game.Card{Name: "Forget", TypeLine: "Sorcery", OracleID: oldB826Forget, ManaCost: "{U}{U}", Colors: []string{"U"}}, "UU", game.CastSpellParams{}); err == nil {
		t.Error("Forget was cast without a target player")
	}
}

// --- Prophecy ----------------------------------------------------------------

func oldB826ProphecyCard() game.Card {
	return game.Card{Name: "Prophecy", TypeLine: "Sorcery", OracleID: oldB826Prophecy, ManaCost: "{W}", Colors: []string{"W"}}
}

func TestOldB826ProphecyGainsLifeOnALandThenDrawsAtTheNextUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB826Seats(g)
	effectstest.SeedLibrary(me, "A", "B")
	effectstest.SeedLibraryTop(opp, "Forest", "Basic Land — Forest", "")
	life, lib := me.Life, opp.Library.Size()
	if _, err := oldB826Cast(t, g, oldB826ProphecyCard(), "W", game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1 for a revealed land", life, me.Life)
	}
	if opp.Library.Size() != lib {
		t.Errorf("their library %d -> %d: the card was revealed, not drawn", lib, opp.Library.Size())
	}
	oldB826DrawsAtTheNextUpkeep(t, g, me)
}

func TestOldB826ProphecyGainsNothingOnANonland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB826Seats(g)
	effectstest.SeedLibraryTop(opp, "Bear", "Creature — Bear", "{1}{G}")
	life := me.Life
	if _, err := oldB826Cast(t, g, oldB826ProphecyCard(), "W", game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d -> %d, want unchanged for a nonland", life, me.Life)
	}
}

func TestOldB826ProphecyRefusesYourselfAsTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB826Seats(g)
	if _, err := oldB826Cast(t, g, oldB826ProphecyCard(), "W", game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(me.ID)}}); err == nil {
		t.Error("Prophecy targeted its own controller")
	}
}

// --- Renewal -----------------------------------------------------------------

func oldB826RenewalCard() game.Card {
	return game.Card{Name: "Renewal", TypeLine: "Sorcery", OracleID: oldB826Renewal, ManaCost: "{2}{G}", Colors: []string{"G"}}
}

func TestOldB826RenewalSacrificesALandFetchesABasicUntappedAndDrawsLater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB826Seats(g)
	effectstest.SeedLibrary(me, "A", "B")
	old := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	basic := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	if _, err := oldB826Cast(t, g, oldB826RenewalCard(), "CCG", game.CastSpellParams{SacrificeIDs: []uuid.UUID{old}}); err != nil {
		t.Fatal(err)
	}
	if g.Battlefield.Contains(old) {
		t.Error("the land was not sacrificed as the spell was cast")
	}
	effectstest.PassPriorityAroundTable(t, g)
	open := effectstest.ChkB1OpenSearch(g)
	if open == nil {
		t.Fatal("no library search opened")
	}
	if err := g.ResolveSearchLibrary(open.ID, open.Chooser, []uuid.UUID{basic}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.FindBattlefieldCardForTest(g, basic)
	if c == nil || c.Tapped {
		t.Errorf("the basic land is not on the battlefield untapped: %+v", c)
	}
	oldB826DrawsAtTheNextUpkeep(t, g, me)
}

func TestOldB826RenewalIsRefusedWithNoLandToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	if _, err := oldB826Cast(t, g, oldB826RenewalCard(), "CCG", game.CastSpellParams{}); err == nil {
		t.Error("Renewal was cast without sacrificing a land")
	}
}

// --- Winter Sky --------------------------------------------------------------

func TestOldB826WinterSkyDamagesEveryoneOnAWinAndDrawsOnALoss(t *testing.T) {
	won, lost := 0, 0
	for i := 0; i < 40 && (won == 0 || lost == 0); i++ {
		g := effectstest.NewCatalogGame(t)
		me, opp := oldB826Seats(g)
		for _, p := range g.Seats {
			effectstest.SeedLibrary(p, "A", "B", "C")
		}
		bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
		runt := effectstest.B12Creature(g, me.ID, "Runt", "Creature — Rat", 1, 1)
		hands := make([]int, len(g.Seats))
		lives := make([]int, len(g.Seats))
		if _, err := oldB826Cast(t, g, game.Card{Name: "Winter Sky", TypeLine: "Sorcery", OracleID: oldB826WinterSky, ManaCost: "{R}", Colors: []string{"R"}}, "R", game.CastSpellParams{}); err != nil {
			t.Fatal(err)
		}
		for j, p := range g.Seats {
			hands[j], lives[j] = p.Hand.Size(), p.Life
		}
		effectstest.PassPriorityAroundTable(t, g)
		ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if ask == nil {
			t.Fatal("no coin call was asked")
		}
		flips := len(effectstest.RandomEvents(g, game.EventFlipCoin))
		if err := g.ResolveCoinCall(ask.ID, me.ID, "heads"); err != nil {
			t.Fatalf("ResolveCoinCall: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		evs := effectstest.RandomEvents(g, game.EventFlipCoin)
		if len(evs) != flips+1 {
			t.Fatalf("%d coins flipped, want 1", len(evs)-flips)
		}
		if evs[len(evs)-1].Won {
			won++
			if g.Battlefield.Contains(runt) {
				t.Error("a won flip left the 1/1 alive")
			}
			if !g.Battlefield.Contains(bear) {
				t.Error("a won flip killed the 2/2 (1 damage only)")
			}
			for j, p := range g.Seats {
				if p.Life != lives[j]-1 || p.Hand.Size() != hands[j] {
					t.Errorf("seat %d after a won flip: life %d -> %d, hand %d -> %d; want -1 life and no draw", j, lives[j], p.Life, hands[j], p.Hand.Size())
				}
			}
		} else {
			lost++
			if !g.Battlefield.Contains(runt) || !g.Battlefield.Contains(bear) {
				t.Error("a lost flip dealt damage")
			}
			for j, p := range g.Seats {
				if p.Life != lives[j] || p.Hand.Size() != hands[j]+1 {
					t.Errorf("seat %d after a lost flip: life %d -> %d, hand %d -> %d; want a card each", j, lives[j], p.Life, hands[j], p.Hand.Size())
				}
			}
		}
	}
	if won == 0 || lost == 0 {
		t.Fatalf("saw %d won and %d lost flips; want both", won, lost)
	}
}

// --- Truce -------------------------------------------------------------------

func TestOldB826TruceAsksEachPlayerAndPaysForTheShortfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	for _, p := range g.Seats {
		effectstest.SeedLibrary(p, "A", "B", "C", "D")
	}
	if _, err := oldB826Cast(t, g, game.Card{Name: "Truce", TypeLine: "Instant", OracleID: oldB826Truce, ManaCost: "{2}{W}", Colors: []string{"W"}}, "CCW", game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	hands := make([]int, len(g.Seats))
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		hands[i], lives[i] = p.Hand.Size(), p.Life
	}
	effectstest.PassPriorityAroundTable(t, g)
	answers := []int{2, 1, 0, 2}
	for i := 0; i < len(g.Seats); i++ {
		seat := g.Seats[(g.Turn.ActiveSeat+i)%len(g.Seats)]
		pa := effectstest.AnswerNumber(t, g, seat.ID, answers[i])
		if pa.Min != 0 || pa.Max != 2 {
			t.Errorf("seat %d prompt %d..%d, want 0..2", i, pa.Min, pa.Max)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	for i := 0; i < len(g.Seats); i++ {
		idx := (g.Turn.ActiveSeat + i) % len(g.Seats)
		p := g.Seats[idx]
		if got := p.Hand.Size() - hands[idx]; got != answers[i] {
			t.Errorf("seat %d drew %d, want %d", i, got, answers[i])
		}
		if got := p.Life - lives[idx]; got != 2*(2-answers[i]) {
			t.Errorf("seat %d gained %d life, want %d", i, got, 2*(2-answers[i]))
		}
	}
}

func TestOldB826TruceIsAnInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB826Seats(g)
	for _, p := range g.Seats {
		effectstest.SeedLibrary(p, "A", "B", "C")
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	c := game.Card{InstanceID: uuid.New(), Name: "Truce", TypeLine: "Instant", OracleID: oldB826Truce, ManaCost: "{2}{W}", Colors: []string{"W"}, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(c)
	effectstest.FloatForTest(g, me, "CCW")
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("Truce could not be cast in the beginning of combat step: %v", err)
	}
}

func TestOldB826TruceIsRefusedWithoutItsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	if _, err := oldB826Cast(t, g, game.Card{Name: "Truce", TypeLine: "Instant", OracleID: oldB826Truce, ManaCost: "{2}{W}", Colors: []string{"W"}}, "W", game.CastSpellParams{}); err == nil {
		t.Error("Truce was cast with one mana")
	}
}

// --- Irini Sengir ------------------------------------------------------------

func oldB826Price(t *testing.T, g *game.Game, seat *game.Player, typeLine, cost string, colors ...string) int {
	t.Helper()
	base, err := game.ParseCost(cost)
	if err != nil {
		t.Fatalf("ParseCost(%q): %v", cost, err)
	}
	out, err := g.ApplyCostModifiers(base, game.CostQuery{
		Card: game.Card{
			InstanceID: uuid.New(), Name: "Spell", TypeLine: typeLine, ManaCost: cost, Colors: colors,
			Owner: seat.ID, Controller: seat.ID,
		},
		Controller: seat.ID,
		FromZone:   game.ZoneHand,
	})
	if err != nil {
		t.Fatalf("ApplyCostModifiers: %v", err)
	}
	return out.ManaValue()
}

func TestOldB826IriniTaxesGreenAndWhiteEnchantmentSpellsOfEveryone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Irini Sengir", "Legendary Creature — Vampire Dwarf", oldB826Irini, 2, 2)
	for _, tc := range []struct {
		name     string
		who      *game.Player
		typeLine string
		cost     string
		colors   []string
		want     int
	}{
		{"own green enchantment", me, "Enchantment", "{1}{G}", []string{"G"}, 4},
		{"opponent's white enchantment", opp, "Enchantment", "{1}{W}", []string{"W"}, 4},
		{"green and white enchantment pays once", me, "Enchantment", "{G}{W}", []string{"G", "W"}, 4},
		{"green Aura", me, "Enchantment — Aura", "{G}", []string{"G"}, 3},
		{"blue enchantment", me, "Enchantment", "{1}{U}", []string{"U"}, 2},
		{"green creature", me, "Creature — Bear", "{1}{G}", []string{"G"}, 2},
		{"white sorcery", me, "Sorcery", "{W}", []string{"W"}, 1},
		{"colorless enchantment", opp, "Enchantment", "{2}", nil, 2},
	} {
		if got := oldB826Price(t, g, tc.who, tc.typeLine, tc.cost, tc.colors...); got != tc.want {
			t.Errorf("%s: costs %d, want %d", tc.name, got, tc.want)
		}
	}
}

func TestOldB826IriniRefusesAGreenEnchantmentWithoutTheExtraTwo(t *testing.T) {
	for _, tc := range []struct {
		pool    string
		refused bool
	}{{"G", true}, {"CG", true}, {"CCG", false}} {
		g := effectstest.NewCatalogGame(t)
		me, _ := oldB826Seats(g)
		effectstest.B12Push(g, me.ID, "Irini Sengir", "Legendary Creature — Vampire Dwarf", oldB826Irini, 2, 2)
		_, err := oldB826Cast(t, g, game.Card{Name: "Verdant Ward", TypeLine: "Enchantment", ManaCost: "{G}", Colors: []string{"G"}}, tc.pool, game.CastSpellParams{})
		if (err != nil) != tc.refused {
			t.Errorf("pool %q: error %v, want refused %v", tc.pool, err, tc.refused)
		}
	}
}

// --- Aysen Crusader ----------------------------------------------------------

func TestOldB826AysenCrusaderCountsYourSoldiersAndWarriors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crusader := effectstest.B12Push(g, me.ID, "Aysen Crusader", "Creature — Human Knight", oldB826Aysen, 0, 0)
	if p, tt := effectstest.PtOf(t, g, crusader); p != 2 || tt != 2 {
		t.Errorf("alone: %d/%d, want 2/2", p, tt)
	}
	effectstest.B12Creature(g, me.ID, "Soldier", "Creature — Human Soldier", 1, 1)
	effectstest.B12Creature(g, me.ID, "Warrior", "Creature — Human Warrior", 1, 1)
	both := effectstest.B12Creature(g, me.ID, "Both", "Creature — Human Soldier Warrior", 1, 1)
	effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	effectstest.B12Creature(g, opp.ID, "Their Soldier", "Creature — Human Soldier", 1, 1)
	if p, tt := effectstest.PtOf(t, g, crusader); p != 5 || tt != 5 {
		t.Errorf("with three of mine: %d/%d, want 5/5 (a Soldier Warrior counts once, theirs and Elves not at all)", p, tt)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, both); err != nil {
		t.Fatal(err)
	}
	if p, tt := effectstest.PtOf(t, g, crusader); p != 4 || tt != 4 {
		t.Errorf("after one leaves: %d/%d, want 4/4", p, tt)
	}
}

// --- An-Havva Constable ------------------------------------------------------

func oldB826Green(g *game.Game, owner uuid.UUID, name string, colors ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Colors: colors, Owner: owner, Controller: owner,
	})
}

func TestOldB826AnHavvaConstableToughnessIsOnePlusGreenCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	constable := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "An-Havva Constable", TypeLine: "Creature — Human", OracleID: oldB826AnHavva,
		Power: 2, Toughness: 1, Colors: []string{"G"}, Owner: me.ID, Controller: me.ID,
	})
	if p, tt := effectstest.PtOf(t, g, constable); p != 2 || tt != 2 {
		t.Errorf("alone: %d/%d, want 2/2 (itself is green)", p, tt)
	}
	oldB826Green(g, me.ID, "Mine", "G")
	theirs := oldB826Green(g, opp.ID, "Theirs", "G")
	oldB826Green(g, opp.ID, "Blue", "U")
	if p, tt := effectstest.PtOf(t, g, constable); p != 2 || tt != 4 {
		t.Errorf("with three green creatures: %d/%d, want 2/4", p, tt)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, theirs); err != nil {
		t.Fatal(err)
	}
	if _, tt := effectstest.PtOf(t, g, constable); tt != 3 {
		t.Errorf("after one leaves: toughness %d, want 3", tt)
	}
}

// --- Hungry Mist -------------------------------------------------------------

func oldB826MistUpkeep(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mist := effectstest.B12Push(g, me.ID, "Hungry Mist", "Creature — Elemental", oldB826Mist, 6, 2)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	return g, me, mist
}

func TestOldB826HungryMistStaysWhenItsUpkeepCostIsPaid(t *testing.T) {
	g, me, mist := oldB826MistUpkeep(t)
	effectstest.B06AddMana(me, "G", "G")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("no pay prompt for the upkeep trigger")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mist) {
		t.Error("Hungry Mist left after its controller paid {G}{G}")
	}
}

func TestOldB826HungryMistIsSacrificedWhenNotPaid(t *testing.T) {
	g, me, mist := oldB826MistUpkeep(t)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		effectstest.AnswerPayUnless(t, g, me.ID, false)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mist) {
		t.Error("Hungry Mist stayed without payment")
	}
}

// --- Ghost Hounds ------------------------------------------------------------

func oldB826HasFirstStrike(g *game.Game, id uuid.UUID) bool {
	has := false
	g.ReadSnapshot(func() {
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.InstanceID == id {
				has = game.HasKeyword(&c, "first strike")
			}
		}
	})
	return has
}

func oldB826HoundsOf(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Push(g, owner, "Ghost Hounds", "Creature — Dog Spirit", oldB826Hounds, 1, 1)
}

func oldB826Other(g *game.Game, owner uuid.UUID, color string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Other", TypeLine: "Creature — Bear", Power: 0, Toughness: 9,
		Colors: []string{color}, Owner: owner, Controller: owner,
	})
}

func TestOldB826GhostHoundsGainsFirstStrikeWhenItBlocksAWhiteCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	white := oldB826Other(g, me.ID, "W")
	hounds := oldB826HoundsOf(g, opp.ID)
	if oldB826HasFirstStrike(g, hounds) {
		t.Fatal("the Hounds began with first strike")
	}
	effectstest.Pr7bAttack(t, g, opp.ID, white)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{hounds: white})
	if !oldB826HasFirstStrike(g, hounds) {
		t.Error("the Hounds did not gain first strike blocking a white creature")
	}
}

func TestOldB826GhostHoundsGainsFirstStrikeWhenAWhiteCreatureBlocksIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hounds := oldB826HoundsOf(g, me.ID)
	white := oldB826Other(g, opp.ID, "W")
	effectstest.Pr7bAttack(t, g, opp.ID, hounds)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{white: hounds})
	if !oldB826HasFirstStrike(g, hounds) {
		t.Error("the Hounds did not gain first strike when blocked by a white creature")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if oldB826HasFirstStrike(g, hounds) {
		t.Error("first strike outlasted the turn")
	}
}

func TestOldB826GhostHoundsIgnoresANonwhiteCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hounds := oldB826HoundsOf(g, me.ID)
	green := oldB826Other(g, opp.ID, "G")
	effectstest.Pr7bAttack(t, g, opp.ID, hounds)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{green: hounds})
	if oldB826HasFirstStrike(g, hounds) {
		t.Error("the Hounds gained first strike from a green blocker")
	}
}

func TestOldB826GhostHoundsHasVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hounds := oldB826HoundsOf(g, me.ID)
	effectstest.Pr7bAttack(t, g, opp.ID, hounds)
	c := effectstest.FindBattlefieldCardForTest(g, hounds)
	if c == nil || c.Tapped {
		t.Errorf("the Hounds tapped to attack: %+v", c)
	}
}

// --- Willow Faerie and Ihsan's Shade -----------------------------------------

func TestOldB826WillowFaerieFliesOverGroundBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	faerie := effectstest.B12Push(g, me.ID, "Willow Faerie", "Creature — Faerie", oldB826Faerie, 1, 2)
	ground := effectstest.B12Creature(g, opp.ID, "Ground", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(faerie, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if effectstest.BlockRefusal(t, g, ground, faerie) == "" {
		t.Error("a creature without flying or reach blocked the Faerie")
	}
}

func TestOldB826IhsansShadeCannotBeTargetedByWhiteSpells(t *testing.T) {
	for _, white := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		shade := effectstest.B12Push(g, opp.ID, "Ihsan's Shade", "Legendary Creature — Shade Knight", oldB826Shade, 5, 5)
		card := game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", Colors: []string{"R"}, OracleID: effectstest.LightningBoltOracle}
		pool := "R"
		if white {
			// A white instant that targets a creature: Swords to Plowshares.
			card = game.Card{Name: "Swords to Plowshares", TypeLine: "Instant", ManaCost: "{W}", Colors: []string{"W"}, OracleID: "b1544f21-7e98-461b-aed5-e748b0168c52"}
			pool = "W"
		}
		_, err := oldB826Cast(t, g, card, pool, game.CastSpellParams{Targets: effectstest.CardTarget(shade)})
		if white && err == nil {
			t.Error("a white spell targeted Ihsan's Shade")
		}
		if !white && err != nil {
			t.Errorf("a red spell could not target Ihsan's Shade: %v", err)
		}
		_ = me
	}
}
