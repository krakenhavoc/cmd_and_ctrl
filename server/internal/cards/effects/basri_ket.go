package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Basri Ket — Legendary Planeswalker — Basri {1}{W}{W}, loyalty 3
// (Jumpstart Selesnya deck request, #4893):
//
//	"+1: Put a +1/+1 counter on up to one target creature. It gains
//	 indestructible until end of turn.
//	 −2: Whenever one or more nontoken creatures attack this turn, create
//	 that many 1/1 white Soldier creature tokens that are tapped and
//	 attacking.
//	 −6: You get an emblem with "At the beginning of combat on your turn,
//	 create a 1/1 white Soldier creature token, then put a +1/+1 counter on
//	 each creature you control.""
//
// +1 is the counter, then the until-end-of-turn indestructible, on the one
// target if it is still legal; with no target chosen it does nothing.
//
// −2 creates a delayed triggered ability (CR 603.7) that repeats for the
// rest of the turn and fires once per attack declaration that includes a
// nontoken creature (DelayedOnEvent{Repeats, OncePerBatch}, Dalkovan
// Encampment's shape). It is not tied to Basri, so it still fires if Basri
// has left. "That many" is the nontoken creatures declared in that batch,
// read off the event log as the trigger resolves, so one removed from
// combat in response still counts. The Soldiers enter tapped and attacking
// and their controller chooses what each attacks (CR 508.4); they were
// never declared, so they fire nothing that watches attacks, this trigger
// included.
//
// −6's emblem makes the Soldier first, then counts the creatures you
// control, so the new Soldier gets a counter too.
//
// No simplification.
func init() {
	const emblemText = "At the beginning of combat on your turn, create a 1/1 white Soldier creature token, then put a +1/+1 counter on each creature you control."
	Register(Spec{
		OracleID:        "497d7ab8-07f1-4c8a-a45a-52b9a4d9400f",
		Name:            "Basri Ket",
		Completeness:    CompletenessFull,
		StartingLoyalty: 3,
		Emblem: &EmblemSpec{
			Label: "Basri Ket emblem",
			Text:  emblemText,
			Triggered: []game.TriggeredAbility{
				AtBeginningOfYourCombat("Basri Ket emblem — create a Soldier, then a +1/+1 counter on each creature you control",
					Do(CreateToken{Template: TokenCard("1/1 white Soldier"), N: 1},
						Instruction(putAPlusOneCounterOnEachCreatureYouControlStep))),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Put a +1/+1 counter on up to one target creature. It gains indestructible until end of turn.",
				Cost:    LoyaltyCost(1),
				Targets: TargetCreature("up to one target creature").WithCount(0, 1),
				Effect:  Do(plusOneCounterThenKeywordOnTheTarget("indestructible", "Basri Ket — indestructible until end of turn")...),
			},
			{
				Label: "−2: Whenever one or more nontoken creatures attack this turn, create that many 1/1 white Soldier creature tokens that are tapped and attacking.",
				Cost:  LoyaltyCost(-2),
				Effect: Do(DelayedOnEvent{
					Label:        "Basri Ket — create that many 1/1 white Soldiers, tapped and attacking",
					On:           []game.EventKind{game.EventAttack},
					Condition:    nontokenCreaturesAttackCondition,
					Body:         basriKetSoldiersBody,
					Repeats:      true,
					OncePerBatch: true,
				}),
			},
			{
				Label:  "−6: You get an emblem with \"" + emblemText + "\"",
				Cost:   LoyaltyCost(-6),
				Effect: Do(CreateEmblem{}),
			},
		},
	})
}
