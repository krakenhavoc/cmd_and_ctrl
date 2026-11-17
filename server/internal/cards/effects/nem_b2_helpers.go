package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// nem_b2_helpers.go — the shapes Nemesis batch 2 (#5618) prints more than
// once: the Percher block rule, the spellshaper cost, the Mercenary,
// Rebel and tribal recruiters, the Silkenfist untap, "if you control a
// Swamp, you may sacrifice a creature rather than pay this spell's mana
// cost", and "each player loses N life". Its own file, as #231 asks of a
// batch, so a concurrent batch never collides here.
//
// Append-only: add a builder, never change what one means.

// nemBlocksOnlyFlyers is "This creature can block only creatures with
// flying." (Belbe's Percher, Battlefield Percher, Stronghold Zeppelin) on
// the permanent that prints it, or "Enchanted creature can block only
// creatures with flying." (Air Bladder) with OnAttached. It is Cloud
// Spirit's block rule, a restriction judged as blockers are declared
// (CR 509.1b).
func nemBlocksOnlyFlyers(scope BlockScope) []game.BlockRule {
	return []game.BlockRule{
		CantBlockAttackers(scope, OnMatching(WithoutKeyword("flying")), "it can block only creatures with flying"),
	}
}

// nemSpellshaperCost is a Nemesis Spellshaper's cost, "<mana>, {T},
// Discard a card": the card is discarded as the ability is activated
// (CR 602.2b), and the {T} is a creature's, so the Spellshaper waits out
// summoning sickness (CR 302.6).
func nemSpellshaperCost(mana string) game.AbilityCost {
	return Plus(ManaCost(mana), TapCost(), DiscardACard())
}

// nemRecruiterRow is the Nemesis recruiters' "<cost>, {T}: Search your
// library for a <subtype> permanent card [with mana value N or less], put
// it onto the battlefield, then shuffle." With maxMV < 0 the clause has
// no mana value bound (Moggcatcher, Seahunter, Skyshroud Poacher);
// otherwise it is the Ramosian recruiters' body at N (Defiant Falcon,
// Rathi Fiend, Rathi Intimidator). The search names a quality, so the
// searcher may find nothing (CR 701.23b), and the library is shuffled
// either way.
func nemRecruiterRow(name, cost, article, subtype string, maxMV int) ActivatedAbility {
	what := article + " " + subtype + " permanent card"
	effect := nemSearchForASubtypePermanent(subtype, name+" — "+what+", onto the battlefield")
	if maxMV >= 0 {
		what += " with mana value " + strconv.Itoa(maxMV) + " or less"
		effect = searchForASubtypePermanentWithManaValueAtMost(subtype, maxMV, name+" — "+what+", onto the battlefield")
	}
	return ActivatedAbility{
		Label:   cost + ", {T}: Search your library for " + what + ", put it onto the battlefield, then shuffle.",
		Purpose: game.Purpose{Answers: game.AnswerValue},
		Cost:    Plus(ManaCost(cost), TapCost()),
		Effect:  effect,
	}
}

// nemSearchForASubtypePermanent is "Search your library for a <subtype>
// permanent card, put it onto the battlefield, then shuffle" with no
// mana value bound: a permanent card (CR 110.4) of that subtype, read
// through the lens the search lends the source's text (HasSubtypeWord),
// so a changeling card counts.
func nemSearchForASubtypePermanent(subtype, reason string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return SearchLibrary{
			Player:    item.Controller,
			Predicate: func(c game.Card) bool { return isPermanentCard(c) && c.HasSubtypeWord(subtype) },
			Dest:      game.ZoneBattlefield,
			Limit:     1,
			Shuffle:   true,
			Reason:    reason,
			Source:    item.SourceCardID,
		}.Apply(NewContext(g, item))
	}
}

// nemUntapWhenBlocked is the Silkenfist Fighter and Silkenfist Order
// trigger, "Whenever this creature becomes blocked, untap it." It
// triggers once however many creatures block (CR 509.3c), and untapping
// does not remove the creature from combat (CR 506.4 lists what does), so
// it stays attacking and blocked.
func nemUntapWhenBlocked(name string) game.TriggeredAbility {
	return On(game.EventBecomesBlocked, Self, name+" — untap it",
		func(g *game.Game, item *game.StackItem) error {
			return UntapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
		})
}

// nemSacrificeACreatureWithASwamp is Dark Triumph's and Mind Swords'
// "If you control a Swamp, you may sacrifice a creature rather than pay
// this spell's mana cost." The Swamp is checked as the spell is cast
// (CR 601.2b); the creature is sacrificed as a cost with the spell
// already on the stack (CR 601.2h), so it stays dead if the spell is
// countered.
func nemSacrificeACreatureWithASwamp() game.AlternativeCost {
	alt := SacrificeInstead(1, "a creature", 0, Creature())
	alt.Label = "Sacrifice a creature (you control a Swamp)"
	alt.Condition = ControlsA("Swamp")
	return alt
}

// nemEachPlayerLoses is "each player loses N life": every live player,
// the controller included, in seat order. Life loss, not damage.
func nemEachPlayerLoses(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, id := range tablePlayers(ctx) {
			if err := g.ChangePlayerLifeForEffect(ctx.Source(), id, -n); err != nil {
				return err
			}
		}
		return nil
	}
}

// nemMayTapOrUntapTheTarget is "You may tap or untap target <permanent>."
// as an activated ability's body (Trickster Mage): the controller picks
// tap it, untap it or leave it as the ability resolves, through Tidal
// Bore's and Rejoinder's prompt. A target that left is illegal (CR
// 608.2b), and nothing is asked.
func nemMayTapOrUntapTheTarget(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return socB23MayTapOrUntap(ctx, name, id)
	}
}

// nemTargetRestrictedThisTurn is "Target creature can't attack this
// turn" (Netter en-Dal) and "can't attack or block this turn" (Off
// Balance's spell body goes through the same restriction) as an
// activated ability's body: the restriction ends in the cleanup step
// (CR 514.2).
func nemTargetRestrictedThisTurn(r game.Restriction, label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return RestrictUntilEOT{Target: id, Restrictions: r, Label: label}.Apply(ctx)
	}
}

// tapAllOtherCreatures is "tap all other creatures" (Timbermare,
// Shrieking Mogg): every creature on the battlefield but the source, as
// the ability resolves.
func tapAllOtherCreatures(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range MatchingBattlefield(ctx, And(Creature(), NotSelf(item.SourceCardID))) {
		ids = append(ids, c.InstanceID)
	}
	return tapEach(ctx, ids)
}
