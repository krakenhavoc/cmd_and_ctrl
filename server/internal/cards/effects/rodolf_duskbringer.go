package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rodolf Duskbringer — Legendary Creature — Vampire Angel {5}{B},
// 4/4:
//
//	"Flying, deathtouch, lifelink
//	 Whenever you gain life, Rodolf Duskbringer gains indestructible
//	 until end of turn.
//	 At the beginning of your end step, you may pay {1}{W/B}. When you
//	 do, return target creature card with mana value X or less from
//	 your graveyard to the battlefield, where X is the amount of life
//	 you gained this turn."
//
// Three printed keywords, a lifegain trigger and an end-step rebuy,
// and the third one is the reason this card is worth reading.
//
// "You may pay {1}{W/B}. When you do, …" is the MayPay / CR 603.12
// pair in its purest printed form: the payment is the whole condition
// and the return is a SECOND triggered ability created while the
// first resolves. Two things fall out of that and both are printed
// behaviour rather than engine detail:
//
//   - The target is chosen when the reflexive trigger goes on the
//     stack (CR 603.3d), which is after the mana is spent — so X is
//     read then, and a creature card put into the graveyard in
//     response to the end-step trigger is a legal choice.
//   - The reflexive trigger is answerable. An opponent gets a window
//     between the payment and the reanimation.
//
// X is the running LifeGained tally, read live by the target
// predicate, so it widens as the turn's lifegain accumulates and the
// picker never offers a card the bound does not reach. X of 0 leaves
// only zero-cost creature cards legal; with none, the reflexive
// trigger has no legal target and is removed (CR 603.3d), which is
// the printed outcome — the mana is still spent, exactly as it would
// be in paper.
//
// The indestructible grant is a Layer 6 keyword until end of turn on
// Rodolf itself. It fires once per lifegain EVENT, not once per
// point, and a Rodolf that has left the battlefield gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "26a716b2-a2e2-405b-accd-df572416d8bf",
		Name:            "Rodolf Duskbringer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch", "lifelink"},
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Rodolf Duskbringer — gains indestructible until end of turn", rodolfGainsIndestructible),
			AtYourEndStep("Rodolf Duskbringer — pay {1}{W/B} to return a creature card", rodolfEndStepOffer),
		},
	})
}

// rodolfGainsIndestructible is the lifegain trigger's body.
func rodolfGainsIndestructible(g *game.Game, item *game.StackItem) error {
	return GrantKeywordUntilEOT{
		Target:   item.SourceCardID,
		Keywords: []string{"indestructible"},
		Label:    "Rodolf Duskbringer — indestructible until end of turn",
	}.Apply(NewContext(g, item))
}

// rodolfEndStepOffer queues the optional cost. MayPay only ASKS —
// the trigger's resolution is over by the time the controller
// answers, so the "when you do" half has to live inside OnPay rather
// than after this call.
func rodolfEndStepOffer(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}{W/B}",
		Question: "Rodolf Duskbringer — pay {1}{W/B} to return a creature card from your graveyard to the battlefield?",
		OnPay:    rodolfReturnFromGraveyard,
	}.Apply(NewContext(g, item))
}

// rodolfReturnFromGraveyard is the CR 603.12 reflexive trigger. It is
// raised only from OnPay, so "when you do" really does mean the
// payment happened.
func rodolfReturnFromGraveyard(ctx *Context) error {
	return ReflexiveTrigger{
		Label: "Rodolf Duskbringer — return a creature card from your graveyard to the battlefield",
		Targets: TargetCardInGraveyard(
			"target creature card in your graveyard with mana value at most the life you gained this turn",
			YouOwn(), Creature(), ManaValueAtMostLifeGainedThisTurn()),
		Effect: returnFirstLegalGraveyardTargetToBattlefield,
	}.Apply(ctx)
}
