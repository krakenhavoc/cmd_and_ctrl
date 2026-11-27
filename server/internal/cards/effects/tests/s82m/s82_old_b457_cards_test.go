package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b457_cards_test.go — Commander 2011, pump and combat tricks,
// instants and sorceries, other creatures and card draw, S82 older sets, batch
// 457 (#5574). Each test plays the real catalog entry.

const (
	oldB457FallenAngel = "1869a326-8b48-4331-b969-5ce83cf4ff8a"
	oldB457Avatar      = "4854f836-2247-43c3-adbc-1d1ee5eaf120"
	oldB457Murmurs     = "1e8be8b4-f7be-49cb-b32e-d92015dea7a2"
	oldB457Ray         = "adf41d29-403d-4936-9689-3148d103e700"
	oldB457Cadets      = "32dc847a-d8e9-4bc5-aa73-dfe69ff330a2"
	oldB457Oni         = "ed8bc0f2-5a65-4eb4-bf88-3642a55b6889"
	oldB457Research    = "9243b539-5f9e-4bb3-9973-3d48366aa7cc"
	oldB457Gorger      = "df23a4a5-eddd-4b0e-9f7e-a3733ce49c0d"
	oldB457Riddle      = "a5e37c6e-ae42-4c98-ba20-840ea3c4a6ab"
	oldB457Dominus     = "f5e18b27-c5cc-4f5e-b767-4b1b43e00fd0"
	oldB457Gomazoa     = "0c82ae16-f995-4ac3-adb4-a5cc93a6e6f2"
	oldB457Bond        = "7de8ffa5-f6e5-4987-b3f8-6d6fc83f8fa8"
	oldB457Force       = "8add6996-6c5c-4aa2-93d4-a39efcfd22b2"
	oldB457Bear        = "Creature — Bear"
)

func TestOldB457EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB457FallenAngel: "Fallen Angel",
		oldB457Avatar:      "Avatar of Fury",
		oldB457Murmurs:     "Murmurs from Beyond",
		oldB457Ray:         "Ray of Command",
		oldB457Cadets:      "Goblin Cadets",
		oldB457Oni:         "Razorjaw Oni",
		oldB457Research:    "Perilous Research",
		oldB457Gorger:      "Trench Gorger",
		oldB457Riddle:      "Riddlekeeper",
		oldB457Dominus:     "Dominus of Fealty",
		oldB457Gomazoa:     "Gomazoa",
		oldB457Bond:        "Martyr's Bond",
		oldB457Force:       "Celestial Force",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB457Land(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: owner, Controller: owner,
	})
}

func oldB457Controller(g *game.Game, id uuid.UUID) uuid.UUID {
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		return uuid.Nil
	}
	return c.Controller
}

// --- Fallen Angel -----------------------------------------------------------------

func TestOldB457FallenAngelSacrificesACreatureForPlusTwoPlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	angel := effectstest.B12Push(g, me.ID, "Fallen Angel", "Creature — Angel", oldB457FallenAngel, 3, 3)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", oldB457Bear, 1, 1)
	effectstest.P7Activate(t, g, me, angel, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	if _, ok := effectstest.BattlefieldCard(g, fodder); ok {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	c := effectstest.LayeredCard(t, g, angel)
	if c.CurrentPower() != 5 || c.CurrentToughness() != 4 {
		t.Errorf("the Angel is %d/%d, want 5/4", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestOldB457FallenAngelCannotBeActivatedWithoutSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	angel := effectstest.B12Push(g, me.ID, "Fallen Angel", "Creature — Angel", oldB457FallenAngel, 3, 3)
	if err := g.ActivateCatalogAbility(me.ID, angel, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the ability was activated with no creature named to sacrifice")
	}
}

// --- Avatar of Fury ---------------------------------------------------------------

func TestOldB457AvatarOfFuryCostsSixLessWhenAnOpponentHasSevenLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 7; i++ {
		oldB457Land(g, opp.ID)
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{R}{R}")
	effectstest.P1g2Cast(t, g, "Avatar of Fury", "Creature — Avatar", oldB457Avatar, "{6}{R}{R}", game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Avatar of Fury"); got != 1 {
		t.Errorf("%d Avatars on the battlefield, want 1 after paying {R}{R}", got)
	}
}

func TestOldB457AvatarOfFuryCostsFullWithSixLandsOrSplitAcrossOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	// Four lands each, eight in all, but never seven for one opponent.
	for _, p := range g.Seats[1:3] {
		for i := 0; i < 4; i++ {
			oldB457Land(g, p.ID)
		}
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{R}{R}")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Avatar of Fury", TypeLine: "Creature — Avatar", OracleID: oldB457Avatar,
		ManaCost: "{6}{R}{R}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the Avatar was cast for {R}{R} with no single opponent at seven lands")
	}
}

func TestOldB457AvatarOfFuryFirebreathes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	avatar := effectstest.B12Push(g, me.ID, "Avatar of Fury", "Creature — Avatar", oldB457Avatar, 6, 6)
	effectstest.FloatMana(t, g, me, "{R}{R}")
	effectstest.P7Activate(t, g, me, avatar, 0, game.ActivateAbilityParams{})
	effectstest.P7Activate(t, g, me, avatar, 0, game.ActivateAbilityParams{})
	if c := effectstest.LayeredCard(t, g, avatar); c.CurrentPower() != 8 {
		t.Errorf("power %d, want 8 after two activations", c.CurrentPower())
	}
}

// --- Murmurs from Beyond ------------------------------------------------------------

func TestOldB457MurmursPutsTheOpponentsChoiceInTheGraveyardAndTheRestInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.StockLibraryFor(g, me, 10)
	a := oldB456TopCardB457(me, "Top A")
	b := oldB456TopCardB457(me, "Top B")
	c := oldB456TopCardB457(me, "Top C")
	hand := len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Murmurs from Beyond", "Instant", oldB457Murmurs, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// The caster picks which opponent chooses, when more than one could.
	for _, ch := range g.PendingChoices {
		if ch != nil && ch.Kind == game.PendingChoiceOptionPick && ch.Chooser == me.ID {
			for i, o := range ch.PickOptions {
				if o.Player == opp.ID {
					if err := g.ResolveOptionPick(ch.ID, me.ID, i); err != nil {
						t.Fatalf("choose the opponent: %v", err)
					}
				}
			}
			break
		}
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, b)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(b) {
		t.Error("the chosen card did not go to the graveyard")
	}
	if !me.Hand.Contains(a) || !me.Hand.Contains(c) {
		t.Error("the other two did not go to the hand")
	}
	if got := len(me.Hand.Cards) - hand; got != 2 {
		t.Errorf("hand grew by %d, want 2", got)
	}
}

func oldB456TopCardB457(p *game.Player, name string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = "Sorcery"
	p.Library.PushTop(c)
	return c.InstanceID
}

// --- Ray of Command -------------------------------------------------------------------

func TestOldB457RayOfCommandStealsUntapsAndTapsItWhenControlEnds(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB457Bear, 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bear).Tapped = true })
	effectstest.CastCatalogSpell(t, g, "Ray of Command", "Instant", oldB457Ray, effectstest.CardTargets(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if oldB457Controller(g, bear) != me.ID {
		t.Fatal("the caster did not gain control")
	}
	if c, _ := effectstest.BattlefieldCard(g, bear); c.Tapped || !game.HasKeyword(&c, "haste") {
		t.Errorf("the stolen creature is tapped=%v haste=%v, want untapped with haste", c.Tapped, game.HasKeyword(&c, "haste"))
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, bear)
	if !ok || c.Controller != opp.ID {
		t.Fatalf("control did not return (ok=%v)", ok)
	}
	if !c.Tapped {
		t.Error("the creature was not tapped when control was lost")
	}
}

func TestOldB457RayOfCommandRefusesYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", oldB457Bear, 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Ray of Command", "Instant", oldB457Ray, effectstest.CardTargets(mine)); err == nil {
		t.Error("a creature its caster controls was accepted")
	}
}

// --- Goblin Cadets ----------------------------------------------------------------------

func TestOldB457GoblinCadetsGoToTheOpponentWhenBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	cadets := effectstest.B12Push(g, me.ID, "Goblin Cadets", "Creature — Goblin", oldB457Cadets, 2, 1)
	blocker := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB457Bear, 2, 2)
	effectstest.Pr7bAttack(t, g, opp.ID, cadets)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blocker: cadets})
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB457Controller(g, cadets) != opp.ID {
		t.Errorf("the Cadets are controlled by %s, want the opponent", oldB457Controller(g, cadets))
	}
}

// --- Razorjaw Oni ------------------------------------------------------------------------

func TestOldB457RazorjawOniStopsBlackCreaturesBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Razorjaw Oni", "Creature — Demon Spirit", oldB457Oni, 4, 5)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", oldB457Bear, 2, 2)
	black := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Black Bear", TypeLine: oldB457Bear,
		Colors: []string{"B"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	green := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Bear", TypeLine: oldB457Bear,
		Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(black, attacker); err == nil {
		t.Error("a black creature was allowed to block")
	}
	if err := g.DeclareBlocker(green, attacker); err != nil {
		t.Errorf("a green creature was refused as a blocker: %v", err)
	}
}

// --- Perilous Research ---------------------------------------------------------------------

func TestOldB457PerilousResearchDrawsTwoThenSacrificesAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.StockLibraryFor(g, me, 10)
	keep := effectstest.B12Creature(g, me.ID, "Keep", oldB457Bear, 2, 2)
	lose := effectstest.B12Creature(g, me.ID, "Lose", oldB457Bear, 1, 1)
	hand := len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Perilous Research", "Instant", oldB457Research, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, lose)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards) - hand; got != 2 {
		t.Errorf("hand grew by %d, want 2 (the spell left it too)", got)
	}
	if _, ok := effectstest.BattlefieldCard(g, lose); ok {
		t.Error("the chosen permanent was not sacrificed")
	}
	if _, ok := effectstest.BattlefieldCard(g, keep); !ok {
		t.Error("the wrong permanent was sacrificed")
	}
}

// --- Trench Gorger ---------------------------------------------------------------------------

func TestOldB457TrenchGorgerExilesLandsAndTakesTheirCountAsBaseSize(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	var lands []uuid.UUID
	for i := 0; i < 3; i++ {
		c := game.NewCard("Library Forest", me.ID)
		c.TypeLine = "Basic Land — Forest"
		me.Library.PushTop(c)
		lands = append(lands, c.InstanceID)
	}
	spare := game.NewCard("Library Bear", me.ID)
	spare.TypeLine = oldB457Bear
	me.Library.PushTop(spare)
	gorger := effectstest.EnterFromHand(t, g, me.ID, "Trench Gorger", "Creature — Leviathan", oldB457Gorger)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.AnswerSearchByID(t, g, me.ID, lands...)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, gorger)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("the Gorger is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	exiled := 0
	for _, e := range g.Exile.Cards {
		for _, id := range lands {
			if e.InstanceID == id {
				exiled++
			}
		}
	}
	if exiled != 3 {
		t.Errorf("%d lands exiled, want 3", exiled)
	}
}

func TestOldB457TrenchGorgerKeepsItsSizeWhenYouDeclineTheSearch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	plain := effectstest.LayeredCard(t, g, effectstest.EnterFromHand(t, g, me.ID, "Plain Leviathan", "Creature — Leviathan", ""))
	gorger := effectstest.EnterFromHand(t, g, me.ID, "Trench Gorger", "Creature — Leviathan", oldB457Gorger)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, gorger)
	if c.CurrentPower() != plain.CurrentPower() || c.CurrentToughness() != plain.CurrentToughness() {
		t.Errorf("the Gorger is %d/%d after declining, want the size it entered with (%d/%d)", c.CurrentPower(), c.CurrentToughness(), plain.CurrentPower(), plain.CurrentToughness())
	}
}

// --- Riddlekeeper ------------------------------------------------------------------------------

func TestOldB457RiddlekeeperMillsTheAttackersController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.StockLibraryFor(g, me, 20)
	effectstest.StockLibraryFor(g, opp, 20)
	effectstest.B12Push(g, opp.ID, "Riddlekeeper", "Creature — Homunculus", oldB457Riddle, 1, 4)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", oldB457Bear, 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Sneaker", oldB457Bear, 2, 2)
	third := g.Seats[2]
	before := me.Library.Size()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(other, third.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - me.Library.Size(); got != 2 {
		t.Errorf("the attacker's controller milled %d, want 2 (one attacker aimed at Riddlekeeper's controller)", got)
	}
}

// --- Dominus of Fealty ----------------------------------------------------------------------------

func TestOldB457DominusStealsAPermanentAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Dominus of Fealty", "Creature — Spirit Avatar", oldB457Dominus, 4, 4)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB457Bear, 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bear).Tapped = true })
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 2)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 3)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB457Controller(g, bear) != me.ID {
		t.Fatal("control did not change")
	}
	if c, _ := effectstest.BattlefieldCard(g, bear); c.Tapped {
		t.Error("the stolen permanent is still tapped")
	}
}

// --- Gomazoa --------------------------------------------------------------------------------------

func TestOldB457GomazoaTucksItselfAndTheAttackerItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", oldB457Bear, 2, 2)
	gomazoa := effectstest.B12Push(g, opp.ID, "Gomazoa", "Creature — Jellyfish", oldB457Gomazoa, 0, 3)
	effectstest.Pr7bAttack(t, g, opp.ID, attacker)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{gomazoa: attacker})
	effectstest.P7Activate(t, g, opp, gomazoa, 0, game.ActivateAbilityParams{})
	for _, id := range []uuid.UUID{attacker, gomazoa} {
		if _, ok := effectstest.BattlefieldCard(g, id); ok {
			t.Errorf("%s is still on the battlefield", id)
		}
	}
	// The libraries are shuffled afterwards, so only that they hold the cards.
	if !me.Library.Contains(attacker) {
		t.Error("the attacker is not in its owner's library")
	}
	if !opp.Library.Contains(gomazoa) {
		t.Error("Gomazoa is not in its owner's library")
	}
}

func TestOldB457GomazoaAloneTucksOnlyItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bystander := effectstest.B12Creature(g, me.ID, "Bystander", oldB457Bear, 2, 2)
	gomazoa := effectstest.B12Push(g, me.ID, "Gomazoa", "Creature — Jellyfish", oldB457Gomazoa, 0, 3)
	effectstest.P7Activate(t, g, me, gomazoa, 0, game.ActivateAbilityParams{})
	if _, ok := effectstest.BattlefieldCard(g, gomazoa); ok {
		t.Error("Gomazoa stayed")
	}
	if _, ok := effectstest.BattlefieldCard(g, bystander); !ok {
		t.Error("an unrelated creature was tucked")
	}
}

// --- Martyr's Bond -----------------------------------------------------------------------------------

func TestOldB457MartyrsBondMakesEachOpponentSacrificeAPermanentOfASharedType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	s := g.Seats
	effectstest.B12Push(g, s[0].ID, "Martyr's Bond", "Enchantment", oldB457Bond, 0, 0)
	mine := effectstest.B12Creature(g, s[0].ID, "My Bear", oldB457Bear, 1, 1)
	cr1 := effectstest.B12Creature(g, s[1].ID, "Creature One", oldB457Bear, 1, 1)
	land1 := oldB457Land(g, s[1].ID)
	cr2 := effectstest.B12Creature(g, s[2].ID, "Creature Two", oldB457Bear, 1, 1)
	effectstest.P1Destroy(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, s[1].ID, cr1)
	effectstest.AnswerSacrifice(t, g, s[2].ID, cr2)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{cr1, cr2} {
		if _, ok := effectstest.BattlefieldCard(g, id); ok {
			t.Errorf("creature %s was not sacrificed", id)
		}
	}
	if _, ok := effectstest.BattlefieldCard(g, land1); !ok {
		t.Error("an opponent's land was sacrificed for a creature")
	}
}

func TestOldB457MartyrsBondIgnoresALandOfYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	s := g.Seats
	effectstest.B12Push(g, s[0].ID, "Martyr's Bond", "Enchantment", oldB457Bond, 0, 0)
	land := oldB457Land(g, s[0].ID)
	theirs := effectstest.B12Creature(g, s[1].ID, "Their Bear", oldB457Bear, 1, 1)
	effectstest.P1Destroy(t, g, land)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, theirs); !ok {
		t.Error("an opponent sacrificed for a land going to the graveyard")
	}
}

// --- Celestial Force ------------------------------------------------------------------------------------

func TestOldB457CelestialForceGainsThreeAtEveryUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Celestial Force", "Creature — Elemental", oldB457Force, 7, 7)
	life := me.Life
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d -> %d after another player's upkeep, want +3", life, me.Life)
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+6 {
		t.Errorf("life %d after two upkeeps, want %d", me.Life, life+6)
	}
}
