package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// yshtola_deck_sweep_c_test.go — the third slice of the "Y'shtola,
// Night's Blessed" deck request (#3657): Summon: Good King Mog XII and
// Hildibrand Manderville // Gentleman's Rise.

const (
	ysCMogOracle       = "2148820c-85b6-4598-adf2-10873001a779"
	ysCHildibrandOracl = "6df03db9-e4b3-4931-94e3-c582d2b04110"
)

func ysCCastMog(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := effectstest.TdCastFromHand(t, g, game.Card{Name: "Summon: Good King Mog XII", TypeLine: "Enchantment Creature — Saga Moogle",
		OracleID: ysCMogOracle, ManaCost: "{4}{W}", Power: 4, Toughness: 4}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestYsCMogChapterIMakesTwoLifelinkMoogles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mog := ysCCastMog(t, g)
	if n := effectstest.CountBattlefieldByName(g, "Moogle"); n != 2 {
		t.Fatalf("Moogle tokens %d, want 2", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Moogle" {
			v := effectstest.S58p6View(t, g, c.InstanceID)
			if v.CurrentPower() != 1 || v.CurrentToughness() != 2 || !game.HasKeyword(&v, "lifelink") || c.Controller != me.ID {
				t.Errorf("Moogle is %d/%d lifelink=%v, want 1/2 lifelink under my control", v.CurrentPower(), v.CurrentToughness(), game.HasKeyword(&v, "lifelink"))
			}
		}
	}
	if effectstest.LoreCountersOn(g, mog) != 1 {
		t.Errorf("lore counters %d, want 1", effectstest.LoreCountersOn(g, mog))
	}
}

func TestYsCMogChapterTwoCopiesATokenForEachNoncreatureSpellThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	mog := ysCCastMog(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat) // chapter II, on a later turn of mine
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LoreCountersOn(g, mog) != 2 {
		t.Fatalf("lore counters %d, want 2", effectstest.LoreCountersOn(g, mog))
	}
	before := effectstest.CountBattlefieldByName(g, "Moogle")

	effectstest.B24CastNoncreature(t, g, "Spell", "Instant", "{U}")
	effectstest.DrcSettle(t, g, g.Seats[seat].ID)
	if got := effectstest.CountBattlefieldByName(g, "Moogle"); got != before+1 {
		t.Errorf("Moogles %d → %d after a noncreature spell, want one more", before, got)
	}
	effectstest.CastAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	effectstest.DrcSettle(t, g, g.Seats[seat].ID)
	if got := effectstest.CountBattlefieldByName(g, "Moogle"); got != before+1 {
		t.Errorf("a creature spell copied a token: %d Moogles", got)
	}
	effectstest.B24CastNoncreature(t, g, "Another", "Instant", "{U}")
	effectstest.DrcSettle(t, g, g.Seats[seat].ID)
	if got := effectstest.CountBattlefieldByName(g, "Moogle"); got != before+2 {
		t.Errorf("each spell this turn copies: %d Moogles, want %d", got, before+2)
	}
}

func TestYsCMogChapterFourPutsTwoCountersOnEachOtherMoogle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	mog := ysCCastMog(t, g)
	for i := 0; i < 3; i++ {
		effectstest.AdvanceToPrecombatMainOf(t, g, seat)
		effectstest.DrcSettle(t, g, g.Seats[seat].ID)
	}
	for _, c := range g.Battlefield.Cards {
		switch {
		case c.Name == "Moogle":
			if got := effectstest.CountersOn(g, c.InstanceID, game.CounterPlusOne); got != 2 {
				t.Errorf("a Moogle token has %d counters, want 2", got)
			}
		case c.InstanceID == mog:
			t.Error("the Saga should be sacrificed after chapter IV")
		}
	}
}

func ysCHildibrand(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(), OracleID: ysCHildibrandOracl, Layout: game.LayoutAdventure,
		Owner: owner, Controller: owner,
		Faces: []game.Face{
			{Name: "Hildibrand Manderville", TypeLine: "Legendary Creature — Human Detective", ManaCost: "{1}{W}", Colors: []string{"W"}, Power: 2, Toughness: 2},
			{Name: "Gentleman's Rise", TypeLine: "Instant — Adventure", ManaCost: "{2}{B}", Colors: []string{"B"}},
		},
	}
	c.SetFace(0)
	return c
}

func TestYsCHildibrandPumpsCreatureTokensOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hild := ysCHildibrand(me.ID)
	hild.Power, hild.Toughness = 2, 2
	g.Battlefield.PushTop(hild)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: hild.InstanceID, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
	})
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 white Soldier"), 1) })
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	if v := effectstest.S58p6View(t, g, bear); v.CurrentPower() != 2 {
		t.Errorf("a nontoken creature got the anthem: %d power", v.CurrentPower())
	}
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Controller == me.ID {
			if v := effectstest.S58p6View(t, g, c.InstanceID); v.CurrentPower() != 2 || v.CurrentToughness() != 2 {
				t.Errorf("token is %d/%d, want 2/2", v.CurrentPower(), v.CurrentToughness())
			}
		}
	}
}

func TestYsCGentlemansRiseMakesAZombieAndAdventuresIntoExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hild := ysCHildibrand(me.ID)
	me.Hand.PushTop(hild)
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, hild.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("Gentleman's Rise: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Zombie"); n != 1 {
		t.Errorf("Zombie tokens %d, want 1", n)
	}
	if !effectstest.InExile(g, hild.InstanceID) {
		t.Error("the card did not go to exile on an adventure")
	}
}

func TestYsCHildibrandCanBeCastAsAnAdventureFromTheGraveyardAfterItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hild := ysCHildibrand(me.ID)
	hild.Power, hild.Toughness = 2, 2
	g.Battlefield.PushTop(hild)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: hild.InstanceID, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
	})
	effectstest.KillCreature(t, g, me.ID, hild.InstanceID)
	effectstest.DrcSettle(t, g, me.ID)
	effectstest.AdvanceToMain(t, g)
	if !me.Graveyard.Contains(hild.InstanceID) {
		t.Fatal("setup: Hildibrand should be in the graveyard")
	}

	// Only the Adventure half is castable from there: the creature is not.
	if err := g.CastSpell(me.ID, hild.InstanceID, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Error("the creature half was castable from the graveyard")
	}
	if err := g.CastSpell(me.ID, hild.InstanceID, game.CastSpellParams{FromZone: "graveyard", Face: 1}); err != nil {
		t.Fatalf("casting Gentleman's Rise from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Zombie"); n != 1 {
		t.Errorf("Zombie tokens %d, want 1", n)
	}
	if !effectstest.InExile(g, hild.InstanceID) {
		t.Error("the Adventure did not go to exile after resolving from the graveyard")
	}
}
