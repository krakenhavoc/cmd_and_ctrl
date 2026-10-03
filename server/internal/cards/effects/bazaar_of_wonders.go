package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bazaar of Wonders — World Enchantment {3}{U}{U}:
//
//	"When this enchantment enters, exile all graveyards.
//	 Whenever a player casts a spell, counter it if a card with the
//	 same name is in a graveyard or a nontoken permanent with the same
//	 name is on the battlefield."
//
// The entry trigger exiles every seat's graveyard (Bojuka Bog's
// helper, once per seat). The cast trigger watches every player's
// spells and asks the question as it resolves, not as it triggers —
// the "if" is the effect's, not an intervening-if — so a matching card
// put into a graveyard in response still counters the spell.
//
// "The same name" is CR 201.2's: a card in a graveyard is matched
// through game.CardNameMatches, so a split card has both its names; a
// permanent has only the name of the face that is up (a transformed
// permanent is not its front face); a face-down spell or permanent has
// none (CR 708.2a); and a token is not a nontoken permanent. The
// Bazaar itself is a nontoken permanent on the battlefield, so a
// second Bazaar cast while it is there is countered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7fca65b8-01fe-4858-97c6-f97ae82cd801",
		Name:         "Bazaar of Wonders",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Bazaar of Wonders — exile all graveyards", b02ExileAllGraveyards),
			On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID != uuid.Nil
			}, "Bazaar of Wonders — counter it if a card or nontoken permanent shares its name", bazaarOfWondersCounter),
		},
	})
}

func bazaarOfWondersCounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	spellID := ctx.Trigger().Event.CardID
	spell, ok := g.LookupCardForEffect(spellID)
	if !ok || spell.FaceDown || g.StackItemForEffect(spellID) == nil {
		return nil
	}
	if !bazaarNameIsTaken(g, spell.Name) {
		return nil
	}
	return CounterTarget{StackID: spellID}.Apply(ctx)
}

// bazaarNameIsTaken reports whether a card named `name` is in any
// graveyard, or a nontoken, face-up permanent named it is on the
// battlefield.
func bazaarNameIsTaken(g *game.Game, name string) bool {
	for _, p := range g.Seats {
		if p == nil || p.Graveyard == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if game.CardNameMatches(c, name) {
				return true
			}
		}
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if !c.IsToken() && !c.FaceDownIsPermanent() && sameFaceUpName(c, name) {
			return true
		}
	}
	return false
}

// sameFaceUpName reports whether a permanent's face-up name is `name`
// (CR 201.2), case- and space-insensitively.
func sameFaceUpName(c game.Card, name string) bool {
	want := strings.TrimSpace(name)
	return want != "" && strings.EqualFold(strings.TrimSpace(c.Name), want)
}
