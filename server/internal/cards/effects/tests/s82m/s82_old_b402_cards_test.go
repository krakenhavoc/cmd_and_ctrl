package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b402_cards_test.go — Dragon's Maze fuse cards, S82 OLD batch 402
// (#5564): thirteen split cards (twenty-six halves), each half and the fused
// casts that matter driven through the engine.

const (
	dgmTurnBurn         = "41841bbf-1c51-494b-b299-c997cce88e44"
	dgmFleshBlood       = "0741151a-fec0-4ed3-9295-0d0f56d24165"
	dgmProtectServe     = "2e0cd5a5-3aee-4b32-873a-e5a40bfba3e9"
	dgmFarAway          = "e4b1ef6a-6f05-472e-aaed-1c24dd6605c7"
	dgmProfitLoss       = "c41e8a99-e4a8-4323-a59d-265266e29fa9"
	dgmToilTrouble      = "97759579-b676-46e8-b566-fcad10751920"
	dgmAliveWell        = "57ad3c3a-6ac7-4a35-bc07-00f6ea0988c5"
	dgmArmedDangerous   = "a0593a5b-6f76-4eb3-a1b1-9794d2d36c92"
	dgmReadyWilling     = "44c7c6bb-2592-4d7c-8485-1815093e70be"
	dgmBreakingEntering = "043370fd-9cfe-47ce-8019-e915cee1ae95"
	dgmBeckCall         = "5b8472c8-a7e7-46f2-aecf-e954082a5ef5"
	dgmGiveTake         = "67a82d5e-a9dc-4e8c-9d23-c6d6aa1e57d1"
	dgmDownDirty        = "eba21e3b-e2b2-4e0b-82e3-f0849943fd89"
)

func TestS82OldB402EveryHalfIsRegisteredFull(t *testing.T) {
	for oracle, names := range map[string][2]string{
		dgmTurnBurn: {"Turn", "Burn"}, dgmFleshBlood: {"Flesh", "Blood"}, dgmProtectServe: {"Protect", "Serve"},
		dgmFarAway: {"Far", "Away"}, dgmProfitLoss: {"Profit", "Loss"}, dgmToilTrouble: {"Toil", "Trouble"},
		dgmAliveWell: {"Alive", "Well"}, dgmArmedDangerous: {"Armed", "Dangerous"}, dgmReadyWilling: {"Ready", "Willing"},
		dgmBreakingEntering: {"Breaking", "Entering"}, dgmBeckCall: {"Beck", "Call"}, dgmGiveTake: {"Give", "Take"},
		dgmDownDirty: {"Down", "Dirty"},
	} {
		for i, id := range []string{oracle, oracle + "#1"} {
			spec, ok := effects.Lookup(id)
			if !ok {
				t.Errorf("%s is not registered", names[i])
				continue
			}
			if spec.Name != names[i] || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
				t.Errorf("%s: name %q, completeness %s, caveats %v", names[i], spec.Name, spec.Completeness, spec.Caveats)
			}
		}
	}
}

// dgmSplit is a split card in `owner`'s hand: the two halves' names, type lines
// and mana costs. The oracle text only has to carry the fuse reminder.
func dgmSplit(g *game.Game, owner *game.Player, oracle, typeLine string, left, right [2]string) uuid.UUID {
	const fuse = "\nFuse (You may cast one or both halves of this card from your hand.)"
	c := game.Card{
		InstanceID: uuid.New(), OracleID: oracle, Layout: game.LayoutSplit, Owner: owner.ID, Controller: owner.ID,
		Faces: []game.Face{
			{Name: left[0], TypeLine: typeLine, ManaCost: left[1], OracleText: "." + fuse},
			{Name: right[0], TypeLine: typeLine, ManaCost: right[1], OracleText: "." + fuse},
		},
	}
	c.SettleImported()
	owner.Hand.PushTop(c)
	return c.InstanceID
}

func dgmCast(t *testing.T, g *game.Game, me *game.Player, card uuid.UUID, params game.CastSpellParams) {
	t.Helper()
	if err := g.CastSpell(me.ID, card, params); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func dgmCard(id uuid.UUID, slot int) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id, Slot: slot}
}

func dgmPlayer(id uuid.UUID, slot int) game.TargetRef {
	return game.TargetRef{Kind: game.TargetPlayer, ID: id, Slot: slot}
}

func dgmHas(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

func TestS82OldB402TurnMakesARedWeirdAndBurnHitsAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ogre := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Ogre", TypeLine: "Creature — Ogre", Power: 4, Toughness: 4, Keywords: []string{"flying"}})
	card := dgmSplit(g, me, dgmTurnBurn, "Instant", [2]string{"Turn", "{2}{U}"}, [2]string{"Burn", "{1}{R}"})
	dgmCast(t, g, me, card, game.CastSpellParams{Targets: []game.TargetRef{dgmCard(ogre, 0)}})
	c := effectstest.LayeredCard(t, g, ogre)
	if p, th := c.CurrentPower(), c.CurrentToughness(); p != 0 || th != 1 {
		t.Errorf("Turn left %d/%d, want 0/1", p, th)
	}
	if !dgmHas(c.Effective().Subtypes, "Weird") || dgmHas(c.Effective().Subtypes, "Ogre") {
		t.Errorf("subtypes %v, want only Weird", c.Effective().Subtypes)
	}
	if effectstest.HasEffectiveKeyword(t, g, ogre, "flying") {
		t.Error("it kept its abilities")
	}
	if !dgmHas(effectstest.EffectiveColors(t, g, ogre), "R") {
		t.Error("it is not red")
	}
	// Burn, from the same card's other half, in a second game.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	effectstest.AdvanceToMain(t, g2)
	card2 := dgmSplit(g2, me2, dgmTurnBurn, "Instant", [2]string{"Turn", "{2}{U}"}, [2]string{"Burn", "{1}{R}"})
	before := opp2.Life
	dgmCast(t, g2, me2, card2, game.CastSpellParams{Face: 1, Targets: []game.TargetRef{dgmPlayer(opp2.ID, 0)}})
	if opp2.Life != before-2 {
		t.Errorf("Burn: %d -> %d, want -2", before, opp2.Life)
	}
}

func TestS82OldB402FleshExilesAndCountersByPowerBloodBites(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	dead := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Giant", TypeLine: "Creature — Giant", Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	card := dgmSplit(g, me, dgmFleshBlood, "Sorcery", [2]string{"Flesh", "{3}{B}{G}"}, [2]string{"Blood", "{R}{G}"})
	dgmCast(t, g, me, card, game.CastSpellParams{Targets: []game.TargetRef{dgmCard(dead, 0), dgmCard(bear, 1)}})
	if opp.Graveyard.Contains(dead) {
		t.Error("the creature card was not exiled")
	}
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 5 {
		t.Errorf("%d counters, want 5", got)
	}
	// Blood: the 7/7 bear (2 + 5) deals 7 damage to the opponent.
	card2 := dgmSplit(g, me, dgmFleshBlood, "Sorcery", [2]string{"Flesh", "{3}{B}{G}"}, [2]string{"Blood", "{R}{G}"})
	before := opp.Life
	dgmCast(t, g, me, card2, game.CastSpellParams{Face: 1, Targets: []game.TargetRef{dgmCard(bear, 0), dgmPlayer(opp.ID, 1)}})
	if opp.Life != before-7 {
		t.Errorf("Blood: %d -> %d, want -7", before, opp.Life)
	}
}

func TestS82OldB402ProtectAndServe(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	card := dgmSplit(g, me, dgmProtectServe, "Instant", [2]string{"Protect", "{2}{W}"}, [2]string{"Serve", "{1}{U}"})
	dgmCast(t, g, me, card, game.CastSpellParams{Fuse: true, Targets: []game.TargetRef{dgmCard(bear, 0), dgmCard(bear, 1)}})
	c := effectstest.LayeredCard(t, g, bear)
	if p, th := c.Effective().Power, c.Effective().Toughness; p != -2 || th != 6 {
		t.Errorf("fused Protect // Serve left %d/%d, want -2/6", p, th)
	}
}

func TestS82OldB402FarBouncesAndAwayMakesThemSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	dgmCast(t, g, me, dgmSplit(g, me, dgmFarAway, "Instant", [2]string{"Far", "{1}{U}"}, [2]string{"Away", "{2}{B}"}),
		game.CastSpellParams{Targets: []game.TargetRef{dgmCard(ogre, 0)}})
	if g.Battlefield.Contains(ogre) || !opp.Hand.Contains(ogre) {
		t.Error("Far did not return the creature to its owner's hand")
	}
	lone := effectstest.B12Creature(g, opp.ID, "Lone Bear", "Creature — Bear", 2, 2)
	dgmCast(t, g, me, dgmSplit(g, me, dgmFarAway, "Instant", [2]string{"Far", "{1}{U}"}, [2]string{"Away", "{2}{B}"}),
		game.CastSpellParams{Face: 1, Targets: []game.TargetRef{dgmPlayer(opp.ID, 0)}})
	effectstest.AnswerSacrifice(t, g, opp.ID, lone)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(lone) {
		t.Error("Away did not make the player sacrifice their creature")
	}
}

func TestS82OldB402ProfitAndLossScopeTheirCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	mouse := effectstest.B12Creature(g, opp.ID, "Mouse", "Creature — Mouse", 1, 1)
	dgmCast(t, g, me, dgmSplit(g, me, dgmProfitLoss, "Instant", [2]string{"Profit", "{1}{W}"}, [2]string{"Loss", "{2}{B}"}),
		game.CastSpellParams{Fuse: true})
	if c := effectstest.LayeredCard(t, g, mine); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("my creature %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if c := effectstest.LayeredCard(t, g, theirs); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("their creature %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
	if g.Battlefield.Contains(mouse) {
		t.Error("their 1/1 survived Loss")
	}
}

func TestS82OldB402FusedToilThenTroubleCountsTheDrawnCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	hand, life := opp.Hand.Size(), opp.Life
	dgmCast(t, g, me, dgmSplit(g, me, dgmToilTrouble, "Sorcery", [2]string{"Toil", "{2}{B}"}, [2]string{"Trouble", "{2}{R}"}),
		game.CastSpellParams{Fuse: true, Targets: []game.TargetRef{dgmPlayer(opp.ID, 0), dgmPlayer(opp.ID, 1)}})
	if opp.Hand.Size() != hand+2 {
		t.Errorf("hand %d -> %d, want +2", hand, opp.Hand.Size())
	}
	// Toil: -2 life; Trouble: damage equal to the new hand size.
	if want := life - 2 - (hand + 2); opp.Life != want {
		t.Errorf("life %d -> %d, want %d", life, opp.Life, want)
	}
}

func TestS82OldB402FusedAliveThenWellCountsTheCentaur(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	dgmCast(t, g, me, dgmSplit(g, me, dgmAliveWell, "Sorcery", [2]string{"Alive", "{3}{G}"}, [2]string{"Well", "{W}"}),
		game.CastSpellParams{Fuse: true})
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Centaur"); n != 1 {
		t.Errorf("%d Centaurs, want 1", n)
	}
	if me.Life != life+4 {
		t.Errorf("life %d -> %d, want +4 (two creatures)", life, me.Life)
	}
}

func TestS82OldB402ArmedAndDangerous(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	dgmCast(t, g, me, dgmSplit(g, me, dgmArmedDangerous, "Sorcery", [2]string{"Armed", "{1}{R}"}, [2]string{"Dangerous", "{3}{G}"}),
		game.CastSpellParams{Targets: []game.TargetRef{dgmCard(bear, 0)}})
	if c := effectstest.LayeredCard(t, g, bear); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("Armed left %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("no double strike")
	}
}

func TestS82OldB402DangerousLuresTheTargetForTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Seats1684(g)
	bear := effectstest.ReqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	other := effectstest.ReqCreature(g, me.ID, "Other Bear", "", 2, 2, "haste")
	a := effectstest.ReqCreature(g, opp.ID, "Wall A", "", 0, 4)
	b := effectstest.ReqCreature(g, opp.ID, "Wall B", "", 0, 4)
	effectstest.AdvanceToMain(t, g)
	dgmCast(t, g, me, dgmSplit(g, me, dgmArmedDangerous, "Sorcery", [2]string{"Armed", "{1}{R}"}, [2]string{"Dangerous", "{3}{G}"}),
		game.CastSpellParams{Face: 1, Targets: []game.TargetRef{dgmCard(bear, 0)}})
	effectstest.ReqToBlockers(t, g, bear, other)
	effectstest.ReqRefusal(t, effectstest.ReqBlock(t, g, other, a))
	effectstest.ReqRefusal(t, g.PassPriority())
	if err := effectstest.ReqBlock(t, g, bear, a, b); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with both blocking: %v", err)
	}
}

func TestS82OldB402ReadyUntapsAndGrantsIndestructibleWillingGrantsDeathtouchAndLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bear).Tapped = true })
	dgmCast(t, g, me, dgmSplit(g, me, dgmReadyWilling, "Instant", [2]string{"Ready", "{1}{G}{W}"}, [2]string{"Willing", "{1}{W}{B}"}),
		game.CastSpellParams{Fuse: true})
	if effectstest.TappedOnBattlefield(t, g, bear) {
		t.Error("Ready did not untap the creature")
	}
	for _, kw := range []string{"indestructible", "deathtouch", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("no %s", kw)
		}
	}
}

func TestS82OldB402FusedBreakingThenEnteringMayTakeAMilledCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	opp.Library.Cards = nil
	var top uuid.UUID
	for i := 0; i < 8; i++ {
		id := uuid.New()
		opp.Library.PushTop(game.Card{InstanceID: id, Name: "Milled Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
		top = id
	}
	card := dgmSplit(g, me, dgmBreakingEntering, "Sorcery", [2]string{"Breaking", "{U}{B}"}, [2]string{"Entering", "{4}{B}{R}"})
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Fuse: true, Targets: []game.TargetRef{dgmPlayer(opp.ID, 0)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != 0 {
		t.Fatalf("library %d, want 0 after milling eight", opp.Library.Size())
	}
	effectstest.AnswerChooseCards(t, g, me.ID, top)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.FindBattlefieldCardForTest(g, top)
	if c == nil || c.Controller != me.ID {
		t.Fatal("the milled creature did not enter under my control")
	}
	if !effectstest.HasEffectiveKeyword(t, g, top, "haste") {
		t.Error("no haste")
	}
}

func TestS82OldB402FusedBeckThenCallDrawsForEachBird(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	card := dgmSplit(g, me, dgmBeckCall, "Sorcery", [2]string{"Beck", "{G}{U}"}, [2]string{"Call", "{4}{W}{U}"})
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Fuse: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	for i := 0; i < 64 && !(effectstest.StackFullyEmpty(g) && len(g.PendingChoices) == 0); i++ {
		if c := effectstest.LatestTriggerPrompt(g, me.ID); c != nil {
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
			continue
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Bird"); n != 4 {
		t.Errorf("%d Birds, want 4", n)
	}
	// Four draws came in (the split card was cast after `hand` was read).
	if got := me.Hand.Size(); got != hand+4 {
		t.Errorf("hand %d -> %d, want %d", hand, got, hand+4)
	}
}

func TestS82OldB402FusedGiveThenTakeDrawsThreeAndRemovesTheCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	dgmCast(t, g, me, dgmSplit(g, me, dgmGiveTake, "Sorcery", [2]string{"Give", "{2}{G}"}, [2]string{"Take", "{2}{U}"}),
		game.CastSpellParams{Fuse: true, Targets: []game.TargetRef{dgmCard(bear, 0), dgmCard(bear, 1)}})
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters left, want 0", got)
	}
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d -> %d, want %d", hand, got, hand+3)
	}
}

func TestS82OldB402FusedDownThenDirtyReturnsADiscardedCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	me.Hand.Cards = nil
	keep := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: keep, Name: "Old Friend", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	card := dgmSplit(g, me, dgmDownDirty, "Sorcery", [2]string{"Down", "{3}{B}"}, [2]string{"Dirty", "{2}{G}"})
	dgmCast(t, g, me, card, game.CastSpellParams{Fuse: true, Targets: []game.TargetRef{dgmPlayer(me.ID, 0), dgmCard(keep, 1)}})
	if !me.Hand.Contains(keep) {
		t.Error("Dirty did not return the card to hand")
	}
}
