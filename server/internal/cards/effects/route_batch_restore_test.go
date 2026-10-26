package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// route_batch_restore_test.go — #4253 slice 9: a route batch (an exile, a
// mill, a discard, manifest dread's two moves, a discard that became an
// entry) carries the rest of the batch as data, so a table waiting on a CR
// 616 prompt in the middle of it is a restore point. Each test pauses the
// batch on an ordering prompt over its SECOND card, takes a STRICT restore
// (the boot path), answers on the restored table, and checks the first
// card stayed where it went, the second moved after the answer, and the
// card's own "then" ran off the disk.

// graveyardOrderProbeOracles are two test cards, each "if a card named
// Needle would be put into a graveyard, it is put into a graveyard
// instead": two declared effects, so a Needle's move to a graveyard (a
// mill, a discard, a plain move) asks for their order (CR 616.1).
var graveyardOrderProbeOracles = []string{"test-4253-graveyard-order-a", "test-4253-graveyard-order-b"}

// entryOrderProbeOracles are the same over an Obstinate Baloth entering
// the battlefield.
var entryOrderProbeOracles = []string{"test-4253-entry-order-a", "test-4253-entry-order-b"}

func init() {
	for i, oracle := range graveyardOrderProbeOracles {
		Register(Spec{
			OracleID: oracle,
			Name:     "Graveyard Order Probe " + string(rune('A'+i)),
			Replacements: []game.ReplacementEffect{{
				Watches: []game.EventKind{game.EventZoneMove, game.EventDiscardCard},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
					if (ev.Kind != game.RepEventMove && ev.Kind != game.RepEventDiscard) || ev.NewZone != game.ZoneGraveyard {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.Name == "Needle"
				},
				Replace: func(*game.ReplacementEvent, *game.Game, *game.Card) error { return nil },
				Label:   "Graveyard Order Probe " + string(rune('A'+i)),
			}},
		})
	}
	for i, oracle := range entryOrderProbeOracles {
		Register(Spec{
			OracleID: oracle,
			Name:     "Entry Order Probe " + string(rune('A'+i)),
			Replacements: []game.ReplacementEffect{{
				Watches: []game.EventKind{game.EventZoneMove},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
					if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.Name == "Obstinate Baloth"
				},
				Replace: func(*game.ReplacementEvent, *game.Game, *game.Card) error { return nil },
				Label:   "Entry Order Probe " + string(rune('A'+i)),
			}},
		})
	}
}

func seedOrderProbes(g *game.Game, who uuid.UUID, oracles []string, name string) {
	for i, oracle := range oracles {
		seedReplacementPermanent(g, oracle, name+" "+string(rune('A'+i)), who)
	}
}

func seatByID(g *game.Game, id uuid.UUID) *game.Player {
	for _, p := range g.Seats {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// --- a discard ---------------------------------------------------------------

// An opponent's Mind Rot over a filler and a Needle: the Needle's discard
// asks for the order, the restored table has the filler in the graveyard
// and the Needle in hand, and the answer discards it.
func TestAMultiCardDiscardIsARestorePointAtTheSecondCardsOrder(t *testing.T) {
	d := newDBOTable(t)
	filler := d.filler(d.victim)
	needle := d.handCard(d.victim, "Needle", "Sorcery", "{1}", "", 0, 0)
	seedOrderProbes(d.g, d.victim.ID, graveyardOrderProbeOracles, "Graveyard Order Probe")
	castCatalogSpell(t, d.g, "Mind Rot", "Sorcery", mindRotOracleID, []game.TargetRef{{Kind: game.TargetPlayer, ID: d.victim.ID}})
	passPriorityAroundTable(t, d.g)
	answerDiscard(t, d.g, d.victim.ID, filler, needle)

	g := strictRestoreWithFrame(t, d.g, game.ReplacementResumeDiscard)
	victim := seatByID(g, d.victim.ID)
	if !victim.Graveyard.Contains(filler) {
		t.Fatal("the filler is not in the graveyard on the restored table")
	}
	if !victim.Hand.Contains(needle) {
		t.Fatal("the Needle left the hand before the restored answer")
	}
	answerExileOrder(t, g)
	if !victim.Graveyard.Contains(needle) {
		t.Fatal("the Needle was not discarded after the restored answer")
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts open after the discard finished", len(g.PendingChoices))
	}
}

// A random discard with a registered "then", told the cards discarded this
// way: the keyed door's tail runs off the disk after a restore mid-batch.
var testDiscardedGainLife = CardTail("test/4253-discarded-gain-life", func(ctx *Context, o game.TailOutcome) error {
	return ctx.Game.ChangePlayerLifeForEffect(uuid.Nil, ctx.Controller(), len(o.IDs))
})

func TestAKeyedDiscardsThenRunsAfterARestoreMidBatch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	d := dboTable{g: g, caster: me, victim: me}
	d.handCard(me, "Needle", "Sorcery", "{1}", "", 0, 0)
	d.handCard(me, "Needle", "Sorcery", "{1}", "", 0, 0)
	seedOrderProbes(g, me.ID, graveyardOrderProbeOracles, "Graveyard Order Probe")
	life := me.Life
	g.WithWriteLock(func() {
		if err := g.DiscardRandomKeyedForEffect(me.ID, 2, game.KeyedTail{Ref: testDiscardedGainLife, Player: me.ID}); err != nil {
			t.Fatalf("DiscardRandomKeyedForEffect: %v", err)
		}
	})
	g = strictRestoreWithFrame(t, g, game.ReplacementResumeDiscard)
	answerExileOrder(t, g)
	g = strictRestoreWithFrame(t, g, game.ReplacementResumeDiscard)
	answerExileOrder(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hama := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hama, the Bloodbender", TypeLine: "Legendary Creature — Human Warlock",
		OracleID: "ef80e174-78e8-412d-a6fe-89c7af746669", Owner: me.ID, Controller: me.ID,
	})
	third := libTop(opp, game.Card{Name: "Ponder", TypeLine: "Sorcery"})
	needle := pushLibraryTopForTest(opp, "Needle")
	first := pushLibraryTopForTest(opp, "First")
	seedOrderProbes(g, opp.ID, graveyardOrderProbeOracles, "Graveyard Order Probe")
	item := &game.StackItem{ID: uuid.New(), Controller: me.ID, SourceCardID: hama, Label: "Hama",
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}
	g.WithWriteLock(func() {
		if err := hamaEnters(g, item); err != nil {
			t.Fatalf("hamaEnters: %v", err)
		}
	})

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	opp = seatByID(g, opp.ID)
	if !opp.Graveyard.Contains(first) {
		t.Fatal("the first milled card is not in the graveyard on the restored table")
	}
	if !opp.Library.Contains(needle) || !opp.Library.Contains(third) {
		t.Fatal("the restored table milled the Needle or the third card before the answer")
	}
	answerExileOrder(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	next := pushLibraryTopForTest(me, "Next")
	needle := pushLibraryTopForTest(me, "Needle")
	forest := libTop(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	seedExileOrderProbes(g, me.ID)
	life := me.Life
	g.WithWriteLock(func() {
		if err := g.MillToZoneKeyedForEffect(me.ID, 0, game.ZoneExile, UntilNonland,
			game.KeyedTail{Ref: testDiscardedGainLife, Player: me.ID}); err != nil {
			t.Fatalf("MillToZoneKeyedForEffect: %v", err)
		}
	})

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	me = g.Seats[0]
	if _, ok := findInZone(g.Exile, forest); !ok {
		t.Fatal("the Forest is not in exile on the restored table")
	}
	if !me.Library.Contains(needle) {
		t.Fatal("the restored table exiled the Needle before the answer")
	}
	answerExileOrder(t, g)
	if _, ok := findInZone(g.Exile, needle); !ok {
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	first, needle := uuid.New(), uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: first, Name: "First", TypeLine: "Creature — Test", Owner: opp.ID, Controller: opp.ID})
	opp.Graveyard.PushTop(game.Card{InstanceID: needle, Name: "Needle", TypeLine: "Creature — Test", Owner: opp.ID, Controller: opp.ID})
	seedExileOrderProbes(g, opp.ID)
	life := me.Life
	castCatalogSpell(t, g, "Crypt Incursion", "Instant", b42CryptIncursionOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	me, opp = seatByID(g, me.ID), seatByID(g, opp.ID)
	if _, ok := findInZone(g.Exile, first); !ok {
		t.Fatal("the first creature card is not in exile on the restored table")
	}
	if !opp.Graveyard.Contains(needle) {
		t.Fatal("the restored table exiled the Needle before the answer")
	}
	answerExileOrder(t, g)
	if _, ok := findInZone(g.Exile, needle); !ok {
		t.Fatal("the Needle was not exiled after the restored answer")
	}
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	needle := pushLibraryTopForTest(me, "Needle")
	first := pushLibraryTopForTest(me, "First")
	seedOrderProbes(g, me.ID, graveyardOrderProbeOracles, "Graveyard Order Probe")
	castCatalogSpell(t, g, "Manifest Dread", "Sorcery", manifestDreadOracle, nil)
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceChooseCards {
		t.Fatalf("pending %+v, want the manifest dread pick", g.PendingChoices)
	}
	g = restoreRoundTrip(t, g, true)
	pick := g.PendingChoices[0]
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{first}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	me = seatByID(g, me.ID)
	c, onField := findInZone(g.Battlefield, first)
	if !onField || !c.FaceDown {
		t.Fatalf("the First is not face down on the battlefield on the restored table (found %v)", onField)
	}
	if !me.Library.Contains(needle) {
		t.Fatal("the restored table moved the Needle before the answer")
	}
	mark := len(g.Events)
	answerExileOrder(t, g)
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
	d := newDBOTable(t)
	filler := d.filler(d.victim)
	baloth := d.handCard(d.victim, "Obstinate Baloth", "Creature — Beast", "{2}{G}{G}", obstinateBalothOracle, 4, 4)
	seedOrderProbes(d.g, d.victim.ID, entryOrderProbeOracles, "Entry Order Probe")
	life := d.victim.Life
	castCatalogSpell(t, d.g, "Mind Rot", "Sorcery", mindRotOracleID, []game.TargetRef{{Kind: game.TargetPlayer, ID: d.victim.ID}})
	passPriorityAroundTable(t, d.g)
	answerDiscard(t, d.g, d.victim.ID, filler, baloth)

	g := strictRestoreWithFrame(t, d.g, game.ReplacementResumeDiscardEntry)
	victim := seatByID(g, d.victim.ID)
	if !victim.Graveyard.Contains(filler) {
		t.Fatal("the filler is not in the graveyard on the restored table")
	}
	if !victim.Hand.Contains(baloth) {
		t.Fatal("the Baloth left the hand before the restored answer")
	}
	mark := len(g.Events)
	answerExileOrder(t, g)
	if _, ok := findInZone(g.Battlefield, baloth); !ok {
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
	passPriorityAroundTable(t, g)
	if victim.Life != life+4 {
		t.Errorf("life %d -> %d, want +4 from the enters trigger", life, victim.Life)
	}
}
