package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Incubation // Incongruity — split card, oracle 363edb16-009f-4cc2-99e3-2b15db84e7af:
//
//	Incubation {G/U} Sorcery: "Look at the top five cards of your library.
//	 You may reveal a creature card from among them and put it into your
//	 hand. Put the rest on the bottom of your library in a random order."
//	Incongruity {1}{G}{U} Instant: "Exile target creature. That creature's
//	 controller creates a 3/3 green Frog Lizard creature token."
//
// Two specs, Incongruity under "<oracle_id>#1" as Connive // Concoct is.
// Incubation is Commune with Nature's dig with Monumental Henge's random
// bottom order. Incongruity is Pongify's shape with exile: the target's
// controller is read before the exile and is handed the token. A target
// that is gone or illegal means the spell does not resolve (CR 608.2b).
func init() {
	const oracle = "363edb16-009f-4cc2-99e3-2b15db84e7af"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Incubation",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 5),
				Match:    effects.Creature(),
				Max:      1,
				Optional: true,
				Label:    "Incubation — reveal a creature card and put it into your hand",
				Key:      effects.TakeRevealRestBottomRandom,
			}.Apply(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Incongruity",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.RemovalReturning(0, game.TargetReturn{CreatureTokens: 1, TokenPower: 3, TokenToughness: 3})), Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			controller, ok := effects.ControllerOfTarget(ctx, target)
			if !ok {
				return nil
			}
			if err := (effects.ExileTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateToken{Controller: controller, Template: effects.TokenCard("3/3 green Frog Lizard"), N: 1}.Apply(ctx)
		},
	})
}
