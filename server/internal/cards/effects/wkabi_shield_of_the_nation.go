package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// W'Kabi, Shield of the Nation — Legendary Creature — Human Warrior Hero
// {1}{G}, 2/3 (Marvel Super Heroes Commander, #4455):
//
//	"Vigilance
//	 Whenever you attack with your commander, if you control an artifact
//	 with mana value 4 or greater, create a 4/4 green Rhino creature token
//	 with trample."
//
// Jocasta's "whenever you attack with your commander" (the declaration is
// yours and a commander you control is among the attackers), collapsed to
// one trigger per declaration (CR 603.2c). The "if" is an intervening if
// (CR 603.4): in the trigger condition, so a board without the artifact
// never triggers, and again as the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "66950068-adf4-42d7-ad72-fdb4a0e58922",
		Name:            "W'Kabi, Shield of the Nation",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(InterveningIf(CondBoard{
				Name: "you control an artifact with mana value 4 or greater",
				Test: func(ctx *Context) bool { return wkabiControlsABigArtifact(ctx.Game, ctx.Controller()) },
			}, On(game.EventAttack, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return msc13AttackedWithYourCommander(ev, source, ch, g) && wkabiControlsABigArtifact(g, source.Controller)
			}, "W'Kabi, Shield of the Nation — create a 4/4 green Rhino creature token with trample",
				Do(CreateToken{Template: TokenCard("4/4 green Rhino with trample"), N: 1})))),
		},
	})
}

// wkabiControlsABigArtifact is "you control an artifact with mana value 4 or
// greater".
func wkabiControlsABigArtifact(g *game.Game, player uuid.UUID) bool {
	return countControlled(g, player, func(c game.Card) bool {
		if !c.IsArtifact() {
			return false
		}
		mv, ok := g.ManaValueForEffect(c)
		return ok && mv >= 4
	}) > 0
}
