package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// create_tokens_per_group_together_test.go — #4867, the cards the ADR 0061
// amendment of 2026-11-05 left creating once per player (amendment of
// 2026-11-06). Encore's "for each opponent, create a token ... that attacks
// that opponent" and Furygale Flocking's pairs are ONE creation of one
// tagged group per opponent (game.TokenGroup.Tag), and Fractured Identity's
// "each player other than its controller creates a token" is one creation
// window per player entering as one event
// (game.CreateTokensForPlayersTogetherForEffect). Either way the tokens are
// created at the same time (CR 101.4, CR 608.2f) and each newcomer sees the
// others enter (CR 603.6a).

// ptgBindings is the opponent each token must attack (its
// ModAttackPlayerIfAble), keyed by token.
func ptgBindings(g *game.Game, toks []game.Card) map[uuid.UUID]uuid.UUID {
	out := map[uuid.UUID]uuid.UUID{}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	g.ReadSnapshot(func() {
		for _, tok := range toks {
			c, ok := g.LookupCardForEffect(tok.InstanceID)
			if !ok {
				continue
			}
			for _, r := range c.Effective().AttackRequirements {
				if r.MustAttack != uuid.Nil {
					out[tok.InstanceID] = r.MustAttack
				}
			}
		}
	})
	return out
}

// ptgNamed is the tokens named `name` that `controller` controls.
func ptgNamed(g *game.Game, controller uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range tctTokens(g, controller) {
		if c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

// ptgOneEntry fails unless every token entered in the same event.
func ptgOneEntry(t *testing.T, toks []game.Card) {
	t.Helper()
	for _, tok := range toks {
		if tok.EntryOrdinal != toks[0].EntryOrdinal {
			t.Fatal("the tokens of one creation did not enter in one event")
		}
	}
}

// ptgPerOpponent fails unless every token is bound to an opponent of me and
// each opponent has exactly `each` tokens bound to them.
func ptgPerOpponent(t *testing.T, g *game.Game, me uuid.UUID, toks []game.Card, each int) {
	t.Helper()
	bound := ptgBindings(g, toks)
	perOpp := map[uuid.UUID]int{}
	for _, tok := range toks {
		opp, ok := bound[tok.InstanceID]
		if !ok || opp == me {
			t.Fatalf("token %s is bound to %v, want an opponent", tok.InstanceID, opp)
		}
		perOpp[opp]++
	}
	if len(perOpp) != len(g.Seats)-1 {
		t.Fatalf("tokens are bound to %d opponents, want %d", len(perOpp), len(g.Seats)-1)
	}
	for opp, n := range perOpp {
		if n != each {
			t.Errorf("%d tokens bound to %v, want %d", n, opp, each)
		}
	}
}

func ptgPeregrinTook(g *game.Game, p uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Peregrin Took", TypeLine: "Legendary Creature — Halfling Citizen",
		OracleID: peregrinTookOracle, Power: 2, Toughness: 3, Owner: p, Controller: p,
	})
}

// Impulsive Pilferer's encore makes its three copies in one creation: they
// enter together, each is bound to its own opponent, and a doubler doubles
// each opponent's share, both of which attack that opponent. A Food that
// Peregrin Took adds to the creation is not an encore token: it is neither
// bound nor sacrificed.
func TestEncoreCreatesEveryOpponentsTokenAtTheSameTime(t *testing.T) {
	for _, tc := range []struct {
		name     string
		doubled  bool
		peregrin bool
		each     int
	}{
		{"one per opponent", false, false, 1},
		{"doubled by Doubling Season", true, false, 2},
		{"with Peregrin Took's Food", false, true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			if tc.doubled {
				seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
			}
			if tc.peregrin {
				ptgPeregrinTook(g, me.ID)
			}
			activateFromGraveyard(t, g, "Impulsive Pilferer", "Creature — Goblin Pirate",
				impulsivePilfererOracle, 1, 1, "{C}{C}{C}{R}", game.ActivateAbilityParams{})
			passPriorityAroundTable(t, g)

			toks := ptgNamed(g, me.ID, "Impulsive Pilferer")
			if want := tc.each * (len(g.Seats) - 1); len(toks) != want {
				t.Fatalf("%d tokens, want %d", len(toks), want)
			}
			ptgOneEntry(t, toks)
			ptgPerOpponent(t, g, me.ID, toks, tc.each)
			if !tc.peregrin {
				return
			}
			food := ptgNamed(g, me.ID, "Food")
			if len(food) != 1 {
				t.Fatalf("%d Food, want Peregrin Took's one", len(food))
			}
			if food[0].EntryOrdinal != toks[0].EntryOrdinal {
				t.Error("the Food did not enter with the rest of the creation")
			}
			if b := ptgBindings(g, food); len(b) != 0 {
				t.Error("the Food is bound to attack: it is not an encore token")
			}
			bound := ptgBindings(g, toks)
			advanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
			for _, tok := range toks {
				if err := g.DeclareAttacker(tok.InstanceID, bound[tok.InstanceID]); err != nil {
					t.Fatalf("declare each token at its own opponent: %v", err)
				}
			}
			if err := g.PassPriority(); err != nil {
				t.Fatalf("pass with every token at its own opponent: %v", err)
			}
			advanceTo(t, g, game.StepEnd)
			passPriorityAroundTable(t, g)
			if n := len(ptgNamed(g, me.ID, "Impulsive Pilferer")); n != 0 {
				t.Errorf("%d encore tokens survived the end step", n)
			}
			// The sacrificed copies' death triggers make Treasures, and
			// Peregrin Took a Food with each, so count the first Food by ID.
			if !g.Battlefield.Contains(food[0].InstanceID) {
				t.Error("the creation's Food was sacrificed: encore sacrifices only its own tokens")
			}
		})
	}
}

// A creature token that has Soul Warden's text, made by encore's shape (one
// tagged group per opponent, the encore tail binding them): each token sees
// the other two enter (CR 603.6a) and is still bound to its own opponent.
func TestEncoreShapedSoulWardensSeeEachOtherEnter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	var groups []game.TokenGroup
	for _, s := range g.Seats[1:] {
		groups = append(groups, game.TokenGroup{Template: cntSoulWardenToken(), Count: 1, Tag: s.ID})
	}
	g.WithWriteLock(func() {
		if err := g.CreateTokensKeyedForEffect(game.TokenCreation{Controller: me.ID, Groups: groups},
			game.KeyedTail{Ref: encoreBindTokens, Player: me.ID}); err != nil {
			t.Fatal(err)
		}
	})
	tctSettle(t, g, me.ID)

	toks := ptgNamed(g, me.ID, "Soul Warden")
	if len(toks) != 3 {
		t.Fatalf("%d tokens, want 3", len(toks))
	}
	ptgOneEntry(t, toks)
	ptgPerOpponent(t, g, me.ID, toks, 1)
	if got := me.Life - life; got != 3*2 {
		t.Errorf("gained %d life, want 6: each Soul Warden token sees the other two enter", got)
	}
}

// Encore's one creation paused on the CR 616 order a Doubling Season and a
// Parallel Lives ask is a restore point: the tags ride the paused event's
// groups, and the restored answer makes four tokens per opponent, entering
// together, each bound to the opponent its group names.
func TestEncoreTaggedCreationIsARestorePointAtTheTokenOrderPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ds, pl := tokenOrderBoard(g, me.ID)
	activateFromGraveyard(t, g, "Impulsive Pilferer", "Creature — Goblin Pirate",
		impulsivePilfererOracle, 1, 1, "{C}{C}{C}{R}", game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !orderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeCreateTokens)
	me = g.Seats[g.Turn.ActiveSeat]
	if n := len(ptgNamed(g, me.ID, "Impulsive Pilferer")); n != 0 {
		t.Fatalf("the restored table made %d tokens before the order", n)
	}
	answerOrderBySource(t, g, ds, pl)
	toks := ptgNamed(g, me.ID, "Impulsive Pilferer")
	if want := 4 * (len(g.Seats) - 1); len(toks) != want {
		t.Fatalf("%d tokens, want %d", len(toks), want)
	}
	ptgOneEntry(t, toks)
	ptgPerOpponent(t, g, me.ID, toks, 4)
}

// Furygale Flocking's pairs are one creation: all six Elementals enter
// together, two bound to each opponent; doubled, four to each.
func TestFurygaleFlockingCreatesEveryOpponentsPairAtTheSameTime(t *testing.T) {
	for _, tc := range []struct {
		name    string
		doubled bool
		each    int
	}{
		{"two per opponent", false, 2},
		{"doubled by Doubling Season", true, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			advanceToMain(t, g)
			if tc.doubled {
				seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
			}
			castCatalogSpell(t, g, "Furygale Flocking", "Sorcery", socB6FurygaleOracle, nil)
			passPriorityAroundTable(t, g)

			toks := ptgNamed(g, me.ID, "Elemental")
			if want := tc.each * (len(g.Seats) - 1); len(toks) != want {
				t.Fatalf("%d Elementals, want %d", len(toks), want)
			}
			ptgOneEntry(t, toks)
			ptgPerOpponent(t, g, me.ID, toks, tc.each)
			for _, tok := range toks {
				if !hasEffectiveKeyword(t, g, tok.InstanceID, "haste") {
					t.Error("an Elemental lacks haste")
				}
			}
		})
	}
}

// Fractured Identity on a Soul Warden: the three other players' copies are
// created at the same time (CR 101.4, CR 608.2f) and enter in one event, so
// each copy sees the other two enter (CR 603.6a). A Doubling Season doubles
// only its controller's copy (each player's creation is their own, and the
// replacement effects applying to it are that player's, CR 701.7b), and that
// player's two copies enter with the rest.
func TestFracturedIdentityCopiesEnterAtTheSameTime(t *testing.T) {
	for _, tc := range []struct {
		name    string
		doubled bool
		// copies and life are per seat, the victim's controller (seat 1)
		// getting none.
		copies [4]int
		life   [4]int
	}{
		{"one copy each", false, [4]int{1, 0, 1, 1}, [4]int{2, 0, 2, 2}},
		{"Doubling Season under seat 2", true, [4]int{1, 0, 2, 1}, [4]int{3, 0, 6, 3}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			opp := g.Seats[1]
			warden := tctSoulWarden(g, opp.ID)
			if tc.doubled {
				seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", g.Seats[2].ID)
			}
			var life [4]int
			for i, s := range g.Seats {
				life[i] = s.Life
			}
			castCatalogSpell(t, g, "Fractured Identity", "Sorcery", b43FracturedIdentityOracle, cardRefs(warden))
			tctSettle(t, g, g.Seats[0].ID)
			if g.Battlefield.Contains(warden) {
				t.Fatal("the Soul Warden is exiled")
			}

			var all []game.Card
			for i, s := range g.Seats {
				toks := ptgNamed(g, s.ID, "Soul Warden")
				if len(toks) != tc.copies[i] {
					t.Errorf("seat %d has %d copies, want %d", i, len(toks), tc.copies[i])
				}
				all = append(all, toks...)
				if got := s.Life - life[i]; got != tc.life[i] {
					t.Errorf("seat %d gained %d life, want %d: each copy sees every other copy enter", i, got, tc.life[i])
				}
			}
			ptgOneEntry(t, all)
		})
	}
}

// A player whose creation window asks a question (Doubling Season beside
// Parallel Lives) is not waited for: the other players' copies enter
// together at once, and that player's copies enter together on their own
// once the order is answered, which is what every player's did before
// (CreateTokensForPlayersTogetherForEffect's fallback). The paused creation
// is the ordinary creation frame, so it is a restore point.
func TestFracturedIdentityDoesNotWaitForAPlayersOrderPrompt(t *testing.T) {
	g := newCatalogGame(t)
	opp, asked := g.Seats[1], g.Seats[2]
	bear := b43Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	ds, pl := tokenOrderBoard(g, asked.ID)
	castCatalogSpell(t, g, "Fractured Identity", "Sorcery", b43FracturedIdentityOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if !orderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}
	first := append(ptgNamed(g, g.Seats[0].ID, "Grizzly Bears"), ptgNamed(g, g.Seats[3].ID, "Grizzly Bears")...)
	if len(first) != 2 {
		t.Fatalf("%d copies under the players who were not asked, want 2", len(first))
	}
	ptgOneEntry(t, first)
	if n := len(ptgNamed(g, asked.ID, "Grizzly Bears")); n != 0 {
		t.Fatalf("the asked player has %d copies before answering", n)
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeCreateTokens)
	answerOrderBySource(t, g, ds, pl)
	later := ptgNamed(g, g.Seats[2].ID, "Grizzly Bears")
	if len(later) != 4 {
		t.Fatalf("the asked player has %d copies, want four (doubled twice)", len(later))
	}
	ptgOneEntry(t, later)
	if later[0].EntryOrdinal == first[0].EntryOrdinal {
		t.Error("the asked player's copies entered with the others, before the order was answered")
	}
}
