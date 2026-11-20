package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spell Queller — Creature — Spirit {1}{W}{U}, 2/3 (Innistrad Remastered):
//
//	"Flash
//	 Flying
//	 When this creature enters, exile target spell with mana value 4 or less.
//	 When this creature leaves the battlefield, the exiled card's owner may
//	 cast that card without paying its mana cost."
//
// Two triggers. The enter trigger exiles the spell with the link Smirking
// Spelljacker uses (the card's ExiledWith names this OBJECT, CR 400.7). The
// leave trigger reads the link off the source's last-known object
// (ctx.SourceRef), so it still finds the card once Queller is gone, and
// offers the cast to the card's OWNER, who may decline and leave it in
// exile. A spell that was countered or left the stack before the enter
// trigger resolved is not exiled and links nothing.
func init() {
	exile := Targeting(
		WhenThisEnters("Spell Queller — exile target spell with mana value 4 or less",
			gapExileTargetSpellLinked),
		TargetSpell("target spell with mana value 4 or less", ManaValueLE(4)))
	leave := On(game.EventLTB, Self, "Spell Queller — the exiled card's owner may cast it without paying its mana cost",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			ref, ok := ctx.SourceRef()
			if !ok {
				return nil
			}
			var cards []game.Card
			for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
				cards = append(cards, *c)
			}
			for _, c := range cards {
				if err := g.CastExiledFreeKeepingForEffect(c.Owner, item.SourceCardID, c.InstanceID); err != nil {
					return err
				}
			}
			return nil
		})
	Register(Spec{
		OracleID:        "135135f2-1643-40db-91f2-bb7b956d212c",
		Name:            "Spell Queller",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered:       []game.TriggeredAbility{exile, leave},
	})
}
