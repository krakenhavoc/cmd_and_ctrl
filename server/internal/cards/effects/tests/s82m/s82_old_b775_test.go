package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b775_test.go — S82 OLD batch 775 (Weatherlight: graveyard recursion,
// #5629), each card driven through the real cast, activation or trigger path.

const (
	oldb775ShatteredCrypt = "77bef290-7229-420a-a5c6-8c1b6fea63df"
	oldb775StrandsOfNight = "0e8ae1f7-8926-414a-8434-db0b0f877505"
	oldb775Hierophant     = "e98361d9-1e3a-464e-bac6-75314fc49f32"
	oldb775HarvestWurm    = "cd539212-ef7b-4f78-8d62-c42289aedc51"
	oldb775Furnace        = "4aafb7e7-b9fc-44ae-9c54-da11fd9dcc17"
	oldb775Barishi        = "ffa93a79-7a10-4331-91de-eabbc4fee17e"
	oldb775SoulShepherd   = "78dbcc18-291c-4b2c-9576-f8dc93914936"
)

func TestOldB775EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldb775ShatteredCrypt: "Shattered Crypt", oldb775StrandsOfNight: "Strands of Night", oldb775Hierophant: "Sylvan Hierophant",
		oldb775HarvestWurm: "Harvest Wurm", oldb775Furnace: "Phyrexian Furnace", oldb775Barishi: "Barishi", oldb775SoulShepherd: "Soul Shepherd",
	} {
		effectstest.CmdrB21AssertFull(t, oracle, name)
	}
}

func oldb775Main(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	return g, me, opp
}

func oldb775Graveyard(p *game.Player, typeLines ...string) []uuid.UUID {
	var ids []uuid.UUID
	for _, tl := range typeLines {
		id := uuid.New()
		p.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Card " + tl, TypeLine: tl, Owner: p.ID, Controller: p.ID})
		ids = append(ids, id)
	}
	return ids
}

func oldb775Refs(ids ...uuid.UUID) []game.TargetRef {
	var refs []game.TargetRef
	for _, id := range ids {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return refs
}

func TestOldB775ShatteredCryptReturnsXCreatureCardsAndLosesXLife(t *testing.T) {
	g, me, _ := oldb775Main(t)
	ids := oldb775Graveyard(me, "Creature — Bear", "Creature — Elf", "Sorcery", "Creature — Wolf")
	bear, elf, sorcery := ids[0], ids[1], ids[2]
	effectstest.FillPoolColored(me, "B", 2)
	effectstest.FillPool(me, 2)
	card := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: card, Name: "Shattered Crypt", TypeLine: "Sorcery", OracleID: oldb775ShatteredCrypt,
		ManaCost: "{X}{B}{B}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, XValue: 2, Targets: oldb775Refs(bear)}); err == nil {
		t.Fatal("one target accepted for X = 2")
	}
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, XValue: 2, Targets: oldb775Refs(bear, sorcery)}); err == nil {
		t.Fatal("a non-creature card accepted as a target")
	}
	life := me.Life
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, XValue: 2, Targets: oldb775Refs(bear, elf)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) || !me.Hand.Contains(elf) || me.Hand.Contains(sorcery) {
		t.Error("the wrong cards came back to the hand")
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
}

func TestOldB775StrandsOfNightCostsAndReanimates(t *testing.T) {
	g, me, _ := oldb775Main(t)
	strands := effectstest.PushCatalogPermanent(g, me.ID, "Strands of Night", "Enchantment", oldb775StrandsOfNight, false)
	swamp := effectstest.B12Push(g, me.ID, "Swamp", "Basic Land — Swamp", "", 0, 0)
	forest := effectstest.B12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	ids := oldb775Graveyard(me, "Creature — Bear", "Sorcery")
	bear, sorcery := ids[0], ids[1]
	effectstest.FillPoolColored(me, "B", 2)
	act := func(sac uuid.UUID, target uuid.UUID) error {
		return g.ActivateCatalogAbility(me.ID, strands, 0, game.ActivateAbilityParams{Strict: true,
			SacrificeIDs: []uuid.UUID{sac}, Targets: oldb775Refs(target)})
	}
	if err := act(forest, bear); err == nil {
		t.Fatal("sacrificed a Forest to a Swamp clause")
	}
	if err := act(swamp, sorcery); err == nil {
		t.Fatal("a sorcery card was a legal target")
	}
	life := me.Life
	if err := act(swamp, bear); err != nil {
		t.Fatal(err)
	}
	if g.Battlefield.Contains(swamp) || me.Life != life-2 {
		t.Errorf("the Swamp is still there (%v) or life is %d, want %d", g.Battlefield.Contains(swamp), me.Life, life-2)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the creature card did not return to the battlefield")
	}
	var controller uuid.UUID
	g.ReadSnapshot(func() { controller = effectstest.FindBattlefieldCardForTest(g, bear).Controller })
	if controller != me.ID {
		t.Error("the creature returned under the wrong player")
	}
}

func TestOldB775SoulShepherdExilesAnyCreatureCardForOneLife(t *testing.T) {
	g, me, _ := oldb775Main(t)
	shepherd := effectstest.B12Push(g, me.ID, "Soul Shepherd", "Creature — Human Cleric", oldb775SoulShepherd, 2, 1)
	ids := oldb775Graveyard(me, "Creature — Bear", "Sorcery")
	bear, sorcery := ids[0], ids[1]
	effectstest.FillPoolColored(me, "W", 2)
	run := func(card uuid.UUID) error {
		return g.ActivateCatalogAbility(me.ID, shepherd, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{card}})
	}
	if err := run(sorcery); err == nil {
		t.Fatal("a sorcery paid for a creature clause")
	}
	life := me.Life
	if err := run(bear); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) || me.Life != life+1 {
		t.Errorf("bear exiled %v, life %d want %d", g.Exile.Contains(bear), me.Life, life+1)
	}
}

func TestOldB775PhyrexianFurnaceExilesTheBottomCard(t *testing.T) {
	g, me, opp := oldb775Main(t)
	furnace := effectstest.PushCatalogPermanent(g, me.ID, "Phyrexian Furnace", "Artifact", oldb775Furnace, false)
	ids := oldb775Graveyard(opp, "Creature — Bear", "Sorcery", "Instant")
	if err := g.ActivateCatalogAbility(me.ID, furnace, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(ids[0]) || !opp.Graveyard.Contains(ids[1]) || !opp.Graveyard.Contains(ids[2]) {
		t.Error("the bottom card was not the one exiled")
	}
}

func TestOldB775PhyrexianFurnaceSacrificesToExileAndDraw(t *testing.T) {
	g, me, opp := oldb775Main(t)
	furnace := effectstest.PushCatalogPermanent(g, me.ID, "Phyrexian Furnace", "Artifact", oldb775Furnace, false)
	ids := oldb775Graveyard(opp, "Creature — Bear", "Sorcery")
	effectstest.FillPool(me, 1)
	hand := len(me.Hand.Cards)
	if err := g.ActivateCatalogAbility(me.ID, furnace, 1, game.ActivateAbilityParams{Strict: true, Targets: oldb775Refs(ids[1])}); err != nil {
		t.Fatal(err)
	}
	if g.Battlefield.Contains(furnace) {
		t.Error("the Furnace was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(ids[1]) || !opp.Graveyard.Contains(ids[0]) {
		t.Error("the targeted card was not the one exiled")
	}
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d, want %d: no card drawn", len(me.Hand.Cards), hand+1)
	}
}

func TestOldB775SylvanHierophantExilesItselfAndReturnsAnotherCreatureCard(t *testing.T) {
	g, me, _ := oldb775Main(t)
	hiero := effectstest.B12Push(g, me.ID, "Sylvan Hierophant", "Creature — Human Cleric", oldb775Hierophant, 1, 2)
	ids := oldb775Graveyard(me, "Creature — Bear", "Sorcery")
	bear := ids[0]
	effectstest.KillCreature(t, g, me.ID, hiero)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(hiero) || me.Graveyard.Contains(hiero) || me.Hand.Contains(hiero) {
		t.Error("the Hierophant was not exiled")
	}
	if !me.Hand.Contains(bear) {
		t.Error("the other creature card did not return to the hand")
	}
}

func TestOldB775SylvanHierophantNeverTargetsItself(t *testing.T) {
	g, me, _ := oldb775Main(t)
	hiero := effectstest.B12Push(g, me.ID, "Sylvan Hierophant", "Creature — Human Cleric", oldb775Hierophant, 1, 2)
	effectstest.KillCreature(t, g, me.ID, hiero)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(hiero) {
		t.Error("the Hierophant returned itself")
	}
}

func TestOldB775BarishiShufflesCreatureCardsIntoTheLibrary(t *testing.T) {
	g, me, _ := oldb775Main(t)
	barishi := effectstest.B12Push(g, me.ID, "Barishi", "Creature — Elemental", oldb775Barishi, 4, 3)
	ids := oldb775Graveyard(me, "Creature — Bear", "Sorcery", "Artifact Creature — Golem")
	lib := len(me.Library.Cards)
	effectstest.KillCreature(t, g, me.ID, barishi)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(barishi) {
		t.Error("Barishi was not exiled")
	}
	if me.Graveyard.Contains(ids[0]) || me.Graveyard.Contains(ids[2]) || !me.Graveyard.Contains(ids[1]) {
		t.Error("the creature cards, and only they, were not shuffled away")
	}
	if got := len(me.Library.Cards) - lib; got != 2 {
		t.Errorf("the library grew by %d, want 2 (Barishi itself is exiled, not shuffled)", got)
	}
}

func TestOldB775HarvestWurmStaysForABasicLandCard(t *testing.T) {
	g, me, _ := oldb775Main(t)
	ids := oldb775Graveyard(me, "Land", "Basic Land — Forest")
	nonbasic, basic := ids[0], ids[1]
	wurm := effectstest.CastAndResolveCreature(t, g, "Harvest Wurm", "Creature — Wurm", oldb775HarvestWurm)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if p == nil {
		t.Fatal("the Wurm asked for no choice")
	}
	if len(p.ChooseCards) != 1 || p.ChooseCards[0] != basic {
		t.Errorf("offered %v, want only the basic land card %v (not %v)", p.ChooseCards, basic, nonbasic)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, basic)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wurm) {
		t.Error("the Wurm was sacrificed after the return was paid")
	}
	if !me.Hand.Contains(basic) {
		t.Error("the basic land card did not reach the hand")
	}
}

func TestOldB775HarvestWurmIsSacrificedWhenTheReturnIsDeclined(t *testing.T) {
	g, me, _ := oldb775Main(t)
	basic := oldb775Graveyard(me, "Basic Land — Forest")[0]
	wurm := effectstest.CastAndResolveCreature(t, g, "Harvest Wurm", "Creature — Wurm", oldb775HarvestWurm)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wurm) {
		t.Error("the Wurm stayed without the payment")
	}
	if !me.Graveyard.Contains(basic) {
		t.Error("the basic land card left the graveyard on a declined payment")
	}
}

func TestOldB775HarvestWurmIsSacrificedWithNoBasicLandCard(t *testing.T) {
	g, me, _ := oldb775Main(t)
	oldb775Graveyard(me, "Land", "Creature — Bear")
	wurm := effectstest.CastAndResolveCreature(t, g, "Harvest Wurm", "Creature — Wurm", oldb775HarvestWurm)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wurm) {
		t.Error("the Wurm stayed with no basic land card to return")
	}
}
