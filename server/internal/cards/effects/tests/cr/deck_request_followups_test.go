package testscr

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	master := effectstest.B12Push(g, me.ID, "Master of the Wild Hunt", "Creature — Human Shaman", followMaster, 3, 3)
	a := effectstest.B12Creature(g, me.ID, "Wolf A", "Creature — Wolf", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Wolf B", "Creature — Wolf", 3, 3)
	tapped := effectstest.B12Creature(g, me.ID, "Tapped Wolf", "Creature — Wolf", 20, 20)
	noncreature := effectstest.B12Permanent(g, me.ID, "Wolf kindred", "Kindred Enchantment — Wolf")
	other := effectstest.B12Creature(g, opp.ID, "Other Wolf", "Creature — Wolf", 10, 10)
	target := effectstest.B12Creature(g, opp.ID, "Target", "Creature", 4, 5)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(tapped) })
	if err := g.ActivateCatalogAbility(me.ID, master, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}, Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.NumberPrompt(t, g, opp.ID); c.PayAmount.Max != 4 {
		t.Fatalf("return power: %+v", c.PayAmount)
	}
	if effectstest.GaLayered(t, g, target).DamageMarked != 5 {
		t.Fatal("Wolves must deal damage before the allocation; lethal target must still deal damage back")
	}
	if effectstest.GaLayered(t, g, noncreature).Tapped || effectstest.GaLayered(t, g, other).Tapped {
		t.Fatal("tapped a noncreature or opposing Wolf")
	}
	if !effectstest.GaLayered(t, g, a).Tapped || !effectstest.GaLayered(t, g, b).Tapped {
		t.Fatal("participating Wolves not tapped")
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	c := effectstest.NumberPrompt(t, g, opp.ID)
	if err := g.ResolvePayAmount(c.ID, me.ID, 2); err == nil {
		t.Fatal("activator chose opponent's allocation")
	}
	if err := g.ResolvePayAmount(c.ID, opp.ID, 2); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, target) || effectstest.BattlefieldHas(g, a) {
		t.Fatal("lethal target or Wolf survived")
	}
	if !effectstest.BattlefieldHas(g, b) || effectstest.GaLayered(t, g, b).DamageMarked != 2 {
		t.Fatal("second Wolf did not receive the remainder")
	}
	if !effectstest.BattlefieldHas(g, tapped) || effectstest.GaLayered(t, g, tapped).DamageMarked != 0 {
		t.Fatal("already tapped Wolf participated")
	}
}

func TestDeckFollowupMasterIllegalTargetDoesNotTapWolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	master := effectstest.B12Push(g, me.ID, "Master of the Wild Hunt", "Creature — Human Shaman", followMaster, 3, 3)
	wolf := effectstest.B12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	target := effectstest.PushVanillaCreature(g, opp.ID, "Target", 3, 3)
	if err := g.ActivateCatalogAbility(me.ID, master, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}, Strict: true}); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(target) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.GaLayered(t, g, wolf).Tapped {
		t.Fatal("fizzled ability tapped a Wolf")
	}
}

func TestDeckFollowupMasterUsesItsChangedCreatureType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	master := effectstest.B12Push(g, me.ID, "Master of the Wild Hunt", "Creature — Human Shaman", followMaster, 3, 3)
	elf := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 2, 3)
	wolf := effectstest.B12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 3)
	target := effectstest.PushVanillaCreature(g, opp.ID, "Target", 1, 5)
	effectstest.TwxEvolve(t, g, master, "Wolf", "Elf")
	if err := g.ActivateCatalogAbility(me.ID, master, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}, Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.GaLayered(t, g, elf).Tapped || effectstest.GaLayered(t, g, elf).DamageMarked != 1 || effectstest.GaLayered(t, g, wolf).Tapped || effectstest.GaLayered(t, g, target).DamageMarked != 2 {
		t.Fatal("changed Wolf word did not select Elves for both damage instructions")
	}
}

func TestDeckFollowupMasterUpkeepCreatesWolf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Master of the Wild Hunt", "Creature — Human Shaman", followMaster, 3, 3)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginUpkeep, Actor: me.ID, Step: game.StepUpkeep}) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Wolf") != 1 {
		t.Fatal("missing upkeep Wolf")
	}
}

func TestDeckFollowupSovereignTapsFiveSickZombiesAndStealsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	source := effectstest.B12Push(g, me.ID, "Gravespawn Sovereign", "Creature — Zombie", followSovereign, 3, 3)
	taps := []uuid.UUID{source}
	for i := 0; i < 3; i++ {
		taps = append(taps, effectstest.B12Creature(g, me.ID, "Zombie", "Creature — Zombie", 2, 2))
	}
	// Printed "Zombies", so a Kindred noncreature is legal too.
	taps = append(taps, effectstest.B12Permanent(g, me.ID, "Zombie kindred", "Kindred Enchantment — Zombie"))
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].SummonedThisTurn = true
		}
	})
	target := effectstest.GyCardWithOracle(opp, "Stolen creature", "Creature — Human", "")
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
		if !effectstest.GaLayered(t, g, id).Tapped {
			t.Fatal("Zombie not tapped as cost")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.GaLayered(t, g, target)
	if c.Controller != me.ID || c.Owner != opp.ID {
		t.Fatal("reanimation did not give activator control")
	}
}

func TestDeckFollowupGempalmCyclingTargetsAndCountsAtResolution(t *testing.T) {
	for _, yes := range []bool{false, true} {
		t.Run(map[bool]string{false: "decline", true: "accept"}[yes], func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.B12Creature(g, me.ID, "Zombie", "Creature — Zombie", 2, 2)
			effectstest.B12Creature(g, opp.ID, "Opposing Zombie", "Creature — Zombie", 2, 2)
			_, _ = effectstest.CycleFromHand(t, g, "Gempalm Polluter", "Creature — Zombie", followPolluter, "{B}{B}")
			pick := effectstest.LatestPickTarget(g, me.ID)
			if pick == nil {
				t.Fatal("cycling did not trigger targeted life loss")
			}
			if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
				t.Fatal(err)
			}
			effectstest.B12Permanent(g, opp.ID, "Zombie kindred", "Kindred Enchantment — Zombie")
			life, hand := opp.Life, me.Hand.Size()
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
			effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	source := effectstest.B12Push(g, me.ID, "Smothering Abomination", "Creature — Eldrazi", followAbomination, 4, 3)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 1, 1)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 1, 1)
	artifact := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	before := me.Hand.Size()
	dead := effectstest.PushVanillaCreature(g, me.ID, "Destroyed, not sacrificed", 1, 1)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(dead); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{theirs, artifact, mine, source} {
		g.WithWriteLock(func() {
			if err := g.SacrificePermanentForEffect(id); err != nil {
				t.Fatal(err)
			}
		})
		effectstest.PassPriorityAroundTable(t, g)
	}
	if me.Hand.Size() != before+2 {
		t.Fatalf("drew %d, want own creature and itself", me.Hand.Size()-before)
	}
}

func TestDeckFollowupAbominationUpkeepMaySacrificeItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	source := effectstest.B12Push(g, me.ID, "Smothering Abomination", "Creature — Eldrazi", followAbomination, 4, 3)
	other := effectstest.PushVanillaCreature(g, me.ID, "Other", 1, 1)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginUpkeep, Actor: me.ID, Step: game.StepUpkeep}) })
	before := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, source)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, source) || !effectstest.BattlefieldHas(g, other) || me.Hand.Size() != before+1 {
		t.Fatal("upkeep must allow itself and trigger the draw")
	}
}

func TestDeckFollowupHowlingMoonCombatPumpAndSecondSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedPermanentWithOracle(g, me.ID, "Howling Moon", "Enchantment", followMoon)
	wolf := effectstest.B12Creature(g, me.ID, "Werewolf", "Creature — Human Werewolf", 3, 3)
	wrong := effectstest.PushVanillaCreature(g, me.ID, "Human", 2, 2)
	otherWolf := effectstest.B12Creature(g, opp.ID, "Opposing Wolf", "Creature — Wolf", 2, 2)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventStepBegan, Actor: me.ID, Step: game.StepBeginCombat}) })
	c := effectstest.LatestPickTarget(g, me.ID)
	if c == nil {
		t.Fatal("missing combat target")
	}
	for _, id := range []uuid.UUID{wrong, otherWolf} {
		if err := g.ResolvePickTargets(c.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: id}}); err == nil {
			t.Fatal("accepted a creature outside the controlled Wolf/Werewolf clause")
		}
	}
	effectstest.AnswerTriggerTargets(t, g, me.ID, wolf)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.GaLayered(t, g, wolf).CurrentPower() != 5 {
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
		effectstest.PassPriorityAroundTable(t, g)
	}
	if effectstest.CountBattlefieldNamed(g, me.ID, "Wolf") != 1 {
		t.Fatal("must create exactly one Wolf for opponent's second spell")
	}
}

func TestDeckFollowupUlrichTransformsAndFightsOnlyNonWerewolves(t *testing.T) {
	w := followUlrichCard
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	source := effectstest.IwPut(t, g, w, me)
	target := effectstest.B12Creature(g, opp.ID, "Target", "Creature — Human", 2, 3)
	illegal := effectstest.B12Creature(g, opp.ID, "Werewolf", "Creature — Werewolf", 2, 3)
	effectstest.IwTransform(t, g, source)
	c := effectstest.LatestPickTarget(g, me.ID)
	if c == nil {
		t.Fatal("no transform fight target")
	}
	if err := g.ResolvePickTargets(c.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: illegal}}); err == nil {
		t.Fatal("fight targeted Werewolf")
	}
	effectstest.AnswerTriggerTargets(t, g, me.ID, target)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, target) || !effectstest.BattlefieldHas(g, source) || effectstest.GaLayered(t, g, source).DamageMarked != 2 {
		t.Fatal("transform fight incorrect")
	}
	effectstest.IwTransform(t, g, source)
	effectstest.AnswerTriggerTargets(t, g, me.ID, source)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.GaLayered(t, g, source).CurrentPower() != 8 {
		t.Fatal("front transform did not grant +4/+4")
	}
}

var followUlrichCard = effectstest.IwCard{Oracle: followUlrich, Front: "Ulrich of the Krallenhorde", FrontType: "Legendary Creature — Human Werewolf", Cost: "{3}{R}{G}", FrontText: "Whenever this creature enters or transforms into Ulrich of the Krallenhorde, target creature gets +4/+4 until end of turn.\n" + effectstest.IwFrontText, Fp: "4", Ft: "4", Back: "Ulrich, Uncontested Alpha", BackText: effectstest.IwBackText, Bp: "6", Bt: "6", Colors: []string{"R", "G"}}

func TestDeckFollowupUlrichEntersAndTransformsAtRealUpkeeps(t *testing.T) {
	g, me := effectstest.IwTable(t)
	source := effectstest.ImportAndCast(t, g, followUlrichCard.Row(), me)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, source)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.GaLayered(t, g, source).CurrentPower() != 8 {
		t.Fatal("entry did not pump Ulrich")
	}
	// No opposing creatures means the back's required target cannot be
	// chosen; its optional fight must not leave a stuck question.
	effectstest.IwNextUpkeep(t, g, nil)
	if effectstest.IwFace(g, source) != 1 || len(g.PendingChoices) != 0 {
		t.Fatal("quiet turn did not transform cleanly to the back")
	}
	effectstest.IwNextUpkeep(t, g, map[int]int{g.Turn.ActiveSeat: 2})
	effectstest.AnswerTriggerTargets(t, g, me.ID, source)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IwFace(g, source) != 0 || effectstest.GaLayered(t, g, source).CurrentPower() != 8 {
		t.Fatal("two spells did not transform and pump the front")
	}
}
