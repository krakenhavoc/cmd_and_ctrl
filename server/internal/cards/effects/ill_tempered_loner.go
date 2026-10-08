package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ill-Tempered Loner // Howlpack Avenger — {2}{R}{R} Creature — Human
// Werewolf 3/3 // Creature — Werewolf 4/4 (#2586, ADR 0132):
//
//	Front: "Whenever this creature is dealt damage, it deals that much
//	        damage to any target.
//	        {1}{R}: This creature gets +2/+0 until end of turn.
//	        Daybound"
//	Back:  "Whenever a permanent you control is dealt damage, this
//	        creature deals that much damage to any target.
//	        {1}{R}: This creature gets +2/+0 until end of turn.
//	        Nightbound"
//
// Screaming Nemesis's shape: EventDealDamage names the damaged
// permanent in Target, the controller picks any target as the trigger
// goes on the stack, and the creature deals the event's amount on
// resolution. The source is the Loner / Avenger itself, as printed, and
// one that died to the damage still deals it. "Any target" includes the
// creature itself. The back face watches every permanent its controller
// controls (including itself), so a planeswalker or battle takes part.
//
// DECLARED SIMPLIFICATION, weaker than printed, the same one Screaming
// Nemesis carries: the engine emits one damage event per SOURCE, so a
// creature damaged by two blockers at once triggers twice, once per
// blocker's damage, where the printed card triggers once for the total.
func init() {
	const oracle = "6e0b3317-394d-42cc-a350-cb5ce051787a"
	caveat := "If two or more sources damage it at the same time, it reflects each source's damage separately instead of the total in one go."
	pump := func(name string) ActivatedAbility {
		return ActivatedAbility{
			Label:  "{1}{R}: This creature gets +2/+0 until end of turn.",
			Cost:   ManaCost("{1}{R}"),
			Effect: thisCreatureUntilEOT(name+" — +2/+0", 2, 0),
		}
	}
	reflect := func(name string, when func(ev game.Event, source *game.Card, g *game.Game) bool) game.TriggeredAbility {
		label := name + " — deal that much damage to any target"
		return game.TriggeredAbility{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventDealDamage && ev.Amount > 0 && when(ev, source, g)
			},
			Targets: TargetAny(),
			Key:     label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Amount = ev.Amount
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Params.Amount <= 0 {
					return nil
				}
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: item.Params.Amount}.Apply(ctx)
				}
				return nil
			},
		}
	}
	Register(Spec{
		OracleID:        oracle,
		Name:            "Ill-Tempered Loner",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{caveat},
		PrintedKeywords: []string{"daybound"},
		Triggered: []game.TriggeredAbility{
			reflect("Ill-Tempered Loner", func(ev game.Event, source *game.Card, _ *game.Game) bool {
				return b35SelfWasDealtDamage(ev, source)
			}),
		},
		Activated: []ActivatedAbility{pump("Ill-Tempered Loner")},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Howlpack Avenger",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{caveat},
		PrintedKeywords: []string{"nightbound"},
		Triggered: []game.TriggeredAbility{
			reflect("Howlpack Avenger", func(ev game.Event, source *game.Card, g *game.Game) bool {
				damaged, ok := g.LookupCardForEffect(ev.Target)
				return ok && damaged.Controller == source.Controller && onBattlefield(g, ev.Target)
			}),
		},
		Activated: []ActivatedAbility{pump("Howlpack Avenger")},
	})
}
