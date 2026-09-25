package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Goose Mother — Legendary Creature — Bird Hydra {X}{G}{U}, 2/2
// (EDHREC rank 2928):
//
//	"Flying
//	 The Goose Mother enters with X +1/+1 counters on it.
//	 When The Goose Mother enters, create half X Food tokens, rounded
//	 up.
//	 Whenever The Goose Mother attacks, you may sacrifice a Food. If
//	 you do, draw a card."
//
// The Food commander. Flying rides PrintedKeywords. The X counters are
// the printed CR 614.1c entry clause and ride the CR 614 pipeline as
// one (XCounters, #1002).
//
// "When The Goose Mother enters, create half X Food tokens, rounded
// up" is a printed ENTERS TRIGGER, not an entry replacement — before
// #1312 an enters trigger's Build had nowhere to read the announced X
// from (the resolving spell's StackItem is gone by the time the
// trigger fires), so this used to move the read into OnResolve, a beat
// early, with a declared caveat that the Food arrived without a
// trigger on the stack to respond to. #1357: Build now reads
// source.CastX() (CastProvenance.X, CR 107.3m) and closes over the
// plain int rather than the card, so the trigger is a real object —
// it can be countered, or Goose Mother can be removed in response to
// it (the trigger still resolves on its own last-known-X, CR 603.10),
// which the resolve-time shortcut could not model.
//
// The attack trigger is Springbloom Druid's shape: "you may" is the
// trigger prompt, the Food is chosen through the pick_target prompt
// when the trigger goes on the stack ("a Food you control" — a Food
// you control can never be an illegal pick), and at resolution the
// Food is sacrificed if it is still there and the draw follows only
// from the sacrifice. With no Food the trigger is removed with no
// prompt (CR 603.3d), which matches "you may sacrifice a Food" with
// nothing to sacrifice.
//
// One declared simplification remains, weaker than printed: the Food
// to sacrifice is chosen when the attack trigger goes on the stack,
// not on resolution.
func init() {
	Register(Spec{
		OracleID:     "de595f1b-3f7d-45e0-a31b-ed23e5d1ee48",
		Name:         "The Goose Mother",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"You pick the Food to sacrifice when the attack trigger goes on the stack rather than on resolution.",
		},
		PrintedKeywords:            []string{"flying"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Key:       "The Goose Mother — create half X Food, rounded up",
				// #1312/#1357: read once, here (ADR 0041 P9's fill-in
				// Build), and stamp the plain int on Params.Amount
				// rather than closing over the card, which the
				// Effect must not capture.
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, "The Goose Mother — create half X Food, rounded up")
					item.Params.Amount = source.CastX()
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					x := item.Params.Amount
					if x <= 0 {
						return nil
					}
					return CreateToken{Controller: item.Controller, Template: FoodToken(), N: (x + 1) / 2}.Apply(NewContext(g, item))
				},
			},
			{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclared(ev, source)
				},
				OptionalPrompt: &game.TriggerOptionalPrompt{Question: "The Goose Mother — sacrifice a Food to draw a card?"},
				Targets:        TargetPermanent("a Food you control", And(HasSubtype("Food"), YouControl())),
				Key:            b27GooseMotherAttackLabel,
				Effect:         b27SacrificeChosenThenDraw,
			},
		},
	})
}
