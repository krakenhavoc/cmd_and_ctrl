package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB161ScoutOracle = "51d7b287-90d5-4e3b-a82c-6b98996154e1"
)

func TestOldB161TrustworthyScoutFetchesAnotherFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	other := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: other, Name: "Trustworthy Scout", TypeLine: "Creature — Human Scout", OracleID: oldB161ScoutOracle, Owner: me.ID, Controller: me.ID})
	decoy := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: decoy, Name: "Trustworthy Thing", TypeLine: "Creature — Human", Owner: me.ID, Controller: me.ID})
	id, _ := effectstest.ActivateFromGraveyard(t, g, "Trustworthy Scout", "Creature — Human Scout", oldB161ScoutOracle, 2, 2, "{C}{W}", game.ActivateAbilityParams{})
	if !g.Exile.Contains(id) {
		t.Fatal("the cost did not exile the Scout")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	if effectstest.SearchOptionNamed(g, c, "Trustworthy Thing") != uuid.Nil {
		t.Error("a card with another name was offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Trustworthy Scout")
	if !me.Hand.Contains(other) {
		t.Fatalf("the second Scout is in %q, not the hand", effectstest.ZoneOf(g, other))
	}
}

func TestOldB161TrustworthyScoutDoesNothingOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	scout := effectstest.B12Push(g, me.ID, "Trustworthy Scout", "Creature — Human Scout", oldB161ScoutOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "W", 2)
	if err := g.ActivateCatalogAbility(me.ID, scout, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated from the battlefield")
	}
}

const oldB161GenesisOracle = "15afdd88-05fc-4335-8a55-c6ee702f3bef"

func oldB161Genesis(me *game.Player) uuid.UUID {
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Genesis", TypeLine: "Creature — Incarnation", ManaCost: "{4}{G}",
		OracleID: oldB161GenesisOracle, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	return id
}

func TestOldB161GenesisPaysToReturnACreatureCardFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	genesis := oldB161Genesis(me)
	bear := effectstest.PushCatalogGraveyardCard(me, "Dead Bear", "Creature — Bear", "", 2, 2)
	land := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{G}"); err != nil {
		t.Fatal(err)
	}
	effectstest.CloudDeckPickTargets(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) || me.Graveyard.Contains(bear) {
		t.Fatalf("the Bear is in %q, not the hand", effectstest.ZoneOf(g, bear))
	}
	if !me.Graveyard.Contains(genesis) || me.Hand.Contains(land) {
		t.Error("Genesis or the land moved")
	}
}

func TestOldB161GenesisDecliningTheCostReturnsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB161Genesis(me)
	bear := effectstest.PushCatalogGraveyardCard(me, "Dead Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.CloudDeckPickTargets(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(bear) {
		t.Error("the Bear returned without payment")
	}
}

func TestOldB161GenesisCanTargetItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	genesis := oldB161Genesis(me)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{G}"); err != nil {
		t.Fatal(err)
	}
	effectstest.CloudDeckPickTargets(t, g, me.ID, genesis)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(genesis) {
		t.Fatalf("Genesis is in %q, not the hand", effectstest.ZoneOf(g, genesis))
	}
}

func TestOldB161GenesisDoesNothingOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	onBattlefield := effectstest.PushCatalogPermanent(g, me.ID, "Genesis", "Creature — Incarnation", oldB161GenesisOracle, false)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if n := effectstest.TriggersOnStackFrom(g, onBattlefield) + len(g.PendingChoices); n != 0 {
		t.Fatalf("the battlefield Genesis raised %d triggers or prompts", n)
	}
}

const oldB161VesperlarkOracle = "b801a701-15ce-4792-a3d7-d632e7bd9f72"

func TestOldB161VesperlarkReturnsAPowerOneCreatureWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lark := effectstest.B12Push(g, me.ID, "Vesperlark", "Creature — Elemental", oldB161VesperlarkOracle, 2, 1)
	small := effectstest.PushCatalogGraveyardCard(me, "Small Thing", "Creature — Elf", "", 1, 1)
	big := effectstest.PushCatalogGraveyardCard(me, "Big Thing", "Creature — Ogre", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.KillCreature(t, g, me.ID, lark)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: big}}); err == nil {
		t.Fatal("a power-2 card was accepted as the target")
	}
	effectstest.CloudDeckPickTargets(t, g, me.ID, small)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Fatalf("the power-1 creature is in %q, not on the battlefield", effectstest.ZoneOf(g, small))
	}
	if g.Battlefield.Contains(big) {
		t.Error("the power-2 creature returned")
	}
}

func TestOldB161VesperlarkEvokedReturnsAPowerOneCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	small := effectstest.PushCatalogGraveyardCard(me, "Small Thing", "Creature — Elf", "", 1, 1)
	id := effectstest.CastWithAltCost(t, g, "Vesperlark", "Creature — Elemental", oldB161VesperlarkOracle, "evoke")
	for i := 0; i < 6; i++ {
		effectstest.PickIfAsked(t, g, me.ID, small)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(id) || !me.Graveyard.Contains(id) {
		t.Error("the evoked Vesperlark was not sacrificed")
	}
	if !g.Battlefield.Contains(small) {
		t.Fatalf("the power-1 creature is in %q, not on the battlefield", effectstest.ZoneOf(g, small))
	}
}

const oldB161ReapOracle = "27bc0219-3fc1-4c84-92cc-da2273a5fd16"

func oldB161InHand(me *game.Player, ids ...uuid.UUID) int {
	n := 0
	for _, id := range ids {
		if me.Hand.Contains(id) {
			n++
		}
	}
	return n
}

func TestOldB161ReapThePastReturnsXCardsAtRandomAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushCatalogGraveyardCard(me, "Card A", "Creature — Bear", "", 2, 2)
	b := effectstest.PushCatalogGraveyardCard(me, "Card B", "Creature — Bear", "", 2, 2)
	c := effectstest.PushCatalogGraveyardCard(me, "Card C", "Creature — Bear", "", 2, 2)
	id := effectstest.CastXSpell(t, g, "Reap the Past", "Sorcery", oldB161ReapOracle, "{X}{R}{G}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB161InHand(me, a, b, c); n != 2 {
		t.Fatalf("%d graveyard cards returned, want 2", n)
	}
	if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
		t.Errorf("Reap the Past is in %q, want exile", effectstest.ZoneOf(g, id))
	}
}

func TestOldB161ReapThePastReturnsEverythingWhenXIsLarger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushCatalogGraveyardCard(me, "Card A", "Creature — Bear", "", 2, 2)
	b := effectstest.PushCatalogGraveyardCard(me, "Card B", "Creature — Bear", "", 2, 2)
	id := effectstest.CastXSpell(t, g, "Reap the Past", "Sorcery", oldB161ReapOracle, "{X}{R}{G}", 5, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB161InHand(me, a, b); n != 2 {
		t.Fatalf("%d graveyard cards returned, want both", n)
	}
	if !g.Exile.Contains(id) {
		t.Error("Reap the Past was not exiled")
	}
}

const oldB161StreamOracle = "e5b29094-0bb4-4c18-ad47-dce4d4379a4f"

func TestOldB161StreamOfThoughtMillsFourAndShufflesUpToFourBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.PushCatalogGraveyardCard(me, "Card A", "Creature — Bear", "", 2, 2)
	b := effectstest.PushCatalogGraveyardCard(me, "Card B", "Creature — Bear", "", 2, 2)
	c := effectstest.PushCatalogGraveyardCard(me, "Card C", "Creature — Bear", "", 2, 2)
	spell := effectstest.AvB24Hand(g, me, game.Card{Name: "Stream of Thought", TypeLine: "Sorcery", ManaCost: "{U}", OracleID: oldB161StreamOracle})
	effectstest.FloatForTest(g, me, "U")
	oppLibrary, oppGraveyard, myLibrary := opp.Library.Size(), opp.Graveyard.Size(), me.Library.Size()
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: effectstest.PlayerRefs(opp.ID), Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != oppLibrary-4 {
		t.Errorf("the target's library is %d, want %d", got, oppLibrary-4)
	}
	if got := opp.Graveyard.Size(); got != oppGraveyard+4 {
		t.Errorf("the target's graveyard is %d, want %d", got, oppGraveyard+4)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(a) || me.Graveyard.Contains(b) || !me.Graveyard.Contains(c) {
		t.Error("exactly A and B should have left the graveyard")
	}
	if got := me.Library.Size(); got != myLibrary+2 {
		t.Errorf("your library is %d, want %d", got, myLibrary+2)
	}
}

func TestOldB161StreamOfThoughtReplicatesIntoTwoMills(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.AdvanceToMain(t, g)
	spell := effectstest.AvB24Hand(g, me, game.Card{Name: "Stream of Thought", TypeLine: "Sorcery", ManaCost: "{U}", OracleID: oldB161StreamOracle})
	effectstest.FloatForTest(g, me, "UUUUUCC")
	oppLibrary := opp.Library.Size()
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: effectstest.PlayerRefs(opp.ID), OptionalCosts: []int{0}, Strict: true}); err != nil {
		t.Fatalf("cast with one replicate: %v", err)
	}
	for i := 0; i < 64; i++ {
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, effectstest.PbPlayer(opp.ID)); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
			effectstest.AnswerChooseCards(t, g, me.ID)
			continue
		}
		if effectstest.StackFullyEmpty(g) {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if got := opp.Library.Size(); got != oppLibrary-8 {
		t.Errorf("the target's library is %d, want %d (two mills of four)", got, oppLibrary-8)
	}
}

const oldB161RotwidowOracle = "27b5d9d8-cc19-46be-9862-405cf7ef5408"

func TestOldB161RotwidowPackMakesASpiderAndDrainsPerSpider(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pack := effectstest.B12Push(g, me.ID, "Rotwidow Pack", "Creature — Spider", oldB161RotwidowOracle, 2, 4)
	effectstest.AdvanceToMain(t, g)
	spell := effectstest.PushGraveyardCardTyped(me, "Spent Spell", "Sorcery")
	bear := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{B}{G}")
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	if err := g.ActivateCatalogAbility(me.ID, pack, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{spell}}); err == nil {
		t.Fatal("a sorcery paid \"exile a creature card\"")
	}
	if err := g.ActivateCatalogAbility(me.ID, pack, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Exile.Contains(bear) {
		t.Fatal("the cost did not exile the creature card")
	}
	effectstest.PassPriorityAroundTable(t, g)
	spiders := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && c.HasSubtype("Spider") {
			spiders++
		}
	}
	if spiders != 2 {
		t.Fatalf("%d Spiders, want the Pack and one token", spiders)
	}
	for _, p := range g.Seats {
		want := lives[p.ID]
		if p.ID != me.ID {
			want -= 2
		}
		if p.Life != want {
			t.Errorf("player life %d, want %d", p.Life, want)
		}
	}
}

func TestOldB161RotwidowPackNeedsACreatureCardToExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pack := effectstest.B12Push(g, me.ID, "Rotwidow Pack", "Creature — Spider", oldB161RotwidowOracle, 2, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushGraveyardCardTyped(me, "Spent Spell", "Sorcery")
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{B}{G}")
	if err := g.ActivateCatalogAbility(me.ID, pack, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with no creature card in the graveyard")
	}
}

const oldB161HogaakOracle = "97acd977-164d-49d5-b3a1-72e0d75c9016"

func oldB161HogaakIn(g *game.Game, me *game.Player, zone string) uuid.UUID {
	id := uuid.New()
	c := game.Card{InstanceID: id, Name: "Hogaak, Arisen Necropolis", TypeLine: "Legendary Creature — Avatar", ManaCost: "{5}{B/G}{B/G}",
		OracleID: oldB161HogaakOracle, Power: 8, Toughness: 8, Owner: me.ID, Controller: me.ID}
	if zone == "graveyard" {
		me.Graveyard.PushTop(c)
	} else {
		me.Hand.PushTop(c)
	}
	return id
}

func TestOldB161HogaakIsCastWithConvokeAndDelveAloneFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	hogaak := oldB161HogaakIn(g, me, "graveyard")
	fuel := effectstest.DelveFuel(me, 5)
	black := effectstest.PushTapCostPermanent(g, me.ID, "Black Bear", "Creature — Bear", "B")
	green := effectstest.PushTapCostPermanent(g, me.ID, "Green Bear", "Creature — Bear", "G")
	if err := g.CastSpell(me.ID, hogaak, game.CastSpellParams{Strict: true, FromZone: "graveyard", DelveIDs: fuel, TapIDs: []uuid.UUID{black, green}}); err != nil {
		t.Fatalf("cast from the graveyard with convoke and delve: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(hogaak) {
		t.Fatalf("Hogaak is in %q, not on the battlefield", effectstest.ZoneOf(g, hogaak))
	}
	for _, id := range fuel {
		if !g.Exile.Contains(id) {
			t.Error("a delved card was not exiled")
		}
	}
}

func TestOldB161HogaakCannotBePaidForWithMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	hogaak := oldB161HogaakIn(g, me, "hand")
	effectstest.FloatForTest(g, me, "BGCCCCC")
	if err := g.CastSpell(me.ID, hogaak, game.CastSpellParams{Strict: true}); !errors.Is(err, game.ErrCantSpendMana) {
		t.Fatalf("cast with mana alone: %v, want ErrCantSpendMana", err)
	}
	fuel := effectstest.DelveFuel(me, 4)
	black := effectstest.PushTapCostPermanent(g, me.ID, "Black Bear", "Creature — Bear", "B")
	green := effectstest.PushTapCostPermanent(g, me.ID, "Green Bear", "Creature — Bear", "G")
	if err := g.CastSpell(me.ID, hogaak, game.CastSpellParams{Strict: true, DelveIDs: fuel, TapIDs: []uuid.UUID{black, green}}); !errors.Is(err, game.ErrCantSpendMana) {
		t.Fatalf("cast with one generic left to pay from the pool: %v, want ErrCantSpendMana", err)
	}
	if me.Hand.Contains(hogaak) == false {
		t.Error("a refused cast moved Hogaak")
	}
}
