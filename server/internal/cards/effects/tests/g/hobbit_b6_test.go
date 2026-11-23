package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Radagast of Rhosgobel --------------------------------------------

func TestHobRadagastDiscountsOnlyTheFirstCreatureSpellEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.HobPush(t, g, me.ID, "Radagast of Rhosgobel", "Legendary Creature — Avatar Wizard", 2, 5)
	if got := effectstest.SelfPricedMV(t, g, me, "", "Creature — Bear", "{3}{G}"); got != 2 {
		t.Errorf("the first creature spell costs %d, want 2", got)
	}
	if got := effectstest.SelfPricedMV(t, g, me, "", "Instant", "{3}{G}"); got != 4 {
		t.Errorf("a noncreature spell costs %d, want 4", got)
	}
	if got := effectstest.SelfPricedMV(t, g, me, "", "Creature — Bear", "{G}"); got != 1 {
		t.Errorf("a one-mana creature costs %d, want 1 (only generic mana is reduced)", got)
	}
	if got := effectstest.SelfPricedMV(t, g, opp, "", "Creature — Bear", "{3}{G}"); got != 4 {
		t.Errorf("an opponent's creature spell costs %d, want 4", got)
	}

	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "GG")
	effectstest.CastFromHandForTest(t, g, me, "First Bear", "Creature — Bear", "{2}{G}", "", effectstest.Strict)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.SelfPricedMV(t, g, me, "", "Creature — Bear", "{3}{G}"); got != 4 {
		t.Errorf("the second creature spell costs %d, want 4", got)
	}
	if got := effectstest.HobB6BattlefieldCount(g, me.ID, "First Bear"); got != 1 {
		t.Errorf("the discounted bear was not cast for {G}{G}: %d on the battlefield", got)
	}
}

func TestHobRadagastLetsTheFirstCreatureSpellBeCastAtInstantSpeedOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	radagast := effectstest.HobPush(t, g, me.ID, "Radagast of Rhosgobel", "Legendary Creature — Avatar Wizard", 2, 5)
	// The beginning of combat is a step where only an instant-speed spell can be cast.
	if _, err := hobFlashCast(t, g, "First Bear", "Creature — Bear", "", 2, 2); err != nil {
		t.Fatalf("the first creature spell could not be cast with flash: %v", err)
	}
	if _, err := hobFlashCast(t, g, "Second Bear", "Creature — Bear", "", 2, 2); err == nil {
		t.Error("the second creature spell was cast with flash")
	}
	if got := effectstest.HobB6BattlefieldCount(g, me.ID, "Second Bear"); got != 0 {
		t.Errorf("the second bear entered the battlefield (%d)", got)
	}
	_ = radagast
}

func TestHobRadagastGivesNoFlashToANoncreatureSpellOrAfterItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	radagast := effectstest.HobPush(t, g, me.ID, "Radagast of Rhosgobel", "Legendary Creature — Avatar Wizard", 2, 5)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	id := effectstest.HandCardFull(me, "Mystery Sorcery", "Sorcery", "", "", nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err == nil {
		t.Error("a sorcery was cast at instant speed")
	}
	effectstest.HobDestroy(t, g, radagast)
	if _, err := hobFlashCast(t, g, "Late Bear", "Creature — Bear", "", 2, 2); err == nil {
		t.Error("a creature spell was cast with flash after Radagast left")
	}
}

// --- Through the Forest Gate ------------------------------------------

// hobB6SeedForestGate stacks the top of `p`'s library with four land
// cards among four spells (top first: land, spell, land, spell, ...), the
// 20-card filler below them, and a land at the very bottom: the twenty-first
// card, which a "top twenty" look must not reach.
func hobB6SeedForestGate(p *game.Player) (lands []uuid.UUID, far uuid.UUID) {
	for i := 3; i >= 0; i-- {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Creature — Bear", Owner: p.ID, Controller: p.ID})
		id := uuid.New()
		p.Library.PushTop(game.Card{InstanceID: id, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: p.ID, Controller: p.ID})
		lands = append([]uuid.UUID{id}, lands...)
	}
	// 4*2 pushed + the filler already there: pad to twenty above the far land.
	for p.Library.Size() < 20 {
		p.Library.PushBottom(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
	far = uuid.New()
	p.Library.PushBottom(game.Card{InstanceID: far, Name: "Far Forest", TypeLine: "Basic Land — Forest", Owner: p.ID, Controller: p.ID})
	return lands, far
}

func hobB6CastForestGate(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if _, err := effectstest.HobB3Cast(t, g, me, "Through the Forest Gate", "Sorcery", "{6}{G}{G}", "{G}{G}{C}{C}{C}{C}{C}{C}", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestHobThroughTheForestGatePutsChosenLandsOntoTheBattlefieldTappedShufflesAndGains8(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lands, far := hobB6SeedForestGate(me)
	sizeBefore := me.Library.Size()
	var topBefore []uuid.UUID
	g.ReadSnapshot(func() {
		for i := 0; i < 8; i++ {
			topBefore = append(topBefore, me.Library.Cards[me.Library.Size()-1-i].InstanceID)
		}
	})
	life := effectstest.HobB6Life(g, me)
	hobB6CastForestGate(t, g, me)

	prompt := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("the Gate offered no land choice")
	}
	if len(prompt.ChooseCards) != 4 {
		t.Errorf("%d cards offered, want the four lands among the top twenty (not the twenty-first)", len(prompt.ChooseCards))
	}
	for _, id := range prompt.ChooseCards {
		if id == far {
			t.Error("the land at depth 21 was offered")
		}
	}
	if prompt.ChooseMin != 0 || prompt.ChooseMax != 0 && prompt.ChooseMax < 4 {
		t.Errorf("bounds %d..%d, want any number from zero", prompt.ChooseMin, prompt.ChooseMax)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, lands[0], lands[1], lands[2])
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range lands[:3] {
		c, ok := effectstest.BattlefieldCard(g, id)
		if !ok {
			t.Fatalf("a chosen land did not enter the battlefield")
		}
		if !c.Tapped {
			t.Error("a land entered untapped")
		}
	}
	if _, ok := effectstest.BattlefieldCard(g, lands[3]); ok {
		t.Error("the unchosen land entered the battlefield")
	}
	if got := effectstest.HobB6Life(g, me); got != life+8 {
		t.Errorf("life %d -> %d, want +8", life, got)
	}
	if got := me.Library.Size(); got != sizeBefore-3 {
		t.Errorf("library %d -> %d, want three fewer", sizeBefore, got)
	}
	same := true
	g.ReadSnapshot(func() {
		for i, id := range topBefore {
			if me.Library.Cards[me.Library.Size()-1-i].InstanceID != id {
				same = false
			}
		}
	})
	if same {
		t.Error("the library was not shuffled")
	}
}

func TestHobThroughTheForestGateWithNoLandStillShufflesAndGains8(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := effectstest.HobB6Life(g, me)
	hobB6CastForestGate(t, g, me)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a land choice was offered with no land to find")
	}
	if got := effectstest.HobB6Life(g, me); got != life+8 {
		t.Errorf("life %d -> %d, want +8", life, got)
	}
}

func TestHobThroughTheForestGateMayPutNoLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lands, _ := hobB6SeedForestGate(me)
	life := effectstest.HobB6Life(g, me)
	hobB6CastForestGate(t, g, me)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range lands {
		if _, ok := effectstest.BattlefieldCard(g, id); ok {
			t.Error("a land entered although none was chosen")
		}
	}
	if got := effectstest.HobB6Life(g, me); got != life+8 {
		t.Errorf("life %d -> %d, want +8", life, got)
	}
}

// --- Thranduil, the Strategist ----------------------------------------

func TestHobThranduilGivesOtherElvesATapAbilityForGOrU(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thranduil := effectstest.HobPush(t, g, me.ID, "Thranduil, the Strategist", "Legendary Creature — Elf Noble", 4, 4)
	elf := effectstest.B12Push(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", "", 1, 1)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Their Elf", "Creature — Elf Druid", "", 1, 1)
	if n := effectstest.GrantedManaRowsOf(g, elf); n != 1 {
		t.Errorf("your other Elf has %d granted mana rows, want 1", n)
	}
	for name, id := range map[string]uuid.UUID{"Thranduil himself": thranduil, "a Bear": bear, "an opposing Elf": theirs} {
		if n := effectstest.GrantedManaRowsOf(g, id); n != 0 {
			t.Errorf("%s has %d granted mana rows, want 0", name, n)
		}
	}
	for _, color := range []string{"G", "U"} {
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == elf {
					g.Battlefield.Cards[i].Tapped = false
				}
			}
		})
		if err := effectstest.TapGrantedMana(t, g, me.ID, elf, color); err != nil {
			t.Fatalf("tapping for {%s}: %v", color, err)
		}
		if got := effectstest.GmrPoolCount(me, color); got != 1 {
			t.Errorf("pool has %d {%s}, want 1", got, color)
		}
	}
	if c, _ := effectstest.BattlefieldCard(g, elf); !c.Tapped {
		t.Error("the Elf did not tap for mana")
	}
}

func TestHobThranduilLandfallMakesAnElfThatAlsoTapsForMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.HobPush(t, g, me.ID, "Thranduil, the Strategist", "Legendary Creature — Elf Noble", 4, 4)
	effectstest.B12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HobB6BattlefieldCount(g, me.ID, "Elf"); got != 1 {
		t.Fatalf("%d Elf tokens, want 1", got)
	}
	var token uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.Name == "Elf" {
				token = c.InstanceID
				if c.Power != 1 || c.Toughness != 1 || !c.HasColorWord("G") || !c.IsToken() {
					t.Errorf("token = %dx%d, colours %v, token %v; want a 1/1 green token", c.Power, c.Toughness, c.Colors, c.IsToken())
				}
			}
		}
	})
	if n := effectstest.GrantedManaRowsOf(g, token); n != 1 {
		t.Errorf("the Elf token has %d granted mana rows, want 1", n)
	}
	if got := effectstest.HobB6BattlefieldCount(g, opp.ID, "Elf"); got != 0 {
		t.Errorf("the opponent has %d Elf tokens, want 0", got)
	}
}

// --- Aragorn, the Uniter ----------------------------------------------

// hobB6AragornCast puts Aragorn on the battlefield and casts a spell of
// the given cost and type, paying with `mana`.
func hobB6AragornCast(t *testing.T, g *game.Game, me *game.Player, cost, mana string) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.HobPush(t, g, me.ID, "Aragorn, the Uniter", "Legendary Creature — Human Noble", 5, 5)
	effectstest.FloatForTest(g, me, mana)
	effectstest.CastFromHandForTest(t, g, me, "Spell", "Instant", cost, "", effectstest.Strict)
}

func TestHobAragornAWhiteSpellMakesAHumanSoldier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hobB6AragornCast(t, g, me, "{W}", "W")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HobB6BattlefieldCount(g, me.ID, "Human Soldier"); got != 1 {
		t.Errorf("%d Human Soldier tokens, want 1", got)
	}
}

func TestHobAragornABlueSpellScries2(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hobB6AragornCast(t, g, me, "{U}", "U")
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("scry looked at %d cards, want 2", len(c.ScryCards))
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
}

func TestHobAragornARedSpellDeals3ToTargetOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life, mine := effectstest.HobB6Life(g, opp), effectstest.HobB6Life(g, me)
	hobB6AragornCast(t, g, me, "{R}", "R")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HobB6Life(g, opp); got != life-3 {
		t.Errorf("the opponent's life %d -> %d, want -3", life, got)
	}
	if got := effectstest.HobB6Life(g, me); got != mine {
		t.Errorf("my life changed %d -> %d", mine, got)
	}
}

func TestHobAragornAGreenSpellPumpsATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	hobB6AragornCast(t, g, me, "{G}", "G")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.AnswerPickTarget(t, g, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 6 || tough != 6 {
		t.Errorf("the bear is %d/%d, want 6/6", p, tough)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the bear is %d/%d after the turn ends, want 2/2", p, tough)
	}
}

func TestHobAragornAMulticolouredSpellTriggersEachColourOnceAndColourlessNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hobB6AragornCast(t, g, me, "{W}{U}", "WU")
	// Two of Aragorn's triggers fire at once; the controller orders them if
	// the engine finds the order matters.
	if effectstest.TriggerOrderPrompt(g) != nil {
		effectstest.AnswerTriggerOrderInOfferedOrder(t, g)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Fatal("the blue half did not scry")
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HobB6BattlefieldCount(g, me.ID, "Human Soldier"); got != 1 {
		t.Errorf("%d Human Soldier tokens, want 1", got)
	}

	effectstest.FloatForTest(g, me, "CC")
	effectstest.CastFromHandForTest(t, g, me, "Colourless Spell", "Instant", "{C}{C}", "", effectstest.Strict)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HobB6BattlefieldCount(g, me.ID, "Human Soldier"); got != 1 {
		t.Errorf("a colourless spell triggered something: %d Human Soldier tokens", got)
	}
	if effectstest.ScryChoiceFor(g, me.ID) != nil {
		t.Error("a colourless spell scried")
	}
}

// --- Gollum, the Abandoned --------------------------------------------

func TestHobGollumCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gollum := effectstest.HobPush(t, g, opp.ID, "Gollum, the Abandoned", "Legendary Creature — Halfling Horror", 2, 2)
	other := effectstest.PushVanillaCreature(g, opp.ID, "Blocker", 2, 2)
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(gollum, attacker); err == nil {
		t.Error("Gollum blocked")
	}
	if err := g.DeclareBlocker(other, attacker); err != nil {
		t.Errorf("another creature could not block: %v", err)
	}
}

func TestHobGollumExilesAnOpponentsGraveyardCardAndEachOpponentLoses2(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushCatalogGraveyardCard(opp, "Their Card", "Creature — Bear", "", 2, 2)
	mine := effectstest.PushCatalogGraveyardCard(me, "My Card", "Creature — Bear", "", 2, 2)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = effectstest.HobB6Life(g, p)
	}
	hobCast(t, g, "Gollum, the Abandoned", "Legendary Creature — Halfling Horror", 2, 2)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.AnswerPickTarget(t, g, theirs)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(theirs) {
		t.Error("the opponent's card is still in their graveyard")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("my own graveyard card was exiled")
	}
	for _, p := range g.Seats {
		want := lives[p.ID]
		if p.ID != me.ID {
			want -= 2
		}
		if got := effectstest.HobB6Life(g, p); got != want {
			t.Errorf("seat %s life %d, want %d", p.ID, got, want)
		}
	}
}

func TestHobGollumStillDrainsWithNothingToExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushCatalogGraveyardCard(me, "My Card", "Creature — Bear", "", 2, 2)
	life := effectstest.HobB6Life(g, opp)
	hobCast(t, g, "Gollum, the Abandoned", "Legendary Creature — Halfling Horror", 2, 2)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Errorf("a target prompt opened with no card in an opponent's graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(mine) {
		t.Error("my own graveyard card was exiled")
	}
	if got := effectstest.HobB6Life(g, opp); got != life-2 {
		t.Errorf("the opponent's life %d -> %d, want -2", life, got)
	}
}

func TestHobGollumReturnsFromTheGraveyardForTwoAndASacrificeAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	gollum := effectstest.PushCatalogGraveyardCard(me, "Gollum, the Abandoned", "Legendary Creature — Halfling Horror", effectstest.HobOracle(t, "Gollum, the Abandoned"), 2, 2)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	rock := effectstest.PushArtifactForTest(g, me.ID, "Mind Stone")
	land := hobB3Perm(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	activate := func(sac ...uuid.UUID) error {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
			t.Fatal(err)
		}
		return g.ActivateCatalogAbility(me.ID, gollum, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: sac})
	}
	if err := activate(); err == nil {
		t.Error("activated without sacrificing")
	}
	if err := activate(land); err == nil {
		t.Error("sacrificed a land")
	}
	if err := activate(rock); err != nil {
		t.Fatalf("sacrificing an artifact: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(gollum) {
		t.Fatal("Gollum did not return to hand")
	}
	if _, ok := effectstest.BattlefieldCard(g, rock); ok {
		t.Error("the artifact was not sacrificed")
	}

	// A creature pays too, and not at instant speed.
	second := effectstest.PushCatalogGraveyardCard(me, "Gollum, the Abandoned", "Legendary Creature — Halfling Horror", effectstest.HobOracle(t, "Gollum, the Abandoned"), 2, 2)
	gollum = second
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := activate(bear); err == nil {
		t.Error("activated at instant speed")
	}
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	if err := activate(bear); err != nil {
		t.Fatalf("sacrificing a creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(second) {
		t.Error("Gollum did not return to hand the second time")
	}
}

// --- Mirkwood Elk ------------------------------------------------------

func TestHobMirkwoodElkEntersAndReturnsAnElfGainingLifeEqualToItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := effectstest.PushCatalogGraveyardCard(me, "Elvish Warrior", "Creature — Elf Warrior", "", 3, 2)
	bear := effectstest.PushCatalogGraveyardCard(me, "Bear", "Creature — Bear", "", 5, 5)
	theirs := effectstest.PushCatalogGraveyardCard(opp, "Their Elf", "Creature — Elf Warrior", "", 4, 4)
	life := effectstest.HobB6Life(g, me)
	hobCast(t, g, "Mirkwood Elk", "Creature — Elk", 6, 6)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.AnswerPickTarget(t, g, elf)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(elf) {
		t.Error("the Elf did not return to hand")
	}
	if !me.Graveyard.Contains(bear) || !opp.Graveyard.Contains(theirs) {
		t.Error("a card that is not my Elf left a graveyard")
	}
	if got := effectstest.HobB6Life(g, me); got != life+3 {
		t.Errorf("life %d -> %d, want +3", life, got)
	}
}

func TestHobMirkwoodElkWithNoElfDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogGraveyardCard(me, "Bear", "Creature — Bear", "", 5, 5)
	life, hand := effectstest.HobB6Life(g, me), me.Hand.Size()
	hobCast(t, g, "Mirkwood Elk", "Creature — Elk", 6, 6)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Error("a target prompt opened with no Elf card in the graveyard")
	}
	if got := effectstest.HobB6Life(g, me); got != life {
		t.Errorf("life changed %d -> %d", life, got)
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d -> %d, want no card gained (the Elk was put there to be cast)", hand, got)
	}
}

func TestHobMirkwoodElkAttackingReturnsAnElfToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elk := effectstest.HobPush(t, g, me.ID, "Mirkwood Elk", "Creature — Elk", 6, 6)
	elf := effectstest.PushCatalogGraveyardCard(me, "Elvish Warrior", "Creature — Elf Warrior", "", 2, 2)
	life := effectstest.HobB6Life(g, me)
	effectstest.DeclareAttack(t, g, opp.ID, elk)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.AnswerPickTarget(t, g, elf)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(elf) {
		t.Error("the Elf did not return to hand when the Elk attacked")
	}
	if got := effectstest.HobB6Life(g, me); got != life+2 {
		t.Errorf("life %d -> %d, want +2", life, got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, elk, "trample") {
		t.Error("the Elk lacks trample")
	}
}

// --- Celeborn the Wise -------------------------------------------------

func TestHobCelebornScriesOnceWhenElvesAttackAndGrowsPerCardLookedAt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	celeborn := effectstest.HobPush(t, g, me.ID, "Celeborn the Wise", "Legendary Creature — Elf Noble", 3, 3)
	elf := effectstest.B12Push(g, me.ID, "Elvish Scout", "Creature — Elf Scout", "", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, celeborn, elf)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("attacking with Elves did not scry")
	}
	if len(c.ScryCards) != 1 {
		t.Errorf("scry looked at %d cards, want 1", len(c.ScryCards))
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) != nil {
		t.Error("two attacking Elves scried twice")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, celeborn); p != 4 || tough != 4 {
		t.Errorf("Celeborn is %d/%d, want 4/4", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, elf); p != 1 || tough != 1 {
		t.Errorf("the other Elf is %d/%d, want 1/1", p, tough)
	}
}

func TestHobCelebornDoesNotScryForANonElfAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.HobPush(t, g, me.ID, "Celeborn the Wise", "Legendary Creature — Elf Noble", 3, 3)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) != nil {
		t.Error("a Bear attacking made Celeborn scry")
	}
}

func TestHobCelebornGetsOneCounterPerCardAnyScryLookedAt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	celeborn := effectstest.HobPush(t, g, me.ID, "Celeborn the Wise", "Legendary Creature — Elf Noble", 3, 3)
	hobB6AragornCast(t, g, me, "{U}", "U")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, celeborn); p != 5 || tough != 5 {
		t.Errorf("Celeborn is %d/%d after a scry 2, want 5/5", p, tough)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, celeborn); p != 3 || tough != 3 {
		t.Errorf("Celeborn is %d/%d after the turn, want 3/3", p, tough)
	}
}

// --- Fateful Discovery -------------------------------------------------

func TestHobFatefulDiscoveryDrawsForEachArtifactYouControlThatEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.HobPush(t, g, me.ID, "Fateful Discovery", "Enchantment", 0, 0)
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 colorless Servo artifact"), 2); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d -> %d, want +2 for two artifact tokens", hand, got)
	}
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 green Snake"), 1); err != nil {
			t.Fatal(err)
		}
		if err := g.CreateTokenForEffect(opp.ID, effects.TokenCard("1/1 colorless Servo artifact"), 1); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand = %d, want a nonartifact and an opponent's artifact to draw nothing (%d)", got, hand+2)
	}
}

// --- Fog on the Barrow-Downs ------------------------------------------

func hobB6Fog(t *testing.T, g *game.Game, me *game.Player, host uuid.UUID) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := effectstest.CastFromHandForTest(t, g, me, "Fog on the Barrow-Downs", "Enchantment — Aura", "", effectstest.HobOracle(t, "Fog on the Barrow-Downs"),
		game.CastSpellParams{Targets: effectstest.CardRefs(host)})
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestHobFogOnTheBarrowDownsMakesTheCreatureASpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Push(g, opp.ID, "Hill Giant Archer", "Artifact Creature — Giant Archer", "", 3, 3)
	aura := hobB6Fog(t, g, me, host)
	got := effectstest.GaLayered(t, g, host)
	if !got.HasSubtype("Spirit") || got.HasSubtype("Giant") || got.HasSubtype("Archer") {
		t.Errorf("the host is %q, want a Spirit and no other creature type", got.TypeLine)
	}
	if !got.IsCreature() || !got.IsArtifact() {
		t.Error("the host lost a card type")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, host); p != 3 || tough != 3 {
		t.Errorf("the host is %d/%d, want unchanged 3/3", p, tough)
	}
	// Destroy the Aura and the creature is its old self again.
	effectstest.HobDestroy(t, g, aura)
	back := effectstest.GaLayered(t, g, host)
	if !back.HasSubtype("Giant") || back.HasSubtype("Spirit") {
		t.Errorf("after the Aura left the host is %q, want a Giant Archer", back.TypeLine)
	}
}

func TestHobFogOnTheBarrowDownsStopsTheCreatureAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.PushVanillaCreature(g, me.ID, "Host", 2, 2)
	free := effectstest.PushVanillaCreature(g, me.ID, "Free", 2, 2)
	hobB6Fog(t, g, me, host)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(host, opp.ID); err == nil {
		t.Error("the enchanted creature attacked")
	}
	if err := g.DeclareAttacker(free, opp.ID); err != nil {
		t.Errorf("another creature could not attack: %v", err)
	}
}

func TestHobFogOnTheBarrowDownsStopsTheCreatureBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blocker := effectstest.PushVanillaCreature(g, opp.ID, "Blocker", 2, 2)
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 1, 1)
	hobB6Fog(t, g, me, blocker)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err == nil {
		t.Error("the enchanted creature blocked")
	}
}

// --- Thorin's Last Stand ----------------------------------------------

func hobB6ThorinCast(t *testing.T, g *game.Game, modes []int, targets ...uuid.UUID) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.HandCardFull(me, "Thorin's Last Stand", "Instant", "", effectstest.HobOracle(t, "Thorin's Last Stand"), nil)
	err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: modes, Targets: effectstest.CardRefs(targets...)})
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return err
}

func TestHobThorinsLastStandPumpsYourCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.PushVanillaCreature(g, me.ID, "A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "B", 1, 1)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	if err := hobB6ThorinCast(t, g, []int{0}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	for id, want := range map[uuid.UUID][2]int{a: {4, 3}, b: {3, 2}, theirs: {2, 2}} {
		if p, tough := effectstest.EffectivePTOf(t, g, id); p != want[0] || tough != want[1] {
			t.Errorf("%s is %d/%d, want %d/%d", id, p, tough, want[0], want[1])
		}
	}
	late := effectstest.PushVanillaCreature(g, me.ID, "Late", 2, 2)
	if p, tough := effectstest.EffectivePTOf(t, g, late); p != 2 || tough != 2 {
		t.Errorf("a creature that arrived later is %d/%d, want 2/2 (CR 611.2c)", p, tough)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, a); p != 2 || tough != 2 {
		t.Errorf("A is %d/%d after the turn, want 2/2", p, tough)
	}
}

func TestHobThorinsLastStandDestroysAnArtifactOrEnchantmentAndGains2(t *testing.T) {
	for _, tc := range []struct {
		Name, TypeLine string
	}{{"artifact", "Artifact"}, {"enchantment", "Enchantment"}} {
		t.Run(tc.Name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			target := hobB3Perm(g, opp.ID, game.Card{Name: "Target", TypeLine: tc.TypeLine})
			life := effectstest.HobB6Life(g, me)
			if err := hobB6ThorinCast(t, g, []int{1}, target); err != nil {
				t.Fatalf("cast: %v", err)
			}
			if _, ok := effectstest.BattlefieldCard(g, target); ok {
				t.Errorf("the %s was not destroyed", tc.Name)
			}
			if got := effectstest.HobB6Life(g, me); got != life+2 {
				t.Errorf("life %d -> %d, want +2", life, got)
			}
		})
	}
}

func TestHobThorinsLastStandCantTargetACreatureOrLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	land := hobB3Perm(g, opp.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	if err := hobB6ThorinCast(t, g, []int{1}, bear); err == nil {
		t.Error("targeted a creature")
	}
	if err := hobB6ThorinCast(t, g, []int{1}, land); err == nil {
		t.Error("targeted a land")
	}
}
