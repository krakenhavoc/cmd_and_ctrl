package legal_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// sacrifice_n_test.go — #747 (ADR 0020 addendum §15): the enumerator
// reads N off the sacrifice clause at all three cost sites, offers
// nothing when the pool is short, and offers ONE payment per move for
// N >= 2, taken in the policy-neutral order (tokens first, then lower
// mana value, then the source last, then board order).

const (
	oracleSavvyHunter  = "602132c2-8ee8-41f8-bfac-cb17d32203f5"
	oracleKuldotha     = "b0f99367-4313-4bb1-a9fd-7711bc4ce40e"
	oracleTeysa        = "8191342b-b25e-4c4d-8f69-aee662148ff4"
	oracleSacNAltar    = "legal-test-sacrifice-n-altar"
	oracleSacNRites    = "legal-test-sacrifice-n-rites"
	sacNAltarLabel     = "Sacrifice two creatures: Add {C}{C}{C}"
	sacNRitesSpellName = "Two-Creature Rites"
)

func init() {
	effects.Register(effects.Spec{
		OracleID: oracleSacNAltar,
		Name:     "Two-Creature Altar",
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{SacrificeOther: effects.SacrificeN(2, "two creatures", effects.Creature()).SacrificeOther},
			Produced: "{C}{C}{C}",
			Label:    sacNAltarLabel,
		}},
	})
	effects.Register(effects.Spec{
		OracleID:       oracleSacNRites,
		Name:           sacNRitesSpellName,
		AdditionalCost: effects.SacrificeNCost(2, "two creatures", effects.Creature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{N: 2}.Apply(ctx)
		},
	})
}

func sacrificeIDsOf(t *testing.T, m legal.Move) []string {
	t.Helper()
	var p struct {
		SacrificeIDs []string `json:"sacrifice_ids"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("bad params %s: %v", string(m.Params), err)
	}
	return p.SacrificeIDs
}

func movesFrom(moves []legal.Move, source uuid.UUID, kind legal.Kind) []legal.Move {
	var out []legal.Move
	for _, m := range moves {
		if m.Source == source && m.Kind == kind {
			out = append(out, m)
		}
	}
	return out
}

func token(name, typeLine string) game.Card {
	return game.Card{Name: name, TypeLine: "Token " + typeLine}
}

func TestSacrificeNOffersNothingBelowNAndOnePaymentAtN(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	hunter := battlefieldCard(g, active, game.Card{Name: "Savvy Hunter", TypeLine: "Creature — Human Warrior", OracleID: oracleSavvyHunter, Power: 3, Toughness: 3})
	food1 := battlefieldCard(g, active, token("Food", "Artifact — Food"))
	advanceTo(t, g, game.StepPrecombatMain)

	if got := movesFrom(legal.EnumerateFor(g, active.ID), hunter, legal.KindActivate); len(got) != 0 {
		t.Fatalf("one Food for a two-Food cost is no move (#544): %v", labels(got))
	}
	food2 := battlefieldCard(g, active, token("Food", "Artifact — Food"))
	food3 := battlefieldCard(g, active, token("Food", "Artifact — Food"))
	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	got := movesFrom(moves, hunter, legal.KindActivate)
	if len(got) != 1 {
		t.Fatalf("three Foods choose two is ONE payment, not every combination: %v", labels(got))
	}
	ids := sacrificeIDsOf(t, got[0])
	if len(ids) != 2 || ids[0] != food1.String() || ids[1] != food2.String() {
		t.Errorf("sacrifice_ids = %v, want the first two Foods in board order (%s, %s; not %s)", ids, food1, food2, food3)
	}
}

// Tokens before nontokens, then lower mana value, then the source
// last. Kuldotha Forgemaster is an artifact, so it is a candidate for
// its own cost; it is given mana value 0 here so only the source rule
// can put it behind the other zero-cost artifact.
func TestSacrificeNOrderIsTokensThenManaValueThenSourceLast(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	forge := battlefieldCard(g, active, game.Card{Name: "Kuldotha Forgemaster", TypeLine: "Artifact Creature — Construct", OracleID: oracleKuldotha, Power: 3, Toughness: 5})
	battlefieldCard(g, active, game.Card{Name: "Wurmcoil", TypeLine: "Artifact Creature — Phyrexian Wurm", ManaCost: "{6}", Power: 6, Toughness: 6})
	ring := battlefieldCard(g, active, game.Card{Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}"})
	ornithopter := battlefieldCard(g, active, game.Card{Name: "Ornithopter", TypeLine: "Artifact Creature — Thopter", ManaCost: "{0}"})
	treasure := battlefieldCard(g, active, token("Treasure", "Artifact — Treasure"))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	got := movesFrom(moves, forge, legal.KindActivate)
	if len(got) != 1 {
		t.Fatalf("want one Forgemaster payment, got %v", labels(got))
	}
	want := []string{treasure.String(), ornithopter.String(), forge.String()}
	if ids := sacrificeIDsOf(t, got[0]); fmt.Sprint(ids) != fmt.Sprint(want) {
		t.Errorf("sacrifice_ids = %v, want the Treasure token, then the zero-cost Ornithopter, then the zero-cost source (%v); the Sol Ring (%s) is fourth", ids, want, ring)
	}
}

// One payment per move keeps the shared expansion budget for targets:
// a three-white-creature cost with five white creatures still reaches
// every target rather than spending the budget on sacrifice sets for
// the first one.
func TestSacrificeNTargetedAbilityReachesEveryTarget(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	clearHand(active)
	teysa := battlefieldCard(g, active, game.Card{Name: "Teysa, Orzhov Scion", TypeLine: "Legendary Creature — Human Advisor", OracleID: oracleTeysa, ManaCost: "{1}{W}{B}", Power: 2, Toughness: 3})
	for i := 0; i < 5; i++ {
		battlefieldCard(g, active, game.Card{Name: fmt.Sprintf("Soldier %d", i), TypeLine: "Creature — Soldier", ManaCost: "{W}", Colors: []string{"W"}, Power: 1, Toughness: 1})
	}
	for i := 0; i < 3; i++ {
		battlefieldCard(g, opp, creature(fmt.Sprintf("Bear %d", i), "{1}{G}", 2, 2))
	}
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	got := movesFrom(moves, teysa, legal.KindActivate)
	// Nine creatures on the table (Teysa, five Soldiers, three Bears),
	// each a legal target, each with the one payment.
	if len(got) != 9 {
		t.Errorf("want one activation per target creature (9), got %d: %v", len(got), labels(got))
	}
	for _, m := range got {
		if n := len(sacrificeIDsOf(t, m)); n != 3 {
			t.Errorf("%q pays %d creatures, want 3", m.Label, n)
		}
	}
}

func TestSacrificeNManaAbilityNamesEveryPermanentPaid(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	altar := battlefieldCard(g, active, game.Card{Name: "Two-Creature Altar", TypeLine: "Artifact", OracleID: oracleSacNAltar})
	battlefieldCard(g, active, creature("Goblin", "{R}", 1, 1))
	advanceTo(t, g, game.StepPrecombatMain)
	if got := movesFrom(legal.EnumerateFor(g, active.ID), altar, legal.KindMana); len(got) != 0 {
		t.Fatalf("one creature for a two-creature mana cost is no move: %v", labels(got))
	}
	battlefieldCard(g, active, creature("Bear", "{1}{G}", 2, 2))
	battlefieldCard(g, active, creature("Elf", "{G}", 1, 1))
	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	got := movesFrom(moves, altar, legal.KindMana)
	if len(got) != 1 {
		t.Fatalf("want one mana payment, got %v", labels(got))
	}
	// Mana value breaks the tie between nontokens: Goblin {R} and Elf
	// {G} (1) before Bear {1}{G} (2), in board order.
	if want := "Two-Creature Altar: " + sacNAltarLabel + " (sacrificing Goblin, Elf)"; got[0].Label != want {
		t.Errorf("label = %q, want %q", got[0].Label, want)
	}
}

// ADR 0020 addendum §16 parity: the protocol view ships
// sacrifice_options.cards in the order this package takes its payment
// from, so for N >= 2 the first N entries a human's "Choose for me"
// takes are exactly the set the enumerator offers a bot — at each of
// the three sites.
func TestSacrificeNViewOrderMatchesTheEnumeratorsPayment(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	hunter := battlefieldCard(g, active, game.Card{Name: "Savvy Hunter", TypeLine: "Creature — Human Warrior", OracleID: oracleSavvyHunter, ManaCost: "{1}{B}{G}", Power: 3, Toughness: 3})
	altar := battlefieldCard(g, active, game.Card{Name: "Two-Creature Altar", TypeLine: "Artifact", OracleID: oracleSacNAltar})
	rites := handCard(active, game.Card{Name: sacNRitesSpellName, TypeLine: "Sorcery", ManaCost: "{B}", OracleID: oracleSacNRites})
	battlefieldCard(g, active, basic("Swamp", "Swamp"))
	battlefieldCard(g, active, game.Card{Name: "Bakery", TypeLine: "Artifact — Food", ManaCost: "{1}"})
	battlefieldCard(g, active, token("Food", "Artifact — Food"))
	battlefieldCard(g, active, game.Card{Name: "Pantry", TypeLine: "Artifact — Food", ManaCost: "{0}"})
	battlefieldCard(g, active, token("Food", "Artifact — Food"))
	battlefieldCard(g, active, creature("Ogre", "{2}{R}", 3, 3))
	battlefieldCard(g, active, game.Card{Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1})
	battlefieldCard(g, active, creature("Goblin", "{R}", 1, 1))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	view := protocol.ViewOfGameFor(g, active.ID.String())
	var board, hand []protocol.CardView
	board = view.Battlefield.Cards
	for _, s := range view.Seats {
		if s.ID == active.ID.String() {
			hand = s.Hand.Cards
		}
	}
	find := func(cards []protocol.CardView, id uuid.UUID) protocol.CardView {
		for _, c := range cards {
			if c.InstanceID == id.String() {
				return c
			}
		}
		t.Fatalf("card %s missing from the view", id)
		return protocol.CardView{}
	}
	sites := []struct {
		name  string
		moves []legal.Move
		opts  *protocol.LegalTargetsView
	}{
		{"activated ability", movesFrom(moves, hunter, legal.KindActivate), find(board, hunter).ActivatedAbilities[0].SacrificeOptions},
		{"mana ability", movesFrom(moves, altar, legal.KindMana), find(board, altar).ManaAbilities[0].SacrificeOptions},
		{"additional cost", movesFrom(moves, rites, legal.KindCast), find(hand, rites).AdditionalCost.SacrificeOptions},
	}
	for _, s := range sites {
		if len(s.moves) != 1 || s.opts == nil || len(s.opts.Cards) < 2 {
			t.Errorf("%s: want one move and at least two view options, got %v and %+v", s.name, labels(s.moves), s.opts)
			continue
		}
		if got, want := fmt.Sprint(sacrificeIDsOf(t, s.moves[0])), fmt.Sprint(s.opts.Cards[:2]); got != want {
			t.Errorf("%s: the enumerator pays %s but the view's first two options are %s", s.name, got, want)
		}
	}
}

func TestSacrificeNAdditionalCostOffersOneCastAtN(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	rites := handCard(active, game.Card{Name: sacNRitesSpellName, TypeLine: "Sorcery", ManaCost: "{B}", OracleID: oracleSacNRites})
	battlefieldCard(g, active, basic("Swamp", "Swamp"))
	battlefieldCard(g, active, creature("Goblin", "{R}", 1, 1))
	advanceTo(t, g, game.StepPrecombatMain)
	if got := movesFrom(legal.EnumerateFor(g, active.ID), rites, legal.KindCast); len(got) != 0 {
		t.Fatalf("one creature for a two-creature additional cost is no cast: %v", labels(got))
	}
	battlefieldCard(g, active, creature("Bear", "{1}{G}", 2, 2))
	battlefieldCard(g, active, creature("Elf", "{G}", 1, 1))
	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	got := movesFrom(moves, rites, legal.KindCast)
	if len(got) != 1 {
		t.Fatalf("want one cast, got %v", labels(got))
	}
	if n := len(sacrificeIDsOf(t, got[0])); n != 2 {
		t.Errorf("the cast pays %d creatures, want 2", n)
	}
}
