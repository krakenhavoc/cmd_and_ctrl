package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reg Barclay, Fidgety Fantasist — Legendary Creature — Human Artificer
// {2}{U}{R}, 1/5 (Star Trek Commander, #4974):
//
//	"Seek out — Whenever you draw your second card each turn, create a
//	 token that's a copy of another target creature you control, except
//	 it isn't legendary, it's an Illusion artifact in addition to its
//	 other types, and it has haste. Exile that token at the beginning of
//	 the next end step."
//
// "Seek out" is an ability word (CR 207.2c); the trigger is Atlantean
// Cavalry's second-card read of the turn's draw tally. The target is
// another creature you control, chosen as the trigger goes on the stack
// (CR 603.3d). The three exceptions are copiable values of the token
// (CR 707.9b): no Legendary supertype, Artifact and Illusion added, and
// haste. The exile is a delayed trigger on that token alone (CR 603.7).
// A target gone by resolution makes no token (CR 608.2b).
//
// No simplification.
func init() {
	seek := On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return b41DrewYourSecondCardThisTurn(ev, source, g)
	}, "Reg Barclay, Fidgety Fantasist — create a hasty Illusion artifact token copy of another target creature you control",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return trekB40CreateCopyExiledAtNextEndStep(ctx, item.Controller, id, trekB40IllusionArtifactExcept,
				"Reg Barclay, Fidgety Fantasist — exile the token")
		})
	seek.Targets = Another(TargetCreature("another target creature you control", YouControl()))
	Register(Spec{
		OracleID:     "adcf571f-c465-44c6-8941-48313810e732",
		Name:         "Reg Barclay, Fidgety Fantasist",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{seek},
	})
}
