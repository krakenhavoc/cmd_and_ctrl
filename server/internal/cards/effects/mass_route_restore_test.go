package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mass_route_restore_test.go — #4253 slice 10a: the mass verbs, the single
// sacrifice and bounce, the tuck and destroy callers, the exile whose "then"
// grants a cast permission or records an "until" return, an exchange's exit
// and a mill's amount window are restore points. Each test pauses on a CR
// 616 ordering prompt, takes a STRICT restore (the boot path), answers on
// the restored table, and checks the rest of the card ran off the disk.

// libraryHandOrderProbeOracles are two test cards, each "if a card named
// Needle would be put into a library or a hand, it is put there instead":
// two declared effects, so a Needle's tuck or bounce asks for their order
// (CR 616.1).
var libraryHandOrderProbeOracles = []string{"test-4253-10a-library-hand-order-a", "test-4253-10a-library-hand-order-b"}

// millPlusOneProbeOracles are two test cards, each "if a player would mill
// one or more cards, they mill that many plus one instead": two declared
// mill-amount replacements, so a mill's amount window asks the milling
// player for their order (CR 616.1).
var millPlusOneProbeOracles = []string{"test-4253-10a-mill-plus-one-a", "test-4253-10a-mill-plus-one-b"}

func init() {
	for i, oracle := range millPlusOneProbeOracles {
		Register(Spec{
			OracleID: oracle,
			Name:     "Mill Plus One Probe " + string(rune('A'+i)),
			Replacements: []game.ReplacementEffect{{
				Watches: []game.EventKind{game.EventMill},
				AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
					return ev.Kind == game.RepEventMill && ev.MillCount > 0
				},
				Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
					ev.MillCount++
					return nil
				},
				Label: "Mill Plus One Probe " + string(rune('A'+i)),
			}},
		})
	}
	for i, oracle := range libraryHandOrderProbeOracles {
		Register(Spec{
			OracleID: oracle,
			Name:     "Library Hand Order Probe " + string(rune('A'+i)),
			Replacements: []game.ReplacementEffect{{
				Watches: []game.EventKind{game.EventZoneMove},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
					if ev.Kind != game.RepEventMove || (ev.NewZone != game.ZoneLibrary && ev.NewZone != game.ZoneHand) {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.Name == "Needle"
				},
				Replace: func(*game.ReplacementEvent, *game.Game, *game.Card) error { return nil },
				Label:   "Library Hand Order Probe " + string(rune('A'+i)),
			}},
		})
	}
}

const (
	fumigateOracleID = "b17ea905-0696-4e58-b564-557e87236e27"
	chaosWarpOracle  = "07a0cba9-8768-4fd9-a3d5-b0f83b4bf8e8"
	terminusOracleID = "3dd196b6-a85a-4e3e-bb57-ec34241f8117"
	sunfallOracleID  = "fb3f5097-0da0-458c-8508-60823567e2da"
)

// tableSeats is the active seat and the one after it.
func tableSeats(g *game.Game) (*game.Player, *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// --- the mass verbs ---------------------------------------------------------

// Fumigate over a Bear and a Needle (DestroyAllMatching.ThenTail, the shared
// mass/gain-life-for-each): the Needle's destruction asks, the restored
// table has the Bear in the graveyard, and the answer destroys the Needle
// and gains 2.
func TestFumigateIsARestorePointAtTheSecondCreaturesOrder(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := tableSeats(g)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	needle := b12Creature(g, opp.ID, "Needle", "Creature — Test", 1, 1)
	seedOrderProbes(g, opp.ID, graveyardOrderProbeOracles, "Graveyard Order Probe")
	life := me.Life
	castCatalogSpell(t, g, "Fumigate", "Sorcery", fumigateOracleID, nil)
	passPriorityAroundTable(t, g)

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	me, opp = seatByID(g, me.ID), seatByID(g, opp.ID)
	if !opp.Graveyard.Contains(bear) {
		t.Fatal("the Bear is not in the graveyard on the restored table")
	}
	if !g.Battlefield.Contains(needle) {
		t.Fatal("the restored table destroyed the Needle before the answer")
	}
	answerExileOrder(t, g)
	if !opp.Graveyard.Contains(needle) {
		t.Fatal("the Needle was not destroyed after the restored answer")
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2 for the two creatures destroyed", life, me.Life)
	}
}

// Sunfall over a Bear and a Needle (ExileAllMatching.ThenTail): the
// restored answer exiles the Needle and incubates 2.
func TestSunfallIsARestorePointMidSweep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := tableSeats(g)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	needle := b12Creature(g, opp.ID, "Needle", "Creature — Test", 1, 1)
	seedExileOrderProbes(g, opp.ID)
	castCatalogSpell(t, g, "Sunfall", "Sorcery", sunfallOracleID, nil)
	passPriorityAroundTable(t, g)

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	if _, ok := findInZone(g.Exile, bear); !ok {
		t.Fatal("the Bear is not in exile on the restored table")
	}
	answerExileOrder(t, g)
	if _, ok := findInZone(g.Exile, needle); !ok {
		t.Fatal("the Needle was not exiled after the restored answer")
	}
	passPriorityAroundTable(t, g)
	incubators := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && IsToken(c) {
			incubators++
		}
	}
	if incubators != 1 {
		t.Errorf("%d tokens for the caster after the sweep, want the one Incubator", incubators)
	}
}

// ReturnAllToHand with a keyed "then" over a filler and a Needle: the
// Needle's bounce asks, and the restored answer runs the tail with both.
func TestReturnAllToHandsKeyedThenRunsAfterARestore(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := tableSeats(g)
	filler := b12Creature(g, me.ID, "Filler", "Creature — Test", 1, 1)
	needle := b12Creature(g, me.ID, "Needle", "Creature — Test", 1, 1)
	seedOrderProbes(g, me.ID, libraryHandOrderProbeOracles, "Library Hand Order Probe")
	life := me.Life
	g.WithWriteLock(func() {
		ctx := keyedContext(g, me.ID, uuid.Nil)
		if err := (ReturnAllToHand{Cards: []uuid.UUID{filler, needle},
			ThenTail: game.KeyedTail{Ref: testDiscardedGainLife, Player: me.ID}}).Apply(ctx); err != nil {
			t.Fatalf("ReturnAllToHand: %v", err)
		}
	})
	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	answerExileOrder(t, g)
	me = seatByID(g, me.ID)
	if !me.Hand.Contains(filler) || !me.Hand.Contains(needle) {
		t.Fatal("both cards are not in hand after the restored answer")
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2 from the keyed then", life, me.Life)
	}
}

// --- the single-permanent verbs ---------------------------------------------

// A BounceToHand and a SacrificePermanent with a keyed "then", each paused
// on the Needle's ordering prompt: the restored answer moves it and tells
// the tail it went (one ID).
func TestTheSingleVerbsKeyedThenRunsAfterARestore(t *testing.T) {
	for _, tc := range []struct {
		name   string
		probes []string
		apply  func(ctx *Context, needle uuid.UUID, tail game.KeyedTail) error
		landed func(g *game.Game, p *game.Player, needle uuid.UUID) bool
	}{
		{"bounce", libraryHandOrderProbeOracles,
			func(ctx *Context, needle uuid.UUID, tail game.KeyedTail) error {
				return BounceToHand{Target: needle, ThenTail: tail}.Apply(ctx)
			},
			func(_ *game.Game, p *game.Player, needle uuid.UUID) bool { return p.Hand.Contains(needle) }},
		{"sacrifice", graveyardOrderProbeOracles,
			func(ctx *Context, needle uuid.UUID, tail game.KeyedTail) error {
				return SacrificePermanent{Target: needle, ThenTail: tail}.Apply(ctx)
			},
			func(_ *game.Game, p *game.Player, needle uuid.UUID) bool { return p.Graveyard.Contains(needle) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, _ := tableSeats(g)
			needle := b12Creature(g, me.ID, "Needle", "Creature — Test", 1, 1)
			seedOrderProbes(g, me.ID, tc.probes, "Order Probe")
			life := me.Life
			g.WithWriteLock(func() {
				if err := tc.apply(keyedContext(g, me.ID, uuid.Nil), needle, game.KeyedTail{Ref: testDiscardedGainLife, Player: me.ID}); err != nil {
					t.Fatalf("apply: %v", err)
				}
			})
			g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
			if !g.Battlefield.Contains(needle) {
				t.Fatal("the Needle moved before the restored answer")
			}
			answerExileOrder(t, g)
			me = seatByID(g, me.ID)
			if !tc.landed(g, me, needle) {
				t.Fatal("the Needle did not go after the restored answer")
			}
			if me.Life != life+1 {
				t.Errorf("life %d -> %d, want +1: the tail is told the one card that went", life, me.Life)
			}
		})
	}
}

// --- the tuck callers ---------------------------------------------------------

// Chaos Warp on a Needle whose owner's library is otherwise empty: the
// tuck asks, the restored table still has the Needle on the battlefield,
// and the answer tucks it, shuffles, reveals it and puts it back onto the
// battlefield (then/chaos-warp-shuffle-reveal, run off the disk).
func TestChaosWarpIsARestorePointAtTheTuck(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := tableSeats(g)
	needle := b12Creature(g, opp.ID, "Needle", "Creature — Test", 1, 1)
	g.WithWriteLock(func() { opp.Library.Cards = nil })
	seedOrderProbes(g, opp.ID, libraryHandOrderProbeOracles, "Library Hand Order Probe")
	castCatalogSpell(t, g, "Chaos Warp", "Instant", chaosWarpOracle, []game.TargetRef{{Kind: game.TargetCard, ID: needle}})
	passPriorityAroundTable(t, g)

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	opp = seatByID(g, opp.ID)
	if !g.Battlefield.Contains(needle) || opp.Library.Size() != 0 {
		t.Fatal("the restored table tucked the Needle before the answer")
	}
	answerExileOrder(t, g)
	back := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Needle" && c.Owner == opp.ID {
			back++
		}
	}
	if back != 1 || opp.Library.Size() != 0 {
		t.Fatalf("after the restored answer %d Needles are on the battlefield and %d cards in the library, want the revealed Needle back and an empty library",
			back, opp.Library.Size())
	}
}

// Terminus over two of an opponent's creatures, the second a Needle: the
// Needle's tuck asks, and the restored answer tucks it and opens the
// owner's bottom-order prompt over both (tuck/each-owner-orders).
func TestTerminusIsARestorePointAtTheSecondTuck(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := tableSeats(g)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	needle := b12Creature(g, opp.ID, "Needle", "Creature — Test", 1, 1)
	seedOrderProbes(g, opp.ID, libraryHandOrderProbeOracles, "Library Hand Order Probe")
	castCatalogSpell(t, g, "Terminus", "Sorcery", terminusOracleID, nil)
	passPriorityAroundTable(t, g)

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	opp = seatByID(g, opp.ID)
	if !opp.Library.Contains(bear) || !g.Battlefield.Contains(needle) {
		t.Fatal("the restored table is not between the two tucks")
	}
	answerExileOrder(t, g)
	var order *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c.Kind == game.PendingChoicePutInLibrary && c.Chooser == opp.ID {
			order = c
		}
	}
	if order == nil || len(order.ScryCards) != 2 {
		t.Fatalf("the owner's bottom-order prompt over both did not open: %+v", g.PendingChoices)
	}
}

// --- an "until" exile and a cast permission ------------------------------------

// Oblivion Ring over a Needle (ExileUntil's keyed tail, exile-until/record-
// return): the exile asks, and after the restored answer the Needle is in
// exile with its return recorded, so destroying the Ring brings it back.
func TestOblivionRingIsARestorePointAtTheExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := tableSeats(g)
	needle := b12Creature(g, opp.ID, "Needle", "Creature — Test", 1, 1)
	seedExileOrderProbes(g, opp.ID)
	ring := castCatalogSpell(t, g, "Oblivion Ring", "Enchantment", g2OblivionRingOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, needle)
	passPriorityAroundTable(t, g)

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	answerExileOrder(t, g)
	if _, ok := findInZone(g.Exile, needle); !ok {
		t.Fatal("the Needle is not in exile after the restored answer")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(ring) })
	g.RunStateChecksForTest()
	back := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Needle" && c.Controller == opp.ID {
			back++
		}
	}
	if back != 1 {
		t.Fatalf("after the Ring left %d Needles are back, want 1: the return was not recorded off the disk", back)
	}
}

// Airbend on a Needle (exile/grant-permission-over-landed, the permission as
// data): the restored answer exiles it and its owner may cast it.
func TestAirbendsGrantSurvivesARestoreAtTheExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := tableSeats(g)
	needle := b12Creature(g, opp.ID, "Needle", "Creature — Test", 1, 1)
	seedExileOrderProbes(g, opp.ID)
	before := len(opp.CastPermissions)
	g.WithWriteLock(func() {
		if err := (Airbend{Target: needle}).Apply(keyedContext(g, me.ID, uuid.Nil)); err != nil {
			t.Fatalf("Airbend: %v", err)
		}
	})
	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	answerExileOrder(t, g)
	opp = seatByID(g, opp.ID)
	if _, ok := findInZone(g.Exile, needle); !ok {
		t.Fatal("the Needle is not in exile after the restored answer")
	}
	granted := false
	for _, p := range opp.CastPermissions[before:] {
		for _, c := range p.Cards {
			if c.ID == needle {
				granted = true
			}
		}
	}
	if !granted {
		t.Fatalf("the owner holds no permission over the exiled Needle: %+v", opp.CastPermissions)
	}
}

// --- an exchange --------------------------------------------------------------

// Goblin Welder's exchange of a Needle artifact for one in the graveyard:
// the entry settles, the Needle's sacrifice asks, the restored table has
// both cards where they were, and the answer sacrifices the Needle and
// lands the returning artifact (replacement-resume/move-exchange).
func TestAnExchangeIsARestorePointAtItsExit(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := tableSeats(g)
	welder := pushCatalogPermanent(g, me.ID, "Goblin Welder", "Creature — Goblin Artificer", goblinWelderOracle, false)
	needle := pushCatalogPermanent(g, me.ID, "Needle", "Artifact", "", false)
	back := welderGraveyardArtifact(me, "Returning Relic", "")
	seedOrderProbes(g, me.ID, graveyardOrderProbeOracles, "Graveyard Order Probe")
	g.WithWriteLock(func() {
		if err := g.ExchangeArtifactWithGraveyardForEffect(welder, needle, back); err != nil {
			t.Fatalf("exchange: %v", err)
		}
	})

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveExchange)
	me = seatByID(g, me.ID)
	if !g.Battlefield.Contains(needle) || !me.Graveyard.Contains(back) {
		t.Fatal("the restored table moved a card of the exchange before the answer")
	}
	answerExileOrder(t, g)
	if !me.Graveyard.Contains(needle) {
		t.Fatal("the Needle was not sacrificed after the restored answer")
	}
	if c, ok := battlefieldCard(g, back); !ok || c.Controller != me.ID {
		t.Fatal("the returning artifact did not land with the exit: the settled entry was not restored")
	}
}

// --- a mill-amount window ----------------------------------------------------

// Two mill-amount replacements over an opponent's keyed mill of one: the
// amount window asks that opponent for the order (CR 616.1), the restored
// table has milled nothing, and the answer mills three and runs the keyed
// tail with them (replacement-resume/mill).
func TestAMillAmountWindowIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := tableSeats(g)
	seedOrderProbes(g, me.ID, millPlusOneProbeOracles, "Mill Plus One Probe")
	life, grave := me.Life, opp.Graveyard.Size()
	g.WithWriteLock(func() {
		if err := g.MillToZoneKeyedForEffect(opp.ID, 1, game.ZoneGraveyard, game.MillUntil{},
			game.KeyedTail{Ref: testDiscardedGainLife, Player: me.ID}); err != nil {
			t.Fatalf("MillToZoneKeyedForEffect: %v", err)
		}
	})
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Chooser != opp.ID {
		t.Fatalf("pending %+v, want the milling player's ordering prompt", g.PendingChoices)
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMill)
	me, opp = seatByID(g, me.ID), seatByID(g, opp.ID)
	if opp.Graveyard.Size() != grave {
		t.Fatal("the restored table milled before the answer")
	}
	answerExileOrder(t, g)
	if got := opp.Graveyard.Size() - grave; got != 3 {
		t.Errorf("milled %d after the restored answer, want 3 (one plus one plus one)", got)
	}
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3 from the keyed then", life, me.Life)
	}
}
