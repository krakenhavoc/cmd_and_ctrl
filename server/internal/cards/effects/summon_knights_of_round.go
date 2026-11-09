package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Summon: Knights of Round — Enchantment Creature — Saga Knight
// {6}{W}{W}, 3/3 (Final Fantasy, #4462):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after V.)
//	 I, II, III, IV — Create three 2/2 white Knight creature tokens.
//	 V — Ultimate End — Other creatures you control get +2/+2 until end
//	 of turn. Put an indestructible counter on each of them.
//	 Indestructible"
//
// A Saga creature, Summon: Bahamut's shape. Chapter V picks "other
// creatures you control" once, as it resolves (CR 611.2c): those get
// +2/+2 until end of turn and an indestructible counter, and a creature
// that arrives later gets neither.
//
// No simplification.
func init() {
	const name = "Summon: Knights of Round"
	knights := func(chapter int) game.TriggeredAbility {
		return ChapterTrigger(chapter, SagaChapterLabel(name, chapter, "create three 2/2 white Knight creature tokens"),
			Do(CreateToken{Template: TokenCard("2/2 white Knight"), N: 3}))
	}
	Register(Spec{
		OracleID:        "73122880-d7f4-422f-a336-18773bcb7bfe",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Triggered: []game.TriggeredAbility{
			knights(1), knights(2), knights(3), knights(4),
			ChapterTrigger(5, SagaChapterLabel(name, 5, "Ultimate End — other creatures you control get +2/+2 until end of turn and an indestructible counter"),
				summonKnightsUltimateEnd),
		},
	})
}

// summonKnightsUltimateEnd is chapter V.
func summonKnightsUltimateEnd(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != item.SourceCardID && c.Controller == item.Controller && c.IsCreature() {
			ids = append(ids, c)
		}
	}
	for _, c := range ids {
		if err := (BoostUntilEOT{Target: c.InstanceID, Power: 2, Toughness: 2, Label: summonKnightsUltimateEndLabel}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
		if err := (AddCounter{Target: c.InstanceID, Kind: game.CounterIndestructible, N: 1}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

const summonKnightsUltimateEndLabel = "Summon: Knights of Round — Ultimate End: +2/+2 until end of turn"
