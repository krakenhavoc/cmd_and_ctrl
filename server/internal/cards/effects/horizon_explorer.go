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
// every tapland. The attack trigger is Adeline's "whenever you
// attack" dedup (one trigger per declaration, however many
// attackers), and the Lander is b16LanderToken, whose search
// ability rides the template.
//
// Two declared simplifications, both weaker than printed:
//   - A land fetched from a library (or returned from a graveyard)
//     that carries an enters-tapped effect of its own — a fetched
//     Guildgate — still enters tapped. Those entry sites cannot pause
//     for the ordering prompt two applicable effects would raise, and
//     an unanswerable prompt strands the land (see the helper). A
//     fetched basic under a "tapped" fetch enters untapped, as
//     printed.
//   - "Whenever you attack a player" is once per declaration:
//     attacking two players at once makes one Lander, not two.
func init() {
	Register(Spec{
		OracleID:     "e8d20361-d9b7-4f9c-8ec5-3ac7c460dcb2",
		Name:         "Horizon Explorer",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"A land fetched from your library that has its own \"enters tapped\" still enters tapped.",
			"Attacking two players at once makes one Lander, not two.",
		},
		Replacements: []game.ReplacementEffect{b16LandsYouControlEnterUntapped()},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b16YouAttackedAPlayer(ev, source, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Horizon Explorer — create a Lander token",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{Controller: item.Controller, Template: b16LanderToken(), N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
