package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Veinwitch Coven — Creature — Vampire Warlock {2}{B}, 3/3 (EDHREC
// rank 4176):
//
//	"Menace
//	 Whenever you gain life, you may pay {B}. If you do, return target
//	 creature card from your graveyard to your hand."
//
// A three-mana 3/3 that turns every lifelink trigger into a Raise
// Dead. In an aristocrats deck with a Blood Artist or a Zulaport
// Cutthroat it recurs a creature on every single death, which is why
// it costs three rather than five: the {B} per activation is the
// whole brake.
//
// Three things make the card, and each is a different mechanism:
//
//   - Menace is a printed keyword string.
//   - "Whenever you gain life" is per EVENT, not per point — one
//     lifegain of six is one trigger (Archangel of Thune's reading,
//     and the same helper).
//   - "You may pay {B}. If you do, …" is a real optional cost paid
//     DURING RESOLUTION, which is MayPay. The target, though, is
//     chosen when the trigger goes on the stack (CR 603.3d), and that
//     ordering is observable: an opponent who exiles the graveyard in
//     response makes the trigger fizzle whether or not you would have
//     paid, and you keep the {B}.
//
// Because the target is picked at announce, a trigger with no legal
// target — an empty graveyard — is never put on the stack at all,
// which is CR 603.3d and needs nothing here.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ab478ac9-af59-4df1-afef-5e9806c06643",
		Name:            "Veinwitch Coven",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverYouGainLife("Veinwitch Coven — pay {B} to return a creature card", func(g *game.Game, item *game.StackItem) error {
					if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
						return nil
					}
					id := item.Targets[0].ID
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{B}",
						Question: "Veinwitch Coven — pay {B} to return a creature card from your graveyard to your hand?",
						OnPay: func(ctx *Context) error {
							return ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(ctx)
						},
					}.Apply(NewContext(g, item))
				}),
				TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
			),
		},
	})
}
