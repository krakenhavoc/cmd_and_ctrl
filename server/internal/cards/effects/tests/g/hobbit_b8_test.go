package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4454, The Hobbit (HOB / HOC) cards, half B, batch 8. Every test drives
// the registered catalog row through the engine.

// hobB8Opponents is every seat but `me`.
func hobB8Opponents(g *game.Game, me *game.Player) []*game.Player {
	var out []*game.Player
	for _, p := range g.Seats {
		if p.ID != me.ID {
			out = append(out, p)
		}
	}
	return out
}

// hobB8Dwarf puts a vanilla Dwarf of `owner`'s onto the battlefield.
func hobB8Dwarf(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.HobB4Creature(g, owner, "Dwarf Guard", "Creature — Dwarf Warrior", 2, 2)
}

// --- Dáin of the Ancient Halls ----------------------------------------

func TestHobDainDamagesEachOpponentByTheDwarvesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	dain := effectstest.HobPush(t, g, me.ID, "Dáin of the Ancient Halls", "Legendary Creature — Dwarf Noble", 4, 5)
	hobB8Dwarf(g, me.ID)
	hobB8Dwarf(g, me.ID)
	effectstest.HobB4Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	opps := hobB8Opponents(g, me)
	hobB8Dwarf(g, opps[0].ID) // an opponent's Dwarf is not counted
	life := map[uuid.UUID]int{}
	for _, p := range opps {
		life[p.ID] = p.Life
	}
	effectstest.DeclareAttack(t, g, opps[0].ID, dain)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range opps {
		if p.Life != life[p.ID]-3 {
			t.Errorf("%s: life %d -> %d, want 3 damage (Dáin and two other Dwarves)", p.Name, life[p.ID], p.Life)
		}
	}
}

func TestHobDainIsVigilantAndHasteful(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	dain := hobCast(t, g, "Dáin of the Ancient Halls", "Legendary Creature — Dwarf Noble", 4, 5)
	opp := hobB8Opponents(g, me)[0]
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, dain)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d -> %d, want 1 damage (Dáin alone) the turn he arrives: haste", life, opp.Life)
	}
	if c, ok := effectstest.BattlefieldCard(g, dain); !ok || c.Tapped {
		t.Error("Dáin attacked and tapped: he has vigilance")
	}
}

// --- Thorin, King of Durin's Folk -------------------------------------

func TestHobThorinMakesATreasureWhenHeOrAnotherDwarfEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	hobCast(t, g, "Thorin, King of Durin's Folk", "Legendary Creature — Dwarf Noble", 4, 4)
	if n := effectstest.BattlefieldCountNamed(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("%d Treasures after Thorin entered, want 1", n)
	}
	effectstest.RfdCast(t, g, "Dwarf Guard", "Creature — Dwarf Warrior", "", 2, 2)
	if n := effectstest.BattlefieldCountNamed(g, me.ID, "Treasure"); n != 2 {
		t.Errorf("%d Treasures after another Dwarf entered, want 2", n)
	}
	effectstest.RfdCast(t, g, "Bear", "Creature — Bear", "", 2, 2)
	if n := effectstest.BattlefieldCountNamed(g, me.ID, "Treasure"); n != 2 {
		t.Errorf("%d Treasures after a Bear entered, want still 2", n)
	}
}

func TestHobThorinPumpsOtherDwarvesPerArtifactToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	opp := hobB8Opponents(g, me)[0]
	thorin := effectstest.HobPush(t, g, me.ID, "Thorin, King of Durin's Folk", "Legendary Creature — Dwarf Noble", 4, 4)
	dwarf := hobB8Dwarf(g, me.ID)
	bear := effectstest.HobB4Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := hobB8Dwarf(g, opp.ID)
	if p := effectstest.EffectivePower(t, g, dwarf); p != 2 {
		t.Fatalf("the Dwarf is %d/x with no artifact tokens, want 2", p)
	}
	effectstest.PushTreasure(g, me)
	effectstest.PushTreasure(g, me)
	effectstest.PushTreasure(g, opp)                        // an opponent's token does not count
	effectstest.B12Permanent(g, me.ID, "Relic", "Artifact") // a nontoken artifact does not count
	for name, c := range map[string]struct {
		id   uuid.UUID
		want int
	}{
		"another Dwarf":     {dwarf, 4},
		"Thorin himself":    {thorin, 4},
		"a Bear":            {bear, 2},
		"an opponent Dwarf": {theirs, 2},
	} {
		if got := effectstest.EffectivePower(t, g, c.id); got != c.want {
			t.Errorf("%s has power %d, want %d", name, got, c.want)
		}
	}
	if tg := effectstest.EffectiveToughness(t, g, dwarf); tg != 2 {
		t.Errorf("the Dwarf's toughness is %d, want 2: the bonus is +1/+0", tg)
	}
}

// --- Gandalf, Wandering Wizard ----------------------------------------

func TestHobGandalfShufflesIntoTheLibraryAndDrawsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	gandalf := effectstest.HobPush(t, g, me.ID, "Gandalf, Wandering Wizard", "Legendary Creature — Avatar Wizard", 4, 5)
	effectstest.AdvanceToMain(t, g)
	hand, lib := me.Hand.Size(), me.Library.Size()
	if err := effectstest.HobB4Activate(t, g, me.ID, gandalf, 0, "{C}{C}{C}{C}{C}", nil); err == nil {
		t.Fatal("activated {6} with only five mana")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, gandalf, 0, "{C}", nil); err != nil {
		t.Fatalf("activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(gandalf) {
		t.Error("Gandalf is still on the battlefield")
	}
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d cards, want 3", got)
	}
	// Gandalf joined the library (+1), then three cards left it.
	if got := lib + 1 - 3; me.Library.Size() != got {
		t.Errorf("library %d, want %d", me.Library.Size(), got)
	}
}

func TestHobGandalfsOwnerDrawsNotTheActivator(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	opp := hobB8Opponents(g, me)[0]
	// I own Gandalf but an opponent controls him.
	gandalf := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Gandalf, Wandering Wizard", TypeLine: "Legendary Creature — Avatar Wizard",
		OracleID: effectstest.HobOracle(t, "Gandalf, Wandering Wizard"), Power: 4, Toughness: 5, Owner: me.ID, Controller: opp.ID,
	})
	effectstest.AdvanceToMain(t, g)
	mine, theirs := me.Hand.Size(), opp.Hand.Size()
	if err := effectstest.HobB4Activate(t, g, opp.ID, gandalf, 0, "{C}{C}{C}{C}{C}{C}", nil); err != nil {
		t.Fatalf("activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - mine; got != 3 {
		t.Errorf("the owner drew %d cards, want 3", got)
	}
	if got := opp.Hand.Size() - theirs; got != 0 {
		t.Errorf("the controller drew %d cards, want none", got)
	}
	if !me.Library.Contains(gandalf) && g.Battlefield.Contains(gandalf) {
		t.Error("Gandalf stayed on the battlefield")
	}
}

func TestHobGandalfLeftInResponseStillDrawsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	gandalf := effectstest.HobPush(t, g, me.ID, "Gandalf, Wandering Wizard", "Legendary Creature — Avatar Wizard", 4, 5)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	if err := effectstest.HobB4Activate(t, g, me.ID, gandalf, 0, "{C}{C}{C}{C}{C}{C}", nil); err != nil {
		t.Fatalf("activation: %v", err)
	}
	effectstest.HobDestroy(t, g, gandalf)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d cards, want 3", got)
	}
	if !me.Graveyard.Contains(gandalf) {
		t.Error("Gandalf should be in the graveyard: there was nothing to shuffle in")
	}
}

func TestHobGandalfWardCountersAnOpponentsSpellUnlessPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	opp := hobB8Opponents(g, me)[0]
	gandalf := effectstest.PushCatalogPermanent(g, me.ID, "Gandalf, Wandering Wizard", "Legendary Creature — Avatar Wizard",
		effectstest.HobOracle(t, "Gandalf, Wandering Wizard"), false)
	effectstest.AdvanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := effectstest.CastAtWardedCreature(t, g, opp, gandalf)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment prompt for the spell's controller")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(gandalf) {
		t.Error("a declined ward counters the spell: Gandalf should live")
	}
	if !opp.Graveyard.Contains(blade) {
		t.Error("the countered spell should be in its owner's graveyard")
	}
}

// --- Bilbo, Unexpected Adventurer -------------------------------------

func hobB8Bilbo(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	return effectstest.HobPush(t, g, me.ID, "Bilbo, Unexpected Adventurer", "Legendary Creature — Halfling Rogue", 2, 2)
}

func TestHobBilboCantBeBlockedByPowerThreeOrGreater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	opp := hobB8Opponents(g, me)[0]
	bilbo := hobB8Bilbo(t, g, me)
	big := effectstest.HobB4Creature(g, opp.ID, "Big", "Creature — Ogre", 3, 3)
	small := effectstest.HobB4Creature(g, opp.ID, "Small", "Creature — Rat", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bilbo, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(big, bilbo); err == nil {
		t.Error("a power 3 creature blocked Bilbo")
	}
	if err := g.DeclareBlocker(small, bilbo); err != nil {
		t.Errorf("a power 2 creature may block Bilbo: %v", err)
	}
}

func TestHobBilboReturnsAGraveyardPermanentUnderItsOwnersControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	opp := hobB8Opponents(g, me)[0]
	bilbo := hobB8Bilbo(t, g, me)
	card := effectstest.PushGraveyardPermanent(opp, "Cheap Rat", "Creature — Rat", "{1}{B}")
	effectstest.AttackWith(t, g, opp.ID, bilbo)
	effectstest.PickTriggerTarget(t, g, me.ID, card)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, card)
	if !ok {
		t.Fatal("the card did not reach the battlefield")
	}
	if c.Controller != opp.ID {
		t.Errorf("it returned under %v's control, want its owner's (the opponent)", c.Controller)
	}
}

func TestHobBilboRejectsLandsAndExpensivePermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	opp := hobB8Opponents(g, me)[0]
	bilbo := hobB8Bilbo(t, g, me)
	land := effectstest.PushGraveyardPermanent(opp, "Forest", "Basic Land — Forest", "")
	pricey := effectstest.PushGraveyardPermanent(me, "Pricey Giant", "Creature — Giant", "{3}{R}")
	spell := effectstest.PushGraveyardPermanent(me, "Old Bolt", "Instant", "{R}")
	fine := effectstest.PushGraveyardPermanent(me, "Fine Artifact", "Artifact", "{3}")
	effectstest.AttackWith(t, g, opp.ID, bilbo)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no pick_target prompt")
	}
	for name, id := range map[string]uuid.UUID{"a land": land, "mana value 4": pricey, "an instant": spell} {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: id}); err == nil {
			t.Errorf("%s was accepted as the target", name)
			return
		}
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: fine}); err != nil {
		t.Fatalf("a mana value 3 artifact in my own graveyard is a legal target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fine) {
		t.Error("the artifact did not reach the battlefield")
	}
}

func TestHobBilboMayChooseNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	opp := hobB8Opponents(g, me)[0]
	bilbo := hobB8Bilbo(t, g, me)
	card := effectstest.PushGraveyardPermanent(opp, "Cheap Rat", "Creature — Rat", "{1}{B}")
	effectstest.AttackWith(t, g, opp.ID, bilbo)
	effectstest.B26DeclinePickTarget(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(card) || !opp.Graveyard.Contains(card) {
		t.Error("the card moved although no target was chosen")
	}
}

func TestHobBilboDoesNothingWhenBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	opp := hobB8Opponents(g, me)[0]
	bilbo := hobB8Bilbo(t, g, me)
	blocker := effectstest.HobB4Creature(g, opp.ID, "Small", "Creature — Rat", 1, 4)
	card := effectstest.PushGraveyardPermanent(opp, "Cheap Rat", "Creature — Rat", "{1}{B}")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bilbo, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, bilbo); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || g.Battlefield.Contains(card) {
		t.Error("combat damage to a creature triggered Bilbo")
	}
}

func TestHobBilboTriggersOnCombatDamageToABattle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	owner := hobB8Opponents(g, me)[0]
	bilbo := hobB8Bilbo(t, g, me)
	battle := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Siege", TypeLine: "Battle — Siege",
		Owner: owner.ID, Controller: owner.ID, ProtectorPlayerID: hobB8Opponents(g, me)[1].ID,
		Counters: map[string]int{game.CounterDefense: 9},
	})
	card := effectstest.PushGraveyardPermanent(me, "Cheap Rat", "Creature — Rat", "{1}{B}")
	effectstest.DeclareAttack(t, g, battle, bilbo)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PickTriggerTarget(t, g, me.ID, card)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(card) {
		t.Error("combat damage to a battle did not trigger Bilbo")
	}
}

// --- Lake-town Toymaker -----------------------------------------------

// hobB8DrawTo draws for `p` until this turn's tally holds `n` cards.
func hobB8DrawTo(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() {
		for g.TurnTallyFor(p.ID).CardsDrawn < n {
			_ = g.DrawNForEffect(p.ID, 1)
		}
	})
}

func hobB8Toymaker(t *testing.T, g *game.Game, me *game.Player) (toymaker, bear uuid.UUID) {
	t.Helper()
	toymaker = effectstest.HobPush(t, g, me.ID, "Lake-town Toymaker", "Creature — Human Artificer", 3, 4)
	bear = effectstest.SeedBear(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	return toymaker, bear
}

func TestHobToymakerPumpsAnotherCreatureAfterTwoDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	toymaker, bear := hobB8Toymaker(t, g, me)
	hobB8DrawTo(g, me, 2)
	basePower := effectstest.EffectivePower(t, g, bear)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.ResolvePickTarget(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: toymaker}); err == nil {
		t.Error("the Toymaker targeted itself: it says another creature")
	}
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != basePower+3 {
		t.Errorf("power %d, want %d", got, basePower+3)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the creature did not gain first strike")
	}
	if got := effectstest.EffectivePower(t, g, toymaker); got != 3 {
		t.Errorf("the Toymaker itself has power %d, want 3", got)
	}
}

func TestHobToymakerDoesNothingAfterOnlyOneDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	_, bear := hobB8Toymaker(t, g, me)
	// Drawing one more card than the tally now holds would leave it short
	// of two only if it started at zero: pin the tally at exactly one.
	g.WithWriteLock(func() {
		g.TurnTally.Players = map[uuid.UUID]game.PlayerTurnTally{me.ID: {CardsDrawn: 1}}
	})
	basePower := effectstest.EffectivePower(t, g, bear)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("the trigger went on the stack with one card drawn")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != basePower {
		t.Errorf("power %d, want the unchanged %d", got, basePower)
	}
}

func TestHobToymakerAloneHasNothingToTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.HobPush(t, g, me.ID, "Lake-town Toymaker", "Creature — Human Artificer", 3, 4)
	effectstest.AdvanceToMain(t, g)
	hobB8DrawTo(g, me, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a target prompt opened with no other creature to choose")
	}
}

// --- Gleaming Splendor ------------------------------------------------

func TestHobSplendorMakesATreasureWhenAnOpponentDrawsTheirSecondCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	opp := hobB8Opponents(g, me)[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Gleaming Splendor", "Enchantment", effectstest.HobOracle(t, "Gleaming Splendor"), false)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.BattlefieldCountNamed(g, me.ID, "Treasure"); n != 0 {
		t.Fatalf("%d Treasures after the first draw, want none", n)
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.BattlefieldCountNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasures after the second and third draws, want exactly 1 (only the second counts)", n)
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 3) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.BattlefieldCountNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasures after my own draws, want still 1", n)
	}
}

func TestHobSplendorTwoTargetPlayersEachDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	opps := hobB8Opponents(g, me)
	splendor := effectstest.PushCatalogPermanent(g, me.ID, "Gleaming Splendor", "Enchantment", effectstest.HobOracle(t, "Gleaming Splendor"), false)
	effectstest.AdvanceToMain(t, g)
	players := func(ids ...uuid.UUID) []game.TargetRef {
		var refs []game.TargetRef
		for _, id := range ids {
			refs = append(refs, game.TargetRef{Kind: game.TargetPlayer, ID: id})
		}
		return refs
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, splendor, 0, "{C}{C}", players(opps[0].ID, opps[1].ID)); err == nil {
		t.Fatal("activated {2}{W} without white mana")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, splendor, 0, "{W}", players(opps[0].ID)); err == nil {
		t.Error("one target player was accepted: it says two")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, splendor, 0, "", players(opps[0].ID, opps[0].ID)); err == nil {
		t.Error("the same player twice was accepted")
	}
	first, second, mine := opps[0].Hand.Size(), me.Hand.Size(), opps[2].Hand.Size()
	if err := effectstest.HobB4Activate(t, g, me.ID, splendor, 0, "", players(me.ID, opps[0].ID)); err != nil {
		t.Fatalf("activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := opps[0].Hand.Size() - first; got != 1 {
		t.Errorf("the targeted opponent drew %d, want 1", got)
	}
	if got := me.Hand.Size() - second; got != 1 {
		t.Errorf("I drew %d, want 1: I may target myself", got)
	}
	if got := opps[2].Hand.Size() - mine; got != 0 {
		t.Errorf("an untargeted player drew %d", got)
	}
}

// --- Elrond, Moon-Reader ----------------------------------------------

func hobB8Elrond(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	return effectstest.HobPush(t, g, me.ID, "Elrond, Moon-Reader", "Legendary Creature — Elf Noble", 3, 3)
}

func TestHobElrondDrawsOnceEachTurnWhenYouActivateACreaturesAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	hobB8Elrond(t, g, me)
	provisioner := effectstest.HobPush(t, g, me.ID, "Dwarven Provisioner", "Creature — Dwarf Citizen", 2, 2)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	for i := 0; i < 2; i++ {
		if err := effectstest.HobB4Activate(t, g, me.ID, provisioner, 0, "{W}{C}{C}{C}", nil); err != nil {
			t.Fatalf("activation %d: %v", i+1, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards for two activations in one turn, want exactly 1", got)
	}
}

func TestHobElrondIgnoresAbilitiesOfNoncreaturesAndOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	opp := hobB8Opponents(g, me)[0]
	hobB8Elrond(t, g, me)
	bear := effectstest.SeedBear(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	blade := effectstest.SeedEquipment(g, me.ID, "Dúnedain Blade", effectstest.HobOracle(t, "Dúnedain Blade"))
	harper := effectstest.HobPush(t, g, opp.ID, "Elvenking's Harper", "Creature — Elf Bard", 2, 2)
	hand := me.Hand.Size()
	if err := effectstest.HobB4Activate(t, g, me.ID, blade, 1, "{C}{C}{C}", effectstest.CardRefs(bear)); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := effectstest.HobB4Activate(t, g, opp.ID, harper, 0, "{U}{C}{C}{C}{C}", effectstest.CardRefs(harper)); err != nil {
		t.Fatalf("the opponent's activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d cards for an Equipment's ability and an opponent's, want none", got)
	}
}

func TestHobElrondCountsAManaAbilityOfACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	hobB8Elrond(t, g, me)
	survivor := effectstest.HobPush(t, g, me.ID, "Pelargir Survivor", "Creature — Human Peasant", 1, 3)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	if err := g.ActivateManaAbility(me.ID, survivor, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards for a creature's mana ability, want 1", got)
	}
}

func TestHobElrondCountsACreatureSacrificedToPayForItsOwnAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	opp := hobB8Opponents(g, me)[0]
	hobB8Elrond(t, g, me)
	library := effectstest.PushCatalogPermanent(g, me.ID, "Living Library", "Artifact Creature — Book Illusion", effectstest.RfpbLivingLibraryOracle, false)
	victim := effectstest.HobB4Creature(g, opp.ID, "Victim", "Creature — Rat", 1, 1)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	if err := effectstest.HobB4Activate(t, g, me.ID, library, 0, "{C}{C}{C}{C}{C}{C}", effectstest.CardRefs(victim)); err != nil {
		t.Fatalf("activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(library) {
		t.Fatal("the Library was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1: the sacrificed permanent was a creature as the ability was activated", got)
	}
}

func TestHobElrondFlickersUpToTwoOtherNonlandPermanentsAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	opp := hobB8Opponents(g, me)[0]
	elrond := hobB8Elrond(t, g, me)
	bear := effectstest.SeedBear(g, me.ID)
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	third := effectstest.SeedBear(g, me.ID)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.HobB4Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	for name, targets := range map[string][]uuid.UUID{
		"Elrond himself":          {elrond},
		"a land":                  {land},
		"an opponent's permanent": {theirs},
		"three permanents":        {bear, relic, third},
	} {
		if err := effectstest.HobB4Activate(t, g, me.ID, elrond, 0, "", effectstest.CardRefs(targets...)); err == nil {
			t.Errorf("%s was accepted as a target", name)
		}
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, elrond, 0, "{U}{U}{C}{C}{C}{C}{C}", effectstest.CardRefs(bear, relic)); err != nil {
		t.Fatalf("activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(relic) {
		t.Fatal("the targets are still on the battlefield before the end step")
	}
	if !g.Battlefield.Contains(third) {
		t.Error("an untargeted permanent left")
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.BattlefieldCountNamed(g, me.ID, "Relic"); n != 1 {
		t.Errorf("%d Relics returned, want 1", n)
	}
	if n := effectstest.BattlefieldCountNamed(g, me.ID, "Bear"); n != 2 {
		t.Errorf("%d Bears on the battlefield, want 2 (one stayed, one returned)", n)
	}
}

func TestHobElrondMayFlickerNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	elrond := hobB8Elrond(t, g, me)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.HobB4Activate(t, g, me.ID, elrond, 0, "{U}{U}{C}{C}{C}{C}{C}", nil); err != nil {
		t.Fatalf("activation with no targets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(elrond) {
		t.Error("Elrond left the battlefield")
	}
}

// --- Wizard's Staff ---------------------------------------------------

func hobB8Staff(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	return effectstest.SeedEquipment(g, me.ID, "Wizard's Staff", effectstest.HobOracle(t, "Wizard's Staff"))
}

func TestHobStaffEquipsAWizardForOneAndAnyCreatureForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.AdvanceToMain(t, g)
	staff := hobB8Staff(t, g, me)
	wizard := effectstest.HobB4Creature(g, me.ID, "Wizard Guy", "Creature — Human Wizard", 1, 1)
	bear := effectstest.SeedBear(g, me.ID)
	// A refused activation leaves its mana floating: one {C} stays in the
	// pool through the next two attempts.
	if err := effectstest.HobB4Activate(t, g, me.ID, staff, 0, "{C}", effectstest.CardRefs(bear)); err == nil {
		t.Fatal("Equip Wizard {1} accepted a non-Wizard")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, staff, 1, "", effectstest.CardRefs(bear)); err == nil {
		t.Fatal("Equip {3} was activated for one mana")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, staff, 0, "", effectstest.CardRefs(wizard)); err != nil {
		t.Fatalf("Equip Wizard {1}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if host := effectstest.AttachmentHostOf(t, g, staff); host.ID != wizard {
		t.Fatalf("the Staff is on %v, want the Wizard", host)
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, staff, 1, "{C}{C}{C}", effectstest.CardRefs(bear)); err != nil {
		t.Fatalf("Equip {3}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if host := effectstest.AttachmentHostOf(t, g, staff); host.ID != bear {
		t.Fatalf("the Staff is on %v, want the Bear", host)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "prowess") {
		t.Error("the equipped creature did not gain prowess")
	}
	if effectstest.HasEffectiveKeyword(t, g, wizard, "prowess") {
		t.Error("the creature the Staff left still has prowess")
	}
}

func TestHobStaffDoublesTheEquippedCreaturesTriggers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	opp := hobB8Opponents(g, me)[0]
	dain := effectstest.HobPush(t, g, me.ID, "Dáin of the Ancient Halls", "Legendary Creature — Dwarf Noble", 4, 5)
	staff := hobB8Staff(t, g, me)
	effectstest.AdvanceToMain(t, g)
	effectstest.EquipTo(t, g, me.ID, staff, dain)
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, dain)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d -> %d, want 2 damage: Dáin's attack trigger fires twice", life, opp.Life)
	}
}

func TestHobStaffDoesNotDoubleAnUnequippedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	opp := hobB8Opponents(g, me)[0]
	dain := effectstest.HobPush(t, g, me.ID, "Dáin of the Ancient Halls", "Legendary Creature — Dwarf Noble", 4, 5)
	bear := effectstest.SeedBear(g, me.ID)
	staff := hobB8Staff(t, g, me)
	effectstest.AdvanceToMain(t, g)
	effectstest.EquipTo(t, g, me.ID, staff, bear)
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, dain)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d -> %d, want 1 damage: the Staff is on the Bear, so Dáin's trigger fires once", life, opp.Life)
	}
}

func TestHobStaffGivesProwessAndProwessTriggersTwice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	opp := hobB8Opponents(g, me)[0]
	bear := effectstest.SeedBear(g, me.ID)
	staff := hobB8Staff(t, g, me)
	effectstest.AdvanceToMain(t, g)
	effectstest.EquipTo(t, g, me.ID, staff, bear)
	base := effectstest.EffectivePower(t, g, bear)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != base+2 {
		t.Errorf("power %d -> %d, want +2: prowess triggered twice for one noncreature spell", base, got)
	}
}

func TestHobStaffDoublesADiesTriggerOfTheEquippedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	kingfisher := effectstest.HobPush(t, g, me.ID, "Ithilien Kingfisher", "Creature — Bird", 2, 1)
	staff := hobB8Staff(t, g, me)
	effectstest.AdvanceToMain(t, g)
	effectstest.EquipTo(t, g, me.ID, staff, kingfisher)
	hand := me.Hand.Size()
	effectstest.HobDestroy(t, g, kingfisher)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2: the dies trigger of the creature that wore the Staff fires twice", got)
	}
}

// --- Arwen, Mortal Queen ----------------------------------------------

func TestHobArwenEntersIndestructibleAndSpendsTheCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	bear := effectstest.SeedBear(g, me.ID)
	arwen := hobCast(t, g, "Arwen, Mortal Queen", "Legendary Creature — Elf Noble", 2, 2)
	if c, _ := effectstest.BattlefieldCard(g, arwen); c.Counters[game.CounterIndestructible] != 1 {
		t.Fatalf("Arwen entered with %v, want an indestructible counter", c.Counters)
	}
	effectstest.HobDestroy(t, g, arwen)
	if !g.Battlefield.Contains(arwen) {
		t.Fatal("an indestructible Arwen was destroyed")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, arwen, 0, "{C}", effectstest.CardRefs(arwen)); err == nil {
		t.Fatal("Arwen targeted herself: it says another creature")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, arwen, 0, "", effectstest.CardRefs(bear)); err != nil {
		t.Fatalf("activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"the target": bear, "Arwen": arwen} {
		c, ok := effectstest.BattlefieldCard(g, id)
		if !ok {
			t.Fatalf("%s left the battlefield", name)
		}
		if c.Counters[game.CounterPlusOne] != 1 || c.Counters[game.CounterLifelink] != 1 {
			t.Errorf("%s has counters %v, want a +1/+1 and a lifelink counter", name, c.Counters)
		}
	}
	if c, _ := effectstest.BattlefieldCard(g, arwen); c.Counters[game.CounterIndestructible] != 0 {
		t.Errorf("Arwen still has %d indestructible counters", c.Counters[game.CounterIndestructible])
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("the target did not gain indestructible until end of turn")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, arwen, 0, "{C}", effectstest.CardRefs(bear)); err == nil {
		t.Error("activated again with no indestructible counter to remove")
	}
	effectstest.HobDestroy(t, g, arwen)
	if g.Battlefield.Contains(arwen) {
		t.Error("Arwen, with her counter spent, survived destruction")
	}
}

func TestHobArwenFizzlesWhenTheTargetIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	bear := effectstest.SeedBear(g, me.ID)
	arwen := hobCast(t, g, "Arwen, Mortal Queen", "Legendary Creature — Elf Noble", 2, 2)
	if err := effectstest.HobB4Activate(t, g, me.ID, arwen, 0, "{C}", effectstest.CardRefs(bear)); err != nil {
		t.Fatalf("activation: %v", err)
	}
	effectstest.HobDestroy(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, arwen); c.Counters[game.CounterPlusOne] != 0 {
		t.Errorf("Arwen got %v although her target was gone: the ability does nothing", c.Counters)
	}
}

// --- The Arkenstone // Seek the Heart ---------------------------------

// hobB8Adventure is The Arkenstone // Seek the Heart in `me`'s hand.
func hobB8Adventure(t *testing.T, me *game.Player) game.Card {
	t.Helper()
	card := game.Card{
		InstanceID: uuid.New(), OracleID: effectstest.HobOracle(t, "The Arkenstone"), Layout: game.LayoutAdventure,
		Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "The Arkenstone", TypeLine: "Legendary Artifact", ManaCost: "{5}"},
			{Name: "Seek the Heart", TypeLine: "Sorcery — Adventure", ManaCost: "{2}{W}", Colors: []string{"W"}},
		},
	}
	card.SetFace(0)
	me.Hand.PushTop(card)
	return card
}

func TestHobArkenstoneGivesCreaturesYouControlPlusOneAndDrawsAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	opp := hobB8Opponents(g, me)[0]
	bear := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.SeedBear(g, opp.ID)
	before, theirBefore := effectstest.EffectivePower(t, g, bear), effectstest.EffectivePower(t, g, theirs)
	effectstest.PushCatalogPermanent(g, me.ID, "The Arkenstone", "Legendary Artifact", effectstest.HobOracle(t, "The Arkenstone"), false)
	if got := effectstest.EffectivePower(t, g, bear); got != before+1 {
		t.Errorf("my Bear has power %d, want %d", got, before+1)
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != theirBefore {
		t.Errorf("an opponent's Bear has power %d, want the unchanged %d", got, theirBefore)
	}
	hand := me.Hand.Size()
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards at my end step, want 1", got)
	}
}

func TestHobSeekTheHeartFetchesALegendaryCreatureThenTheArtifactIsCastFromExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	card := hobB8Adventure(t, me)
	effectstest.StapleLibraryCard(me, "Plain Elf", "Creature — Elf")
	legend := effectstest.StapleLibraryCard(me, "Legendary Elf", "Legendary Creature — Elf Noble")
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Seek the Heart: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, legend)
	if !me.Hand.Contains(legend) {
		t.Fatal("the legendary creature did not reach the hand")
	}
	if !g.Exile.Contains(card.InstanceID) {
		t.Fatal("the resolved Adventure is not in exile")
	}
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast The Arkenstone from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(card.InstanceID) {
		t.Error("The Arkenstone did not enter the battlefield")
	}
}

func TestHobSeekTheHeartFindsOnlyLegendaryCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	card := hobB8Adventure(t, me)
	artifact := effectstest.StapleLibraryCard(me, "Legendary Relic", "Legendary Artifact")
	plain := effectstest.StapleLibraryCard(me, "Plain Elf", "Creature — Elf")
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Seek the Heart: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(artifact) || me.Hand.Contains(plain) {
		t.Error("a card that is not a legendary creature reached the hand")
	}
}
