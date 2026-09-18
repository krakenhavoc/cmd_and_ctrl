package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Horizon Explorer — Creature — Insect Scout {2}{G}, 2/4 (EDHREC rank
// 1811):
//
//	"Lands you control enter untapped.
//	 Whenever you attack a player, create a Lander token. (It's an
//	 artifact with "{2}, {T}, Sacrifice this token: Search your
//	 library for a basic land card, put it onto the battlefield
//	 tapped, then shuffle.")"
//
// The landfall deck's Amulet on a body. The first line is the
// b16LandsYouControlEnterUntapped replacement: it clears the
// enters-tapped flag once something has set it, which is how the
// CR 616 order comes out as printed without an ordering prompt on
// every tapland. The attack trigger is the "one or more … A PLAYER"
// guard: one declaration is one batch, so however many creatures
// attack one player they make one Lander, and attacking a SECOND
// player is a second occurrence and a second Lander — "Horizon
// Explorer's last ability will trigger once for each player you
// attack" (ruling), which is CR 603.2c and the engine's
// OncePerBatchPerPlayer (#784). The Lander is b16LanderToken, whose
// search ability rides the template.
//
// One declared simplification, weaker than printed: a land fetched
// from a library (or returned from a graveyard) that carries an
// enters-tapped effect of its own — a fetched Guildgate — still
// enters tapped. Those entry sites cannot pause for the ordering
// prompt two applicable effects would raise, and an unanswerable
// prompt strands the land (see the helper). A fetched basic under a
// "tapped" fetch enters untapped, as printed.
func init() {
	Register(Spec{
		OracleID:     "e8d20361-d9b7-4f9c-8ec5-3ac7c460dcb2",
		Name:         "Horizon Explorer",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"A land fetched from your library that has its own \"enters tapped\" still enters tapped.",
		},
		Replacements: []game.ReplacementEffect{b16LandsYouControlEnterUntapped()},
		Triggered: []game.TriggeredAbility{
			OncePerBatchPerPlayer(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b16YouAttackedAPlayer(ev, source, g)
			}, "Horizon Explorer — create a Lander token", Do(CreateToken{Template: b16LanderToken(), N: 1}))),
		},
	})
}
