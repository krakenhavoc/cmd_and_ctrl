package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_ring_pool_a_test.go — ADR 0114 PR 5, pool A: the Ring cards
// whose other text is an ordinary spell, an enters or dies trigger, or
// a sibling's ability.

const (
	BilboRetiredBurglarOracle  = "2e7e3be5-04dd-4ff4-b121-3e65b3e00699"
	BirthdayEscapeOracle       = "4f958678-2a32-4a30-8fe7-47df9ada6b2f"
	BombadilsSongOracle        = "60092fb4-98db-4f2d-a5d7-24b5f11c8acc"
	ClaimThePreciousOracle     = "8b0a0991-2327-45ea-b47f-c31db367fa1d"
	DreadfulAsTheStormOracle   = "bd9f4f20-cdef-4a20-b4cb-4f25c137a786"
	EnragedHuornOracle         = "8a7dbf4e-f121-40c5-a688-3398e04c1311"
	FieryInscriptionOracle     = "111889d4-bcca-4b1f-ad48-3077e2f5136f"
	GollumsBiteOracle          = "cf0492de-cec5-4455-839f-212246b7e9ea"
	HorsesOfTheBruinenOracle   = "100d1c9e-15aa-4bab-9fb3-76c2cd6abc2f"
	InheritedEnvelopeOracle    = "2f1ae834-f6b8-458c-ac0d-3510ddb85193"
	MirrormereGuardianOracle   = "97c188c4-3f27-428e-a4f6-4213f44898fc"
	RangersFirebrandOracle     = "796d3836-d820-440c-b0ad-5907b3cfcb10"
	RelentlessRohirrimOracle   = "1212cc74-ff66-479e-8e1b-504446e1c6d9"
	RohirrimLancerOracle       = "504fac88-1674-4157-becb-0204ab8844bf"
	SamsDesperateRescueOracle  = "5a3336f1-614a-4b3f-8752-522a34d0417a"
	SoothingOfSmeagolOracle    = "35a17b42-9d3f-49ef-8218-2780c5ef5f08"
	StalwartsOfOsgiliathOracle = "d41b2a4f-e137-4e61-99eb-6261ab3edd03"
	TookReaperOracle           = "bfd8daa9-029f-41aa-a947-cdfde3f4d6b6"
	UrukHaiBerserkerOracle     = "0e353a41-b523-4c22-9bda-d10460ba99c7"
	TheBlackBreathOracle       = "c0afbb32-dc50-436a-bd24-2990c103a4f8"
)

// ringCastSettled casts a catalog spell with targets and settles the
// stack, failing the test on a ring_bearer prompt (arrange at most one
// creature).
func RingCastSettled(t *testing.T, g *game.Game, name, typeLine, oracle string, targets ...game.TargetRef) {
	t.Helper()
	CastCatalogSpell(t, g, name, typeLine, oracle, targets)
	RingSettle(t, g)
	if RingPrompt(g, g.Seats[g.Turn.ActiveSeat].ID) != nil {
		t.Fatal("setup: the tempt asked for a Ring-bearer; arrange one creature")
	}
}
