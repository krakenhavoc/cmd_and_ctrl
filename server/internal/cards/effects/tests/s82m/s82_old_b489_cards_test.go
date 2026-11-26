package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 489 (Worldwake, #5582): Vapor Snare, the three Zendikons,
// Perimeter Captain, Scrib Nibblers, Rumbling Aftershocks, Slavering Nulls,
// Voyager Drake, Tideforce Elemental, Jwari Shapeshifter and Dead Reckoning.

const (
	oldB489Snare      = "29ae5540-5fdf-48be-86ba-b4db0fe19f6c"
	oldB489Corrupted  = "3902048e-b7bf-4676-be92-e590bc76c729"
	oldB489Crusher    = "68a1a67e-9d40-47dd-b282-6c1b15465f62"
	oldB489Wind       = "d6786f80-818c-433d-8978-81814171dd39"
	oldB489Captain    = "05608055-d97a-4c8f-833d-47b3dd1ea255"
	oldB489Nibblers   = "458800d3-47b7-4f28-8371-25c8d1e60011"
	oldB489Aftershock = "851a2214-5f87-4e5b-b099-1c5c90784a97"
	oldB489Nulls      = "c6d39445-dfc2-42c3-afd2-016fdf008f0e"
	oldB489Drake      = "b2e68a8c-cf07-4c2c-8e23-7602a8ccf831"
	oldB489Tideforce  = "bda87ca2-91d0-48fa-bb82-a3396f0438ac"
	oldB489Jwari      = "0b0c02eb-4d9e-4fea-bad6-9ef5d7bb5295"
	oldB489Reckoning  = "2a4b77f0-81ac-48a6-961b-bd63e99f74cf"
)

func TestOldB489EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB489Snare: "Vapor Snare", oldB489Corrupted: "Corrupted Zendikon", oldB489Crusher: "Crusher Zendikon",
		oldB489Wind: "Wind Zendikon", oldB489Captain: "Perimeter Captain", oldB489Nibblers: "Scrib Nibblers",
		oldB489Aftershock: "Rumbling Aftershocks", oldB489Nulls: "Slavering Nulls", oldB489Drake: "Voyager Drake",
		oldB489Tideforce: "Tideforce Elemental", oldB489Jwari: "Jwari Shapeshifter", oldB489Reckoning: "Dead Reckoning",
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

func oldB489Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	id := effectstest.CastCatalogSpell(t, g, name, typeLine, oracle, targets)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestOldB489ZendikonsAnimateTheLandAndReturnItWhenItDies(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, subtype, color, keyword string
		power, toughness                      int
	}{
		{"Corrupted Zendikon", oldB489Corrupted, "Ooze", "B", "", 3, 3},
		{"Crusher Zendikon", oldB489Crusher, "Beast", "R", "trample", 4, 2},
		{"Wind Zendikon", oldB489Wind, "Elemental", "U", "flying", 2, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			land := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
			oldB489Cast(t, g, tc.name, "Enchantment — Aura", tc.oracle, effectstest.CardTarget(land))
			c := effectstest.B12Card(t, g, land)
			if !c.IsLand() || !c.IsCreature() || !c.HasSubtype(tc.subtype) {
				t.Fatalf("the enchanted land is %q, want a land %s creature", c.TypeLine, tc.subtype)
			}
			if p, tt := effectstest.PtOf(t, g, land); p != tc.power || tt != tc.toughness {
				t.Errorf("the land is %d/%d, want %d/%d", p, tt, tc.power, tc.toughness)
			}
			if got := effectstest.EffectiveColors(t, g, land); len(got) != 1 || got[0] != tc.color {
				t.Errorf("colours %v, want %s", got, tc.color)
			}
			if tc.keyword != "" && !effectstest.HasEffectiveKeyword(t, g, land, tc.keyword) {
				t.Errorf("the land lacks %s", tc.keyword)
			}
			if tc.keyword == "" && (effectstest.HasEffectiveKeyword(t, g, land, "flying") || effectstest.HasEffectiveKeyword(t, g, land, "trample")) {
				t.Error("the land has a keyword the card does not give")
			}
			effectstest.KillCreature(t, g, me.ID, land)
			effectstest.PassPriorityAroundTable(t, g)
			if !effectstest.HandHasID(me, land) {
				t.Error("the land was not returned to its owner's hand")
			}
		})
	}
}

func TestOldB489ZendikonRefusesAnythingButALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Wind Zendikon", "Enchantment — Aura", oldB489Wind, effectstest.CardTarget(bear)); err == nil {
		t.Error("Wind Zendikon enchanted a creature that is not a land")
	}
}

func TestOldB489VaporSnareStealsAndCostsALandEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.SeedBear(g, opp.ID)
	land := effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	snare := oldB489Cast(t, g, "Vapor Snare", "Enchantment — Aura", oldB489Snare, effectstest.CardTarget(theirs))
	if got := effectstest.ControllerOf(t, g, theirs); got != me.ID {
		t.Fatalf("controller %s, want the Aura's controller", got)
	}
	// Next upkeep of mine: pay with the land and the Aura stays.
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityUntilChoice(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for returning a land")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, land) || g.Battlefield.Contains(land) {
		t.Error("the land was not returned to hand")
	}
	if !g.Battlefield.Contains(snare) {
		t.Fatal("Vapor Snare was sacrificed although a land was returned")
	}
	// The following upkeep: no land to return, so the Aura is sacrificed.
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(snare) {
		t.Error("Vapor Snare stayed with no land to return")
	}
	if got := effectstest.ControllerOf(t, g, theirs); got != opp.ID {
		t.Errorf("control did not go home when the Aura left: %s", got)
	}
}

func TestOldB489VaporSnareIsSacrificedWhenThePaymentIsDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.SeedBear(g, opp.ID)
	land := effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	snare := oldB489Cast(t, g, "Vapor Snare", "Enchantment — Aura", oldB489Snare, effectstest.CardTarget(theirs))
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityUntilChoice(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for returning a land")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(snare) {
		t.Error("Vapor Snare stayed although the payment was declined")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land left although the payment was declined")
	}
}

func TestOldB489PerimeterCaptainGainsLifeWhenADefenderBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.HmlCreature(g, me.ID, "Attacker", "", 1, 1)
	captain := effectstest.HmlCreature(g, opp.ID, "Perimeter Captain", oldB489Captain, 0, 4)
	nonDefender := effectstest.HmlCreature(g, opp.ID, "Bear", "", 2, 2)
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(captain, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, opp.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life+2 {
		t.Errorf("life %d → %d, want +2 when a creature with defender blocks", life, opp.Life)
	}
	_ = nonDefender
}

func TestOldB489PerimeterCaptainIgnoresABlockerWithoutDefender(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.HmlCreature(g, me.ID, "Attacker", "", 1, 1)
	effectstest.HmlCreature(g, opp.ID, "Perimeter Captain", oldB489Captain, 0, 4)
	bear := effectstest.HmlCreature(g, opp.ID, "Bear", "", 2, 2)
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(bear, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("life %d → %d, want no gain from a blocker without defender", life, opp.Life)
	}
}

// oldB489OnTop puts a card on top of p's library.
func oldB489OnTop(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

func TestOldB489ScribNibblersExilesTheTopCardAndGainsLifeForALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nib := effectstest.PushCatalogPermanent(g, me.ID, "Scrib Nibblers", "Creature — Rat", oldB489Nibblers, false)
	land := oldB489OnTop(opp, "Swamp", "Basic Land — Swamp")
	life := me.Life
	effectstest.B16Activate(t, g, me.ID, nib, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, land) {
		t.Error("the land card was not exiled")
	}
	if me.Life != life+1 {
		t.Errorf("life %d → %d, want +1 for exiling a land card", life, me.Life)
	}
	if !effectstest.TappedForTest(t, g, nib) {
		t.Error("the {T} cost did not tap Scrib Nibblers")
	}
	// A nonland top card exiles with no life.
	effectstest.UntapForTest(g, nib)
	spell := oldB489OnTop(opp, "Bear", "Creature — Bear")
	effectstest.B16Activate(t, g, me.ID, nib, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, spell) || me.Life != life+1 {
		t.Errorf("nonland: exiled %v, life %d (want exiled, %d)", effectstest.InExile(g, spell), me.Life, life+1)
	}
}

func TestOldB489LandfallMayUntapScribNibblersAndTideforceElemental(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Scrib Nibblers", "Creature — Rat", oldB489Nibblers},
		{"Tideforce Elemental", "Creature — Elemental", oldB489Tideforce},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			effectstest.TapCard(g, id)
			effectstest.AdvanceToMain(t, g)
			effectstest.PlayLandFromHand(t, g, "Forest", "")
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.TappedForTest(t, g, id) {
				t.Error("landfall did not untap the creature")
			}
		})
	}
}

func TestOldB489SlaveringNullsMakesThePlayerDiscardOnlyWithASwamp(t *testing.T) {
	for _, swamp := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		nulls := effectstest.HmlCreature(g, me.ID, "Slavering Nulls", oldB489Nulls, 2, 1)
		if swamp {
			effectstest.PushLandFor(g, me.ID, "Swamp", "Basic Land — Swamp")
		}
		effectstest.PushCatalogHandCard(opp, "Card A", "Sorcery", "")
		effectstest.DealCombatDamageToPlayer(g, nulls, opp.ID, 2)
		if swamp {
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		}
		effectstest.PassPriorityAroundTable(t, g)
		asked := effectstest.LatestChooseCardsFor(g, opp.ID) != nil
		if asked != swamp {
			t.Errorf("swamp=%v: the damaged player was asked to discard = %v", swamp, asked)
		}
	}
}

func TestOldB489TideforceElementalTapsOrUntapsAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ele := effectstest.PushCatalogPermanent(g, me.ID, "Tideforce Elemental", "Creature — Elemental", oldB489Tideforce, false)
	victim := effectstest.HmlCreature(g, opp.ID, "Victim", "", 2, 2)
	effectstest.B10AddMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, ele, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(victim)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedForTest(t, g, victim) {
		t.Error("the target was not tapped")
	}
	if !effectstest.TappedForTest(t, g, ele) {
		t.Error("the {T} cost did not tap the Elemental")
	}
	// It cannot target itself.
	effectstest.UntapForTest(g, ele)
	effectstest.B10AddMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, ele, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(ele)}); err == nil {
		t.Error("Tideforce Elemental targeted itself")
	}
}

func oldB489DrakeCard() game.Card {
	return game.Card{Name: "Voyager Drake", TypeLine: "Creature — Drake", OracleID: oldB489Drake, ManaCost: "{3}{U}", Power: 3, Toughness: 3}
}

func TestOldB489VoyagerDrakeGivesFlyingToUpToXCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
	if _, err := effectstest.Msc17Cast(t, g, me, oldB489DrakeCard(), game.CastSpellParams{OptionalCosts: []int{0, 0}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatalf("no target prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c}}); err == nil {
		t.Fatal("'up to X' with X=2 accepted three targets")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}); err != nil {
		t.Fatalf("pick two: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, a, "flying") || !effectstest.HasEffectiveKeyword(t, g, b, "flying") {
		t.Error("a chosen creature did not gain flying")
	}
	if effectstest.HasEffectiveKeyword(t, g, c, "flying") {
		t.Error("an unchosen creature gained flying")
	}
}

func TestOldB489VoyagerDrakeUnkickedAsksNoTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	if _, err := effectstest.Msc17Cast(t, g, me, oldB489DrakeCard(), game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Error("an unkicked Voyager Drake asked for a target")
	}
}

func TestOldB489RumblingAftershocksDamagesByTheTimesAKickedSpellWasKicked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Rumbling Aftershocks", "Enchantment", oldB489Aftershock, false)
	life := opp.Life
	if _, err := effectstest.Msc17Cast(t, g, me, oldB489DrakeCard(), game.CastSpellParams{OptionalCosts: []int{0, 0, 0}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatalf("no target prompt for the trigger: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("the opponent lost %d life, want 3 (kicked three times)", got)
	}
}

func TestOldB489RumblingAftershocksIgnoresAnUnkickedSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Rumbling Aftershocks", "Enchantment", oldB489Aftershock, false)
	life := opp.Life
	if _, err := effectstest.Msc17Cast(t, g, me, oldB489DrakeCard(), game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil || opp.Life != life {
		t.Errorf("an unkicked spell triggered Rumbling Aftershocks (prompt %v, life %d → %d)", p != nil, life, opp.Life)
	}
}

func TestOldB489JwariShapeshifterEntersAsACopyOfAnAlly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ally := effectstest.SeedCopyableCreature(g, me.ID, "Kor Cartographer", "Creature — Kor Scout Ally", 2, 2)
	effectstest.SeedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	id := effectstest.CastCatalogSpell(t, g, "Jwari Shapeshifter", "Creature — Shapeshifter Ally", oldB489Jwari, nil)
	prompt := 0
	for i := 0; i < 32 && g.Battlefield.Contains(id) == false; i++ {
		if c := effectstest.CopyPrompt(g); c != nil {
			prompt++
			if err := g.ResolveCopyTarget(c.ID, c.Chooser, ally); err != nil {
				t.Fatalf("ResolveCopyTarget: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if prompt == 0 {
		t.Fatal("no copy prompt")
	}
	if got := effectstest.B12Card(t, g, id); got.Effective().Name != "Kor Cartographer" {
		t.Errorf("entered as %q, want a copy of Kor Cartographer", got.Effective().Name)
	}
}

func TestOldB489DeadReckoningTucksTheCardAndDealsItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Fallen Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 3)
	oldB489Cast(t, g, "Dead Reckoning", "Sorcery", oldB489Reckoning, []game.TargetRef{{Kind: game.TargetCard, ID: dead}, {Kind: game.TargetCard, ID: victim}})
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(me.Library.Cards); n == 0 || me.Library.Cards[n-1].InstanceID != dead {
		t.Error("the creature card is not on top of the library")
	}
	if g.Battlefield.Contains(victim) {
		t.Error("3 damage (the card's power) did not kill the 3/3")
	}
}

func TestOldB489DeadReckoningDealsNothingWhenTheTuckIsDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Fallen Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 3)
	oldB489Cast(t, g, "Dead Reckoning", "Sorcery", oldB489Reckoning, []game.TargetRef{{Kind: game.TargetCard, ID: dead}, {Kind: game.TargetCard, ID: victim}})
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Error("damage was dealt although the card was not put on the library")
	}
	if _, ok := effectstest.GraveyardCard(me, dead); !ok {
		t.Error("the card left the graveyard although the tuck was declined")
	}
}
