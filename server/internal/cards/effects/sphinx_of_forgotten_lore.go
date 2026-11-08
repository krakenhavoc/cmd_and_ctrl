package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sphinx of Forgotten Lore — Creature — Sphinx {2}{U}{U}, 3/3
// (Foundations, #4465):
//
//	"Flash
//	 Flying
//	 Whenever this creature attacks, target instant or sorcery card in
//	 your graveyard gains flashback until end of turn. The flashback cost
//	 is equal to that card's mana cost."
//
// Snapcaster Mage's grant (ADR 0066) on an attack trigger: the permission
// names the targeted card object, its cost is the card's own mana cost,
// and it is flashback, so the card is exiled however it leaves the stack
// (CR 702.34a). Flash and flying ride PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ddb7e868-8f1d-4d57-9ee2-ebadb17b398a",
		Name:            "Sphinx of Forgotten Lore",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{Targeting(
			WheneverThisAttacks("Sphinx of Forgotten Lore — target card in your graveyard gains flashback", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				return GrantFlashbackToCard{
					Target: ts[0].ID,
					Label:  "Flashback — its mana cost (Sphinx of Forgotten Lore)",
				}.Apply(ctx)
			}),
			TargetCardInGraveyard("target instant or sorcery card in your graveyard",
				YouOwn(), Or(Instant(), Sorcery())),
		)},
	})
}
