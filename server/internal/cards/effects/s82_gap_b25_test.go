package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const giverOfRunesOracle = "4ec85850-f274-4c0c-9a03-0488267caa14"

func TestGiverOfRunesGrantsColorlessOrAColorToAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	giver := pushCatalogPermanent(g, me.ID, "Giver of Runes", "Creature — Kor Cleric", giverOfRunesOracle, false)
	bear := aiBear(g, me.ID)

	// "Another": it may not target itself.
	if err := g.ActivateCatalogAbility(me.ID, giver, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: giver}},
	}); err == nil {
		t.Fatal("Giver of Runes targeted itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, giver, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := pendingOfKind(g, game.PendingChoiceColor)
	if c == nil || len(c.ColorOptions) != 6 {
		t.Fatalf("want the colorless-or-color prompt, got %+v", c)
	}
	answerColor(t, g, me.ID, "C")
	if got := protectionsOn(t, g, bear); len(got) != 1 || got[0] != "colorless" {
		t.Fatalf("protections = %q, want [colorless]", got)
	}
	if got := protectionsOn(t, g, giver); len(got) != 0 {
		t.Errorf("Giver protected itself: %q", got)
	}
	// It tapped.
	if err := g.ActivateCatalogAbility(me.ID, giver, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Error("a tapped Giver of Runes activated again")
	}

}

const (
	ionaShieldOfEmeriaOracle = "5e816472-561d-43d1-ad28-80a829e2a91a"
	ethersworCanonistOracle  = "ec3471ad-b65c-4e35-a7d1-68da9555a9ec"
	containmentPriestOracle  = "8a76a9b9-3127-45fe-b20f-a8f643276281"
	jeskaiBallerOracle       = "e789e7bf-d2f9-41e8-976f-24321e8da928"
	nobleOxOracle            = "b152753f-0516-4e21-b823-608e9d0c9c82"
	newMasterOfArmsOracle    = "d44b7b43-4674-47cd-b346-7db93a14b74c"
	katakiWarsWageOracle     = "27ffa6e9-c161-4f53-8b41-7fae5f199f41"
)

// gapCastFromHand puts `c` in the hand of the active player, moves to the
// precombat main phase and casts it (strictly), returning the card id.
func gapCastFromHand(t *testing.T, g *game.Game, c game.Card) (uuid.UUID, error) {
	t.Helper()
	p := g.Seats[g.Turn.ActiveSeat]
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Hand.PushTop(c)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return c.InstanceID, g.CastSpell(p.ID, c.InstanceID, game.CastSpellParams{Strict: true})
}

func TestIonaStopsOpponentsCastingTheChosenColor(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushChosenColorPermanent(t, g, other.ID, "Iona, Shield of Emeria", "Legendary Creature — Angel", ionaShieldOfEmeriaOracle, "R")

	if _, err := gapCastFromHand(t, g, game.Card{Name: "Shock", TypeLine: "Instant", Colors: []string{"R"}}); err == nil {
		t.Error("an opponent cast a spell of the chosen color")
	}
	if _, err := gapCastFromHand(t, g, game.Card{Name: "Opt", TypeLine: "Instant", Colors: []string{"U"}}); err != nil {
		t.Errorf("a spell of another color was refused: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, err := gapCastFromHand(t, g, game.Card{Name: "Ornithopter", TypeLine: "Artifact Creature — Thopter"}); err != nil {
		t.Errorf("a colorless spell was refused: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Iona's controller is not an opponent.
	pushChosenColorPermanent(t, g, caster.ID, "Iona, Shield of Emeria", "Legendary Creature — Angel", ionaShieldOfEmeriaOracle, "G")
	if _, err := gapCastFromHand(t, g, game.Card{Name: "Giant Growth", TypeLine: "Instant", Colors: []string{"G"}}); err != nil {
		t.Errorf("Iona's controller was restricted by their own Iona: %v", err)
	}
}

func TestEtherswornCanonistLimitsNonartifactSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Ethersworn Canonist", "Artifact Creature — Human Cleric", ethersworCanonistOracle, false)

	if _, err := gapCastFromHand(t, g, game.Card{Name: "Opt", TypeLine: "Instant", Colors: []string{"U"}}); err != nil {
		t.Fatalf("the first nonartifact spell was refused: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, err := gapCastFromHand(t, g, game.Card{Name: "Shock", TypeLine: "Instant", Colors: []string{"R"}}); err == nil {
		t.Error("a second nonartifact spell was cast")
	}
	if _, err := gapCastFromHand(t, g, game.Card{Name: "Ornithopter", TypeLine: "Artifact Creature — Thopter"}); err != nil {
		t.Errorf("an artifact spell was refused: %v", err)
	}
}

func TestEtherswornCanonistAnArtifactSpellDoesNotUseItUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Ethersworn Canonist", "Artifact Creature — Human Cleric", ethersworCanonistOracle, false)
	if _, err := gapCastFromHand(t, g, game.Card{Name: "Ornithopter", TypeLine: "Artifact Creature — Thopter"}); err != nil {
		t.Fatalf("artifact spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, err := gapCastFromHand(t, g, game.Card{Name: "Opt", TypeLine: "Instant", Colors: []string{"U"}}); err != nil {
		t.Errorf("the first nonartifact spell was refused after an artifact spell: %v", err)
	}
}

func TestContainmentPriestExilesCreaturesThatWereNotCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Containment Priest", "Creature — Human Cleric", containmentPriestOracle, false)

	// Put from a graveyard: exiled instead.
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, dead); err != nil {
		t.Fatalf("move: %v", err)
	}
	if g.Battlefield.Contains(dead) {
		t.Error("a creature that was not cast entered the battlefield")
	}
	if !g.Exile.Contains(dead) {
		t.Error("the creature was not exiled instead")
	}

	// A noncreature permanent is not affected.
	rock := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: rock, Name: "Mind Stone", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, rock); err != nil {
		t.Fatalf("move: %v", err)
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("an artifact was exiled")
	}

	// A cast creature spell resolves normally.
	cast, err := gapCastFromHand(t, g, game.Card{Name: "Runeclaw Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cast) {
		t.Error("a creature that was cast did not enter the battlefield")
	}

	// A token is not a nontoken creature.
	g.WithWriteLock(func() {
		if err := (CreateToken{Template: TokenCard("1/1 white Soldier"), N: 1}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatalf("token: %v", err)
		}
	})
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Soldier" {
			found = true
		}
	}
	if !found {
		t.Error("a token was exiled")
	}
}

func TestJeskaiBallerMakesAnAthleteWhenCast(t *testing.T) {
	g := newCatalogGame(t)
	id, err := gapCastFromHand(t, g, game.Card{Name: "Jeskai Baller", TypeLine: "Creature — Human Athlete", OracleID: jeskaiBallerOracle,
		Colors: []string{"W"}, Power: 2, Toughness: 3})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	// The trigger is on the stack above the spell: the token exists before it resolves.
	passPriorityAroundTable(t, g)
	athletes := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Athlete" && c.IsToken() {
			athletes++
		}
	}
	if athletes != 1 {
		t.Errorf("Athlete tokens = %d, want 1", athletes)
	}
	if !g.Battlefield.Contains(id) {
		t.Error("the Baller did not enter the battlefield (rebound exiles an instant or sorcery, CR 702.88a)")
	}
}

func TestNobleOxBlocksEveryUnblockedAttackerOnYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ox := pushCatalogPermanent(g, opp.ID, "Noble Ox", "Creature — Noble Ox", nobleOxOracle, false)
	atks := mbAttack(t, g, me, opp, 3)
	if err := mbBlock(g, pushVanillaCreature(g, opp.ID, "Chump", 1, 1), atks[0]); err != nil {
		t.Fatalf("block: %v", err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	var n int
	g.WithWriteLock(func() { n = g.BlockUnblockedAttackersForEffect(ox, opp.ID) })
	if n != 2 {
		t.Fatalf("Ox blocked %d attackers, want the 2 unblocked ones", n)
	}
	got := mbBlocked(t, g, ox)
	if len(got) != 2 {
		t.Fatalf("Ox blocks %d, want 2", len(got))
	}
	for _, a := range atks[1:] {
		if !g.BlockedAttackerForEffect(a) {
			t.Errorf("attacker %v is not blocked", a)
		}
	}
	// Run it again: nothing is left unblocked.
	g.WithWriteLock(func() { n = g.BlockUnblockedAttackersForEffect(ox, opp.ID) })
	if n != 0 {
		t.Errorf("second pass blocked %d, want 0", n)
	}
}

func TestNobleOxDoesNothingForAttackersOnSomeoneElse(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	third := g.Seats[2]
	ox := pushCatalogPermanent(g, third.ID, "Noble Ox", "Creature — Noble Ox", nobleOxOracle, false)
	mbAttack(t, g, me, opp, 2)
	var n int
	g.WithWriteLock(func() { n = g.BlockUnblockedAttackersForEffect(ox, third.ID) })
	if n != 0 {
		t.Errorf("Ox blocked %d creatures attacking another player", n)
	}
}

func TestNobleOxETBTriggerIsRegistered(t *testing.T) {
	spec, ok := Lookup(nobleOxOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Triggered) != 1 || len(spec.Static) != 1 {
		t.Fatalf("Noble Ox spec = %+v", spec)
	}
	g := newCatalogGame(t)
	opp := g.Seats[1]
	ox := pushCatalogPermanent(g, opp.ID, "Noble Ox", "Creature — Noble Ox", nobleOxOracle, false)
	if got := mbCapacity(t, g, ox); got != 0 {
		t.Errorf("capacity = %d, want 0 (any number)", got)
	}
}

func TestNewMasterOfArmsPreventsDamageFromTappedBlockers(t *testing.T) {
	for _, tapIt := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		master := pushCatalogPermanent(g, me.ID, "New Master of Arms", "Creature — Human Soldier", newMasterOfArmsOracle, false)
		g.WithWriteLock(func() {
			c := findBattlefieldCardForTest(g, master)
			c.Power, c.Toughness = 2, 2
		})
		blocker := pushVanillaCreature(g, opp.ID, "Big Blocker", 3, 5)
		declareAttack(t, g, opp.ID, master)
		advanceTo(t, g, game.StepDeclareBlockers)
		if err := mbBlock(g, blocker, master); err != nil {
			t.Fatalf("block: %v", err)
		}
		if err := g.FinishBlocks(opp.ID); err != nil {
			t.Fatalf("FinishBlocks: %v", err)
		}
		if tapIt {
			dcMana(t, g, me.ID, "{W}{W}")
			if err := g.ActivateCatalogAbility(me.ID, master, 0, game.ActivateAbilityParams{
				Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: blocker}},
			}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			passPriorityAroundTable(t, g)
			if !e2Card(t, g, blocker).Tapped {
				t.Fatal("the blocker was not tapped")
			}
		}
		advanceTo(t, g, game.StepEndCombat)
		alive := g.Battlefield.Contains(master)
		if tapIt && !alive {
			t.Error("a tapped blocker's combat damage was not prevented")
		}
		if !tapIt && alive {
			t.Error("control: an untapped 3-power blocker should have killed the 2/2 Master")
		}
	}
}

func TestNewMasterOfArmsCannotTapACreatureThatIsNotBlockingIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	master := pushCatalogPermanent(g, me.ID, "New Master of Arms", "Creature — Human Soldier", newMasterOfArmsOracle, false)
	idle := pushVanillaCreature(g, opp.ID, "Idle", 1, 1)
	dcMana(t, g, me.ID, "{W}{W}")
	if err := g.ActivateCatalogAbility(me.ID, master, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: idle}},
	}); err == nil {
		t.Error("tapped a creature that is not blocking")
	}
}

func TestKatakiGivesEveryArtifactAnUpkeepTax(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		pushCatalogPermanent(g, opp.ID, "Kataki, War's Wage", "Legendary Creature — Spirit", katakiWarsWageOracle, false)
		mine := pushCatalogPermanent(g, me.ID, "Mind Stone", "Artifact", "", false)
		bear := pushCatalogPermanent(g, me.ID, "Grizzly Bears", "Creature — Bear", "", false)
		walkToUpkeepOfSeat0(t, g)
		passPriorityAroundTable(t, g)
		if pay {
			dcMana(t, g, me.ID, "{W}")
		}
		answerPayUnless(t, g, me.ID, pay)
		if got := g.Battlefield.Contains(mine); got != pay {
			t.Errorf("paid %v: my artifact on the battlefield = %v", pay, got)
		}
		if !g.Battlefield.Contains(bear) {
			t.Errorf("paid %v: a nonartifact permanent was taxed", pay)
		}
	}
}
