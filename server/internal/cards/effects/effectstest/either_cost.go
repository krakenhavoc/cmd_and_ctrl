package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// either_cost_test.go — ADR 0100 sub-PR 3: either/or additional costs
// (CR 601.2b, 601.2f–h) and the discard record (owner decision 6). The
// plan arithmetic is tested next to the code in game/; these are the
// board-level rules — what the announcement must name, what each branch
// pays, and what the resolution reads back — and one test per card.

const (
	DemandAnswersOracle      = "c11e84a1-dbda-429b-8cd6-fd0deaefc689"
	BoneShardsOracle         = "6e760cfe-45b9-4a3f-b2e2-4ca6ec2bd13a"
	LightningAxeOracle       = "81b90905-fbc0-426a-a084-c3300533abb4"
	GrabThePrizeOracle       = "a37387f0-9d57-49ff-aa38-2c6b359e8916"
	BitterTriumphOracle      = "776341cb-d2ec-423f-9250-92dc8bd8d503"
	EatenAliveOracle         = "d437ecc2-2fd3-4ad3-b23e-217f55e58dae"
	SparkHarvestOracle       = "25756e33-25e5-4976-8cce-a39fa2c4b007"
	LashOfTheBalrogOracle    = "66499407-3ad2-464b-bbfe-95d860f53616"
	AnnihilatingGlareOracle  = "a75087aa-41bd-4e0f-a118-61b5209356f5"
	DeadlyPrecisionOracle    = "d23ca3d2-0731-4df4-a3c6-69ddf0c58601"
	StirUpTroubleOracle      = "dda607bd-f419-4b7f-b052-a5ce6ce22bfe"
	FinalPaymentOracle       = "e47d57dc-2e69-4939-8aec-077595f2ae05"
	PumpkinBombardmentOracle = "37827e84-9f5f-49ed-b939-0cf80dac82e7"
	BogslithersEmbraceOracle = "21695c70-f746-4d77-b6d3-e2713a5c929e"
	WildUnravelingOracle     = "dd1a4219-437b-494f-ae0e-1d3181ffcd63"
	MorkrutBehemothOracle    = "76c2571d-7d45-4c7a-99d1-302e2b26f8f6"
	BayouGroffOracle         = "e83b2790-3cec-429c-a65d-6f4c7d025d37"
	MinionMissileOracle      = "670cf656-5c72-4d28-b5e6-dc77781dc001"
	LouisoixsSacrificeOracle = "42d8ab49-703a-4fd3-84d0-0a3ac1eafca8"
	RedirectLightningOracle  = "c62b5c22-e058-436f-9575-a01cb2112829"
	LethalThrowdownOracle    = "4ab565bb-3188-4d23-93d8-5de67fc0d056"
	SoulsOfTheLostOracle     = "c05d5a51-8e39-4475-9743-1adc60283c5a"
)

func Branch(i int) *int { return &i }

// eitherPermanent puts a permanent the active seat controls onto the
// battlefield and returns its ID.
func EitherPermanent(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}
