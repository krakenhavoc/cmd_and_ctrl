package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gap29StonySilence  = "12c63a45-7eb2-4262-b571-3358182f5610"
	gap29CommonBlack   = "f64eff0e-dda7-457b-adf9-aa4e537f4f58"
	gap29OnceUponATime = "d00ade82-caab-43c6-9756-6e9e6edfc519"
	gap29Seedtime      = "3ba50802-57d0-450e-9e2f-a46a09d2f6bb"
	gap29SummonersPact = "782dcbc4-5a11-42f5-b8d9-4572e36d45b3"
	gap29Whiteout      = "9e8de3e1-1159-460f-99b7-1043e5814597"
	gap29AncientGrudge = "306593b0-6ea8-476f-a3f2-e17876c1bab4"
)

func TestGapB29CardsAreRegisteredFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		gap29StonySilence:  "Stony Silence",
		gap29CommonBlack:   "Common Black Removal",
		gap29OnceUponATime: "Once Upon a Time",
		gap29Seedtime:      "Seedtime",
		gap29SummonersPact: "Summoner's Pact",
		gap29Whiteout:      "Whiteout",
		gap29AncientGrudge: "Ancient Grudge",
	} {
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

// Stony Silence stops a Sol Ring's mana ability, mana abilities included.
func TestGapB29StonySilenceStopsArtifactAbilities(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	solRing := b21Push(g, me.ID, "Sol Ring", "Artifact", actSolRingOracle, 0, 0)
	if err := g.ActivateManaAbility(me.ID, solRing, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("Sol Ring before: %v", err)
	}
	b21Push(g, me.ID, "Stony Silence", "Enchantment", gap29StonySilence, 0, 0, "W")
	g.WithWriteLock(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == solRing {
				c.Tapped = false
			}
		}
	})
	if err := g.ActivateManaAbility(me.ID, solRing, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrCantActivate) {
		t.Errorf("Sol Ring under Stony Silence: err = %v, want ErrCantActivate", err)
	}
}

// Ancient Grudge destroys an artifact, and flashes back for {G}.
func TestGapB29AncientGrudgeDestroysAndFlashesBack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	spell := castCatalogSpell(t, g, "Ancient Grudge", "Instant", gap29AncientGrudge, cardTarget(rock))
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(rock) {
		t.Fatalf("the artifact was not destroyed")
	}
	if !me.Graveyard.Contains(spell) {
		t.Fatalf("Ancient Grudge is not in the graveyard")
	}
	rock2 := b12Permanent(g, opp.ID, "Rock 2", "Artifact")
	apaMana(me, "G")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{
		Strict: true, FromZone: "graveyard", AlternativeCost: "flashback", Targets: cardTarget(rock2),
	}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(rock2) || !g.Exile.Contains(spell) {
		t.Errorf("flashback did not destroy the artifact and exile the spell")
	}
}

func TestGapB29AncientGrudgeRefusesANonArtifact(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	err := castCatalogSpellErrStrict(t, g, "Ancient Grudge", "Instant", gap29AncientGrudge, cardTarget(bear))
	if err == nil {
		t.Errorf("Ancient Grudge was cast at a creature")
	}
}

// Whiteout strips flying from every creature, and returns from the
// graveyard for a snow land sacrifice.
func TestGapB29WhiteoutStripsFlyingAndReturns(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flyer := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1,
		Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	spell := castCatalogSpell(t, g, "Whiteout", "Snow Instant", gap29Whiteout, nil)
	passPriorityAroundTable(t, g)
	if c := apaLive(g, flyer); c == nil || game.HasKeyword(c, "flying") {
		t.Fatalf("the flyer kept flying")
	}
	if !me.Graveyard.Contains(spell) {
		t.Fatalf("Whiteout is not in the graveyard")
	}
	snow := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Snow-Covered Forest", TypeLine: "Basic Snow Land — Forest",
		Owner: me.ID, Controller: me.ID,
	})
	if err := g.ActivateCatalogAbility(me.ID, spell, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{snow}}); err != nil {
		t.Fatalf("activate from graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(spell) {
		t.Errorf("Whiteout did not return to hand")
	}
}

// Summoner's Pact fetches a green creature and schedules the upkeep debt.
func TestGapB29SummonersPactFetchesAndSchedulesTheDebt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	green := pushLibraryCardForTest(me, game.Card{Name: "Elf", TypeLine: "Creature — Elf", Colors: []string{"G"}})
	red := pushLibraryCardForTest(me, game.Card{Name: "Goblin", TypeLine: "Creature — Goblin", Colors: []string{"R"}})
	castCatalogSpell(t, g, "Summoner's Pact", "Instant", gap29SummonersPact, nil)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, green)
	_ = red
	found := false
	for _, d := range g.DelayedTriggers {
		if d != nil && d.Controller == me.ID && d.Body.Key() == "pact/pay-or-lose" {
			found = true
		}
	}
	if !found {
		t.Error("no pact payment scheduled")
	}
	if !me.Hand.Contains(green) {
		t.Errorf("the green creature is not in hand")
	}
}

// Once Upon a Time's free cast is offered only for the first spell.
func TestGapB29OnceUponATimeFreeOnlyAsTheFirstSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	first := handCardFull(me, "Once Upon a Time", "Instant", "{1}{G}", gap29OnceUponATime, []string{"G"})
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Strict: true, AlternativeCost: "free"}); err != nil {
		t.Fatalf("free cast as the first spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	second := handCardFull(me, "Once Upon a Time", "Instant", "{1}{G}", gap29OnceUponATime, []string{"G"})
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{Strict: true, AlternativeCost: "free"}); err == nil {
		t.Errorf("the free cast was offered for the second spell")
	}
}

// Seedtime: only on your turn, and the extra turn needs an opponent's blue spell.
func TestGapB29SeedtimeTakesAnExtraTurnAfterABlueOpponentSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	g.WithWriteLock(func() {
		if g.SpellsCastThisTurn == nil {
			g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{}
		}
		g.SpellsCastThisTurn[g.Seats[1].ID] = game.CastTally{Total: 1, Colors: "U"}
	})
	castCatalogSpell(t, g, "Seedtime", "Instant", gap29Seedtime, nil)
	passPriorityAroundTable(t, g)
	endTurn(t, g)
	assertTurnOf(t, g, me, true, "after Seedtime with a blue opposing spell")
}

func TestGapB29SeedtimeDoesNothingWithoutABlueOpponentSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	castCatalogSpell(t, g, "Seedtime", "Instant", gap29Seedtime, nil)
	passPriorityAroundTable(t, g)
	endTurn(t, g)
	assertTurnOf(t, g, (me+1)%len(g.Seats), false, "after Seedtime with no blue opposing spell")
}

// Common Black Removal: destroys the creature, then the chosen bullet.
func TestGapB29CommonBlackRemovalFoodAndMill(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	castCatalogSpellWithModesTargets(t, g, gap29CommonBlack, []int{0}, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bear) {
		t.Fatalf("the creature was not destroyed")
	}
	if n, _ := countTokensNamed(g, me.ID, "Food"); n != 1 {
		t.Errorf("Food tokens %d, want 1", n)
	}
}

func TestGapB29CommonBlackRemovalMillsItsControllerByPower(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	before := opp.Library.Size()
	castCatalogSpellWithModesTargets(t, g, gap29CommonBlack, []int{3}, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != before-3 {
		t.Errorf("library %d → %d, want -3", before, got)
	}
}
