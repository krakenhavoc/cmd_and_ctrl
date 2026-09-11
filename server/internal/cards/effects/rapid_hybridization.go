package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rapid Hybridization — Instant {U} (EDHREC rank 207):
//
//	"Destroy target creature. It can't be regenerated. That
//	 creature's controller creates a 3/3 green Frog Lizard creature
//	 token."
//
// Pongify's twin — the same one-mana blue answer with a different
// token name, and the same shape: destroy first, then hand the
// VICTIM the vanilla 3/3. The controller is read BEFORE the destroy,
// for the reason beast_within.go gives: afterwards the card is in a
// graveyard and its controller field is stale for a creature that
// had changed hands.
//
// "It can't be regenerated" is a no-op because regeneration is not
// modelled — see terminate.go / pongify.go for the note to revisit.
func init() {
	Register(Spec{
		OracleID: "06692cd9-ac2f-4a32-8fd1-043ba3c0fe71",
		Name:     "Rapid Hybridization",
		Targets:  TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			controller, ok := controllerOfTarget(ctx, target)
			if !ok {
				return nil
			}
			if err := (DestroyTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			return CreateToken{
				Controller: controller,
				Template:   GreenFrogLizardToken(),
				N:          1,
			}.Apply(ctx)
		},
	})
}

// GreenFrogLizardToken is Rapid Hybridization's 3/3 green Frog
// Lizard. Lives here rather than in tokens.go, like Pongify's Ape,
// so a concurrent batch editing that file doesn't collide with this
// one; it has no other consumer.
func GreenFrogLizardToken() game.Card {
	return game.Card{
		Name:      "Frog Lizard",
		TypeLine:  "Token Creature — Frog Lizard",
		Power:     3,
		Toughness: 3,
	}
}
