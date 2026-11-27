package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 630 (Apocalypse: pump and combat tricks): Minotaur Tactician,
// Glade Gnarr, Angelfire Crusader, Planar Despair, Ceta Disciple, Kavu Glider,
// Ana Sanctuary, Bog Gnarr, Kavu Mauler, Dega Disciple, Ana Disciple.

const (
	oldB630Tactician = "197fd850-0a08-4e82-9249-a8d719436f77"
	oldB630Glade     = "1a42fd1d-a2c3-40c4-8c34-7a3559aa9caf"
	oldB630Crusader  = "1a587b8d-fcb4-47ca-9689-610897117166"
	oldB630Despair   = "1ba36042-8902-4abb-a4ae-a8d26d81a0de"
	oldB630Ceta      = "1f085c2d-bd1b-4889-882f-367e0e970f44"
	oldB630Glider    = "425050a7-41d2-4850-bb45-3a83275d3c50"
	oldB630AnaSanct  = "51f17fe1-1cf7-4362-b9a2-8ec225d41b03"
	oldB630Bog       = "537ebea5-f805-4f78-9151-58eb43c4a62d"
	oldB630Mauler    = "6ffbcaba-5437-4fb6-a2d6-e94b1e6dc1d2"
	oldB630Dega      = "c6c53277-daad-4aa3-b29a-11378d550651"
	oldB630AnaDisc   = "f543dfcd-015e-48bc-851d-08002d0241fa"
)

func TestOldB630EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB630Tactician: "Minotaur Tactician", oldB630Glade: "Glade Gnarr", oldB630Crusader: "Angelfire Crusader",
		oldB630Despair: "Planar Despair", oldB630Ceta: "Ceta Disciple", oldB630Glider: "Kavu Glider",
		oldB630AnaSanct: "Ana Sanctuary", oldB630Bog: "Bog Gnarr", oldB630Mauler: "Kavu Mauler",
		oldB630Dega: "Dega Disciple", oldB630AnaDisc: "Ana Disciple",
	} {
		effectstest.CmdrB21AssertFull(t, oracle, name)
	}
}

func oldB630Colored(g *game.Game, owner uuid.UUID, name, typeLine string, p, tough int, colors ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: p, Toughness: tough, Colors: colors, Owner: owner, Controller: owner})
}

func oldB630Activate(t *testing.T, g *game.Game, who *game.Player, id uuid.UUID, idx int, mana string, targets ...uuid.UUID) error {
	t.Helper()
	g.WithWriteLock(func() { who.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, who, mana)
	params := game.ActivateAbilityParams{Strict: true}
	if len(targets) > 0 {
		params.Targets = effectstest.CardTargets(targets...)
	}
	err := g.ActivateCatalogAbility(who.ID, id, idx, params)
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return err
}

func oldB630Size(t *testing.T, g *game.Game, id uuid.UUID, p, tough int) {
	t.Helper()
	if gp, gt := effectstest.PtOf(t, g, id); gp != p || gt != tough {
		t.Errorf("creature is %d/%d, want %d/%d", gp, gt, p, tough)
	}
}

// --- Minotaur Tactician ----------------------------------------------------

func TestOldB630TacticianGrowsForEachOfAWhiteAndABlueCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tac := effectstest.B12Push(g, me.ID, "Minotaur Tactician", "Creature — Minotaur", oldB630Tactician, 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, tac, "haste") {
		t.Error("no haste")
	}
	oldB630Size(t, g, tac, 1, 1)
	oldB630Colored(g, opp.ID, "Their White", "Creature — Soldier", 1, 1, "W")
	oldB630Size(t, g, tac, 1, 1) // an opponent's white creature does not count
	white := oldB630Colored(g, me.ID, "My White", "Creature — Soldier", 1, 1, "W")
	oldB630Size(t, g, tac, 2, 2)
	oldB630Colored(g, me.ID, "My Second White", "Creature — Soldier", 1, 1, "W")
	oldB630Size(t, g, tac, 2, 2) // one +1/+1 per printed static, not per creature
	oldB630Colored(g, me.ID, "My Blue", "Creature — Merfolk", 1, 1, "U")
	oldB630Size(t, g, tac, 3, 3)
	effectstest.KillCreature(t, g, me.ID, white)
	oldB630Size(t, g, tac, 3, 3) // the second white creature still counts
}

// --- Glade Gnarr / Bog Gnarr -------------------------------------------------

func oldB630Cast(t *testing.T, g *game.Game, p *game.Player, name string, colors ...string) {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Sorcery", Colors: colors, Owner: p.ID, Controller: p.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB630GnarrsGrowOnlyForTheirOwnColourOfSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	glade := effectstest.B12Push(g, me.ID, "Glade Gnarr", "Creature — Beast", oldB630Glade, 4, 4)
	bog := effectstest.B12Push(g, me.ID, "Bog Gnarr", "Creature — Beast", oldB630Bog, 2, 2)
	oldB630Cast(t, g, me, "Red Thing", "R")
	oldB630Size(t, g, glade, 4, 4)
	oldB630Size(t, g, bog, 2, 2)
	oldB630Cast(t, g, me, "Blue Thing", "U")
	oldB630Size(t, g, glade, 6, 6)
	oldB630Size(t, g, bog, 2, 2)
	oldB630Cast(t, g, me, "Black Thing", "B")
	oldB630Size(t, g, glade, 6, 6)
	oldB630Size(t, g, bog, 4, 4)
	oldB630Cast(t, g, me, "Another Blue Thing", "U")
	oldB630Size(t, g, glade, 8, 8) // each blue spell is a trigger of its own
}

func TestOldB630GnarrBonusEndsWithTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	glade := effectstest.B12Push(g, me.ID, "Glade Gnarr", "Creature — Beast", oldB630Glade, 4, 4)
	oldB630Cast(t, g, me, "Blue Thing", "U")
	oldB630Size(t, g, glade, 6, 6)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	oldB630Size(t, g, glade, 4, 4)
}

// --- Angelfire Crusader / Kavu Glider ------------------------------------------

func TestOldB630CrusaderPumpsForRedManaAsOftenAsYouLike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	c := effectstest.B12Push(g, me.ID, "Angelfire Crusader", "Creature — Human Soldier Knight", oldB630Crusader, 2, 3)
	effectstest.AdvanceToMain(t, g)
	if err := oldB630Activate(t, g, me, c, 0, "W"); err == nil {
		t.Error("activated with white mana")
	}
	for i := 0; i < 3; i++ {
		if err := oldB630Activate(t, g, me, c, 0, "R"); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
	}
	oldB630Size(t, g, c, 5, 3)
}

func TestOldB630GliderToughensForWhiteAndFliesForBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	k := effectstest.B12Push(g, me.ID, "Kavu Glider", "Creature — Kavu", oldB630Glider, 2, 1)
	effectstest.AdvanceToMain(t, g)
	if effectstest.HasEffectiveKeyword(t, g, k, "flying") {
		t.Fatal("flies before the ability")
	}
	if err := oldB630Activate(t, g, me, k, 0, "W"); err != nil {
		t.Fatalf("white: %v", err)
	}
	if err := oldB630Activate(t, g, me, k, 0, "W"); err != nil {
		t.Fatalf("white again: %v", err)
	}
	oldB630Size(t, g, k, 2, 3)
	if err := oldB630Activate(t, g, me, k, 1, "U"); err != nil {
		t.Fatalf("blue: %v", err)
	}
	if !effectstest.HasEffectiveKeyword(t, g, k, "flying") {
		t.Error("no flying after {U}")
	}
	if err := oldB630Activate(t, g, me, k, 1, "R"); err == nil {
		t.Error("the flying ability took red mana")
	}
}

// --- Planar Despair --------------------------------------------------------------

func TestOldB630PlanarDespairShrinksEveryCreatureByTheDomainCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for _, land := range []string{"Swamp", "Swamp", "Forest", "Island"} {
		effectstest.PushCatalogPermanent(g, me.ID, land, "Basic Land — "+land, "", false)
	}
	effectstest.PushCatalogPermanent(g, opp.ID, "Mountain", "Basic Land — Mountain", "", false) // not yours, not counted
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 5, 5)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 3)
	effectstest.CastCatalogSpell(t, g, "Planar Despair", "Sorcery", oldB630Despair, nil)
	effectstest.PassPriorityAroundTable(t, g)
	oldB630Size(t, g, big, 2, 2)
	if g.Battlefield.Contains(small) {
		t.Error("the 2/3 survived -3/-3")
	}
}

func TestOldB630PlanarDespairDoesNothingWithoutABasicLandType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Wastes", "Land", "", false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Planar Despair", "Sorcery", oldB630Despair, nil)
	effectstest.PassPriorityAroundTable(t, g)
	oldB630Size(t, g, bear, 2, 2)
}

// --- Disciples --------------------------------------------------------------------

func TestOldB630CetaDiscipleBoostsATargetAndFiltersManaIntoAnyColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	d := effectstest.PushCatalogPermanent(g, me.ID, "Ceta Disciple", "Creature — Merfolk Wizard", oldB630Ceta, false)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := oldB630Activate(t, g, me, d, 0, "G", victim); err == nil {
		t.Error("activated the pump with green mana")
	}
	if err := oldB630Activate(t, g, me, d, 0, "R", victim); err != nil {
		t.Fatalf("activate: %v", err)
	}
	oldB630Size(t, g, victim, 4, 2)
	if !effectstest.IsTapped(g, d) {
		t.Error("the Disciple did not tap")
	}
	d2 := effectstest.PushCatalogPermanent(g, me.ID, "Ceta Disciple", "Creature — Merfolk Wizard", oldB630Ceta, false)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "G")
	effectstest.ActivateManaFor(t, g, me.ID, d2, 0, game.ManaAbilityParams{Colors: []string{"U"}})
	if got := me.ManaPool.CountColor("U"); got != 1 || me.ManaPool.CountColor("G") != 0 {
		t.Errorf("pool has %d blue and %d green, want {G} turned into one {U}", got, me.ManaPool.CountColor("G"))
	}
}

func TestOldB630DiscipleCannotActivateWhileSummoningSick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	d := effectstest.PushCatalogPermanent(g, me.ID, "Ceta Disciple", "Creature — Merfolk Wizard", oldB630Ceta, true)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := oldB630Activate(t, g, me, d, 0, "R", victim); err == nil {
		t.Error("a summoning-sick Disciple paid {T}")
	}
}

func TestOldB630DegaDiscipleShrinksOrPumpsAPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	d := effectstest.PushCatalogPermanent(g, me.ID, "Dega Disciple", "Creature — Human Wizard", oldB630Dega, false)
	d2 := effectstest.PushCatalogPermanent(g, me.ID, "Dega Disciple", "Creature — Human Wizard", oldB630Dega, false)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 3)
	effectstest.AdvanceToMain(t, g)
	if err := oldB630Activate(t, g, me, d, 0, "B", victim); err != nil {
		t.Fatalf("shrink: %v", err)
	}
	oldB630Size(t, g, victim, 1, 3)
	if err := oldB630Activate(t, g, me, d2, 1, "R", victim); err != nil {
		t.Fatalf("pump: %v", err)
	}
	oldB630Size(t, g, victim, 3, 3)
	if err := oldB630Activate(t, g, me, d, 1, "R", victim); err == nil {
		t.Error("a tapped Disciple activated again")
	}
}

func TestOldB630AnaDiscipleGrantsFlyingOrShrinks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	d := effectstest.PushCatalogPermanent(g, me.ID, "Ana Disciple", "Creature — Human Wizard", oldB630AnaDisc, false)
	d2 := effectstest.PushCatalogPermanent(g, me.ID, "Ana Disciple", "Creature — Human Wizard", oldB630AnaDisc, false)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 3)
	effectstest.AdvanceToMain(t, g)
	if err := oldB630Activate(t, g, me, d, 0, "U", victim); err != nil {
		t.Fatalf("flying: %v", err)
	}
	if !effectstest.HasEffectiveKeyword(t, g, victim, "flying") {
		t.Error("the target has no flying")
	}
	if err := oldB630Activate(t, g, me, d2, 1, "B", victim); err != nil {
		t.Fatalf("shrink: %v", err)
	}
	oldB630Size(t, g, victim, 1, 3)
}

// --- Kavu Mauler -----------------------------------------------------------------

func TestOldB630MaulerGrowsForEachOtherAttackingKavuOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	m := effectstest.B12Push(g, me.ID, "Kavu Mauler", "Creature — Kavu", oldB630Mauler, 4, 4)
	k1 := effectstest.B12Creature(g, me.ID, "Kavu One", "Creature — Kavu", 1, 1)
	k2 := effectstest.B12Creature(g, me.ID, "Kavu Two", "Creature — Kavu", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	idle := effectstest.B12Creature(g, me.ID, "Idle Kavu", "Creature — Kavu", 1, 1)
	_ = idle
	if !effectstest.HasEffectiveKeyword(t, g, m, "trample") {
		t.Error("no trample")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, m, k1, k2, bear)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	oldB630Size(t, g, m, 6, 6)
}

func TestOldB630MaulerAttackingAloneStaysTheSame(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	m := effectstest.B12Push(g, me.ID, "Kavu Mauler", "Creature — Kavu", oldB630Mauler, 4, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, m)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	oldB630Size(t, g, m, 4, 4)
}

// --- Ana Sanctuary ---------------------------------------------------------------

func oldB630Sanctuary(t *testing.T, colors ...[]string) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Ana Sanctuary", "Enchantment", oldB630AnaSanct, false)
	for i, c := range colors {
		oldB630Colored(g, me.ID, "Idol "+string(rune('A'+i)), "Enchantment", 0, 0, c...)
	}
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	return g, me, victim
}

func TestOldB630AnaSanctuaryNeedsABlueOrBlackPermanent(t *testing.T) {
	g, me, victim := oldB630Sanctuary(t, []string{"R"}, nil)
	effectstest.BeginUpkeep(t, g, me.ID)
	oldB630Size(t, g, victim, 2, 2)
}

func TestOldB630AnaSanctuaryPumpsOneOrFiveByColours(t *testing.T) {
	for _, tc := range []struct {
		name   string
		colors [][]string
		want   int
	}{
		{"blue only", [][]string{{"U"}}, 3},
		{"black only", [][]string{{"B"}}, 3},
		{"blue and black", [][]string{{"U"}, {"B"}}, 7},
		{"one blue black permanent", [][]string{{"U", "B"}}, 7},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, victim := oldB630Sanctuary(t, tc.colors...)
			effectstest.BeginUpkeep(t, g, me.ID)
			effectstest.AnswerTriggerTargets(t, g, me.ID, victim)
			effectstest.PassPriorityAroundTable(t, g)
			oldB630Size(t, g, victim, tc.want, tc.want)
		})
	}
}
