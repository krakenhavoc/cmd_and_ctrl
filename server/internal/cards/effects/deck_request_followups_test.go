package effects

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	followMoon        = "a814d4c7-2ec0-4278-8704-23b39dcead4e"
	followMaster      = "dd25cab9-15cc-40f5-9b68-8f345dd7e952"
	followUlrich      = "0d397c05-a680-4274-972f-6a5f778b5133"
	followPolluter    = "bd532f8d-789f-4ee1-a2f3-ae30e8282c76"
	followSovereign   = "4235c99b-50de-4190-8938-3ce5ec5c5248"
	followAbomination = "c3db28e5-6162-4642-a1c9-a7ace0737950"
)

func TestDeckFollowupMasterDividesByTargetsControllerAndRestores(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	master := b12Push(g, me.ID, "Master of the Wild Hunt", "Creature — Human Shaman", followMaster, 3, 3)
	a := b12Creature(g, me.ID, "Wolf A", "Creature — Wolf", 2, 2)
	b := b12Creature(g, me.ID, "Wolf B", "Creature — Wolf", 3, 3)
	tapped := b12Creature(g, me.ID, "Tapped Wolf", "Creature — Wolf", 20, 20)
	noncreature := b12Permanent(g, me.ID, "Wolf kindred", "Kindred Enchantment — Wolf")
	other := b12Creature(g, opp.ID, "Other Wolf", "Creature — Wolf", 10, 10)
	target := b12Creature(g, opp.ID, "Target", "Creature", 4, 5)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(tapped) })
	if err := g.ActivateCatalogAbility(me.ID, master, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}, Strict: true}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if c := numberPrompt(t, g, opp.ID); c.PayAmount.Max != 4 {
		t.Fatalf("return power: %+v", c.PayAmount)
	}
	if gaLayered(t, g, target).DamageMarked != 5 {
		t.Fatal("Wolves must deal damage before the allocation; lethal target must still deal damage back")
	}
	if gaLayered(t, g, noncreature).Tapped || gaLayered(t, g, other).Tapped {
		t.Fatal("tapped a noncreature or opposing Wolf")
	}
	if !gaLayered(t, g, a).Tapped || !gaLayered(t, g, b).Tapped {
		t.Fatal("participating Wolves not tapped")
	}
	g = restoreRoundTrip(t, g, true)
	c := numberPrompt(t, g, opp.ID)
	if err := g.ResolvePayAmount(c.ID, me.ID, 2); err == nil {
		t.Fatal("activator chose opponent's allocation")
	}
	if err := g.ResolvePayAmount(c.ID, opp.ID, 2); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, target) || battlefieldHas(g, a) {
		t.Fatal("lethal target or Wolf survived")
	}
	if !battlefieldHas(g, b) || gaLayered(t, g, b).DamageMarked != 2 {
		t.Fatal("second Wolf did not receive the remainder")
	}
	if !battlefieldHas(g, tapped) || gaLayered(t, g, tapped).DamageMarked != 0 {
		t.Fatal("already tapped Wolf participated")
	}
}

func TestDeckFollowupMasterIllegalTargetDoesNotTapWolves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	master := b12Push(g, me.ID, "Master of the Wild Hunt", "Creature — Human Shaman", followMaster, 3, 3)
	wolf := b12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	target := pushVanillaCreature(g, opp.ID, "Target", 3, 3)
	if err := g.ActivateCatalogAbility(me.ID, master, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}, Strict: true}); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(target) })
	passPriorityAroundTable(t, g)
	if gaLayered(t, g, wolf).Tapped {
		t.Fatal("fizzled ability tapped a Wolf")
	}
}

func TestDeckFollowupMasterUsesItsChangedCreatureType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	master := b12Push(g, me.ID, "Master of the Wild Hunt", "Creature — Human Shaman", followMaster, 3, 3)
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf", 2, 3)
	wolf := b12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 3)
	target := pushVanillaCreature(g, opp.ID, "Target", 1, 5)
	twxEvolve(t, g, master, "Wolf", "Elf")
	if err := g.ActivateCatalogAbility(me.ID, master, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}, Strict: true}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !gaLayered(t, g, elf).Tapped || gaLayered(t, g, elf).DamageMarked != 1 || gaLayered(t, g, wolf).Tapped || gaLayered(t, g, target).DamageMarked != 2 {
		t.Fatal("changed Wolf word did not select Elves for both damage instructions")
	}
}

func TestDeckFollowupMasterUpkeepCreatesWolf(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Master of the Wild Hunt", "Creature — Human Shaman", followMaster, 3, 3)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginUpkeep, Actor: me.ID, Step: game.StepUpkeep}) })
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Wolf") != 1 {
		t.Fatal("missing upkeep Wolf")
	}
}

func TestDeckFollowupSovereignTapsFiveSickZombiesAndStealsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	source := b12Push(g, me.ID, "Gravespawn Sovereign", "Creature — Zombie", followSovereign, 3, 3)
	taps := []uuid.UUID{source}
	for i := 0; i < 3; i++ {
		taps = append(taps, b12Creature(g, me.ID, "Zombie", "Creature — Zombie", 2, 2))
	}
	// Printed "Zombies", so a Kindred noncreature is legal too.
	taps = append(taps, b12Permanent(g, me.ID, "Zombie kindred", "Kindred Enchantment — Zombie"))
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].SummonedThisTurn = true
		}
	})
	target := gyCardWithOracle(opp, "Stolen creature", "Creature — Human", "")
	opp.Graveyard.Cards[len(opp.Graveyard.Cards)-1].Power = 2
	opp.Graveyard.Cards[len(opp.Graveyard.Cards)-1].Toughness = 2
	params := game.ActivateAbilityParams{TapIDs: taps[:4], Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}, Strict: true}
	if err := g.ActivateCatalogAbility(me.ID, source, 0, params); err == nil {
		t.Fatal("four Zombies paid five cost")
	}
	params.TapIDs = taps
	if err := g.ActivateCatalogAbility(me.ID, source, 0, params); err != nil {
		t.Fatal(err)
	}
	for _, id := range taps {
		if !gaLayered(t, g, id).Tapped {
			t.Fatal("Zombie not tapped as cost")
		}
	}
	passPriorityAroundTable(t, g)
	c := gaLayered(t, g, target)
	if c.Controller != me.ID || c.Owner != opp.ID {
		t.Fatal("reanimation did not give activator control")
	}
}

func TestDeckFollowupGempalmCyclingTargetsAndCountsAtResolution(t *testing.T) {
	for _, yes := range []bool{false, true} {
		t.Run(map[bool]string{false: "decline", true: "accept"}[yes], func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			b12Creature(g, me.ID, "Zombie", "Creature — Zombie", 2, 2)
			b12Creature(g, opp.ID, "Opposing Zombie", "Creature — Zombie", 2, 2)
			_, _ = cycleFromHand(t, g, "Gempalm Polluter", "Creature — Zombie", followPolluter, "{B}{B}")
			pick := latestPickTarget(g, me.ID)
			if pick == nil {
				t.Fatal("cycling did not trigger targeted life loss")
			}
			if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
				t.Fatal(err)
			}
			b12Permanent(g, opp.ID, "Zombie kindred", "Kindred Enchantment — Zombie")
			life, hand := opp.Life, me.Hand.Size()
			answerLatestTriggerPrompt(t, g, me.ID, yes)
			passPriorityAroundTable(t, g)
			want := life
			if yes {
				want -= 3
			}
			if opp.Life != want {
				t.Fatalf("life %d, want %d", opp.Life, want)
			}
			if me.Hand.Size() != hand+1 {
				t.Fatal("cycling did not draw independently")
			}
		})
	}
}

func TestDeckFollowupAbominationDrawsOnlyForOwnCreatureSacrifices(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	source := b12Push(g, me.ID, "Smothering Abomination", "Creature — Eldrazi", followAbomination, 4, 3)
	mine := pushVanillaCreature(g, me.ID, "Mine", 1, 1)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 1, 1)
	artifact := b12Permanent(g, me.ID, "Rock", "Artifact")
	before := me.Hand.Size()
	dead := pushVanillaCreature(g, me.ID, "Destroyed, not sacrificed", 1, 1)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(dead); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{theirs, artifact, mine, source} {
		g.WithWriteLock(func() {
			if err := g.SacrificePermanentForEffect(id); err != nil {
				t.Fatal(err)
			}
		})
		passPriorityAroundTable(t, g)
	}
	if me.Hand.Size() != before+2 {
		t.Fatalf("drew %d, want own creature and itself", me.Hand.Size()-before)
	}
}

func TestDeckFollowupAbominationUpkeepMaySacrificeItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	source := b12Push(g, me.ID, "Smothering Abomination", "Creature — Eldrazi", followAbomination, 4, 3)
	other := pushVanillaCreature(g, me.ID, "Other", 1, 1)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginUpkeep, Actor: me.ID, Step: game.StepUpkeep}) })
	before := me.Hand.Size()
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, me.ID, source)
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, source) || !battlefieldHas(g, other) || me.Hand.Size() != before+1 {
		t.Fatal("upkeep must allow itself and trigger the draw")
	}
}

func TestDeckFollowupHowlingMoonCombatPumpAndSecondSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedPermanentWithOracle(g, me.ID, "Howling Moon", "Enchantment", followMoon)
	wolf := b12Creature(g, me.ID, "Werewolf", "Creature — Human Werewolf", 3, 3)
	wrong := pushVanillaCreature(g, me.ID, "Human", 2, 2)
	otherWolf := b12Creature(g, opp.ID, "Opposing Wolf", "Creature — Wolf", 2, 2)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventStepBegan, Actor: me.ID, Step: game.StepBeginCombat}) })
	c := latestPickTarget(g, me.ID)
	if c == nil {
		t.Fatal("missing combat target")
	}
	for _, id := range []uuid.UUID{wrong, otherWolf} {
		if err := g.ResolvePickTargets(c.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: id}}); err == nil {
			t.Fatal("accepted a creature outside the controlled Wolf/Werewolf clause")
		}
	}
	answerTriggerTargets(t, g, me.ID, wolf)
	passPriorityAroundTable(t, g)
	if gaLayered(t, g, wolf).CurrentPower() != 5 {
		t.Fatal("missing combat +2/+2")
	}
	for _, tc := range []struct {
		player uuid.UUID
		n      int
	}{{opp.ID, 1}, {me.ID, 2}, {opp.ID, 2}, {opp.ID, 3}} {
		g.WithWriteLock(func() {
			if g.SpellsCastThisTurn == nil {
				g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{}
			}
			g.SpellsCastThisTurn[tc.player] = game.CastTally{Total: tc.n}
			g.EmitEvent(game.Event{Kind: game.EventCast, Actor: tc.player, CardID: uuid.New()})
		})
		passPriorityAroundTable(t, g)
	}
	if countBattlefieldNamed(g, me.ID, "Wolf") != 1 {
		t.Fatal("must create exactly one Wolf for opponent's second spell")
	}
}

func TestDeckFollowupUlrichTransformsAndFightsOnlyNonWerewolves(t *testing.T) {
	w := followUlrichCard
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	source := iwPut(t, g, w, me)
	target := b12Creature(g, opp.ID, "Target", "Creature — Human", 2, 3)
	illegal := b12Creature(g, opp.ID, "Werewolf", "Creature — Werewolf", 2, 3)
	iwTransform(t, g, source)
	c := latestPickTarget(g, me.ID)
	if c == nil {
		t.Fatal("no transform fight target")
	}
	if err := g.ResolvePickTargets(c.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: illegal}}); err == nil {
		t.Fatal("fight targeted Werewolf")
	}
	answerTriggerTargets(t, g, me.ID, target)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, target) || !battlefieldHas(g, source) || gaLayered(t, g, source).DamageMarked != 2 {
		t.Fatal("transform fight incorrect")
	}
	iwTransform(t, g, source)
	answerTriggerTargets(t, g, me.ID, source)
	passPriorityAroundTable(t, g)
	if gaLayered(t, g, source).CurrentPower() != 8 {
		t.Fatal("front transform did not grant +4/+4")
	}
}

var followUlrichCard = iwCard{oracle: followUlrich, front: "Ulrich of the Krallenhorde", frontType: "Legendary Creature — Human Werewolf", cost: "{3}{R}{G}", frontText: "Whenever this creature enters or transforms into Ulrich of the Krallenhorde, target creature gets +4/+4 until end of turn.\n" + iwFrontText, fp: "4", ft: "4", back: "Ulrich, Uncontested Alpha", backText: iwBackText, bp: "6", bt: "6", colors: []string{"R", "G"}}

func TestDeckFollowupUlrichEntersAndTransformsAtRealUpkeeps(t *testing.T) {
	g, me := iwTable(t)
	source := importAndCast(t, g, followUlrichCard.row(), me)
	passPriorityAroundTable(t, g)
	answerTriggerTargets(t, g, me.ID, source)
	passPriorityAroundTable(t, g)
	if gaLayered(t, g, source).CurrentPower() != 8 {
		t.Fatal("entry did not pump Ulrich")
	}
	// No opposing creatures means the back's required target cannot be
	// chosen; its optional fight must not leave a stuck question.
	iwNextUpkeep(t, g, nil)
	if iwFace(g, source) != 1 || len(g.PendingChoices) != 0 {
		t.Fatal("quiet turn did not transform cleanly to the back")
	}
	iwNextUpkeep(t, g, map[int]int{g.Turn.ActiveSeat: 2})
	answerTriggerTargets(t, g, me.ID, source)
	passPriorityAroundTable(t, g)
	if iwFace(g, source) != 0 || gaLayered(t, g, source).CurrentPower() != 8 {
		t.Fatal("two spells did not transform and pump the front")
	}
}
