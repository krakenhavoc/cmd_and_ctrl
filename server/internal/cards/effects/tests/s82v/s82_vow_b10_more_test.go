package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_vow_b10_more_test.go — the rest of Innistrad: Crimson Vow batch 10
// (#5495): the triggered, mill and library-look cards.

// --- Dying to Serve -----------------------------------------------------------

func vowB10Discard(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		id := effectstest.HandCardForTest(p, "Discarded", vowB10SorceryTypeLine, "")
		g.WithWriteLock(func() {
			if err := g.DiscardCardsKeyedForEffect(p.ID, []uuid.UUID{id}, uuid.Nil, game.KeyedTail{}); err != nil {
				t.Fatalf("discard: %v", err)
			}
		})
	}
}

func vowB10Zombies(g *game.Game, p *game.Player) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == p.ID && c.Name == "Zombie" {
			out = append(out, c)
		}
	}
	return out
}

func TestVowB10DyingToServeMakesOneTappedZombieOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Dying to Serve", TypeLine: "Enchantment", OracleID: vowB10DyingOracle, Owner: me.ID, Controller: me.ID})
	vowB10Discard(t, g, me, 1)
	effectstest.PassPriorityAroundTable(t, g)
	zs := vowB10Zombies(g, me)
	if len(zs) != 1 {
		t.Fatalf("%d Zombie tokens after a discard, want 1", len(zs))
	}
	if !zs[0].Tapped || zs[0].Power != 2 || zs[0].Toughness != 2 {
		t.Errorf("token is tapped=%v %d/%d, want a tapped 2/2", zs[0].Tapped, zs[0].Power, zs[0].Toughness)
	}
	vowB10Discard(t, g, me, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(vowB10Zombies(g, me)); n != 1 {
		t.Errorf("%d Zombies after a second discard the same turn, want still 1", n)
	}
}

func TestVowB10DyingToServeIgnoresAnOpponentsDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Dying to Serve", TypeLine: "Enchantment", OracleID: vowB10DyingOracle, Owner: me.ID, Controller: me.ID})
	vowB10Discard(t, g, opp, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(vowB10Zombies(g, me)); n != 0 {
		t.Errorf("%d Zombies from an opponent's discard, want 0", n)
	}
}

// A batch of cards discarded together is one trigger.
func TestVowB10DyingToServeOneOrMoreIsOneTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Dying to Serve", TypeLine: "Enchantment", OracleID: vowB10DyingOracle, Owner: me.ID, Controller: me.ID})
	ids := []uuid.UUID{
		effectstest.HandCardForTest(me, "A", vowB10SorceryTypeLine, ""),
		effectstest.HandCardForTest(me, "B", vowB10SorceryTypeLine, ""),
		effectstest.HandCardForTest(me, "C", vowB10SorceryTypeLine, ""),
	}
	g.WithWriteLock(func() {
		if err := g.DiscardCardsKeyedForEffect(me.ID, ids, uuid.Nil, game.KeyedTail{}); err != nil {
			t.Fatalf("discard: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(vowB10Zombies(g, me)); n != 1 {
		t.Errorf("%d Zombies from three simultaneous discards, want 1", n)
	}
}

// --- Wretched Throng ----------------------------------------------------------

func TestVowB10WretchedThrongFindsAnotherWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	throng := effectstest.B12Push(g, me.ID, "Wretched Throng", "Creature — Zombie Horror", vowB10ThrongOracle, 2, 1)
	other := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: other, Name: "Wretched Throng", TypeLine: "Creature — Zombie Horror", OracleID: vowB10ThrongOracle, Owner: me.ID, Controller: me.ID})
	decoy := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: decoy, Name: "Wretched Thing", TypeLine: "Creature — Zombie", Owner: me.ID, Controller: me.ID})
	effectstest.KillCreature(t, g, me.ID, throng)
	effectstest.PassPriorityAroundTable(t, g)
	search := effectstest.SearchChoiceFor(g, me.ID)
	if search == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	if err := g.BeginSearchLibrary(search.ID, me.ID); err != nil {
		t.Fatalf("begin the search: %v", err)
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{decoy}); err == nil {
		t.Fatal("a card with another name was accepted")
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{other}); err != nil {
		t.Fatalf("search: %v", err)
	}
	g.SettleResolution()
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(other) {
		t.Fatalf("the second Wretched Throng is in %q, not the hand", effectstest.ZoneOf(g, other))
	}
	if me.Hand.Contains(decoy) {
		t.Error("a card with another name was fetched")
	}
}

func TestVowB10WretchedThrongMayDeclineTheSearch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	throng := effectstest.B12Push(g, me.ID, "Wretched Throng", "Creature — Zombie Horror", vowB10ThrongOracle, 2, 1)
	other := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: other, Name: "Wretched Throng", TypeLine: "Creature — Zombie Horror", OracleID: vowB10ThrongOracle, Owner: me.ID, Controller: me.ID})
	effectstest.KillCreature(t, g, me.ID, throng)
	effectstest.PassPriorityAroundTable(t, g)
	search := effectstest.SearchChoiceFor(g, me.ID)
	if search == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	if err := g.DeclineSearchLibrary(search.ID, me.ID); err != nil {
		t.Fatalf("decline the search: %v", err)
	}
	g.SettleResolution()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(other) {
		t.Error("the search was declined but a card was taken")
	}
}

// --- Archghoul of Thraben -----------------------------------------------------

func TestVowB10ArchghoulTakesATopZombieWhenAZombieDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Archghoul of Thraben", vowB10ZombieTypeLine, vowB10ArchghoulOracle, 3, 2)
	victim := effectstest.B12Push(g, me.ID, "Other Zombie", vowB10ZombieTypeLine, "", 1, 1)
	top := effectstest.SpmB17Top(me, game.Card{Name: "Top Zombie", TypeLine: "Creature — Zombie"})
	effectstest.KillCreature(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(top) {
		t.Fatalf("the top Zombie is in %q, not the hand", effectstest.ZoneOf(g, top))
	}
}

func TestVowB10ArchghoulTriggersOnItsOwnDeathAndOffersTheGraveyardForANonZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ghoul := effectstest.B12Push(g, me.ID, "Archghoul of Thraben", vowB10ZombieTypeLine, vowB10ArchghoulOracle, 3, 2)
	top := effectstest.SpmB17Top(me, game.Card{Name: "Top Bear", TypeLine: vowB10CreatureTypeLine})
	effectstest.KillCreature(t, g, me.ID, ghoul)
	effectstest.PassPriorityAroundTable(t, g)
	// A non-Zombie goes straight to the graveyard offer.
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(top) {
		t.Fatalf("the declined card is in %q, want the graveyard", effectstest.ZoneOf(g, top))
	}
	if me.Hand.Contains(top) {
		t.Error("a non-Zombie went to the hand")
	}
}

func TestVowB10ArchghoulDecliningBothLeavesTheCardOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Archghoul of Thraben", vowB10ZombieTypeLine, vowB10ArchghoulOracle, 3, 2)
	victim := effectstest.B12Push(g, me.ID, "Other Zombie", vowB10ZombieTypeLine, "", 1, 1)
	top := effectstest.SpmB17Top(me, game.Card{Name: "Top Zombie", TypeLine: "Creature — Zombie"})
	effectstest.KillCreature(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got, err := me.Library.Top(); err != nil || got.InstanceID != top {
		t.Error("the card declined twice is not on top of the library")
	}
}

func TestVowB10ArchghoulIgnoresANonZombieAndAnOpponentsZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Archghoul of Thraben", vowB10ZombieTypeLine, vowB10ArchghoulOracle, 3, 2)
	bear := effectstest.B12Push(g, me.ID, "Bear", vowB10CreatureTypeLine, "", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Their Zombie", vowB10ZombieTypeLine, "", 1, 1)
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("a creature that was not your Zombie asked for a look")
	}
}

// --- Skywarp Skaab ------------------------------------------------------------

func vowB10CastSkaab(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	if _, err := vowB10Cast(t, g, me, game.Card{Name: "Skywarp Skaab", TypeLine: "Creature — Zombie Drake", ManaCost: "{3}{U}{U}", OracleID: vowB10SkaabOracle, Power: 2, Toughness: 5}, "CCCUU", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestVowB10SkywarpSkaabExilesTwoCreatureCardsToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := vowB10GraveyardCreatures(me, 3)
	rock := effectstest.PushCatalogGraveyardCard(me, "Dead Rock", vowB10ArtifactTypeLine, "", 0, 0)
	hand := me.Hand.Size()
	vowB10CastSkaab(t, g, me)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.AnswerChooseCards(t, g, me.ID, ids[0], ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range ids[:2] {
		if effectstest.ZoneOf(g, id) != game.ZoneExile {
			t.Errorf("a chosen creature card is in %q, want exile", effectstest.ZoneOf(g, id))
		}
	}
	if effectstest.ZoneOf(g, ids[2]) != game.ZoneGraveyard || effectstest.ZoneOf(g, rock) != game.ZoneGraveyard {
		t.Error("an unchosen card left the graveyard")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (Skaab cast, one card drawn)", me.Hand.Size(), hand+1)
	}
}

func TestVowB10SkywarpSkaabDeclinedExilesNothingAndDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := vowB10GraveyardCreatures(me, 2)
	hand := me.Hand.Size()
	vowB10CastSkaab(t, g, me)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range ids {
		if effectstest.ZoneOf(g, id) != game.ZoneGraveyard {
			t.Errorf("a creature card is in %q after declining", effectstest.ZoneOf(g, id))
		}
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d: nothing is drawn", me.Hand.Size(), hand)
	}
}

func TestVowB10SkywarpSkaabWithOneCreatureCardDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := vowB10GraveyardCreatures(me, 1)
	hand := me.Hand.Size()
	vowB10CastSkaab(t, g, me)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOf(g, ids[0]) != game.ZoneGraveyard || me.Hand.Size() != hand {
		t.Error("one creature card is not enough to pay for the draw")
	}
}

// --- Screaming Swarm ----------------------------------------------------------

func TestVowB10ScreamingSwarmMillsAsManyAsAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	swarm := effectstest.B12Push(g, me.ID, "Screaming Swarm", "Creature — Bird Horror", vowB10SwarmOracle, 4, 4)
	a := effectstest.PushVanillaCreature(g, me.ID, "A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "B", 2, 2)
	effectstest.AdvanceToMain(t, g)
	library, graveyard := opp.Library.Size(), opp.Graveyard.Size()
	effectstest.DeclareAttack(t, g, opp.ID, swarm, a, b)
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := library - opp.Library.Size(); got != 3 {
		t.Errorf("milled %d, want 3 (one per attacker)", got)
	}
	if got := opp.Graveyard.Size() - graveyard; got != 3 {
		t.Errorf("graveyard grew by %d, want 3", got)
	}
}

func TestVowB10ScreamingSwarmTuckSecondFromTheTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMainForCost(t, g)
	id := effectstest.PushCatalogGraveyardCard(me, "Screaming Swarm", "Creature — Bird Horror", vowB10SwarmOracle, 4, 4)
	effectstest.FloatForTest(g, me, "CCU")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	library := me.Library.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOf(g, id) != game.ZoneLibrary {
		t.Fatalf("Swarm is in %q, want the library", effectstest.ZoneOf(g, id))
	}
	if me.Library.Size() != library+1 {
		t.Errorf("library %d, want %d", me.Library.Size(), library+1)
	}
	if got := me.Library.Cards[me.Library.Size()-2].InstanceID; got != id {
		t.Error("the Swarm is not second from the top")
	}
}

// --- Scattered Thoughts -------------------------------------------------------

func TestVowB10ScatteredThoughtsTakesTwoOfTheTopFourAndBinsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	d := effectstest.SpmB17Top(me, game.Card{Name: "D", TypeLine: vowB10SorceryTypeLine})
	c := effectstest.SpmB17Top(me, game.Card{Name: "C", TypeLine: vowB10SorceryTypeLine})
	b := effectstest.SpmB17Top(me, game.Card{Name: "B", TypeLine: vowB10SorceryTypeLine})
	a := effectstest.SpmB17Top(me, game.Card{Name: "A", TypeLine: vowB10SorceryTypeLine})
	below := me.Library.Cards[me.Library.Size()-5].InstanceID
	effectstest.CastCatalogSpell(t, g, "Scattered Thoughts", "Instant", vowB10ScatteredOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, a, c)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, c} {
		if !me.Hand.Contains(id) {
			t.Errorf("a chosen card is in %q, want the hand", effectstest.ZoneOf(g, id))
		}
	}
	for _, id := range []uuid.UUID{b, d} {
		if !me.Graveyard.Contains(id) {
			t.Errorf("an unchosen card is in %q, want the graveyard", effectstest.ZoneOf(g, id))
		}
	}
	if effectstest.ZoneOf(g, below) != game.ZoneLibrary {
		t.Error("the fifth card left the library")
	}
}

// --- Witness the Future -------------------------------------------------------

func TestVowB10WitnessTheFutureShufflesAPlayersCardsBackThenTakesOneOfFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := []uuid.UUID{
		effectstest.PushCatalogGraveyardCard(opp, "T1", vowB10CreatureTypeLine, "", 1, 1),
		effectstest.PushCatalogGraveyardCard(opp, "T2", vowB10CreatureTypeLine, "", 1, 1),
	}
	d := effectstest.SpmB17Top(me, game.Card{Name: "D", TypeLine: vowB10SorceryTypeLine})
	c := effectstest.SpmB17Top(me, game.Card{Name: "C", TypeLine: vowB10SorceryTypeLine})
	b := effectstest.SpmB17Top(me, game.Card{Name: "B", TypeLine: vowB10SorceryTypeLine})
	a := effectstest.SpmB17Top(me, game.Card{Name: "A", TypeLine: vowB10SorceryTypeLine})
	oppLibrary := opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Witness the Future", vowB10SorceryTypeLine, vowB10WitnessOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
		{Kind: game.TargetCard, ID: theirs[0], Slot: 1},
		{Kind: game.TargetCard, ID: theirs[1], Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, b)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range theirs {
		if effectstest.ZoneOf(g, id) != game.ZoneLibrary {
			t.Errorf("a targeted card is in %q, want the target player's library", effectstest.ZoneOf(g, id))
		}
	}
	if opp.Library.Size() != oppLibrary+2 {
		t.Errorf("their library %d, want %d", opp.Library.Size(), oppLibrary+2)
	}
	if !me.Hand.Contains(b) {
		t.Errorf("the chosen card is in %q, want the hand", effectstest.ZoneOf(g, b))
	}
	// The rest go to the bottom of YOUR library, not the graveyard.
	n := me.Library.Size()
	bottom := map[uuid.UUID]bool{}
	for i := 0; i < 3 && i < n; i++ {
		bottom[me.Library.Cards[i].InstanceID] = true
	}
	for _, id := range []uuid.UUID{a, c, d} {
		if !bottom[id] {
			t.Errorf("an unchosen card is not among the bottom three of the library (zone %q)", effectstest.ZoneOf(g, id))
		}
	}
}

// --- Retrieve -----------------------------------------------------------------

func TestVowB10RetrieveReturnsACreatureAndANoncreaturePermanentThenExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushCatalogGraveyardCard(me, "Dead Bear", vowB10CreatureTypeLine, "", 2, 2)
	rock := effectstest.PushCatalogGraveyardCard(me, "Dead Rock", vowB10ArtifactTypeLine, "", 0, 0)
	bolt := effectstest.PushCatalogGraveyardCard(me, "Dead Bolt", "Instant", "", 0, 0)
	spell := effectstest.CastCatalogSpell(t, g, "Retrieve", vowB10SorceryTypeLine, vowB10RetrieveOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear, Slot: 0},
		{Kind: game.TargetCard, ID: rock, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) || !me.Hand.Contains(rock) {
		t.Errorf("bear in %q, rock in %q; both belong in the hand", effectstest.ZoneOf(g, bear), effectstest.ZoneOf(g, rock))
	}
	if effectstest.ZoneOf(g, bolt) != game.ZoneGraveyard {
		t.Error("an untargeted card left the graveyard")
	}
	if effectstest.ZoneOf(g, spell) != game.ZoneExile {
		t.Errorf("Retrieve is in %q, want exile", effectstest.ZoneOf(g, spell))
	}
}

func TestVowB10RetrieveTakesEitherHalfAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.PushCatalogGraveyardCard(me, "Dead Rock", vowB10ArtifactTypeLine, "", 0, 0)
	spell := effectstest.CastCatalogSpell(t, g, "Retrieve", vowB10SorceryTypeLine, vowB10RetrieveOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: rock, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) {
		t.Errorf("the rock is in %q, want the hand", effectstest.ZoneOf(g, rock))
	}
	if effectstest.ZoneOf(g, spell) != game.ZoneExile {
		t.Errorf("Retrieve is in %q, want exile", effectstest.ZoneOf(g, spell))
	}
}

func TestVowB10RetrieveRefusesACreatureInTheNoncreatureSlot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushCatalogGraveyardCard(me, "Dead Bear", vowB10CreatureTypeLine, "", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Retrieve", vowB10SorceryTypeLine, vowB10RetrieveOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear, Slot: 1},
	}); err == nil {
		t.Error("a creature card was accepted as the noncreature permanent")
	}
}
