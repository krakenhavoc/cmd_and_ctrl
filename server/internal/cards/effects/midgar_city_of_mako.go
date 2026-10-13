package effects

import (
	"github.com/google/uuid"

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
const midgarOracleID = "4e34a49d-f031-48ac-a458-97b79124b76c"

func init() {
	Register(Spec{
		OracleID:     midgarOracleID,
		Name:         "Midgar, City of Mako",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
	})
	Register(Spec{
		OracleID:     midgarOracleID + "#1",
		Name:         "Reactor Raid",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if !controlsAnArtifactOrCreature(ctx.Game, item.Controller) {
				return nil
			}
			return MayChoice{
				Question: "Reactor Raid — sacrifice an artifact or creature to draw two cards?",
				YesLabel: "Sacrifice",
				NoLabel:  "Don't",
				OnYes:    reactorRaidSacrifice,
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
func reactorRaidSacrifice(ctx *Context) error {
	item, me := ctx.Item, ctx.Controller()
	return ctx.Game.PlayersSacrificeThenForEffect(item.SourceCardID, []uuid.UUID{me},
		sacrificeSpec("an artifact or creature", Or(Artifact(), Creature())),
		"Reactor Raid — sacrifice an artifact or creature",
		func(g *game.Game, sacrificed game.PromptedSacrifices) error {
			if !sacrificed.Sacrificed(me) {
				return nil
			}
			return DrawCards{Player: me, N: 2}.Apply(NewContext(g, item))
		})
}
