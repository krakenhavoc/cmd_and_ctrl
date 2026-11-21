package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsh "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/h"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- a discard ---------------------------------------------------------------

// An opponent's Mind Rot over a filler and a Needle: the Needle's discard
// asks for the order, the restored table has the filler in the graveyard
// and the Needle in hand, and the answer discards it.
func TestAMultiCardDiscardIsARestorePointAtTheSecondCardsOrder(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	filler := d.Filler(d.Victim)
	needle := d.HandCard(d.Victim, "Needle", "Sorcery", "{1}", "", 0, 0)
	effectstest.SeedOrderProbes(d.G, d.Victim.ID, effectstest.GraveyardOrderProbeOracles, "Graveyard Order Probe")
	effectstest.CastCatalogSpell(t, d.G, "Mind Rot", "Sorcery", effectstest.MindRotOracleID, []game.TargetRef{{Kind: game.TargetPlayer, ID: d.Victim.ID}})
	effectstest.PassPriorityAroundTable(t, d.G)
	effectstest.AnswerDiscard(t, d.G, d.Victim.ID, filler, needle)

	g := effectstest.StrictRestoreWithFrame(t, d.G, game.ReplacementResumeDiscard)
	victim := effectstest.SeatByID(g, d.Victim.ID)
	if !victim.Graveyard.Contains(filler) {
		t.Fatal("the filler is not in the graveyard on the restored table")
	}
	if !victim.Hand.Contains(needle) {
		t.Fatal("the Needle left the hand before the restored answer")
	}
	effectstest.AnswerExileOrder(t, g)
	if !victim.Graveyard.Contains(needle) {
		t.Fatal("the Needle was not discarded after the restored answer")
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts open after the discard finished", len(g.PendingChoices))
	}
}

func TestAKeyedDiscardsThenRunsAfterARestoreMidBatch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	d := effectstest.DboTable{G: g, Caster: me, Victim: me}
	d.HandCard(me, "Needle", "Sorcery", "{1}", "", 0, 0)
	d.HandCard(me, "Needle", "Sorcery", "{1}", "", 0, 0)
	effectstest.SeedOrderProbes(g, me.ID, effectstest.GraveyardOrderProbeOracles, "Graveyard Order Probe")
	life := me.Life
	g.WithWriteLock(func() {
		if err := g.DiscardRandomKeyedForEffect(me.ID, 2, game.KeyedTail{Ref: effectstest.TestDiscardedGainLife, Player: me.ID}); err != nil {
			t.Fatalf("DiscardRandomKeyedForEffect: %v", err)
		}
	})
	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeDiscard)
	effectstest.AnswerExileOrder(t, g)
	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeDiscard)
	effectstest.AnswerExileOrder(t, g)
	me = g.Seats[0]
	if me.Graveyard.Size() < 2 {
		t.Fatalf("graveyard %d after both answers, want both Needles", me.Graveyard.Size())
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2 from the keyed then", life, me.Life)
	}
}

// --- a mill ------------------------------------------------------------------

// Hama, the Bloodbender's "target opponent mills three cards; exile up to
// one noncreature, nonland card from their graveyard": the second milled
// card asks, the restored table has the first in the graveyard and the
// rest on the library, and the answer finishes the mill and opens Hama's
// pick, the mill's registered tail.
func TestAMillIsARestorePointAtTheSecondCardsOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hama := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hama, the Bloodbender", TypeLine: "Legendary Creature — Human Warlock",
		OracleID: "ef80e174-78e8-412d-a6fe-89c7af746669", Owner: me.ID, Controller: me.ID,
	})
	third := effectstest.LibTop(opp, game.Card{Name: "Ponder", TypeLine: "Sorcery"})
	needle := effectstest.PushLibraryTopForTest(opp, "Needle")
	first := effectstest.PushLibraryTopForTest(opp, "First")
	effectstest.SeedOrderProbes(g, opp.ID, effectstest.GraveyardOrderProbeOracles, "Graveyard Order Probe")
	item := &game.StackItem{ID: uuid.New(), Controller: me.ID, SourceCardID: hama, Label: "Hama",
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}
	g.WithWriteLock(func() {
		if err := cardsh.HamaEnters(g, item); err != nil {
			t.Fatalf("hamaEnters: %v", err)
		}
	})

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	opp = effectstest.SeatByID(g, opp.ID)
	if !opp.Graveyard.Contains(first) {
		t.Fatal("the first milled card is not in the graveyard on the restored table")
	}
	if !opp.Library.Contains(needle) || !opp.Library.Contains(third) {
		t.Fatal("the restored table milled the Needle or the third card before the answer")
	}
	effectstest.AnswerExileOrder(t, g)
	for _, id := range []uuid.UUID{first, needle, third} {
		if !opp.Graveyard.Contains(id) {
			t.Errorf("%s is not in the graveyard after the restored answer", id)
		}
	}
	var pick *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c.Kind == game.PendingChoiceChooseCards && c.Chooser == me.ID {
			pick = c
		}
	}
	if pick == nil || len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != third {
		t.Fatalf("Hama's pick over the Ponder did not open after the mill: %+v", g.PendingChoices)
	}
}

// "Exile cards from the top of your library until you exile a nonland card"
// as a keyed run (millUntilRun, its stop clause until/nonland-card): the
// second repetition's Needle asks, the restored table has the Forest in
// exile and the Needle on the library, and the answer exiles it, ends the
// run on it (the next card stays) and runs the tail with both cards.
func TestAnExileUntilRunIsARestorePointMidRun(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	next := effectstest.PushLibraryTopForTest(me, "Next")
	needle := effectstest.PushLibraryTopForTest(me, "Needle")
	forest := effectstest.LibTop(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.SeedExileOrderProbes(g, me.ID)
	life := me.Life
	g.WithWriteLock(func() {
		if err := g.MillToZoneKeyedForEffect(me.ID, 0, game.ZoneExile, effects.UntilNonland,
			game.KeyedTail{Ref: effectstest.TestDiscardedGainLife, Player: me.ID}); err != nil {
			t.Fatalf("MillToZoneKeyedForEffect: %v", err)
		}
	})

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	me = g.Seats[0]
	if _, ok := effectstest.FindInZone(g.Exile, forest); !ok {
		t.Fatal("the Forest is not in exile on the restored table")
	}
	if !me.Library.Contains(needle) {
		t.Fatal("the restored table exiled the Needle before the answer")
	}
	effectstest.AnswerExileOrder(t, g)
	if _, ok := effectstest.FindInZone(g.Exile, needle); !ok {
		t.Fatal("the Needle was not exiled after the restored answer")
	}
	if !me.Library.Contains(next) {
		t.Error("the run went past the nonland card")
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2 for the two cards the run exiled", life, me.Life)
	}
}

// --- an exile batch ----------------------------------------------------------

// Crypt Incursion over two creature cards, the second a Needle: the
// Needle's exile asks, the restored table has the first in exile, and the
// answer exiles the Needle and gains 6 life (crypt-incursion/gain-3-per-
// exiled, run off the disk).
func TestCryptIncursionIsARestorePointAtTheSecondCardsOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	first, needle := uuid.New(), uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: first, Name: "First", TypeLine: "Creature — Test", Owner: opp.ID, Controller: opp.ID})
	opp.Graveyard.PushTop(game.Card{InstanceID: needle, Name: "Needle", TypeLine: "Creature — Test", Owner: opp.ID, Controller: opp.ID})
	effectstest.SeedExileOrderProbes(g, opp.ID)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Crypt Incursion", "Instant", effectstest.B42CryptIncursionOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	me, opp = effectstest.SeatByID(g, me.ID), effectstest.SeatByID(g, opp.ID)
	if _, ok := effectstest.FindInZone(g.Exile, first); !ok {
		t.Fatal("the first creature card is not in exile on the restored table")
	}
	if !opp.Graveyard.Contains(needle) {
		t.Fatal("the restored table exiled the Needle before the answer")
	}
	effectstest.AnswerExileOrder(t, g)
	if _, ok := effectstest.FindInZone(g.Exile, needle); !ok {
		t.Fatal("the Needle was not exiled after the restored answer")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+6 {
		t.Errorf("life %d -> %d, want +6 for the two cards exiled", life, me.Life)
	}
}

// --- manifest dread ----------------------------------------------------------

const manifestDreadOracle = "df5a66d1-61f6-44d9-b464-3ee739e78dce"

// Manifest Dread over a First and a Needle: the pick is a restore point,
// the First enters face down, the Needle's move to the graveyard asks, the
// restored table has the First on the battlefield and the Needle on the
// library, and the answer puts it into the graveyard and announces the
// manifest dread.
func TestManifestDreadIsARestorePointAtThePickAndAtTheGraveyardCardsOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	needle := effectstest.PushLibraryTopForTest(me, "Needle")
	first := effectstest.PushLibraryTopForTest(me, "First")
	effectstest.SeedOrderProbes(g, me.ID, effectstest.GraveyardOrderProbeOracles, "Graveyard Order Probe")
	effectstest.CastCatalogSpell(t, g, "Manifest Dread", "Sorcery", manifestDreadOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceChooseCards {
		t.Fatalf("pending %+v, want the manifest dread pick", g.PendingChoices)
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	pick := g.PendingChoices[0]
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{first}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	me = effectstest.SeatByID(g, me.ID)
	c, onField := effectstest.FindInZone(g.Battlefield, first)
	if !onField || !c.FaceDown {
		t.Fatalf("the First is not face down on the battlefield on the restored table (found %v)", onField)
	}
	if !me.Library.Contains(needle) {
		t.Fatal("the restored table moved the Needle before the answer")
	}
	mark := len(g.Events)
	effectstest.AnswerExileOrder(t, g)
	if !me.Graveyard.Contains(needle) {
		t.Fatal("the Needle is not in the graveyard after the restored answer")
	}
	announced := false
	for _, ev := range g.Events[mark:] {
		if ev.Kind == game.EventManifestDread && ev.CardID == first && ev.Target == needle {
			announced = true
		}
	}
	if !announced {
		t.Error("no EventManifestDread naming the manifested First and the graveyarded Needle")
	}
}

// --- Obstinate Baloth --------------------------------------------------------

// An opponent's Mind Rot over a filler and an Obstinate Baloth: the
// Baloth's discard becomes an entry, the entry asks for the order, the
// restored table has the filler in the graveyard and the Baloth in hand,
// and the answer puts it onto the battlefield, logs its discard there and
// gains the 4 life of its enters trigger.
func TestObstinateBalothsEntryIsARestorePointMidDiscard(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	filler := d.Filler(d.Victim)
	baloth := d.HandCard(d.Victim, "Obstinate Baloth", "Creature — Beast", "{2}{G}{G}", effectstest.ObstinateBalothOracle, 4, 4)
	effectstest.SeedOrderProbes(d.G, d.Victim.ID, effectstest.EntryOrderProbeOracles, "Entry Order Probe")
	life := d.Victim.Life
	effectstest.CastCatalogSpell(t, d.G, "Mind Rot", "Sorcery", effectstest.MindRotOracleID, []game.TargetRef{{Kind: game.TargetPlayer, ID: d.Victim.ID}})
	effectstest.PassPriorityAroundTable(t, d.G)
	effectstest.AnswerDiscard(t, d.G, d.Victim.ID, filler, baloth)

	g := effectstest.StrictRestoreWithFrame(t, d.G, game.ReplacementResumeDiscardEntry)
	victim := effectstest.SeatByID(g, d.Victim.ID)
	if !victim.Graveyard.Contains(filler) {
		t.Fatal("the filler is not in the graveyard on the restored table")
	}
	if !victim.Hand.Contains(baloth) {
		t.Fatal("the Baloth left the hand before the restored answer")
	}
	mark := len(g.Events)
	effectstest.AnswerExileOrder(t, g)
	if _, ok := effectstest.FindInZone(g.Battlefield, baloth); !ok {
		t.Fatal("the Baloth is not on the battlefield after the restored answer")
	}
	logged := false
	for _, ev := range g.Events[mark:] {
		if ev.Kind == game.EventDiscardCard && ev.CardID == baloth && ev.NewZone == game.ZoneBattlefield {
			logged = true
		}
	}
	if !logged {
		t.Error("the Baloth's discard was not logged on the battlefield")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if victim.Life != life+4 {
		t.Errorf("life %d -> %d, want +4 from the enters trigger", life, victim.Life)
	}
}
