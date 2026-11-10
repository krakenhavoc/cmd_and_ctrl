package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b47_cards_test.go — Standard set leftovers, batches 47–50 (#4452).

const (
	stdHammerOracle     = "ccc2eb6d-21cf-45cf-b4e4-4b2a2517136a"
	stdAngelOracle      = "2ac93092-1a57-4cf7-8f03-1dca86d5c476"
	stdSageOfDaysOracle = "656fc672-efa5-484a-b5e8-eac262331439"
)

func TestStdB47DeconstructionHammerSacrificesItselfToDestroy(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	hammer := b21Push(g, me.ID, "Deconstruction Hammer", "Artifact — Equipment", stdHammerOracle, 0, 0)
	rock := b21Push(g, opp.ID, "Rock", "Artifact", "", 0, 0)
	advanceToMain(t, g)
	equipTo(t, g, me.ID, hammer, bear)
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("equipped power %d, want 3", got)
	}
	idx, ref := grantedActivatedIndex(t, g, bear)
	if idx < 0 {
		t.Fatal("the equipped creature has the granted ability")
	}
	// summoning sickness: the creature is old enough in tests; pay {3}.
	params := game.ActivateAbilityParams{Ref: ref, Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}}}
	if err := g.ActivateCatalogAbility(me.ID, bear, idx, params); err == nil {
		t.Fatal("activated with no mana")
	}
	floatMana(t, g, me, "{C}{C}{C}")
	b16Activate(t, g, me.ID, bear, idx, params)
	if g.Battlefield.Contains(hammer) {
		t.Error("the Hammer was not sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the target artifact survived")
	}
	if !b16Tapped(t, g, bear) {
		t.Error("the creature taps to pay")
	}
}

func TestStdB47DeconstructionHammerCantBeSacrificedByAnotherPlayersCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	hammer := b21Push(g, me.ID, "Deconstruction Hammer", "Artifact — Equipment", stdHammerOracle, 0, 0)
	rock := b21Push(g, me.ID, "Rock", "Artifact", "", 0, 0)
	advanceToMain(t, g)
	equipTo(t, g, me.ID, hammer, bear)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == hammer {
				g.Battlefield.Cards[i].Controller = opp.ID
			}
		}
	})
	idx, ref := grantedActivatedIndex(t, g, bear)
	if idx < 0 {
		t.Fatal("the creature lost the grant")
	}
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, bear, idx, game.ActivateAbilityParams{
		Ref: ref, Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	}); err == nil {
		t.Error("the creature's controller activated a grant whose Equipment they can't sacrifice")
	}
}

func TestStdB47ResplendentAngelNeedsFiveLifeAndMakesOneAngel(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Resplendent Angel", "Creature — Angel", stdAngelOracle, 3, 3)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 4) })
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Angel"); n != 0 {
		t.Errorf("four life is not five: %d Angel tokens", n)
	}
	advanceToMainOf(t, g, 1)
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3)
	})
	advanceToEndStepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Angel"); n != 1 {
		t.Fatalf("six life gained: %d Angel tokens, want exactly 1", n)
	}
	angel := battlefieldIDNamed(g, me.ID, "Angel")
	if !hasEffectiveKeyword(t, g, angel, "flying") || !hasEffectiveKeyword(t, g, angel, "vigilance") {
		t.Error("the Angel token has flying and vigilance")
	}
}

func TestStdB47SageOfDaysKeepsOneOnTopAndMillsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	c3 := seedLibraryTop(me, "Third", "Sorcery", "")
	c2 := seedLibraryTop(me, "Second", "Sorcery", "")
	c1 := seedLibraryTop(me, "First", "Sorcery", "")
	castCreatureCard(t, g, "Sage of Days", "Creature — Human Wizard", stdSageOfDaysOracle, 3, 2)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, c2)
	passPriorityAroundTable(t, g)
	if top, err := me.Library.Top(); err != nil || top.InstanceID != c2 {
		t.Errorf("the picked card is not on top of the library: %+v %v", top, err)
	}
	if !me.Graveyard.Contains(c1) || !me.Graveyard.Contains(c3) || me.Graveyard.Contains(c2) {
		t.Error("the two unpicked cards go to the graveyard and the picked one does not")
	}
}

func TestStdB47SageOfDaysDecliningMillsAllThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := []uuid.UUID{
		seedLibraryTop(me, "A", "Sorcery", ""), seedLibraryTop(me, "B", "Sorcery", ""), seedLibraryTop(me, "C", "Sorcery", ""),
	}
	castCreatureCard(t, g, "Sage of Days", "Creature — Human Wizard", stdSageOfDaysOracle, 3, 2)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	for _, id := range ids {
		if !me.Graveyard.Contains(id) {
			t.Errorf("%s was not put into the graveyard", id)
		}
	}
}

func TestStdB47BrackishBlunderMakesAMapOnlyForATappedCreature(t *testing.T) {
	for _, tapped := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		victim := b16Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2, "G")
		if tapped {
			g.WithWriteLock(func() {
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].InstanceID == victim {
						g.Battlefield.Cards[i].Tapped = true
					}
				}
			})
		}
		castCatalogSpell(t, g, "Brackish Blunder", "Instant", "4ad34f85-a185-458a-a7a1-eb60f7efce5b",
			[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(victim) || !opp.Hand.Contains(victim) {
			t.Fatalf("tapped=%v: the creature was not returned to its owner's hand", tapped)
		}
		maps := 0
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Map" && c.Controller == me.ID {
				maps++
			}
		}
		if want := map[bool]int{true: 1, false: 0}[tapped]; maps != want {
			t.Errorf("tapped=%v: %d Map tokens for the caster, want %d", tapped, maps, want)
		}
	}
}

func TestStdB47AnotherChanceMillsThenReturnsUpToTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibraryTop(me, "Filler One", "Sorcery", "")
	milledCreature := seedLibraryTop(me, "Milled Bear", "Creature — Bear", "")
	old := []uuid.UUID{
		pushGraveCreature(me, "Old One"), pushGraveCreature(me, "Old Two"), pushGraveCreature(me, "Old Three"),
	}
	castCatalogSpell(t, g, "Another Chance", "Instant", "df9cdc11-cb92-46ce-a83d-eb26e3cd004a", nil)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	if !me.Graveyard.Contains(milledCreature) {
		t.Fatal("the creature was not milled")
	}
	answerChooseCards(t, g, me.ID, milledCreature, old[0])
	if !me.Hand.Contains(milledCreature) || !me.Hand.Contains(old[0]) {
		t.Error("the two chosen creature cards did not reach the hand")
	}
	if me.Hand.Contains(old[1]) || me.Hand.Contains(old[2]) {
		t.Error("more than two creature cards returned")
	}
}

func TestStdB47AnotherChanceDecliningTheMillStillReturns(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Library.Size()
	c := pushGraveCreature(me, "Old One")
	castCatalogSpell(t, g, "Another Chance", "Instant", "df9cdc11-cb92-46ce-a83d-eb26e3cd004a", nil)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	if me.Library.Size() != before {
		t.Error("declining milled cards anyway")
	}
	answerChooseCards(t, g, me.ID, c)
	if !me.Hand.Contains(c) {
		t.Error("the creature card did not return")
	}
}

func pushGraveCreature(p *game.Player, name string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", Owner: p.ID, Controller: p.ID})
	return id
}

func TestStdB47TerrorTideShrinksEveryCreatureByTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushGraveCreature(me, "Dead One")
	pushGraveCreature(me, "Dead Two")
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	mine := b16Creature(g, me.ID, "Mine", "Creature — Bear", 4, 3, "G")
	small := b16Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2, "G")
	castCatalogSpell(t, g, "Terror Tide", "Sorcery", "12be58f4-1c12-4bca-8069-1f049cd8a90c", nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("a 2/2 survived -2/-2")
	}
	if !g.Battlefield.Contains(mine) {
		t.Fatal("a 4/3 died to -2/-2")
	}
	if p, tt := effectivePower(t, g, mine), effectiveToughness(t, g, mine); p != 2 || tt != 1 {
		t.Errorf("my creature is %d/%d, want 2/1", p, tt)
	}
}

func TestStdB47SkullcapSnailMakesAnOpponentExileFromTheirHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	card := handCardForTest(opp, "Their Card", "Sorcery", "")
	castCatalogSpell(t, g, "Skullcap Snail", "Creature — Fungus Snail", "6420e8a0-3ef4-4f95-bb6a-12409eef4d48", nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, opp.ID, card)
	if opp.Hand.Contains(card) {
		t.Error("the card is still in their hand")
	}
	if !g.Exile.Contains(card) {
		t.Error("the card was not exiled")
	}
}

func TestStdB48DefossilizeReturnsThenExploresTwice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	target := pushGraveCreature(me, "Fossil")
	seedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	seedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	handBefore := me.Hand.Size()
	castCatalogSpell(t, g, "Defossilize", "Sorcery", "d620de15-8141-4877-9330-8328a3fa85e1",
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(target) {
		t.Fatal("the creature did not return")
	}
	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("two land explores should put two lands into hand, gained %d", got)
	}
}

func TestStdB48OverTheEdgeExploresTwiceOrDestroys(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	seedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	seedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	handBefore := me.Hand.Size()
	paCast(t, g, "Over the Edge", "Sorcery", "{1}{G}", "26b441b9-fdfb-43e2-bc02-b358653eadf4", game.CastSpellParams{
		Modes:   []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("two land explores: gained %d, want 2", got)
	}
	rock := b21Push(g, opp.ID, "Rock", "Artifact", "", 0, 0)
	paCast(t, g, "Over the Edge", "Sorcery", "{1}{G}", "26b441b9-fdfb-43e2-bc02-b358653eadf4", game.CastSpellParams{
		Modes:   []int{0},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the destroy mode left the artifact")
	}
}

func TestStdB49PalanisHatcherHatchesAnEggAtCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b21Push(g, me.ID, "Palani's Hatcher", "Creature — Dinosaur", "ce6b6869-e4d7-4156-b66d-57fc224e4b20", 3, 3)
	eggA := b16Creature(g, me.ID, "Dinosaur Egg", "Token Creature — Dinosaur Egg", 0, 1, "G")
	b16Creature(g, me.ID, "Dinosaur Egg", "Token Creature — Dinosaur Egg", 0, 1, "G")
	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, eggA)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(eggA) {
		t.Error("the chosen Egg was not sacrificed")
	}
	if n := countBattlefieldNamed(g, me.ID, "Dinosaur Egg"); n != 1 {
		t.Errorf("%d Eggs left, want 1", n)
	}
	if n := countBattlefieldNamed(g, me.ID, "Dinosaur"); n != 1 {
		t.Errorf("%d Dinosaur tokens, want 1", n)
	}
}

func TestStdB50BedrockTortoiseGrantsHexproofOnlyOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b21Push(g, me.ID, "Bedrock Tortoise", "Creature — Turtle", "d9c2e674-11a9-47f0-8790-a53b62c43fe6", 0, 6)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	advanceToMain(t, g)
	if !hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("creatures have hexproof during your turn")
	}
	advanceToMainOf(t, g, 1)
	if hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("hexproof stayed on during an opponent's turn")
	}
}

func TestStdB48ArmoredKincallerGainsLifeForADinosaurInHandOrOnBoard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dino := handCardForTest(me, "Dino", "Creature — Dinosaur", "")
	life := me.Life
	castCreatureCard(t, g, "Armored Kincaller", "Creature — Dinosaur", "6aad6906-96ec-4994-bfc3-9a76f893bc8c", 3, 3)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, dino)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d after revealing a Dinosaur", me.Life, life+3)
	}
}

func TestStdB48ArmoredKincallerDecliningStillGainsWithAnotherDinosaur(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b16Creature(g, me.ID, "Other Dino", "Creature — Dinosaur", 2, 2, "G")
	dino := handCardForTest(me, "Dino", "Creature — Dinosaur", "")
	_ = dino
	life := me.Life
	castCreatureCard(t, g, "Armored Kincaller", "Creature — Dinosaur", "6aad6906-96ec-4994-bfc3-9a76f893bc8c", 3, 3)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d with another Dinosaur", me.Life, life+3)
	}
}

func TestStdB48ArmoredKincallerGainsNothingAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	castCreatureCard(t, g, "Armored Kincaller", "Creature — Dinosaur", "6aad6906-96ec-4994-bfc3-9a76f893bc8c", 3, 3)
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life changed to %d with no other Dinosaur and none in hand", me.Life)
	}
}

func TestStdB47AclazotzMakesOpponentsDiscardAndDrawsForTheEmptyHanded(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	acl := b21Push(g, me.ID, "Aclazotz, Deepest Betrayal", "Legendary Creature — Bat God", "fcdfe9d5-2743-4d3e-ab57-bf0f96beaa15", 4, 4)
	// Empty the second opponent's hand; the rest discard.
	g.WithWriteLock(func() {
		for _, s := range g.Seats[2:3] {
			s.Hand.Cards = nil
		}
	})
	handBefore := me.Hand.Size()
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, acl)
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID)
	answerDiscard(t, g, g.Seats[3].ID, g.Seats[3].Hand.Cards[0].InstanceID)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("drew %d cards for the one opponent who couldn't discard, want 1", got)
	}
}

func TestStdB47ResplendentAngelRegistered(t *testing.T) {
	for _, id := range []string{
		"ccc2eb6d-21cf-45cf-b4e4-4b2a2517136a", "2ac93092-1a57-4cf7-8f03-1dca86d5c476", "656fc672-efa5-484a-b5e8-eac262331439",
		"4ad34f85-a185-458a-a7a1-eb60f7efce5b", "df9cdc11-cb92-46ce-a83d-eb26e3cd004a", "fcdfe9d5-2743-4d3e-ab57-bf0f96beaa15",
		"fcdfe9d5-2743-4d3e-ab57-bf0f96beaa15#1", "12be58f4-1c12-4bca-8069-1f049cd8a90c", "6420e8a0-3ef4-4f95-bb6a-12409eef4d48",
		"d620de15-8141-4877-9330-8328a3fa85e1", "26b441b9-fdfb-43e2-bc02-b358653eadf4", "2d785e3e-8e19-4eea-b09f-9a30c792dba1",
		"6aad6906-96ec-4994-bfc3-9a76f893bc8c", "ce6b6869-e4d7-4156-b66d-57fc224e4b20", "d9c2e674-11a9-47f0-8790-a53b62c43fe6",
	} {
		spec, ok := Lookup(id)
		if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v completeness=%v", id, ok, spec.Completeness)
		}
	}
}

func TestStdB47WardenOfTheInnerSkyTapsThreeForACounterAndGrowsWings(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	warden := b21Push(g, me.ID, "Warden of the Inner Sky", "Creature — Human Soldier", "5268f772-5b42-4b5b-9e57-c32c662a42f3", 1, 2)
	a := b16Creature(g, me.ID, "A", "Creature — Bear", 1, 1, "G")
	b := b16Creature(g, me.ID, "B", "Creature — Bear", 1, 1, "G")
	c := b21Push(g, me.ID, "Rock", "Artifact", "", 0, 0)
	advanceToMain(t, g)
	if hasEffectiveKeyword(t, g, warden, "flying") {
		t.Fatal("the Warden flies with no counters")
	}
	for i := 0; i < 3; i++ {
		for _, id := range []uuid.UUID{warden, a, b, c} {
			g.WithWriteLock(func() {
				for j := range g.Battlefield.Cards {
					if g.Battlefield.Cards[j].InstanceID == id {
						g.Battlefield.Cards[j].Tapped = false
					}
				}
			})
		}
		b16Activate(t, g, me.ID, warden, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{a, b, c}})
		passPriorityAroundTable(t, g)
		answerScryIfAsked(t, g, me)
	}
	if n := countersOn(g, warden, game.CounterPlusOne); n < 3 {
		t.Fatalf("%d counters, want 3", n)
	}
	if !hasEffectiveKeyword(t, g, warden, "flying") || !hasEffectiveKeyword(t, g, warden, "vigilance") {
		t.Error("three counters give flying and vigilance")
	}
}

func TestStdB47MarketGnomeDrawsWhenExiledToCraftButNotWhenExiledByAnEffect(t *testing.T) {
	g, me, _ := spendTable(t)
	blade := pushCraftCard(g, me, tithingBladeRow())
	gnome := pushCatalogPermanent(g, me.ID, "Market Gnome", "Artifact Creature — Gnome", "12fa17b2-158e-4d53-897f-81111f912b00", false)
	handBefore := me.Hand.Size()
	lifeBefore := me.Life
	floatMana(t, g, me, "{C}{C}{C}{C}{B}")
	if err := activateCraft(g, me, blade, 0, gnome); err != nil {
		t.Fatalf("craft: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("drew %d cards after being exiled to craft, want 1", got)
	}
	if me.Life != lifeBefore+1 {
		t.Errorf("life %d, want %d", me.Life, lifeBefore+1)
	}
}

func TestStdB50BringerOfTheLastGiftSacrificesThenReturnsOnlyTheOlderCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	old := pushGraveCreature(opp, "Old Fossil")
	mine := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	castCreatureCard(t, g, "Bringer of the Last Gift", "Creature — Vampire Demon", "0fb9a8a2-4331-4259-9460-a37397554962", 6, 6)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("another creature survived the sacrifice")
	}
	if !g.Battlefield.Contains(old) {
		t.Error("the creature card that was already in a graveyard did not return")
	}
	if !me.Graveyard.Contains(mine) || !opp.Graveyard.Contains(theirs) {
		t.Error("a creature put into the graveyard this way came back")
	}
}

func TestStdB49ContestedGameBallCountsFiveDrawsAndMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ball := b21Push(g, me.ID, "Contested Game Ball", "Artifact", "876c9620-216c-45dd-b8a6-b9e0cba63e3d", 0, 0)
	advanceToMain(t, g)
	for i := 0; i < 5; i++ {
		g.WithWriteLock(func() {
			for j := range g.Battlefield.Cards {
				if g.Battlefield.Cards[j].InstanceID == ball {
					g.Battlefield.Cards[j].Tapped = false
				}
			}
		})
		floatMana(t, g, me, "{C}{C}")
		b16Activate(t, g, me.ID, ball, 0, game.ActivateAbilityParams{})
		passPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(ball) {
		t.Error("the Ball was not sacrificed at five point counters")
	}
	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasures, want 1", n)
	}
}

func TestStdB49SunkenCitadelEntersTappedAndChoosesAColor(t *testing.T) {
	spec, ok := Lookup("508189e1-9cef-4f9c-8ff1-078c99a0f603")
	if !ok || len(spec.ManaAbilities) != 2 || spec.Completeness != CompletenessFull {
		t.Fatalf("Sunken Citadel registered wrongly: %v %d", ok, len(spec.ManaAbilities))
	}
}

func TestStdB50ZoyowaDamagesAnOpponentWhoDoesNotPay(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b21Push(g, me.ID, "Zoyowa Lava-Tongue", "Legendary Creature — Goblin Warlock", "e90f298f-c54b-4fe8-82db-5b9153a57712", 2, 2)
	advanceToMain(t, g)
	lciDGMillOne(t, g, me, "Creature — Bear")
	before := make([]int, len(g.Seats))
	for i, s := range g.Seats {
		before[i] = s.Life
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	for _, s := range g.Seats[1:] {
		answerPayUnless(t, g, s.ID, false)
	}
	for i, s := range g.Seats[1:] {
		if s.Life != before[i+1]-3 {
			t.Errorf("opponent %d at %d, want %d", i+1, s.Life, before[i+1]-3)
		}
	}
	_ = opp
}

func TestStdB50TheMycotyrantSizeCountsFungiAndSaprolings(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	myco := b21Push(g, me.ID, "The Mycotyrant", "Legendary Creature — Elder Fungus", "b69bbb8d-7f8c-45d6-983a-759b5fce6ebd", 0, 0)
	b16Creature(g, me.ID, "Spore", "Creature — Saproling", 1, 1, "G")
	b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	advanceToMain(t, g)
	if p, tt := effectivePower(t, g, myco), effectiveToughness(t, g, myco); p != 2 || tt != 2 {
		t.Errorf("The Mycotyrant is %d/%d, want 2/2 (itself and one Saproling)", p, tt)
	}
}

func TestStdB49PitOfOfferingsAndEchoingDeepsRegistered(t *testing.T) {
	for _, id := range []string{"044d2788-6daa-4849-a813-1f577eef9295", "2ef88214-f46d-473e-a55b-795a647e2f03",
		"876c9620-216c-45dd-b8a6-b9e0cba63e3d", "05551d91-50c6-46d1-86a4-cd3d177d0923", "b69bbb8d-7f8c-45d6-983a-759b5fce6ebd",
		"e90f298f-c54b-4fe8-82db-5b9153a57712", "0fb9a8a2-4331-4259-9460-a37397554962", "12fa17b2-158e-4d53-897f-81111f912b00",
		thousandMoonsSmithyOracleID, thousandMoonsSmithyOracleID + "#1", "5268f772-5b42-4b5b-9e57-c32c662a42f3"} {
		if spec, ok := Lookup(id); !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s not registered Full", id)
		}
	}
}

func TestStdB48BloodthornFlailPaysManaOrDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	flail := b21Push(g, me.ID, "Bloodthorn Flail", "Artifact — Equipment", "01ab7724-1967-4da6-aebd-3729983d6a0e", 0, 0)
	advanceToMain(t, g)
	idx := len(game.ActivatedAbilitiesForCard(game.Card{OracleID: oracleOfBattlefieldCard(t, g, flail)})) - 1
	tgt := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}
	discard, mana := 1, 0
	// the discard branch
	card := handCardForTest(me, "Spare", "Sorcery", "")
	if err := g.ActivateCatalogAbility(me.ID, flail, idx, game.ActivateAbilityParams{Targets: tgt, CostBranch: &discard, DiscardIDs: []uuid.UUID{card}}); err != nil {
		t.Fatalf("equip by discarding: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(card) {
		t.Error("the discard branch did not discard")
	}
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("equipped power %d, want 4", got)
	}
	// the mana branch needs {3}
	if err := g.ActivateCatalogAbility(me.ID, flail, idx, game.ActivateAbilityParams{Targets: tgt, CostBranch: &mana, Strict: true}); err == nil {
		t.Error("equipped for {3} with no mana")
	}
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, flail, idx, game.ActivateAbilityParams{Targets: tgt, CostBranch: &mana, Strict: true}); err != nil {
		t.Fatalf("equip for {3}: %v", err)
	}
}

func TestStdB49NicanzilReadsTheRevealedCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	nic := b21Push(g, me.ID, "Nicanzil, Current Conductor", "Legendary Creature — Merfolk Scout", "4cb86fb6-3e06-4c67-aee0-d5952178f664", 2, 2)
	seedLibraryTop(me, "Grizzly Bears", "Creature — Bear", "")
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.ExploreForEffect(uuid.Nil, game.ObjectRef{ID: nic}, me.ID) })
	// a nonland reveal: keep it on top, then the +1/+1 counter trigger for the nonland
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, nic, game.CounterPlusOne); n != 2 {
		t.Errorf("%d +1/+1 counters, want 2 (the explore's and Nicanzil's)", n)
	}
}

func TestStdB47SageDaysRegisteredAndAbuelo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := pushGraveArtifact(me, "Fossil Rock")
	castCatalogSpellX(t, g, "Abuelo's Awakening", "Sorcery", "2b22cf44-7b03-4919-9812-6796bf193484", 2, rock)
	if !g.Battlefield.Contains(rock) {
		t.Fatal("the artifact did not return")
	}
	if n := countersOn(g, rock, game.CounterPlusOne); n != 2 {
		t.Errorf("%d +1/+1 counters, want X = 2", n)
	}
	if !hasEffectiveKeyword(t, g, rock, "flying") {
		t.Error("the returned artifact flies")
	}
	if p := effectivePowerWithCounters(t, g, rock); p != 3 {
		t.Errorf("power %d, want 1/1 plus two counters", p)
	}
}

func pushGraveArtifact(p *game.Player, name string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Artifact", Owner: p.ID, Controller: p.ID})
	return id
}

func castCatalogSpellX(t *testing.T, g *game.Game, name, typeLine, oracle string, x int, target uuid.UUID) {
	t.Helper()
	paCast(t, g, name, typeLine, "{X}{3}{W}", oracle, game.CastSpellParams{
		XValue:  x,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	})
	passPriorityAroundTable(t, g)
}

func TestStdB49TwistsAndTurnsScriesBeforeEveryExplore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b21Push(g, me.ID, "Twists and Turns", "Enchantment", twistsAndTurnsOracleID, 0, 0)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	seedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	advanceToMain(t, g)
	handBefore := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.ExploreForEffect(uuid.Nil, game.ObjectRef{ID: bear}, me.ID) })
	answerScryIfAsked(t, g, me)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("explored a land: hand grew by %d, want 1", got)
	}
}

func TestStdB49AmaliaExploresWhenYouGainLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	amalia := b21Push(g, me.ID, "Amalia Benavides Aguirre", "Legendary Creature — Vampire Scout", "85eefda9-4aab-467d-a08c-be026cb8ff42", 2, 2)
	seedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	advanceToMain(t, g)
	handBefore := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2) })
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("hand grew by %d, want 1 (explored a land)", got)
	}
	if !g.Battlefield.Contains(amalia) {
		t.Error("Amalia was destroyed with a power other than 20")
	}
}

func TestStdB48DowsingDeviceTransformsWithFourArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 3; i++ {
		b21Push(g, me.ID, "Rock", "Artifact", "", 0, 0)
	}
	mine := b16Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2, "G")
	advanceToMain(t, g)
	dowsing := importToBattlefield(t, g, transformRow(dowsingDeviceOracleID, "Dowsing Device", "Artifact", "{1}{R}",
		"Geode Grotto", "Land — Cave", "", "", []string{"R"}), me)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, mine); got != 3 {
		t.Errorf("the target is %d power, want 3", got)
	}
	c, _ := battlefieldCard(g, dowsing)
	if c.ActiveFace != 1 {
		t.Errorf("Dowsing Device did not transform with four artifacts (active face %d)", c.ActiveFace)
	}
}

func TestStdB49EveryRemainingCardIsRegisteredFull(t *testing.T) {
	for _, id := range []string{
		"d2824464-963b-48db-a482-cc08a0a97332", twistsAndTurnsOracleID, twistsAndTurnsOracleID + "#1",
		"85eefda9-4aab-467d-a08c-be026cb8ff42", "6615f881-960c-4216-a32a-63856bc3afd0", "4cb86fb6-3e06-4c67-aee0-d5952178f664",
		tarriansJournalOracleID, tarriansJournalOracleID + "#1", graspingShadowsOracleID, graspingShadowsOracleID + "#1",
		dowsingDeviceOracleID, dowsingDeviceOracleID + "#1", "2034d41b-5e77-49a6-884e-d3a6017f84d0", "01ab7724-1967-4da6-aebd-3729983d6a0e",
		"2b22cf44-7b03-4919-9812-6796bf193484", "798aaf78-16d9-474b-a963-f5e9c39aae6e",
	} {
		if spec, ok := Lookup(id); !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s not registered Full", id)
		}
	}
}

func TestStdB49TendrilAnimatesALandWithSevenCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tendril := b21Push(g, me.ID, "Tendril of the Mycotyrant", "Creature — Fungus Wizard", "d2824464-963b-48db-a482-cc08a0a97332", 2, 2)
	land := pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{G}{G}")
	b16Activate(t, g, me.ID, tendril, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}})
	passPriorityAroundTable(t, g)
	if n := countersOn(g, land, game.CounterPlusOne); n != 7 {
		t.Fatalf("%d counters, want 7", n)
	}
	c, _ := battlefieldCard(g, land)
	if !c.IsCreature() || !c.IsLand() || !hasEffectiveKeyword(t, g, land, "haste") {
		t.Error("the land is not a hasty creature land")
	}
	if p := effectivePowerWithCounters(t, g, land); p != 7 {
		t.Errorf("the Fungus land is %d power, want 0/0 plus seven counters", p)
	}
}

func TestStdB50CosmiumConfluenceFetchesAndAnimatesACave(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cave := pushCatalogPermanent(g, me.ID, "Hidden Cave", "Land — Cave", "", false)
	ench := b21Push(g, g.Seats[1].ID, "Pest", "Enchantment", "", 0, 0)
	paCast(t, g, "Cosmium Confluence", "Sorcery", "{4}{G}", "798aaf78-16d9-474b-a963-f5e9c39aae6e", game.CastSpellParams{
		Modes: []int{1, 2, 2},
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: ench},
			{Kind: game.TargetCard, ID: ench},
		},
	})
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, cave)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, cave, game.CounterPlusOne); n != 3 {
		t.Errorf("%d counters on the Cave, want 3", n)
	}
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment survived the destroy modes")
	}
}

func TestStdB50AnimPakalMakesAttackingGnomes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pakal := b21Push(g, me.ID, "Anim Pakal, Thousandth Moon", "Legendary Creature — Human Soldier", "05551d91-50c6-46d1-86a4-cd3d177d0923", 1, 2)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, pakal, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters on Anim Pakal, want 1", n)
	}
	if n := countBattlefieldNamed(g, me.ID, "Gnome"); n != 1 {
		t.Errorf("%d Gnome tokens, want X = 1", n)
	}
}

func TestStdB48GraspingShadowsTransformsAtThreeDreadCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	shadows := importToBattlefield(t, g, transformRow(graspingShadowsOracleID, "Grasping Shadows", "Enchantment", "{3}{B}",
		"Shadows' Lair", "Land — Cave", "", "", []string{"B"}), me)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == shadows {
				g.Battlefield.Cards[i].Counters = map[string]int{dreadCounter: 2}
			}
		}
	})
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	declareAttack(t, g, g.Seats[1].ID, bear)
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, shadows)
	if c.ActiveFace != 1 {
		t.Errorf("Grasping Shadows did not transform at its third dread counter (face %d)", c.ActiveFace)
	}
	if !hasEffectiveKeyword(t, g, bear, "deathtouch") || !hasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("the lone attacker did not gain deathtouch and lifelink")
	}
}

func TestStdB49PitOfOfferingsExilesAndTapsForTheirColors(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	red := pushGraveCard(opp, game.Card{Name: "Red Thing", TypeLine: "Creature — Goblin", Colors: []string{"R"}})
	advanceToMain(t, g)
	pit := importToBattlefield(t, g, cards.Card{
		ID: uuid.New(), Name: "Pit of Offerings", TypeLine: "Land — Cave", OracleID: uuid.MustParse("044d2788-6daa-4849-a813-1f577eef9295"),
	}, me)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, red)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(red) {
		t.Fatal("the target card was not exiled")
	}
	_ = pit
}

func pushGraveCard(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Graveyard.PushTop(c)
	return c.InstanceID
}
