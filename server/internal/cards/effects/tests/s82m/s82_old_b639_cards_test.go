package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b639_cards_test.go — Planeshift, batch 639 (#5615). Each test plays
// the real catalog entry and asserts what its text says.

const (
	b639FavorOracle    = "1990753c-06a7-4e1f-bc0e-905a9ba6e293"
	b639GriffinOracle  = "4d74c735-8035-4c42-bfb1-1bc97e5bdd9a"
	b639ScornOracle    = "52cebfb0-0302-40d9-9a3f-f875dbf43659"
	b639DiscipleOracle = "778fb830-2d2b-4f81-a193-e6b652981695"
	b639MischiefOracle = "78e87805-19e3-415b-ad0b-3275183d7297"
	b639ElderOracle    = "8feb27aa-1f62-46e5-ba20-72c712de0862"
	b639ImpOracle      = "d2b1b4dd-e860-40d8-8e07-1dee5be65bc4"
	b639BearType       = "Creature — Bear"
)

func TestOldB639EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b639FavorOracle:    "Planeswalker's Favor",
		b639GriffinOracle:  "Aurora Griffin",
		b639ScornOracle:    "Planeswalker's Scorn",
		b639DiscipleOracle: "Disciple of Kangee",
		b639MischiefOracle: "Planeswalker's Mischief",
		b639ElderOracle:    "Samite Elder",
		b639ImpOracle:      "Volcano Imp",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func b639Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func b639Player(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetPlayer, ID: id} }

func b639Activate(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, index int, mana string, targets ...game.TargetRef) error {
	t.Helper()
	effectstest.DcMana(t, g, me.ID, mana)
	err := effectstest.InvB1Activate(g, me.ID, id, index, targets...)
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return err
}

func b639WantPT(t *testing.T, g *game.Game, id uuid.UUID, p, tough int) {
	t.Helper()
	if gp, gt := effectstest.PtOf(t, g, id); gp != p || gt != tough {
		t.Errorf("P/T is %d/%d, want %d/%d", gp, gt, p, tough)
	}
}

func TestOldB639PlaneswalkersFavorPumpsByTheRevealedCardsManaValue(t *testing.T) {
	g, me, opp := b639Table(t)
	favor := effectstest.B12Push(g, me.ID, "Planeswalker's Favor", "Enchantment", b639FavorOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", b639BearType, 2, 2)
	opp.Hand.Cards = nil
	effectstest.PutInHand(opp, game.Card{Name: "Four Drop", TypeLine: "Sorcery", ManaCost: "{3}{R}"})
	targets := append([]game.TargetRef{b639Player(opp.ID)}, effectstest.CardTarget(bear)...)
	if err := b639Activate(t, g, me, favor, 0, "{G}{G}{G}{G}", targets...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	b639WantPT(t, g, bear, 6, 6)
	if opp.Hand.Size() != 1 {
		t.Errorf("the revealed card left the opponent's hand")
	}
}

func TestOldB639PlaneswalkersFavorWithAnEmptyHandPumpsNothing(t *testing.T) {
	g, me, opp := b639Table(t)
	favor := effectstest.B12Push(g, me.ID, "Planeswalker's Favor", "Enchantment", b639FavorOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", b639BearType, 2, 2)
	opp.Hand.Cards = nil
	targets := append([]game.TargetRef{b639Player(opp.ID)}, effectstest.CardTarget(bear)...)
	if err := b639Activate(t, g, me, favor, 0, "{G}{G}{G}{G}", targets...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	b639WantPT(t, g, bear, 2, 2)
}

func TestOldB639PlaneswalkersFavorNeedsAnOpponentAsItsPlayerTarget(t *testing.T) {
	g, me, _ := b639Table(t)
	favor := effectstest.B12Push(g, me.ID, "Planeswalker's Favor", "Enchantment", b639FavorOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", b639BearType, 2, 2)
	targets := append([]game.TargetRef{b639Player(me.ID)}, effectstest.CardTarget(bear)...)
	if err := b639Activate(t, g, me, favor, 0, "{G}{G}{G}{G}", targets...); err == nil {
		t.Fatal("the controller was accepted as the target opponent")
	}
}

func TestOldB639PlaneswalkersScornShrinksByTheRevealedCardsManaValue(t *testing.T) {
	g, me, opp := b639Table(t)
	scorn := effectstest.B12Push(g, me.ID, "Planeswalker's Scorn", "Enchantment", b639ScornOracle, 0, 0)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", b639BearType, 3, 3)
	opp.Hand.Cards = nil
	effectstest.PutInHand(opp, game.Card{Name: "Three Drop", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 2, Toughness: 2})
	targets := append([]game.TargetRef{b639Player(opp.ID)}, effectstest.CardTarget(victim)...)
	if err := b639Activate(t, g, me, scorn, 0, "{B}{B}{B}{B}", targets...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effectstest.BattlefieldHas(g, victim) {
		t.Error("a 3/3 given -3/-3 survived")
	}
}

func TestOldB639PlaneswalkersMischiefExilesAnInstantAndLetsYouCastItFree(t *testing.T) {
	g, me, opp := b639Table(t)
	mischief := effectstest.B12Push(g, me.ID, "Planeswalker's Mischief", "Enchantment", b639MischiefOracle, 0, 0)
	opp.Hand.Cards = nil
	bolt := effectstest.PutInHand(opp, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: effectstest.LightningBoltOracle})
	if err := b639Activate(t, g, me, mischief, 0, "{U}{U}{U}{U}", b639Player(opp.ID)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Exile.Contains(bolt) || opp.Hand.Contains(bolt) {
		t.Fatal("the revealed instant was not exiled")
	}
	life := opp.Life
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{FromZone: "exile", Targets: []game.TargetRef{b639Player(opp.ID)}}); err != nil {
		t.Fatalf("casting the exiled instant without paying: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("the free Bolt dealt %d, want 3", got)
	}
	// Cast, so the end step has nothing to return.
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Contains(bolt) {
		t.Error("a cast card was returned to its owner's hand")
	}
}

func TestOldB639PlaneswalkersMischiefReturnsAnUncastCardAtTheNextEndStep(t *testing.T) {
	g, me, opp := b639Table(t)
	mischief := effectstest.B12Push(g, me.ID, "Planeswalker's Mischief", "Enchantment", b639MischiefOracle, 0, 0)
	opp.Hand.Cards = nil
	bolt := effectstest.PutInHand(opp, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: effectstest.LightningBoltOracle})
	if err := b639Activate(t, g, me, mischief, 0, "{U}{U}{U}{U}", b639Player(opp.ID)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Exile.Contains(bolt) {
		t.Fatal("the revealed instant was not exiled")
	}
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(bolt) || !opp.Hand.Contains(bolt) {
		t.Error("the uncast card was not returned to its owner's hand")
	}
}

func TestOldB639PlaneswalkersMischiefLeavesAnyOtherCardInHand(t *testing.T) {
	g, me, opp := b639Table(t)
	mischief := effectstest.B12Push(g, me.ID, "Planeswalker's Mischief", "Enchantment", b639MischiefOracle, 0, 0)
	opp.Hand.Cards = nil
	bear := effectstest.PutInHand(opp, game.Card{Name: "Grizzly", TypeLine: b639BearType, ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	if err := b639Activate(t, g, me, mischief, 0, "{U}{U}{U}{U}", b639Player(opp.ID)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !opp.Hand.Contains(bear) || g.Exile.Contains(bear) {
		t.Error("a creature card was exiled by Planeswalker's Mischief")
	}
}

func TestOldB639AuroraGriffinMakesAPermanentWhite(t *testing.T) {
	g, me, opp := b639Table(t)
	griffin := effectstest.B27Push(g, me.ID, "Aurora Griffin", "Creature — Griffin", b639GriffinOracle, "{3}{W}", 2, 2, "W")
	target := effectstest.B27Push(g, opp.ID, "Red Bear", b639BearType, "", "{1}{R}", 2, 2, "R")
	if err := b639Activate(t, g, me, griffin, 0, "{W}", effectstest.CardTarget(target)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := effectstest.TmtB9Colors(t, g, target); got != "W" {
		t.Errorf("colors = %q, want only white", got)
	}
}

func TestOldB639DiscipleOfKangeeGivesFlyingAndBlue(t *testing.T) {
	g, me, opp := b639Table(t)
	disciple := effectstest.B27Push(g, me.ID, "Disciple of Kangee", "Creature — Human Wizard", b639DiscipleOracle, "{2}{W}", 2, 2, "W")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == disciple {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	target := effectstest.B27Push(g, opp.ID, "Red Bear", b639BearType, "", "{1}{R}", 2, 2, "R")
	if err := b639Activate(t, g, me, disciple, 0, "{U}", effectstest.CardTarget(target)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !effectstest.HasEffectiveKeyword(t, g, target, "flying") {
		t.Error("the target did not gain flying")
	}
	if got := effectstest.TmtB9Colors(t, g, target); got != "U" {
		t.Errorf("colors = %q, want only blue", got)
	}
	if !effectstest.FindCardAnywhere(t, g, disciple).Tapped {
		t.Error("the {T} cost was not paid")
	}
}

func TestOldB639SamiteElderGivesYourCreaturesProtectionFromThePermanentsColors(t *testing.T) {
	g, me, opp := b639Table(t)
	elder := effectstest.B27Push(g, me.ID, "Samite Elder", "Creature — Human Cleric", b639ElderOracle, "{2}{W}", 1, 2, "W")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == elder {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	mine := effectstest.B12Creature(g, me.ID, "Mine", b639BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", b639BearType, 2, 2)
	multi := effectstest.B27Push(g, me.ID, "Gold Card", "Artifact", "", "{R}{G}", 0, 0, "R", "G")
	if err := b639Activate(t, g, me, elder, 0, "", effectstest.CardTarget(multi)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for _, id := range []uuid.UUID{mine, elder} {
		for _, kw := range []string{"protection from red", "protection from green"} {
			if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
				t.Errorf("one of your creatures lacks %s", kw)
			}
		}
		if effectstest.HasEffectiveKeyword(t, g, id, "protection from white") {
			t.Error("a creature gained protection from a colour the permanent is not")
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "protection from red") {
		t.Error("an opponent's creature gained protection")
	}
}

func TestOldB639VolcanoImpGainsFirstStrike(t *testing.T) {
	g, me, _ := b639Table(t)
	imp := effectstest.B12Push(g, me.ID, "Volcano Imp", "Creature — Imp", b639ImpOracle, 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, imp, "first strike") {
		t.Fatal("the Imp starts with first strike")
	}
	if err := b639Activate(t, g, me, imp, 0, "{R}{R}"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !effectstest.HasEffectiveKeyword(t, g, imp, "first strike") {
		t.Error("the Imp did not gain first strike")
	}
}
