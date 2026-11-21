package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0108_pr7b_cards_test.go — ADR 0108 PR 7b (#1904): the object-source
// and property families of the not-one-use source shield
// (preventFromSource), each card against the damage it must stop and
// the damage it must not.

const (
	Pr7bFendOff         = "fa8f3827-8cd9-4896-ab0c-26fecacceb40"
	Pr7bKorHaven        = "276cece9-f9f2-46e6-ae76-daddaa2fb9ab"
	Pr7bWarning         = "c5ba0f0f-65c5-4ffa-987a-f320b401ec8f"
	Pr7bRestrain        = "5aa66cb0-86b9-4e85-9a17-df78416d682d"
	Pr7bHorn            = "50a1c14a-003f-424b-bb8e-2e2d51465a90"
	Pr7bSafeguard       = "e310c3ab-a729-404d-944f-9b477258495c"
	Pr7bLadyEvangela    = "8800d672-424b-4a7b-886f-7eb9d7a56cfe"
	Pr7bResistance      = "a95a2d1c-70a9-4eb2-ae30-a07fab26a9c6"
	Pr7bSongstitcher    = "eec72bcf-ccbc-43fc-8bdd-7bf4faa1fad7"
	Pr7bSubdue          = "81bac4b8-277b-415a-9064-a80a68fd7051"
	Pr7bKryShield       = "2c4bd475-b8af-4916-b7a0-68abb8994138"
	Pr7bSoulParry       = "8b36335b-119b-4f90-a5d7-b77d8f47f55f"
	Pr7bSnare           = "55e3dfcb-0999-401a-ae59-9804e12f4cfe"
	Pr7bHallow          = "60b6a31f-043e-4de3-ad37-67eec02114f8"
	Pr7bDromoka         = "3bd8ea71-cbcc-4659-b9a8-88cf27ee12d8"
	Pr7bAzoriusPloy     = "b9b58c3a-5a76-4b6a-884b-74f54f0ca53c"
	Pr7bShieldmageElder = "943bbe5b-76cd-478d-a1fe-d4e557df3469"
	Pr7bStonewise       = "ecfa6791-938a-4159-92bf-ff2b0ce69523"
	Pr7bIgnoble         = "52a61983-53e2-440f-af03-cebe57b102f2"
	Pr7bZealot          = "4107d0e3-608b-4024-b433-40c5d8b63549"
	Pr7bFeint           = "1bb8fe05-abb3-40a8-9e80-5d99ed0e4284"
	Pr7bFallingTimber   = "4ea19457-97ef-4ac8-b67c-41be1109ca73"
	Pr7bSereneSunset    = "e2ef027d-0ef7-4c8b-a594-3918290b02d9"
	Pr7bBorosFuryShield = "20eff2ce-f26d-48a2-8a1f-435a7e2968c8"
	Pr7bChainOfSilence  = "d7b476ed-fc50-4804-8375-488a9e2ab184"
	Pr7bAjanisAid       = "b3779fbe-7701-433b-afad-bebd6096a3c8"
	Pr7bFightingChance  = "fece632b-d8d5-4d3d-a4a7-643293b67538"
	Pr7bLoafingGiant    = "038dce4c-f754-45ef-98a4-4e16f931a65c"
	Pr7bMtendaLion      = "af029853-cfb0-403b-af51-141ba02ae2e4"
	Pr7bHeroism         = "74716642-fc8c-4f62-a556-154ed0b4af0e"
	Pr7bGuardDogs       = "490af644-2e85-49c7-af63-d5f0a9babff8"
	Pr7bRadiantKavu     = "ab46b957-5206-4390-a2f6-aba2e8d1debe"
	Pr7bLuminesce       = "7b0f482d-35a3-47f1-8283-0fd47e30cc31"
	Pr7bEtherealHaze    = "0cdbd9e7-4941-46ac-99d8-ea227181bdf8"
	Pr7bChant           = "e245a736-5f65-4159-8e39-e279e1f8794f"
	Pr7bScarecrow       = "58616e15-531f-4680-8294-35ab7e962c96"
	Pr7bEerie           = "35ba0e70-da7c-4ed9-b8ec-7dd5c5ce110e"

	Pr7bLightningBolt = "4457ed35-7c10-48c8-9776-456485fdf070"
)

// pr7bCreature is pr7Creature with a keyword list and a mana cost.
func Pr7bCreature(g *game.Game, owner uuid.UUID, name string, power int, mana string, colors []string, keywords ...string) uuid.UUID {
	return ApaPush(g, owner, owner, game.Card{Name: name, TypeLine: "Creature — Test", Power: power, Toughness: 4,
		ManaCost: mana, Colors: colors, Keywords: keywords})
}

// pr7bCastNow puts a card in the active seat's hand and casts it where
// the game stands — in combat, with priority — then resolves the stack.
func Pr7bCastNow(t *testing.T, g *game.Game, name, typeLine, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: active.ID, Controller: active.ID})
	if err := g.CastSpell(active.ID, id, params); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	PassPriorityAroundTable(t, g)
	return id
}

func Pr7bTargets(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, len(ids))
	for i, id := range ids {
		out[i] = game.TargetRef{Kind: game.TargetCard, ID: id}
	}
	return out
}

// pr7bAttack declares the attackers against `defender` and locks the
// declaration in, leaving the game in the declare attackers step with
// any attack triggers resolved.
func Pr7bAttack(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	LockInAttacks(t, g)
	PassPriorityAroundTable(t, g)
}

// pr7bBlock walks to the declare blockers step, declares the blocks
// (blocker → attacker) and locks them in, leaving priority with the
// active player in that step.
func Pr7bBlock(t *testing.T, g *game.Game, blocks map[uuid.UUID]uuid.UUID) {
	t.Helper()
	AdvanceTo(t, g, game.StepDeclareBlockers)
	for b, a := range blocks {
		if err := g.DeclareBlocker(b, a); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	FinishBlockDeclarations(t, g)
	PassPriorityAroundTable(t, g)
}

// pr7bDamageStep walks into the combat damage step, where the damage is
// dealt, and runs its priority boundary.
func Pr7bDamageStep(t *testing.T, g *game.Game) {
	t.Helper()
	AdvanceTo(t, g, game.StepCombatDamage)
	g.RunStateChecksForTest()
}

func Pr7bPrompt(t *testing.T, g *game.Game, kind game.PendingChoiceKind) *game.PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == kind {
			return c
		}
	}
	t.Fatalf("no %s prompt open", kind)
	return nil
}

func Pr7bSourceShields(g *game.Game) int {
	n := 0
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == game.ModPreventFromSource {
				n++
			}
		}
	}
	return n
}

// firstPendingTarget is a trigger's open target pick, if any.
func FirstPendingTarget(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			return c
		}
	}
	return nil
}
