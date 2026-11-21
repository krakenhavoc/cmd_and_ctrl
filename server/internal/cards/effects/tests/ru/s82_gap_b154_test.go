package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b154_test.go — GAP batch 154 (Jumpstart 2022 creatures, #5948),
// played through the catalog and the real stack.

const (
	gapB154Kibo        = "0b9be4fa-5238-4afd-a9f6-f9022e67e5ab"
	gapB154Kenessos    = "45b3a028-5705-4dc8-bfab-04bb5e01eea6"
	gapB154Isu         = "b34823a3-7bee-4e71-a3df-9b533c8ddf15"
	gapB154Ardoz       = "bcd5e175-2ef0-48f4-b9bd-14383dd234e0"
	gapB154ArmsDealer  = "8ebc4198-7317-4ffb-b8b8-14733c2077ff"
	gapB154Blyte       = "a5e9142e-76bf-40d4-aeaa-ffe197d38f25"
	gapB154Squadron    = "5783e71e-2497-4241-9d1a-8bced4745b1d"
	gapB154Gremlin     = "cd8a90a7-5502-405f-a49e-ad4252872595"
	gapB154Dragonheart = "3f0dc347-adac-428b-8f04-858799e91243"
	gapB154Bolt        = "4457ed35-7c10-48c8-9776-456485fdf070"
	gapB154Bear        = "Creature — Bear"
)

func TestGapB154ArdozPumpsEachCreatureThatEntersUnderYourControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ardoz := gapB154Enter(t, g, me, "Ardoz, Cobbler of War", "Legendary Creature — Goblin Shaman", gapB154Ardoz, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if live := effectstest.ApaLive(g, ardoz); live.CurrentPower() != 3 {
		t.Errorf("Ardoz entered and is %d/%d, want 3/1", live.CurrentPower(), live.CurrentToughness())
	}
	bear := gapB154Enter(t, g, me, "Bear", gapB154Bear, "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if live := effectstest.ApaLive(g, bear); live.CurrentPower() != 4 {
		t.Errorf("a creature that entered is %d/%d, want 4/2", live.CurrentPower(), live.CurrentToughness())
	}
	theirs := gapB154Enter(t, g, opp, "Their Bear", gapB154Bear, "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if live := effectstest.ApaLive(g, theirs); live.CurrentPower() != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", live.CurrentPower(), live.CurrentToughness())
	}
}

func TestGapB154ArdozMakesAHastyGoblinAtSorcerySpeedOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ardoz := effectstest.B12Push(g, me.ID, "Ardoz, Cobbler of War", "Legendary Creature — Goblin Shaman", gapB154Ardoz, 1, 1)
	effectstest.FloatForTest(g, me, "RRRR")
	effectstest.ApaActivate(t, g, me, ardoz, 0, game.ActivateAbilityParams{})
	n, _ := effectstest.CountTokensNamed(g, me.ID, "Goblin")
	if n != 1 {
		t.Fatalf("%d Goblin tokens, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Goblin" {
			if live := effectstest.ApaLive(g, c.InstanceID); live.CurrentPower() != 3 || !effectstest.P1g2HasKeyword(g, c.InstanceID, "haste") {
				t.Errorf("the token is %d/%d (haste %v), want a hasty 3/1 (Ardoz pumps it as it enters)",
					live.CurrentPower(), live.CurrentToughness(), effectstest.P1g2HasKeyword(g, c.InstanceID, "haste"))
			}
		}
	}
	// Not at sorcery speed: with a spell on the stack it is refused.
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", gapB154Bolt, []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})
	effectstest.FloatForTest(g, me, "RRRR")
	if err := g.ActivateCatalogAbility(me.ID, ardoz, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("Ardoz's ability was activated with a spell on the stack")
	}
}

func TestGapB154ArmsDealerSacrificesAGoblinForFourDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	dealer := effectstest.B12Push(g, me.ID, "Arms Dealer", "Creature — Goblin Rogue", gapB154ArmsDealer, 1, 1)
	goblin := effectstest.B12Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", gapB154Bear, 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", gapB154Bear, 4, 4)
	effectstest.FloatForTest(g, me, "RR")
	if err := g.ActivateCatalogAbility(me.ID, dealer, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(big), SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("Arms Dealer sacrificed a non-Goblin")
	}
	effectstest.ApaActivate(t, g, me, dealer, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(big), SacrificeIDs: []uuid.UUID{goblin}})
	if g.Battlefield.Contains(goblin) {
		t.Error("the Goblin was not sacrificed")
	}
	if g.Battlefield.Contains(big) {
		t.Error("the target survived 4 damage")
	}
	// It may sacrifice itself; the damage is then dealt as it last existed.
	big2 := effectstest.B12Creature(g, opp.ID, "Big", gapB154Bear, 4, 4)
	effectstest.FloatForTest(g, me, "RR")
	effectstest.ApaActivate(t, g, me, dealer, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(big2), SacrificeIDs: []uuid.UUID{dealer}})
	if g.Battlefield.Contains(dealer) || g.Battlefield.Contains(big2) {
		t.Error("sacrificing itself should still deal 4 damage to the target")
	}
}

func TestGapB154BlyteGrowsFromDamageYourSourcesDealToYouAndSpendsCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	blyte := effectstest.B12Push(g, me.ID, "Auntie Blyte, Bad Influence", "Legendary Creature — Devil Advisor", gapB154Blyte, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", gapB154Bolt, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, blyte, game.CounterPlusOne); got != 3 {
		t.Fatalf("Blyte has %d counters after my own Bolt hit me for 3, want 3", got)
	}
	// An opponent's source is not "a source you control".
	pinger := effectstest.B12Creature(g, opp.ID, "Pinger", gapB154Bear, 1, 1)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(pinger, me.ID, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, blyte, game.CounterPlusOne); got != 3 {
		t.Errorf("Blyte has %d counters after an opponent's source hit me, want 3", got)
	}
	// {1}{R}, {T}, remove 2: 2 damage to any target.
	effectstest.FloatForTest(g, me, "RR")
	life := opp.Life
	effectstest.ApaActivate(t, g, me, blyte, 0, game.ActivateAbilityParams{CounterCounts: []int{2}, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if opp.Life != life-2 {
		t.Errorf("the target is at %d, want %d", opp.Life, life-2)
	}
	if got := effectstest.CounterCount(g, blyte, game.CounterPlusOne); got != 1 {
		t.Errorf("Blyte has %d counters after removing 2, want 1", got)
	}
}

func TestGapB154SquadronIsAsBigAsYourCreatureCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	sq := effectstest.B12Push(g, me.ID, "Battle Squadron", "Creature — Goblin", gapB154Squadron, 0, 0)
	if live := effectstest.ApaLive(g, sq); live.CurrentPower() != 1 || live.CurrentToughness() != 1 {
		t.Errorf("alone it is %d/%d, want 1/1", live.CurrentPower(), live.CurrentToughness())
	}
	effectstest.B12Creature(g, me.ID, "Bear", gapB154Bear, 2, 2)
	effectstest.B12Creature(g, me.ID, "Bear", gapB154Bear, 2, 2)
	effectstest.B12Creature(g, opp.ID, "Theirs", gapB154Bear, 2, 2)
	if live := effectstest.ApaLive(g, sq); live.CurrentPower() != 3 || live.CurrentToughness() != 3 {
		t.Errorf("with two more creatures it is %d/%d, want 3/3", live.CurrentPower(), live.CurrentToughness())
	}
}

func TestGapB154GremlinPingsEachOpponentAndUntapsOnANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	gremlin := effectstest.B12Push(g, me.ID, "Blisterspit Gremlin", "Creature — Gremlin", gapB154Gremlin, 1, 1)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.FloatForTest(g, me, "R")
	effectstest.ApaActivate(t, g, me, gremlin, 0, game.ActivateAbilityParams{})
	for _, p := range g.Seats {
		want := lives[p.ID]
		if p.ID != me.ID {
			want--
		}
		if p.Life != want {
			t.Errorf("%s is at %d, want %d", p.Name, p.Life, want)
		}
	}
	if !effectstest.ApaLive(g, gremlin).Tapped {
		t.Fatal("the Gremlin did not tap")
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", gapB154Bolt, []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ApaLive(g, gremlin).Tapped {
		t.Error("a noncreature spell did not untap the Gremlin")
	}
}

func TestGapB154DragonheartBecomesAFourFourFlyingHastyDragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	heart := effectstest.B12Push(g, me.ID, "Bogardan Dragonheart", "Creature — Human Shaman", gapB154Dragonheart, 2, 2)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", gapB154Bear, 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, heart, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{heart}}); err == nil {
		t.Fatal("the Dragonheart sacrificed itself for 'another creature'")
	}
	effectstest.ApaActivate(t, g, me, heart, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	live := effectstest.ApaLive(g, heart)
	if live.CurrentPower() != 4 || live.CurrentToughness() != 4 || !live.HasSubtype("Dragon") || live.HasSubtype("Human") {
		t.Errorf("it is %d/%d (Dragon %v, Human %v), want a 4/4 Dragon", live.CurrentPower(), live.CurrentToughness(), live.HasSubtype("Dragon"), live.HasSubtype("Human"))
	}
	if !effectstest.P1g2HasKeyword(g, heart, "flying") || !effectstest.P1g2HasKeyword(g, heart, "haste") {
		t.Error("it lacks flying or haste")
	}
}

func TestGapB154KiboMakesBananasThatPayLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	kibo := effectstest.B12Push(g, me.ID, "Kibo, Uktabi Prince", "Legendary Creature — Monkey Noble", gapB154Kibo, 2, 2)
	effectstest.ApaActivate(t, g, me, kibo, 0, game.ActivateAbilityParams{})
	for _, p := range g.Seats {
		if n, _ := effectstest.CountTokensNamed(g, p.ID, "Banana"); n != 1 {
			t.Errorf("%s has %d Bananas, want 1", p.Name, n)
		}
	}
	var banana uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Banana" {
			banana = c.InstanceID
		}
	}
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, banana, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Banana for mana: %v", err)
	}
	effectstest.ResolveManaPickAnyColor(t, g, me.ID)
	if me.ManaPool.Total() != 1 || me.Life != life+2 {
		t.Errorf("pool %d, life %d → %d; want 1 mana and 2 life", me.ManaPool.Total(), life, me.Life)
	}
	if g.Battlefield.Contains(banana) {
		t.Error("the Banana was not sacrificed")
	}
}

func TestGapB154KiboGrowsApesAndMonkeysWhenAnOpponentsArtifactDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	kibo := effectstest.B12Push(g, me.ID, "Kibo, Uktabi Prince", "Legendary Creature — Monkey Noble", gapB154Kibo, 2, 2)
	ape := effectstest.B12Creature(g, me.ID, "Ape", "Creature — Ape", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", gapB154Bear, 2, 2)
	theirApe := effectstest.B12Creature(g, opp.ID, "Their Ape", "Creature — Ape", 3, 3)
	mine := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	gapB154Kill(g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, ape, game.CounterPlusOne); got != 0 {
		t.Errorf("my own artifact dying gave the Ape %d counters", got)
	}
	gapB154Kill(g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"Kibo": kibo, "the Ape": ape} {
		if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("%s has %d counters, want 1", name, got)
		}
	}
	for name, id := range map[string]uuid.UUID{"the Bear": bear, "their Ape": theirApe} {
		if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 0 {
			t.Errorf("%s has %d counters, want 0", name, got)
		}
	}
}

func TestGapB154KiboMakesTheDefenderSacrificeAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kibo := effectstest.B12Push(g, me.ID, "Kibo, Uktabi Prince", "Legendary Creature — Monkey Noble", gapB154Kibo, 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(kibo, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Chooser != opp.ID {
		t.Fatalf("the defending player was not asked to choose an artifact to sacrifice: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(g.PendingChoices[0].ID, opp.ID, []uuid.UUID{rock}); err != nil {
		t.Fatalf("answer the sacrifice: %v", err)
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the defending player kept their only artifact")
	}
}

func TestGapB154KenessosScriesOneMoreAndLooksForSeaCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	kenessos := effectstest.B12Push(g, me.ID, "Kenessos, Priest of Thassa", "Legendary Creature — Merfolk Cleric", gapB154Kenessos, 1, 3)
	g.WithWriteLock(func() { g.ScryKeyedForEffect(me.ID, uuid.Nil, 1, game.KeyedTail{}) })
	if len(g.PendingChoices) != 1 || len(g.PendingChoices[0].ScryCards) != 2 {
		t.Fatalf("scry 1 looked at %+v, want 2 cards", g.PendingChoices)
	}
	if err := g.ResolveScry(g.PendingChoices[0].ID, me.ID, g.PendingChoices[0].ScryCards, nil); err != nil {
		t.Fatal(err)
	}

	kraken := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Kraken", TypeLine: "Creature — Kraken", Power: 8, Toughness: 8})
	effectstest.FloatForTest(g, me, "UUUU")
	effectstest.ApaActivate(t, g, me, kenessos, 0, game.ActivateAbilityParams{})
	c := effectstest.LatestOptionPickFor(g, me.ID)
	if c == nil || len(c.PickOptions) != 3 {
		t.Fatalf("a Kraken on top should offer put / bottom / leave, got %+v", c)
	}
	if err := g.ResolveOptionPick(c.ID, me.ID, 0); err != nil {
		t.Fatal(err)
	}
	if !g.Battlefield.Contains(kraken) {
		t.Error("the Kraken did not enter the battlefield")
	}

	bear := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Bear", TypeLine: gapB154Bear, Power: 2, Toughness: 2})
	g.WithWriteLock(func() {
		me.Library.PushBottom(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	})
	effectstest.FloatForTest(g, me, "UUUU")
	effectstest.ApaActivate(t, g, me, kenessos, 0, game.ActivateAbilityParams{})
	c = effectstest.LatestOptionPickFor(g, me.ID)
	if c == nil || len(c.PickOptions) != 2 {
		t.Fatalf("a Bear on top should offer only bottom / leave, got %+v", c)
	}
	if err := g.ResolveOptionPick(c.ID, me.ID, 0); err != nil {
		t.Fatal(err)
	}
	if top, _ := me.Library.Top(); top.InstanceID == bear {
		t.Errorf("the Bear was not put on the bottom (library %d, pending %d)", me.Library.Size(), len(g.PendingChoices))
	}
}

func TestGapB154IsuPlaysSnowCardsFromTheTopAndGrowsWhenSnowEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	isu := effectstest.B12Push(g, me.ID, "Isu the Abominable", "Legendary Snow Creature — Yeti", gapB154Isu, 5, 5)

	plain := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	if err := g.CastSpell(me.ID, plain, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Fatal("played a nonsnow land from the top")
	}
	snow := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Snow-Covered Forest", TypeLine: "Basic Snow Land — Forest"})
	if err := g.CastSpell(me.ID, snow, game.CastSpellParams{FromZone: "library", Strict: true}); err != nil {
		t.Fatalf("play a snow land from the top: %v", err)
	}
	if !g.Battlefield.Contains(snow) {
		t.Fatal("the snow land did not enter")
	}
	// The snow land entering is a snow permanent entering: pay {G}.
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatForTest(g, me, "G")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, isu, game.CounterPlusOne); got != 1 {
		t.Errorf("Isu has %d counters after a snow permanent entered and {G} was paid, want 1", got)
	}
}

// gapB154Enter puts a card in me's hand and moves it onto the battlefield
// (it enters, so its enters triggers fire).
func gapB154Enter(t *testing.T, g *game.Game, me *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
			Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID})
	})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatalf("enter the battlefield: %v", err)
	}
	return id
}

// gapB154Kill destroys a permanent and settles the state-based actions.
func gapB154Kill(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{id}, game.DestroyOptions{}) })
	g.RunStateChecksForTest()
}
