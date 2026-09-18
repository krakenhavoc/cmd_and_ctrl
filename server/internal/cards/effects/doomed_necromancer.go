package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Doomed Necromancer — 2/2 Creature — Human Cleric Mercenary for
// {2}{B} (EDHREC rank 3982):
//
//	"{B}, {T}, Sacrifice this creature: Return target creature card
//	 from your graveyard to the battlefield."
//
// A three-mana body that turns into a Zombify on a later turn, which
// is why every reanimator deck that can find creatures plays it over
// the sorcery: it is tutorable with a creature tutor, it can be
// blinked, and the sacrifice is a cost, so the reanimation happens
// even if someone answers the Necromancer in response.
//
// It is in the batch as the three-component activation cost — mana,
// tap, and sacrifice-this all on one ability — which nothing else in
// the reanimation family has. All three are paid on activation, so:
//
//   - Summoning sickness applies, because of the {T} (CR 302.6). The
//     Necromancer does nothing the turn it lands.
//   - The Necromancer is already in the graveyard when the ability
//     resolves, and it is not a legal target for its own ability —
//     the target is chosen at activation, when the Necromancer is
//     still on the battlefield.
//   - Killing the Necromancer in response does nothing: the cost has
//     been paid and the ability is on the stack independently.
//
// "From YOUR graveyard" is the narrow half of the family (Zombify's
// clause, not Reanimate's), so the creature returns to its owner —
// who is the activator — and the ability can never take a creature
// out of an opponent's pile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "155422a0-a0cd-4399-8ed9-fa68ac2c80a6",
		Name:         "Doomed Necromancer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{B}, {T}, Sacrifice this creature: Return target creature card from your graveyard to the battlefield.",
			Cost:    Plus(ManaCost("{B}"), TapCost(), SacrificeThis()),
			Targets: targetCreatureInYourGraveyard(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				_, _ = reanimateSingleTarget(ctx, item.Controller)
				return nil
			},
		}},
	})
}
