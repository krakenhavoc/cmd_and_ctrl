package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b635_cards_test.go — Apocalypse, batch 635 (#5614). Each test plays
// the real catalog entry through the engine and asserts what its text says.

const (
	oldB635PenumbraKavu    = "0abcf934-2552-4885-bbe7-38d93ce0e642"
	oldB635HauntedAngel    = "437be0ad-2dd1-42c8-bf62-5af95228b2f8"
	oldB635PenumbraWurm    = "abccb19f-098b-4453-b7bc-838ad5914ebe"
	oldB635JadedResponse   = "23917517-7561-45b3-b156-77cd2f340466"
	oldB635EvasiveAction   = "4543a99d-eefa-470d-976d-11250524ae28"
	oldB635CoastalDrake    = "d4220231-e9ac-4390-8c7b-979549b26d33"
	oldB635StandardBearer  = "1d6e9716-8c3c-443d-afd8-68d0d6c469a5"
	oldB635DwarvenPatrol   = "4369adc8-d8e0-4d51-b2a9-7e4a8dfd9f8f"
	oldB635Cromat          = "376601b6-fe51-4e2d-8ec6-98f965d649a3"
	oldB635TahngarthsGlare = "390cbc61-6291-4e7d-b262-4de57c827702"
	oldB635Suppress        = "b39651ad-1c05-45f8-b036-536c4d7324c0"
	oldB635IceCave         = "c532ecd2-feeb-41f5-bda0-b05e86f86a45"
)

var oldB635Cards = map[string]string{
	oldB635PenumbraKavu: "Penumbra Kavu", oldB635HauntedAngel: "Haunted Angel", oldB635PenumbraWurm: "Penumbra Wurm",
	oldB635JadedResponse: "Jaded Response", oldB635EvasiveAction: "Evasive Action", oldB635CoastalDrake: "Coastal Drake",
	oldB635StandardBearer: "Standard Bearer", oldB635DwarvenPatrol: "Dwarven Patrol", oldB635Cromat: "Cromat",
	oldB635TahngarthsGlare: "Tahngarth's Glare", oldB635Suppress: "Suppress", oldB635IceCave: "Ice Cave",
}

func TestOldB635EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range oldB635Cards {
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

func oldB635Token(g *game.Game, controller uuid.UUID, name string) (game.Card, int) {
	var found game.Card
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name && c.IsToken() {
			found = c
			n++
		}
	}
	return found, n
}

func TestOldB635PenumbraKavuAndWurmLeaveTheirTokenWhenTheyDie(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, token string
		p, kw                         int
		trample                       bool
	}{
		{"Penumbra Kavu", "Creature — Kavu", oldB635PenumbraKavu, "Kavu", 3, 3, false},
		{"Penumbra Wurm", "Creature — Wurm", oldB635PenumbraWurm, "Wurm", 6, 6, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			victim := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.p, tc.kw)
			src := effectstest.B12Creature(g, opp.ID, "Source", "Creature — Giant", 9, 9)
			effectstest.DamageAndSettle(t, g, src, victim, tc.p)
			if g.Battlefield.Contains(victim) {
				t.Fatal("the creature survived lethal damage")
			}
			tok, n := oldB635Token(g, me.ID, tc.token)
			if n != 1 {
				t.Fatalf("%d %s tokens, want 1", n, tc.token)
			}
			if tok.Power != tc.p || tok.Toughness != tc.kw || !tok.HasColor("B") {
				t.Errorf("the token is %d/%d (colors %v), want %d/%d black", tok.Power, tok.Toughness, tok.Colors, tc.p, tc.kw)
			}
			if tc.trample && !effectstest.HasKeywordOnBattlefield(t, g, tok.InstanceID, "trample") {
				t.Error("the Wurm token has no trample")
			}
		})
	}
}

func TestOldB635HauntedAngelExilesItselfAndGivesEveryOtherPlayerAnAngel(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	angel := effectstest.B12Push(g, me.ID, "Haunted Angel", "Creature — Angel", oldB635HauntedAngel, 3, 3)
	src := effectstest.B12Creature(g, g.Seats[1].ID, "Source", "Creature — Giant", 9, 9)
	effectstest.DamageAndSettle(t, g, src, angel, 3)
	if !g.Exile.Contains(angel) {
		t.Error("the Angel was not exiled")
	}
	for i, p := range g.Seats {
		tok, n := oldB635Token(g, p.ID, "Angel")
		want := 1
		if i == 0 {
			want = 0
		}
		if n != want {
			t.Errorf("seat %d controls %d Angel tokens, want %d", i, n, want)
			continue
		}
		if want == 1 && (tok.Power != 3 || tok.Toughness != 3 || !tok.HasColor("B")) {
			t.Errorf("seat %d's token is %d/%d %v, want 3/3 black", i, tok.Power, tok.Toughness, tok.Colors)
		}
	}
}

// oldB635Respond has seat 0 cast `spell` (a creature spell of the given
// colours), then seat 1 respond with the card named `responder`.
func oldB635Respond(t *testing.T, g *game.Game, colours []string, responder, oracle string, mana string) (game.CastSpellParams, uuid.UUID, *game.Player) {
	t.Helper()
	me, them := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	spell := effectstest.HandCardFull(me, "Their Bear", "Creature — Bear", "", "", colours)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the spell: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority to the responder: %v", err)
	}
	effectstest.FloatForTest(g, them, mana)
	resp := effectstest.HandCardFull(them, responder, "Instant", "", oracle, []string{"U"})
	if err := g.CastSpell(them.ID, resp, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(spell)}); err != nil {
		t.Fatalf("cast %s: %v", responder, err)
	}
	return game.CastSpellParams{}, spell, them
}

func TestOldB635JadedResponseCountersOnlyASpellSharingAColorWithYourCreature(t *testing.T) {
	for _, tc := range []struct {
		name    string
		spell   []string
		creat   []string
		counter bool
	}{
		{"shared red", []string{"R"}, []string{"R"}, true},
		{"one of two shared", []string{"R", "G"}, []string{"G", "W"}, true},
		{"no shared colour", []string{"R"}, []string{"U"}, false},
		{"colourless spell", nil, []string{"R"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.TmpB1Push(g, g.Seats[1].ID, "Their Creature", "Creature — Bear", "", 2, 2, tc.creat...)
			_, spell, _ := oldB635Respond(t, g, tc.spell, "Jaded Response", oldB635JadedResponse, "CU")
			effectstest.PassPriorityAroundTable(t, g)
			if resolved := g.Battlefield.Contains(spell); resolved == tc.counter {
				t.Errorf("the spell resolved = %v, want countered = %v", resolved, tc.counter)
			}
		})
	}
}

func TestOldB635EvasiveActionTaxesByDomain(t *testing.T) {
	for _, tc := range []struct {
		lands []string
		pay   bool
		want  string
	}{
		{nil, false, ""},
		{[]string{"Plains", "Island", "Island"}, false, "{2}"},
		{[]string{"Plains", "Island", "Island"}, true, "{2}"},
	} {
		g := effectstest.NewCatalogGame(t)
		for _, ty := range tc.lands {
			effectstest.PushLandFor(g, g.Seats[1].ID, ty, "Basic Land — "+ty)
		}
		_, spell, _ := oldB635Respond(t, g, []string{"R"}, "Evasive Action", oldB635EvasiveAction, "CU")
		if tc.want == "" {
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(spell) {
				t.Error("with no basic land type the spell was countered")
			}
			continue
		}
		if tc.pay {
			effectstest.FloatForTest(g, g.Seats[0], "CC")
		}
		effectstest.PassUntilTaxed(t, g, g.Seats[0].ID)
		prompt := effectstest.AnswerPayUnless(t, g, g.Seats[0].ID, tc.pay)
		if prompt.PayCost != tc.want {
			t.Errorf("the tax is %q, want %s (two basic land types)", prompt.PayCost, tc.want)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if resolved := g.Battlefield.Contains(spell); resolved != tc.pay {
			t.Errorf("pay=%v: the spell resolved = %v", tc.pay, resolved)
		}
	}
}

func TestOldB635CoastalDrakeBouncesOnlyAKavu(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	drake := effectstest.PushCatalogPermanent(g, me.ID, "Coastal Drake", "Creature — Drake", oldB635CoastalDrake, false)
	kavu := effectstest.B12Creature(g, opp.ID, "Kavu", "Creature — Kavu", 3, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CrtActivateRaw(t, g, me.ID, drake, 0, effectstest.CardTarget(bear), "C", "U"); err == nil {
		t.Fatal("a non-Kavu was accepted as the target")
	}
	me.ManaPool = nil
	if err := effectstest.CrtActivateRaw(t, g, me.ID, drake, 0, effectstest.CardTarget(kavu), "C", "U"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(kavu) || !opp.Hand.Contains(kavu) {
		t.Error("the Kavu was not returned to its owner's hand")
	}
	if !effectstest.TappedNow(g, drake) {
		t.Error("the Drake did not tap for its cost")
	}
}

func TestOldB635StandardBearerForcesAnOpponentsTargetsOntoAFlagbearer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bearer := effectstest.B12Push(g, opp.ID, "Standard Bearer", "Creature — Human Flagbearer", oldB635StandardBearer, 1, 1)
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
	if _, err := b281Cast(t, g, me, "Slaying Fire", "Instant", "{2}{R}", effectstest.SlayingFireOracle, "CCR",
		game.CastSpellParams{Targets: b281Target(other)}); err == nil {
		t.Fatal("an opponent's spell targeted another creature while a Flagbearer was available")
	}
	if _, err := b281Cast(t, g, me, "Slaying Fire", "Instant", "{2}{R}", effectstest.SlayingFireOracle, "CCR",
		game.CastSpellParams{Targets: b281Target(bearer)}); err != nil {
		t.Fatalf("targeting the Flagbearer: %v", err)
	}
}

func TestOldB635DwarvenPatrolUntapsOnlyWhenYouCastANonredSpell(t *testing.T) {
	for _, tc := range []struct {
		name    string
		colours []string
		untaps  bool
	}{{"blue", []string{"U"}, true}, {"colourless", nil, true}, {"red", []string{"R"}, false}, {"red and blue", []string{"R", "U"}, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.ToMainForCost(t, g)
			patrol := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Dwarven Patrol", TypeLine: "Creature — Dwarf", OracleID: oldB635DwarvenPatrol,
				Power: 4, Toughness: 2, Tapped: true, Owner: me.ID, Controller: me.ID,
			})
			spell := effectstest.HandCardFull(me, "Spell", "Instant", "", "", tc.colours)
			if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if tapped := effectstest.TappedNow(g, patrol); tapped == tc.untaps {
				t.Errorf("tapped after the cast = %v, want untapped = %v", tapped, tc.untaps)
			}
		})
	}
	spec, _ := effects.Lookup(oldB635DwarvenPatrol)
	if len(spec.UntapStepRestrictions) != 1 {
		t.Errorf("Dwarven Patrol declares %d untap-step restrictions, want 1", len(spec.UntapStepRestrictions))
	}
}

func TestOldB635CromatsFiveAbilities(t *testing.T) {
	spec, _ := effects.Lookup(oldB635Cromat)
	if len(spec.Activated) != 5 {
		t.Fatalf("Cromat has %d activated abilities, want 5", len(spec.Activated))
	}
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	cromat := effectstest.PushCatalogPermanent(g, me.ID, "Cromat", "Legendary Creature — Illusion", oldB635Cromat, false)
	activate := func(idx int, targets []game.TargetRef, mana ...string) {
		t.Helper()
		if err := effectstest.CrtActivateRaw(t, g, me.ID, cromat, idx, targets, mana...); err != nil {
			t.Fatalf("activate %d: %v", idx, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	activate(1, nil, "U", "R")
	if !effectstest.HasEffectiveKeyword(t, g, cromat, "flying") {
		t.Error("{U}{R} did not give flying")
	}
	activate(3, nil, "R", "W")
	if p, th := effectstest.EffectivePTOf(t, g, cromat); p != 2 || th != 2 {
		t.Errorf("after {R}{W} Cromat is %d/%d, want 2/2 (the test body is a 1/1)", p, th)
	}
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2)
	effectstest.CrtBlock(g, blocker, cromat, me.ID)
	if err := effectstest.CrtActivateRaw(t, g, me.ID, cromat, 0, effectstest.CardTarget(bystander), "W", "B"); err == nil {
		t.Fatal("a creature not in combat with Cromat was accepted")
	}
	me.ManaPool = nil
	activate(0, effectstest.CardTarget(blocker), "W", "B")
	if g.Battlefield.Contains(blocker) || !g.Battlefield.Contains(bystander) {
		t.Error("{W}{B} should destroy exactly the blocker")
	}
	activate(4, nil, "G", "U")
	if g.Battlefield.Contains(cromat) {
		t.Fatal("Cromat is still on the battlefield after {G}{U}")
	}
	if top := effectstest.LibraryTopIDs(me, 1); len(top) != 1 || top[0] != cromat {
		t.Error("Cromat is not on top of its owner's library")
	}
}

func TestOldB635TahngarthsGlareHasEachPlayerOrderTheOthersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "M1", TypeLine: "Land"}, game.Card{Name: "M2", TypeLine: "Land"},
		game.Card{Name: "M3", TypeLine: "Land"}, game.Card{Name: "M4", TypeLine: "Land"})
	effectstest.SeedSearchLibrary(opp,
		game.Card{Name: "T1", TypeLine: "Land"}, game.Card{Name: "T2", TypeLine: "Land"},
		game.Card{Name: "T3", TypeLine: "Land"}, game.Card{Name: "T4", TypeLine: "Land"})
	effectstest.CastCatalogSpell(t, g, "Tahngarth's Glare", "Sorcery", oldB635TahngarthsGlare,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	mineAsked, theirsAsked := effectstest.PutInLibraryChoiceFor(g, me.ID), effectstest.PutInLibraryChoiceFor(g, opp.ID)
	if mineAsked == nil || theirsAsked == nil {
		t.Fatalf("each player should be asked to order three cards (me %v, opponent %v)", mineAsked != nil, theirsAsked != nil)
	}
	if len(mineAsked.ScryCards) != 3 || len(theirsAsked.ScryCards) != 3 {
		t.Fatalf("the prompts hold %d and %d cards, want 3 each", len(mineAsked.ScryCards), len(theirsAsked.ScryCards))
	}
	for _, id := range mineAsked.ScryCards {
		if !opp.Library.Contains(id) {
			t.Error("I was shown a card that is not in the opponent's library")
		}
	}
	for _, id := range theirsAsked.ScryCards {
		if !me.Library.Contains(id) {
			t.Error("the opponent was shown a card that is not in my library")
		}
	}
	reverse := func(ids []uuid.UUID) []uuid.UUID {
		out := make([]uuid.UUID, len(ids))
		for i, id := range ids {
			out[len(ids)-1-i] = id
		}
		return out
	}
	if err := g.ResolvePutInLibrary(mineAsked.ID, me.ID, nil, reverse(mineAsked.ScryCards)); err != nil {
		t.Fatalf("my ordering: %v", err)
	}
	if err := g.ResolvePutInLibrary(theirsAsked.ID, opp.ID, nil, reverse(theirsAsked.ScryCards)); err != nil {
		t.Fatalf("the opponent's ordering: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	top := func(p *game.Player, want []uuid.UUID) bool {
		got := effectstest.LibraryTopIDs(p, len(want))
		for _, id := range want {
			found := false
			for _, g2 := range got {
				found = found || g2 == id
			}
			if !found {
				return false
			}
		}
		return true
	}
	if !top(opp, mineAsked.ScryCards) || !top(me, theirsAsked.ScryCards) {
		t.Error("the looked-at cards are not back on top of their owners' libraries")
	}
	if me.Library.Size() != 4 || opp.Library.Size() != 4 {
		t.Errorf("library sizes %d and %d, want 4 each", me.Library.Size(), opp.Library.Size())
	}
}

func TestOldB635SuppressReturnsTheHandAtTheEndStepOfTheTargetsNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.ToMainForCost(t, g)
	var cards []uuid.UUID
	for i := 0; i < 3; i++ {
		id := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Held", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		cards = append(cards, id)
	}
	hand := opp.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Suppress", "Sorcery", oldB635Suppress, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != 0 {
		t.Fatalf("the target's hand has %d cards after Suppress, want 0", opp.Hand.Size())
	}
	for _, id := range cards {
		if !g.Exile.Contains(id) {
			t.Fatal("a card of the hand was not exiled")
		}
	}
	// Nothing comes back at the end of my own turn.
	for i := 0; i < 60 && !(g.Turn.ActiveSeat == 1 && g.Turn.Step == game.StepEnd); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if !(g.Turn.ActiveSeat == 1 && g.Turn.Step == game.StepEnd) && g.Turn.Step == game.StepEnd {
			effectstest.PassPriorityAroundTable(t, g)
			if g.Turn.ActiveSeat == 0 && opp.Hand.Size() != 0 {
				t.Fatal("the cards returned at my own end step")
			}
		}
	}
	if g.Turn.ActiveSeat != 1 || g.Turn.Step != game.StepEnd {
		t.Fatalf("never reached the opponent's end step (seat %d, %s)", g.Turn.ActiveSeat, g.Turn.Step)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range cards {
		if !opp.Hand.Contains(id) {
			t.Error("a suppressed card did not return to its owner's hand")
		}
	}
	if opp.Hand.Size() < hand {
		t.Errorf("hand %d, want at least %d", opp.Hand.Size(), hand)
	}
}

func oldB635IceCaveSpell(t *testing.T, g *game.Game) (uuid.UUID, uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	cave := effectstest.PushCatalogPermanent(g, me.ID, "Ice Cave", "Enchantment", oldB635IceCave, false)
	effectstest.FloatForTest(g, me, "CR")
	spell := effectstest.HandCardFull(me, "A Bear", "Creature — Bear", "{1}{R}", "", []string{"R"})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	return cave, spell
}

func TestOldB635IceCaveOffersEveryOtherPlayerTheSpellsManaCost(t *testing.T) {
	for _, tc := range []struct {
		name   string
		payer  int // seat index, -1 for nobody
		asked  []int
		counts bool
	}{
		{"nobody pays", -1, []int{1, 2, 3}, false},
		{"the second player pays", 2, []int{1, 2}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			_, spell := oldB635IceCaveSpell(t, g)
			who := uuid.Nil
			if tc.payer >= 0 {
				who = g.Seats[tc.payer].ID
				effectstest.FloatForTest(g, g.Seats[tc.payer], "CR")
			}
			effectstest.PassPriorityAroundTable(t, g)
			asked := effectstest.AppAnswerWith(t, g, effectstest.AppOnly(who))
			if len(asked) != len(tc.asked) {
				t.Fatalf("asked %d players, want %d", len(asked), len(tc.asked))
			}
			for i, seat := range tc.asked {
				if asked[i] != g.Seats[seat].ID {
					t.Errorf("link %d went to the wrong player", i)
				}
			}
			resolved := g.Battlefield.Contains(spell)
			if resolved == tc.counts {
				t.Errorf("the spell resolved = %v, want countered = %v", resolved, tc.counts)
			}
		})
	}
}
