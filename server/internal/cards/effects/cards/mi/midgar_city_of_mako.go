package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Midgar, City of Mako // Reactor Raid — Land — Town, with a Sorcery —
// Adventure half {2}{B} (oracle 4e34a49d…, #3213, the Kuja deck):
//
//	Midgar, City of Mako
//	  "This land enters tapped.
//	   {T}: Add {B}."
//	Reactor Raid
//	  "You may sacrifice an artifact or creature. If you do, draw two
//	   cards. (Then exile this card. You may play the land later from
//	   exile.)"
//
// Jidoor's shape (a Town land with an Adventure half; the engine plays
// the land from exile afterwards, spending the land drop). Reactor
// Raid's "you may ... if you do" is a yes/no first and the prompted
// sacrifice second, with the draw as the sacrifice's continuation so it
// pays out only for a permanent that really left the battlefield. With
// no artifact or creature to give up the question is not asked and
// nothing is drawn.
//
// No simplification.
const MidgarOracleID = "4e34a49d-f031-48ac-a458-97b79124b76c"

func init() {
	effects.Register(effects.Spec{
		OracleID:     MidgarOracleID,
		Name:         "Midgar, City of Mako",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     MidgarOracleID + "#1",
		Name:         "Reactor Raid",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if !controlsAnArtifactOrCreature(ctx.Game, item.Controller) {
				return nil
			}
			return effects.MayChoice{
				Question: "Reactor Raid — sacrifice an artifact or creature to draw two cards?",
				YesLabel: "Sacrifice",
				NoLabel:  "Don't",
				Then:     reactorRaidAnswer,
			}.Apply(ctx)
		},
	})
}

func controlsAnArtifactOrCreature(g *game.Game, player uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && (c.IsArtifact() || c.IsCreature()) {
			return true
		}
	}
	return false
}

// reactorRaidSacrifice is the "if you do" branch.
func reactorRaidSacrifice(ctx *effects.Context) error {
	me := ctx.Controller()
	return ctx.PlayersSacrificeThen([]uuid.UUID{me},
		effects.SacrificeSpec("an artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		"Reactor Raid — sacrifice an artifact or creature",
		reactorRaidDraw, ctx.RunData([]uuid.UUID{me}))
}

// reactorRaidDraw is "if you do, draw two cards" once the sacrifice has
// landed, keyed (#4290): the controller is the run's carry.
var reactorRaidDraw = effects.RunThen("prompt-run/reactor-raid-draw", func(ctx *effects.Context, r game.RunDone) error {
	if len(r.Carry) == 0 || !r.Sacrifices().Sacrificed(r.Carry[0]) {
		return nil
	}
	return effects.DrawCards{Player: r.Carry[0], N: 2}.Apply(ctx)
})

// reactorRaidAnswer is the "yes": sacrifice an artifact or creature, then
// draw two cards. Keyed (#4291).
var reactorRaidAnswer = effects.MayYes("confirm/reactor-raid-sacrifice", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	return reactorRaidSacrifice(ctx)
})
