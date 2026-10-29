package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4454, The Hobbit (HOB / HOC) cards, half B, batch 6. Every test drives
// the registered catalog row through the engine.

// hobB6BattlefieldCount is how many permanents named `name` the player
// controls.
func hobB6BattlefieldCount(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == controller && c.Name == name {
				n++
			}
		}
	})
	return n
}

// hobB6Life is a player's life total, read under the game's lock.
func hobB6Life(g *game.Game, p *game.Player) int {
	life := 0
	g.ReadSnapshot(func() { life = p.Life })
	return life
}

// --- Radagast of Rhosgobel --------------------------------------------

func TestHobRadagastDiscountsOnlyTheFirstCreatureSpellEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hobPush(t, g, me.ID, "Radagast of Rhosgobel", "Legendary Creature — Avatar Wizard", 2, 5)
	if got := selfPricedMV(t, g, me, "", "Creature — Bear", "{3}{G}"); got != 2 {
		t.Errorf("the first creature spell costs %d, want 2", got)
	}
	if got := selfPricedMV(t, g, me, "", "Instant", "{3}{G}"); got != 4 {
		t.Errorf("a noncreature spell costs %d, want 4", got)
	}
	if got := selfPricedMV(t, g, me, "", "Creature — Bear", "{G}"); got != 1 {
		t.Errorf("a one-mana creature costs %d, want 1 (only generic mana is reduced)", got)
	}
	if got := selfPricedMV(t, g, opp, "", "Creature — Bear", "{3}{G}"); got != 4 {
		t.Errorf("an opponent's creature spell costs %d, want 4", got)
	}

	advanceToMain(t, g)
	floatForTest(g, me, "GG")
	castFromHandForTest(t, g, me, "First Bear", "Creature — Bear", "{2}{G}", "", strict)
	passPriorityAroundTable(t, g)
	if got := selfPricedMV(t, g, me, "", "Creature — Bear", "{3}{G}"); got != 4 {
		t.Errorf("the second creature spell costs %d, want 4", got)
	}
	if got := hobB6BattlefieldCount(g, me.ID, "First Bear"); got != 1 {
		t.Errorf("the discounted bear was not cast for {G}{G}: %d on the battlefield", got)
	}
}

func TestHobRadagastLetsTheFirstCreatureSpellBeCastAtInstantSpeedOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	radagast := hobPush(t, g, me.ID, "Radagast of Rhosgobel", "Legendary Creature — Avatar Wizard", 2, 5)
	// The beginning of combat is a step where only an instant-speed spell can be cast.
	if _, err := hobFlashCast(t, g, "First Bear", "Creature — Bear", "", 2, 2); err != nil {
		t.Fatalf("the first creature spell could not be cast with flash: %v", err)
	}
	if _, err := hobFlashCast(t, g, "Second Bear", "Creature — Bear", "", 2, 2); err == nil {
		t.Error("the second creature spell was cast with flash")
	}
	if got := hobB6BattlefieldCount(g, me.ID, "Second Bear"); got != 0 {
		t.Errorf("the second bear entered the battlefield (%d)", got)
	}
	_ = radagast
}

func TestHobRadagastGivesNoFlashToANoncreatureSpellOrAfterItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	radagast := hobPush(t, g, me.ID, "Radagast of Rhosgobel", "Legendary Creature — Avatar Wizard", 2, 5)
	advanceTo(t, g, game.StepBeginCombat)
	id := handCardFull(me, "Mystery Sorcery", "Sorcery", "", "", nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err == nil {
		t.Error("a sorcery was cast at instant speed")
	}
	hobDestroy(t, g, radagast)
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
	advanceTo(t, g, game.StepPrecombatMain)
	if _, err := hobB3Cast(t, g, me, "Through the Forest Gate", "Sorcery", "{6}{G}{G}", "{G}{G}{C}{C}{C}{C}{C}{C}", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestHobThroughTheForestGatePutsChosenLandsOntoTheBattlefieldTappedShufflesAndGains8(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lands, far := hobB6SeedForestGate(me)
	sizeBefore := me.Library.Size()
	var topBefore []uuid.UUID
	g.ReadSnapshot(func() {
		for i := 0; i < 8; i++ {
			topBefore = append(topBefore, me.Library.Cards[me.Library.Size()-1-i].InstanceID)
		}
	})
	life := hobB6Life(g, me)
	hobB6CastForestGate(t, g, me)

	prompt := chooseCardsChoiceFor(g, me.ID)
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
	answerChooseCards(t, g, me.ID, lands[0], lands[1], lands[2])
	passPriorityAroundTable(t, g)

	for _, id := range lands[:3] {
		c, ok := battlefieldCard(g, id)
		if !ok {
			t.Fatalf("a chosen land did not enter the battlefield")
		}
		if !c.Tapped {
			t.Error("a land entered untapped")
		}
	}
	if _, ok := battlefieldCard(g, lands[3]); ok {
		t.Error("the unchosen land entered the battlefield")
	}
	if got := hobB6Life(g, me); got != life+8 {
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := hobB6Life(g, me)
	hobB6CastForestGate(t, g, me)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a land choice was offered with no land to find")
	}
	if got := hobB6Life(g, me); got != life+8 {
		t.Errorf("life %d -> %d, want +8", life, got)
	}
}

func TestHobThroughTheForestGateMayPutNoLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lands, _ := hobB6SeedForestGate(me)
	life := hobB6Life(g, me)
	hobB6CastForestGate(t, g, me)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	for _, id := range lands {
		if _, ok := battlefieldCard(g, id); ok {
			t.Error("a land entered although none was chosen")
		}
	}
	if got := hobB6Life(g, me); got != life+8 {
		t.Errorf("life %d -> %d, want +8", life, got)
	}
}

// --- Thranduil, the Strategist ----------------------------------------

func TestHobThranduilGivesOtherElvesATapAbilityForGOrU(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thranduil := hobPush(t, g, me.ID, "Thranduil, the Strategist", "Legendary Creature — Elf Noble", 4, 4)
	elf := b12Push(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", "", 1, 1)
	bear := b12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := b12Push(g, opp.ID, "Their Elf", "Creature — Elf Druid", "", 1, 1)
	if n := grantedManaRowsOf(g, elf); n != 1 {
		t.Errorf("your other Elf has %d granted mana rows, want 1", n)
	}
	for name, id := range map[string]uuid.UUID{"Thranduil himself": thranduil, "a Bear": bear, "an opposing Elf": theirs} {
		if n := grantedManaRowsOf(g, id); n != 0 {
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
		if err := tapGrantedMana(t, g, me.ID, elf, color); err != nil {
			t.Fatalf("tapping for {%s}: %v", color, err)
		}
		if got := gmrPoolCount(me, color); got != 1 {
			t.Errorf("pool has %d {%s}, want 1", got, color)
		}
	}
	if c, _ := battlefieldCard(g, elf); !c.Tapped {
		t.Error("the Elf did not tap for mana")
	}
}

func TestHobThranduilLandfallMakesAnElfThatAlsoTapsForMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hobPush(t, g, me.ID, "Thranduil, the Strategist", "Legendary Creature — Elf Noble", 4, 4)
	b12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := hobB6BattlefieldCount(g, me.ID, "Elf"); got != 1 {
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
	if n := grantedManaRowsOf(g, token); n != 1 {
		t.Errorf("the Elf token has %d granted mana rows, want 1", n)
	}
	if got := hobB6BattlefieldCount(g, opp.ID, "Elf"); got != 0 {
		t.Errorf("the opponent has %d Elf tokens, want 0", got)
	}
}

// --- Aragorn, the Uniter ----------------------------------------------

// hobB6AragornCast puts Aragorn on the battlefield and casts a spell of
// the given cost and type, paying with `mana`.
func hobB6AragornCast(t *testing.T, g *game.Game, me *game.Player, cost, mana string) {
	t.Helper()
	advanceToMain(t, g)
	hobPush(t, g, me.ID, "Aragorn, the Uniter", "Legendary Creature — Human Noble", 5, 5)
	floatForTest(g, me, mana)
	castFromHandForTest(t, g, me, "Spell", "Instant", cost, "", strict)
}

func TestHobAragornAWhiteSpellMakesAHumanSoldier(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hobB6AragornCast(t, g, me, "{W}", "W")
	passPriorityAroundTable(t, g)
	if got := hobB6BattlefieldCount(g, me.ID, "Human Soldier"); got != 1 {
		t.Errorf("%d Human Soldier tokens, want 1", got)
	}
}

func TestHobAragornABlueSpellScries2(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hobB6AragornCast(t, g, me, "{U}", "U")
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("scry looked at %d cards, want 2", len(c.ScryCards))
	}
	answerScryKeepAll(t, g, me.ID)
}

func TestHobAragornARedSpellDeals3ToTargetOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life, mine := hobB6Life(g, opp), hobB6Life(g, me)
	hobB6AragornCast(t, g, me, "{R}", "R")
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		answerPickTargetPlayer(t, g, opp.ID)
	}
	passPriorityAroundTable(t, g)
	if got := hobB6Life(g, opp); got != life-3 {
		t.Errorf("the opponent's life %d -> %d, want -3", life, got)
	}
	if got := hobB6Life(g, me); got != mine {
		t.Errorf("my life changed %d -> %d", mine, got)
	}
}

func TestHobAragornAGreenSpellPumpsATargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	hobB6AragornCast(t, g, me, "{G}", "G")
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		answerPickTarget(t, g, bear)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 6 || tough != 6 {
		t.Errorf("the bear is %d/%d, want 6/6", p, tough)
	}
	advanceToNextSeatsTurn(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the bear is %d/%d after the turn ends, want 2/2", p, tough)
	}
}

func TestHobAragornAMulticolouredSpellTriggersEachColourOnceAndColourlessNone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hobB6AragornCast(t, g, me, "{W}{U}", "WU")
	// Two of Aragorn's triggers fire at once; the controller orders them if
	// the engine finds the order matters.
	if triggerOrderPrompt(g) != nil {
		answerTriggerOrderInOfferedOrder(t, g)
	}
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatal("the blue half did not scry")
	}
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := hobB6BattlefieldCount(g, me.ID, "Human Soldier"); got != 1 {
		t.Errorf("%d Human Soldier tokens, want 1", got)
	}

	floatForTest(g, me, "CC")
	castFromHandForTest(t, g, me, "Colourless Spell", "Instant", "{C}{C}", "", strict)
	passPriorityAroundTable(t, g)
	if got := hobB6BattlefieldCount(g, me.ID, "Human Soldier"); got != 1 {
		t.Errorf("a colourless spell triggered something: %d Human Soldier tokens", got)
	}
	if scryChoiceFor(g, me.ID) != nil {
		t.Error("a colourless spell scried")
	}
}

// --- Gollum, the Abandoned --------------------------------------------

func TestHobGollumCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gollum := hobPush(t, g, opp.ID, "Gollum, the Abandoned", "Legendary Creature — Halfling Horror", 2, 2)
	other := pushVanillaCreature(g, opp.ID, "Blocker", 2, 2)
	attacker := pushVanillaCreature(g, me.ID, "Attacker", 1, 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(gollum, attacker); err == nil {
		t.Error("Gollum blocked")
	}
	if err := g.DeclareBlocker(other, attacker); err != nil {
		t.Errorf("another creature could not block: %v", err)
	}
}

func TestHobGollumExilesAnOpponentsGraveyardCardAndEachOpponentLoses2(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushCatalogGraveyardCard(opp, "Their Card", "Creature — Bear", "", 2, 2)
	mine := pushCatalogGraveyardCard(me, "My Card", "Creature — Bear", "", 2, 2)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = hobB6Life(g, p)
	}
	hobCast(t, g, "Gollum, the Abandoned", "Legendary Creature — Halfling Horror", 2, 2)
	if latestPickTarget(g, me.ID) != nil {
		answerPickTarget(t, g, theirs)
	}
	passPriorityAroundTable(t, g)
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
		if got := hobB6Life(g, p); got != want {
			t.Errorf("seat %s life %d, want %d", p.ID, got, want)
		}
	}
}

func TestHobGollumStillDrainsWithNothingToExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushCatalogGraveyardCard(me, "My Card", "Creature — Bear", "", 2, 2)
	life := hobB6Life(g, opp)
	hobCast(t, g, "Gollum, the Abandoned", "Legendary Creature — Halfling Horror", 2, 2)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Errorf("a target prompt opened with no card in an opponent's graveyard")
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(mine) {
		t.Error("my own graveyard card was exiled")
	}
	if got := hobB6Life(g, opp); got != life-2 {
		t.Errorf("the opponent's life %d -> %d, want -2", life, got)
	}
}

func TestHobGollumReturnsFromTheGraveyardForTwoAndASacrificeAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	gollum := pushCatalogGraveyardCard(me, "Gollum, the Abandoned", "Legendary Creature — Halfling Horror", hobOracle(t, "Gollum, the Abandoned"), 2, 2)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	rock := pushArtifactForTest(g, me.ID, "Mind Stone")
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
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(gollum) {
		t.Fatal("Gollum did not return to hand")
	}
	if _, ok := battlefieldCard(g, rock); ok {
		t.Error("the artifact was not sacrificed")
	}

	// A creature pays too, and not at instant speed.
	second := pushCatalogGraveyardCard(me, "Gollum, the Abandoned", "Legendary Creature — Halfling Horror", hobOracle(t, "Gollum, the Abandoned"), 2, 2)
	gollum = second
	advanceTo(t, g, game.StepBeginCombat)
	if err := activate(bear); err == nil {
		t.Error("activated at instant speed")
	}
	advanceTo(t, g, game.StepPostcombatMain)
	if err := activate(bear); err != nil {
		t.Fatalf("sacrificing a creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(second) {
		t.Error("Gollum did not return to hand the second time")
	}
}

// --- Mirkwood Elk ------------------------------------------------------

func TestHobMirkwoodElkEntersAndReturnsAnElfGainingLifeEqualToItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := pushCatalogGraveyardCard(me, "Elvish Warrior", "Creature — Elf Warrior", "", 3, 2)
	bear := pushCatalogGraveyardCard(me, "Bear", "Creature — Bear", "", 5, 5)
	theirs := pushCatalogGraveyardCard(opp, "Their Elf", "Creature — Elf Warrior", "", 4, 4)
	life := hobB6Life(g, me)
	hobCast(t, g, "Mirkwood Elk", "Creature — Elk", 6, 6)
	if latestPickTarget(g, me.ID) != nil {
		answerPickTarget(t, g, elf)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(elf) {
		t.Error("the Elf did not return to hand")
	}
	if !me.Graveyard.Contains(bear) || !opp.Graveyard.Contains(theirs) {
		t.Error("a card that is not my Elf left a graveyard")
	}
	if got := hobB6Life(g, me); got != life+3 {
		t.Errorf("life %d -> %d, want +3", life, got)
	}
}

func TestHobMirkwoodElkWithNoElfDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogGraveyardCard(me, "Bear", "Creature — Bear", "", 5, 5)
	life, hand := hobB6Life(g, me), me.Hand.Size()
	hobCast(t, g, "Mirkwood Elk", "Creature — Elk", 6, 6)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Error("a target prompt opened with no Elf card in the graveyard")
	}
	if got := hobB6Life(g, me); got != life {
		t.Errorf("life changed %d -> %d", life, got)
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d -> %d, want no card gained (the Elk was put there to be cast)", hand, got)
	}
}

func TestHobMirkwoodElkAttackingReturnsAnElfToo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elk := hobPush(t, g, me.ID, "Mirkwood Elk", "Creature — Elk", 6, 6)
	elf := pushCatalogGraveyardCard(me, "Elvish Warrior", "Creature — Elf Warrior", "", 2, 2)
	life := hobB6Life(g, me)
	declareAttack(t, g, opp.ID, elk)
	if latestPickTarget(g, me.ID) != nil {
		answerPickTarget(t, g, elf)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(elf) {
		t.Error("the Elf did not return to hand when the Elk attacked")
	}
	if got := hobB6Life(g, me); got != life+2 {
		t.Errorf("life %d -> %d, want +2", life, got)
	}
	if !hasEffectiveKeyword(t, g, elk, "trample") {
		t.Error("the Elk lacks trample")
	}
}

// --- Celeborn the Wise -------------------------------------------------

func TestHobCelebornScriesOnceWhenElvesAttackAndGrowsPerCardLookedAt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	celeborn := hobPush(t, g, me.ID, "Celeborn the Wise", "Legendary Creature — Elf Noble", 3, 3)
	elf := b12Push(g, me.ID, "Elvish Scout", "Creature — Elf Scout", "", 1, 1)
	declareAttack(t, g, opp.ID, celeborn, elf)
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("attacking with Elves did not scry")
	}
	if len(c.ScryCards) != 1 {
		t.Errorf("scry looked at %d cards, want 1", len(c.ScryCards))
	}
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) != nil {
		t.Error("two attacking Elves scried twice")
	}
	if p, tough := effectivePTOf(t, g, celeborn); p != 4 || tough != 4 {
		t.Errorf("Celeborn is %d/%d, want 4/4", p, tough)
	}
	if p, tough := effectivePTOf(t, g, elf); p != 1 || tough != 1 {
		t.Errorf("the other Elf is %d/%d, want 1/1", p, tough)
	}
}

func TestHobCelebornDoesNotScryForANonElfAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hobPush(t, g, me.ID, "Celeborn the Wise", "Legendary Creature — Elf Noble", 3, 3)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) != nil {
		t.Error("a Bear attacking made Celeborn scry")
	}
}

func TestHobCelebornGetsOneCounterPerCardAnyScryLookedAt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	celeborn := hobPush(t, g, me.ID, "Celeborn the Wise", "Legendary Creature — Elf Noble", 3, 3)
	hobB6AragornCast(t, g, me, "{U}", "U")
	passPriorityAroundTable(t, g)
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, celeborn); p != 5 || tough != 5 {
		t.Errorf("Celeborn is %d/%d after a scry 2, want 5/5", p, tough)
	}
	advanceToNextSeatsTurn(t, g)
	if p, tough := effectivePTOf(t, g, celeborn); p != 3 || tough != 3 {
		t.Errorf("Celeborn is %d/%d after the turn, want 3/3", p, tough)
	}
}

// --- Fateful Discovery -------------------------------------------------

func TestHobFatefulDiscoveryDrawsForEachArtifactYouControlThatEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hobPush(t, g, me.ID, "Fateful Discovery", "Enchantment", 0, 0)
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, TokenCard("1/1 colorless Servo artifact"), 2); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d -> %d, want +2 for two artifact tokens", hand, got)
	}
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, TokenCard("1/1 green Snake"), 1); err != nil {
			t.Fatal(err)
		}
		if err := g.CreateTokenForEffect(opp.ID, TokenCard("1/1 colorless Servo artifact"), 1); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand = %d, want a nonartifact and an opponent's artifact to draw nothing (%d)", got, hand+2)
	}
}

// --- Fog on the Barrow-Downs ------------------------------------------

func hobB6Fog(t *testing.T, g *game.Game, me *game.Player, host uuid.UUID) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	id := castFromHandForTest(t, g, me, "Fog on the Barrow-Downs", "Enchantment — Aura", "", hobOracle(t, "Fog on the Barrow-Downs"),
		game.CastSpellParams{Targets: cardRefs(host)})
	passPriorityAroundTable(t, g)
	return id
}

func TestHobFogOnTheBarrowDownsMakesTheCreatureASpirit(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := b12Push(g, opp.ID, "Hill Giant Archer", "Artifact Creature — Giant Archer", "", 3, 3)
	aura := hobB6Fog(t, g, me, host)
	got := gaLayered(t, g, host)
	if !got.HasSubtype("Spirit") || got.HasSubtype("Giant") || got.HasSubtype("Archer") {
		t.Errorf("the host is %q, want a Spirit and no other creature type", got.TypeLine)
	}
	if !got.IsCreature() || !got.IsArtifact() {
		t.Error("the host lost a card type")
	}
	if p, tough := effectivePTOf(t, g, host); p != 3 || tough != 3 {
		t.Errorf("the host is %d/%d, want unchanged 3/3", p, tough)
	}
	// Destroy the Aura and the creature is its old self again.
	hobDestroy(t, g, aura)
	back := gaLayered(t, g, host)
	if !back.HasSubtype("Giant") || back.HasSubtype("Spirit") {
		t.Errorf("after the Aura left the host is %q, want a Giant Archer", back.TypeLine)
	}
}

func TestHobFogOnTheBarrowDownsStopsTheCreatureAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := pushVanillaCreature(g, me.ID, "Host", 2, 2)
	free := pushVanillaCreature(g, me.ID, "Free", 2, 2)
	hobB6Fog(t, g, me, host)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(host, opp.ID); err == nil {
		t.Error("the enchanted creature attacked")
	}
	if err := g.DeclareAttacker(free, opp.ID); err != nil {
		t.Errorf("another creature could not attack: %v", err)
	}
}

func TestHobFogOnTheBarrowDownsStopsTheCreatureBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blocker := pushVanillaCreature(g, opp.ID, "Blocker", 2, 2)
	attacker := pushVanillaCreature(g, me.ID, "Attacker", 1, 1)
	hobB6Fog(t, g, me, blocker)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err == nil {
		t.Error("the enchanted creature blocked")
	}
}

// --- Thorin's Last Stand ----------------------------------------------

func hobB6ThorinCast(t *testing.T, g *game.Game, modes []int, targets ...uuid.UUID) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := handCardFull(me, "Thorin's Last Stand", "Instant", "", hobOracle(t, "Thorin's Last Stand"), nil)
	err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: modes, Targets: cardRefs(targets...)})
	if err == nil {
		passPriorityAroundTable(t, g)
	}
	return err
}

func TestHobThorinsLastStandPumpsYourCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushVanillaCreature(g, me.ID, "A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "B", 1, 1)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	if err := hobB6ThorinCast(t, g, []int{0}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	for id, want := range map[uuid.UUID][2]int{a: {4, 3}, b: {3, 2}, theirs: {2, 2}} {
		if p, tough := effectivePTOf(t, g, id); p != want[0] || tough != want[1] {
			t.Errorf("%s is %d/%d, want %d/%d", id, p, tough, want[0], want[1])
		}
	}
	late := pushVanillaCreature(g, me.ID, "Late", 2, 2)
	if p, tough := effectivePTOf(t, g, late); p != 2 || tough != 2 {
		t.Errorf("a creature that arrived later is %d/%d, want 2/2 (CR 611.2c)", p, tough)
	}
	advanceToNextSeatsTurn(t, g)
	if p, tough := effectivePTOf(t, g, a); p != 2 || tough != 2 {
		t.Errorf("A is %d/%d after the turn, want 2/2", p, tough)
	}
}

func TestHobThorinsLastStandDestroysAnArtifactOrEnchantmentAndGains2(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
	}{{"artifact", "Artifact"}, {"enchantment", "Enchantment"}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			target := hobB3Perm(g, opp.ID, game.Card{Name: "Target", TypeLine: tc.typeLine})
			life := hobB6Life(g, me)
			if err := hobB6ThorinCast(t, g, []int{1}, target); err != nil {
				t.Fatalf("cast: %v", err)
			}
			if _, ok := battlefieldCard(g, target); ok {
				t.Errorf("the %s was not destroyed", tc.name)
			}
			if got := hobB6Life(g, me); got != life+2 {
				t.Errorf("life %d -> %d, want +2", life, got)
			}
		})
	}
}

func TestHobThorinsLastStandCantTargetACreatureOrLand(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	land := hobB3Perm(g, opp.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	if err := hobB6ThorinCast(t, g, []int{1}, bear); err == nil {
		t.Error("targeted a creature")
	}
	if err := hobB6ThorinCast(t, g, []int{1}, land); err == nil {
		t.Error("targeted a land")
	}
}
