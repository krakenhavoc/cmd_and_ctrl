package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b220_test.go — Commander Legends: Battle for Baldur's Gate red creatures,
// GAP batch 220 (#5948). Each test plays the real catalog entry.

const (
	gapB220LivaanOracle     = "ee565d5c-7a3a-481c-91aa-b982cf384220"
	gapB220NalfeshneeOracle = "97a85a3e-258e-49b7-8ebe-7efee808da7a"
	gapB220PhoenixOracle    = "bd0dcb8c-d550-4034-8597-0aba6f6e3b3f"
	gapB220SwashOracle      = "66f0f43c-829d-44d6-adda-d8885bfbfbaf"
	gapB220FanaticsOracle   = "6eef7583-0716-49b5-b70e-8feb0e7c99fb"
	gapB220ThiefOracle      = "ddf34a38-29d5-4bf3-9b08-e7c9579a4200"
	gapB220WyllOracle       = "4fb254cf-b60c-4a3e-b753-4213cb572837"
	gapB220YoungRedOracle   = "2f865ae9-9328-4f6f-924f-d5e0a65aaa96"
	gapB220LozhanOracle     = "be2075af-1cd8-42fe-a29b-a0fd2b047a44"
)

func TestGapB220EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB220LivaanOracle:          "Livaan, Cultist of Tiamat",
		gapB220NalfeshneeOracle:      "Nalfeshnee",
		gapB220PhoenixOracle:         "Nemesis Phoenix",
		gapB220SwashOracle:           "Swashbuckler Extraordinaire",
		gapB220FanaticsOracle:        "Tiamat's Fanatics",
		gapB220ThiefOracle:           "Warehouse Thief",
		gapB220WyllOracle:            "Wyll, Blade of Frontiers",
		gapB220YoungRedOracle:        "Young Red Dragon",
		gapB220YoungRedOracle + "#1": "Bathe in Gold",
		gapB220LozhanOracle:          "Lozhan, Dragons' Legacy",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func gapB220Read(g *game.Game, id uuid.UUID) (c game.Card) {
	g.ReadSnapshot(func() { c, _ = g.LookupCardForEffect(id) })
	return c
}

func gapB220PickTarget(t *testing.T, g *game.Game, chooser uuid.UUID, ref game.TargetRef) {
	t.Helper()
	if pick := effectstest.LatestPickTarget(g, chooser); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, chooser, ref); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
}

// --- Livaan ---

func TestGapB220LivaanPumpsByTheManaValueOfANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Livaan, Cultist of Tiamat", "Legendary Creature — Dragon Shaman", gapB220LivaanOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	spell := game.Card{InstanceID: uuid.New(), Name: "Three Sorcery", TypeLine: "Sorcery", ManaCost: "{2}{R}", Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(spell)
	effectstest.B06AddMana(me, "R", "R", "R")
	if err := g.CastSpell(me.ID, spell.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	gapB220PickTarget(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear})
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, bear); p != 5 {
		t.Errorf("the target has power %d, want 5 (2 plus the spell's mana value 3)", p)
	}
}

func TestGapB220LivaanIgnoresACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Livaan, Cultist of Tiamat", "Legendary Creature — Dragon Shaman", gapB220LivaanOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	cast := game.Card{InstanceID: uuid.New(), Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{2}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(cast)
	effectstest.B06AddMana(me, "R", "R")
	if err := g.CastSpell(me.ID, cast.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, bear); p != 2 {
		t.Errorf("a creature spell pumped the target to %d", p)
	}
}

// --- Nalfeshnee ---

func TestGapB220NalfeshneeCopiesAPermanentSpellCastFromExileAsAHastyTemporaryToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMainForTest(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Nalfeshnee", "Creature — Beast Demon", gapB220NalfeshneeOracle, false)
	id := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: id, Name: "Exiled Bear", TypeLine: "Creature — Bear", ManaCost: "{1}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		if _, err := g.ExileTopWithPermissionForEffect(me.ID, me.ID, 1, game.CastPermission{}); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	effectstest.B06AddMana(me, "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	bears := func() (n int, token uuid.UUID) {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Exiled Bear" {
				n++
				if c.IsToken() {
					token = c.InstanceID
				}
			}
		}
		return
	}
	n, token := bears()
	if n != 2 || token == uuid.Nil {
		t.Fatalf("%d Exiled Bears with token %v, want the card and a token copy", n, token)
	}
	if !effectstest.HasEffectiveKeyword(t, g, token, "haste") {
		t.Error("the copy has no haste")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := bears(); n != 1 {
		t.Errorf("%d Exiled Bears after the end step, want the original only", n)
	}
}

func TestGapB220NalfeshneeIgnoresASpellCastFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Nalfeshnee", "Creature — Beast Demon", gapB220NalfeshneeOracle, false)
	hand := me.Hand
	card := game.Card{InstanceID: uuid.New(), Name: "Hand Bear", TypeLine: "Creature — Bear", ManaCost: "{1}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}
	hand.PushTop(card)
	effectstest.B06AddMana(me, "C")
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Hand Bear" {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d Hand Bears, want 1", n)
	}
}

// --- Nemesis Phoenix ---

func gapB220PhoenixBoard(t *testing.T, attackBoth bool) (g *game.Game, me *game.Player, phoenix uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "Attacker A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Attacker B", "Creature — Bear", 2, 2)
	phoenix = uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: phoenix, Name: "Nemesis Phoenix", TypeLine: "Creature — Phoenix", OracleID: gapB220PhoenixOracle,
			ManaCost: "{3}{R}{R}", Power: 4, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
		t.Fatalf("declare: %v", err)
	}
	second := g.Seats[1].ID
	if attackBoth {
		second = g.Seats[2].ID
	}
	if err := g.DeclareAttacker(b, second); err != nil {
		t.Fatalf("declare: %v", err)
	}
	return g, me, phoenix
}

func TestGapB220NemesisPhoenixReturnsTappedAndAttackingWhenAttackingTwoOpponents(t *testing.T) {
	g, me, phoenix := gapB220PhoenixBoard(t, true)
	effectstest.B06AddMana(me, "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, phoenix, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionPick && c.Chooser == me.ID {
			// "What does it attack?": the first option is an opponent.
			if err := g.ResolveOptionPick(c.ID, me.ID, 0); err != nil {
				t.Fatalf("ResolveOptionPick: %v", err)
			}
			break
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := gapB220Read(g, phoenix)
	if !g.Battlefield.Contains(phoenix) || !c.Tapped || c.AttackingTarget == uuid.Nil {
		t.Errorf("Phoenix on battlefield %v, tapped %v, attacking %v; want tapped and attacking", g.Battlefield.Contains(phoenix), c.Tapped, c.AttackingTarget != uuid.Nil)
	}
}

func TestGapB220NemesisPhoenixNeedsTwoOpponentsBeingAttacked(t *testing.T) {
	g, me, phoenix := gapB220PhoenixBoard(t, false)
	effectstest.B06AddMana(me, "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, phoenix, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the ability was activated while attacking a single opponent")
	}
}

func TestGapB220NemesisPhoenixNeedsTheDeclareAttackersStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	phoenix := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: phoenix, Name: "Nemesis Phoenix", TypeLine: "Creature — Phoenix", OracleID: gapB220PhoenixOracle,
			ManaCost: "{3}{R}{R}", Power: 4, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	effectstest.B06AddMana(me, "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, phoenix, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the ability was activated in the main phase")
	}
}

// --- Swashbuckler Extraordinaire ---

func TestGapB220SwashbucklerMakesATreasureWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Swashbuckler Extraordinaire", TypeLine: "Creature — Dragon Rogue Warrior", OracleID: gapB220SwashOracle, Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Treasure"); got != 1 {
		t.Errorf("%d Treasure tokens, want 1", got)
	}
}

func TestGapB220SwashbucklerSacrificesTreasuresForDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	swash := effectstest.B12Push(g, me.ID, "Swashbuckler Extraordinaire", "Creature — Dragon Rogue Warrior", gapB220SwashOracle, 2, 2)
	treasure := effectstest.PushCatalogPermanent(g, me.ID, "Treasure", "Token Artifact — Treasure", "", false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, swash)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, treasure)
	gapB220PickTarget(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(treasure) {
		t.Fatal("the Treasure was not sacrificed")
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("the chosen creature has no double strike")
	}
}

func TestGapB220SwashbucklerSacrificingNothingDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	swash := effectstest.B12Push(g, me.ID, "Swashbuckler Extraordinaire", "Creature — Dragon Rogue Warrior", gapB220SwashOracle, 2, 2)
	treasure := effectstest.PushCatalogPermanent(g, me.ID, "Treasure", "Token Artifact — Treasure", "", false)
	effectstest.DeclareAttack(t, g, opp.ID, swash)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(treasure) || effectstest.HasEffectiveKeyword(t, g, swash, "double strike") {
		t.Error("something happened without a sacrifice")
	}
}

// --- Tiamat's Fanatics ---

func TestGapB220TiamatsFanaticsHasHasteAndMyriad(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fanatics := effectstest.B12Push(g, me.ID, "Tiamat's Fanatics", "Creature — Dragon Warrior", gapB220FanaticsOracle, 4, 3)
	effectstest.DeclareAttack(t, g, opp.ID, fanatics)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	copies := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Tiamat's Fanatics" && c.InstanceID != fanatics && c.IsToken() {
			copies++
		}
	}
	if copies != 2 {
		t.Errorf("%d myriad token copies, want 2", copies)
	}
	if spec, _ := effects.Lookup(gapB220FanaticsOracle); len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "haste" {
		t.Errorf("keywords %v, want haste", spec.PrintedKeywords)
	}
}

// --- Warehouse Thief ---

func TestGapB220WarehouseThiefSacrificesToExileTheTopCardForTwoTurns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	thief := effectstest.PushCatalogPermanent(g, me.ID, "Warehouse Thief", "Creature — Tiefling Rogue", gapB220ThiefOracle, false)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	top := uuid.New()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: top, Name: "Top Card", TypeLine: "Sorcery", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	})
	effectstest.B06AddMana(me, "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, thief, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) || !g.Exile.Contains(top) {
		t.Fatal("the creature was not sacrificed, or the top card was not exiled")
	}
	if effectstest.GrantedPermissionOn(g, me.ID, top, game.ZoneExile) == nil {
		t.Error("no permission to play the exiled card")
	}
}

// --- Wyll ---

func TestGapB220WyllRollsOneMoreDieAndGrowsOncePerRoll(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wyll := effectstest.B12Push(g, me.ID, "Wyll, Blade of Frontiers", "Legendary Creature — Human Warlock", gapB220WyllOracle, 1, 1)
	from := effectstest.SfkEventCount(g)
	effectstest.D5681Roll(t, g, me.ID, uuid.New(), 20, 1, 0)
	if got := effectstest.D5681Rolled(g, from); len(got) != 1 {
		t.Errorf("one die announced %d rolls, want 1 (two rolled, the lowest ignored)", len(got))
	}
	if got := effectstest.CounterOn(g, wyll, game.CounterPlusOne); got != 1 {
		t.Errorf("Wyll has %d +1/+1 counters after one roll, want 1", got)
	}
	effectstest.D5681Roll(t, g, me.ID, uuid.New(), 20, 2, 0)
	if got := effectstest.CounterOn(g, wyll, game.CounterPlusOne); got != 2 {
		t.Errorf("Wyll has %d +1/+1 counters after rolling two dice at once, want 2 (one per roll)", got)
	}
	from = effectstest.SfkEventCount(g)
	effectstest.D5681Roll(t, g, opp.ID, uuid.New(), 20, 1, 0)
	if got := effectstest.D5681Rolled(g, from); len(got) != 1 {
		t.Errorf("an opponent's die announced %d rolls, want 1 (not replaced)", len(got))
	}
	if got := effectstest.CounterOn(g, wyll, game.CounterPlusOne); got != 2 {
		t.Errorf("an opponent's roll grew Wyll to %d counters", got)
	}
}

// --- Young Red Dragon // Bathe in Gold ---

func TestGapB220YoungRedDragonCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "My Attacker", "Creature — Bear", 2, 2)
	dragon := effectstest.B12Push(g, opp.ID, "Young Red Dragon", "Creature — Dragon", gapB220YoungRedOracle, 3, 2)
	effectstest.BrAttack(t, g, attacker)
	if err := g.DeclareBlocker(dragon, attacker); err == nil {
		t.Error("Young Red Dragon blocked")
	}
}

func TestGapB220BatheInGoldMakesATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Bathe in Gold", "Instant — Adventure", gapB220YoungRedOracle+"#1", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Treasure"); got != 1 {
		t.Errorf("%d Treasure tokens, want 1", got)
	}
}

// --- Lozhan ---

func TestGapB220LozhanDealsDamageEqualToAnAdventureSpellsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Lozhan, Dragons' Legacy", "Legendary Creature — Dragon Shaman", gapB220LozhanOracle, false)
	life := opp.Life
	adventure := game.Card{InstanceID: uuid.New(), Name: "Bathe in Gold", TypeLine: "Instant — Adventure", Layout: game.LayoutAdventure,
		OracleID: gapB220YoungRedOracle + "#1", ManaCost: "{1}{R}", Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(adventure)
	effectstest.B06AddMana(me, "R", "R")
	if err := g.CastSpell(me.ID, adventure.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	gapB220PickTarget(t, g, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the opponent's life is %d, want %d (damage equal to mana value 2)", opp.Life, life-2)
	}
}

func TestGapB220LozhanDealsDamageForADragonSpellButNotAnOtherSpell(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		damage   int
	}{{"Creature — Dragon", 3}, {"Creature — Bear", 0}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		effectstest.PushCatalogPermanent(g, me.ID, "Lozhan, Dragons' Legacy", "Legendary Creature — Dragon Shaman", gapB220LozhanOracle, false)
		life := opp.Life
		card := game.Card{InstanceID: uuid.New(), Name: "Test Spell", TypeLine: tc.typeLine, ManaCost: "{2}{R}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}
		me.Hand.PushTop(card)
		effectstest.B06AddMana(me, "R", "R", "R")
		if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		gapB220PickTarget(t, g, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
		effectstest.PassPriorityAroundTable(t, g)
		if got := life - opp.Life; got != tc.damage {
			t.Errorf("%s: %d damage, want %d", tc.typeLine, got, tc.damage)
		}
	}
}

func TestGapB220LozhanCannotTargetACommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Lozhan, Dragons' Legacy", "Legendary Creature — Dragon Shaman", gapB220LozhanOracle, false)
	cmdr := effectstest.B12Creature(g, opp.ID, "Their Commander", "Legendary Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, opp.ID, "Plain Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == cmdr {
				g.Battlefield.Cards[i].IsCommander = true
			}
		}
	})
	spec, _ := effects.Lookup(gapB220LozhanOracle)
	var clause *game.TargetSpec
	for _, tr := range spec.Triggered {
		if tr.Targets != nil {
			clause = tr.Targets
		}
	}
	legal := effectstest.TaLegalForSpec(g, me.ID, clause)
	if legal[cmdr] {
		t.Error("a commander was a legal target")
	}
	if !legal[plain] {
		t.Error("a creature that is not a commander was refused")
	}
}
