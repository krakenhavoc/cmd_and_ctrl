package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eden, Seat of the Sanctum — Land — Town (EDHREC rank 3714):
//
//	"{T}: Add {C}.
//	 {5}, {T}: Mill two cards. Then you may sacrifice this land. When
//	 you do, return another target permanent card from your graveyard
//	 to your hand."
//
// A colourless utility land that turns into a Regrowth for
// permanents. The mana is a plain {C}. The second ability is ONE
// CR 602 activation with a decision in the middle of it and a
// reflexive trigger after the decision, and it is now written that
// way — the two gaps it was waiting on are both closed:
//
//   - the reflexive trigger (#636): "When you do, return another
//     target permanent card..." is a CR 603.12 trigger created by the
//     resolving ability, on the branch where the sacrifice actually
//     happened. It gets its own CR 603.3d target pick and its own
//     response window above the ability that made it.
//   - the free yes/no at resolution (#796): "Then you MAY sacrifice
//     this land" is a MayChoice addressed to the controller, asked
//     AFTER the mill.
//
// Until #796 this shipped as two activations that shared the {5} and
// the tap, with the sacrifice chosen up front (Insidious Fungus's
// split) — so the controller decided whether to sacrifice before
// seeing the two milled cards, and the card to return was picked
// before the mill and could never be one of the two cards the mill
// had just put there. Both simplifications are gone: the question is
// asked after the mill, and the reflexive trigger's target is chosen
// after the sacrifice, so a permanent card the mill just binned is a
// legal pick exactly as it is in paper.
//
// "Another" is OtherThan(source): Eden is in the graveyard by the
// time the reflexive trigger picks, and without the exclusion it
// would be a legal target for itself.
func init() {
	Register(Spec{
		OracleID:     "84856b92-5ce8-47f3-9a1c-78d6a3e26aca",
		Name:         "Eden, Seat of the Sanctum",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{
			{
				Label:  "{5}, {T}: Mill two cards. Then you may sacrifice this land. When you do, return another target permanent card from your graveyard to your hand.",
				Cost:   Plus(ManaCost("{5}"), TapCost()),
				Effect: edenMillTwoThenMaySacrifice,
			},
		},
	})
}

// edenMillTwoThenMaySacrifice is Eden's printed body: mill two, then
// ask.
//
// Caller holds g.mu.
func edenMillTwoThenMaySacrifice(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (MillCards{Player: item.Controller, N: 2}).Apply(ctx); err != nil {
		return err
	}
	return MayChoice{
		Question: "Eden, Seat of the Sanctum — sacrifice it to return a permanent card from your graveyard?",
		YesLabel: "Sacrifice Eden",
		NoLabel:  "Keep it",
		OnYes:    edenSacrificeThenReturn,
	}.Apply(ctx)
}

// edenSacrificeThenReturn is the "if you do" branch: Eden is
// sacrificed, and the CR 603.12 reflexive trigger that follows picks
// the permanent card to return.
//
// A package-level function reading everything off the Context it is
// handed — the undo contract every continuation in the catalog
// follows. The one thing it captures is the source's instance ID,
// into the "another" predicate, and that is a scalar.
//
// Caller holds g.mu.
func edenSacrificeThenReturn(ctx *Context) error {
	source := ctx.Source()
	if source == uuid.Nil {
		return nil
	}
	if err := ctx.Game.SacrificePermanentForEffect(source); err != nil {
		return err
	}
	return ReflexiveTrigger{
		Label:   "Eden, Seat of the Sanctum — return a permanent card from your graveyard",
		Targets: TargetCardInGraveyard("another target permanent card from your graveyard", YouOwn(), Permanent(), OtherThan(source)),
		Effect:  edenReturnChosenFromGraveyard,
	}.Apply(ctx)
}

// edenReturnChosenFromGraveyard is the reflexive trigger's body: the
// permanent card chosen when it went on the stack comes back to hand,
// re-checked at resolution (CR 608.2b) like every other target.
//
// Caller holds g.mu.
func edenReturnChosenFromGraveyard(g *game.Game, item *game.StackItem) error {
	return b34ReturnChosenGraveyardCardToHand(NewContext(g, item))
}
