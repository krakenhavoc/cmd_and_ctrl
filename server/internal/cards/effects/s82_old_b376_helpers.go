package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b376_helpers.go — shared by Born of the Gods batch 376 (#5559).

// SameNameAsCardExiledWithThis is the clause "target spell with the same name as a card
// exiled with this creature" (Mindreaver, CR 607.2a): the spell's name must be the name of
// a card in exile linked to the clause's own source object. See
// game.TargetSpec.SameNameAsExiledWithSource.
func SameNameAsCardExiledWithThis(spec *game.TargetSpec) *game.TargetSpec {
	spec.SameNameAsExiledWithSource = true
	return spec
}

// ExileTopCardsOfTargetPlayerLinked is "exile the top N cards of target player's library"
// linked to the resolving ability's source (CR 607.2a): the first still-legal player
// target's top N, as one exile. A library with fewer cards exiles what it has.
func ExileTopCardsOfTargetPlayerLinked(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetPlayer {
				continue
			}
			p := g.PlayerByIDForEffect(t.ID)
			if p == nil || p.Library == nil {
				return nil
			}
			var ids []uuid.UUID
			for i := 0; i < n && i < len(p.Library.Cards); i++ {
				// The library's top is the LAST element.
				ids = append(ids, p.Library.Cards[len(p.Library.Cards)-1-i].InstanceID)
			}
			if len(ids) == 0 {
				return nil
			}
			link, _ := ctx.SourceRef()
			return g.ExileLinkedKeyedForEffect(link, ids, game.KeyedTail{})
		}
		return nil
	}
}

// CantBlockUnlessYouControlAnother is "This creature can't block unless you control another
// <subtype>" (Felhide Brawler, CR 509.1b): the blocker is the permanent that prints the rule,
// and the rule bars it unless its controller controls a different permanent with the type.
// The subtype is read through the source's text (CR 612.2).
func CantBlockUnlessYouControlAnother(subtype, label string) game.BlockRule {
	self := OnSelf()
	return game.BlockRule{
		Reason: game.BlockReasonCantBlockAttacker,
		Label:  label,
		Pair: func(g *game.Game, _, blocker, source *game.Card) bool {
			if !self(g, blocker, source) {
				return false
			}
			want := source.TextSubtype(subtype)
			for i := range g.Battlefield.Cards {
				c := &g.Battlefield.Cards[i]
				if c.InstanceID != source.InstanceID && c.Controller == source.Controller && c.HasSubtype(want) {
					return false
				}
			}
			return true
		},
	}
}

// PowerLessThanIslandsYouControl is the forbidden-blocker test of Kraken of the Straits:
// "creatures with power less than the number of Islands you control", "you" being the
// attacker's controller. Counted off the battlefield directly: a block check may hold only the
// read lock.
func PowerLessThanIslandsYouControl() BlockerAgainstAttacker {
	return func(g *game.Game, attacker, blocker *game.Card) bool {
		want := attacker.TextSubtype("Island")
		n := 0
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.Controller == attacker.Controller && c.HasSubtype(want) {
				n++
			}
		}
		return blocker.PowerForComparison() < n
	}
}
