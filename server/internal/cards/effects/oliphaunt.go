package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oliphaunt — Creature — Elephant {5}{R}, 6/4 (EDHREC rank 4155):
//
//	"Trample
//	 Whenever this creature attacks, another target creature you
//	 control gets +2/+0 and gains trample until end of turn.
//	 Mountaincycling {1} ({1}, Discard this card: Search your library
//	 for a Mountain card, reveal it, put it into your hand, then
//	 shuffle.)"
//
// A six-mana 6/4 trampler that is really played as a one-mana
// Mountain tutor: the body is a late-game top-deck and the
// mountaincycling is the reason it makes the deck.
//
// The attack trigger is until-end-of-turn continuous effects — one of
// the five mechanics the 2026-09-18 re-triage freed (#279):
//
//   - "+2/+0" is a Layer 7c modification, snapshotted at resolution
//     against the one creature the trigger targeted.
//   - "gains trample" is a Layer 6 grant on the same creature.
//
// "ANOTHER target creature you control" is exact. The clause is built
// per trigger through AnotherTarget (TriggeredAbility.TargetsFrom,
// which is handed the source), so the picker excludes THIS Oliphaunt
// by instance rather than by name: a second Oliphaunt — a Clone or a
// token copy of this one — is a legal target, as printed, and the
// Oliphaunt whose trigger it is never is. The resolution re-check
// (CR 608.2b) runs the same clause.
//
// Mountaincycling {1} arrived with #660, the same Typecycling
// constructor every other cycling card uses. "A Mountain card" is any
// land with the Mountain type, not only a basic one, so the search
// predicate is IsLandWithSubtype("Mountain") rather than
// IsBasicLand-scoped.
func init() {
	Register(Spec{
		OracleID:        "186b2256-4af3-48cb-96b0-b0e80a7ee6dc",
		Name:            "Oliphaunt",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []ActivatedAbility{
			Typecycling("Mountaincycling", "{1}", "a Mountain card", IsLandWithSubtype("Mountain")),
		},
		Triggered: []game.TriggeredAbility{
			oliphauntAttackTrigger(),
		},
	})
}

// oliphauntAttackTrigger is split out so TargetsFrom can be set
// directly — Targeting only takes a static *game.TargetSpec, and this
// clause needs the trigger's own source (see the card comment).
func oliphauntAttackTrigger() game.TriggeredAbility {
	t := WheneverThisAttacks("Oliphaunt — +2/+0 and trample until end of turn", func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		if err := (BoostUntilEOT{Target: id, Power: 2, Label: "Oliphaunt — +2/+0"}).Apply(ctx); err != nil {
			return err
		}
		return GrantKeywordUntilEOT{
			Target:   id,
			Keywords: []string{"trample"},
			Label:    "Oliphaunt — trample",
		}.Apply(ctx)
	})
	t.TargetsFrom = AnotherTarget(func(other CardPredicate) *game.TargetSpec {
		return TargetCreature("another target creature you control", YouControl(), other)
	})
	return t
}
