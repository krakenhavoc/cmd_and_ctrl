package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alien Invasion — Enchantment {1}{R}{R}{G} (Marvel Super Heroes):
//
//	"At the beginning of combat on your turn, create a 1/1 red Alien
//	 creature token with haste and "This token attacks each combat if
//	 able." Put a +1/+1 counter on it for each invasion counter on this
//	 enchantment, then put an invasion counter on this enchantment."
//
// The Alien is a catalog token (alienToken): the haste is a printed keyword
// and the attack requirement is the token's own static ability, so a restore
// keeps both. The counters are a keyed tail that runs once the token has
// really landed (a creation can pause on a CR 616 prompt): it reads the
// invasion counters on the enchantment as they are then, puts that many
// +1/+1 counters on the new token in one placement, and only afterwards puts
// the next invasion counter on the enchantment, so the first Alien is a 1/1,
// the second a 2/2, the third a 3/3. The invasion counter is put on even if
// the token somehow did not arrive, as printed ("then" is not "if you do").
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7c854a07-e031-458e-a335-9c31865020b8",
		Name:         "Alien Invasion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Alien Invasion — create a 1/1 Alien, grown by the invasion counters, then add an invasion counter", alienInvasionBegin),
		},
	})
}

// alienInvasionCounter is the counter kind Alien Invasion counts.
const alienInvasionCounter = "invasion"

// alienInvasionBegin is the trigger's resolution.
func alienInvasionBegin(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups:     []game.TokenGroup{{Template: ctx.TextToken(alienToken()), Count: 1}},
	}, ctx.ItemTail(alienInvasionGrow))
}

// alienInvasionGrow is "put a +1/+1 counter on it for each invasion counter
// on this enchantment, then put an invasion counter on this enchantment" as
// a keyed tail. An on-disk key.
var alienInvasionGrow = CardTail("alien-invasion/grow", func(ctx *Context, o game.TailOutcome) error {
	self, ok := ctx.Game.LookupCardForEffect(ctx.Source())
	if !ok {
		return nil
	}
	if n := self.Counters[alienInvasionCounter]; n > 0 && len(o.IDs) > 0 {
		if err := (AddCounter{Target: o.IDs[0], Kind: game.CounterPlusOne, N: n}).Apply(ctx); err != nil {
			return err
		}
	}
	return AddCounter{Target: ctx.Source(), Kind: alienInvasionCounter, N: 1}.Apply(ctx)
})

// alienToken is the 1/1 red Alien creature token with haste and "This token
// attacks each combat if able."
func alienToken() game.Card { return tokenFromCatalog(printedAlienToken) }

// printedAlienToken is the Alien as PRINTED.
func printedAlienToken() tokenTemplate {
	return tokenTemplate{
		Slug: "alien",
		Card: game.Card{
			Name:      "Alien",
			TypeLine:  "Token Creature — Alien",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"R"},
			Keywords:  []string{"haste"},
		},
		Static: []game.StaticAbility{AttacksEachCombat()},
		Text:   "Haste\nThis token attacks each combat if able.",
	}
}
