package testsru

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b148_test.go — GAP batch 148 (#5948): Jumpstart 2022 creatures
// Dread Rider, Dread Slaver, Ghoulraiser, Hooded Assassin, Kalastria
// Nightwatch and Kraul Swarm.

const (
	gap148Rider     = "e5c15ef5-6795-40c2-a131-a73f74142314"
	gap148Slaver    = "8110f092-41b0-4e53-a7d7-f7e3cbc4a52a"
	gap148Ghoul     = "614f748b-5d31-440d-a7c0-3ba12dc63c24"
	gap148Assassin  = "a3f76c99-b608-425e-83ca-675acb7c2693"
	gap148Nightwach = "f610748a-8b2c-4c7d-b495-b4677ff4c423"
	gap148Swarm     = "ac97d8fc-15fa-4a60-96d3-4c09deafa83a"
)

func TestGap148EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gap148Rider: "Dread Rider", gap148Slaver: "Dread Slaver", gap148Ghoul: "Ghoulraiser",
		gap148Assassin: "Hooded Assassin", gap148Nightwach: "Kalastria Nightwatch", gap148Swarm: "Kraul Swarm",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func gap148GraveyardCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID, Power: 2, Toughness: 2})
	return id
}

func TestGap148GhoulraiserReturnsOnlyAZombieCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := gap148GraveyardCard(me, "Dead Bear", "Creature — Bear")
	z1 := gap148GraveyardCard(me, "Dead Zombie", "Creature — Zombie")
	z2 := gap148GraveyardCard(me, "Other Zombie", "Creature — Zombie Knight")
	effectstest.CastAndResolveCreature(t, g, "Ghoulraiser", "Creature — Zombie", gap148Ghoul)
	effectstest.PassPriorityAroundTable(t, g)
	inHand := 0
	for _, id := range []uuid.UUID{z1, z2} {
		if me.Hand.Contains(id) {
			inHand++
		}
	}
	if inHand != 1 {
		t.Fatalf("%d Zombie cards returned, want exactly 1", inHand)
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the non-Zombie creature card left the graveyard")
	}
	if me.Graveyard.Size() != 2 {
		t.Errorf("graveyard size %d, want 2", me.Graveyard.Size())
	}
}

func TestGap148GhoulraiserWithNoZombieDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := gap148GraveyardCard(me, "Dead Bear", "Creature — Bear")
	hand := me.Hand.Size()
	effectstest.CastAndResolveCreature(t, g, "Ghoulraiser", "Creature — Zombie", gap148Ghoul)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(bear) || me.Hand.Size() != hand {
		t.Errorf("a card moved with no Zombie in the graveyard: hand %d, want %d", me.Hand.Size(), hand)
	}
}

func TestGap148NightwatchFliesUntilEndOfTurnWhenYouGainLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	watch := effectstest.PushCatalogPermanent(g, me.ID, "Kalastria Nightwatch", "Creature — Vampire Warrior Ally", gap148Nightwach, false)
	if effectstest.HasKeywordOnBattlefield(t, g, watch, "flying") {
		t.Fatal("flying before any life was gained")
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 3) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasKeywordOnBattlefield(t, g, watch, "flying") {
		t.Fatal("an opponent's life gain gave flying")
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, watch, "flying") {
		t.Fatal("no flying after gaining life")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if effectstest.HasKeywordOnBattlefield(t, g, watch, "flying") {
		t.Error("flying lasted past the end of the turn")
	}
}

func TestGap148KraulSwarmReturnsFromTheGraveyardForACreatureDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	swarm := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: swarm, Name: "Kraul Swarm", TypeLine: "Creature — Insect Warrior", OracleID: gap148Swarm,
		Owner: me.ID, Controller: me.ID, Power: 4, Toughness: 1})
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Not a Creature", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})

	effectstest.FloatForTest(g, me, "BCC")
	if err := g.ActivateCatalogAbility(me.ID, swarm, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{spell}}); err == nil {
		t.Fatal("discarded a noncreature card to pay")
	}
	if err := g.ActivateCatalogAbility(me.ID, swarm, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without discarding")
	}
	if err := g.ActivateCatalogAbility(me.ID, swarm, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the creature card was not discarded as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(swarm) {
		t.Error("Kraul Swarm did not return to hand")
	}
}

func TestGap148DreadRiderDrainsThreeForAGraveyardCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	rider := effectstest.PushCatalogPermanent(g, me.ID, "Dread Rider", "Creature — Spirit Knight", gap148Rider, false)
	spell := gap148GraveyardCard(me, "Spent Spell", "Sorcery")
	bear := gap148GraveyardCard(me, "Dead Bear", "Creature — Bear")
	target := game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}

	effectstest.FloatForTest(g, me, "BC")
	target.ExileIDs = []uuid.UUID{spell}
	if err := g.ActivateCatalogAbility(me.ID, rider, 0, target); err == nil {
		t.Fatal("a sorcery paid \"exile a creature card\"")
	}
	self := game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{bear}, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}
	if err := g.ActivateCatalogAbility(me.ID, rider, 0, self); err == nil {
		t.Fatal("targeted its own controller, not an opponent")
	}
	target.ExileIDs = []uuid.UUID{bear}
	meLife, oppLife := me.Life, opp.Life
	if err := g.ActivateCatalogAbility(me.ID, rider, 0, target); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Exile.Contains(bear) || !me.Graveyard.Contains(spell) {
		t.Fatal("the cost did not exile exactly the creature card")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-3 || me.Life != meLife {
		t.Errorf("life me %d opp %d, want %d and %d", me.Life, opp.Life, meLife, oppLife-3)
	}
	again := game.ActivateAbilityParams{Strict: true, Targets: target.Targets}
	effectstest.FloatForTest(g, me, "BC")
	if err := g.ActivateCatalogAbility(me.ID, rider, 0, again); err == nil {
		t.Error("activated again while tapped and with no creature card to exile")
	}
}

func gap148AssassinEnters(t *testing.T, g *game.Game, me *game.Player) *game.PendingChoice {
	t.Helper()
	effectstest.CastAndResolveCreature(t, g, "Hooded Assassin", "Creature — Human Assassin", gap148Assassin)
	mode := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if mode == nil {
		t.Fatalf("no mode prompt from the enters trigger: %+v", g.PendingChoices)
	}
	return mode
}

func TestGap148HoodedAssassinCounterMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mode := gap148AssassinEnters(t, g, me)
	effectstest.AnswerMode(t, g, mode, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	var id uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.OracleID == gap148Assassin {
			id = c.InstanceID
		}
	}
	if id == uuid.Nil {
		t.Fatal("the Assassin did not enter")
	}
	if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

func TestGap148HoodedAssassinDestroysOnlyADamagedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	hurt := effectstest.B12Creature(g, opp.ID, "Hurt Ogre", "Creature — Ogre", 3, 5)
	fresh := effectstest.B12Creature(g, opp.ID, "Fresh Ogre", "Creature — Ogre", 3, 5)
	effectstest.Pr6Damage(t, g, fresh, hurt, 1)

	mode := gap148AssassinEnters(t, g, me)
	effectstest.AnswerMode(t, g, mode, me.ID, 1)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: fresh}); err == nil {
		t.Fatal("an undamaged creature was accepted as the target")
	}
	effectstest.PickCard(t, g, me.ID, hurt)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hurt) {
		t.Error("the damaged creature survived")
	}
	if !g.Battlefield.Contains(fresh) {
		t.Error("the undamaged creature was destroyed")
	}
}

func TestGap148DreadSlaverReturnsWhatItDamagedAsABlackZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	slaver := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Dread Slaver", TypeLine: "Creature — Zombie Horror",
		OracleID: gap148Slaver, Power: 3, Toughness: 5, Owner: me.ID, Controller: me.ID, Colors: []string{"B"}})
	victim := effectstest.B12Creature(g, opp.ID, "Doomed Elf", "Creature — Elf", 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Other Elf", "Creature — Elf", 1, 1)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Elf", 1, 1)

	effectstest.Pr6Damage(t, g, slaver, victim, 2)
	effectstest.Pr6Damage(t, g, other, bystander, 1)
	effectstest.PassPriorityAroundTable(t, g)

	c, ok := effectstest.BattlefieldCard(g, victim)
	if !ok {
		t.Fatal("the creature the Slaver damaged did not come back")
	}
	if c.Controller != me.ID {
		t.Errorf("controller = %v, want the Slaver's controller", c.Controller)
	}
	eff := effectstest.LayeredCard(t, g, victim).Effective()
	if !slices.Contains(eff.Subtypes, "Zombie") || !slices.Contains(eff.Subtypes, "Elf") || !slices.Contains(eff.Colors, "B") {
		t.Errorf("returned creature has subtypes %v colors %v, want a black Elf Zombie", eff.Subtypes, eff.Colors)
	}
	if g.Battlefield.Contains(bystander) || !opp.Graveyard.Contains(bystander) {
		t.Error("a creature damaged by something else came back")
	}
	if !g.Battlefield.Contains(slaver) {
		t.Error("the Slaver left the battlefield")
	}
}

func TestGap148DreadSlaverIgnoresACreatureItDidNotDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Dread Slaver", TypeLine: "Creature — Zombie Horror",
		OracleID: gap148Slaver, Power: 3, Toughness: 5, Owner: me.ID, Controller: me.ID})
	victim := effectstest.B12Creature(g, opp.ID, "Doomed Elf", "Creature — Elf", 2, 2)
	effectstest.KillCreature(t, g, opp.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || !opp.Graveyard.Contains(victim) {
		t.Error("a creature the Slaver never damaged came back")
	}
}
