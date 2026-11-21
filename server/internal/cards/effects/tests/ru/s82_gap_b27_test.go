package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	mirrisGuileOracle       = "7f89c0ee-b914-406c-8a3f-98424a52ae14"
	triumphOfFerocityOracle = "eabcbaca-6df9-455a-85e5-32d6ad4d1b44"
	counterbalanceOracle    = "088dfd02-c0e0-41d1-9923-efe023fb2ad1"
	penumbraUmbraOracle     = "5cce6146-94a6-442e-9890-6f2914cbcebb"
)

func TestMirrisGuileLooksAtTheTopThreeOnYourUpkeep(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.PushCatalogPermanent(g, me.ID, "Mirri's Guile", "Enchantment", mirrisGuileOracle, false)
		effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
		effectstest.WalkToUpkeepOfSeat0(t, g)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.LookAtTopChoiceFor(g, me.ID)
		if !yes {
			if c != nil {
				t.Error("declined Mirri's Guile still looked at cards")
			}
			continue
		}
		if c == nil || len(c.ScryCards) != 3 {
			t.Fatalf("look-at prompt = %+v, want three cards", c)
		}
		rev := []uuid.UUID{c.ScryCards[2], c.ScryCards[1], c.ScryCards[0]}
		if err := g.ResolveLookAtTop(c.ID, me.ID, rev); err != nil {
			t.Fatalf("ResolveLookAtTop: %v", err)
		}
		if got := effectstest.LibraryTopNames(me, 3); got[0] != "C" || got[2] != "A" {
			t.Errorf("library top is %v after reversing the three, want C B A", got)
		}
	}
}

func TestTriumphOfFerocityDrawsOnlyForTheGreatestPower(t *testing.T) {
	for _, tc := range []struct {
		name         string
		mine, theirs int
		draws        bool
	}{
		{"mine is bigger", 4, 2, true},
		{"tied for the greatest", 3, 3, true},
		{"theirs is bigger", 3, 4, false},
	} {
		g := effectstest.NewCatalogGame(t)
		me, them := g.Seats[0], g.Seats[1]
		effectstest.PushCatalogPermanent(g, me.ID, "Triumph of Ferocity", "Enchantment", triumphOfFerocityOracle, false)
		effectstest.PushVanillaCreature(g, me.ID, "Mine", tc.mine, tc.mine)
		effectstest.PushVanillaCreature(g, them.ID, "Theirs", tc.theirs, tc.theirs)
		effectstest.WalkToUpkeepOfSeat0(t, g)
		before := len(me.Hand.Cards)
		effectstest.PassPriorityAroundTable(t, g)
		got := len(me.Hand.Cards) - before
		if tc.draws && got != 1 || !tc.draws && got != 0 {
			t.Errorf("%s: drew %d cards, want draws=%v", tc.name, got, tc.draws)
		}
	}
}

func TestTriumphOfFerocityDrawsNothingWithNoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Triumph of Ferocity", "Enchantment", triumphOfFerocityOracle, false)
	effectstest.WalkToUpkeepOfSeat0(t, g)
	before := len(me.Hand.Cards)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards) - before; got != 0 {
		t.Errorf("drew %d with no creatures on the battlefield", got)
	}
}

func TestCounterbalanceCountersASpellWithTheRevealedManaValue(t *testing.T) {
	for _, tc := range []struct {
		name    string
		topCost string
		yes     bool
		counter bool
	}{
		{"same mana value", "{1}{G}", true, true},
		{"different mana value", "{3}", true, false},
		{"declined", "{1}{G}", false, false},
	} {
		g := effectstest.NewCatalogGame(t)
		me, them := g.Seats[0], g.Seats[1]
		effectstest.PushCatalogPermanent(g, me.ID, "Counterbalance", "Enchantment", counterbalanceOracle, false)
		effectstest.SeedLibraryTop(me, "Top Card", "Creature — Bear", tc.topCost)
		effectstest.AdvanceToMain(t, g)
		rock := game.NewCard("Test Rock", them.ID)
		rock.TypeLine = "Artifact"
		rock.ManaCost = "{2}"
		rock.Keywords = []string{"flash"}
		them.Hand.PushTop(rock)
		effectstest.DcMana(t, g, them.ID, "{C}{C}")
		if err := g.CastSpell(them.ID, rock.InstanceID, game.CastSpellParams{}); err != nil {
			t.Fatalf("%s: cast: %v", tc.name, err)
		}
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, tc.yes)
		effectstest.PassPriorityAroundTable(t, g)
		onField := g.Battlefield.Contains(rock.InstanceID)
		if tc.counter == onField {
			t.Errorf("%s: spell on the battlefield = %v, want countered = %v", tc.name, onField, tc.counter)
		}
		if tc.counter && !them.Graveyard.Contains(rock.InstanceID) {
			t.Errorf("%s: the countered spell is not in its owner's graveyard", tc.name)
		}
		if top := effectstest.LibraryTopNames(me, 1); len(top) == 0 || top[0] != "Top Card" {
			t.Errorf("%s: the revealed card did not stay on top: %v", tc.name, top)
		}
	}
}

func TestCounterbalanceIgnoresYourOwnSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Counterbalance", "Enchantment", counterbalanceOracle, false)
	effectstest.SeedLibraryTop(me, "Top Card", "Creature — Bear", "{2}")
	rock := game.NewCard("Test Rock", me.ID)
	rock.TypeLine = "Artifact"
	rock.ManaCost = "{2}"
	if _, err := gapCastFromHandNoMana(t, g, rock); err != nil {
		t.Fatalf("cast: %v", err)
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt {
			t.Fatal("Counterbalance triggered off its controller's own spell")
		}
	}
}

func gapCastFromHandNoMana(t *testing.T, g *game.Game, c game.Card) (uuid.UUID, error) {
	t.Helper()
	p := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	p.Hand.PushTop(c)
	effectstest.DcMana(t, g, p.ID, "{C}{C}")
	return c.InstanceID, g.CastSpell(p.ID, c.InstanceID, game.CastSpellParams{})
}

func TestPenumbraUmbraCopiesTheCreatureInBlackWhenItsAuraIsDestroyed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Runeclaw Bear", TypeLine: "Creature — Bear", Colors: []string{"G"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	umbra := effectstest.PushAuraOnLand(t, g, me.ID, "Penumbra Umbra", penumbraUmbraOracle, bear)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatal(err)
		}
	})
	if !g.Battlefield.Contains(bear) || !me.Graveyard.Contains(umbra) {
		t.Fatal("umbra armor should destroy the Aura instead of the creature")
	}
	effectstest.PassPriorityAroundTable(t, g)
	var token *game.Card
	for i := range g.Battlefield.Cards {
		c := g.Battlefield.Cards[i]
		if c.IsToken() && c.Name == "Runeclaw Bear" {
			token = &g.Battlefield.Cards[i]
		}
	}
	if token == nil {
		t.Fatal("no token copy of the enchanted creature was created")
	}
	if cols := token.EffectiveColors(); len(cols) != 1 || cols[0] != "B" {
		t.Errorf("token colours = %v, want black only", cols)
	}
	if token.Controller != me.ID {
		t.Errorf("the token is controlled by %v, want the Aura's controller", token.Controller)
	}
	if p := effectstest.EffectivePower(t, g, token.InstanceID); p != 2 {
		t.Errorf("token power = %d, want 2", p)
	}
}
