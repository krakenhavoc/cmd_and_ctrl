package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Barrensteppe Siege — Enchantment {2}{W}{B}:
//
//	"As this enchantment enters, choose Abzan or Mardu.
//	 • Abzan — At the beginning of your end step, put a +1/+1 counter
//	   on each creature you control.
//	 • Mardu — At the beginning of your end step, if a creature died
//	   under your control this turn, each opponent sacrifices a
//	   creature of their choice."
//
// A #1572 anchor-word card: each bullet is an end-step trigger gated
// on its word (ADR 0071).
//
// Mardu's "if" is an intervening if (CR 603.4): checked as the trigger
// would fire and again as it resolves, against the per-turn tally of
// creatures that died under the controller's control (tokens count,
// CR 700.4). The sacrifice is EachPlayerSacrifices: each opponent
// picks their own creature, and it is neither targeted nor optional.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "228d3307-69a9-45ce-a610-77754ed50877",
		Name:         "Barrensteppe Siege",
		Completeness: CompletenessFull,
		AsEnters:     ChooseOptionAsEnters("Barrensteppe Siege", "Abzan", "Mardu"),
		Triggered: []game.TriggeredAbility{
			WhenChosen("Abzan", AtYourEndStep("Barrensteppe Siege — put a +1/+1 counter on each creature you control",
				b11PutCounterOnEachCreatureYouControl)),
			WhenChosen("Mardu", On(game.EventBeginEndStep,
				AllOf(ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return aCreatureDiedUnderYourControlThisTurn(g, source.Controller)
				}),
				"Barrensteppe Siege — each opponent sacrifices a creature", barrensteppeSiegeEdict)),
		},
	})
}

// aCreatureDiedUnderYourControlThisTurn is Mardu's intervening if.
func aCreatureDiedUnderYourControlThisTurn(g *game.Game, controller uuid.UUID) bool {
	return g.TurnTallyFor(controller).CreaturesDied > 0
}

// barrensteppeSiegeEdict is the Mardu body, the "if" re-checked first.
func barrensteppeSiegeEdict(g *game.Game, item *game.StackItem) error {
	if !aCreatureDiedUnderYourControlThisTurn(g, item.Controller) {
		return nil
	}
	return EachPlayerSacrifices{ExceptController: true, Match: Creature(), Label: "a creature"}.Apply(NewContext(g, item))
}
