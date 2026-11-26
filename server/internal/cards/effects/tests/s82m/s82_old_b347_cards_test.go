package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b347_cards_test.go — Khans of Tarkir, batch 347 (S82 older sets,
// #5554): the ferocious cards. Every test plays the real catalog entry through
// the engine, once with a power-4 creature on the table and once without.

const (
	oldB347DragonGrip = "1e83a750-bf86-468e-8342-2c4eb7700dfd"
	oldB347Heir       = "35cb88aa-6ed9-4a2d-9bbc-5937e1df4f4a"
	oldB347ForceAway  = "4770ec88-bc7b-457e-8f67-ae46e2cf4caa"
	oldB347Barrage    = "71a756f3-cd1e-4047-90ac-eda096b90647"
	oldB347Unwritten  = "b0060b32-b856-4ff7-a113-cec2dc8f8dc6"
	oldB347Claws      = "b8cf8614-14f4-4c20-9697-2e5ea34e0039"
	oldB347IcyBlast   = "f49302c5-8510-4360-841c-a59f53f87e0b"
	oldB347FeedClan   = "f70f1f2e-7277-4bc9-b0bb-b44a1da2c62f"
)

func TestOldB347EveryCardIsRegisteredFull(t *testing.T) {
	for name, oracle := range map[string]string{
		"Dragon Grip": oldB347DragonGrip, "Heir of the Wilds": oldB347Heir, "Force Away": oldB347ForceAway,
		"Barrage of Boulders": oldB347Barrage, "See the Unwritten": oldB347Unwritten, "Crater's Claws": oldB347Claws,
		"Icy Blast": oldB347IcyBlast, "Feed the Clan": oldB347FeedClan,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not in the catalog", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q completeness %v caveats %d, want Full with none", name, spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// oldB347Table seeds a main-phase table; with ferocious it also puts a 4/4 on
// the active player's side.
func oldB347Table(t *testing.T, ferocious bool) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if ferocious {
		effectstest.B12Creature(g, me.ID, "Big Bear", "Creature — Bear", 4, 4)
	} else {
		effectstest.B12Creature(g, me.ID, "Small Bear", "Creature — Bear", 3, 3)
	}
	effectstest.AdvanceToMain(t, g)
	return g, me, opp
}

func TestOldB347DragonGripGivesPlusTwoAndFirstStrike(t *testing.T) {
	g, me, _ := oldB347Table(t, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastFromHandForTest(t, g, me, "Dragon Grip", "Enchantment — Aura", "", oldB347DragonGrip, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.MkmB1PT(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("enchanted bear = %d/%d, want 4/2", p, tough)
	}
	effectstest.AssertKeywords(t, g, bear, "first strike")
}

func TestOldB347DragonGripHasFlashOnlyWithFerocious(t *testing.T) {
	for _, ferocious := range []bool{false, true} {
		g, me, opp := oldB347Table(t, ferocious)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		// An opponent's turn: only flash can cast it.
		effectstest.AdvanceToNextSeatsTurn(t, g)
		effectstest.AdvanceToMainOf(t, g, (g.Turn.ActiveSeat))
		if g.Seats[g.Turn.ActiveSeat].ID != opp.ID {
			t.Fatalf("expected the opponent's turn")
		}
		_, err := effectstest.CastFromHandErr(g, me, "Dragon Grip", "Enchantment — Aura", "", oldB347DragonGrip, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(bear)})
		if ferocious && err != nil {
			t.Errorf("ferocious: the Aura could not be cast with flash: %v", err)
		}
		if !ferocious && err == nil {
			t.Error("without ferocious the Aura was cast at instant speed")
		}
	}
}

func TestOldB347HeirOfTheWildsGetsPlusOneOnlyWithFerocious(t *testing.T) {
	for _, ferocious := range []bool{false, true} {
		g, me, opp := oldB347Table(t, ferocious)
		heir := effectstest.B12Push(g, me.ID, "Heir of the Wilds", "Creature — Human Warrior", oldB347Heir, 2, 2)
		effectstest.DeclareAttack(t, g, opp.ID, heir)
		effectstest.PassPriorityAroundTable(t, g)
		want := 2
		if ferocious {
			want = 3
		}
		if p, _ := effectstest.MkmB1PT(t, g, heir); p != want {
			t.Errorf("ferocious=%v: power %d, want %d", ferocious, p, want)
		}
	}
}

func TestOldB347ForceAwayBouncesAndLootsOnlyWithFerocious(t *testing.T) {
	for _, ferocious := range []bool{false, true} {
		g, me, opp := oldB347Table(t, ferocious)
		theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		effectstest.CastFromHandForTest(t, g, me, "Force Away", "Instant", "", oldB347ForceAway, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(theirs)})
		effectstest.PassPriorityAroundTable(t, g)
		if !opp.Hand.Contains(theirs) {
			t.Fatalf("ferocious=%v: the creature was not returned to its owner's hand", ferocious)
		}
		asked := effectstest.LatestConfirmFor(g, me.ID) != nil
		if asked != ferocious {
			t.Errorf("ferocious=%v: loot offered = %v", ferocious, asked)
		}
		if ferocious {
			before := me.Hand.Size()
			effectstest.AnswerMayChoice(t, g, me.ID, true)
			if me.Hand.Size() != before+1 {
				t.Errorf("hand %d → %d, want one card drawn", before, me.Hand.Size())
			}
		}
	}
}

func TestOldB347BarrageOfBouldersPingsOpponentsAndStopsBlockingWithFerocious(t *testing.T) {
	for _, ferocious := range []bool{false, true} {
		g, me, opp := oldB347Table(t, ferocious)
		mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		effectstest.CastFromHandForTest(t, g, me, "Barrage of Boulders", "Sorcery", "", oldB347Barrage, game.CastSpellParams{Strict: true})
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.Pr6Marked(g, theirs) != 1 || effectstest.Pr6Marked(g, mine) != 0 {
			t.Errorf("ferocious=%v: damage on theirs %d / mine %d, want 1 / 0", ferocious, effectstest.Pr6Marked(g, theirs), effectstest.Pr6Marked(g, mine))
		}
		g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
		cantBlock := effectstest.FindBattlefieldCardForTest(g, theirs).Effective().Restrictions&game.CantBlock != 0
		if cantBlock != ferocious {
			t.Errorf("ferocious=%v: their creature can't block = %v", ferocious, cantBlock)
		}
	}
}

func TestOldB347SeeTheUnwrittenPutsOneCreatureOrTwoWithFerocious(t *testing.T) {
	for _, ferocious := range []bool{false, true} {
		g, me, _ := oldB347Table(t, ferocious)
		var creatures []uuid.UUID
		for i := 0; i < 8; i++ {
			tl := "Creature — Bear"
			if i >= 3 {
				tl = "Sorcery"
			}
			id := effectstest.PlTop(me, "Card", tl, "")
			if i < 3 {
				creatures = append(creatures, id)
			}
		}
		// PushTop in order: the last pushed is on top; all eight are in the top eight.
		effectstest.CastFromHandForTest(t, g, me, "See the Unwritten", "Sorcery", "", oldB347Unwritten, game.CastSpellParams{Strict: true})
		effectstest.PassPriorityAroundTable(t, g)
		picks := creatures[:2]
		if err := func() error {
			pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
			if pick == nil {
				for _, c := range g.PendingChoices {
					t.Logf("pending %s %q", c.Kind, c.Reason)
				}
				t.Logf("stack %d gy %d lib %d", g.Stack.Size(), me.Graveyard.Size(), me.Library.Size())
				for _, ev := range g.Events {
					if ev.Kind == game.EventEffectError {
						t.Logf("error: %s", ev.ErrorMsg)
					}
				}
				t.Fatalf("ferocious=%v: no choose-cards prompt; creatures on battlefield %v", ferocious, g.Battlefield.Contains(creatures[0]))
			}
			return g.ResolveChooseCards(pick.ID, me.ID, picks)
		}(); err == nil && !ferocious {
			t.Error("two creature cards were accepted without ferocious")
		} else if err != nil && ferocious {
			t.Errorf("ferocious: two creature cards refused: %v", err)
		}
		if !ferocious {
			effectstest.AnswerChooseCards(t, g, me.ID, creatures[0])
		}
		on := 0
		for _, id := range creatures {
			if g.Battlefield.Contains(id) {
				on++
			}
		}
		want := 1
		if ferocious {
			want = 2
		}
		if on != want {
			t.Errorf("ferocious=%v: %d creature cards entered, want %d", ferocious, on, want)
		}
		if !me.Graveyard.Contains(creatures[2]) {
			t.Errorf("ferocious=%v: the rest did not go to the graveyard", ferocious)
		}
	}
}

func TestOldB347CratersClawsDealsXOrXPlusTwo(t *testing.T) {
	for _, ferocious := range []bool{false, true} {
		g, me, opp := oldB347Table(t, ferocious)
		life := opp.Life
		effectstest.CastFromHandForTest(t, g, me, "Crater's Claws", "Sorcery", "{X}{R}", oldB347Claws, game.CastSpellParams{XValue: 2, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
		effectstest.PassPriorityAroundTable(t, g)
		want := 2
		if ferocious {
			want = 4
		}
		if got := life - opp.Life; got != want {
			t.Errorf("ferocious=%v: %d damage, want %d", ferocious, got, want)
		}
	}
}

func TestOldB347IcyBlastTapsAndFreezesWithFerocious(t *testing.T) {
	for _, ferocious := range []bool{false, true} {
		g, me, opp := oldB347Table(t, ferocious)
		a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
		b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
		spare := effectstest.B12Creature(g, opp.ID, "Spare", "Creature — Bear", 2, 2)
		effectstest.CastFromHandForTest(t, g, me, "Icy Blast", "Instant", "{X}{U}", oldB347IcyBlast, game.CastSpellParams{XValue: 2, Targets: effectstest.CardTargets(a, b)})
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.B16Tapped(t, g, a) || !effectstest.B16Tapped(t, g, b) || effectstest.B16Tapped(t, g, spare) {
			t.Fatalf("ferocious=%v: tapped = %v %v %v, want true true false", ferocious,
				effectstest.B16Tapped(t, g, a), effectstest.B16Tapped(t, g, b), effectstest.B16Tapped(t, g, spare))
		}
		effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
		effectstest.PassPriorityAroundTable(t, g)
		stillTapped := effectstest.B16Tapped(t, g, a) && effectstest.B16Tapped(t, g, b)
		if stillTapped != ferocious {
			t.Errorf("ferocious=%v: still tapped after their untap step = %v", ferocious, stillTapped)
		}
	}
}

func TestOldB347FeedTheClanGainsFiveOrTen(t *testing.T) {
	for _, ferocious := range []bool{false, true} {
		g, me, _ := oldB347Table(t, ferocious)
		life := me.Life
		effectstest.CastFromHandForTest(t, g, me, "Feed the Clan", "Instant", "", oldB347FeedClan, game.CastSpellParams{Strict: true})
		effectstest.PassPriorityAroundTable(t, g)
		want := 5
		if ferocious {
			want = 10
		}
		if got := me.Life - life; got != want {
			t.Errorf("ferocious=%v: gained %d, want %d", ferocious, got, want)
		}
	}
}
