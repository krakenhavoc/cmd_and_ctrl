package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_source_cards_test.go — the catalog half of #1212: three cards
// that read the payment that made them.
//
// The engine-side contract is pinned in game/mana_source_test.go. What
// is pinned here is that the three printed shapes reach it: a
// permanent asking about the SOURCE of its mana (Hired Hexblade), a
// permanent asking about a COLOUR (Gruul Scrapper), and a spell asking
// about a colour while it resolves (Ribbons of Night).

const (
	hiredHexbladeOracle = "f3a0f155-05d8-465c-b0ef-35aa12e93013"
	gruulScrapperOracle = "878e8980-a2cc-4335-aafe-f5166ef48f79"
	ribbonsOfNightOracl = "f074c0d0-8455-484d-ae75-820fdfbc4740"
)

// crackATreasureFor spawns a real Treasure token, cracks it, and
// answers its colour pick — leaving one token of `color` in the pool
// that remembers it came from a Treasure.
func crackATreasureFor(t *testing.T, g *game.Game, p *game.Player, color string) {
	t.Helper()
	tmpl, ok := Tokens().TokenTemplate("Treasure")
	if !ok {
		t.Fatal("no Treasure template")
	}
	ids, err := g.SpawnCards(p.ID, p.ID, game.ZoneBattlefield, tmpl, 1)
	if err != nil || len(ids) != 1 {
		t.Fatalf("SpawnCards: %v (%d ids)", err, len(ids))
	}
	if err := g.ActivateManaAbility(p.ID, ids[0], 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("crack the Treasure: %v", err)
	}
	pick := riderLatestManaPick(g, p.ID)
	if pick == nil {
		t.Fatal("no colour pick queued by the Treasure")
	}
	if err := g.ResolveManaChoice(pick.ID, p.ID, color); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
}

// castFromHandForTest puts a card in hand and casts it under `params`.
func castFromHandForTest(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: p.ID, Controller: p.ID,
		})
	})
	if err := g.CastSpell(p.ID, id, params); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// --- Hired Hexblade: the SOURCE, read by a permanent ----------------

// Cast off a cracked Treasure and a Swamp, the Hexblade draws a card
// and loses a life. Cast off two Swamps it is a 2/2 and nothing else.
//
// The Treasure is in a graveyard — in fact it has ceased to exist
// (CR 111.7) — before the spell is even announced, and the spell has
// finished resolving before the trigger does. Both facts are recovered
// from the snapshot, which is the whole of #1212.
func TestHiredHexbladeDrawsOnlyOffTreasureMana(t *testing.T) {
	for _, tc := range []struct {
		name     string
		treasure bool
		draws    int
	}{
		{"a Treasure paid for it", true, 1},
		{"two Swamps paid for it", false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			if tc.treasure {
				crackATreasureFor(t, g, me, "B")
				floatForTest(g, me, "B")
			} else {
				floatForTest(g, me, "BB")
			}
			handBefore, lifeBefore := me.Hand.Size(), me.Life

			castFromHandForTest(t, g, me, "Hired Hexblade", "Creature — Elf Warlock",
				"{1}{B}", hiredHexbladeOracle, game.CastSpellParams{Strict: true})
			passPriorityAroundTable(t, g)

			if got := me.Hand.Size() - handBefore; got != tc.draws {
				t.Errorf("drew %d cards, want %d", got, tc.draws)
			}
			if got := lifeBefore - me.Life; got != tc.draws {
				t.Errorf("lost %d life, want %d — the draw and the loss are one clause", got, tc.draws)
			}
		})
	}
}

// CR 603.4: an intervening if gates whether the ability triggers at
// all, so a Hexblade paid for without Treasure mana never puts a
// trigger on the stack and nobody gets a window to respond to one.
func TestHiredHexbladeWithoutTreasureManaTriggersNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	floatForTest(g, me, "BB")

	castFromHandForTest(t, g, me, "Hired Hexblade", "Creature — Elf Warlock",
		"{1}{B}", hiredHexbladeOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)

	for _, ev := range g.Events {
		if ev.Kind == game.EventTrigger {
			t.Errorf("a trigger was announced for a Hexblade nothing triggered: %+v", ev)
		}
	}
}

// --- Gruul Scrapper: a COLOUR, read by a permanent ------------------

// {R} paid into the generic half of {3}{G} makes the Scrapper hasty;
// paying the same four mana without a red one does not.
func TestGruulScrapperGainsHasteOnlyWhenRedWasSpent(t *testing.T) {
	for _, tc := range []struct {
		name  string
		pool  string
		haste bool
	}{
		{"a Mountain paid part of the generic half", "GRCC", true},
		{"no red mana at all", "GCCC", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			floatForTest(g, me, tc.pool)

			id := castFromHandForTest(t, g, me, "Gruul Scrapper", "Creature — Human Berserker",
				"{3}{G}", gruulScrapperOracle, game.CastSpellParams{Strict: true})
			passPriorityAroundTable(t, g)

			card := b12Card(t, g, id)
			var got bool
			g.WithWriteLock(func() { got = game.HasKeyword(&card, "haste") })
			if got != tc.haste {
				t.Errorf("haste = %v, want %v", got, tc.haste)
			}
		})
	}
}

// --- Ribbons of Night: a COLOUR, read by the spell ------------------

// The damage and the life happen either way; only the draw is the
// rider, and it reads the record off the stack item while the spell is
// still resolving.
func TestRibbonsOfNightDrawsOnlyWhenBlueWasSpent(t *testing.T) {
	for _, tc := range []struct {
		name  string
		pool  string
		draws int
	}{
		{"an Island paid part of the generic half", "BUCCC", 1},
		{"no blue mana at all", "BCCCC", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			advanceToMain(t, g)
			floatForTest(g, me, tc.pool)

			victim := uuid.New()
			g.WithWriteLock(func() {
				g.Battlefield.PushTop(game.Card{
					InstanceID: victim, Name: "Bear", TypeLine: "Creature — Bear",
					Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
				})
			})
			handBefore, lifeBefore := me.Hand.Size(), me.Life

			castFromHandForTest(t, g, me, "Ribbons of Night", "Sorcery",
				"{4}{B}", ribbonsOfNightOracl, game.CastSpellParams{
					Strict:  true,
					Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
				})
			passPriorityAroundTable(t, g)

			if me.Life-lifeBefore != 4 {
				t.Errorf("gained %d life, want 4 — the life is not conditional", me.Life-lifeBefore)
			}
			if g.Battlefield.Contains(victim) {
				t.Error("the Bear survived 4 damage")
			}
			if got := me.Hand.Size() - handBefore; got != tc.draws {
				t.Errorf("drew %d cards, want %d", got, tc.draws)
			}
		})
	}
}

// --- the registry ---------------------------------------------------

// Every card this PR converted is registered under the oracle ID the
// Scryfall dump gives it, and the two that declare a source wish
// declare the right one.
func TestManaSourceCardsAreRegistered(t *testing.T) {
	for _, tc := range []struct {
		oracle string
		name   string
		wish   game.ManaSourceKinds
	}{
		{hiredHexbladeOracle, "Hired Hexblade", game.ManaSourceTreasure},
		{gruulScrapperOracle, "Gruul Scrapper", 0},
		{ribbonsOfNightOracl, "Ribbons of Night", 0},
	} {
		spec, ok := Lookup(tc.oracle)
		if !ok {
			t.Errorf("%s is not registered under %s", tc.name, tc.oracle)
			continue
		}
		if spec.Name != tc.name {
			t.Errorf("%s registered as %q", tc.oracle, spec.Name)
		}
		if spec.WantsManaFrom != tc.wish {
			t.Errorf("%s declares WantsManaFrom %b, want %b", tc.name, spec.WantsManaFrom, tc.wish)
		}
	}
}
