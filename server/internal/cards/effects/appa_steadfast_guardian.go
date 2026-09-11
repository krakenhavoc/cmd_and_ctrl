package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Appa, Steadfast Guardian — 3/4 Legendary Creature — Bison Ally
// for {2}{W}{W}:
//
//	"Flash
//	Flying
//	When Appa enters, airbend any number of other target nonland
//	permanents you control. (Exile them. While each one is exiled,
//	its owner may cast it for {2} rather than its mana cost.)
//	Whenever you cast a spell from exile, create a 1/1 white Ally
//	creature token."
//
// The two abilities are the two halves of one plan: airbend your
// own board in response to a wrath, then rebuy it at {2} a piece
// and collect an Ally for each rebuy. The second ability is why
// airbending your OWN permanents is a payoff rather than a cost.
//
// Both halves needed engine work that landed with this card:
//
//   - "Any number of other target …" is a genuinely multi-target
//     trigger (Min 0, Max unbounded). The pick_target prompt and
//     the client's multi-pick banner already handled the shape;
//     nothing in the catalog had used it from a trigger before.
//   - "Whenever you cast a spell FROM EXILE" reads EventCast's new
//     OldZone. CR 601.2a puts the card on the stack and it stops
//     remembering where it came from, so the fact has to be stamped
//     at emit time — see events.go.
//
// Shape notes and simplifications:
//
//   - "OTHER" is not expressible in the target clause (TargetSpec is
//     built at init(), NotSelf needs a live InstanceID), so the
//     picker offers Appa and the Effect skips him. Same gap and
//     same remedy as Deputy of Acquittals.
//   - A trigger with no legal target is dropped before the prompt
//     (CR 603.3d, engine-wide). For a Min-0 clause the printed card
//     would instead put the ability on the stack targeting nothing.
//     No card in the catalog can observe the difference.
//   - "Create a 1/1 WHITE Ally" — the token template carries no
//     color, matching every other template in tokens.go. Nothing
//     reads a token's color yet.
//   - The Ally trigger fires once per spell cast from exile,
//     including a spell Appa's own airbend put there, which is the
//     printed loop.
func init() {
	Register(Spec{
		OracleID:        "03c141ca-11e4-4927-a6bd-980ee1203c73",
		Name:            "Appa, Steadfast Guardian",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The picker lets you choose Appa itself as a target, but it is skipped — only your other permanents get airbent."},
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.CardID == source.InstanceID
				},
				Targets: TargetPermanent(
					"any number of other target nonland permanents you control",
					Nonland(), YouControl(),
				).WithCount(0, 0),
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Appa, Steadfast Guardian — airbend your permanents",
						func(g *game.Game, item *game.StackItem) error {
							ctx := NewContext(g, item)
							for _, t := range item.Targets {
								if t.Kind != game.TargetCard || t.ID == item.SourceCardID {
									continue
								}
								// CR 608.2b per slot: some targets may
								// have become illegal while the trigger
								// sat on the stack. The ability does as
								// much as it can.
								if !g.TargetStillLegalForEffect(item, t) {
									continue
								}
								if err := (Airbend{Target: t.ID}).Apply(ctx); err != nil {
									return err
								}
							}
							return nil
						})
				},
			},
			{
				Watches: []game.EventKind{game.EventCast},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Actor == source.Controller && ev.OldZone == game.ZoneExile
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Appa, Steadfast Guardian — create a 1/1 Ally",
						func(g *game.Game, item *game.StackItem) error {
							return CreateToken{
								Controller: item.Controller,
								Template:   WhiteAllyToken(),
								N:          1,
							}.Apply(NewContext(g, item))
						})
				},
			},
		},
	})
}
