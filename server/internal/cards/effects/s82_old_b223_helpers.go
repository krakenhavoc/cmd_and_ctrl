package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b223_helpers.go — shared bodies for Ixalan, batch 223 (#5532): the
// raid triggers of the Pirates and enchantments.

// OldB223RaidAtYourEndStep is "Raid — At the beginning of your end step, if you
// attacked this turn, <effect>": an intervening "if" (CR 603.4) read from the
// turn's attack record as the step begins and again as the ability resolves.
func OldB223RaidAtYourEndStep(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventBeginEndStep, AllOf(ByYou, TmcB6YouAttackedThisTurn), label,
		Do(If{Cond: AvatarB11CondYouAttacked{}, Then: []Applier{runEffect{fn: effect}}}))
}

// OldB223RaidWhenThisEnters is "Raid — When this creature enters, if you
// attacked this turn, <effect>", an intervening "if" read as it enters and
// again as it resolves.
func OldB223RaidWhenThisEnters(label string, effect Effect) game.TriggeredAbility {
	return InterveningIf(AvatarB11CondYouAttacked{}, On(game.EventETB, AllOf(Self, TmcB6YouAttackedThisTurn), label, effect))
}

// OldB223MayLoot is "you may draw a card. If you do, discard a card" (Sage of
// the Falls's loot): the discard happens only if the draw was taken.
func OldB223MayLoot(question string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return MayChoice{
			Question: question,
			Draws:    1,
			YesLabel: "Draw",
			NoLabel:  "Decline",
			Then:     CelestusAnswer,
		}.Apply(NewContext(g, item))
	}
}
