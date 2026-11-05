package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// log_self_sacrifice_test.go — #4427, ADR 0209 §3: the client's quick
// "crack" cue plays, in place of the sacrifice's altar and knife, when a
// permanent is sacrificed as a cost of its own ability: a fetch land
// cracked, a Treasure sacrificed for its mana (CR 111.10a), a Clue for its
// card (CR 111.10f). The log already wrote a `sacrifice` line for each,
// in the batch the ability was activated in, and nothing on it said it
// was a cost, or whose. It now carries, additively, `cause` "cost" for a
// sacrifice paid as a cost (CR 118.1, 601.2h, 602.2b), and, when the cost
// is the sacrificed permanent's own ability's, `source` is that permanent
// itself, its own card_id. Nothing new is revealed: the activation is a
// public action, and the source is the card the line already names. This
// pins, for every viewer, a spectator included:
//
//   - a fetch land (Evolving Wilds): its sacrifice is a cost of its own;
//   - a Treasure's mana ability, activated by hand: a cost of its own,
//     after the token has ceased to exist (CR 111.7);
//   - a Clue's ability: a cost of its own;
//   - two Treasures the auto-tapper sacrifices to pay for a spell as it
//     is cast (CR 601.2g): two lines, each a cost of its own, in the
//     cast's batch;
//   - a creature sacrificed to Viscera Seer's ability: a cost, of
//     another permanent's ability, so no source;
//   - a sacrifice an effect makes, and a sandbox sacrifice: no cause.

// sacrificesOf is every `sacrifice` entry viewer sees for card id.
func sacrificesOf(g *game.Game, viewer string, id uuid.UUID) []LogEvent {
	var out []LogEvent
	for _, e := range logEntriesOfKind(ViewOfGameFor(g, viewer).Log, LogSacrifice) {
		if e.CardID == id.String() {
			out = append(out, e)
		}
	}
	return out
}

// assertSacrifice checks that every viewer sees exactly one `sacrifice`
// entry for id, in a batch, with the wire cause and source given ("" for
// absent), and returns what the first viewer saw.
func assertSacrifice(t *testing.T, g *game.Game, id uuid.UUID, cause, source string) LogEvent {
	t.Helper()
	var first LogEvent
	for i, viewer := range allViewers(g) {
		got := sacrificesOf(g, viewer, id)
		if len(got) != 1 {
			t.Fatalf("viewer %q: %d sacrifice entries for the card, want 1: %+v", viewer, len(got), got)
		}
		e := got[0]
		raw := wireFields(t, e)
		for key, w := range map[string]string{"cause": cause, "source": source} {
			v, ok := raw[key]
			if w == "" {
				if ok {
					t.Errorf("viewer %q: wire field %q = %v, want it absent (entry %v)", viewer, key, v, raw)
				}
				continue
			}
			if v != w {
				t.Errorf("viewer %q: wire field %q = %v, want %v (entry %v)", viewer, key, v, w, raw)
			}
		}
		if e.Batch == 0 {
			t.Errorf("viewer %q: the sacrifice entry has no batch: %+v", viewer, e)
		}
		assertNoUUID(t, e.Text)
		if i == 0 {
			first = e
		}
	}
	return first
}

// selfSacToken creates one token from tmpl under p and returns its ID.
func selfSacToken(t *testing.T, g *game.Game, p *game.Player, tmpl game.Card) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	g.WithWriteLock(func() {
		ids, err := g.CreateTokensForEffect(p.ID, tmpl, 1, game.TokenEntryOptions{})
		if err != nil || len(ids) != 1 {
			t.Fatalf("create %s: %v %v", tmpl.Name, ids, err)
		}
		id = ids[0]
	})
	return id
}

// pushCatalogPermanent puts a public catalog permanent onto the
// battlefield under owner.
func pushCatalogPermanent(g *game.Game, owner, opp *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
			Owner: owner.ID, Controller: owner.ID, Power: 1, Toughness: 1,
			KnownBy: map[uuid.UUID]bool{owner.ID: true, opp.ID: true},
		})
	})
	return id
}

func TestASacrificePaidForItsOwnAbilityNamesItselfAsItsSource(t *testing.T) {
	t.Run("a fetch land", func(t *testing.T) {
		g, me, opp := attachTable(t)
		wilds := pushCatalogPermanent(g, me, opp, "Evolving Wilds", "Land", "a75445d3-1303-4bb5-89ad-26ea93fecd48")
		if err := g.ActivateCatalogAbility(me.ID, wilds, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate Evolving Wilds: %v", err)
		}
		e := assertSacrifice(t, g, wilds, "cost", wilds.String())
		// The activation is the same moment; the search comes later, as it
		// resolves (log_library_search_test.go).
		var activated bool
		for _, a := range logEntriesOfKind(ViewOfGameFor(g, opp.ID.String()).Log, LogActivate) {
			if a.CardID == wilds.String() {
				activated = true
				if a.Batch != e.Batch {
					t.Errorf("the activation's batch %d, its sacrifice's %d", a.Batch, e.Batch)
				}
			}
		}
		if !activated {
			t.Error("no activate entry for Evolving Wilds")
		}
	})

	t.Run("a Treasure's mana ability, after the token is gone", func(t *testing.T) {
		g, me, _ := attachTable(t)
		tr := selfSacToken(t, g, me, effects.TreasureToken())
		if err := g.ActivateManaAbility(me.ID, tr, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
			t.Fatalf("activate the Treasure: %v", err)
		}
		g.ReadSnapshot(func() {
			if g.Battlefield.Contains(tr) {
				t.Fatal("the Treasure is still on the battlefield")
			}
		})
		e := assertSacrifice(t, g, tr, "cost", tr.String())
		if e.Text != "P1 sacrificed Treasure" {
			t.Errorf("text %q", e.Text)
		}
	})

	t.Run("a Clue's ability", func(t *testing.T) {
		g, me, _ := attachTable(t)
		clue := selfSacToken(t, g, me, effects.ClueToken())
		g.WithWriteLock(func() {
			me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
		})
		if err := g.ActivateCatalogAbility(me.ID, clue, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate the Clue: %v", err)
		}
		assertSacrifice(t, g, clue, "cost", clue.String())
	})

	t.Run("two Treasures the auto-tapper spends on a spell as it is cast", func(t *testing.T) {
		g, me, opp := attachTable(t)
		spell := game.NewCard("Two Drop", me.ID)
		spell.TypeLine = "Sorcery"
		spell.ManaCost = "{2}"
		spell.KnownBy = map[uuid.UUID]bool{me.ID: true, opp.ID: true}
		g.WithWriteLock(func() {
			me.ManaPool = nil
			me.Hand.PushTop(spell)
		})
		t1 := selfSacToken(t, g, me, effects.TreasureToken())
		t2 := selfSacToken(t, g, me, effects.TreasureToken())
		if err := g.CastSpell(me.ID, spell.InstanceID, game.CastSpellParams{AutoTap: true}); err != nil {
			t.Fatalf("cast with Treasures: %v", err)
		}
		a := assertSacrifice(t, g, t1, "cost", t1.String())
		b := assertSacrifice(t, g, t2, "cost", t2.String())
		if a.Batch != b.Batch {
			t.Errorf("one payment's sacrifices in batches %d and %d", a.Batch, b.Batch)
		}
		casts := logEntriesOfKind(ViewOfGameFor(g, opp.ID.String()).Log, LogCast)
		if len(casts) != 1 || casts[0].Batch != a.Batch || casts[0].Seq < b.Seq {
			t.Errorf("the cast entry %+v: want one, after the payment, in its batch %d", casts, a.Batch)
		}
	})
}

func TestASacrificeThatIsNotItsOwnCostNamesNoSelf(t *testing.T) {
	t.Run("a creature sacrificed to another permanent's ability", func(t *testing.T) {
		g, me, opp := attachTable(t)
		seer := pushCatalogPermanent(g, me, opp, "Viscera Seer", "Creature — Vampire Wizard", "f82a4e85-526d-4456-b700-7760043a31be")
		bear := pushPermanent(g, me, opp, "Bear", "Creature — Bear")
		if err := g.ActivateCatalogAbility(me.ID, seer, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err != nil {
			t.Fatalf("activate Viscera Seer: %v", err)
		}
		assertSacrifice(t, g, bear, "cost", "")
	})

	t.Run("an effect's sacrifice", func(t *testing.T) {
		g, me, opp := attachTable(t)
		bear := pushPermanent(g, me, opp, "Bear", "Creature — Bear")
		g.WithWriteLock(func() {
			if err := g.SacrificePermanentForEffect(bear); err != nil {
				t.Fatalf("SacrificePermanentForEffect: %v", err)
			}
		})
		assertSacrifice(t, g, bear, "", "")
	})

	t.Run("a sandbox sacrifice", func(t *testing.T) {
		g, me, opp := attachTable(t)
		bear := pushPermanent(g, me, opp, "Bear", "Creature — Bear")
		if err := g.SacrificePermanent(me.ID, bear); err != nil {
			t.Fatalf("SacrificePermanent: %v", err)
		}
		assertSacrifice(t, g, bear, "", "")
	})
}
