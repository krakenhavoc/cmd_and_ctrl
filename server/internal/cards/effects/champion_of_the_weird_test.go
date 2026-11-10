package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const championOfTheWeirdOracle = "1726f9cd-6f18-4509-b95b-bca82ba795a4"

func TestChampionOfTheWeirdIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(championOfTheWeirdOracle)
	if !ok || spec.Name != "Champion of the Weird" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// castChampionOfTheWeird casts it strictly for {3}{B}, beholding `reveal`.
func castChampionOfTheWeird(t *testing.T, g *game.Game, reveal ...uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	advanceTo(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	b06AddMana(me, "B", "C", "C", "C")
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Champion of the Weird", TypeLine: "Creature — Goblin Berserker",
		ManaCost: "{3}{B}", OracleID: championOfTheWeirdOracle, Power: 5, Toughness: 5,
		Owner: me.ID, Controller: me.ID,
	})
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, RevealIDs: reveal})
}

// Beholding a Goblin card exiles it. The activation costs 1 life and two
// -1/-1 counters on a creature you control; the target opponent then puts
// two -1/-1 counters on a creature of their choice. Targeting yourself is
// refused. When the Champion leaves, the Goblin card returns.
func TestChampionOfTheWeirdMakesAnOpponentBlightTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	goblin := handCard(me, "Boggart Card", "Creature — Goblin")
	a := pushVanillaCreature(g, opp.ID, "A", 2, 2)
	b := pushVanillaCreature(g, opp.ID, "B", 4, 4)

	champ, err := castChampionOfTheWeird(t, g, goblin)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !inExile(g, goblin) {
		t.Fatal("the beheld Goblin card was not exiled")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(champ) {
		t.Fatal("the Champion did not resolve")
	}

	err = g.ActivateCatalogAbility(me.ID, champ, 0, game.ActivateAbilityParams{Strict: true,
		Targets:   []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
		BlightIDs: []uuid.UUID{champ}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("targeting yourself: err = %v, want ErrIllegalTarget", err)
	}
	if got := minusCounters(g, champ); got != 0 {
		t.Fatalf("a refused activation put %d counters on the Champion", got)
	}

	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, champ, 0, game.ActivateAbilityParams{Strict: true,
		Targets:   []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		BlightIDs: []uuid.UUID{champ}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Life != life-1 {
		t.Errorf("life %d after activating, want %d", me.Life, life-1)
	}
	if got := minusCounters(g, champ); got != 2 {
		t.Errorf("-1/-1 counters on the Champion: %d, want 2 (blight 2 paid)", got)
	}
	monSettle(t, g)
	p := blightPrompt(g, opp.ID)
	if p == nil {
		t.Fatal("the target opponent is not asked to blight")
	}
	if p.ChooseMin != 1 || p.ChooseMax != 1 || !hasID(p.ChooseCards, a) || !hasID(p.ChooseCards, b) || hasID(p.ChooseCards, champ) {
		t.Errorf("prompt %d..%d over %v: want one of the opponent's creatures", p.ChooseMin, p.ChooseMax, p.ChooseCards)
	}
	answerChooseCards(t, g, opp.ID, b)
	monSettle(t, g)
	if got := minusCounters(g, b); got != 2 {
		t.Errorf("-1/-1 counters on the opponent's choice: %d, want 2", got)
	}
	if got := minusCounters(g, a); got != 0 {
		t.Errorf("-1/-1 counters on the other creature: %d, want 0", got)
	}

	b27Kill(g, champ)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(goblin) {
		t.Error("the exiled Goblin did not return to its owner's hand")
	}
}

// The behold is mandatory and takes a Goblin only.
func TestChampionOfTheWeirdRefusesANonGoblin(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elf := handCard(me, "Elf Card", "Creature — Elf")
	if _, err := castChampionOfTheWeird(t, g); err == nil {
		t.Error("cast without beholding a Goblin")
	}
	if _, err := castChampionOfTheWeird(t, g, elf); err == nil {
		t.Error("beheld an Elf")
	}
}
