package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cosmic Cube — {5} Artifact (Marvel Super Heroes, #4455):
//
//	"Ward {2}
//	 Whenever you attack, look at the top six cards of your library. You
//	 may cast a spell from among them with mana value less than or equal
//	 to the greatest power among attacking creatures you control without
//	 paying its mana cost. Put the rest on the bottom of your library in
//	 a random order."
//
// Aetherworks Marvel's look-and-cast (free_library_cast.go) with its
// eligibility narrowed to the mana value cap. "Whenever you attack" is
// one trigger per declaration (OncePerBatch); the cap is read as the
// trigger resolves, from the creatures that are attacking then
// (CR 608.2h: the greatest power is the one of the creatures now in
// combat; with none left the cap is 0, so only a mana value 0 spell
// qualifies). A land is not a spell and is never offered. The cap is the
// spell's, judged on the face being cast (#4509): a modal DFC whose back
// face costs more is cast only as its front.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0866a630-d962-4aa9-a15f-08b2947540ef",
		Name:         "Cosmic Cube",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Cosmic Cube — ward {2}"),
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Cosmic Cube — look at the top six cards and cast one free", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				limit := cosmicCubeGreatestAttackingPower(g, ctx.Controller())
				return lookAtTopAndOfferFreeCast(g, ctx.Controller(), ctx.Source(), cosmicCubeFreeCast, 6,
					spellWithManaValueAtMost(limit))
			})),
		},
	})
}

var cosmicCubeFreeCast = registerFreeLibraryCastOffer("choose-cards/cosmic-cube-cast", "Cosmic Cube")

// cosmicCubeGreatestAttackingPower is the greatest power among attacking
// creatures `you` control, post-layer; 0 with none.
func cosmicCubeGreatestAttackingPower(g *game.Game, you uuid.UUID) int {
	best := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != you || !c.IsCreature() || c.AttackingTarget == uuid.Nil {
			continue
		}
		if p := c.CurrentPower(); p > best {
			best = p
		}
	}
	return best
}
