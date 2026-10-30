package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msh_a8_cards_test.go — Marvel Super Heroes, half A, eighth batch (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mshA8SecretInvasionOracle = "25443ea4-3177-4e29-a2b1-f3fee60b3205"
	mshA8SpeedballOracle      = "53d3ef8f-8e0f-4119-8cdc-a0bf076b4e91"
	mshA8AdaptoidOracle       = "f93fa6ec-1f13-4109-aa60-11e148f7ba88"
	mshA8SerumOracle          = "6f2ba142-5e7c-4709-ae03-e028cb2a9872"
	mshA8SwordsmanOracle      = "2f15abe7-54aa-40ed-a15d-5b827f1c3923"
	mshA8TaskmasterOracle     = "dcd97730-b793-4c0c-b158-625ac3656b4a"
	mshA8AntManOracle         = "6c7e4e51-099a-4398-a4b7-e2f0ddd36429"
	mshA8KingpinOracle        = "3eeeb159-aa45-4a9b-9618-8542604b9830"
	mshA8SentryOracle         = "1fe549c3-79be-40ef-b57c-de59f24df5ab"
	mshA8SquirrelGirlOracle   = "80312910-5d53-44f1-9982-e46dc7532abb"
	mshA8ThorOracle           = "ec9b74d0-e4fe-4764-b628-992a62f2a1cf"
	mshA8CompoundOracle       = "99c70f4e-de8a-426d-99aa-17b2f87625ba"
	mshA8StratagemOracle      = "170ee399-d3cd-40ef-a340-193c0f0f6002"
)

func TestMshA8EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		mshA8SecretInvasionOracle: "Secret Invasion",
		mshA8SpeedballOracle:      "Speedball, New Warrior",
		mshA8AdaptoidOracle:       "Super-Adaptoid",
		mshA8SerumOracle:          "Super-Soldier Serum",
		mshA8SwordsmanOracle:      "Swordsman, Sharp Scoundrel",
		mshA8TaskmasterOracle:     "Taskmaster, Mercenary Mimic",
		mshA8AntManOracle:         "The Astonishing Ant-Man",
		mshA8KingpinOracle:        "The Kingpin of Crime",
		mshA8SentryOracle:         "The Sentry, Golden Guardian",
		mshA8SquirrelGirlOracle:   "The Unbeatable Squirrel Girl",
		mshA8ThorOracle:           "Thor Odinson",
		mshA8CompoundOracle:       "Training Compound",
		mshA8StratagemOracle:      "Trickster's Stratagem",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// mshA8Layered reads a permanent with the layer pass applied.
func mshA8Layered(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	return gaLayered(t, g, id)
}

// mshA8Count counts a player's permanents of a name.
func mshA8Count(g *game.Game, controller uuid.UUID, name string) int {
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

// --- Thor Odinson ----------------------------------------------------------------------------

func TestMshA8ThorHasFlyingVigilanceAndTwoProwesses(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	thor := pushCatalogPermanent(g, me.ID, "Thor Odinson", "Legendary Creature — God Warrior Hero", mshA8ThorOracle, false)
	for _, kw := range []string{"flying", "vigilance"} {
		if !mshA3HasKeyword(g, thor, kw) {
			t.Errorf("Thor has no %s", kw)
		}
	}
	if n := prowessCountOf(t, g, thor); n != 2 {
		t.Fatalf("Thor has %d instances of prowess, want 2", n)
	}
	base := effectivePower(t, g, thor)
	castNoncreature(t, g)
	if got := prowessItemsFrom(g, thor); got != 2 {
		t.Fatalf("one noncreature spell put %d prowess triggers on the stack, want 2", got)
	}
	settleProwess(t, g)
	if got := effectivePower(t, g, thor); got != base+2 {
		t.Errorf("Thor's power %d after one noncreature spell, want %d", got, base+2)
	}
}

// --- Training Compound -----------------------------------------------------------------------

func TestMshA8TrainingCompoundTapsForRedOrGreenWithABasicLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := pushCatalogPermanent(g, me.ID, "Training Compound", "Land", mshA8CompoundOracle, false)
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err == nil {
		t.Fatal("the coloured ability worked with no basic land and without having entered this turn")
	}
	pushCatalogPermanent(g, me.ID, "Mountain", "Basic Land — Mountain", "", false)
	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	b10ResolveAllManaPicks(t, g, me.ID, "G")
	if got := poolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("tapped for {G} with a basic land: pool %v", got)
	}
}

func TestMshA8TrainingCompoundTapsForColourOnTheTurnItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := playLandFromHand(t, g, "Training Compound", mshA8CompoundOracle)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	mshA3TapSecondRow(t, g, me.ID, land, "R")
	if got := poolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("tapped for {R} the turn it entered: pool %v", got)
	}
}

func TestMshA8TrainingCompoundAlwaysTapsForColourless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := pushCatalogPermanent(g, me.ID, "Training Compound", "Land", mshA8CompoundOracle, false)
	advanceToMain(t, g)
	b28TapForMana(t, g, me.ID, land, "")
	if got := poolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("colourless ability: pool %v", got)
	}
}

// --- The Unbeatable Squirrel Girl ------------------------------------------------------------

func TestMshA8SquirrelGirlMakesASquirrelWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b12PlayFromHand(t, g, "The Unbeatable Squirrel Girl", "Legendary Creature — Squirrel Human Hero", mshA8SquirrelGirlOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := mshA8Count(g, me.ID, "Squirrel"); got != 1 {
		t.Fatalf("%d Squirrel tokens after she entered, want 1", got)
	}
}

func TestMshA8SquirrelGirlMakesASquirrelWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	girl := pushCatalogPermanent(g, me.ID, "The Unbeatable Squirrel Girl", "Legendary Creature — Squirrel Human Hero", mshA8SquirrelGirlOracle, false)
	attackWith(t, g, opp.ID, girl)
	passPriorityAroundTable(t, g)
	if got := mshA8Count(g, me.ID, "Squirrel"); got != 1 {
		t.Fatalf("%d Squirrel tokens after she attacked, want 1", got)
	}
}

func TestMshA8SquirrelGirlActivationMakesAsManySquirrelsAsYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	girl := pushCatalogPermanent(g, me.ID, "The Unbeatable Squirrel Girl", "Legendary Creature — Squirrel Human Hero", mshA8SquirrelGirlOracle, false)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 green Squirrel"), 2) })
	advanceToMain(t, g)
	// A non-Squirrel does not count.
	pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	floatMana(t, g, me, "{C}{G}{G}{G}")
	b16Activate(t, g, me.ID, girl, 0, game.ActivateAbilityParams{})
	// She and two tokens are three Squirrels, so three more arrive.
	if got := mshA8Count(g, me.ID, "Squirrel"); got != 5 {
		t.Errorf("%d Squirrel tokens, want 5 (two plus X = 3)", got)
	}
}

// --- The Astonishing Ant-Man -----------------------------------------------------------------

func TestMshA8AntManGrowsWhenYouDrawAndSpendsCountersOnInsects(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ant := pushCatalogPermanent(g, me.ID, "The Astonishing Ant-Man", "Legendary Creature — Human Scientist Hero", mshA8AntManOracle, false)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 3) })
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, ant); got != 3 {
		t.Fatalf("%d +1/+1 counters after three draws, want 3", got)
	}
	floatMana(t, g, me, "{C}{C}{G}")
	b16Activate(t, g, me.ID, ant, 0, game.ActivateAbilityParams{CounterCounts: []int{2}})
	if got := mshA8Count(g, me.ID, "Insect"); got != 2 {
		t.Errorf("%d Insect tokens, want 2", got)
	}
	if got := plusOnes(g, ant); got != 1 {
		t.Errorf("%d +1/+1 counters left, want 1", got)
	}
	if !b16Tapped(t, g, ant) {
		t.Error("the ability's {T} cost was not paid")
	}
}

func TestMshA8AntManIgnoresAnOpponentsDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	ant := pushCatalogPermanent(g, me.ID, "The Astonishing Ant-Man", "Legendary Creature — Human Scientist Hero", mshA8AntManOracle, false)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 2) })
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, ant); got != 0 {
		t.Errorf("an opponent's draws put %d counters on Ant-Man", got)
	}
}

// --- The Sentry, Golden Guardian -------------------------------------------------------------

func TestMshA8SentryGivesTheTargetOpponentTheVoid(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	advanceToMain(t, g)
	sentry := b12PlayFromHand(t, g, "The Sentry, Golden Guardian", "Legendary Creature — Human Hero", mshA8SentryOracle, game.CastSpellParams{})
	pick := game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
	assistCardSettle(t, g, me.ID, &pick, true)
	for _, kw := range []string{"flying", "vigilance", "indestructible"} {
		if !mshA3HasKeyword(g, sentry, kw) {
			t.Errorf("The Sentry has no %s", kw)
		}
	}
	if got := mshA8Count(g, me.ID, "The Void"); got != 0 {
		t.Errorf("the caster got %d Voids", got)
	}
	if got := mshA8Count(g, opp.ID, "The Void"); got != 1 {
		t.Fatalf("the target opponent controls %d Voids, want 1", got)
	}
	var voidID uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "The Void" {
				voidID = c.InstanceID
			}
		}
	})
	void := mshA8Layered(t, g, voidID)
	if !void.IsToken() || !void.IsLegendary() || void.CurrentPower() != 5 || void.CurrentToughness() != 5 ||
		!void.HasSubtype("Horror") || !void.HasSubtype("Villain") || len(void.Colors) != 1 || void.Colors[0] != "B" {
		t.Errorf("The Void is %+v, want a legendary 5/5 black Horror Villain token", void)
	}
	for _, kw := range []string{"flying", "indestructible"} {
		if !mshA3HasKeyword(g, voidID, kw) {
			t.Errorf("The Void has no %s", kw)
		}
	}
	if len(void.Effective().AttackRequirements) == 0 {
		t.Error("The Void has no attack requirement (it attacks each combat if able)")
	}
}

// --- Trickster's Stratagem -------------------------------------------------------------------

// mshA8Stratagem casts Trickster's Stratagem at an opponent's creature and
// (optionally) one of mine, and returns the opponent, their creature and mine.
func mshA8Stratagem(t *testing.T, g *game.Game, withMine bool) (opp *game.Player, theirs, mine uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs = b12Creature(g, opp.ID, "Their Creature", "Creature — Bear", 2, 2)
	mine = b12Creature(g, me.ID, "My Creature", "Creature — Bear", 2, 2)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}
	if withMine {
		targets = append(targets, game.TargetRef{Kind: game.TargetCard, ID: mine})
	}
	castCatalogSpell(t, g, "Trickster's Stratagem", "Sorcery", mshA8StratagemOracle, targets)
	passPriorityAroundTable(t, g)
	return opp, theirs, mine
}

func TestMshA8StratagemOwnerTucksThenYourCreatureConnives(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	lib := seedLibrary(opp, "Top", "Next")
	drawn := seedLibraryTop(me, "Drawn Spell", "Instant", "")
	handBefore := me.Hand.Size()
	_, theirs, mine := mshA8Stratagem(t, g, true)

	ask := putInLibraryChoiceFor(g, opp.ID)
	if ask == nil || ask.LibraryPlacement != game.LibraryPlaceTopOrBottom || ask.LibraryTopDepth != 2 {
		t.Fatalf("the OWNER is asked, with the top lane at depth 2: %+v", ask)
	}
	if putInLibraryChoiceFor(g, me.ID) != nil {
		t.Fatal("the caster does not choose")
	}
	// The connive waits for the owner's answer.
	if me.Hand.Size() != handBefore || discardChoiceFor(g, me.ID) != nil {
		t.Fatalf("the connive began before the owner chose: hand %d → %d", handBefore, me.Hand.Size())
	}
	if err := g.ResolvePutInLibrary(ask.ID, opp.ID, nil, []uuid.UUID{theirs}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if onBattlefield(g, theirs) {
		t.Fatal("their creature stayed on the battlefield")
	}
	if got := libraryTopIDs(opp, 3); !sameIDs(got, []uuid.UUID{lib[0], theirs, lib[1]}) {
		t.Errorf("top three %v, want [Top Their-Creature Next]", got)
	}
	// Then my creature connives: draw, discard the nonland, counter.
	if me.Hand.Size() != handBefore+1 {
		t.Fatalf("the connive drew no card: hand %d → %d", handBefore, me.Hand.Size())
	}
	answerDiscard(t, g, me.ID, drawn)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, mine); got != 1 {
		t.Errorf("%d +1/+1 counters on my creature, want 1 (a nonland was discarded)", got)
	}
}

func TestMshA8StratagemOwnerMayPutItOnTheBottom(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seedLibrary(opp, "Top", "Next")
	_, theirs, _ := mshA8Stratagem(t, g, false)
	ask := putInLibraryChoiceFor(g, opp.ID)
	if ask == nil {
		t.Fatal("the owner was not asked")
	}
	if err := g.ResolvePutInLibrary(ask.ID, opp.ID, []uuid.UUID{theirs}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if opp.Library.Cards[0].InstanceID != theirs {
		t.Error("it is not on the bottom of its owner's library")
	}
}

func TestMshA8StratagemWithNoCreatureOfYoursNothingConnives(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seedLibrary(opp, "Top", "Next")
	handBefore := me.Hand.Size()
	_, theirs, _ := mshA8Stratagem(t, g, false)
	ask := putInLibraryChoiceFor(g, opp.ID)
	if ask == nil {
		t.Fatal("the owner was not asked")
	}
	if err := g.ResolvePutInLibrary(ask.ID, opp.ID, []uuid.UUID{theirs}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if me.Hand.Size() != handBefore || discardChoiceFor(g, me.ID) != nil {
		t.Errorf("a connive happened with no second target: hand %d → %d", handBefore, me.Hand.Size())
	}
}

// --- Swordsman, Sharp Scoundrel --------------------------------------------------------------

func mshA8PickTargets(t *testing.T, g *game.Game, chooser uuid.UUID, ids ...uuid.UUID) {
	t.Helper()
	b04WaitForPick(t, g, chooser)
	p := latestPickTarget(g, chooser)
	refs := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	if err := g.ResolvePickTargets(p.ID, chooser, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// mshA8PickOne answers one clause's pick_target prompt, without passing priority.
func mshA8PickOne(t *testing.T, g *game.Game, chooser uuid.UUID, ids ...uuid.UUID) {
	t.Helper()
	b04WaitForPick(t, g, chooser)
	p := latestPickTarget(g, chooser)
	refs := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	if err := g.ResolvePickTargets(p.ID, chooser, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
}

// mshA8ConfirmAsked passes priority until a yes/no prompt is open for the player.
func mshA8ConfirmAsked(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	for i := 0; i < 12; i++ {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceConfirm && c.Chooser == me.ID {
				return
			}
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatalf("no yes/no prompt for %s: %+v", me.Name, g.PendingChoices)
}

func mshA8AttachedTo(t *testing.T, g *game.Game, attachment uuid.UUID) uuid.UUID {
	t.Helper()
	return b12Card(t, g, attachment).AttachedTo.ID
}

func TestMshA8SwordsmanAttachesAnEquipmentWhenAnotherVillainEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Swordsman, Sharp Scoundrel", "Legendary Creature — Human Hero Villain", mshA8SwordsmanOracle, false)
	blade := b12Permanent(g, me.ID, "Test Blade", "Artifact — Equipment")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Henchman", "Creature — Human Villain", "", nil)
	passPriorityAroundTable(t, g)
	mshA8PickOne(t, g, me.ID, blade)
	mshA8PickOne(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := mshA8AttachedTo(t, g, blade); got != bear {
		t.Errorf("the Equipment is attached to %v, want the Bear", got)
	}
}

func TestMshA8SwordsmanIgnoresANonVillainAndTheOpponentsVillain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Swordsman, Sharp Scoundrel", "Legendary Creature — Human Hero Villain", mshA8SwordsmanOracle, false)
	b12Permanent(g, me.ID, "Test Blade", "Artifact — Equipment")
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("a non-Villain entering asked for targets")
	}
	castFlashCreatureFrom(t, g, opp, game.Card{Name: "Their Henchman", TypeLine: "Creature — Human Villain", Keywords: []string{"flash"}})
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("an opponent's Villain entering asked for targets")
	}
}

func TestMshA8SwordsmanEquippedCreatureConnivesWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Swordsman, Sharp Scoundrel", "Legendary Creature — Human Hero Villain", mshA8SwordsmanOracle, false)
	blade := b12Permanent(g, me.ID, "Test Blade", "Artifact — Equipment")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AttachForEffect(blade, game.TargetRef{Kind: game.TargetCard, ID: bear})
	})
	drawn := seedLibraryTop(me, "Drawn Spell", "Instant", "")
	handBefore := me.Hand.Size()
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore+1 {
		t.Fatalf("the equipped attacker did not draw: hand %d → %d", handBefore, me.Hand.Size())
	}
	answerDiscard(t, g, me.ID, drawn)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, bear); got != 1 {
		t.Errorf("%d +1/+1 counters on the equipped creature, want 1", got)
	}
}

func TestMshA8SwordsmanUnequippedAttackerDoesNotConnive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Swordsman, Sharp Scoundrel", "Legendary Creature — Human Hero Villain", mshA8SwordsmanOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	handBefore := me.Hand.Size()
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore {
		t.Errorf("an unequipped attacker connived: hand %d → %d", handBefore, me.Hand.Size())
	}
}

// --- The Kingpin of Crime --------------------------------------------------------------------

func TestMshA8KingpinHasExtort(t *testing.T) {
	spec, _ := Lookup(mshA8KingpinOracle)
	found := false
	for _, tr := range spec.Triggered {
		if tr.Keyword == KeywordExtort {
			found = true
		}
	}
	if !found {
		t.Error("The Kingpin of Crime has no extort trigger")
	}
}

func TestMshA8KingpinPayingLifeMakesToughnessDealTheDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "The Kingpin of Crime", "Legendary Creature — Human Villain", mshA8KingpinOracle, true)
	wall := b12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 4)
	brute := b12Creature(g, me.ID, "Brute", "Creature — Ogre", 3, 3)
	life := me.Life
	declareAttack(t, g, opp.ID, wall, brute)
	mshA8ConfirmAsked(t, g, me)
	answerConfirmFor(t, g, me, true)
	passPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Fatalf("life %d, want %d (paid 2)", me.Life, life-2)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	// The 0/4 Wall assigns 4 by toughness; the 3/3 still assigns 3.
	if want := 40 - 4 - 3; opp.Life != want {
		t.Errorf("opponent life %d, want %d", opp.Life, want)
	}
}

func TestMshA8KingpinDecliningPaysNothingAndChangesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "The Kingpin of Crime", "Legendary Creature — Human Villain", mshA8KingpinOracle, true)
	wall := b12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 4)
	life := me.Life
	declareAttack(t, g, opp.ID, wall)
	mshA8ConfirmAsked(t, g, me)
	answerConfirmFor(t, g, me, false)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d, want %d", me.Life, life)
	}
	if opp.Life != 40 {
		t.Errorf("opponent life %d, want 40 (the Wall deals 0 by power)", opp.Life)
	}
}

func TestMshA8KingpinEffectEndsAtEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "The Kingpin of Crime", "Legendary Creature — Human Villain", mshA8KingpinOracle, true)
	wall := b12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 4)
	declareAttack(t, g, opp.ID, wall)
	mshA8ConfirmAsked(t, g, me)
	answerConfirmFor(t, g, me, true)
	passPriorityAroundTable(t, g)
	if !gaLayered(t, g, wall).Effective().DamageByToughness {
		t.Fatal("the Wall does not assign damage by toughness this turn")
	}
	advanceToNextSeatsTurn(t, g)
	if gaLayered(t, g, wall).Effective().DamageByToughness {
		t.Error("the effect outlasted the turn")
	}
}

// --- Taskmaster, Mercenary Mimic -------------------------------------------------------------

func mshA8Taskmaster(g *game.Game) uuid.UUID {
	me := g.Seats[0]
	return b12Push(g, me.ID, "Taskmaster, Mercenary Mimic", "Legendary Creature — Human Mercenary Villain", mshA8TaskmasterOracle, 3, 5)
}

func TestMshA8TaskmasterCopiesACreatureOnTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	task := mshA8Taskmaster(g)
	dragon := b12Creature(g, opp.ID, "Elder Dragon", "Artifact Creature — Dragon", 5, 5)
	advanceToMain(t, g)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: dragon}, true)
	c := gaLayered(t, g, task)
	if c.Name != "Taskmaster, Mercenary Mimic" || !c.IsLegendary() || !c.IsCreature() {
		t.Errorf("he is %q %q, want the legendary creature Taskmaster, Mercenary Mimic", c.Name, c.TypeLine)
	}
	if c.IsArtifact() || c.HasSubtype("Dragon") || !c.HasSubtype("Human") || !c.HasSubtype("Mercenary") || !c.HasSubtype("Villain") {
		t.Errorf("type line %q, want exactly a legendary Human Mercenary Villain creature", c.TypeLine)
	}
	if c.CurrentPower() != 5 || c.CurrentToughness() != 5 {
		t.Errorf("he is %d/%d, want the Dragon's 5/5", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestMshA8TaskmasterCopiesACreatureCardInAGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	task := mshA8Taskmaster(g)
	card := uuid.New()
	opp.Graveyard.PushTop(game.Card{
		InstanceID: card, Name: "Fallen Giant", TypeLine: "Creature — Giant", Power: 7, Toughness: 6,
		Owner: opp.ID, Controller: opp.ID,
	})
	advanceToMain(t, g)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: card}, true)
	c := gaLayered(t, g, task)
	if c.Name != "Taskmaster, Mercenary Mimic" || c.CurrentPower() != 7 || c.CurrentToughness() != 6 {
		t.Errorf("he is %q %d/%d, want Taskmaster as a 7/6", c.Name, c.CurrentPower(), c.CurrentToughness())
	}
	if c.HasSubtype("Giant") {
		t.Errorf("type line %q kept the Giant", c.TypeLine)
	}
	if !opp.Graveyard.Contains(card) {
		t.Error("the copied card left the graveyard")
	}
}

func TestMshA8TaskmasterNoTargetChangesNothing(t *testing.T) {
	g := newCatalogGame(t)
	task := mshA8Taskmaster(g)
	b12Creature(g, g.Seats[1].ID, "Elder Dragon", "Creature — Dragon", 5, 5)
	advanceToMain(t, g)
	pick := pickTargetPrompt(t, g)
	if err := g.ResolvePickTargets(pick.ID, pick.Chooser, nil); err != nil {
		t.Fatalf("ResolvePickTargets(none): %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := gaLayered(t, g, task); c.CurrentPower() != 3 || c.CurrentToughness() != 5 {
		t.Errorf("he is %d/%d, want 3/5", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestMshA8TaskmasterRevertsWhenYourNextTurnBegins(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	task := mshA8Taskmaster(g)
	dragon := b12Creature(g, opp.ID, "Elder Dragon", "Creature — Dragon", 5, 5)
	advanceToMain(t, g)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: dragon}, true)
	advanceToNextSeatsTurn(t, g)
	for g.Turn.ActiveSeat != 0 {
		if c := gaLayered(t, g, task); c.CurrentPower() != 5 {
			t.Fatalf("the copy ended during seat %d's turn", g.Turn.ActiveSeat)
		}
		advanceToNextSeatsTurn(t, g)
	}
	for i := 0; i < 8 && g.Turn.Step != game.StepUpkeep; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if c := gaLayered(t, g, task); c.CurrentPower() != 3 || c.CurrentToughness() != 5 {
		t.Errorf("still a copy on my next turn: %d/%d", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Super-Soldier Serum ---------------------------------------------------------------------

func mshA8Serum(t *testing.T, g *game.Game) (bear uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	bear = b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Super-Soldier Serum", "Enchantment — Aura", mshA8SerumOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	return bear
}

func TestMshA8SerumPumpsAndGrantsKeywordsAndTypes(t *testing.T) {
	g := newCatalogGame(t)
	bear := mshA8Serum(t, g)
	c := gaLayered(t, g, bear)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	for _, kw := range []string{"first strike", "vigilance"} {
		if !mshA3HasKeyword(g, bear, kw) {
			t.Errorf("the Bear has no %s", kw)
		}
	}
	if !c.IsLegendary() || !c.HasSubtype("Soldier") || !c.HasSubtype("Bear") || !c.IsCreature() {
		t.Errorf("type line %q, want a legendary Soldier Bear creature", c.TypeLine)
	}
}

func TestMshA8SerumAttackAttachesAnyNumberOfYourEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	bear := mshA8Serum(t, g)
	one := b12Permanent(g, me.ID, "Blade One", "Artifact — Equipment")
	two := b12Permanent(g, me.ID, "Blade Two", "Artifact — Equipment")
	theirs := b12Permanent(g, opp.ID, "Their Blade", "Artifact — Equipment")
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	pick := pickTargetPrompt(t, g)
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.PickTargetCards {
		offered[id] = true
	}
	if !offered[one] || !offered[two] || offered[theirs] {
		t.Fatalf("offered %v, want both of mine and not theirs", pick.PickTargetCards)
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: one}, {Kind: game.TargetCard, ID: two},
	}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if mshA8AttachedTo(t, g, one) != bear || mshA8AttachedTo(t, g, two) != bear {
		t.Error("both Equipment should be attached to the enchanted creature")
	}
}

func TestMshA8SerumBlockAttachesEquipmentToo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := mshA8Serum(t, g)
	blade := b12Permanent(g, me.ID, "Blade One", "Artifact — Equipment")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBlock, CardID: bear, Actor: me.ID})
	})
	passPriorityAroundTable(t, g)
	mshA8PickTargets(t, g, me.ID, blade)
	if got := mshA8AttachedTo(t, g, blade); got != bear {
		t.Errorf("the Equipment is attached to %v, want the enchanted creature", got)
	}
}

func TestMshA8SerumAttackWithNoEquipmentAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := mshA8Serum(t, g)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			if err := g.ResolvePickTargets(c.ID, c.Chooser, nil); err != nil {
				t.Fatalf("ResolvePickTargets(none): %v", err)
			}
		}
	}
	passPriorityAroundTable(t, g)
}

// --- Speedball, New Warrior ------------------------------------------------------------------

func TestMshA8SpeedballGetsPlusTwoAndMayRedirectTheSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ball := pushCatalogPermanent(g, me.ID, "Speedball, New Warrior", "Legendary Creature — Human Hero", mshA8SpeedballOracle, false)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	base := effectivePower(t, g, ball)
	spell := castCatalogSpell(t, g, "Djeru's Resolve", "Instant", pr7aDjerusResolve, []game.TargetRef{pr7aCard(ball)})
	passPriorityAroundTable(t, g)
	prompt := latestRetarget(g, me.ID)
	if prompt == nil {
		t.Fatal("Speedball offered no retarget")
	}
	if prompt.PickTargetMin != 0 {
		t.Errorf(`"you may choose new targets" offers min %d, want 0`, prompt.PickTargetMin)
	}
	if err := g.ResolveRetarget(prompt.ID, me.ID, []game.TargetRef{pr7aCard(other)}); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
	if got := g.StackMeta[spell].Targets; len(got) != 1 || got[0].ID != other {
		t.Fatalf("the spell's target did not move: %+v", got)
	}
	if got := effectivePower(t, g, ball); got != base+2 {
		t.Errorf("Speedball's power %d, want %d (+2/+2 until end of turn)", got, base+2)
	}
	if got := effectiveToughness(t, g, ball); got != 3 {
		t.Errorf("Speedball's toughness %d, want 3", got)
	}
}

func TestMshA8SpeedballRetargetMayBeDeclined(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ball := pushCatalogPermanent(g, me.ID, "Speedball, New Warrior", "Legendary Creature — Human Hero", mshA8SpeedballOracle, false)
	b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	spell := castCatalogSpell(t, g, "Djeru's Resolve", "Instant", pr7aDjerusResolve, []game.TargetRef{pr7aCard(ball)})
	passPriorityAroundTable(t, g)
	prompt := latestRetarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no retarget prompt")
	}
	if err := g.ResolveRetarget(prompt.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	if got := g.StackMeta[spell].Targets; len(got) != 1 || got[0].ID != ball {
		t.Errorf("declining moved the target: %+v", got)
	}
}

func TestMshA8SpeedballIgnoresASpellThatTargetsAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ball := pushCatalogPermanent(g, me.ID, "Speedball, New Warrior", "Legendary Creature — Human Hero", mshA8SpeedballOracle, false)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	base := effectivePower(t, g, ball)
	castCatalogSpell(t, g, "Djeru's Resolve", "Instant", pr7aDjerusResolve, []game.TargetRef{pr7aCard(other)})
	passPriorityAroundTable(t, g)
	if latestRetarget(g, me.ID) != nil || effectivePower(t, g, ball) != base {
		t.Error("a spell aimed at another creature triggered Speedball")
	}
}

// --- Secret Invasion -------------------------------------------------------------------------

func mshA8SecretInvasion(t *testing.T, g *game.Game) (aura, host, dragon uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	host = b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	dragon = b12Creature(g, opp.ID, "Elder Dragon", "Creature — Dragon", 5, 4)
	aura = castCatalogSpell(t, g, "Secret Invasion", "Enchantment — Aura", mshA8SecretInvasionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	passPriorityAroundTable(t, g)
	mshA8PickTargets(t, g, me.ID, dragon)
	return aura, host, dragon
}

func TestMshA8SecretInvasionExilesTheCreatureAndCopiesIt(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	_, host, dragon := mshA8SecretInvasion(t, g)
	if b12ZoneOf(g, dragon) != game.ZoneExile {
		t.Fatalf("the Dragon is in %q, want exile", b12ZoneOf(g, dragon))
	}
	c := gaLayered(t, g, host)
	if c.Name != "Elder Dragon" || c.CurrentPower() != 5 || c.CurrentToughness() != 4 || !c.HasSubtype("Dragon") {
		t.Errorf("the enchanted creature is %q %d/%d %q, want a copy of the Elder Dragon", c.Name, c.CurrentPower(), c.CurrentToughness(), c.TypeLine)
	}
	_ = opp
}

func TestMshA8SecretInvasionEndsWhenTheAuraLeaves(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	aura, host, dragon := mshA8SecretInvasion(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	g.RunStateChecksForTest()
	if n := mshA8Count(g, opp.ID, "Elder Dragon"); n != 1 {
		t.Errorf("the Dragon came back %d times under its owner's control, want 1", n)
	}
	if b12ZoneOf(g, dragon) == game.ZoneBattlefield {
		t.Log("the returned Dragon is a new object")
	}
	if c := gaLayered(t, g, host); c.Name != "My Bear" || c.CurrentPower() != 2 {
		t.Errorf("the creature is %q %d/%d, want the Bear again", c.Name, c.CurrentPower(), c.CurrentToughness())
	}
}

func TestMshA8SecretInvasionCannotTargetTheEnchantedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	host := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other Bear", "Creature — Bear", 3, 3)
	castCatalogSpell(t, g, "Secret Invasion", "Enchantment — Aura", mshA8SecretInvasionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.PickTargetCards {
		offered[id] = true
	}
	if offered[host] || !offered[other] {
		t.Errorf("offered %v, want the other creature and not the enchanted one", pick.PickTargetCards)
	}
}

// --- Super-Adaptoid --------------------------------------------------------------------------

func mshA8Adaptoid(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b12Push(g, owner, "Super-Adaptoid", "Legendary Artifact Creature — Robot Villain", mshA8AdaptoidOracle, 0, 2)
}

// mshA8CastAdaptoid casts Super-Adaptoid with its printed toughness of 2
// (a cast card carries none, and a 0-toughness creature dies).
func mshA8CastAdaptoid(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return castFromHand(t, g, game.Card{
		Name: "Super-Adaptoid", TypeLine: "Legendary Artifact Creature — Robot Villain",
		OracleID: mshA8AdaptoidOracle, Power: 0, Toughness: 2,
	}, game.CastSpellParams{})
}

func mshA8KeywordCreature(g *game.Game, owner uuid.UUID, keywords ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Target Creature", TypeLine: "Creature — Test",
		Power: 1, Toughness: 1, Keywords: keywords, Owner: owner, Controller: owner,
	})
}

func TestMshA8AdaptoidPowerIsTheLegendaryCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ada := mshA8Adaptoid(g, me.ID)
	if got := effectivePower(t, g, ada); got != 1 {
		t.Fatalf("alone, power %d, want 1 (itself)", got)
	}
	b12Creature(g, me.ID, "Hero One", "Legendary Creature — Hero", 2, 2)
	b12Creature(g, me.ID, "Hero Two", "Legendary Creature — Hero", 2, 2)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Hero", "Legendary Creature — Hero", 2, 2)
	if got := effectivePower(t, g, ada); got != 3 {
		t.Errorf("power %d, want 3 (itself and two other legendary creatures of mine)", got)
	}
	if got := effectiveToughness(t, g, ada); got != 2 {
		t.Errorf("toughness %d, want 2", got)
	}
}

func TestMshA8AdaptoidEntersAndTakesEachKeywordTheTargetHas(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := mshA8KeywordCreature(g, g.Seats[1].ID,
		"haste", "flying", "first strike", "double strike", "deathtouch", "indestructible",
		"lifelink", "menace", "reach", "trample", "vigilance")
	ada := mshA8CastAdaptoid(t, g)
	passPriorityAroundTable(t, g)
	mshA8PickTargets(t, g, me.ID, target)
	for _, kw := range mshA8AdaptoidKeywords {
		if got := countersOn(g, ada, kw); got != 1 {
			t.Errorf("%d %s counters, want 1", got, kw)
		}
		if !mshA3HasKeyword(g, ada, kw) {
			t.Errorf("Super-Adaptoid does not have %s", kw)
		}
	}
}

func TestMshA8AdaptoidOnlyTakesTheKeywordsTheTargetHas(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := mshA8KeywordCreature(g, g.Seats[1].ID, "flying", "lifelink")
	ada := mshA8CastAdaptoid(t, g)
	passPriorityAroundTable(t, g)
	mshA8PickTargets(t, g, me.ID, target)
	for _, kw := range mshA8AdaptoidKeywords {
		want := 0
		if kw == game.CounterFlying || kw == game.CounterLifelink {
			want = 1
		}
		if got := countersOn(g, ada, kw); got != want {
			t.Errorf("%d %s counters, want %d", got, kw, want)
		}
	}
}

func TestMshA8AdaptoidAttackingGivesNothingForAKeywordItAlreadyHas(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	ada := mshA8Adaptoid(g, me.ID)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ada, game.CounterFlying, 1) })
	target := mshA8KeywordCreature(g, opp.ID, "flying", "reach")
	declareAttack(t, g, opp.ID, ada)
	mshA8PickTargets(t, g, me.ID, target)
	if got := countersOn(g, ada, game.CounterFlying); got != 1 {
		t.Errorf("%d flying counters, want 1 (it already had flying)", got)
	}
	if got := countersOn(g, ada, game.CounterReach); got != 1 {
		t.Errorf("%d reach counters, want 1", got)
	}
}

func TestMshA8AdaptoidCannotTargetItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ada := mshA8CastAdaptoid(t, g)
	other := mshA8KeywordCreature(g, g.Seats[1].ID, "flying")
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	for _, id := range pick.PickTargetCards {
		if id == ada {
			t.Error("Super-Adaptoid was offered as its own target")
		}
	}
	found := false
	for _, id := range pick.PickTargetCards {
		found = found || id == other
	}
	if !found {
		t.Error("the other creature was not offered")
	}
}
