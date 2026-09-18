package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Earth King — Legendary Creature — Human Noble Ally {3}{G}, 2/2
// (EDHREC rank 2742):
//
//	"When The Earth King enters, create a 4/4 green Bear creature
//	 token.
//	 Whenever one or more creatures you control with power 4 or
//	 greater attack, search your library for up to that many basic
//	 land cards, put them onto the battlefield tapped, then shuffle."
//
// The Bear it brings is its own first attacker. Two triggers:
//
//   - The ETB makes the 4/4.
//   - "One or more … attack" is one trigger per combat: EventAttack
//     fires once per attacker, so the first attacker with power 4 or
//     more queues the ability and the rest of the declaration — one
//     batch (OncePerBatch, the Adeline dedup; see AGENTS.md §7) — is
//     declined. "That many" is read as the ability RESOLVES —
//     the count of attacking creatures the controller controls with
//     power 4 or more at that moment, so a pump in response widens
//     the search and a shrink narrows it. The search is "up to", so
//     the searcher may take fewer, and the lands enter tapped through
//     the search's own clause.
//
// No simplification.
const b26EarthKingSearchLabel = "The Earth King — search for basic lands"

func init() {
	Register(Spec{
		OracleID:     "d5770b0f-4493-42e9-a2d7-0e74a28da4ba",
		Name:         "The Earth King",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("The Earth King — create a 4/4 Bear", Do(CreateToken{Template: TokenCard("4/4 green Bear"), N: 1})),
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b26CreatureYouControlWithPowerAtLeastAttacked(ev, source, g, 4)
			}, b26EarthKingSearchLabel, func(g *game.Game, item *game.StackItem) error {
				n := b26AttackingCreaturesYouControlWithPowerAtLeast(g, item.Controller, 4)
				return b26SearchBasicsOntoBattlefieldTapped(g, item, n, "The Earth King — up to that many basic lands, tapped")
			})),
		},
	})
}
