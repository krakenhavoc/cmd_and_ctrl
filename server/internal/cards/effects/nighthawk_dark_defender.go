package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nighthawk, Dark Defender — Legendary Creature — Human Hero {1}{W}, 2/1
// (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Whenever Nighthawk or another Hero you control enters, target creature
//	 gets +1/+1 until end of turn."
//
// Flying rides PrintedKeywords. The trigger fires for Nighthawk himself and
// for any other Hero entering under his controller (the type is read from
// the entering permanent), and the target is chosen as it goes on the
// stack, any creature.
//
// No simplification.
func init() {
	const label = "Nighthawk, Dark Defender — target creature gets +1/+1 until end of turn"
	Register(Spec{
		OracleID:        "eb8a373a-ba15-4b3f-959a-b7c708fea616",
		Name:            "Nighthawk, Dark Defender",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == source.InstanceID {
					return true
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.Controller == source.Controller && HasSubtype("Hero")(g, source.Controller, c)
			}, label, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: label}.Apply(ctx)
			}), TargetCreature("target creature")), ForTargets(BenefitUntilEOT(0, 1, 1))),
		},
	})
}
