package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ace, Fearless Rebel — Legendary Creature — Human Rebel {3}{G}, 2/2 (Doctor Who):
//
//	"Nitro-9 — Whenever Ace attacks, you may sacrifice an artifact. When you
//	 do, put a +1/+1 counter on Ace, then it fights up to one target creature
//	 defending player controls.
//	 Doctor's companion"
//
// Bolg of the North's and Armix's shape: the attack trigger asks "you may
// sacrifice an artifact" as it resolves (CR 608.2c) and an artifact that really
// went raises the printed "when you do" as a CR 603.12 reflexive trigger. That
// trigger chooses its own "up to one target creature defending player
// controls" as it goes on the stack (the defending player rides its params),
// puts the counter on Ace, then Ace fights the target if one was chosen and is
// still legal. If Ace has left the battlefield by then, neither half happens.
// Doctor's companion is a deck-construction rule read off the oracle text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "50d0d870-fe3d-44a2-ad6e-93307a1fb468",
		Name:         "Ace, Fearless Rebel",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Nitro-9 — you may sacrifice an artifact", func(g *game.Game, item *game.StackItem) error {
				return effects.MaySacrificeOne(g, item, "Ace — sacrifice an artifact to put a +1/+1 counter on Ace and have it fight up to one target creature defending player controls?",
					effects.Artifact(), false, aceFearlessRebelSacrificed)
			}),
		},
	})
}

// aceFearlessRebelSacrificed raises the reflexive trigger once an artifact has
// really gone, keyed (#4290).
var aceFearlessRebelSacrificed = effects.RegisterSacrificedOne("prompt-run/ace-fearless-rebel-sacrifice", func(ctx *effects.Context, _ uuid.UUID) error {
	return effects.ReflexiveTrigger{
		Label:  "Nitro-9 — put a +1/+1 counter on Ace, then it fights up to one target creature defending player controls",
		Body:   aceFearlessRebelFightBody,
		Params: game.EffectParams{Player: effects.B17DefendingPlayer(ctx.Game, ctx.Trigger().Event)},
	}.Apply(ctx)
})

var aceFearlessRebelFightBody = game.ReflexiveBody("ace-fearless-rebel/counter-and-fight",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		if !effects.B09SourceStillOnBattlefield(g, item) {
			return nil
		}
		if err := effects.PlusOneCountersOnThis(1)(g, item); err != nil {
			return err
		}
		return effects.B30SourceFightsFirstLegalTarget(g, item)
	}),
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return effects.TargetCreature("up to one target creature defending player controls", effects.ControlledBy(p.Player)).WithCount(0, 1)
	})
