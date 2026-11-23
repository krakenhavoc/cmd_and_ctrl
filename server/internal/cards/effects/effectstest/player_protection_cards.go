package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// player_protection_cards_test.go — #1197's four cards, each proved
// through the engine rather than by reading its own Spec back.
//
// The assertions are deliberately about what a PLAYER at the table
// would notice: an opponent's Lightning Bolt is refused at announce,
// the life total does not move, the shield is gone on your next turn.
// A test that only checked `spec.PlayerKeywords` would pass with
// every choke point backed out.

const (
	LeylineOfSanctityOracle = "492e0e6c-8c27-4376-938b-f8a8b6205810"
	AegisOfTheGodsOracle    = "c5bfc1b9-a55d-4608-a6f7-bb62cb8dc3c6"
	TeferisProtectionOracle = "0d4ecdb1-ec90-497f-a7a4-1c68092b8757"
	AbsoluteVirtueOracle    = "d4fddf20-6b3a-42c3-a245-002dacbb4725"
)

// playerAbilities reads a seat's abilities the way every consumer
// does, under the read lock.
func PlayerAbilities(g *game.Game, p *game.Player) []string {
	var out []string
	g.ReadSnapshot(func() { out = g.PlayerAbilitiesForEffect(p) })
	return out
}

func HasPlayerAbility(list []string, want string) bool {
	for _, a := range list {
		if a == want {
			return true
		}
	}
	return false
}

// boltAtPlayerErr is b08CastBolt's announce half: the cast that is
// SUPPOSED to be refused hands its error back.
func BoltAtPlayerErr(t *testing.T, g *game.Game, caster *game.Player, target uuid.UUID) error {
	t.Helper()
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Lightning Bolt", TypeLine: "Instant",
		OracleID: LightningBoltOracle, Colors: []string{"R"}, ManaCost: "{R}",
		Owner: caster.ID, Controller: caster.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return g.CastSpell(caster.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: target}},
	})
}

// advanceOneTurnForTest walks the cursor until the active seat
// changes. The effects package has no turn helper of its own; this is
// duration_test.go's, locally.
func AdvanceOneTurnForTest(t *testing.T, g *game.Game) {
	t.Helper()
	start := g.Turn.ActiveSeat
	for i := 0; i < 64; i++ {
		if g.Turn.ActiveSeat != start {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatalf("cursor never left seat %d's turn", start)
}
