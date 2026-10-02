package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// land_play_restriction_cards_test.go — ADR 0109 §4 (#1895), the card
// half: each printed shape refused at the land play (CR 101.2) with its
// printed clause, the card left in hand, and the same answer from the
// gate the enumerator and the view ask. The three-way agreement is
// asserted in internal/legal and internal/protocol against the same cards.

const (
	lpAggressiveMiningOracle   = "ff321c13-03ae-4cb9-b371-1242290f8433"
	lpCityInABottleOracle      = "a83f25e3-4d84-4c9b-ab12-19b8d326e459"
	lpExperimentalFrenzyOracle = "3715e8b8-31df-499b-8a91-1bcd1199d6eb"
	lpMoonholdOracle           = "9b515bb8-6e37-4b2d-8d56-bb517c3b267c"
	lpPardicMinerOracle        = "e1788b54-4cd9-459f-b2c9-06773e68e9e3"
	lpRockJockeyOracle         = "09fff6a3-6a54-4101-a587-55d7113b9639"
	lpSolfataraOracle          = "6053a192-fcf9-4b06-9f46-e84eaadaa882"
	lpTerritorialDisputeOracle = "a1785817-f17b-471b-a63b-866e7972df1f"
	lpTurfWoundOracle          = "8dc48180-1482-40c4-8c5f-1724018e9c5b"
	lpWardOfBonesOracle        = "c3ea1497-63ac-46bb-95e9-d98b93a880b3"
)

// lpHandLand puts a land card in p's hand.
func lpHandLand(p *game.Player, name string) uuid.UUID {
	id := uuid.New()
	typeLine := "Basic Land — Forest"
	if name != "Forest" {
		typeLine = "Land"
	}
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// lpPlay plays a land out of hand through the engine's own path.
func lpPlay(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID) error {
	t.Helper()
	advanceToMain(t, g)
	return g.CastSpell(p.ID, id, game.CastSpellParams{})
}

// lpGate asks the land-play gate for a land card in `zone`.
func lpGate(g *game.Game, p *game.Player, name string, zone game.ZoneKind) error {
	var err error
	g.WithWriteLock(func() {
		err = g.LandPlayGateLocked(p.ID, game.Card{Name: name, TypeLine: "Land"}, zone)
	})
	return err
}

// lpAssertRefused pins the error SHAPE: the sentinel, the printed clause,
// and the card still in hand with no land drop spent.
func lpAssertRefused(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("the land play was allowed, want a refusal mentioning %q", want)
	}
	if !errors.Is(err, game.ErrCantPlayLand) {
		t.Fatalf("refusal is %v, want it to wrap game.ErrCantPlayLand", err)
	}
	var cant *game.CantPlayLandError
	if !errors.As(err, &cant) || !strings.Contains(cant.Reason, want) {
		t.Fatalf("refusal = %v, want a *game.CantPlayLandError mentioning %q", err, want)
	}
	if !p.Hand.Contains(id) {
		t.Error("a refused land play did not leave the card in hand")
	}
	if g.LandsPlayedThisTurnFor(p.ID) != 0 {
		t.Error("a refused land play spent a land drop")
	}
}

// Territorial Dispute: "Players can't play lands" binds everyone, and an
// extra land drop does not lift it (CR 101.2's own example).
func TestTerritorialDisputeStopsEveryLandPlay(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Territorial Dispute", "Enchantment", lpTerritorialDisputeOracle, false)
	g.WithWriteLock(func() { g.GrantAdditionalLandPlayForEffect(me.ID, 2) })
	land := lpHandLand(me, "Forest")
	err := lpPlay(t, g, me, land)
	lpAssertRefused(t, g, me, land, err, "Players can't play lands.")
	if !strings.Contains(err.Error(), "Territorial Dispute") {
		t.Errorf("the refusal does not name the card: %v", err)
	}
	if lpGate(g, g.Seats[1], "Forest", game.ZoneHand) == nil {
		t.Error("an opponent may play a land under Territorial Dispute")
	}
}

// Aggressive Mining: "You can't play lands" is the CONTROLLER's, so one an
// opponent controls restricts them and not the player whose turn it is. Its
// ability draws two, once each turn.
func TestAggressiveMiningBindsItsControllerAndDrawsTwice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	apaPush(g, me.ID, opp.ID, game.Card{Name: "Aggressive Mining", TypeLine: "Enchantment", OracleID: lpAggressiveMiningOracle})
	if err := lpPlay(t, g, me, lpHandLand(me, "Forest")); err != nil {
		t.Fatalf("an Aggressive Mining the opponent controls stopped its owner's land play: %v", err)
	}
	if lpGate(g, opp, "Forest", game.ZoneHand) == nil {
		t.Error("the controller of Aggressive Mining may play a land")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	mining := pushCatalogPermanent(g2, me2.ID, "Aggressive Mining", "Enchantment", lpAggressiveMiningOracle, false)
	land := lpHandLand(me2, "Forest")
	lpAssertRefused(t, g2, me2, land, lpPlay(t, g2, me2, land), "You can't play lands.")

	a := seedLandOnBattlefield(g2, me2.ID, "Forest", "Basic Land — Forest")
	b := seedLandOnBattlefield(g2, me2.ID, "Forest", "Basic Land — Forest")
	hand := me2.Hand.Size()
	b16Activate(t, g2, me2.ID, mining, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{a}})
	if got := me2.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d → %d, want +2", hand, got)
	}
	if err := g2.ActivateCatalogAbility(me2.ID, mining, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{b}}); err == nil {
		t.Error("the second activation this turn was allowed")
	}
}

// City in a Bottle: lands and spells with an Arabian Nights name are
// refused, anything else is not, and the state trigger sacrifices every
// other nontoken permanent with such a name (a token, and the City itself,
// are left alone).
func TestCityInABottle(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	desertOnBoard := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Desert", TypeLine: "Land"})
	oasis := apaPush(g, me.ID, me.ID, game.Card{Name: "Oasis", TypeLine: "Land"})
	token := apaPush(g, me.ID, me.ID, game.Card{Name: "Desert", TypeLine: "Token Land"})
	city := apaPush(g, me.ID, me.ID, game.Card{Name: "City in a Bottle", OracleID: lpCityInABottleOracle, TypeLine: "Artifact"})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{desertOnBoard, oasis} {
		if g.Battlefield.Contains(id) {
			t.Errorf("an Arabian Nights permanent survived City in a Bottle: %v", id)
		}
	}
	if !g.Battlefield.Contains(token) || !g.Battlefield.Contains(city) {
		t.Error("a token, or the City itself, was sacrificed")
	}
	if !opp.Graveyard.Contains(desertOnBoard) {
		t.Error("the opponent's Desert was not sacrificed by its controller")
	}

	forest := lpHandLand(me, "Forest")
	if err := lpPlay(t, g, me, forest); err != nil {
		t.Fatalf("a land with another name was refused: %v", err)
	}
	library := lpHandLand(me, "Library of Alexandria")
	g.WithWriteLock(func() { g.GrantAdditionalLandPlayForEffect(me.ID, 1) })
	err := g.CastSpell(me.ID, library, game.CastSpellParams{})
	lpAssertRefusedAfterDrop(t, g, me, library, err, "Players can't play lands with a name originally printed in the Arabian Nights expansion.")

	twister := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: twister, Name: "Desert Twister", TypeLine: "Sorcery", ManaCost: "{4}{G}{G}",
		Owner: me.ID, Controller: me.ID})
	err = g.CastSpell(me.ID, twister, game.CastSpellParams{})
	assertCantCast(t, err, "Arabian Nights")
	if lpGate(g, opp, "Ydwen Efreet", game.ZoneGraveyard) == nil {
		t.Error("an opponent may play an Arabian Nights land from a graveyard")
	}
}

// lpAssertRefusedAfterDrop is lpAssertRefused for a play made after
// another land play this turn.
func lpAssertRefusedAfterDrop(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, err error, want string) {
	t.Helper()
	if !errors.Is(err, game.ErrCantPlayLand) {
		t.Fatalf("refusal is %v, want it to wrap game.ErrCantPlayLand", err)
	}
	var cant *game.CantPlayLandError
	if !errors.As(err, &cant) || !strings.Contains(cant.Reason, want) {
		t.Fatalf("refusal = %v, want %q", err, want)
	}
	if !p.Hand.Contains(id) {
		t.Error("a refused land play did not leave the card in hand")
	}
}

// Turf Wound: the target can't play lands this turn, the caster draws, and
// the record is gone next turn.
func TestTurfWoundBansThePlayerForTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Turf Wound", "Instant", lpTurfWoundOracle, pr6Player(opp.ID))
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d after the cast and the draw, want +1", hand, got)
	}
	err := lpGate(g, opp, "Forest", game.ZoneHand)
	if err == nil || !strings.Contains(err.Error(), "You can't play lands this turn — Turf Wound") {
		t.Fatalf("the target's gate = %v, want the Turf Wound clause", err)
	}
	if err := lpGate(g, me, "Forest", game.ZoneHand); err != nil {
		t.Errorf("the caster is banned too: %v", err)
	}
	g.WithWriteLock(func() {
		if got := g.LandPlayBanFor(opp.ID); !strings.Contains(got, "Turf Wound") {
			t.Errorf("the seat's banner = %q", got)
		}
	})
	advanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if err := lpGate(g, opp, "Forest", game.ZoneHand); err != nil {
		t.Errorf("the ban outlived the turn: %v", err)
	}
}

// Solfatara: Turf Wound's ban, then a draw at the next upkeep, whoever's it
// is.
func TestSolfataraBansAndDrawsAtTheNextUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	castCatalogSpell(t, g, "Solfatara", "Instant", lpSolfataraOracle, pr6Player(opp.ID))
	passPriorityAroundTable(t, g)
	if lpGate(g, opp, "Forest", game.ZoneHand) == nil {
		t.Fatal("the target may still play a land")
	}
	if n := len(g.DelayedTriggers); n != 1 {
		t.Fatalf("delayed triggers = %d, want the upkeep draw", n)
	}
	hand := me.Hand.Size()
	advanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d at the next upkeep, want +1", hand, got)
	}
}

// Pardic Miner: the sacrifice is the cost, and the ban survives its source.
func TestPardicMinerSacrificesToBanALandPlay(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	miner := pushCatalogPermanent(g, me.ID, "Pardic Miner", "Creature — Dwarf", lpPardicMinerOracle, false)
	b16Activate(t, g, me.ID, miner, 0, game.ActivateAbilityParams{Targets: pr6Player(opp.ID)})
	if g.Battlefield.Contains(miner) {
		t.Error("Pardic Miner was not sacrificed")
	}
	err := lpGate(g, opp, "Forest", game.ZoneHand)
	if err == nil || !strings.Contains(err.Error(), "Pardic Miner") {
		t.Fatalf("the ban = %v, want it to name Pardic Miner after the Miner has gone", err)
	}
}

// Moonhold: red buys the land ban, white the creature-spell ban, and both
// colours buy both.
func TestMoonholdReadsTheColoursSpent(t *testing.T) {
	cases := []struct {
		pool            string
		wantLandBan     bool
		wantCreatureBan bool
	}{
		{"RCC", true, false},
		{"WCC", false, true},
		{"RWC", true, true},
	}
	for _, tc := range cases {
		t.Run(tc.pool, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			advanceToMain(t, g)
			floatForTest(g, me, tc.pool)
			id := uuid.New()
			me.Hand.PushTop(game.Card{InstanceID: id, Name: "Moonhold", TypeLine: "Instant", ManaCost: "{2}{R/W}",
				OracleID: lpMoonholdOracle, Owner: me.ID, Controller: me.ID})
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: pr6Player(opp.ID)}); err != nil {
				t.Fatalf("cast Moonhold: %v", err)
			}
			passPriorityAroundTable(t, g)
			if banned := lpGate(g, opp, "Forest", game.ZoneHand) != nil; banned != tc.wantLandBan {
				t.Errorf("land ban = %v, want %v", banned, tc.wantLandBan)
			}
			var creatureErr error
			g.WithWriteLock(func() {
				creatureErr = g.CastGateLocked(opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear"}, game.ZoneHand, game.CastSpellParams{})
			})
			if banned := creatureErr != nil; banned != tc.wantCreatureBan {
				t.Errorf("creature-spell ban = %v, want %v (%v)", banned, tc.wantCreatureBan, creatureErr)
			}
			var instantErr error
			g.WithWriteLock(func() {
				instantErr = g.CastGateLocked(opp.ID, game.Card{Name: "Shock", TypeLine: "Instant"}, game.ZoneHand, game.CastSpellParams{})
			})
			if instantErr != nil {
				t.Errorf("a noncreature spell was banned: %v", instantErr)
			}
		})
	}
}

// Rock Jockey: it can't be cast after a land play, and a land can't be
// played once it was cast this turn — but not if it has left the battlefield
// since (a new object was not cast, CR 400.7).
func TestRockJockeyLocksTheTurnBothWays(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if err := lpPlay(t, g, me, lpHandLand(me, "Forest")); err != nil {
		t.Fatalf("setup land play: %v", err)
	}
	jockey := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: jockey, Name: "Rock Jockey", TypeLine: "Creature — Goblin", ManaCost: "{2}{R}",
		OracleID: lpRockJockeyOracle, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	assertCantCast(t, g.CastSpell(me.ID, jockey, game.CastSpellParams{}), "played a land this turn")

	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	advanceToMain(t, g2)
	jockey2 := uuid.New()
	me2.Hand.PushTop(game.Card{InstanceID: jockey2, Name: "Rock Jockey", TypeLine: "Creature — Goblin", ManaCost: "{2}{R}",
		OracleID: lpRockJockeyOracle, Power: 3, Toughness: 3, Owner: me2.ID, Controller: me2.ID})
	if err := g2.CastSpell(me2.ID, jockey2, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Rock Jockey with no land played: %v", err)
	}
	passPriorityAroundTable(t, g2)
	land := lpHandLand(me2, "Forest")
	lpAssertRefused(t, g2, me2, land, g2.CastSpell(me2.ID, land, game.CastSpellParams{}), "if this creature was cast this turn")
	g2.WithWriteLock(func() {
		g2.EmitEvent(game.Event{Kind: game.EventLTB, CardID: jockey2, NewZone: game.ZoneHand})
	})
	if err := lpGate(g2, me2, "Forest", game.ZoneHand); err != nil {
		t.Errorf("a Jockey that left the battlefield still locks the turn: %v", err)
	}
}

// Ward of Bones: an opponent who controls MORE of a type than the Ward's
// controller can't cast that type or (for lands) play lands; a tie, or
// fewer, is free.
func TestWardOfBonesComparesPermanentCounts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Ward of Bones", "Artifact", lpWardOfBonesOracle, false)
	cast := func(typeLine string) error {
		var err error
		g.WithWriteLock(func() {
			err = g.CastGateLocked(opp.ID, game.Card{Name: "X", TypeLine: typeLine}, game.ZoneHand, game.CastSpellParams{})
		})
		return err
	}
	if cast("Creature — Bear") != nil || lpGate(g, opp, "Forest", game.ZoneHand) != nil {
		t.Fatal("an opponent with nothing is restricted")
	}
	seedCreature(g, "A", opp.ID)
	if err := cast("Creature — Bear"); err == nil {
		t.Error("an opponent with more creatures may cast creature spells")
	}
	if err := cast("Artifact"); err != nil {
		t.Errorf("creature counts restricted artifact spells: %v", err)
	}
	seedCreature(g, "B", me.ID)
	if err := cast("Creature — Bear"); err != nil {
		t.Errorf("a tie in creatures restricts: %v", err)
	}
	seedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	err := lpGate(g, opp, "Forest", game.ZoneHand)
	if err == nil || !strings.Contains(err.Error(), "controls more lands") {
		t.Errorf("an opponent with more lands: %v", err)
	}
	if lpGate(g, me, "Forest", game.ZoneHand) != nil {
		t.Error("the Ward's controller is restricted")
	}
	seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	if err := lpGate(g, opp, "Forest", game.ZoneHand); err != nil {
		t.Errorf("a tie in lands restricts: %v", err)
	}
}

// Experimental Frenzy: no plays from your hand, plays from the top of the
// library still work, and the commander in the command zone is untouched.
func TestExperimentalFrenzyClosesTheHandAndOpensTheLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Experimental Frenzy", "Enchantment", lpExperimentalFrenzyOracle, false)
	handLand := lpHandLand(me, "Forest")
	lpAssertRefused(t, g, me, handLand, lpPlay(t, g, me, handLand), "You can't play lands from your hand.")

	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Shock", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	assertCantCast(t, g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: pr6Player(g.Seats[1].ID)}), "from your hand")

	top := seedLibraryTop(me, "Mountain", "Basic Land — Mountain", "")
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: string(game.ZoneLibrary)}); err != nil {
		t.Fatalf("play a land from the top of the library: %v", err)
	}
	if !g.Battlefield.Contains(top) {
		t.Error("the land from the library top did not enter")
	}
}
