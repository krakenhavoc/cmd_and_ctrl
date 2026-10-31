package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crimson Cowl, Master of Evil — Legendary Creature — Human Villain
// {2}{B}, 1/3 (Marvel Super Heroes Commander, #4455):
//
//	"Whenever one or more nontoken Villains you control attack a player,
//	 you create a 2/1 black Villain creature token with menace."
//
// Attuma's "attack a player" guard (CR 603.2c): one declaration collapses
// to a single token per player attacked, however many nontoken Villains go
// at that player, and a second player attacked is a second token. Crimson
// Cowl is a nontoken Villain himself, so he counts. A token Villain
// attacking alone does not trigger it. The token is the shared "2/1 black
// Villain with menace" row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2962cfc6-6557-4aee-a95a-8ca3bb06eb29",
		Name:         "Crimson Cowl, Master of Evil",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatchPerPlayer(On(game.EventAttack, msc5NontokenVillainAttackAPlayer,
				"Crimson Cowl, Master of Evil — create a 2/1 black Villain creature token with menace",
				Do(CreateToken{Template: TokenCard("2/1 black Villain with menace"), N: 1}))),
		},
	})
}

// msc5NontokenVillainAttackAPlayer is the When of "whenever one or more
// nontoken Villains you control attack a player".
func msc5NontokenVillainAttackAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !b16YouAttackedAPlayer(ev, source, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && !c.IsToken() && c.HasSubtype(source.TextSubtype("Villain"))
}
