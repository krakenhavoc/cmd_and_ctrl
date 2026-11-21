package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hashaton_test.go — the Hashaton, Scarab's Fist decklist batch: the
// commander's token-copy engine, the reanimation package, and the
// conditional enters-tapped land cycles.
//
// Two assertions in here are the point of the whole batch:
//
//   - Reanimating from an OPPONENT's graveyard has to put the
//     creature under the CASTER's control. It used to route to the
//     graveyard's owner, which handed the creature straight back.
//   - A conditional enters-tapped land has to ENTER tapped, not
//     enter untapped and get tapped. Both leave Tapped == true, so
//     the discriminator is the tap event count, exactly as in
//     temples_test.go.

const (
	HashatonOracle        = "db266661-f783-4907-9e52-6963eec05431"
	ReanimateOracle       = "a044474a-cd72-4e9d-bd8d-a08f2de9cdc0"
	ZombifyOracle         = "bb95db4d-5017-4121-bf79-d68476602d8c"
	LateToDinnerOracle    = "31bf199b-dfb1-428e-96a5-eb25104e2b43"
	DarkslickShoresOracle = "a2b48695-f7d7-42ce-a8a0-2a723428542a"
	GlacialFortressOracle = "027dd013-baa7-4111-b3c9-f4d1414e9c45"
	DesertedBeachOracle   = "f0ec8681-da50-466b-8cdd-1dc710deccd9"
	TestOgreOracle        = "00000000-0000-4000-8000-00000000ffff"
)

// seedGraveyardCreature puts a creature card with a real mana cost
// into a player's graveyard, so a reanimation target exists and its
// mana value is something other than zero.
func SeedGraveyardCreature(p *game.Player, name, manaCost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Creature — Zombie Giant",
		ManaCost:   manaCost,
		Power:      3,
		Toughness:  3,
		Owner:      p.ID,
		Controller: p.ID,
	})
	return id
}

// seedLandOnBattlefield drops a land straight onto the battlefield
// without going through the play path — for setting up the board
// state a conditional enters-tapped land reads.
func SeedLandOnBattlefield(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

// emptyHandToLibrary clears a player's hand so a random discard is
// deterministic. The tests below are about what the discard
// triggers, not about which card is picked.
func EmptyHandToLibrary(g *game.Game, p *game.Player) {
	g.WithWriteLock(func() {
		for p.Hand.Size() > 0 {
			c, err := p.Hand.Top()
			if err != nil {
				return
			}
			_, _ = game.MoveCard(p.Hand, p.Library, c.InstanceID)
		}
	})
}

// reanimatedTypeChoiceFor returns the queued creature-type prompt
// addressed to a player, or nil.
func ReanimatedTypeChoiceFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceCreatureType && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// --- Hashaton ----------------------------------------------------

// hashatonDiscardOgre seeds Hashaton on the battlefield under `me`,
// empties the hand, puts one distinctive creature card in it and
// discards it. Returns the discarded card's ID.
//
// The ogre's printed values are all deliberately unlike the token's:
// 2/5, red, legendary, with subtypes. Every one of them is something
// the "except it's a 4/4 black Zombie" clause has to overwrite or
// keep, so the assertions can tell a real copy from a template.
func HashatonDiscardOgre(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	PushCatalogPermanent(g, me.ID, "Hashaton, Scarab's Fist",
		"Legendary Creature — Zombie Wizard", HashatonOracle, false)
	EmptyHandToLibrary(g, me)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       "Test Ogre",
		TypeLine:   "Legendary Creature — Ogre Warrior",
		OracleID:   TestOgreOracle,
		ManaCost:   "{2}{R}",
		Power:      2,
		Toughness:  5,
		Colors:     []string{"R"},
		Owner:      me.ID,
		Controller: me.ID,
	})
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	PassPriorityAroundTable(t, g)
	return id
}
