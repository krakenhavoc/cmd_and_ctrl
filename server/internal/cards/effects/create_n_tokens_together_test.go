package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// create_n_tokens_together_test.go — #4867 (ADR 0061 amendment
// 2026-11-05): "create N tokens" creates them at the same time
// (CR 701.7a, CR 608.2f), so they enter the battlefield in one event.
// Every permanent on the battlefield, the newcomers included, is checked
// for that one event (CR 603.6a): a Soul Warden token sees the other
// tokens of its own creation enter, a "whenever one or more tokens you
// control enter" ability triggers once (CR 603.2c), and a per-creature
// trigger outside the creation still triggers once per token. A token
// doubler still applies to the instruction (CR 616.1) and its extra
// tokens enter with the rest. The probes are token_copies_together_test.go's.

// cntSoulWardenToken is a token that has Soul Warden's text.
func cntSoulWardenToken() game.Card {
	return game.Card{
		Name: "Soul Warden", OracleID: b04SoulWardenOracle, TypeLine: "Token Creature — Human Cleric", Power: 1, Toughness: 1,
	}
}

func TestCreateNTokensEntersThemTogether(t *testing.T) {
	for _, tc := range []struct {
		name    string
		doubled bool
		tokens  int
		// life is what the Soul Wardens gain: the one already on the
		// battlefield sees every token, and each token sees every other.
		life int
	}{
		{"three", false, 3, 3 + 3*2},
		{"three, doubled by Doubling Season", true, 6, 6 + 6*5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			tctSoulWarden(g, me.ID)
			tctMisterFantastic(g, me.ID)
			if tc.doubled {
				seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
			}
			life, hand := me.Life, me.Hand.Size()

			g.WithWriteLock(func() {
				if err := g.CreateTokenForEffect(me.ID, cntSoulWardenToken(), 3); err != nil {
					t.Fatal(err)
				}
			})
			tctSettle(t, g, me.ID)

			tokens := tctTokens(g, me.ID)
			if len(tokens) != tc.tokens {
				t.Fatalf("%d tokens, want %d", len(tokens), tc.tokens)
			}
			for _, tok := range tokens {
				if tok.EntryOrdinal != tokens[0].EntryOrdinal {
					t.Fatal("the tokens did not enter in one event")
				}
			}
			if got := me.Life - life; got != tc.life {
				t.Errorf("gained %d life, want %d: each Soul Warden token sees the others enter (CR 603.6a)", got, tc.life)
			}
			if got := me.Hand.Size() - hand; got != 1 {
				t.Errorf("drew %d cards, want 1: the tokens entering together are one event (CR 603.2c)", got)
			}
		})
	}
}

// Cleaver Skaab's "create two tokens that are copies of the sacrificed
// creature" makes both copies in one creation: two Soul Warden copies
// see each other enter.
func TestCleaverSkaabCreatesItsTwoCopiesAtTheSameTime(t *testing.T) {
	for _, tc := range []struct {
		name    string
		doubled bool
		copies  int
	}{
		{"two", false, 2},
		{"doubled by Doubling Season", true, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			advanceToMain(t, g)
			skaab := pushCatalogPermanent(g, me.ID, "Cleaver Skaab", "Creature — Zombie Horror", wrCleaverSkaabOracle, false)
			// A Soul Warden that is a Zombie, so the Skaab can sacrifice it.
			warden := b27Push(g, me.ID, "Soul Warden", "Creature — Zombie Cleric", b04SoulWardenOracle, "{W}", 1, 1, "W")
			tctSoulWarden(g, me.ID) // watches; is not sacrificed
			tctMisterFantastic(g, me.ID)
			if tc.doubled {
				seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
			}
			life, hand := me.Life, me.Hand.Size()

			fillPoolColored(me, "C", 3)
			b16Activate(t, g, me.ID, skaab, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{warden}})
			tctSettle(t, g, me.ID)

			tokens := tctTokens(g, me.ID)
			if len(tokens) != tc.copies {
				t.Fatalf("%d copies, want %d", len(tokens), tc.copies)
			}
			for _, tok := range tokens {
				if tok.OracleID != b04SoulWardenOracle {
					t.Errorf("a token is %q, not a copy of the sacrificed Soul Warden", tok.Name)
				}
				if tok.EntryOrdinal != tokens[0].EntryOrdinal {
					t.Error("the copies did not enter in one event")
				}
			}
			// The watching Soul Warden sees every copy; each copy sees the
			// others (CR 603.6a).
			want := tc.copies + tc.copies*(tc.copies-1)
			if got := me.Life - life; got != want {
				t.Errorf("gained %d life, want %d: each copy sees the others enter", got, want)
			}
			if got := me.Hand.Size() - hand; got != 1 {
				t.Errorf("drew %d cards, want 1: the copies entering together are one event (CR 603.2c)", got)
			}
		})
	}
}

// Skyclave Relic's two tapped copies enter together, after the Relic
// itself, and both keep the creation's "tapped".
func TestSkyclaveRelicCreatesItsTwoCopiesAtTheSameTime(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tctMisterFantastic(g, me.ID)
	samiGiveColorless(me, 3)
	hand := me.Hand.Size()
	relic, err := castWithOptionalCosts(t, g, "Skyclave Relic", "Artifact", samiSkyclaveRelic, nil, []int{0}, nil)
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	tctSettle(t, g, me.ID)

	tokens := tctTokens(g, me.ID)
	if len(tokens) != 2 {
		t.Fatalf("%d copies, want 2", len(tokens))
	}
	original := findBattlefieldCardForTest(g, relic)
	if original == nil {
		t.Fatal("the Relic is not on the battlefield")
	}
	for _, tok := range tokens {
		if !tok.Tapped {
			t.Error("a copy entered untapped")
		}
		if tok.EntryOrdinal != tokens[0].EntryOrdinal {
			t.Error("the copies did not enter in one event")
		}
		if tok.EntryOrdinal == original.EntryOrdinal {
			t.Error("a copy entered with the Relic, before the Relic's trigger resolved")
		}
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1: the copies entering together are one event (CR 603.2c)", got)
	}
}

// A copy whose entry asks a question (a Clone's "enter as a copy?") asks
// it in its entry window, before any copy of the creation lands, and each
// question is a restore point: the paused creation is the simultaneous
// entry's frame (ReplacementResumeMoveTokenBatch), with both staged copies
// written. Answered on the restored table, both copies land together.
func TestCleaverSkaabCopiesPausedOnAnEntryQuestionSurviveARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	skaab := pushCatalogPermanent(g, me.ID, "Cleaver Skaab", "Creature — Zombie Horror", wrCleaverSkaabOracle, false)
	bears := seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	// A Clone that is a Zombie, so the Skaab can sacrifice it; it copies
	// nothing, so its copies are Clones and ask.
	clone := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Clone", TypeLine: "Creature — Zombie Shapeshifter", OracleID: oracleClone,
		Owner: me.ID, Controller: me.ID,
	})

	fillPoolColored(me, "C", 3)
	b16Activate(t, g, me.ID, skaab, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{clone}})
	passPriorityAroundTable(t, g)
	for q := 0; q < 2; q++ {
		if copyPrompt(g) == nil {
			t.Fatalf("copy %d did not ask", q)
		}
		g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveTokenBatch)
		if n := len(battlefieldIDsNamed(g, "Grizzly Bears")); n != 1 {
			t.Fatalf("question %d: %d Grizzly Bears on the restored table, want 1: no copy lands before every entry is settled", q, n)
		}
		c := copyPrompt(g)
		if c == nil {
			t.Fatalf("question %d: the restored table has no question", q)
		}
		if err := g.ResolveCopyTarget(c.ID, c.Chooser, bears); err != nil {
			t.Fatalf("ResolveCopyTarget: %v", err)
		}
	}
	me = g.Seats[0]
	tokens := tctTokens(g, me.ID)
	if len(tokens) != 2 {
		t.Fatalf("%d copies after the answers, want 2", len(tokens))
	}
	for _, tok := range tokens {
		if tok.Name != "Grizzly Bears" {
			t.Errorf("a copy is %q, want a Grizzly Bears", tok.Name)
		}
	}
	if tokens[0].EntryOrdinal != tokens[1].EntryOrdinal {
		t.Error("the copies did not enter in one event")
	}
	if c := copyPrompt(g); c != nil {
		t.Errorf("a third question is open: %+v", c)
	}
}
