package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// player_keywords_if_cards_test.go — #2013: a permanent that gives its
// controller hexproof only while a condition holds. Proved through the
// engine's targeting gate, from both sides of the condition.

const (
	gruulSpellbreakerOracle = "827b3ed0-09b3-4d89-8b81-dd89cc002782"
	captainAmericaOracle    = "ac1b6ffc-12a9-4d9c-a9fc-68a839a40738"
)

// opponentMayTarget asks the targeting gate whether a spell `caster`
// controls may choose `target` — the predicate the announce gate, the
// resolution re-check and the picker all share.
func opponentMayTarget(g *game.Game, caster, target uuid.UUID) bool {
	var ok bool
	g.ReadSnapshot(func() { ok = g.CanPlayerBeTargetedByForEffect(target, game.SourceChooser(caster)) })
	return ok
}

// TestGruulSpellbreakerShieldsYouOnlyDuringYourTurn is the card: on its
// controller's turn the controller and the creature have hexproof, and on
// an opponent's turn neither does.
func TestGruulSpellbreakerShieldsYouOnlyDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if g.Turn.ActiveSeat != 0 {
		t.Fatalf("the game should start on seat 0's turn, got %d", g.Turn.ActiveSeat)
	}
	breaker := pushCatalogPermanent(g, me.ID, "Gruul Spellbreaker", "Creature — Ogre Warrior", gruulSpellbreakerOracle, false)

	// My turn: I have hexproof, and so does the creature.
	if got := playerAbilities(g, me); !hasPlayerAbility(got, KeywordHexproof) {
		t.Fatalf("on its controller's turn the controller has %v, want hexproof", got)
	}
	if !effectiveAbilitiesContain(t, g, breaker, KeywordHexproof) {
		t.Error("on its controller's turn the Spellbreaker has no hexproof")
	}
	if opponentMayTarget(g, opp.ID, me.ID) {
		t.Error("an opponent may target the Spellbreaker's controller during their turn")
	}
	if !opponentMayTarget(g, me.ID, me.ID) {
		t.Error("the controller can't target themselves under their own hexproof")
	}
	if got := playerAbilities(g, opp); len(got) != 0 {
		t.Errorf("the opponent was given %v", got)
	}

	// The opponent's turn: neither has it.
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if got := playerAbilities(g, me); hasPlayerAbility(got, KeywordHexproof) {
		t.Errorf("on an opponent's turn the controller still has %v", got)
	}
	if effectiveAbilitiesContain(t, g, breaker, KeywordHexproof) {
		t.Error("on an opponent's turn the Spellbreaker still has hexproof")
	}
	if !opponentMayTarget(g, opp.ID, me.ID) {
		t.Error("on their own turn the opponent can't target the Spellbreaker's controller")
	}
	// A Bolt cast on that turn is accepted by the announce gate.
	if err := boltAtPlayerErr(t, g, opp, me.ID); err != nil {
		t.Errorf("Bolt at the Spellbreaker's controller on the opponent's turn: %v", err)
	}
}

// TestGruulSpellbreakerOnAnOpponentsSideShieldsThemNotYou: the
// condition is the controller's turn, so a Spellbreaker an opponent
// controls gives nothing on my turn and the Bolt goes through.
func TestGruulSpellbreakerOnAnOpponentsSideShieldsThemNotYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, opp.ID, "Gruul Spellbreaker", "Creature — Ogre Warrior", gruulSpellbreakerOracle, false)
	if got := playerAbilities(g, opp); hasPlayerAbility(got, KeywordHexproof) {
		t.Errorf("on my turn the opponent's Spellbreaker gives them %v", got)
	}
	if err := boltAtPlayerErr(t, g, me, opp.ID); err != nil {
		t.Errorf("Bolt at the opponent on my turn: %v", err)
	}
}

// TestCaptainAmericaShieldsYouWhileHeHasAShieldCounter: he enters with
// the counter, you and your other Heroes have hexproof while it is
// there, and the first destroy effect spends it and ends both.
func TestCaptainAmericaShieldsYouWhileHeHasAShieldCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var opp *game.Player
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opp = p
			break
		}
	}
	captain := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: captain, Name: "Captain America, Super-Soldier", OracleID: captainAmericaOracle,
		TypeLine: "Legendary Creature — Human Soldier Hero", ManaCost: "{1}{W}{W}",
		Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	hero := pushCatalogPermanent(g, me.ID, "Test Hero", "Creature — Human Hero", "", false)
	soldier := pushCatalogPermanent(g, me.ID, "Test Soldier", "Creature — Human Soldier", "", false)
	theirHero := pushCatalogPermanent(g, opp.ID, "Their Hero", "Creature — Human Hero", "", false)

	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, captain, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Captain America: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := counterCount(g, captain, game.CounterShield); n != 1 {
		t.Fatalf("Captain America entered with %d shield counters, want 1", n)
	}

	if got := playerAbilities(g, me); !hasPlayerAbility(got, KeywordHexproof) {
		t.Errorf("with a shield counter on him his controller has %v, want hexproof", got)
	}
	if opponentMayTarget(g, opp.ID, me.ID) {
		t.Error("an opponent may target Captain America's controller while he is shielded")
	}
	if !effectiveAbilitiesContain(t, g, hero, KeywordHexproof) {
		t.Error("another Hero you control has no hexproof")
	}
	if effectiveAbilitiesContain(t, g, captain, KeywordHexproof) {
		t.Error("Captain America gave himself hexproof; the grant is to OTHER Heroes")
	}
	if effectiveAbilitiesContain(t, g, soldier, KeywordHexproof) {
		t.Error("a non-Hero was given hexproof")
	}
	if effectiveAbilitiesContain(t, g, theirHero, KeywordHexproof) {
		t.Error("an opponent's Hero was given hexproof")
	}

	// A destroy effect spends the counter (CR 122.1c) and he stays.
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(captain); err != nil {
			t.Fatalf("destroy Captain America: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Captain America, Super-Soldier") == uuid.Nil {
		t.Fatal("the shield counter did not save him")
	}
	if n := counterCount(g, captain, game.CounterShield); n != 0 {
		t.Fatalf("the destroy left %d shield counters", n)
	}
	if got := playerAbilities(g, me); hasPlayerAbility(got, KeywordHexproof) {
		t.Errorf("with no shield counter his controller still has %v", got)
	}
	if !opponentMayTarget(g, opp.ID, me.ID) {
		t.Error("with no shield counter an opponent still can't target his controller")
	}
	if effectiveAbilitiesContain(t, g, hero, KeywordHexproof) {
		t.Error("with no shield counter the other Hero still has hexproof")
	}
}

// TestPlayerKeywordsIfRefusesAnEntryWithNoCondition: a conditional
// player ability with no condition would grant nothing, silently, so
// Register stops the boot.
func TestPlayerKeywordsIfRefusesAnEntryWithNoCondition(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("checkPlayerKeywords accepted a PlayerKeywordsIf entry with no condition")
		}
	}()
	checkPlayerKeywords(Spec{Name: "Test", PlayerKeywordsIf: []game.PlayerKeywordIf{{Keyword: KeywordHexproof, Label: "x"}}})
}

// TestPlayerKeywordsIfRefusesATokenNobodyReads: the vocabulary is
// PlayerKeywords', so shroud is refused here too.
func TestPlayerKeywordsIfRefusesATokenNobodyReads(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("checkPlayerKeywords accepted shroud as a conditional player ability")
		}
	}()
	checkPlayerKeywords(Spec{Name: "Test", PlayerKeywordsIf: []game.PlayerKeywordIf{
		YouHaveWhile("shroud", "You have shroud.", func(*game.Game, game.Card) bool { return true }),
	}})
}
