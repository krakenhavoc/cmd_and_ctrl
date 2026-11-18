package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// obscuraAscendancySoul is the counter this enchantment collects. An on-disk
// identity: never renamed, never reused.
const obscuraAscendancySoul = "soul"

// Obscura Ascendancy — Enchantment {W}{U}{B} (Streets of New Capenna, #5491):
//
//	"Whenever you cast a spell, if its mana value is equal to 1 plus the
//	 number of soul counters on this enchantment, put a soul counter on this
//	 enchantment, then create a 2/2 white Spirit creature token with flying.
//	 As long as there are five or more soul counters on this enchantment,
//	 Spirits you control get +3/+3."
//
// The intervening "if" (CR 603.4) is read as the spell is cast and again as
// the trigger resolves, against the counters then on the same object; the
// spell's mana value is the one it had on the stack when it was cast (the
// trigger records it, CR 608.2h), so a countered spell does not change it.
// The anthem is Layer 7c over Spirits you control, on only while the
// enchantment has five or more soul counters.
//
// No simplification.
func init() {
	const label = "Obscura Ascendancy — put a soul counter on this enchantment, then create a 2/2 white Spirit with flying"
	anthem := TribalAnthem(TribeFilter{Tribes: []string{"Spirit"}, YoursOnly: true}, 3, 3)
	spirits := anthem.AppliesTo
	anthem.AppliesTo = func(target *game.Card, g *game.Game, source *game.Card) bool {
		return source.Counters[obscuraAscendancySoul] >= 5 && spirits(target, g, source)
	}
	anthem.Label = "As long as there are five or more soul counters on this enchantment, Spirits you control get +3/+3."
	Register(Spec{
		OracleID:     "34ec084d-2c16-452f-8dcc-06641c80f7d1",
		Name:         "Obscura Ascendancy",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{anthem},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && spellManaValueForEffect(g, ev.CardID) == 1+source.Counters[obscuraAscendancySoul]
			},
			Key: label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Amount = spellManaValueForEffect(g, ev.CardID)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Controller == uuid.Nil || sourceIsNewObject(g, item) || !onBattlefield(g, item.SourceCardID) {
					return nil
				}
				c, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok || item.Params.Amount != 1+c.Counters[obscuraAscendancySoul] {
					return nil
				}
				ctx := NewContext(g, item)
				if err := (AddCounter{Target: item.SourceCardID, Kind: obscuraAscendancySoul, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return CreateToken{Template: TokenCard("2/2 white Spirit with flying"), N: 1}.Apply(ctx)
			},
		}},
	})
}
