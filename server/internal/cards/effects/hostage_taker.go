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
// "Until this creature leaves the battlefield" is two abilities
// (CR 610.3), Ossification's shape: the entry trigger exiles, and a
// leave trigger returns whatever that exile took, read back off the
// event log through b27ExiledWith on the entry trigger's label. The
// return is under the card's OWNER's control.
//
// Three details the printed card depends on:
//
//   - "ANOTHER target" is built per trigger from the source's own
//     instance (TargetsFrom + OtherThan), so Hostage Taker cannot
//     exile itself, and a second Hostage Taker can take the first.
//   - CR 610.3c: if Hostage Taker has already left the battlefield when
//     its entry trigger resolves, nothing is exiled at all. Otherwise
//     the card would be exiled with nothing left to bring it back.
//   - The cast permission is ExileWithPermission's airbend grant,
//     pointed at the Taker's controller rather than the owner: CastOnly
//     (the text says cast), for as long as the card stays in exile, and
//     payable with mana of any TYPE (AnyType, #1573), so a {C} in a
//     stolen card's cost is payable with any mana. A card cast this way
//     has left exile, so the leave trigger no longer returns it. The
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
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, hostageTakerExileLabel, hostageTakerExile)
				},
			},
			On(game.EventLTB, Self, "Hostage Taker — return the exiled card",
				b41ReturnCardsExiledWithToTheBattlefield(hostageTakerExileLabel)),
		},
	})
}

const hostageTakerExileLabel = "Hostage Taker — exile another target creature or artifact until this creature leaves the battlefield"

// hostageTakerExile is the entry trigger's resolution.
func hostageTakerExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	// CR 610.3c: a Hostage Taker that has already left exiles nothing.
	if info, ok := ctx.SourcePermanent(); !ok || info.Left {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return ExileWithPermission{
			Target:      t.ID,
			GrantTo:     item.Controller,
			CastOnly:    true,
			AnyType:     true,
			WhileExiled: true,
		}.Apply(ctx)
	}
	return nil
}
