package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Colorstorm Stallion — Creature — Elemental Horse {1}{U}{R}, 3/3
// (Secrets of Strixhaven promo, #4834):
//
//	"Ward {1}, haste
//	 Opus — Whenever you cast an instant or sorcery spell, this creature
//	 gets +1/+1 until end of turn. If five or more mana was spent to cast
//	 that spell, create a token that's a copy of this creature."
//
// Opus is an ability word (CR 207.2c); the ability is one trigger on your
// instant and sorcery casts. The mana spent on the spell is counted as the
// ability triggers (the cast path records it before the event, and a waived
// payment asks what was paid: ReadsYourCasts), and carried on the trigger,
// so the ruling holds: the ability resolves before the spell, and even if
// that spell is countered or has left the stack, the count it was cast with
// is what the "if" reads. It is a clause of the effect, not an intervening
// "if", so the +1/+1 happens either way.
//
// The pump comes first and the copy second. The token is a copy of the
// Stallion's copiable values as the trigger resolves, or as it last existed
// if it has left (CR 707.2, CR 608.2h): it does not copy the +1/+1, and it
// has Opus too (leitmotifComposerCopy, the shared "copy of this creature").
//
// No simplification.
func init() {
	label := "Colorstorm Stallion — Opus: +1/+1 until end of turn; with five or more mana spent, copy this creature"
	opus := On(game.EventCast, YouCast(Or(Instant(), Sorcery())), label,
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			if sourceIsStillThisPermanent(g, item) {
				if err := (BoostUntilEOT{
					Target:    item.SourceCardID,
					Power:     1,
					Toughness: 1,
					Label:     "Colorstorm Stallion — +1/+1 until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			if item.Params.Amount < 5 {
				return nil
			}
			return leitmotifComposerCopy(g, item)
		})
	opus.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = g.StackItemPaidForEffect(ev.CardID).Spent().Total()
		return item
	}
	Register(Spec{
		OracleID:        "47ee6837-2e2f-4344-ab9f-6b8887874f77",
		Name:            "Colorstorm Stallion",
		ReadsManaSpent:  game.ReadsYourCasts,
		PrintedKeywords: []string{"haste"},
		Completeness:    CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{1}"), "Colorstorm Stallion — ward {1}"),
			opus,
		},
	})
}
