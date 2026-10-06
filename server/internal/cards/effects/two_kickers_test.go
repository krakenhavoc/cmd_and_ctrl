package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// two_kickers_test.go — #2153, CR 702.33b: "Kicker [A] and/or [B]" is
// two kicker abilities, each a separate once-only payment. The engine
// half (validation, record, price, copy) is in
// game/two_kickers_test.go; this file is the declaration, the
// "kicked with its [A] kicker" readers (CR 702.33f) and the cards.

const (
	archangelOfWrathOracle = "022e97af-2a3a-4e13-9b6b-d34fcc8cf168"
	thornscapeOracle       = "8d4d6806-cb01-49e3-91cc-fb0e7f7a8684"
	sunscapeOracle         = "c8fd2232-e499-4889-a2ca-e486ea87d02b"
)

// settleAnsweringPicks resolves everything on the stack, answering
// each of `me`'s pick_target prompts on the way: one that offers
// players is aimed at `victim`, one that offers only cards takes the
// first. Two enters abilities that trigger together are ordered as
// offered (CR 603.3b). Returns how many target prompts were answered —
// one per targeted trigger that fired.
func settleAnsweringPicks(t *testing.T, g *game.Game, me, victim uuid.UUID) int {
	t.Helper()
	asked := 0
	for i := 0; i < 16; i++ {
		passPriorityAroundTable(t, g)
		if c := openTriggerOrder(g, me); c != nil {
			if err := g.ResolveTriggerOrder(c.ID, me, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		p := latestPickTarget(g, me)
		if p == nil {
			if stackFullyEmpty(g) {
				return asked
			}
			continue
		}
		pick := game.TargetRef{Kind: game.TargetPlayer, ID: victim}
		if len(p.PickTargetPlayers) == 0 {
			if len(p.PickTargetCards) == 0 {
				t.Fatalf("pick_target prompt %q offers nothing", p.Reason)
			}
			pick = game.TargetRef{Kind: game.TargetCard, ID: p.PickTargetCards[0]}
		}
		if err := g.ResolvePickTarget(p.ID, me, pick); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
		asked++
	}
	t.Fatal("the stack did not settle")
	return asked
}

func openTriggerOrder(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerOrder && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func seedArtifactFor(g *game.Game, controller uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bauble", TypeLine: "Artifact",
		Owner: controller, Controller: controller,
	})
}

// TestKickersDeclaresTwoKickerCosts pins the constructor: two entries,
// printed order, both keyed kicker and paid at most once — not a
// multikicker relabelled, which is what Urborg Lhurgoyf used to be.
func TestKickersDeclaresTwoKickerCosts(t *testing.T) {
	for _, oracle := range []string{urborgLhurgoyfOracle, thornscapeOracle, sunscapeOracle, archangelOfWrathOracle} {
		costs := game.OptionalCostsFor(oracle)
		if len(costs) != 2 {
			t.Fatalf("%s: %d optional costs, want 2", oracle, len(costs))
		}
		for i, oc := range costs {
			if oc.Key != game.KickerKey || oc.MaxPayments() != 1 || oc.Label != "Kicker "+oc.ManaCost {
				t.Errorf("%s cost %d = %+v, want a once-only kicker labelled by its mana", oracle, i, oc)
			}
		}
	}
	if got := game.OptionalCostsFor(urborgLhurgoyfOracle); got[0].ManaCost != "{U}" || got[1].ManaCost != "{B}" {
		t.Errorf("Urborg Lhurgoyf's kickers = %q, %q; want {U} then {B}, as printed", got[0].ManaCost, got[1].ManaCost)
	}
}

// TestRegisterHoldsKickersToThePrintedShapes: two kickers are fine;
// a third, two with the same mana, or a manaless second are not — and
// every other key is still unique per card.
func TestRegisterHoldsKickersToThePrintedShapes(t *testing.T) {
	registerForTest(t, Spec{OracleID: "test-two-kickers-ok", Name: "Two Kickers", OptionalCosts: Kickers("{R}", "{W}")})

	mustPanic(t, "at most two", func() {
		Register(Spec{OracleID: "test-three-kickers", Name: "Three Kickers",
			OptionalCosts: append(Kickers("{R}", "{W}"), Kicker("{G}"))})
	})
	mustPanic(t, "different mana cost", func() {
		Register(Spec{OracleID: "test-same-kickers", Name: "Same Kickers", OptionalCosts: Kickers("{R}", "{R}")})
	})
	mustPanic(t, "different mana cost", func() {
		Register(Spec{OracleID: "test-manaless-kicker", Name: "Manaless Kicker",
			OptionalCosts: []game.AdditionalCost{Kicker("{R}"), KickerSacrifice("a creature", Creature())}})
	})
	mustPanic(t, "two optional costs keyed", func() {
		Register(Spec{OracleID: "test-two-buybacks", Name: "Two Buybacks",
			OptionalCosts: []game.AdditionalCost{Buyback("{2}"), Buyback("{3}")}})
	})
}

// TestKickedWithReadsWhichKickerPaidAndSurvivesACopy is CR 702.33f on
// a SPELL, through ctx.KickedWith, and CR 707.10b: a Reverberate copy
// of a spell kicked with its second kicker was kicked with that kicker
// too, and only that one.
func TestKickedWithReadsWhichKickerPaidAndSurvivesACopy(t *testing.T) {
	type read struct {
		r, u, unknown bool
		times         int
	}
	var reads []read
	const fixture = "test-two-kicker-sorcery"
	registerForTest(t, Spec{
		OracleID:      fixture,
		Name:          "Two-Kicker Sorcery",
		OptionalCosts: Kickers("{R}", "{U}"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			reads = append(reads, read{ctx.KickedWith("{R}"), ctx.KickedWith("{U}"), ctx.KickedWith("{G}"), ctx.KickedTimes()})
			return nil
		},
	})

	for _, tc := range []struct {
		name     string
		optional []int
		want     read
	}{
		{"neither", nil, read{}},
		{"first", []int{0}, read{r: true, times: 1}},
		{"second", []int{1}, read{u: true, times: 1}},
		{"both", []int{0, 1}, read{r: true, u: true, times: 2}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reads = nil
			g := newCatalogGame(t)
			spell, err := castWithOptionalCosts(t, g, "Two-Kicker Sorcery", "Sorcery", fixture, nil, tc.optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			castCatalogSpell(t, g, "Reverberate", "Instant", reverberateOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
			passPriorityAroundTable(t, g)
			if len(reads) != 2 {
				t.Fatalf("resolved %d times, want the copy and the original", len(reads))
			}
			for i, got := range reads {
				if got != tc.want {
					t.Errorf("resolution %d read %+v, want %+v", i, got, tc.want)
				}
			}
		})
	}
}

// TestThornscapeBattlemageEachKickerTriggersItsOwnAbility: kicked with
// {R} it pings, with {W} it destroys an artifact, with both it does
// both, and unkicked it triggers nothing (CR 603.4).
func TestThornscapeBattlemageEachKickerTriggersItsOwnAbility(t *testing.T) {
	for _, tc := range []struct {
		name      string
		optional  []int
		damage    int
		destroyed bool
	}{
		{"unkicked", nil, 0, false},
		{"kicked with {R}", []int{0}, 2, false},
		{"kicked with {W}", []int{1}, 0, true},
		{"kicked with both", []int{0, 1}, 2, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
			bauble := seedArtifactFor(g, victim.ID)
			life := victim.Life
			if _, err := castWithOptionalCosts(t, g, "Thornscape Battlemage", "Creature — Elf Wizard",
				thornscapeOracle, nil, tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			asked := settleAnsweringPicks(t, g, me.ID, victim.ID)
			if asked != len(tc.optional) {
				t.Errorf("answered %d target prompts, want %d — one per kicker paid", asked, len(tc.optional))
			}
			if got := life - victim.Life; got != tc.damage {
				t.Errorf("victim lost %d life, want %d", got, tc.damage)
			}
			if gone := findBattlefieldCardByID(g, bauble) == nil; gone != tc.destroyed {
				t.Errorf("artifact destroyed = %v, want %v", gone, tc.destroyed)
			}
		})
	}
}

// TestArchangelOfWrathKickedTwiceIsBothKickers: "if it was kicked"
// fires for either kicker, "if it was kicked twice" only for both
// (CR 702.33d), and lifelink gains what the Archangel deals.
func TestArchangelOfWrathKickedTwiceIsBothKickers(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		damage   int
	}{
		{"unkicked", nil, 0},
		{"kicked with {B}", []int{0}, 2},
		{"kicked with {R}", []int{1}, 2},
		{"kicked twice", []int{0, 1}, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
			life, mine := victim.Life, me.Life
			if _, err := castWithOptionalCosts(t, g, "Archangel of Wrath", "Creature — Angel",
				archangelOfWrathOracle, nil, tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			settleAnsweringPicks(t, g, me.ID, victim.ID)
			if got := life - victim.Life; got != tc.damage {
				t.Errorf("victim lost %d life, want %d", got, tc.damage)
			}
			if got := me.Life - mine; got != tc.damage {
				t.Errorf("caster gained %d life, want %d (lifelink, CR 702.15b)", got, tc.damage)
			}
		})
	}
}

// TestSunscapeBattlemageMultiSymbolKickers: the same linked reads with
// kickers of more than one symbol.
func TestSunscapeBattlemageMultiSymbolKickers(t *testing.T) {
	for _, tc := range []struct {
		name      string
		optional  []int
		drawn     int
		destroyed bool
	}{
		{"unkicked", nil, 0, false},
		{"kicked with {1}{G}", []int{0}, 0, true},
		{"kicked with {2}{U}", []int{1}, 2, false},
		{"kicked with both", []int{0, 1}, 2, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
			flyer := pushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 1,
				Keywords: []string{"flying"}, Owner: victim.ID, Controller: victim.ID,
			})
			if _, err := castWithOptionalCosts(t, g, "Sunscape Battlemage", "Creature — Human Wizard",
				sunscapeOracle, nil, tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			hand := me.Hand.Size()
			settleAnsweringPicks(t, g, me.ID, victim.ID)
			if got := me.Hand.Size() - hand; got != tc.drawn {
				t.Errorf("drew %d, want %d", got, tc.drawn)
			}
			if gone := findBattlefieldCardByID(g, flyer) == nil; gone != tc.destroyed {
				t.Errorf("flyer destroyed = %v, want %v", gone, tc.destroyed)
			}
		})
	}
}

// TestATokenCopyOfAThornscapeKeepsWhichKickerPaid is CR 707.10b /
// CR 400.7d on a permanent spell: Double Major on a Thornscape
// Battlemage kicked with {W} only makes a token whose {W} ability
// triggers and whose {R} ability does not.
func TestATokenCopyOfAThornscapeKeepsWhichKickerPaid(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	seedArtifactFor(g, victim.ID)
	seedArtifactFor(g, victim.ID)
	life := victim.Life

	spell, err := castWithOptionalCosts(t, g, "Thornscape Battlemage", "Creature — Elf Wizard",
		thornscapeOracle, nil, []int{1}, nil)
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	castCatalogSpell(t, g, "Double Major", "Instant", oracleDoubleMajor,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	if asked := settleAnsweringPicks(t, g, me.ID, victim.ID); asked != 2 {
		t.Errorf("answered %d target prompts, want 2 — the {W} ability of the card and of its copy", asked)
	}

	tokens, cards := copiesOnBattlefield(g, "Thornscape Battlemage")
	if len(tokens) != 1 || len(cards) != 1 {
		t.Fatalf("got %d token copies and %d cards, want one of each", len(tokens), len(cards))
	}
	if !game.CardKickedWith(tokens[0], "{W}") || game.CardKickedWith(tokens[0], "{R}") {
		t.Errorf("token kicked with {W}=%v {R}=%v, want true and false",
			game.CardKickedWith(tokens[0], "{W}"), game.CardKickedWith(tokens[0], "{R}"))
	}
	artifacts := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bauble" {
			artifacts++
		}
	}
	if artifacts != 0 {
		t.Errorf("%d artifacts survive, want 0 — the card and its copy each destroy one", artifacts)
	}
	if victim.Life != life {
		t.Errorf("victim lost %d life; the {R} ability was not paid for", life-victim.Life)
	}
}
