package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hostage Taker — Creature — Human Pirate {2}{U}{B}, 2/3 (Edea
// steal-and-sac deck, #1565):
//
//	"When this creature enters, exile another target creature or
//	 artifact until this creature leaves the battlefield. You may cast
//	 that card for as long as it remains exiled, and mana of any type
//	 can be spent to cast that spell."
//
// "Until this creature leaves the battlefield" is CR 610.3, through
// ExileUntil (#1729): the entry trigger exiles, and the return is a
// one-shot effect performed immediately after the Taker leaves — not a
// trigger on the stack. The rulings of 2017-09-29 say both halves of
// that: "the exiled card returns to the battlefield immediately after
// Hostage Taker leaves the battlefield. Nothing happens between the two
// events, including state-based actions", and if the Taker's owner
// leaves the game "the exiled card will return to the battlefield under
// its owner's control. Because the one-shot effect that returns the
// card isn't an ability that goes on the stack, it won't cease to exist
// along with the leaving player's spells and abilities on the stack."
// The return is under the card's OWNER's control (CR 610.3c).
//
// The leave trigger the card carried before #1729 stays as a legacy
// row (UntilThisLeavesLegacyReturn), for a card exiled by an older
// binary.
//
// Three details the printed card depends on:
//
//   - "ANOTHER target" is built per trigger from the source's own
//     instance (TargetsFrom + OtherThan), so Hostage Taker cannot
//     exile itself, and a second Hostage Taker can take the first.
//   - CR 610.3b: if Hostage Taker has already left the battlefield when
//     its entry trigger resolves, nothing is exiled at all. Otherwise
//     the card would be exiled with nothing left to bring it back.
//   - The cast permission is ExileWithPermission's airbend grant,
//     pointed at the Taker's controller rather than the owner: CastOnly
//     (the text says cast), for as long as the card stays in exile, and
//     payable with mana of any TYPE (AnyType, #1573), so a {C} in a
//     stolen card's cost is payable with any mana. A card cast this way
//     has left exile, so the Taker leaving no longer returns it. The
//     creature you cast stays yours; you control the spell and the
//     permanent.
func init() {
	Register(Spec{
		OracleID:     "c5c2d209-e3ef-4b0d-85f5-e7402dcf09eb",
		Name:         "Hostage Taker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: b06SelfETB,
				TargetsFrom: func(_ game.TriggerContext, source *game.Card, _ *game.Game) *game.TargetSpec {
					return TargetPermanent("another target creature or artifact",
						Or(Creature(), Artifact()), OtherThan(source.InstanceID))
				},
				Key:    hostageTakerExileLabel,
				Effect: hostageTakerExile,
			},
			UntilThisLeavesLegacyReturn("Hostage Taker — return the exiled card", hostageTakerExileLabel),
		},
	})
}

const hostageTakerExileLabel = "Hostage Taker — exile another target creature or artifact until this creature leaves the battlefield"

// hostageTakerExile is the entry trigger's resolution.
func hostageTakerExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	// CR 610.3b: a Hostage Taker that has already left exiles nothing.
	if info, ok := ctx.SourcePermanent(); !ok || info.Left {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		grant := ExileWithPermission{
			GrantTo:     item.Controller,
			CastOnly:    true,
			AnyType:     true,
			WhileExiled: true,
		}.permission()
		return ExileUntil{
			Target:     t.ID,
			ThisLeaves: true,
			Label:      "Hostage Taker — the exiled card returns when Hostage Taker leaves the battlefield",
			Grant:      &grant,
		}.Apply(ctx)
	}
	return nil
}
