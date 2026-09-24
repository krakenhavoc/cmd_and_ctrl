package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Venser, Shaper Savant — Legendary Creature — Human Wizard {2}{U}{U},
// 2/2 (EDHREC rank 1967):
//
//	"Flash
//	 When Venser enters, return target spell or permanent to its
//	 owner's hand."
//
// The flash Man-o'-War that also answers a spell. The target clause is
// TargetSpellOrPermanent — one pick across the stack and the
// battlefield, Aang, Swift Savior's two-zone shape — with no narrowing
// on either half: any spell, any permanent, Venser included.
//
// A SPELL is returned, not countered (ReturnSpellToHand, Reprieve's
// verb): a spell that can't be countered still goes, nothing keyed to
// "countered" fires, and a copy of a spell ceases to exist in hand
// (CR 707.10a). A PERMANENT is bounced; a token bounced this way
// ceases to exist and a commander goes where its owner chooses
// (CR 903.9). Both halves re-check the target on resolution (CR
// 608.2b), so a spell that resolved in the meantime is not chased
// into the graveyard.
func init() {
	Register(Spec{
		OracleID:        "0f41cefc-d6ff-4db7-ba35-502b7e081de1",
		Name:            "Venser, Shaper Savant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets:   TargetSpellOrPermanent("target spell or permanent", nil, nil),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Venser, Shaper Savant — return target spell or permanent to its owner's hand",
					ReturnTargetSpellOrPermanentToHand)
			},
		}},
	})
}

// ReturnTargetSpellOrPermanentToHand is the body behind "return target
// spell or permanent to its owner's hand": a spell is returned from
// the stack WITHOUT being countered (ReturnSpellToHand — a spell that
// can't be countered still goes, and nothing keyed to a counter
// fires), a permanent is bounced. Reads the first still-legal target
// (CR 608.2b), so a target that left in response does nothing.
func ReturnTargetSpellOrPermanentToHand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		z := g.FindCardZoneForEffect(t.ID)
		if z == nil {
			return nil
		}
		switch z.Kind {
		case game.ZoneStack:
			return ReturnSpellToHand{StackID: t.ID}.Apply(ctx)
		case game.ZoneBattlefield:
			return BounceToHand{Target: t.ID}.Apply(ctx)
		}
		return nil
	}
	return nil
}
