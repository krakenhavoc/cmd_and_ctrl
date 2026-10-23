package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tombstone Stairwell — World Enchantment {2}{B}{B}:
//
//	"Cumulative upkeep {1}{B}
//	 At the beginning of each upkeep, if this enchantment is on the
//	 battlefield, each player creates a 2/2 black Zombie creature token
//	 with haste named Tombspawn for each creature card in their graveyard.
//	 At the beginning of each end step and when this enchantment leaves
//	 the battlefield, destroy all tokens created with this enchantment.
//	 They can't be regenerated."
//
// #1997, one of the world enchantments ADR 0109 PR 7 left out. The
// cumulative upkeep is the ordinary constructor. The upkeep trigger fires
// on every player's upkeep; its "if this enchantment is on the
// battlefield" is checked again as it resolves (CR 603.4), so a
// Stairwell sacrificed to its own cumulative upkeep first makes nothing.
// Each player, in APNAP order, creates one Tombspawn per creature card
// in their graveyard as it resolves, under their own control.
//
// "Tokens created with this enchantment" is the token's own link to the
// Stairwell OBJECT (game.Card.CreatedWith, CR 607.1d), stamped as the
// token is made (game.TokenGroup.CreatedWith). So a Tombspawn whose
// control changed is still destroyed, a token copy of one is not, and a
// Stairwell that left and came back is a new object whose end-step
// trigger leaves the old one's Tombspawn alone (CR 400.7). The leaves
// trigger reads the object that left from its own item, so it still
// finds them. The destruction ignores regeneration (CR 701.19c).
//
// A world permanent: the world rule (CR 704.5k, ADR 0109 §8) puts it
// into its owner's graveyard when a newer one enters, and its leaves
// trigger then destroys its Tombspawn.
//
// No simplifications.
const tombstoneStairwellToken = "2/2 black Zombie with haste named Tombspawn"

func init() {
	Register(Spec{
		OracleID:     "9aaf1149-eccb-434a-a1a7-14b93fa50141",
		Name:         "Tombstone Stairwell",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			CumulativeUpkeep("Tombstone Stairwell — cumulative upkeep {1}{B}", "{1}{B}"),
			AtEachUpkeep("Tombstone Stairwell — each player creates a Tombspawn for each creature card in their graveyard",
				tombstoneStairwellUpkeep),
			OnAny([]game.EventKind{game.EventBeginEndStep, game.EventLTB},
				func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					if ev.Kind == game.EventLTB {
						return Self(ev, source, lki, g)
					}
					return true
				},
				"Tombstone Stairwell — destroy all tokens created with it; they can't be regenerated",
				tombstoneStairwellDestroy),
		},
	})
}

// tombstoneStairwellUpkeep is the upkeep trigger's resolution: the
// intervening "if" re-checked (the same object still on the
// battlefield), then each player's Tombspawn in APNAP order.
func tombstoneStairwellUpkeep(g *game.Game, item *game.StackItem) error {
	if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
		return nil
	}
	ref, ok := NewContext(g, item).SourceRef()
	if !ok {
		return nil
	}
	return tombstoneStairwellCreate(g, item, game.PermissionCardRef(ref), apnapPlayers(g))
}

// tombstoneStairwellCreate creates the first player's Tombspawn and
// moves on to the next player once they have landed, so a CR 616 order
// one creation asks for is answered before the next is made (CR 608.2c).
func tombstoneStairwellCreate(g *game.Game, item *game.StackItem, link game.PermissionCardRef, order []uuid.UUID) error {
	for len(order) > 0 {
		player := order[0]
		order = order[1:]
		n := len(graveyardIDs(g, player, func(c game.Card) bool { return c.IsCreature() }))
		if n == 0 {
			continue
		}
		rest := order
		ctx := NewContext(g, item)
		return g.CreateTokensThenForEffect(game.TokenCreation{
			Controller: player,
			Source:     item.SourceCardID,
			Groups:     []game.TokenGroup{{Template: ctx.TextToken(TokenCard(tombstoneStairwellToken)), Count: n, CreatedWith: link}},
		}, func(g *game.Game, _ []uuid.UUID) error {
			return tombstoneStairwellCreate(g, item, link, rest)
		})
	}
	return nil
}

// tombstoneStairwellDestroy destroys every token linked to the source
// OBJECT the item names: the Stairwell on the battlefield for the end
// step, the one that left for the leaves trigger.
func tombstoneStairwellDestroy(g *game.Game, item *game.StackItem) error {
	ref, ok := NewContext(g, item).SourceRef()
	if !ok {
		return nil
	}
	ids := g.TokensCreatedWithForEffect(game.PermissionCardRef(ref))
	if len(ids) == 0 {
		return nil
	}
	g.DestroyPermanentsForEffect(ids, game.DestroyOptions{CantBeRegenerated: true})
	return nil
}
