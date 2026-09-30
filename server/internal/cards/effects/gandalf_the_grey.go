package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// gandalfTheGreyLabel is the trigger's stack label, and with it the key
// its "hasn't been chosen" memory is kept under.
const gandalfTheGreyLabel = "Gandalf the Grey — you cast an instant or sorcery"

// Gandalf the Grey — Legendary Creature — Avatar Wizard {3}{U}{R}, 3/4:
//
//	"Whenever you cast an instant or sorcery spell, choose one that
//	 hasn't been chosen —
//	 • You may tap or untap target permanent.
//	 • Gandalf deals 3 damage to each opponent.
//	 • Copy target instant or sorcery spell you control. You may choose
//	   new targets for the copy.
//	 • Put Gandalf on top of its owner's library."
//
// ChooseOneNotChosen (ADR 0097) with no duration: the memory is on this
// Gandalf for as long as it is on the battlefield, so four instants or
// sorceries use the four bullets once each — the last of which puts
// Gandalf on top of its library. Recast, it is a new object with no
// memory (CR 400.7) and may choose all four again; that loop is the
// card.
//
// "You may tap or untap" is a pick between the two made as the bullet
// resolves; tapping a tapped permanent and untapping an untapped one do
// nothing, so the "may" needs no third option. The copy is the spell
// the trigger targets — normally the one that triggered it — with the
// CR 707.10c offer of new targets. "Put Gandalf on top of its owner's
// library" moves only the Gandalf that triggered: one that left and
// came back is a new object and stays where it is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dfd12e3c-2b2a-4461-a08d-09d4e6c4626e",
		Name:         "Gandalf the Grey",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			gandalfTheGreyTrigger(),
		},
	})
}

func gandalfTheGreyTrigger() game.TriggeredAbility {
	t := WheneverYouCast(Or(Instant(), Sorcery()), gandalfTheGreyLabel,
		func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = ChooseOneNotChosen(
		ModeDoing("You may tap or untap target permanent.", TargetPermanent("target permanent"),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return tapOrUntapPermanent(ctx, "Gandalf the Grey", t.ID)
			}),
		ModeDoing("Gandalf deals 3 damage to each opponent.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return damageToEachOpponent(ctx.Game, item, 3)
			}),
		ModeDoing("Copy target instant or sorcery spell you control. You may choose new targets for the copy.",
			instantOrSorcerySpell("target instant or sorcery spell you control", YouControl()),
			func(item *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return CopySpell{StackID: t.ID, Controller: item.Controller, ChooseNewTargets: true}.Apply(ctx)
			}),
		ModeDoing("Put Gandalf on top of its owner's library.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				if sourceIsNewObject(ctx.Game, item) || !onBattlefield(ctx.Game, item.SourceCardID) {
					return nil
				}
				return ctx.Game.TuckToLibraryForEffect(item.SourceCardID, false)
			}),
	)
	return t
}
