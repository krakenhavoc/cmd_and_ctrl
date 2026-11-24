package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b016_cards_test.go — Secret Lair Drop, "Instants and sorceries"
// (S82 OLD batch 16, #5462): each card driven through a real cast.

const (
	oldB016VoyageOracle   = "00f82f84-81a5-45e5-9c35-9d627a180950"
	oldB016BalanceOracle  = "17fa98cd-ed8f-483f-9525-7e989a82ebb2"
	oldB016ChanceOracle   = "324af60d-142b-4b6c-afe9-6aa36267d5ee"
	oldB016GraspOracle    = "6d56aeb1-0a50-46b6-abdb-cd6575a98dc3"
	oldB016SymmetryOracle = "6dfb50d7-359b-4644-90e3-7ecdacee7c12"
	oldB016ByForceOracle  = "7ce950fb-e065-45d3-aeff-f8c475460754"
	oldB016ClaimOracle    = "9bb82829-05e8-4cdd-91e4-b837a620bfc8"
	oldB016EntranceOracle = "dda8bf18-9396-45e6-829c-8691f49f7290"
	oldB016CacophonyOrcl  = "e0f0f290-c83f-4d97-9f5d-786c7136b0d1"
	oldB016GustOracle     = "ecf1e378-d6bc-4ee0-800e-eb8a955d781f"
	oldB016WarpOracle     = "50a228a2-c8b6-4416-b0ab-417926a9b9b6"
)

func TestS82OldB016AllAreFullWithNoCaveats(t *testing.T) {
	for _, id := range []string{oldB016VoyageOracle, oldB016BalanceOracle, oldB016ChanceOracle, oldB016GraspOracle,
		oldB016SymmetryOracle, oldB016ByForceOracle, oldB016ClaimOracle, oldB016EntranceOracle, oldB016CacophonyOrcl, oldB016GustOracle, oldB016WarpOracle} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Balance -----------------------------------------------------------------

func TestS82OldB016BalanceEvensLandsCardsAndCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	// Everyone else has more of everything than the opponent, who has the fewest.
	for i := 0; i < 3; i++ {
		effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
		effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	}
	effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, p := range g.Seats[2:] {
		effectstest.B12Permanent(g, p.ID, "Forest", "Basic Land — Forest")
		effectstest.B12Creature(g, p.ID, "Bear", "Creature — Bear", 2, 2)
	}
	for len(me.Hand.Cards) < 5 {
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	for len(opp.Hand.Cards) > 1 {
		opp.Hand.Cards = opp.Hand.Cards[:len(opp.Hand.Cards)-1]
	}
	for _, p := range g.Seats[2:] {
		for len(p.Hand.Cards) > 2 {
			p.Hand.Cards = p.Hand.Cards[:len(p.Hand.Cards)-1]
		}
		for len(p.Hand.Cards) < 2 {
			p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
		}
	}
	effectstest.CastCatalogSpell(t, g, "Balance", "Sorcery", oldB016BalanceOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 60; i++ {
		answered := false
		for _, seat := range g.Seats {
			c := effectstest.LatestChooseCards(g, seat.ID)
			if c == nil || len(c.ChooseCards) == 0 {
				continue
			}
			n := c.ChooseMin
			if n == 0 {
				n = 1
			}
			if n > len(c.ChooseCards) {
				n = len(c.ChooseCards)
			}
			if err := g.ResolveChooseCards(c.ID, seat.ID, c.ChooseCards[:n]); err != nil {
				t.Fatalf("ResolveChooseCards: %v", err)
			}
			answered = true
			break
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !answered && effectstest.StackFullyEmpty(g) {
			break
		}
	}
	lands := func(p *game.Player) int { return effectsCount(g, p.ID, game.Card.IsLand) }
	creatures := func(p *game.Player) int { return effectsCount(g, p.ID, game.Card.IsCreature) }
	for _, p := range g.Seats {
		if lands(p) != 1 || creatures(p) != 1 {
			t.Errorf("seat %v ends with %d lands and %d creatures, want 1 and 1", p.ID, lands(p), creatures(p))
		}
	}
	if len(me.Hand.Cards) != len(opp.Hand.Cards) {
		t.Errorf("hands %d and %d, want equal", len(me.Hand.Cards), len(opp.Hand.Cards))
	}
}

// effectsCount counts the battlefield cards a player controls that match.
func effectsCount(g *game.Game, player uuid.UUID, match func(game.Card) bool) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == player && match(c) {
			n++
		}
	}
	return n
}

// --- Chance for Glory --------------------------------------------------------

func TestS82OldB016ChanceForGloryGrantsIndestructibleAndTakesAnExtraTurnThenLoses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, g.Seats[me].ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, g.Seats[(me+1)%len(g.Seats)].ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Chance for Glory", "Instant", oldB016ChanceOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, mine, "indestructible") {
		t.Error("my creature did not gain indestructible")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, theirs, "indestructible") {
		t.Error("an opponent's creature gained indestructible")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Seats[me].Eliminated {
		t.Fatal("lost in the end step of the turn the spell resolved in")
	}
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, me, true, "the extra turn")
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Seats[me].Eliminated {
		t.Fatal("did not lose at the beginning of the extra turn's end step")
	}
}

// --- Dramatic Entrance -------------------------------------------------------

func TestS82OldB016DramaticEntrancePutsAGreenCreatureFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	green := effectstest.HandCardFull(me, "Big Green", "Creature — Beast", "{5}{G}", "", []string{"G"})
	red := effectstest.HandCardFull(me, "Big Red", "Creature — Beast", "{5}{R}", "", []string{"R"})
	effectstest.CastCatalogSpell(t, g, "Dramatic Entrance", "Instant", oldB016EntranceOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID); p != nil {
		if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{green}); err != nil {
			t.Fatalf("pick: %v", err)
		}
	} else {
		effectstest.AnswerMayChoice(t, g, me.ID, true)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(green) {
		t.Error("the green creature card did not enter")
	}
	if g.Battlefield.Contains(red) {
		t.Error("the red creature card entered")
	}
}

// --- Claim the Firstborn -----------------------------------------------------

func TestS82OldB016ClaimTheFirstbornStealsACheapCreatureUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	cheap := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Cheap", TypeLine: "Creature — Bear",
		ManaCost: "{2}", Power: 2, Toughness: 2, Tapped: true, Owner: opp.ID, Controller: opp.ID})
	dear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Dear", TypeLine: "Creature — Bear",
		ManaCost: "{4}", Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID})
	if err := effectstest.CastCatalogSpellErr(t, g, "Claim the Firstborn", "Sorcery", oldB016ClaimOracle, effectstest.B16TargetCard(dear)); err == nil {
		t.Fatal("targeted a creature with mana value 4")
	}
	effectstest.CastCatalogSpell(t, g, "Claim the Firstborn", "Sorcery", oldB016ClaimOracle, effectstest.B16TargetCard(cheap))
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := effectstest.ApaLive(g, cheap)
	if c.Controller != me.ID || c.Tapped {
		t.Errorf("controller %v tapped %v, want mine and untapped", c.Controller, c.Tapped)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, cheap, "haste") {
		t.Error("the creature lacks haste")
	}
	effectstest.EndTurn(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if got := effectstest.ApaLive(g, cheap).Controller; got != opp.ID {
		t.Errorf("after the turn the creature is controlled by %v, want its owner", got)
	}
}

// --- By Force ----------------------------------------------------------------

func TestS82OldB016ByForceDestroysXTargetArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Permanent(g, opp.ID, "Rock A", "Artifact")
	b := effectstest.B12Permanent(g, me.ID, "Rock B", "Artifact")
	c := effectstest.B12Permanent(g, opp.ID, "Rock C", "Artifact")
	creature := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.Pr8CastX(t, g, "By Force", "Sorcery", oldB016ByForceOracle, 2, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("a targeted artifact survived")
	}
	if !g.Battlefield.Contains(c) || !g.Battlefield.Contains(creature) {
		t.Error("an untargeted permanent was destroyed")
	}
}

// --- Maddening Cacophony -----------------------------------------------------

func TestS82OldB016MaddeningCacophonyMillsEightOrHalfWhenKicked(t *testing.T) {
	for _, tc := range []struct {
		name   string
		kicked bool
	}{{"unkicked", false}, {"kicked", true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			effectstest.AuraSeedLibrary(opp, 20)
			size := opp.Library.Size()
			mine := me.Library.Size()
			params := game.CastSpellParams{}
			cost := "{C}{U}"
			if tc.kicked {
				params.OptionalCosts = []int{0}
				cost = "{C}{C}{C}{C}{U}{U}"
			}
			id := uuid.New()
			me.Hand.PushTop(game.Card{InstanceID: id, Name: "Maddening Cacophony", TypeLine: "Sorcery", OracleID: oldB016CacophonyOrcl,
				ManaCost: "{1}{U}", Colors: []string{"U"}, Owner: me.ID, Controller: me.ID})
			effectstest.AddPool(t, g, me.ID, cost)
			if err := g.CastSpell(me.ID, id, params); err != nil {
				t.Fatalf("cast: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			want := 8
			if tc.kicked {
				want = (size + 1) / 2
			}
			if got := size - opp.Library.Size(); got != want {
				t.Errorf("the opponent milled %d, want %d", got, want)
			}
			if me.Library.Size() != mine {
				t.Errorf("the caster milled")
			}
		})
	}
}

// --- Aether Gust -------------------------------------------------------------

func TestS82OldB016AetherGustLetsTheOwnerPickTopOrBottom(t *testing.T) {
	for _, tc := range []struct {
		name   string
		bottom bool
	}{{"top", false}, {"bottom", true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			opp := g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			red := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Red", TypeLine: "Creature — Bear",
				Power: 2, Toughness: 2, Colors: []string{"R"}, Owner: opp.ID, Controller: opp.ID})
			blue := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Blue", TypeLine: "Creature — Bear",
				Power: 2, Toughness: 2, Colors: []string{"U"}, Owner: opp.ID, Controller: opp.ID})
			if err := effectstest.CastCatalogSpellErr(t, g, "Aether Gust", "Instant", oldB016GustOracle, effectstest.B16TargetCard(blue)); err == nil {
				t.Fatal("targeted a blue permanent")
			}
			lib := opp.Library.Size()
			effectstest.CastCatalogSpell(t, g, "Aether Gust", "Instant", oldB016GustOracle, effectstest.B16TargetCard(red))
			effectstest.PassPriorityAroundTable(t, g)
			var pick *game.PendingChoice
			for _, c := range g.PendingChoices {
				if c != nil && c.Chooser == opp.ID && c.Kind == game.PendingChoicePutInLibrary {
					pick = c
				}
			}
			if pick == nil {
				t.Fatalf("the owner was not asked where to put it (pending: %d)", len(g.PendingChoices))
			}
			var bottom, top []uuid.UUID
			if tc.bottom {
				bottom = []uuid.UUID{red}
			} else {
				top = []uuid.UUID{red}
			}
			if err := g.ResolvePutInLibrary(pick.ID, opp.ID, bottom, top); err != nil {
				t.Fatalf("answer: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(red) || opp.Library.Size() != lib+1 {
				t.Fatalf("the permanent is still on the battlefield or the library is %d, want %d", opp.Library.Size(), lib+1)
			}
			cards := opp.Library.Cards
			if tc.bottom && cards[0].InstanceID != red {
				t.Error("the card is not on the bottom of the library")
			}
			if !tc.bottom && cards[len(cards)-1].InstanceID != red {
				t.Error("the card is not on top of the library")
			}
			if effectstest.ZoneOf(g, red) != game.ZoneLibrary {
				t.Errorf("the card is in %q, want the library", effectstest.ZoneOf(g, red))
			}
		})
	}
}

func TestS82OldB016AetherGustCanTargetARedSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	bolt := effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: bear})
	lib := opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Aether Gust", "Instant", oldB016GustOracle, effectstest.B16TargetCard(bolt))
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == opp.ID && c.Kind == game.PendingChoicePutInLibrary {
			if err := g.ResolvePutInLibrary(c.ID, opp.ID, nil, []uuid.UUID{bolt}); err != nil {
				t.Fatalf("answer: %v", err)
			}
			break
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOf(g, bolt) != game.ZoneLibrary || opp.Library.Size() != lib+1 {
		t.Errorf("the Bolt is in %q, library %d (want library %d)", effectstest.ZoneOf(g, bolt), opp.Library.Size(), lib+1)
	}
	if c, _ := effectstest.BattlefieldCardByID(g, bear); c.DamageMarked != 0 {
		t.Error("the Bolt still resolved")
	}
}

// --- Collective Voyage -------------------------------------------------------

func TestS82OldB016CollectiveVoyageSearchesForTheTotalPaidInBasicLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	s := g.Seats
	lands := make([][]uuid.UUID, len(s))
	for i, p := range s {
		effectstest.StockLibraryFor(g, p, 10)
		for j := 0; j < 4; j++ {
			lands[i] = append(lands[i], effectstest.SeedLibraryTop(p, "Plains", "Basic Land — Plains", ""))
		}
	}
	effectstest.CastCatalogSpell(t, g, "Collective Voyage", "Sorcery", oldB016VoyageOracle, nil)
	effectstest.FillPool(s[0], 2)
	effectstest.FillPool(s[1], 1)
	effectstest.PassPriorityAroundTable(t, g)
	// Seat 0 pays 2 and seat 1 pays 1; everyone else keeps their mana (they have none).
	for i := 0; i < 24; i++ {
		var next *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && (c.Kind == game.PendingChoicePayAmount || c.Kind == game.PendingChoicePayUnless) {
				next = c
				break
			}
		}
		if next == nil {
			break
		}
		amounts := map[uuid.UUID]int{s[0].ID: 2, s[1].ID: 1}
		if next.Kind == game.PendingChoicePayAmount {
			if err := g.ResolvePayAmount(next.ID, next.Chooser, amounts[next.Chooser]); err != nil {
				t.Fatalf("pay amount: %v", err)
			}
		} else if err := g.ResolvePayUnless(next.ID, next.Chooser, true); err != nil {
			t.Fatalf("pay: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	// Each player searches for up to three basic lands, caster first.
	for i, p := range s {
		if c := effectstest.SearchChoiceFor(g, p.ID); c != nil {
			effectstest.AnswerSearchByID(t, g, p.ID, lands[i][:3]...)
			effectstest.PassPriorityAroundTable(t, g)
		} else {
			t.Errorf("seat %d was not asked to search", i)
		}
	}
	for i, p := range s {
		n, tapped := 0, 0
		for _, c := range g.Battlefield.Cards {
			if c.Controller == p.ID && c.Name == "Plains" {
				n++
				if c.Tapped {
					tapped++
				}
			}
		}
		if n != 3 || tapped != 3 {
			t.Errorf("seat %d has %d Plains (%d tapped), want 3 tapped", i, n, tapped)
		}
	}
}

// --- Praetor's Grasp ---------------------------------------------------------

func TestS82OldB016PraetorsGraspExilesACardFaceDownAndLetsYouPlayIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	want := effectstest.SeedLibraryTop(opp, "Juicy Bear", "Creature — Bear", "{1}{G}")
	effectstest.SeedLibraryTop(opp, "Filler", "Sorcery", "{1}")
	effectstest.CastCatalogSpell(t, g, "Praetor's Grasp", "Sorcery", oldB016GraspOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, want)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOf(g, want) != game.ZoneExile {
		t.Fatalf("the card is in %q, want exile", effectstest.ZoneOf(g, want))
	}
	effectstest.AddPool(t, g, me.ID, "{C}{G}")
	if err := g.CastSpell(me.ID, want, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCardByID(g, want)
	if !ok || c.Controller != me.ID {
		t.Errorf("the card is on the battlefield %v under %v, want under me", ok, c.Controller)
	}
}

// --- Scheming Symmetry -------------------------------------------------------

func TestS82OldB016SchemingSymmetryLetsTwoPlayersTutorToTheTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	a, b := g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	wantA := effectstest.SeedLibraryTop(a, "A's Pick", "Sorcery", "{1}")
	effectstest.SeedLibraryTop(a, "Junk A", "Sorcery", "{1}")
	wantB := effectstest.SeedLibraryTop(b, "B's Pick", "Sorcery", "{1}")
	effectstest.SeedLibraryTop(b, "Junk B", "Sorcery", "{1}")
	sizeA, sizeB := a.Library.Size(), b.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Scheming Symmetry", "Sorcery", oldB016SymmetryOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: a.ID}, {Kind: game.TargetPlayer, ID: b.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, a.ID, wantA)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, b.ID, wantB)
	effectstest.PassPriorityAroundTable(t, g)
	if a.Library.Size() != sizeA || b.Library.Size() != sizeB {
		t.Errorf("library sizes %d/%d, want %d/%d", a.Library.Size(), b.Library.Size(), sizeA, sizeB)
	}
	for name, tc := range map[string]struct {
		p    *game.Player
		want uuid.UUID
	}{"A": {a, wantA}, "B": {b, wantB}} {
		if top, _ := tc.p.Library.Top(); top.InstanceID != tc.want {
			t.Errorf("%s's pick is not on top of their library (top is %q)", name, top.Name)
		}
	}
}

// --- Warp World --------------------------------------------------------------

func TestS82OldB016WarpWorldReplacesEachPlayersPermanentsWithTheTopOfTheirLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, enchanter, sorcerer := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	fill := func(p *game.Player, n int, name, typeLine string) {
		for i := 0; i < n; i++ {
			p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
		}
	}
	p0 := me.Library.Size()
	fill(me, 10, "Zed", "Creature — Bear")
	fill(enchanter, 10, "Charm", "Enchantment")
	fill(sorcerer, 10, "Spell", "Sorcery")
	effectstest.B12Permanent(g, me.ID, "My Plains", "Basic Land — Plains")
	myBear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Token", TypeLine: "Token Creature — Goblin",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.B12Permanent(g, enchanter.ID, "Their Rock", "Artifact")
	effectstest.B12Creature(g, sorcerer.ID, "Their Bear", "Creature — Bear", 2, 2)
	e0, s0 := enchanter.Library.Size(), sorcerer.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Warp World", "Sorcery", oldB016WarpOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 10 && len(g.PendingChoices) > 0; i++ {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoicePutInLibrary {
				if err := g.ResolvePutInLibrary(c.ID, c.Chooser, c.ScryCards, nil); err != nil {
					t.Fatalf("order: %v", err)
				}
				break
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	_ = myBear
	mine := effectsCount(g, me.ID, func(c game.Card) bool { return true })
	if mine != 3 {
		t.Errorf("I control %d permanents, want 3 (two cards and a token's worth, all revealed creatures)", mine)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() {
			t.Error("my token came back: it ceased to exist in the library")
		}
	}
	if got := me.Library.Size(); got != p0+10-3+2 {
		t.Errorf("my library is %d, want %d", got, p0+10-3+2)
	}
	if n := effectsCount(g, enchanter.ID, func(c game.Card) bool { return true }); n != 1 {
		t.Errorf("the enchantment player controls %d permanents, want 1", n)
	}
	if got := enchanter.Library.Size() + effectsCount(g, enchanter.ID, func(c game.Card) bool { return true }); got != e0+1 {
		t.Errorf("the enchantment player's cards are not conserved: library+battlefield = %d", got)
	}
	if n := effectsCount(g, sorcerer.ID, func(c game.Card) bool { return c.TypeLine == "Sorcery" }); n != 0 {
		t.Errorf("a sorcery card entered the battlefield")
	}
	if got := sorcerer.Library.Size() + effectsCount(g, sorcerer.ID, func(c game.Card) bool { return true }); got != s0+1 {
		t.Errorf("the sorcery player's cards are not conserved: library+battlefield = %d, want %d", got, s0+1)
	}
}
