package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// borrowed_abilities.go — the card side of ADR 0155 (#1557): "has all
// activated abilities of …". The engine half, the layer-6 borrow step and
// the readers, is server/internal/game/borrowed_abilities.go.
//
// A borrow is a layer-6 static that declares game.StaticAbility.Borrows.
// Build one with HasActivatedAbilitiesOf ("~ has all activated abilities
// of …") or GrantActivatedAbilitiesOf ("[recipients] have all activated
// abilities of …"); the lender sets (battlefield, graveyard, exiled-with,
// library top, chosen permanent) arrive with the cards that use them.

// HasActivatedAbilitiesOf is "~ has all activated abilities of …": a
// layer-6 borrow whose one recipient is its own source.
func HasActivatedAbilitiesOf(b game.AbilityBorrow) game.StaticAbility {
	s := GrantActivatedAbilitiesOf(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return source != nil && target.InstanceID == source.InstanceID
	}, b)
	// The printed clause is the row's player-facing wording.
	s.Label = b.Label
	return s
}

// GrantActivatedAbilitiesOf is "[recipients] have all activated abilities
// of …" (Hazel's Brewmaster's Foods, Agatha's Soul Cauldron's creatures):
// a layer-6 borrow over the ordinary StaticAbility predicate.
func GrantActivatedAbilitiesOf(appliesTo func(target *game.Card, g *game.Game, source *game.Card) bool, b game.AbilityBorrow) game.StaticAbility {
	borrow := b
	return game.StaticAbility{
		Layer:     game.Layer6Ability,
		AppliesTo: appliesTo,
		Borrows:   &borrow,
		Label:     b.Label,
	}
}

// LenderPredicate picks the lenders of a borrow. It is a CardPredicate that
// also sees the borrowing source, for the clauses that compare with it
// ("that don't have the same name as this creature"). `source.Controller`
// is "you".
type LenderPredicate func(g *game.Game, source *game.Card, c game.Card) bool

// ByController lifts a CardPredicate (YouControl(), Creature() …) to a
// LenderPredicate whose caster is the borrowing source's controller.
func ByController(p CardPredicate) LenderPredicate {
	return func(g *game.Game, source *game.Card, c game.Card) bool {
		return p(g, source.Controller, game.ThroughTextOf(source, c))
	}
}

// BattlefieldLenders is "… the activated abilities of [permanents matching
// P]": every permanent on the battlefield that matches. The borrowing
// source itself is never its own lender, because its own abilities are
// already its own ("each other creature", "creatures you control that
// don't have the same name"). Types, control and counters are read after
// layers 1-5 have settled, and the lender's rows after its own layer 6
// (ADR 0155 Decision 3).
func BattlefieldLenders(p LenderPredicate) func(g *game.Game, source *game.Card) []*game.Card {
	return func(g *game.Game, source *game.Card) []*game.Card {
		if g == nil || g.Battlefield == nil || source == nil {
			return nil
		}
		var out []*game.Card
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID == source.InstanceID || !p(g, source, *c) {
				continue
			}
			out = append(out, c)
		}
		return out
	}
}

// LenderGraveyards says whose graveyards a graveyard lender reads.
type LenderGraveyards uint8

const (
	// YourGraveyard is "cards in your graveyard" (Thranduil, Trazyn): the
	// borrowing source's controller's.
	YourGraveyard LenderGraveyards = iota
	// AllGraveyards is "cards in all graveyards" (Necrotic Ooze, Mirran
	// Safehouse).
	AllGraveyards
)

// GraveyardLenders is "… the activated abilities of [cards matching P] in
// [your / all] graveyards". A card there has no layered characteristic, so
// the engine reads its printed abilities through its catalog entry and its
// designation gate (CR 113.6 keeps an ability that works only on the
// battlefield from being lent as usable there, which the activation path
// enforces). The borrowing source is never its own lender.
func GraveyardLenders(whose LenderGraveyards, p LenderPredicate) func(g *game.Game, source *game.Card) []*game.Card {
	return func(g *game.Game, source *game.Card) []*game.Card {
		if g == nil || source == nil {
			return nil
		}
		var out []*game.Card
		for _, seat := range g.Seats {
			if seat == nil || seat.Graveyard == nil {
				continue
			}
			if whose == YourGraveyard && seat.ID != source.Controller {
				continue
			}
			for i := range seat.Graveyard.Cards {
				c := &seat.Graveyard.Cards[i]
				if c.InstanceID == source.InstanceID || !p(g, source, *c) {
					continue
				}
				out = append(out, c)
			}
		}
		return out
	}
}

// BorrowedAbilitiesSpendableAsAnyColor is the rider "You may spend mana as
// though it were mana of any color to activate those abilities." for the
// borrow whose printed clause is `borrowLabel` (Drana and Linvala): the
// borrower's controller, for the rows that borrow lent and no others
// (ADR 0155 Decision 7).
func BorrowedAbilitiesSpendableAsAnyColor(label, borrowLabel string) []game.AnyColorSpendStatic {
	return []game.AnyColorSpendStatic{{
		Label: label,
		Whose: game.AnyColorSpendYou,
		Covers: func(ctx game.ManaSpendContext) bool {
			return ctx.AbilityOrigin.LentByBorrow(borrowLabel)
		},
	}}
}

// BorrowedAbilitiesSpendableAsAnyType is the wider rider "Mana of any
// type can be spent to activate <those> abilities." (Sharkey, Tyrant of
// the Shire): BorrowedAbilitiesSpendableAsAnyColor with AnyType, so
// colourless is a type too and coloured mana may also pay a {C} symbol
// in a lent cost (CR 106.1b no longer stops it).
func BorrowedAbilitiesSpendableAsAnyType(label, borrowLabel string) []game.AnyColorSpendStatic {
	out := BorrowedAbilitiesSpendableAsAnyColor(label, borrowLabel)
	out[0].AnyType = true
	return out
}

// checkStaticBorrows is ADR 0155 Decision 1's refusal. A borrow is an
// ability-adding effect, CR 613.1f's layer 6, and the engine applies it
// in its own step at the end of that bucket, so:
//
//   - declared in any other layer it would never be reached;
//   - it stands alone — a removal, a grant or an Apply beside it would be
//     applied in the borrow step's slot rather than its own timestamp's;
//   - it names its lenders and its printed clause, or it lends nothing
//     and shows nothing.
func checkStaticBorrows(card, what string, statics []game.StaticAbility) {
	for i, s := range statics {
		if s.Borrows == nil {
			continue
		}
		where := fmt.Sprintf("effects.Register: %q %s %d", card, what, i)
		switch {
		case s.Layer != game.Layer6Ability:
			panic(fmt.Sprintf("%s borrows abilities in layer %d — a borrow is a layer-6 effect (CR 613.1f)", where, s.Layer))
		case s.RemovesAbilities || len(s.GrantAbilities) > 0 || s.Apply != nil:
			panic(where + " borrows abilities beside a removal, a grant or an Apply — declare those as a static of their own")
		case s.AffectsSpells:
			panic(where + " borrows abilities for spells — a borrow lends to permanents")
		case s.AppliesTo == nil:
			panic(where + " borrows abilities for nothing — it has no AppliesTo")
		case s.Borrows.Lenders == nil:
			panic(where + " borrows abilities from nothing — it has no Lenders")
		case s.Borrows.Label == "":
			panic(where + " borrows abilities with no printed Label")
		}
	}
}

// --- exile lenders (ADR 0155 PR 3, #3201) ---------------------------------

// ExiledWithLenders is "… the activated abilities of all [cards matching
// P] exiled with this [permanent]": the cards in exile whose
// Card.ExiledWith names the borrowing source OBJECT (CR 607.2a), in exile
// order. The link is stamped by a linked exile (ExileTarget.Linked,
// Game.ExileLinkedForEffect), so a card another effect exiled is not a
// lender, and a source that left and came back is a new object with no
// claim on what its earlier self exiled (CR 400.7). A card that leaves
// exile loses the link with the move.
//
// The static that uses it declares DependsOnExile, so the borrow is
// re-read when a linked card leaves exile for a hand, a library or the
// stack; a counter change on an exiled card always re-reads it.
func ExiledWithLenders(p LenderPredicate) func(g *game.Game, source *game.Card) []*game.Card {
	return func(g *game.Game, source *game.Card) []*game.Card {
		if g == nil || source == nil {
			return nil
		}
		var out []*game.Card
		for _, c := range g.ExiledWithForEffect(game.PermissionCardRef{ID: source.InstanceID, Epoch: source.ObjectEpoch}) {
			if c.InstanceID != source.InstanceID && p(g, source, *c) {
				out = append(out, c)
			}
		}
		return out
	}
}

// ExileLenders is "… the activated abilities of all cards in exile
// [matching P]", with no link to the source: Rex, Cyber-Hound's "all
// cards in exile with brain counters on them", Mairsil's "all cards you
// own in exile with cage counters on them". Every card in exile that
// matches lends, whoever exiled it.
func ExileLenders(p LenderPredicate) func(g *game.Game, source *game.Card) []*game.Card {
	return func(g *game.Game, source *game.Card) []*game.Card {
		if g == nil || g.Exile == nil || source == nil {
			return nil
		}
		var out []*game.Card
		for i := range g.Exile.Cards {
			c := &g.Exile.Cards[i]
			if c.InstanceID != source.InstanceID && p(g, source, *c) {
				out = append(out, c)
			}
		}
		return out
	}
}

// CraftedWithLenders is "… each activated ability of the exiled cards
// used to craft it" (Locus of Enlightenment): CR 702.167c's craft link
// (Card.CraftedWith), the materials still in exile as the objects the
// craft cost put there. A material that left exile is gone from the link
// (CR 400.7).
func CraftedWithLenders() func(g *game.Game, source *game.Card) []*game.Card {
	return func(g *game.Game, source *game.Card) []*game.Card {
		mats := craftMaterialsOf(g, source)
		if len(mats) == 0 {
			return nil
		}
		out := make([]*game.Card, 0, len(mats))
		for i := range mats {
			out = append(out, &mats[i])
		}
		return out
	}
}

// LenderWithCounter is a LenderPredicate for a card with at least one `kind`
// counter on it: Rex's brain counters, Mairsil's cage counters.
func LenderWithCounter(kind string) LenderPredicate {
	return func(_ *game.Game, _ *game.Card, c game.Card) bool {
		return c.Counters[kind] > 0
	}
}

// AnyLender is a LenderPredicate that admits every card: Myr Welder's
// "all cards exiled with it", Territory Forge's "the exiled card".
func AnyLender(*game.Game, *game.Card, game.Card) bool { return true }

// exileLinkedTargetsEffect is exileLinkedTargets as an ability's Effect:
// Duplicant's imprint, Myr Welder's, Territory Forge's.
func exileLinkedTargetsEffect(g *game.Game, item *game.StackItem) error {
	return exileLinkedTargets(NewContext(g, item))
}

// exileLinkedTargets exiles every legal card target of the resolving item
// linked to its source object (ExileTarget.Linked): "exile target
// creature", "exile up to three target land cards". A target that left in
// response is skipped (CR 608.2b).
func exileLinkedTargets(ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (ExileTarget{Target: t.ID, Linked: true}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// --- library-top and chosen-permanent lenders (ADR 0155 PR 4, #3202) -----

// TopOfLibraryLender is "… the activated abilities of the top card of your
// library, as long as it is [a card matching P]" (Conspicuous Snoop's
// Goblin card, Skill Borrower's artifact or creature card): the top card
// of the borrowing source's controller's library, when it matches. The
// card is read where it is, so it lends its printed abilities (ADR 0155
// Decision 3).
//
// The static that uses it declares DependsOnLibraryTop, so the borrow is
// re-read the moment that library's top card changes — a draw, a cast
// from the top, a scry, a shuffle (game/library_top_watch.go).
func TopOfLibraryLender(p LenderPredicate) func(g *game.Game, source *game.Card) []*game.Card {
	return func(g *game.Game, source *game.Card) []*game.Card {
		if g == nil || source == nil {
			return nil
		}
		for _, seat := range g.Seats {
			if seat == nil || seat.ID != source.Controller || seat.Library == nil {
				continue
			}
			n := len(seat.Library.Cards)
			if n == 0 {
				return nil
			}
			top := &seat.Library.Cards[n-1]
			if top.InstanceID == source.InstanceID || !p(g, source, *top) {
				return nil
			}
			return []*game.Card{top}
		}
		return nil
	}
}

// ChosenPermanentLender is "… the activated abilities of the chosen
// permanent" (Scheming Fence): the permanent chosen as the source entered
// (game.Card.ChosenPermanent), while it is still that object on the
// battlefield. Nothing chosen, or a chosen permanent that has left, lends
// nothing.
func ChosenPermanentLender() func(g *game.Game, source *game.Card) []*game.Card {
	return func(g *game.Game, source *game.Card) []*game.Card {
		if g == nil {
			return nil
		}
		if c := g.ChosenPermanentOf(source); c != nil && c.InstanceID != source.InstanceID {
			return []*game.Card{c}
		}
		return nil
	}
}

// ChoosePermanentAsEnters builds the Spec.AsEnters for "As this permanent
// enters, you may choose a [permanent matching P]" (CR 614.12, Scheming
// Fence): a pick of at most one, the empty pick being "choose nothing".
// The answer is stored on the permanent (game.Card.ChosenPermanent) as
// the object it was.
func ChoosePermanentAsEnters(question string, p CardPredicate) func(*game.Card, *Context) error {
	return func(card *game.Card, ctx *Context) error {
		var ids []uuid.UUID
		for _, c := range ctx.Game.BattlefieldCardsForEffect() {
			if c.InstanceID != card.InstanceID && p(ctx.Game, card.Controller, c) {
				ids = append(ids, c.InstanceID)
			}
		}
		ctx.Game.QueueChoosePermanentAsEntersForEffect(card.Controller, card.InstanceID, question, ids)
		return nil
	}
}

// ChosenPermanentCantActivate is "Activated abilities of the chosen
// permanent can't be activated" (Scheming Fence), mana abilities
// included: the gate judges the SOURCE object of an activation, so the
// chosen permanent is stopped and a copy of its ability on the borrower
// is not (that copy's source is the borrower). Nothing chosen stops
// nothing.
func ChosenPermanentCantActivate(label string) game.ActivationRestriction {
	return game.ActivationRestriction{
		Label: label,
		Forbids: func(q game.ActivationQuery) bool {
			if q.FromZone != game.ZoneBattlefield {
				return false
			}
			chosen := q.Source.ChosenPermanent
			return chosen.ID != uuid.Nil && q.Card.InstanceID == chosen.ID && q.Card.ObjectEpoch == chosen.Epoch
		},
	}
}

// BorrowAbilitiesFor is "<affected> gains all activated abilities of
// <lender> until end of turn" from a resolving ability (ADR 0155
// Decision 9): Quicksilver Elemental, Grell Philosopher, Havengul Lich.
// Both sets are fixed as it resolves (CR 608.2h, 611.2c): the affected
// set by Target or Match, the lent rows read off Lender now. A lender with
// no activated ability registers nothing; one the catalog has no entry
// for registers a record that lends nothing and names it, so the
// recipient's menu says "<lender>: not automated" (ADR 0155 owner
// decision 2).
type BorrowAbilitiesFor struct {
	// Target pins the effect to one permanent. Ignored when Match is set.
	Target uuid.UUID
	// Match selects the affected permanents, evaluated once, now.
	Match CardPredicate
	// Lender is the object whose rows are lent.
	Lender *game.Card
	// Printed reads the lender as printed, wherever it is now (Havengul
	// Lich's "that card", the spell as it was cast) rather than with its
	// layered rows on the battlefield (Quicksilver Elemental's target).
	Printed bool
	// SpendFrom is the effect's own "you may spend <colour> mana as
	// though it were mana of any color to activate those abilities" —
	// Grell Philosopher's "U" — or empty.
	SpendFrom string
	// Duration is how long it lasts; zero is until end of turn.
	Duration game.Duration
	// Label is the borrow's printed clause, for the wire's
	// granted_abilities line and the borrow's identity.
	Label string
}

// Apply registers the borrow.
//
// Caller must be inside the resolution frame (holds g.mu write).
func (s BorrowAbilitiesFor) Apply(ctx *Context) error {
	if s.Lender == nil {
		return nil
	}
	var rows []game.LentRow
	if s.Printed {
		rows = game.PrintedLentRowsForEffect(*s.Lender)
	} else {
		rows = game.LentRowsForEffect(s.Lender)
	}
	mod := game.BorrowAbilitiesMod(s.SpendFrom, rows...)
	if len(rows) == 0 {
		m := game.UnautomatedBorrowMod(s.Lender)
		if m == nil {
			return nil
		}
		mod = *m
	}
	d := s.Duration
	if d.IsZero() {
		d = DurationUntilEndOfTurn(ctx)
	}
	return ScopedEffectFor{
		Target:   s.Target,
		Match:    s.Match,
		Mods:     []game.Mod{mod},
		Duration: d,
		Label:    s.Label,
	}.Apply(ctx)
}
