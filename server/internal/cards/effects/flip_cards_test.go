package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flip_cards_test.go — the three proof cards of ADR 0151 (#1892), each
// down the real import road, so the per-half keywords, the colour and
// cost of the flipped half and the "#1" catalog key are the importer's.
// The engine verb is pinned in game/flip_test.go.

// flipRow is a flip card's Scryfall record: the colour and cost at the
// top level and on the top half, the flipped half with neither.
func flipRow(oracleID, cost string, colors, keywords []string, top, flipped cards.CardFace) cards.Card {
	top.ManaCost = cost
	return cards.Card{
		ID:            uuid.New(),
		OracleID:      uuid.MustParse(oracleID),
		Name:          top.Name + " // " + flipped.Name,
		Layout:        "flip",
		TypeLine:      top.TypeLine + " // " + flipped.TypeLine,
		ManaCost:      cost,
		Colors:        colors,
		ColorIdentity: colors,
		Keywords:      keywords,
		CardFaces:     []cards.CardFace{top, flipped},
	}
}

func runeTailRow() cards.Card {
	return flipRow(runeTailOracleID, "{2}{W}", []string{"W"}, nil,
		cards.CardFace{Name: "Rune-Tail, Kitsune Ascendant", TypeLine: "Legendary Creature — Fox Monk",
			OracleText: "When you have 30 or more life, flip Rune-Tail.", Power: "2", Toughness: "2"},
		cards.CardFace{Name: "Rune-Tail's Essence", TypeLine: "Legendary Enchantment",
			OracleText: "Prevent all damage that would be dealt to creatures you control."})
}

func akkiLavarunnerRow() cards.Card {
	return flipRow(akkiLavarunnerOracleID, "{3}{R}", []string{"R"}, []string{"Protection", "Haste"},
		cards.CardFace{Name: "Akki Lavarunner", TypeLine: "Creature — Goblin Warrior",
			OracleText: "Haste\nWhenever this creature deals damage to an opponent, flip it.", Power: "1", Toughness: "1"},
		cards.CardFace{Name: "Tok-Tok, Volcano Born", TypeLine: "Legendary Creature — Goblin Shaman",
			OracleText: "Protection from red\nIf a red source would deal damage to a player, it deals that much damage plus 1 to that player instead.",
			Power:      "2", Toughness: "2"})
}

func homuraRow() cards.Card {
	return flipRow(homuraOracleID, "{4}{R}{R}", []string{"R"}, nil,
		cards.CardFace{Name: "Homura, Human Ascendant", TypeLine: "Legendary Creature — Human Monk",
			OracleText: "Homura can't block.\nWhen Homura dies, return it to the battlefield flipped.", Power: "4", Toughness: "4"},
		cards.CardFace{Name: "Homura's Essence", TypeLine: "Legendary Enchantment",
			OracleText: "Creatures you control get +2/+2 and have flying and \"{R}: This creature gets +1/+0 until end of turn.\""})
}

// flipCardNow reads the permanent of the oracle ID with its layers
// current.
func flipCardNow(t *testing.T, g *game.Game, oracle string) game.Card {
	t.Helper()
	var out game.Card
	found := false
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].OracleID == oracle {
				out, found = g.Battlefield.Cards[i], true
			}
		}
	})
	if !found {
		t.Fatalf("no permanent with oracle %s", oracle)
	}
	return out
}

// Rune-Tail flips when its controller reaches 30 life (a CR 603.8 state
// trigger) and the Essence prevents damage to its controller's
// creatures only. It stays white and {2}{W} (CR 710.1c).
func TestRuneTailFlipsAtThirtyLifeAndPreventsDamageToYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	stSetLife(t, g, me, 20) // a Commander game starts at 40
	rt := importToBattlefield(t, g, runeTailRow(), me)
	bear := apaPush(g, me.ID, me.ID, apaCreature("My Bear", "oracle-my-bear", 2, 2))
	theirs := apaPush(g, opp.ID, opp.ID, apaCreature("Their Bear", "oracle-their-bear", 2, 2))

	stSetLife(t, g, me, 29)
	passPriorityAroundTable(t, g)
	if c := flipCardNow(t, g, runeTailOracleID); c.Flipped {
		t.Fatal("flipped at 29 life")
	}

	stSetLife(t, g, me, 30)
	passPriorityAroundTable(t, g)
	c := flipCardNow(t, g, runeTailOracleID)
	if !c.Flipped || c.Name != "Rune-Tail's Essence" || c.InstanceID != rt {
		t.Fatalf("at 30 life: flipped %v %q (same object %v)", c.Flipped, c.Name, c.InstanceID == rt)
	}
	if c.IsCreature() || c.ManaValue() != 3 || !slices.Equal(c.Effective().Colors, []string{"W"}) {
		t.Errorf("Essence: creature %v mv %d colours %v; want a white enchantment of mana value 3", c.IsCreature(), c.ManaValue(), c.Effective().Colors)
	}

	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, bear, 3); err != nil {
			t.Fatal(err)
		}
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, theirs, 1); err != nil {
			t.Fatal(err)
		}
	})
	if b, ok := findBattlefieldByID(g, bear); !ok || b.DamageMarked != 0 {
		t.Errorf("damage to my creature was not prevented: %+v", b.DamageMarked)
	}
	if b, ok := findBattlefieldByID(g, theirs); !ok || b.DamageMarked != 1 {
		t.Errorf("damage to an opponent's creature was prevented: %d", b.DamageMarked)
	}

	// Falling below 30 and coming back does not flip it back (CR 710.4).
	stSetLife(t, g, me, 10)
	stSetLife(t, g, me, 40)
	passPriorityAroundTable(t, g)
	if c := flipCardNow(t, g, runeTailOracleID); !c.Flipped {
		t.Error("the Essence unflipped")
	}
}

// Akki Lavarunner flips when it deals damage to an opponent. Flipped,
// it loses haste, gains protection from red, becomes legendary (CR
// 710.2's own example, read by a legendary anthem across the flip with
// no other event between), and red sources deal +1 to players.
func TestAkkiLavarunnerFlipsIntoTokTok(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	akki := importToBattlefield(t, g, akkiLavarunnerRow(), me)
	apaPush(g, me.ID, me.ID, game.Card{Name: "Rising of the Day", OracleID: "434a20f8-3f87-4004-9155-0f196fc2257e", TypeLine: "Enchantment"})

	c := flipCardNow(t, g, akkiLavarunnerOracleID)
	if !slices.Contains(c.Keywords, "haste") || game.HasKeyword(&c, "protection from red") {
		t.Errorf("top half keywords: %v", c.Keywords)
	}
	if p := c.Effective().Power; p != 1 {
		t.Errorf("nonlegendary Akki Lavarunner has power %d, want 1 (the anthem is for legendary creatures)", p)
	}

	// Damage to a creature or to its own controller does not flip it.
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(akki, me.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if flipCardNow(t, g, akkiLavarunnerOracleID).Flipped {
		t.Fatal("damage to its own controller flipped it")
	}

	life := opp.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(akki, opp.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	if opp.Life != life-1 {
		t.Fatalf("opponent life %d, want %d (no +1 before the flip)", opp.Life, life-1)
	}
	passPriorityAroundTable(t, g)

	c = flipCardNow(t, g, akkiLavarunnerOracleID)
	if !c.Flipped || c.Name != "Tok-Tok, Volcano Born" || c.InstanceID != akki {
		t.Fatalf("after damaging an opponent: flipped %v %q", c.Flipped, c.Name)
	}
	// Printed keywords: Rising of the Day's own haste grant is layered on
	// top of the effective list, so read the half's printed ones.
	if slices.Contains(c.Keywords, "haste") || !game.HasKeyword(&c, "protection from red") {
		t.Errorf("Tok-Tok printed keywords: %v, want protection from red and no haste", c.Keywords)
	}
	if eff := c.Effective(); eff.Power != 3 || eff.Toughness != 2 {
		t.Errorf("Tok-Tok is %d/%d, want 3/2 (2/2 plus the legendary anthem)", eff.Power, eff.Toughness)
	}
	if c.ManaValue() != 4 || !slices.Equal(c.Effective().Colors, []string{"R"}) {
		t.Errorf("Tok-Tok mv %d colours %v; want 4 and red (CR 710.1c)", c.ManaValue(), c.Effective().Colors)
	}

	// A red source deals +1 to any player; a colourless source does not.
	life = opp.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(akki, opp.ID, 2); err != nil {
			t.Fatal(err)
		}
	})
	if opp.Life != life-3 {
		t.Errorf("red source: opponent lost %d, want 3", life-opp.Life)
	}
	mine := me.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(akki, me.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	if me.Life != mine-2 {
		t.Errorf("red source to Tok-Tok's own controller: lost %d, want 2", mine-me.Life)
	}
	life = opp.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(uuid.Nil, opp.ID, 2); err != nil {
			t.Fatal(err)
		}
	})
	if opp.Life != life-2 {
		t.Errorf("sourceless damage: opponent lost %d, want 2", life-opp.Life)
	}
}

// Homura can't block; when it dies it returns flipped as Homura's
// Essence, a new object, under its owner's control, and the Essence
// pumps, gives flying to and grants firebreathing to your creatures.
func TestHomuraReturnsFlippedAsItsEssence(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	homura := importToBattlefield(t, g, homuraRow(), me)
	bear := apaPush(g, me.ID, me.ID, apaCreature("My Bear", "oracle-my-bear", 2, 2))
	theirs := apaPush(g, opp.ID, opp.ID, apaCreature("Their Bear", "oracle-their-bear", 2, 2))

	if c := flipCardNow(t, g, homuraOracleID); c.Effective().Restrictions&game.CantBlock == 0 {
		t.Error("Homura can block")
	}

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(homura); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)

	c := flipCardNow(t, g, homuraOracleID)
	if !c.Flipped || c.ActiveFace != 1 || c.Name != "Homura's Essence" {
		t.Fatalf("returned as %q (flipped %v face %d), want Homura's Essence flipped", c.Name, c.Flipped, c.ActiveFace)
	}
	if c.Controller != me.ID || c.IsCreature() {
		t.Errorf("controller %v creature %v, want its owner's enchantment", c.Controller, c.IsCreature())
	}
	if c.ManaValue() != 6 || !slices.Equal(c.Effective().Colors, []string{"R"}) {
		t.Errorf("Essence mv %d colours %v; want 6 and red (CR 710.1c)", c.ManaValue(), c.Effective().Colors)
	}
	if me.Graveyard.Contains(homura) {
		t.Error("the card is still in the graveyard")
	}

	var b, o game.Card
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		b = *findBattlefieldCardForTest(g, bear)
		o = *findBattlefieldCardForTest(g, theirs)
	})
	eff := b.Effective()
	if eff.Power != 4 || eff.Toughness != 4 || !game.HasKeyword(&b, "flying") {
		t.Errorf("my bear is %d/%d flying %v, want 4/4 with flying", eff.Power, eff.Toughness, game.HasKeyword(&b, "flying"))
	}
	if !slices.ContainsFunc(eff.GrantedAbilities, func(a game.GrantedAbility) bool { return a.Key == game.GrantKey(homuraGrant) }) {
		t.Errorf("my bear has no granted firebreathing: %+v", eff.GrantedAbilities)
	}
	if oe := o.Effective(); oe.Power != 2 || game.HasKeyword(&o, "flying") || len(oe.GrantedAbilities) != 0 {
		t.Errorf("the opponent's bear was affected: %d/%d %+v", oe.Power, oe.Toughness, oe.GrantedAbilities)
	}
}

// If Homura's card leaves the graveyard in response, nothing returns
// (CR 400.7).
func TestHomuraReturnsNothingIfItLeftTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	homura := importToBattlefield(t, g, homuraRow(), me)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(homura); err != nil {
			t.Fatal(err)
		}
		if err := g.ExileCardForEffect(homura); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == homuraOracleID {
			t.Fatal("Homura returned from exile")
		}
	}
	if c, ok := g.LookupCardForEffect(homura); !ok || c.Flipped || c.ActiveFace != 0 {
		t.Errorf("the exiled card is flipped %v face %d, want unflipped", c.Flipped, c.ActiveFace)
	}
}
