package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_b03_helpers.go — card-side helpers for Commander-deck wave 2,
// batch 3 (#5368).

// Token rows added from this file's init so the shared table's column
// alignment is not rewritten.
func init() {
	for key, card := range map[string]game.Card{
		// Sandwurm Convergence (the printed Wurm of Duskmourn Commander).
		"5/5 green Wurm": {Name: "Wurm", TypeLine: "Token Creature — Wurm", Power: 5, Toughness: 5, Colors: []string{"G"}},
	} {
		if _, dup := tokenTable[key]; dup {
			panic("effects: duplicate token row " + key)
		}
		tokenTable[key] = card
	}
}

// cmdr2B03CreaturesMatchingCantAttackYouOrPlaneswalkers is "<creatures
// matching the queries> can't attack you or planeswalkers you control"
// (CR 508.1c): CreaturesMatchingCantAttackYou's wider sibling, covering
// the controller's planeswalkers (and battles) as well as the player.
func cmdr2B03CreaturesMatchingCantAttackYouOrPlaneswalkers(label string, attackers ...game.PermanentQuery) game.StaticAbility {
	return game.StaticAbility{
		Label: label,
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
			return target.IsCreature()
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			if source == nil {
				return
			}
			c.AttackTargetRestrictions = append(c.AttackTargetRestrictions, game.AttackTargetRestriction{
				Source:                source.InstanceID,
				SourceName:            source.Name,
				NotPlayer:             source.Controller,
				OnlyAttackersMatching: append([]game.PermanentQuery(nil), attackers...),
			})
		},
	}
}

// cmdr2B03AttackingYouOrYourPlaneswalker is "creature that's attacking you
// or a planeswalker you control" (Soul Snare): AttackingYou's wider
// sibling. The attacked object is the caster as a player, or a
// planeswalker on the battlefield the caster controls (CR 506.3).
func cmdr2B03AttackingYouOrYourPlaneswalker() CardPredicate {
	return func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		return attacksYouOrYourPlaneswalker(g, c.AttackingTarget, caster)
	}
}

// cmdr2B03DuringSourcesControllersTurn passes while it is the turn of the
// modifier source's controller: "during your turn" on a static ability.
func cmdr2B03DuringSourcesControllersTurn() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Game != nil && IsYourTurn(q.Game, q.Source.Controller)
	}
}

// cmdr2B03NotAManaAbility passes on an activation that is not a mana
// ability (CR 605.1a): "unless they're mana abilities".
func cmdr2B03NotAManaAbility() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Ability != nil && !q.Ability.Mana
	}
}

// --- creature-type reveals (Descendants' Fury, Heirloom Blade, Kindred Summons) ---

// cmdr2B03AllTypesWord is the tail word of a creature that is every creature
// type (a changeling, CR 702.73a); any other creature's word is its creature
// types joined by "|".
const cmdr2B03AllTypesWord = "*"

// cmdr2B03TypesWord encodes the creature types a creature had as one word a
// keyed tail can carry: "reveal cards until a creature card that shares a
// creature type with the sacrificed creature" reads the sacrificed creature
// after it has gone.
func cmdr2B03TypesWord(subtypes []string, everyType bool) string {
	if everyType {
		return cmdr2B03AllTypesWord
	}
	var types []string
	for _, s := range subtypes {
		if game.IsCreatureType(s) {
			types = append(types, s)
		}
	}
	return strings.Join(types, "|")
}

// cmdr2B03SharesTypeWith is "a creature card that shares a creature type
// with" the creature a cmdr2B03TypesWord describes. An empty word (a
// creature with no creature type) is shared with nothing, and a changeling
// shares with every creature that has any creature type (CR 702.73a).
func cmdr2B03SharesTypeWith(word string) func(game.Card) bool {
	return func(c game.Card) bool {
		if word == "" || !c.IsCreature() {
			return false
		}
		if word == cmdr2B03AllTypesWord {
			return len(game.CreatureTypesOf(&c)) > 0
		}
		for _, t := range strings.Split(word, "|") {
			if c.HasSubtype(t) {
				return true
			}
		}
		return false
	}
}

// cmdr2B03RevealUntilSharingTypeToHand is "reveal cards from the top of your
// library until you reveal a creature card that shares a creature type with
// <it>. Put that card into your hand and the rest on the bottom of your
// library in a random order" (Heirloom Blade); the battlefield half of the
// sentence (Descendants' Fury) is RevealUntilThenPutOntoBattlefield.
func cmdr2B03RevealUntilSharingTypeToHand(ctx *Context, word, reason string) error {
	player := ctx.Controller()
	run, hit := revealUntil(ctx, player, cmdr2B03SharesTypeWith(word), reason)
	return TakeFromLibraryToHand{
		Player: player,
		Cards:  run,
		Match:  func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
		All:    true,
		Key:    takeRestBottomRandom,
	}.Apply(ctx)
}

// cmdr2B03CreaturesYouControlOfType counts the creatures `player` controls
// that have creature type t (a changeling has every one).
func cmdr2B03CreaturesYouControlOfType(g *game.Game, player uuid.UUID, t string) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() && c.HasSubtype(t) {
			n++
		}
	}
	return n
}

// cmdr2B03ShuffleRestIntoLibraryTail is "then shuffle the rest of the revealed
// cards into your library": the rest are still in it, so shuffling is
// putting them in. An on-disk key.
var cmdr2B03ShuffleRestIntoLibraryTail = PutFromLibraryTail("cmdr2-b03/shuffle-rest-into-library",
	func(ctx *Context, res PutFromLibraryResult) error {
		return ctx.Game.ShuffleLibraryForEffect(res.Player)
	})
