package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soulbond_cards_test.go — the four other soulbond cards of #5675, each
// paired through its own enters trigger and then driven.

const (
	sbondDoomWeaverOracle  = "172b94a0-2801-4623-aed8-902591c2d591"
	sbondMindbreakerOracle = "95157a46-6e0a-4800-93d3-46d2842464db"
	sbondSeraphOracle      = "a41e632f-9d36-44d2-b1dd-e17df0d2486e"
	sbondPhalanxOracle     = "d18dfca7-2947-4308-a9ef-501e1b4843f4"
)

// sbondPairCard casts a soulbond card and pairs it with `partner`
// through its enters trigger.
func sbondPairCard(t *testing.T, g *game.Game, me *game.Player, name, typeLine, oracle string, power, toughness int, partner uuid.UUID, keywords ...string) uuid.UUID {
	t.Helper()
	id := sbondHand(me, name, typeLine, oracle, power, toughness, append(keywords, game.KeywordSoulbond)...)
	sbondCast(t, g, me, id)
	nghcPassOnce(t, g)
	answerOwnPermanents(t, g, me.ID, partner)
	passPriorityAroundTable(t, g)
	if sbondPartner(g, id) != partner || sbondPartner(g, partner) != id {
		t.Fatalf("%s is not paired with the chosen creature", name)
	}
	return id
}

// sbondDie puts a permanent into its owner's graveyard and settles the
// stack.
func sbondDie(t *testing.T, g *game.Game, owner, id uuid.UUID) {
	t.Helper()
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: owner}, id); err != nil {
		t.Fatalf("die: %v", err)
	}
	b21SettleStack(t, g)
}

// sbondReady clears summoning sickness, so a creature cast this turn
// may attack.
func sbondReady(g *game.Game, ids ...uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			for _, id := range ids {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].SummonedThisTurn = false
				}
			}
		}
	})
}

// --- Doom Weaver ----------------------------------------------------------

func TestDoomWeaverHasReachAndSoulbond(t *testing.T) {
	g := newCatalogGame(t)
	id := b12Push(g, g.Seats[0].ID, "Doom Weaver", "Creature — Spider Horror", sbondDoomWeaverOracle, 1, 8)
	for _, kw := range []string{"reach", game.KeywordSoulbond} {
		if !hasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("Doom Weaver printed %s", kw)
		}
	}
}

func TestDoomWeaverPairedCreatureDrawsItsPowerWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	weaver := sbondPairCard(t, g, me, "Doom Weaver", "Creature — Spider Horror", sbondDoomWeaverOracle, 1, 8, bear, "reach")

	hand := me.Hand.Size()
	sbondDie(t, g, me.ID, bear)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Fatalf("the paired 3/3 died: drew %d, want 3", got)
	}
	// The Weaver is unpaired now, so its own death draws nothing.
	hand = me.Hand.Size()
	sbondDie(t, g, me.ID, weaver)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("an unpaired Doom Weaver died: drew %d, want 0", got)
	}
}

func TestDoomWeaverItselfDrawsItsPowerWhilePaired(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	weaver := sbondPairCard(t, g, me, "Doom Weaver", "Creature — Spider Horror", sbondDoomWeaverOracle, 1, 8, bear, "reach")
	hand := me.Hand.Size()
	sbondDie(t, g, me.ID, weaver)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Fatalf("a paired Doom Weaver (power 1) died: drew %d, want 1", got)
	}
	hand = me.Hand.Size()
	sbondDie(t, g, me.ID, bear)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("the old partner died unpaired: drew %d, want 0", got)
	}
}

// --- Imperious Mindbreaker ------------------------------------------------

func TestImperiousMindbreakerEachPairedAttackerMillsItsToughness(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Wolf", "Creature — Wolf", 3, 3)
	mind := sbondPairCard(t, g, me, "Imperious Mindbreaker", "Creature — Human Wizard", sbondMindbreakerOracle, 1, 4, bear)
	sbondReady(g, mind)

	before := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		before[i] = p.Graveyard.Size()
	}
	declareAttack(t, g, opp.ID, mind, bear, other)
	b21SettleStack(t, g)
	// Mindbreaker's toughness 4 and the Bear's 2; the unpaired Wolf
	// mills nothing. Every opponent mills, not only the one attacked.
	for i, p := range g.Seats {
		want := 6
		if p.ID == me.ID {
			want = 0
		}
		if got := p.Graveyard.Size() - before[i]; got != want {
			t.Errorf("seat %d milled %d, want %d", i, got, want)
		}
	}
}

func TestImperiousMindbreakerUnpairedMillsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	mind := b12Push(g, me.ID, "Imperious Mindbreaker", "Creature — Human Wizard", sbondMindbreakerOracle, 1, 4)
	before := opp.Graveyard.Size()
	declareAttack(t, g, opp.ID, mind)
	b21SettleStack(t, g)
	if got := opp.Graveyard.Size() - before; got != 0 {
		t.Errorf("an unpaired Mindbreaker milled %d", got)
	}
}

// --- Breathkeeper Seraph --------------------------------------------------

func TestBreathkeeperSeraphHasFlyingAndSoulbond(t *testing.T) {
	g := newCatalogGame(t)
	id := b12Push(g, g.Seats[0].ID, "Breathkeeper Seraph", "Creature — Angel", sbondSeraphOracle, 4, 4)
	for _, kw := range []string{"flying", game.KeywordSoulbond} {
		if !hasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("Breathkeeper Seraph printed %s", kw)
		}
	}
}

// sbondSeraphBoard pairs a Seraph with a Bear and kills the Bear.
func sbondSeraphBoard(t *testing.T) (g *game.Game, me *game.Player, seraph, bear uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	bear = b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	seraph = sbondPairCard(t, g, me, "Breathkeeper Seraph", "Creature — Angel", sbondSeraphOracle, 4, 4, bear, "flying")
	sbondDie(t, g, me.ID, bear)
	if !appInZone(g, bear, game.ZoneGraveyard) {
		t.Fatal("the Bear is not in the graveyard")
	}
	return g, me, seraph, bear
}

// sbondToMyUpkeep walks to seat 0's next upkeep and puts the delayed
// trigger on the stack and resolves it, stopping at its question.
func sbondToMyUpkeep(t *testing.T, g *game.Game) {
	t.Helper()
	endTurn(t, g)
	appToUpkeepOf(t, g, 0)
	for i := 0; i < 16 && !stackFullyEmpty(g); i++ {
		if latestChoiceOfKindFor(g, game.PendingChoiceConfirm, g.Seats[0].ID) != nil {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
}

func TestBreathkeeperSeraphReturnsThePairedCreatureAtYourNextUpkeep(t *testing.T) {
	g, me, seraph, bear := sbondSeraphBoard(t)
	// Not at an opponent's upkeep.
	endTurn(t, g)
	appToUpkeepOf(t, g, 1)
	b21SettleStack(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID) != nil || !appInZone(g, bear, game.ZoneGraveyard) {
		t.Fatal("the return waits for YOUR next upkeep")
	}
	appToUpkeepOf(t, g, 0)
	for i := 0; i < 16 && latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	answerConfirmFor(t, g, me, true)
	b21SettleStack(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the Bear did not return to the battlefield")
	}
	if c, _ := battlefieldCard(g, bear); c.Controller != me.ID {
		t.Error("it returns under its owner's control")
	}
	if sbondPartner(g, bear) != uuid.Nil || sbondPartner(g, seraph) != uuid.Nil {
		t.Error("the returned creature is a new object and is unpaired")
	}
}

func TestBreathkeeperSeraphDecliningLeavesTheCardInTheGraveyard(t *testing.T) {
	g, me, _, bear := sbondSeraphBoard(t)
	sbondToMyUpkeep(t, g)
	answerConfirmFor(t, g, me, false)
	b21SettleStack(t, g)
	if !appInZone(g, bear, game.ZoneGraveyard) {
		t.Error("\"you may\": a decline leaves the card in the graveyard")
	}
}

func TestBreathkeeperSeraphDoesNotReturnACardThatLeftTheGraveyard(t *testing.T) {
	g, me, _, bear := sbondSeraphBoard(t)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneExile}, bear); err != nil {
		t.Fatalf("exile the Bear: %v", err)
	}
	sbondToMyUpkeep(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID) != nil {
		t.Error("the delayed trigger asked about a card that is no longer in the graveyard")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("an exiled card returned")
	}
}

// "Under its owner's control", and "your next upkeep" is the upkeep of
// the player who controlled the creature when it died.
func TestBreathkeeperSeraphReturnsAStolenCreatureToItsOwner(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if !g.GainControlForEffect(uuid.New(), theirs, me.ID, game.IndefiniteDuration(), "test — steal for good") {
			t.Fatal("GainControlForEffect refused")
		}
	})
	g.ReadSnapshot(func() {})
	sbondPairCard(t, g, me, "Breathkeeper Seraph", "Creature — Angel", sbondSeraphOracle, 4, 4, theirs, "flying")
	sbondDie(t, g, opp.ID, theirs)
	sbondToMyUpkeep(t, g)
	answerConfirmFor(t, g, me, true)
	b21SettleStack(t, g)
	c, ok := battlefieldCard(g, theirs)
	if !ok {
		t.Fatal("the creature did not return")
	}
	if c.Controller != opp.ID {
		t.Error("it returns under its OWNER's control, not the Seraph player's")
	}
}

func TestBreathkeeperSeraphUnpairedCreatureDoesNotReturn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	seraph := b12Push(g, me.ID, "Breathkeeper Seraph", "Creature — Angel", sbondSeraphOracle, 4, 4)
	sbondDie(t, g, me.ID, seraph)
	sbondToMyUpkeep(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID) != nil || g.Battlefield.Contains(seraph) {
		t.Error("an unpaired Seraph has no dies ability")
	}
}

// --- Mirage Phalanx -------------------------------------------------------

// sbondTokens lists the tokens `controller` controls that are not in
// `known`.
func sbondTokens(g *game.Game, controller uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.IsToken() {
			out = append(out, c)
		}
	}
	return out
}

func TestMiragePhalanxEachPairedCreatureCopiesItselfAtYourCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	phalanx := sbondPairCard(t, g, me, "Mirage Phalanx", "Creature — Human Soldier", sbondPhalanxOracle, 4, 4, bear)

	advanceTo(t, g, game.StepBeginCombat)
	b21SettleStack(t, g)
	tokens := sbondTokens(g, me.ID)
	if len(tokens) != 2 {
		t.Fatalf("each paired creature makes one copy of itself: %d tokens", len(tokens))
	}
	names := map[string]game.Card{}
	for _, tok := range tokens {
		names[tok.Name] = tok
	}
	copyOfPhalanx, okP := names["Mirage Phalanx"]
	copyOfBear, okB := names["Bear"]
	if !okP || !okB {
		t.Fatalf("one copy of each: %v", names)
	}
	for _, tok := range []game.Card{copyOfPhalanx, copyOfBear} {
		if !hasEffectiveKeyword(t, g, tok.InstanceID, "haste") {
			t.Errorf("the copy of %s has haste", tok.Name)
		}
		if hasEffectiveKeyword(t, g, tok.InstanceID, game.KeywordSoulbond) {
			t.Errorf("the copy of %s lost soulbond", tok.Name)
		}
		if sbondPartner(g, tok.InstanceID) != uuid.Nil {
			t.Errorf("the copy of %s is unpaired: the pair is not copiable", tok.Name)
		}
	}
	if !hasEffectiveKeyword(t, g, phalanx, game.KeywordSoulbond) {
		t.Error("the original keeps soulbond")
	}
	if sbondPartner(g, phalanx) != bear {
		t.Error("making the copies ended the pair")
	}
	if !stackFullyEmpty(g) {
		t.Error("a token without soulbond asked to pair")
	}

	// "Exile it at end of combat."
	advanceTo(t, g, game.StepEndCombat)
	b21SettleStack(t, g)
	if left := sbondTokens(g, me.ID); len(left) != 0 {
		t.Errorf("%d tokens survived the end of combat", len(left))
	}
	if !g.Battlefield.Contains(phalanx) || !g.Battlefield.Contains(bear) {
		t.Error("the originals stay")
	}
}

// A copy "loses soulbond" for good: a creature entering beside the
// token does not trigger it, and the loss survives a restore.
func TestMiragePhalanxCopyNeverAsksToPair(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sbondPairCard(t, g, me, "Mirage Phalanx", "Creature — Human Soldier", sbondPhalanxOracle, 4, 4, bear)
	advanceTo(t, g, game.StepBeginCombat)
	b21SettleStack(t, g)

	g = restoreThroughJSON(t, g)
	me = g.Seats[0]
	var token uuid.UUID
	for _, tok := range sbondTokens(g, me.ID) {
		if tok.Name == "Mirage Phalanx" {
			token = tok.InstanceID
		}
	}
	if token == uuid.Nil {
		t.Fatal("no copy of Mirage Phalanx after the restore")
	}
	if hasEffectiveKeyword(t, g, token, game.KeywordSoulbond) {
		t.Fatal("the restored copy has soulbond again")
	}
	enterFromHand(t, g, me.ID, "Wolf", "Creature — Wolf", "")
	if !stackFullyEmpty(g) {
		t.Error("a copy that lost soulbond triggered for a creature that entered")
	}
}

func TestMiragePhalanxUnpairedMakesNoCopy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	b12Push(g, me.ID, "Mirage Phalanx", "Creature — Human Soldier", sbondPhalanxOracle, 4, 4)
	advanceTo(t, g, game.StepBeginCombat)
	b21SettleStack(t, g)
	if n := len(sbondTokens(g, me.ID)); n != 0 {
		t.Errorf("an unpaired Mirage Phalanx made %d tokens", n)
	}
}
