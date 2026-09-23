package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rabble Rousing — Enchantment {4}{W}:
//
//	"Hideaway 5 (When this enchantment enters, look at the top five
//	 cards of your library, exile one face down, then put the rest on
//	 the bottom in a random order.)
//	 Whenever you attack with one or more creatures, create that many
//	 1/1 green and white Citizen creature tokens. Then if you control
//	 ten or more creatures, you may play the exiled card without paying
//	 its mana cost."
//
// The card #1331 was filed for (deck tracker #1306, "Aang is so
// flashy"), and the first hideaway card in the catalog (ADR 0091).
//
// The attack half is one trigger per attack DECLARATION, however many
// creatures it names: OncePerBatch on EventAttack, keyed on the label.
// "That many" is counted in Build, at the first event of the batch,
// when the whole declaration is already staged (every attacker carries
// its AttackingTarget — see b41AnotherLegendaryCreatureYouControlIsAttacking)
// — so a three-creature swing makes three Citizens and a creature that
// dies before the trigger resolves still counted.
//
// "Then if you control ten or more creatures" is asked AFTER the
// tokens land — the Citizens it just made count toward the ten — which
// is why the check is the token creation's continuation rather than
// the next statement: creation can pause on a CR 616 prompt, and a
// check written after it would read the board before the tokens.
//
// The free play is PlayHiddenCard over the card THIS Rabble Rousing
// hid, captured as the object in Build (CR 607.2a). A grant rather than
// an inline play — ADR 0091's declared deviation, ADR 0066's posture:
// the player plays the card with an ordinary action once the trigger
// has resolved, until end of turn; a land hidden here waits for a main
// phase and the turn's land drop, which the attack step never has.
//
// Declared simplification in ADR 0091, not a caveat: the grant-not-
// inline window every "you may cast it" in the engine shares.
const rabbleRousingOracleID = "2e420001-2b3d-4eb7-b01c-3a49d5e241c2"

const rabbleRousingLabel = "Rabble Rousing — create that many Citizens"

func init() {
	Register(Spec{
		OracleID:     rabbleRousingOracleID,
		Name:         "Rabble Rousing",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Hideaway("Rabble Rousing", 5),
			OncePerBatch(game.TriggeredAbility{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				},
				Key: rabbleRousingLabel,
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					ref := game.ObjectRefOf(*source)
					n := len(b13AttackingCreaturesYouControl(g, source.Controller))
					return game.NewTriggeredItem(source, rabbleRousingLabel, func(g *game.Game, item *game.StackItem) error {
						return rabbleRousingResolve(g, item, ref, n)
					})
				},
			}),
		},
	})
}

// rabbleRousingResolve makes the Citizens, then — with them on the
// battlefield — offers the hidden card if the controller has ten or
// more creatures.
func rabbleRousingResolve(g *game.Game, item *game.StackItem, ref game.PermissionCardRef, n int) error {
	controller := item.Controller
	offer := func(g *game.Game, _ []uuid.UUID) error {
		if b14CreaturesControlled(g, controller) >= 10 {
			grantHiddenPlay(g, controller, ref, "Rabble Rousing — play the exiled card without paying its mana cost")
		}
		return nil
	}
	if n <= 0 {
		return offer(g, nil)
	}
	return g.CreateTokensThenForEffect(game.TokenCreation{
		Controller: controller,
		Source:     item.SourceCardID,
		Groups: []game.TokenGroup{{
			Template: TokenCard("1/1 green and white Citizen"),
			Count:    n,
		}},
	}, offer)
}
