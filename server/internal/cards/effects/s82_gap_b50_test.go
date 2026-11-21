package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapWombatOracle      = "fd077d1a-f395-41c2-b1df-dd5e2f723abf"
	gapOozeOracle        = "4feeb493-8db5-41d3-bd27-486052e61f28"
	gapImskirOracle      = "2494101a-2c6b-45d8-95fe-e62bf9a25130"
	gapMunitionsOracle   = "a0b16b11-69d9-4e94-aae3-298aa44ce3d4"
	gapKaaliaOracle      = "cb8d80c9-ed58-4f2d-aa8c-c383370c7f1a"
	gapShadowspinnerOrcl = "5fcbb5d3-71d1-4946-b806-ba6376338eba"
	gapFrogOracle        = "e157ad1e-be35-47f3-92db-01799f8fb6a5"
	gapSnackerOracle     = "f255492e-4e12-48b4-9c19-8a2a2780a1b7"
	gapOnduOracle        = "dc69a130-967a-4d9f-992f-db7fad2fcd6d"
	gapShilgengarOracle  = "71ac4d8c-5b58-4ea8-996b-f0289717c4bb"
)

// gapB50Float puts n black mana in the pool: generic costs take any colour.
func gapB50Float(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		floatMana(t, g, p, "{B}")
	}
}

func gapB50Stock(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() {
		for i := 0; i < n; i++ {
			p.Library.PushTop(game.NewCard("Filler", p.ID))
		}
	})
}

func gapB50HandCard(p *game.Player, name string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = "Instant"
	p.Hand.PushTop(c)
	return c.InstanceID
}

func gapCountersOf(g *game.Game, id uuid.UUID) int {
	c, _ := battlefieldCard(g, id)
	return c.Counters[game.CounterPlusOne]
}

func gapAddCounter(t *testing.T, g *game.Game, id uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}).Apply(NewContext(g, nil)); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
}

func gapPick(t *testing.T, g *game.Game, who uuid.UUID, target game.TargetRef) {
	t.Helper()
	pick := latestChoiceOfKindFor(g, game.PendingChoicePickTarget, who)
	if pick == nil {
		return
	}
	if err := g.ResolvePickTarget(pick.ID, who, target); err != nil {
		t.Fatal(err)
	}
}

func gapActivateStrict(t *testing.T, g *game.Game, who, id uuid.UUID, idx int, p game.ActivateAbilityParams) error {
	t.Helper()
	p.Strict = true
	return g.ActivateCatalogAbility(who, id, idx, p)
}

func TestGapB50WombatAddsOneCounterToEachPermanentOncePerTurn(t *testing.T) {
	const eelOracle = "75132177-0881-4158-9d65-73f5ba9d4402" // Skitter Eel: {2}{U}: Adapt 2
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wombat := pushCatalogPermanent(g, me.ID, "Cursed Wombat", "Creature — Nightmare Wombat", gapWombatOracle, false)
	mine := pushCatalogPermanent(g, me.ID, "Skitter Eel", "Creature — Fish", eelOracle, false)
	theirs := pushCatalogPermanent(g, opp.ID, "Skitter Eel", "Creature — Fish", eelOracle, false)
	for _, tc := range []struct {
		who  *game.Player
		eel  uuid.UUID
		want int
	}{{me, mine, 3}, {opp, theirs, 2}} {
		for i := 0; i < 3; i++ {
			floatMana(t, g, tc.who, "{U}")
		}
		if err := gapActivateStrict(t, g, tc.who.ID, tc.eel, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		if got := gapCountersOf(g, tc.eel); got != tc.want {
			t.Errorf("%v's Eel has %d counters, want %d", tc.who == me, got, tc.want)
		}
	}
	gapB50Float(t, g, me, 3)
	floatMana(t, g, me, "{G}")
	if err := gapActivateStrict(t, g, me.ID, wombat, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if got := gapCountersOf(g, wombat); got != 3 {
		t.Errorf("wombat counters = %d, want 3 (adapt 2 and one additional)", got)
	}
}

func TestGapB50OozePutsACounterOnAModifiedCreatureWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ooze := pushCatalogPermanent(g, me.ID, "Expanding Ooze", "Creature — Ooze", gapOozeOracle, false)
	plain := pushVanillaCreature(g, me.ID, "Plain", 2, 2)
	modified := pushVanillaCreature(g, me.ID, "Modified", 2, 2)
	gapAddCounter(t, g, modified, 1)
	declareAttack(t, g, opp.ID, ooze)
	gapPick(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: modified})
	passPriorityAroundTable(t, g)
	if got := gapCountersOf(g, modified); got != 2 {
		t.Errorf("modified creature counters = %d, want 2", got)
	}
	if got := gapCountersOf(g, plain); got != 0 {
		t.Errorf("an unmodified creature got a counter")
	}
}

func TestGapB50OozeCannotTargetAnUnmodifiedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ooze := pushCatalogPermanent(g, me.ID, "Expanding Ooze", "Creature — Ooze", gapOozeOracle, false)
	plain := pushVanillaCreature(g, me.ID, "Plain", 2, 2)
	declareAttack(t, g, opp.ID, ooze)
	gapPick(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: plain})
	passPriorityAroundTable(t, g)
	if got := gapCountersOf(g, plain); got != 0 {
		t.Errorf("an unmodified creature was a legal target: counters %d", got)
	}
}

func TestGapB50ImskirDrawsAndLosesHalfTheArtifactsAndSacrificesForDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gapB50Stock(g, me, 6)
	for i := 0; i < 5; i++ {
		pushCatalogPermanent(g, me.ID, "Trinket", "Artifact", "", false)
	}
	hand, life := me.Hand.Size(), me.Life
	id := castCatalogSpell(t, g, "Imskir Iron-Eater", "Legendary Creature — Demon", gapImskirOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 || me.Life != life-2 {
		t.Errorf("hand %d life %d, want %d / %d (five artifacts: X = 2)", me.Hand.Size(), me.Life, hand+2, life-2)
	}
	artifact := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: artifact, Name: "Three-drop", TypeLine: "Artifact", ManaCost: "{3}",
		Owner: me.ID, Controller: me.ID})
	oppLife := opp.Life
	gapB50Float(t, g, me, 3)
	floatMana(t, g, me, "{R}")
	err := gapActivateStrict(t, g, me.ID, id, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		SacrificeIDs: []uuid.UUID{artifact},
	})
	if err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-3 {
		t.Errorf("opponent life %d, want %d (the sacrificed artifact's mana value)", opp.Life, oppLife-3)
	}
}

func TestGapB50MunitionsExpertDealsDamageEqualToYourGoblins(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushDiesCreatureForTest(g, me.ID, "Goblin A", "", "Creature — Goblin", 1, 1)
	pushDiesCreatureForTest(g, me.ID, "Goblin B", "", "Creature — Goblin", 1, 1)
	victim := pushDiesCreatureForTest(g, opp.ID, "Victim", "", "Creature — Bear", 3, 3)
	castCatalogSpell(t, g, "Munitions Expert", "Creature — Goblin", gapMunitionsOracle, nil)
	passPriorityAroundTable(t, g)
	gapPick(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: victim})
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, victim)
	if ok && c.DamageMarked < 3 {
		t.Errorf("victim took %d, want 3 (three Goblins)", c.DamageMarked)
	}
}

func TestGapB50KaaliaPutsADragonInAttackingTheSameOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kaalia := pushCatalogPermanent(g, me.ID, "Kaalia of the Vast", "Legendary Creature — Human Cleric", gapKaaliaOracle, false)
	dragon := gapB50HandCard(me, "Test Dragon")
	bear := gapB50HandCard(me, "Test Bear")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			switch me.Hand.Cards[i].InstanceID {
			case dragon:
				me.Hand.Cards[i].TypeLine = "Creature — Dragon"
			case bear:
				me.Hand.Cards[i].TypeLine = "Creature — Bear"
			}
		}
	})
	declareAttack(t, g, opp.ID, kaalia)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to put a creature card onto the battlefield")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != dragon {
		t.Fatalf("candidates %v, want only the Dragon", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, dragon)
	c, ok := battlefieldCard(g, dragon)
	if !ok || !c.Tapped || c.AttackingTarget != opp.ID {
		t.Errorf("dragon on battlefield=%v tapped=%v attacking=%v, want tapped and attacking %v", ok, c.Tapped, c.AttackingTarget, opp.ID)
	}
}

func TestGapB50HorridShadowspinnerLootsForItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gapB50Stock(g, me, 6)
	spinner := pushCatalogPermanent(g, me.ID, "Horrid Shadowspinner", "Creature — Horror", gapShadowspinnerOrcl, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == spinner {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 2, 3
			}
		}
	})
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, spinner)
	passPriorityAroundTable(t, g)
	ask := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID)
	if ask == nil {
		t.Fatal("no may-draw prompt")
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d (drew two)", me.Hand.Size(), hand+2)
	}
	asked := false
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == me.ID && c.Kind != game.PendingChoiceConfirm {
			asked = true
		}
	}
	if !asked {
		t.Fatal("no discard prompt after drawing")
	}
}

func TestGapB50PsychicFrogTradesCardsForCountersAndFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	frog := pushCatalogPermanent(g, me.ID, "Psychic Frog", "Creature — Frog", gapFrogOracle, false)
	if err := gapActivateStrict(t, g, me.ID, frog, 1, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("exiled three cards from an empty graveyard")
	}
	var spent []uuid.UUID
	for i := 0; i < 3; i++ {
		spent = append(spent, pushGraveyardCardTyped(me, "Spent", "Sorcery"))
	}
	if err := gapActivateStrict(t, g, me.ID, frog, 1, game.ActivateAbilityParams{ExileIDs: spent}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, frog, "flying") {
		t.Error("no flying after exiling three cards")
	}
	if me.Graveyard.Size() != 0 {
		t.Errorf("graveyard has %d cards left", me.Graveyard.Size())
	}
	held := gapB50HandCard(me, "Pitch")
	if err := gapActivateStrict(t, g, me.ID, frog, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{held}}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if got := gapCountersOf(g, frog); got != 1 {
		t.Errorf("counters = %d, want 1", got)
	}
	if !me.Graveyard.Contains(held) {
		t.Error("the discarded card is not in the graveyard")
	}
}

func TestGapB50PsychicFrogDrawsOnCombatDamageToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gapB50Stock(g, me, 4)
	frog := pushCatalogPermanent(g, me.ID, "Psychic Frog", "Creature — Frog", gapFrogOracle, false)
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, frog)
	for i := 0; i < 12 && g.Turn.Step != game.StepEndCombat; i++ {
		passPriorityAroundTable(t, g)
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestGapB50SneakySnackerReturnsOnTheThirdDrawOnly(t *testing.T) {
	for _, draws := range []int{2, 3} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		gapB50Stock(g, me, 6)
		snack := pushGraveyardCardTyped(me, "Sneaky Snacker", "Creature — Faerie Rogue")
		g.WithWriteLock(func() {
			for i := range me.Graveyard.Cards {
				if me.Graveyard.Cards[i].InstanceID == snack {
					me.Graveyard.Cards[i].OracleID = gapSnackerOracle
				}
			}
			_ = g.DrawNForEffect(me.ID, draws-len(g.CardsDrawnThisTurnFor(me.ID)))
		})
		passPriorityAroundTable(t, g)
		c, back := battlefieldCard(g, snack)
		if want := draws == 3; back != want {
			t.Errorf("draws=%d: on battlefield=%v, want %v", draws, back, want)
		}
		if back && !c.Tapped {
			t.Error("it returned untapped")
		}
	}
}

func TestGapB50OnduKnotmasterGrowsOnlyForModifiedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ondu := pushCatalogPermanent(g, me.ID, "Ondu Knotmaster", "Creature — Kor Rogue", gapOnduOracle, false)
	plain := pushVanillaCreature(g, me.ID, "Plain", 2, 2)
	modified := pushVanillaCreature(g, me.ID, "Modified", 2, 2)
	gapAddCounter(t, g, modified, 1)
	killCreature(t, g, me.ID, plain)
	passPriorityAroundTable(t, g)
	if got := gapCountersOf(g, ondu); got != 0 {
		t.Fatalf("an unmodified creature's death gave %d counters", got)
	}
	killCreature(t, g, me.ID, modified)
	passPriorityAroundTable(t, g)
	if got := gapCountersOf(g, ondu); got != 2 {
		t.Errorf("counters = %d, want 2", got)
	}
}

func TestGapB50ThrowALineDividesTwoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := pushVanillaCreature(g, me.ID, "A", 1, 1)
	b := pushVanillaCreature(g, me.ID, "B", 1, 1)
	card := game.Card{
		InstanceID: uuid.New(), OracleID: gapOnduOracle, Layout: game.LayoutAdventure, Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Ondu Knotmaster", TypeLine: "Creature — Kor Rogue", ManaCost: "{2}{W}{B}", Power: 2, Toughness: 2},
			{Name: "Throw a Line", TypeLine: "Sorcery — Adventure", ManaCost: "{W}{B}"},
		},
	}
	card.SetFace(0)
	me.Hand.PushTop(card)
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: cardRefs(a, b),
		Distribution: map[uuid.UUID]int{a: 1, b: 1}}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if gapCountersOf(g, a) != 1 || gapCountersOf(g, b) != 1 {
		t.Errorf("counters %d / %d, want 1 / 1", gapCountersOf(g, a), gapCountersOf(g, b))
	}
	if z := g.FindCardZoneForEffect(card.InstanceID); z == nil || z.Kind != game.ZoneExile {
		t.Error("the Adventure did not go to exile")
	}
}

func TestGapB50ShilgengarMakesBloodAndOneForEachPointOfAnAngelsToughness(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sh := pushCatalogPermanent(g, me.ID, "Shilgengar, Sire of Famine", "Legendary Creature — Elder Demon", gapShilgengarOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	angel := pushDiesCreatureForTest(g, me.ID, "Test Angel", "", "Creature — Angel", 3, 4)
	for _, tc := range []struct {
		victim uuid.UUID
		want   int
	}{{bear, 1}, {angel, 5}} {
		if err := gapActivateStrict(t, g, me.ID, sh, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{tc.victim}}); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		if got := randomTokenCount(g, "Blood"); got != tc.want {
			t.Errorf("after sacrificing %v: %d Blood tokens, want %d", tc.victim == angel, got, tc.want)
		}
	}
	if err := gapActivateStrict(t, g, me.ID, sh, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{sh}}); err == nil {
		t.Error("Shilgengar sacrificed itself to its own ability")
	}
}

func TestGapB50ShilgengarSixBloodReturnsEveryCreatureAsAVampire(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sh := pushCatalogPermanent(g, me.ID, "Shilgengar, Sire of Famine", "Legendary Creature — Elder Demon", gapShilgengarOracle, false)
	var blood []uuid.UUID
	for i := 0; i < 6; i++ {
		id := uuid.New()
		blood = append(blood, id)
		g.Battlefield.PushTop(game.Card{InstanceID: id, Name: "Blood", TypeLine: "Token Artifact — Blood",
			Owner: me.ID, Controller: me.ID})
	}
	pushGraveyardPermanent(me, "Corpse One", "Creature — Zombie", "")
	pushGraveyardPermanent(me, "Corpse Two", "Creature — Elf", "")
	relic := pushGraveyardPermanent(me, "Relic", "Artifact", "")
	gapB50Float(t, g, me, 3)
	if err := gapActivateStrict(t, g, me.ID, sh, 1, game.ActivateAbilityParams{SacrificeIDs: blood[:5]}); err == nil {
		t.Fatal("paid with only five Blood tokens")
	}
	if err := gapActivateStrict(t, g, me.ID, sh, 1, game.ActivateAbilityParams{SacrificeIDs: blood}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	vampires := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Corpse One" || c.Name == "Corpse Two" {
			if c.Counters[game.CounterFinality] != 1 || !c.HasSubtype("Vampire") {
				t.Errorf("%s: finality %d, vampire %v", c.Name, c.Counters[game.CounterFinality], c.HasSubtype("Vampire"))
			}
			vampires++
		}
	}
	if vampires != 2 {
		t.Errorf("%d creatures returned, want 2", vampires)
	}
	if !me.Graveyard.Contains(relic) {
		t.Error("a noncreature card left the graveyard")
	}
}
