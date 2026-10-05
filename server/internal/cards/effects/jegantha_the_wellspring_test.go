package effects

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// jegantha_the_wellspring_test.go — #2170: "This mana can't be spent to
// pay generic mana costs." The engine half is game/mana_no_generic_test.go.

const jeganthaOracle = "f0587c55-06c1-4930-a62e-668f37464ce9"

func jeganthaTapped(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	id := pushCatalogPermanent(g, me.ID, "Jegantha, the Wellspring", "Legendary Creature — Elemental Elk", jeganthaOracle, false)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap Jegantha: %v", err)
	}
	return id
}

func TestJeganthaAddsRestrictedFiveColorMana(t *testing.T) {
	g, me, _ := spendTable(t)
	jeganthaTapped(t, g, me)
	if got := poolColors(me); !slices.Equal(got, []string{"W", "U", "B", "R", "G"}) {
		t.Fatalf("pool = %v, want W U B R G", got)
	}
	for _, tok := range me.ManaPool {
		if !slices.Equal(tok.Restrictions, []string{game.ManaRestrictNoGeneric}) {
			t.Errorf("token %+v is not restricted to non-generic costs", tok)
		}
	}
	spec, ok := Lookup(jeganthaOracle)
	if !ok || spec.Completeness != CompletenessCaveats || len(spec.Caveats) != 1 {
		t.Errorf("Jegantha is %v with caveats %v, want the one companion caveat", spec.Completeness, spec.Caveats)
	}
}

// The headline: WUBRG pays a five-colour spell and not {4}.
func TestJeganthaManaPaysAFiveColorSpellButNotGeneric(t *testing.T) {
	g, me, _ := spendTable(t)
	jeganthaTapped(t, g, me)
	four := handSpell(me, "Four", "Artifact", "{4}")
	five := handSpell(me, "Five Colors", "Sorcery", "{W}{U}{B}{R}{G}")
	refusedForMana(t, g.CastSpell(me.ID, four, game.CastSpellParams{Strict: true}), "Jegantha mana for {4}")
	if len(me.ManaPool) != 5 {
		t.Fatalf("a refused cast spent mana: pool %v", poolColors(me))
	}
	if err := g.CastSpell(me.ID, five, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("{W}{U}{B}{R}{G} off Jegantha: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool %v after the cast, want empty", poolColors(me))
	}
}

// Mixed with other mana: the lands pay the generic part, Jegantha's mana
// the coloured part.
func TestJeganthaManaMixedWithLandsPaysGenericFromTheLands(t *testing.T) {
	g, me, _ := spendTable(t)
	jeganthaTapped(t, g, me)
	basics(g, me, "Mountain", 2)
	spell := handSpell(me, "Mixed", "Sorcery", "{2}{W}{U}")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("{2}{W}{U} off two Mountains and Jegantha: %v", err)
	}
	if got := poolColors(me); !slices.Equal(got, []string{"B", "R", "G"}) {
		t.Errorf("pool %v, want the unspent B R G", got)
	}
}

// Commander tax is generic mana (CR 903.8).
func TestJeganthaManaCannotPayCommanderTax(t *testing.T) {
	g, me, _ := spendTable(t)
	jeganthaTapped(t, g, me)
	id := uuid.New()
	me.Command.PushTop(game.Card{
		InstanceID: id, Name: "Five Color Commander", TypeLine: "Legendary Creature — Dragon",
		ManaCost: "{W}{U}{B}{R}{G}", Owner: me.ID, Controller: me.ID, IsCommander: true,
	})
	me.CommanderCasts[id] = 1
	refusedForMana(t, g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "command", Strict: true}), "the taxed commander off Jegantha alone")
	g.WithWriteLock(func() {
		me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "command", Strict: true}); err != nil {
		t.Fatalf("the taxed commander with two ordinary mana for the tax: %v", err)
	}
}

func TestJeganthaManaAndXSpells(t *testing.T) {
	g, me, _ := spendTable(t)
	jeganthaTapped(t, g, me)
	x := handSpell(me, "X Spell", "Sorcery", "{X}{R}")
	refusedForMana(t, g.CastSpell(me.ID, x, game.CastSpellParams{Strict: true, XValue: 2}), "X=2 off Jegantha alone")
	if err := g.CastSpell(me.ID, x, game.CastSpellParams{Strict: true, XValue: 0}); err != nil {
		t.Fatalf("{X}{R} with X=0 off Jegantha: %v", err)
	}
}

// Kruphix, God of Horizons turns mana you would lose into colourless and
// keeps it; the converted mana keeps its restriction, so it pays {C} but
// not generic.
func TestKruphixConversionKeepsJeganthasRestriction(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Kruphix, God of Horizons", "Legendary Enchantment Creature — God", mkKruphixOracle, false)
	jeganthaTapped(t, g, me)
	mkAdvance(t, g)
	mkWantPool(t, me, map[string]int{"C": 5})
	for _, tok := range me.ManaPool {
		if !slices.Equal(tok.Restrictions, []string{game.ManaRestrictNoGeneric}) {
			t.Fatalf("converted token %+v lost the restriction", tok)
		}
	}
	four := handSpell(me, "Four", "Instant", "{4}")
	refusedForMana(t, g.CastSpell(me.ID, four, game.CastSpellParams{Strict: true}), "converted Jegantha mana for {4}")
	colorless := handSpell(me, "Colorless Symbols", "Instant", "{C}{C}")
	if err := g.CastSpell(me.ID, colorless, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("{C}{C} off converted Jegantha mana: %v", err)
	}
}

func TestJeganthaTokensSurviveASnapshotRoundTrip(t *testing.T) {
	g, me, _ := spendTable(t)
	jeganthaTapped(t, g, me)
	blob, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap game.GameSnapshot
	if err := json.Unmarshal(blob, &snap); err != nil {
		t.Fatal(err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	var pool game.ManaPool
	for _, p := range restored.Seats {
		if p.ID == me.ID {
			pool = p.ManaPool
		}
	}
	if len(pool) != 5 {
		t.Fatalf("restored pool %v", pool)
	}
	one, _ := game.ParseCost("{1}")
	if pool.CanPayFor(one, 0, game.ManaSpendContext{Purpose: game.SpendPurposeCast}) {
		t.Error("the restored token paid generic mana")
	}
}

// spellAgreement asks one board three questions about a hand spell, and
// fails unless they give one answer: the view's legal_actions digest, the
// bot's enumerator and the real (strict, auto-tapped) payment.
func spellAgreement(t *testing.T, name string, g *game.Game, me *game.Player, spell uuid.UUID, want bool) {
	t.Helper()
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].KnownBy == nil {
				me.Hand.Cards[i].KnownBy = map[uuid.UUID]bool{}
			}
			me.Hand.Cards[i].KnownBy[me.ID] = true
		}
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.KnownBy == nil {
				c.KnownBy = map[uuid.UUID]bool{}
			}
			for _, p := range g.Seats {
				c.KnownBy[p.ID] = true
			}
		}
	})
	view := false
	if v := protocol.ViewOfGameFor(g, me.ID.String()); v.LegalActions != nil {
		if src := v.LegalActions.Sources[spell.String()]; src != nil {
			view = slices.Contains(src.Kinds, legal.KindCast)
		}
	}
	bot := false
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type == "cast_spell" && m.Source == spell {
			bot = true
		}
	}
	err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true})
	if view != want || bot != want || (err == nil) != want {
		t.Errorf("%s: view %v, bot %v, payment %v — want all %v", name, view, bot, err, want)
	}
}

func TestJeganthaViewBotAndPaymentAgree(t *testing.T) {
	t.Run("five colours", func(t *testing.T) {
		g, me, _ := spendTable(t)
		jeganthaTapped(t, g, me)
		spell := handSpell(me, "Five Colors", "Sorcery", "{W}{U}{B}{R}{G}")
		spellAgreement(t, "five colours", g, me, spell, true)
	})
	t.Run("generic four", func(t *testing.T) {
		g, me, _ := spendTable(t)
		jeganthaTapped(t, g, me)
		spell := handSpell(me, "Four", "Artifact", "{4}")
		spellAgreement(t, "generic four", g, me, spell, false)
	})
	t.Run("two generic with two Mountains", func(t *testing.T) {
		g, me, _ := spendTable(t)
		jeganthaTapped(t, g, me)
		basics(g, me, "Mountain", 2)
		spell := handSpell(me, "Mixed", "Sorcery", "{2}{W}{U}")
		spellAgreement(t, "two generic with two Mountains", g, me, spell, true)
	})
	t.Run("two generic with one Mountain", func(t *testing.T) {
		g, me, _ := spendTable(t)
		jeganthaTapped(t, g, me)
		basics(g, me, "Mountain", 1)
		spell := handSpell(me, "Mixed", "Sorcery", "{2}{W}{U}")
		spellAgreement(t, "two generic with one Mountain", g, me, spell, false)
	})
	t.Run("Jegantha untapped is not planned", func(t *testing.T) {
		// A restricted mana ability is never auto-tapped, so an untapped
		// Jegantha funds nothing: all three answers agree on "no".
		g, me, _ := spendTable(t)
		pushCatalogPermanent(g, me.ID, "Jegantha, the Wellspring", "Legendary Creature — Elemental Elk", jeganthaOracle, false)
		spell := handSpell(me, "Five Colors", "Sorcery", "{W}{U}{B}{R}{G}")
		spellAgreement(t, "untapped Jegantha", g, me, spell, false)
	})
}
