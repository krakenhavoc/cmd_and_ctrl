package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blitz_cards_test.go — the blitz creatures (CR 702.152, #3772): every
// card offers the right blitz price, and each card's own text does what
// it prints.

const (
	TenaciousUnderdogOracle  = "598e3ed2-659f-4400-b41d-3aee8068fba6"
	RequisitionerOracle      = "469f25de-c180-47e9-9b1b-e23269139bbd"
	GirderGoonsOracle        = "655961a3-94b5-4600-b9f1-f93163d03a49"
	WorkshopWarchiefOracle   = "0695af85-7405-4512-bb97-42517a532ada"
	CaldaiaStrongarmOracle   = "e0e6938b-7abf-4f69-a3d6-82a58912bff4"
	MayhemPatrolOracle       = "f77a9160-1d37-476d-8908-29ddcc982328"
	PlasmaJockeyOracle       = "dedf0d3a-090a-4cc3-b5ee-65b6cb990b72"
	NightClubberOracle       = "4d2abfe1-f5f8-4ed5-915d-bfe98c6321ee"
	RiveteersDecoyOracle     = "1c4918e6-3810-458e-a11b-949bbc43e7cc"
	MezzioMuggerOracle       = "068699a9-dd5c-428b-96eb-fd085dda71e9"
	PugnaciousPugilistOracle = "7c22e6ba-5570-460b-962e-b7c0fbb3f394"
	CaldaiaGuardianOracle    = "fd55d686-510e-4b6c-b8eb-458ee7c19fff"
	WaveOfRatsOracle         = "ccb0f59b-7947-4563-9e16-5d23246f4232"
	StarAthleteOracle        = "fb03befb-b4e4-4ae9-908d-a16d5b03b054"
	JaxisOracle              = "764121bf-535d-40d0-ab9d-65486da31734"
)

// attackWithTrigger declares `attacker` at `defender`, locks the attack
// in, answers the trigger's target (when `target` is set) and resolves.
func AttackWithTrigger(t *testing.T, g *game.Game, attacker, defender, target uuid.UUID) {
	t.Helper()
	AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	LockInAttacks(t, g)
	if target != uuid.Nil {
		AnswerPickTarget(t, g, target)
	}
	PassPriorityAroundTable(t, g)
}
