package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const mirrorformOracle = "81cb888c-c4d3-45a2-a03e-bfd91a2af2be"

func TestMirrorformIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(mirrorformOracle)
	if !ok || spec.Name != "Mirrorform" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// mirrorformCast casts Mirrorform strictly, paying {4}{U}{U} from the pool.
func mirrorformCast(t *testing.T, g *game.Game, target uuid.UUID) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Mirrorform", TypeLine: "Instant", OracleID: mirrorformOracle,
		ManaCost: "{4}{U}{U}", Owner: me.ID, Controller: me.ID})
	advanceTo(t, g, game.StepPrecombatMain)
	b06AddMana(me, "U", "U", "C", "C", "C", "C")
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}})
}

// Every nonland permanent the caster controls becomes a lasting copy of
// the target; lands and the opponent's permanents are left alone.
func TestMirrorformTurnsEveryNonlandPermanentYouControlIntoTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	model := seedCopyableCreature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	theirs := seedCopyableCreature(g, opp.ID, "Isamaru", "Legendary Creature — Dog", 2, 2)
	bear := seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	rock := dcSeed(g, me.ID, "Mind Stone", "Artifact", "oracle-mind-stone", 0, 0)
	land := dcSeed(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)

	if err := mirrorformCast(t, g, model); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	dcCleanup(g)
	for _, id := range []uuid.UUID{bear, rock} {
		if c := dcCard(t, g, id); c.Name != "Hill Giant" || !c.IsCreature() || c.Effective().Power != 3 {
			t.Errorf("%s is %q, want a lasting Hill Giant", id, c.Name)
		}
	}
	if c := dcCard(t, g, land); c.Name != "Forest" || c.IsCopy() {
		t.Errorf("the land became %q", c.Name)
	}
	if c := dcCard(t, g, theirs); c.Name != "Isamaru" {
		t.Errorf("the opponent's creature became %q", c.Name)
	}
}

// An Aura is not a legal target.
func TestMirrorformRefusesAnAura(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	aura := dcSeed(g, me.ID, "Pacifism", "Enchantment — Aura", "oracle-pacifism", 0, 0)
	if err := mirrorformCast(t, g, aura); err == nil {
		t.Fatal("Mirrorform targeted an Aura")
	}
}
