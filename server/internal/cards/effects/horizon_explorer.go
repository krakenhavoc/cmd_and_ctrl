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
// A land FETCHED tapped that also carries its own enters-tapped
// clause — a fetched Guildgate — is the one entry that raises the
// CR 616.1 ordering prompt, because the search seeds the fetching
// effect's flag onto the event and both effects are applicable at
// once. That used to be a declared simplification: the entry site
// could not pause, so the helper left such a land alone and it
// entered tapped. #478 made the search entry resumable and #732 took
// the retreat out; the prompt is answerable now and the land is as
// printed.
func init() {
	Register(Spec{
		OracleID:     "e8d20361-d9b7-4f9c-8ec5-3ac7c460dcb2",
		Name:         "Horizon Explorer",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{b16LandsYouControlEnterUntapped()},
		Triggered: []game.TriggeredAbility{
			OncePerBatchPerPlayer(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b16YouAttackedAPlayer(ev, source, g)
			}, "Horizon Explorer — create a Lander token", Do(CreateToken{Template: b16LanderToken(), N: 1}))),
		},
	})
}
