package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Moonbase — Plane — Moon (Doctor Who, #4870):
//
//	"Low Gravity — All creatures have "{2}: This creature gains flying
//	 until end of turn. Activate only as a sorcery."
//	 Whenever chaos ensues, for each opponent, choose up to one target
//	 creature card in that player's graveyard. Put those cards onto the
//	 battlefield face down under your control. They're 2/2 Cyberman
//	 artifact creatures."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6). The
// static is a layer-6 grant of a catalog bundle (ADR 0093) to every creature,
// each controller activating its own. The chaos ability (CR 311.7) has one
// "up to one" target clause per opponent, bound to that opponent's graveyard
// (the target clauses of perOpponentClauses), and puts each card that is
// still a legal target as it resolves onto the battlefield face down under
// the planar controller's control as a nameless 2/2 Cyberman artifact
// creature (CR 708.2, CybermanBody). Face down, nothing about the card
// triggers as it enters (CR 708.2a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "381f510b-8696-4af1-a6c9-d1175e9f36d0",
		Name:         "The Moonbase",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: theMoonbaseGrant,
			Activated: []ActivatedAbility{{
				Label:        "{2}: This creature gains flying until end of turn. Activate only as a sorcery.",
				Cost:         ManaCost("{2}"),
				SorcerySpeed: true,
				Effect:       thisGainsKeywordUntilEndOfTurn("flying", "The Moonbase — flying until end of turn"),
			}},
			Text: "{2}: This creature gains flying until end of turn. Activate only as a sorcery.",
		}},
		Static: []game.StaticAbility{GrantAbilities(allCreatures, theMoonbaseGrant)},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverChaosEnsues("The Moonbase — for each opponent, put up to one target creature card from their graveyard onto the battlefield face down under your control as a 2/2 Cyberman artifact creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if _, err := g.ReturnFromGraveyardFaceDownForEffect(t.ID, item.Controller, false, CybermanBody()); err != nil {
							if err = nothingIfGone(err); err != nil {
								return err
							}
						}
					}
					return nil
				}),
				planesB2PerOpponentGraveyardClauses("creature card", Creature())),
		},
	})
}

// theMoonbaseGrant is the granted bundle's key. An on-disk key.
const theMoonbaseGrant = "the-moonbase/low-gravity"
